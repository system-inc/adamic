package checker

import (
	"regexp/syntax"
	"sort"
	"strings"
	"unicode"
)

// Serialize the regex compiler's immutable instruction program. No input string,
// match result, lint predicate, finding or edit is passed to this function.
func regexpProgramFacts(out *fields, question string) (string, error) {
	_, pattern, present := strings.Cut(question, "\n")
	if !present {
		out.yes(false)
		out.text("regexp-program requires a pattern")
		return out.String(), nil
	}
	expression, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		out.yes(false)
		out.text(err.Error())
		return out.String(), nil
	}
	program, err := syntax.Compile(expression.Simplify())
	if err != nil {
		out.yes(false)
		out.text(err.Error())
		return out.String(), nil
	}
	out.yes(true)
	out.number(uint64(program.Start))
	out.number(uint64(len(program.Inst)))
	for _, instruction := range program.Inst {
		out.text(instruction.Op.String())
		out.number(uint64(instruction.Out))
		out.number(uint64(instruction.Arg))
		ranges := instruction.Rune
		if len(ranges) == 1 {
			values := []rune{ranges[0]}
			if syntax.Flags(instruction.Arg)&syntax.FoldCase != 0 {
				for next := unicode.SimpleFold(ranges[0]); next != ranges[0]; next = unicode.SimpleFold(next) {
					values = append(values, next)
				}
			}
			sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
			ranges = nil
			for _, value := range values {
				ranges = append(ranges, value, value)
			}
		}
		out.number(uint64(len(ranges)))
		for _, value := range ranges {
			out.number(uint64(value))
		}
	}
	return out.String(), nil
}
