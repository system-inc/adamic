package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/bridge/tsgo/wave07_regex/regexpattern"
	"github.com/system-inc/adamic/bridge/tsgo/wave07_regex/regexsyntax"
	"strings"
	"unicode/utf16"
)

// wave07RegexStructure exposes raw regex character spans and scanner boundaries.
// The native caller decides validity gates, rewrite eligibility and replacement text.
func (p *Program) wave07RegexStructure(node *ast.Node, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("wave07-regex-structure requires flags and pattern")
	}
	pattern := parts[2]
	flags := regexsyntax.ParseRegexFlags(parts[1])
	out := new(fields)
	out.number(1)
	out.text("wave07-regex-structure")
	offset := func(at int) uint64 { return uint64(len(utf16.Encode([]rune(pattern[:at])))) }
	var characters []regexpattern.Character
	scanned := regexpattern.Walk(pattern, flags, func(character regexpattern.Character) bool { characters = append(characters, character); return true })
	out.yes(scanned)
	out.number(uint64(len(characters)))
	for _, character := range characters {
		out.number(offset(character.Start))
		out.number(offset(character.End))
	}
	type boundary struct {
		start, end int
		ok         bool
	}
	var boundaries []boundary
	for at := 0; at < len(pattern); at++ {
		switch pattern[at] {
		case '\\':
			width, ok := regexsyntax.SkipPatternEscape(pattern, at, flags)
			end := at
			if ok {
				end += width
			}
			boundaries = append(boundaries, boundary{at, end, ok})
		case '[':
			end, ok := regexsyntax.ClassEnd(pattern, at, flags)
			if !ok {
				end = at
			}
			boundaries = append(boundaries, boundary{at, end, ok})
		}
	}
	out.number(uint64(len(boundaries)))
	for _, boundary := range boundaries {
		out.number(offset(boundary.start))
		out.number(offset(boundary.end))
		out.yes(boundary.ok)
	}
	return out.String(), nil
}
