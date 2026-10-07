package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	regexpattern "github.com/system-inc/adamic/bridge/tsgo/checker/regex_pattern/pattern"
	regexsyntax "github.com/system-inc/adamic/bridge/tsgo/checker/regex_pattern/syntax"
	"strings"
	"unicode/utf16"
)

// regexPattern projects grammar scan status, character extents and escape/class extents.
func (p *Program) regexPattern(out *fields, node *ast.Node, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 3)
	if len(parts) != 3 || node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression {
		return "", fmt.Errorf("regex-pattern requires an invocation and flags/pattern")
	}
	flags, pattern := regexsyntax.ParseRegexFlags(parts[1]), parts[2]
	var characters []regexpattern.Character
	parsed := regexpattern.Walk(pattern, flags, func(character regexpattern.Character) bool { characters = append(characters, character); return true })
	out.yes(parsed)
	out.number(uint64(len(characters)))
	units := func(index int) uint64 { return uint64(len(utf16.Encode([]rune(pattern[:index])))) }
	for _, character := range characters {
		out.number(units(character.Start))
		out.number(units(character.End))
		out.text(pattern[character.Start:character.End])
	}
	type extent struct {
		start, end int
		ok         bool
	}
	var escapes, classes []extent
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' {
			width, ok := regexsyntax.SkipPatternEscape(pattern, i, flags)
			end := i
			if ok {
				end = i + width
			}
			escapes = append(escapes, extent{i, end, ok})
		}
		if pattern[i] == '[' {
			end, ok := regexsyntax.ClassEnd(pattern, i, flags)
			if !ok {
				end = i
			}
			classes = append(classes, extent{i, end, ok})
		}
	}
	for _, group := range [][]extent{escapes, classes} {
		out.number(uint64(len(group)))
		for _, entry := range group {
			out.number(units(entry.start))
			out.number(units(entry.end))
			out.yes(entry.ok)
		}
	}
	return out.String(), nil
}
