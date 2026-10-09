package doc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// widthCases are the shapes most likely to separate a port from upstream, each for a stated reason.
var widthCases = []string{
	"",
	"plain ascii",
	"tab\there",                   // a control character takes the slow path and counts zero
	"中文字符",                        // wide
	"ｆｕｌｌｗｉｄｔｈ",                   // fullwidth forms
	"é vs é",                     // a combining mark counts zero
	"smile 😀 done",                // astral emoji, a surrogate pair in UTF-16
	"flag 🇺🇸",                     // regional indicator pair
	"family 👨‍👩‍👧",                // a ZWJ sequence is one emoji
	"keycap 1️⃣ and #️⃣",          // keycap sequences, the first branch of emoji-regex
	"heart ❤ vs ❤️",               // text presentation against emoji presentation
	"copyright © and ™",           // narrow emoji candidates
	"ambiguous ± §",               // ambiguous width counts 1
	"skin 👍🏽",                     // a modifier sequence
	"lone \xed\xa0\x80 surrogate", // invalid UTF-8, which both sides decode to U+FFFD
	"ｱｲｳ halfwidth katakana",      // halfwidth counts 1
	"\u00a0nbsp",
	"️ stray selector",
}

// TestStringWidthAgreesWithUpstream measures the port against the fork's own getStringWidth.
//
// Off unless COHERE_PRETTIER_FORK names the fork checkout, because it runs node. Then it checks the
// curated cases and, when COHERE_FORMAT_CORPORA is also set, every distinct line containing a
// non-ASCII byte in those repositories, which is the population that actually reaches the printer.
func TestStringWidthAgreesWithUpstream(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)

	inputs := append([]string(nil), widthCases...)
	seen := map[string]bool{}
	for _, input := range inputs {
		seen[input] = true
	}
	for _, corpus := range strings.Split(os.Getenv("COHERE_FORMAT_CORPORA"), ":") {
		if corpus == "" {
			continue
		}
		inputs = append(inputs, nonASCIILines(t, corpus, seen)...)
	}

	// Batched, because the corpus population is hundreds of megabytes and a single JSON payload
	// overflows node's maximum string length. That overflow is a failure to measure, not a mismatch,
	// and capping the population instead would quietly measure less than the corpus.
	const batchSize = 20000
	expected := make([]int, 0, len(inputs))
	for start := 0; start < len(inputs); start += batchSize {
		end := min(start+batchSize, len(inputs))
		expected = append(expected, upstreamWidths(t, root, inputs[start:end])...)
	}
	mismatches := 0
	for index, input := range inputs {
		if got := StringWidth(input); got != expected[index] {
			mismatches++
			if mismatches <= 20 {
				t.Errorf("StringWidth(%q) = %d, upstream %d", input, got, expected[index])
			}
		}
	}
	t.Logf("%d strings measured, %d curated, %d mismatches", len(inputs), len(widthCases), mismatches)
}

// TestWidthControlCanFail proves the comparison above can disagree, so its agreement means something.
func TestWidthControlCanFail(t *testing.T) {
	t.Parallel()
	if StringWidth("中") != 2 || StringWidth("a") != 1 || StringWidth("😀") != 2 || StringWidth("é") != 1 {
		t.Fatalf("basic widths wrong: 中=%d a=%d 😀=%d e+combining=%d",
			StringWidth("中"), StringWidth("a"), StringWidth("😀"), StringWidth("é"))
	}
}

// upstreamWidths asks the fork's getStringWidth, through node.
//
// stdin is decoded as UTF-8 by setEncoding, not by concatenating Buffer chunks. Concatenation decodes
// each chunk on its own, so a multi-byte character split across a 64 KB boundary became two U+FFFD
// and measured one or two columns wider. The first corpus run reported eight such "mismatches", all on
// non-ASCII lines, all with Go narrower; every one was the instrument splitting a character.
func upstreamWidths(t *testing.T, root string, inputs []string) []int {
	t.Helper()
	script := `
import getStringWidth from "./src/utilities/get-string-width.js";
process.stdin.setEncoding("utf8");
let data = ""; process.stdin.on("data", (chunk) => data += chunk);
process.stdin.on("end", () => process.stdout.write(JSON.stringify(JSON.parse(data).map(getStringWidth))));
`
	encoded, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--input-type=module", "-e", script)
	command.Dir = root
	command.Stdin = bytes.NewReader(encoded)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	var widths []int
	if err := json.Unmarshal(stdout.Bytes(), &widths); err != nil {
		t.Fatal(err)
	}
	if len(widths) != len(inputs) {
		t.Fatalf("upstream returned %d widths for %d inputs", len(widths), len(inputs))
	}
	return widths
}

// nonASCIILines collects distinct lines with a non-ASCII byte from the files the formatter is offered
// under root, nested repositories included, which is the population that actually reaches a printer.
//
// The first version walked every text file under the root and collected 1.08 GB across 607,275 lines,
// most of it data files no ignore layer lets Prettier see. That run failed on node's maximum string
// length, which was a failure to measure rather than a mismatch, and it was measuring the wrong
// population anyway. So the population comes from the same enumeration the format phase uses.
//
// Invalid UTF-8 is skipped: json.Marshal would replace it before node saw it, so the two sides would
// be measuring different strings, which is exactly the comparison that looks clean and proves nothing.
func nonASCIILines(t *testing.T, root string, seen map[string]bool) []string {
	t.Helper()
	engine, err := prettier.New(formatoptions.Default())
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	pending := []string{root}
	// Every corpus walked, logged once the loop ends, so a run that reaches fewer repositories says so.
	var measured []string
	for len(pending) > 0 {
		current := pending[0]
		pending = pending[1:]
		// Nested repositories regardless of current's own ignore lists, as the differential finds them
		// (#k6vebep): a repository under a path current never formats is still a corpus.
		nestedRepositories, err := formatfiles.NestedRepositoriesBelow(current)
		if err != nil {
			t.Fatalf("finding the repositories nested in %s: %v", current, err)
		}
		for _, nested := range nestedRepositories {
			pending = append(pending, filepath.Join(current, nested))
		}

		enumeration, err := engine.Enumerate(current)
		if errors.Is(err, formatoptions.ErrPrettierConfigRemains) {
			t.Logf("skipping %s, not yet adopted: %v", current, err)
			continue
		}
		if err != nil {
			t.Fatalf("enumerating %s: %v", current, err)
		}
		measured = append(measured, fmt.Sprintf("%s (%d)", current, len(enumeration.Files)))
		for _, path := range enumeration.Files {
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 1<<20), 1<<24)
			for scanner.Scan() {
				line := scanner.Text()
				if isPrintableASCII(line) || !utf8.ValidString(line) || seen[line] {
					continue
				}
				seen[line] = true
				lines = append(lines, line)
			}
			file.Close()
		}
	}
	t.Logf("corpora walked: %d\n  %s", len(measured), strings.Join(measured, "\n  "))
	return lines
}
