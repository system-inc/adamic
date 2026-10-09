// Command generate_classes captures the character classes Prettier's markdown printer splits and wraps
// text by, from the bundle the oracle runs, and writes them as Go.
//
// Usage:
//
//	go run ./internal/format/markdown/tools/generate_classes [-output <file.go>] [-check]
//
// splitText (src/language-markdown/utilities.js) splits prose on CJK_REGEXP and classifies each piece
// with PUNCTUATION_REGEXP and a Hangul test, and word.js and mdast.js test a space-separator class. All
// were built from Unicode property escapes, which goja does not implement, so the fork's build rewrote
// each into a literal class from the Node that built it. Those literals are the authority.
//
// The bundle is minified, so the generator does not trust variable names. It finds splitText by the
// split it opens with, reads the names of the three regexes from the way splitText uses them, cuts each
// literal out by that name, and asks goja about every code point, lone surrogates included, because
// upstream tests single UTF-16 units (`s[0]`, `s.at(-1)`) against them.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/system-inc/cohere/internal/format/prettier"
)

func main() {
	output := flag.String("output", filepath.Join("internal", "format", "markdown", "classes_generated.go"), "where to write the Go")
	check := flag.Bool("check", false, "regenerate and fail if the committed file disagrees")
	flag.Parse()

	generated, err := generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *check {
		committed, err := os.ReadFile(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !bytes.Equal(committed, generated) {
			fmt.Fprintf(os.Stderr, "%s is stale: regenerate it\n", *output)
			os.Exit(1)
		}
		return
	}

	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// The way splitText uses each regex, with the name captured.
var (
	cjkUse         = regexp.MustCompile(`split\(new RegExp\(` + "`" + `\(\$\{([A-Za-z_$]{1,3})\.source\}\)` + "`" + `,"u"\)\)`)
	punctuationUse = regexp.MustCompile(`hasLeadingPunctuation:([A-Za-z_$]{1,3})\.test\(s\[0\]\)`)
	hangulUse      = regexp.MustCompile(`if\(([A-Za-z_$]{1,3})\.test\(s\)\)\{n\(\{type:"word",value:s,kind:[A-Za-z_$]{1,3},isCJ:!1`)
	spaceUse       = regexp.MustCompile(`/(\[\\u\{20\}\\u\{A0\}[^/]*\\t\\n\\f\\r\])\$/u\.test`)
)

func generate() ([]byte, error) {
	bundles, err := prettier.Bundles()
	if err != nil {
		return nil, err
	}
	markdown := string(bundles.Files["plugins/markdown.js"])

	splitStart := strings.Index(markdown, `.split(/([\t\n ]+)/)`)
	if splitStart < 0 || strings.Count(markdown, `.split(/([\t\n ]+)/)`) != 1 {
		return nil, fmt.Errorf("could not find splitText's split exactly once")
	}
	splitText := markdown[splitStart : splitStart+1500]

	name := func(pattern *regexp.Regexp, what string) (string, error) {
		match := pattern.FindStringSubmatch(splitText)
		if match == nil {
			return "", fmt.Errorf("could not find how splitText uses %s", what)
		}
		return match[1], nil
	}
	cjkName, err := name(cjkUse, "CJK_REGEXP")
	if err != nil {
		return nil, err
	}
	punctuationName, err := name(punctuationUse, "PUNCTUATION_REGEXP")
	if err != nil {
		return nil, err
	}
	hangulName, err := name(hangulUse, "the Hangul regex")
	if err != nil {
		return nil, err
	}

	cjkLiteral, err := literal(markdown, cjkName)
	if err != nil {
		return nil, err
	}
	punctuationLiteral, err := literal(markdown, punctuationName)
	if err != nil {
		return nil, err
	}
	hangulLiteral, err := literal(markdown, hangulName)
	if err != nil {
		return nil, err
	}

	// CJK_REGEXP is `(?:[cjk])(?:[variation selectors])?`: the two classes are asked about separately.
	cjkParts := regexp.MustCompile(`^\(\?:(\[.*\])\)\(\?:(\[.*\])\)\?$`).FindStringSubmatch(cjkLiteral)
	if cjkParts == nil {
		return nil, fmt.Errorf("CJK_REGEXP does not have the shape (?:[...])(?:[...])?: %.80s", cjkLiteral)
	}

	spaces := spaceUse.FindAllStringSubmatch(markdown, -1)
	if len(spaces) == 0 {
		return nil, fmt.Errorf("could not find the space-separator class")
	}
	for _, space := range spaces[1:] {
		if space[1] != spaces[0][1] {
			return nil, fmt.Errorf("the space-separator class differs between uses")
		}
	}

	runtime := goja.New()
	classes := []struct {
		name        string
		description string
		pattern     string
	}{
		{"cjkRanges", "are the code points CJK_REGEXP's main class matches.", cjkParts[1]},
		{"variationSelectorRanges", "are the code points CJK_REGEXP's optional second class matches.", cjkParts[2]},
		{"punctuationRanges", "are the code points PUNCTUATION_REGEXP matches.", punctuationLiteral},
		{"hangulRanges", "are the code points the printer's Hangul test matches.", hangulLiteral},
		{"spaceSeparatorRanges", "are the code points of the class `[\\p{Space_Separator}\\t\\n\\f\\r]` was rewritten to.", spaces[0][1]},
	}

	var source bytes.Buffer
	source.WriteString("// Code generated by generate_classes. DO NOT EDIT.\n//\n")
	source.WriteString("// Captured from the embedded markdown bundle, run by goja: the rewritten classes the printer splits and\n")
	source.WriteString("// wraps prose by, each asked about every code point. Regenerate with:\n//\n")
	source.WriteString("//\tgo run ./internal/format/markdown/tools/generate_classes\n\n")
	source.WriteString("package markdown\n\n")

	answers := map[string][][2]int{}
	for _, class := range classes {
		ranges, err := codePointRanges(runtime, class.pattern)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", class.name, err)
		}
		answers[class.name] = ranges
		fmt.Fprintf(&source, "// %s %s\nvar %s = [][2]rune{\n", class.name, class.description, class.name)
		for _, entry := range ranges {
			fmt.Fprintf(&source, "\t{0x%04X, 0x%04X},\n", entry[0], entry[1])
		}
		source.WriteString("}\n\n")
	}

	// Controls in both directions.
	for _, probe := range []struct {
		class string
		point int
		want  bool
	}{
		{"cjkRanges", 0x4E2D, true}, {"cjkRanges", 'a', false},
		{"variationSelectorRanges", 0xFE0F, true}, {"variationSelectorRanges", 'a', false},
		{"punctuationRanges", '!', true}, {"punctuationRanges", 0x3002, true}, {"punctuationRanges", 'a', false},
		{"hangulRanges", 0xAC00, true}, {"hangulRanges", 0x4E2D, false},
		{"spaceSeparatorRanges", ' ', true}, {"spaceSeparatorRanges", 0x3000, true}, {"spaceSeparatorRanges", 'a', false},
	} {
		if contains(answers[probe.class], probe.point) != probe.want {
			return nil, fmt.Errorf("control failed: %s U+%04X answered %v, want %v", probe.class, probe.point, !probe.want, probe.want)
		}
	}

	return format.Source(source.Bytes())
}

// literal cuts the regex literal assigned to name out of the bundle, without its slashes or flags.
func literal(markdown string, name string) (string, error) {
	pattern := regexp.MustCompile(`[,; ]` + regexp.QuoteMeta(name) + `=/`)
	matches := pattern.FindAllStringIndex(markdown, -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("found %d assignments of %s to a regex literal, want exactly 1", len(matches), name)
	}
	start := matches[0][1]
	inClass := false
	for index := start; index < len(markdown); index++ {
		switch markdown[index] {
		case '\\':
			index++
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				if !strings.HasPrefix(markdown[index+1:], "u") {
					return "", fmt.Errorf("%s is not a u-flag regex", name)
				}
				return markdown[start:index], nil
			}
		}
	}
	return "", fmt.Errorf("%s's literal never closes", name)
}

// codePointRanges asks a u-flag regex, anchored, about every code point, lone surrogates included.
func codePointRanges(runtime *goja.Runtime, pattern string) ([][2]int, error) {
	runtime.Set("__pattern", pattern)
	value, err := runtime.RunString(`
		(function () {
			const regex = new RegExp("^(?:" + __pattern + ")$", "u");
			const ranges = [];
			let start = -1;
			for (let point = 0; point <= 0x110000; point++) {
				const inside = point <= 0x10FFFF && regex.test(String.fromCodePoint(point));
				if (inside && start < 0) start = point;
				if (!inside && start >= 0) { ranges.push([start, point - 1]); start = -1; }
			}
			return JSON.stringify(ranges);
		})()
	`)
	if err != nil {
		return nil, err
	}
	var ranges [][2]int
	if err := json.Unmarshal([]byte(value.String()), &ranges); err != nil {
		return nil, err
	}
	if len(ranges) == 0 {
		return nil, fmt.Errorf("matched no code point")
	}
	return ranges, nil
}

func contains(ranges [][2]int, point int) bool {
	for _, entry := range ranges {
		if point >= entry[0] && point <= entry[1] {
			return true
		}
	}
	return false
}
