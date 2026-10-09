package printing

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
)

/*
 * The text utilities against upstream's, across the offset conversion.
 *
 * Upstream's run on UTF-16 indices and these on byte offsets, so the comparison maps every start index
 * from bytes to UTF-16 before asking node and maps every answer back. The alphabet is chosen for where
 * the two can disagree: U+2028 and U+2029, which skipNewline treats as newlines and which are three bytes
 * in UTF-8; the non-ASCII whitespace JavaScript's \s matches; a wide character and an astral emoji, which
 * are not whitespace but change where offsets land; and CRLF. The core's toy-language differential is
 * pure ASCII, so it never reached any of this.
 */

var textAlphabet = []string{" ", "\t", "\n", "\r", "\r\n", " ", " ", " ", "　", "\ufeff", "a", "中", "😀", "/", "*", ",", ";", "//", "/*", "*/"}

// byteToUnit and unitToByte convert between offsets at rune boundaries. -1, -2 and len map to themselves.
func offsetMaps(text string) (byteToUnit map[int]int, unitToByte map[int]int) {
	byteToUnit, unitToByte = map[int]int{}, map[int]int{}
	unit := 0
	for offset, character := range text {
		byteToUnit[offset], unitToByte[unit] = unit, offset
		unit += len(utf16.Encode([]rune{character}))
	}
	byteToUnit[len(text)], unitToByte[unit] = unit, len(text)
	for _, special := range []int{-1, -2} {
		byteToUnit[special], unitToByte[special] = special, special
	}
	return byteToUnit, unitToByte
}

type textCase struct {
	Text  string `json:"text"`
	Start int    `json:"start"`
}

type textAnswers struct {
	SkipWhitespaceForward, SkipWhitespaceBackward int
	SkipSpacesForward, SkipSpacesBackward         int
	SkipToLineEnd, SkipEverythingButNewLine       int
	SkipNewlineForward, SkipNewlineBackward       int
	HasNewlineForward, HasNewlineBackward         bool
	IsPreviousLineEmpty                           bool
	SkipInlineComment, SkipTrailingComment        int
}

const textUpstreamScript = `
import { skipWhitespace, skipSpaces, skipToLineEnd, skipEverythingButNewLine } from "./src/utilities/skip.js";
import skipNewline from "./src/utilities/skip-newline.js";
import hasNewline from "./src/utilities/has-newline.js";
import isPreviousLineEmpty from "./src/utilities/is-previous-line-empty.js";
import skipInlineComment from "./src/utilities/skip-inline-comment.js";
import skipTrailingComment from "./src/utilities/skip-trailing-comment.js";
process.stdin.setEncoding("utf8");
let input = ""; process.stdin.on("data", (chunk) => input += chunk);
process.stdin.on("end", () => {
  const f = (value) => value === false ? -2 : value;
  const back = { backwards: true };
  process.stdout.write(JSON.stringify(JSON.parse(input).map(({ text, start }) => ({
    SkipWhitespaceForward: f(skipWhitespace(text, start)), SkipWhitespaceBackward: f(skipWhitespace(text, start, back)),
    SkipSpacesForward: f(skipSpaces(text, start)), SkipSpacesBackward: f(skipSpaces(text, start, back)),
    SkipToLineEnd: f(skipToLineEnd(text, start)), SkipEverythingButNewLine: f(skipEverythingButNewLine(text, start)),
    SkipNewlineForward: f(skipNewline(text, start)), SkipNewlineBackward: f(skipNewline(text, start, back)),
    HasNewlineForward: hasNewline(text, start), HasNewlineBackward: hasNewline(text, start, back),
    IsPreviousLineEmpty: isPreviousLineEmpty(text, start),
    SkipInlineComment: f(skipInlineComment(text, start)), SkipTrailingComment: f(skipTrailingComment(text, start)),
  }))));
});
`

// TestTextAgreesWithUpstream is the differential. Off unless COHERE_PRETTIER_FORK names the fork.
func TestTextAgreesWithUpstream(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)
	random := rand.New(rand.NewSource(20261003))
	type pending struct {
		text      string
		byteStart int
	}
	var cases []textCase
	var originals []pending
	for len(cases) < 20000 {
		var builder strings.Builder
		for count := 1 + random.Intn(10); count > 0; count-- {
			builder.WriteString(textAlphabet[random.Intn(len(textAlphabet))])
		}
		text := builder.String()
		byteToUnit, _ := offsetMaps(text)
		// Starts at every rune boundary and both ends, since edges are where off-by-one lives.
		var starts []int
		for offset := range text {
			starts = append(starts, offset)
		}
		starts = append(starts, len(text), -1)
		start := starts[random.Intn(len(starts))]
		cases = append(cases, textCase{Text: text, Start: byteToUnit[start]})
		originals = append(originals, pending{text: text, byteStart: start})
	}

	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--input-type=module", "-e", textUpstreamScript)
	command.Dir = root
	command.Stdin = bytes.NewReader(encoded)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	var upstream []textAnswers
	if err := json.Unmarshal(stdout.Bytes(), &upstream); err != nil {
		t.Fatal(err)
	}

	mismatches, unmapped, nonASCII := 0, 0, 0
	for index, original := range originals {
		text, start := original.text, original.byteStart
		if !isASCIIOnly(text) {
			nonASCII++
		}
		_, unitToByte := offsetMaps(text)
		want := upstream[index]
		mapped := func(unit int) int {
			if offset, present := unitToByte[unit]; present {
				return offset
			}
			// Upstream stepped one UTF-16 unit back into an astral character and stopped on its low
			// surrogate, a position with no byte offset. The byte version stops at that character's first
			// byte. Every caller only tests which character sits at the position, and neither a lone
			// surrogate nor an astral rune's first byte equals any character these functions look for, so
			// the two are equivalent. That is a claim, so the comparison holds it exactly: Go must answer
			// the start of the very rune upstream's unit sits inside, and these are counted separately.
			if offset, present := unitToByte[unit-1]; present {
				unmapped++
				return offset
			}
			return -100
		}
		got := textAnswers{
			SkipWhitespaceForward: SkipWhitespace(text, start, false), SkipWhitespaceBackward: SkipWhitespace(text, start, true),
			SkipSpacesForward: SkipSpaces(text, start, false), SkipSpacesBackward: SkipSpaces(text, start, true),
			SkipToLineEnd: SkipToLineEnd(text, start, false), SkipEverythingButNewLine: SkipEverythingButNewLine(text, start, false),
			SkipNewlineForward: SkipNewline(text, start, false), SkipNewlineBackward: SkipNewline(text, start, true),
			HasNewlineForward: HasNewline(text, start, false), HasNewlineBackward: HasNewline(text, start, true),
			IsPreviousLineEmpty: IsPreviousLineEmpty(text, start),
			SkipInlineComment:   SkipInlineComment(text, start), SkipTrailingComment: SkipTrailingComment(text, start),
		}
		expected := textAnswers{
			SkipWhitespaceForward: mapped(want.SkipWhitespaceForward), SkipWhitespaceBackward: mapped(want.SkipWhitespaceBackward),
			SkipSpacesForward: mapped(want.SkipSpacesForward), SkipSpacesBackward: mapped(want.SkipSpacesBackward),
			SkipToLineEnd: mapped(want.SkipToLineEnd), SkipEverythingButNewLine: mapped(want.SkipEverythingButNewLine),
			SkipNewlineForward: mapped(want.SkipNewlineForward), SkipNewlineBackward: mapped(want.SkipNewlineBackward),
			HasNewlineForward: want.HasNewlineForward, HasNewlineBackward: want.HasNewlineBackward,
			IsPreviousLineEmpty: want.IsPreviousLineEmpty,
			SkipInlineComment:   mapped(want.SkipInlineComment), SkipTrailingComment: mapped(want.SkipTrailingComment),
		}
		if got != expected {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("text %q start %d\nwant %+v\ngot  %+v", text, start, expected, got)
			}
		}
	}
	t.Logf("%d cases, %d with non-ASCII, %d answers upstream lands mid surrogate pair (held to the containing rune's start), %d mismatches", len(cases), nonASCII, unmapped, mismatches)
	if nonASCII < len(cases)/2 {
		t.Errorf("only %d of %d cases contain non-ASCII text, which is the population this test exists for", nonASCII, len(cases))
	}
}

func isASCIIOnly(text string) bool {
	for index := 0; index < len(text); index++ {
		if text[index] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
