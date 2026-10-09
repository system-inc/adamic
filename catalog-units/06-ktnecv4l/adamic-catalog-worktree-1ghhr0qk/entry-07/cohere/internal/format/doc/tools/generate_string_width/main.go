// Command generate_string_width captures the three tables Prettier's getStringWidth depends on, from
// the exact package versions the fork bundles, and writes them as Go.
//
// Usage:
//
//	go run ./internal/format/doc/tools/generate_string_width -prettier-root ~/Projects/system/prettier [-output <file.go>] [-check]
//
// Width decides where every line containing a wide character or an emoji breaks, so a table that is
// one Unicode version off produces a different line break and a different file, with nothing wrong
// to look at. That is why nothing here reads the packages' source. The wide table is computed by
// calling get-east-asian-width's own predicates on every code point and compressing the answers into
// ranges, so it cannot disagree with the code that consumes it, fast paths included.
//
// The two emoji regexes are copied, but translated. Both carry no `u` flag, so JavaScript matches them
// against UTF-16 code units, and emoji-regex spells every astral emoji as a surrogate pair. Go's
// regexp matches runes. So the Go side matches over a string with one rune per UTF-16 code unit, with
// surrogates moved into plane 16 where they are valid runes, and the regex escapes are moved the same
// way. A Go match is then a JavaScript match, unit for unit.
//
// The translator knows exactly the two escape forms the sources use today, `\uXXXX` and `\xHH`, and
// refuses anything else, as well as any character-class range that straddles the surrogate block,
// which the unit mapping would silently reorder. An upgrade that introduces new regex syntax fails
// here, loudly, rather than compiling into a pattern that means something else.
//
// `-check` regenerates and fails when the committed file disagrees. Node is required to run it and
// never to use the result.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// surrogateBase is where UTF-16 surrogate code units live as runes. Kept in step with stringWidthUnit
// in the doc package, which applies the same mapping to the text being measured.
const surrogateBase = 0x100000

const captureScript = `
import emojiRegex from "emoji-regex";
import { narrowEmojiRegexp } from "narrow-emojis";
import { _isFullWidth as isFullWidth, _isWide as isWide } from "get-east-asian-width";
import { readFileSync } from "node:fs";
// Read from disk rather than require(name/package.json): get-east-asian-width's exports map does not
// expose its package.json, and the version is a header comment, not something to resolve.
const version = (name) => JSON.parse(readFileSync("node_modules/" + name + "/package.json", "utf8")).version;

const ranges = [];
let start = -1;
for (let codePoint = 0; codePoint <= 0x10FFFF; codePoint++) {
  const wide = isFullWidth(codePoint) || isWide(codePoint);
  if (wide && start < 0) start = codePoint;
  if (!wide && start >= 0) { ranges.push([start, codePoint - 1]); start = -1; }
}
if (start >= 0) ranges.push([start, 0x10FFFF]);

const emoji = emojiRegex();
process.stdout.write(JSON.stringify({
  wideRanges: ranges,
  emojiSource: emoji.source, emojiFlags: emoji.flags,
  narrowSource: narrowEmojiRegexp.source, narrowFlags: narrowEmojiRegexp.flags,
  versions: {
    "emoji-regex": version("emoji-regex"),
    "get-east-asian-width": version("get-east-asian-width"),
    "narrow-emojis": version("narrow-emojis"),
  },
}));
`

type capture struct {
	WideRanges   [][2]int          `json:"wideRanges"`
	EmojiSource  string            `json:"emojiSource"`
	EmojiFlags   string            `json:"emojiFlags"`
	NarrowSource string            `json:"narrowSource"`
	NarrowFlags  string            `json:"narrowFlags"`
	Versions     map[string]string `json:"versions"`
}

func main() {
	prettierRoot := flag.String("prettier-root", "", "path to the Prettier fork checkout, whose node_modules hold the width packages")
	output := flag.String("output", filepath.Join("internal", "format", "doc", "string_width_generated.go"), "where to write the Go")
	check := flag.Bool("check", false, "regenerate and fail if the committed file disagrees")
	flag.Parse()

	if *prettierRoot == "" {
		fail("-prettier-root is required")
	}

	command := exec.Command("node", "--input-type=module", "-e", captureScript)
	command.Dir = *prettierRoot
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		fail("running node in %s: %v\n%s", *prettierRoot, err, stderr.String())
	}

	var captured capture
	if err := json.Unmarshal(stdout.Bytes(), &captured); err != nil {
		fail("decoding the capture: %v", err)
	}
	// A flag would change what the patterns mean: `u` makes them match code points, `i` folds case.
	// Only the global flag (which replace() uses and Go's ReplaceAll implies) and no flag are understood.
	if strings.Trim(captured.EmojiFlags, "g") != "" || strings.Trim(captured.NarrowFlags, "g") != "" {
		fail("unexpected regex flags: emoji %q, narrow %q", captured.EmojiFlags, captured.NarrowFlags)
	}

	emoji := translate("emoji-regex", captured.EmojiSource)
	narrow := translate("narrow-emojis", captured.NarrowSource)
	for name, source := range map[string]string{"emoji": emoji, "narrow": narrow} {
		if _, err := regexp.Compile(source); err != nil {
			fail("the translated %s pattern does not compile: %v", name, err)
		}
	}

	var file bytes.Buffer
	fmt.Fprintf(&file, "// Code generated by generate_string_width. DO NOT EDIT.\n//\n")
	fmt.Fprintf(&file, "// Captured from emoji-regex %s, get-east-asian-width %s and narrow-emojis %s, the versions the\n",
		captured.Versions["emoji-regex"], captured.Versions["get-east-asian-width"], captured.Versions["narrow-emojis"])
	fmt.Fprintf(&file, "// Prettier fork bundles. Regenerate with:\n//\n//\tgo run ./internal/format/doc/tools/generate_string_width -prettier-root <fork>\n\n")
	fmt.Fprintf(&file, "package doc\n\n")
	fmt.Fprintf(&file, "// wideRanges are the inclusive code point ranges for which get-east-asian-width's isFullWidth or isWide is true.\n")
	fmt.Fprintf(&file, "var wideRanges = [][2]rune{\n")
	for _, wideRange := range captured.WideRanges {
		fmt.Fprintf(&file, "\t{0x%X, 0x%X},\n", wideRange[0], wideRange[1])
	}
	fmt.Fprintf(&file, "}\n\n")
	fmt.Fprintf(&file, "// emojiPatternSource is emoji-regex, translated to match one rune per UTF-16 code unit.\n")
	fmt.Fprintf(&file, "const emojiPatternSource = %s\n\n", strconv.Quote(emoji))
	fmt.Fprintf(&file, "// narrowEmojiPatternSource is narrow-emojis' narrowEmojiRegexp, translated the same way.\n")
	fmt.Fprintf(&file, "const narrowEmojiPatternSource = %s\n", strconv.Quote(narrow))

	generated, err := format.Source(file.Bytes())
	if err != nil {
		fail("formatting the generated Go: %v", err)
	}

	if *check {
		committed, err := os.ReadFile(*output)
		if err != nil {
			fail("reading %s: %v", *output, err)
		}
		if !bytes.Equal(committed, generated) {
			fail("%s is stale: the width packages changed, regenerate it", *output)
		}
		fmt.Printf("%s matches the packages\n", *output)
		return
	}
	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fail("writing %s: %v", *output, err)
	}
	fmt.Printf("wrote %s: %d wide ranges\n", *output, len(captured.WideRanges))
}

// unitRune maps a UTF-16 code unit to the rune that stands for it on the Go side.
func unitRune(unit int) int {
	if unit >= 0xD800 && unit <= 0xDFFF {
		return surrogateBase + unit - 0xD800
	}
	return unit
}

func isSurrogate(unit int) bool { return unit >= 0xD800 && unit <= 0xDFFF }

// translate rewrites a no-flag JavaScript regex source into Go syntax over unit runes.
func translate(name string, source string) string {
	var out strings.Builder
	inClass := false
	previousUnit := -1
	pendingRange := false

	for index := 0; index < len(source); {
		character := source[index]
		unit := -1
		width := 1

		switch {
		case character == '\\':
			if index+1 >= len(source) {
				fail("%s: dangling backslash", name)
			}
			switch source[index+1] {
			case 'u':
				value, err := strconv.ParseUint(source[index+2:index+6], 16, 32)
				if err != nil {
					fail("%s: bad \\u escape at %d", name, index)
				}
				unit, width = int(value), 6
			case 'x':
				value, err := strconv.ParseUint(source[index+2:index+4], 16, 32)
				if err != nil {
					fail("%s: bad \\x escape at %d", name, index)
				}
				unit, width = int(value), 4
			default:
				fail("%s: escape \\%c at %d is not one the translator knows; extend it deliberately", name, source[index+1], index)
			}
		case character > 0x7F:
			fail("%s: literal non-ASCII at %d; the translator expects escapes only", name, index)
		}

		if unit >= 0 {
			if inClass && pendingRange && previousUnit >= 0 && isSurrogate(previousUnit) != isSurrogate(unit) {
				fail("%s: class range %X-%X straddles the surrogate block, which the unit mapping would reorder", name, previousUnit, unit)
			}
			fmt.Fprintf(&out, `\x{%X}`, unitRune(unit))
			previousUnit = unit
			pendingRange = false
			index += width
			continue
		}

		switch character {
		case '[':
			inClass = true
			previousUnit = -1
		case ']':
			inClass = false
		case '-':
			pendingRange = inClass && previousUnit >= 0
			index++
			out.WriteByte(character)
			continue
		default:
			if inClass {
				previousUnit = int(character)
			}
		}
		pendingRange = false
		out.WriteByte(character)
		index++
	}
	return out.String()
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "generate_string_width: "+format+"\n", arguments...)
	os.Exit(1)
}
