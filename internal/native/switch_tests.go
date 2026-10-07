package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Each case expression runs only if the preceding cases did not match, including
// cases grouped into one body. Its temporaries are released after the comparison.
func (e *emitter) switchTests(of ir.Type, held string, tests []ir.Expression) string {
	constant := true
	for _, test := range tests {
		switch test.(type) {
		case ir.NumberConstant, ir.StringConstant, ir.BooleanConstant:
		default:
			constant = false
		}
	}
	if constant {
		comparisons := []string{}
		for _, test := range tests {
			comparisons = append(comparisons, e.binary(ir.Equal, of, held, e.value(test)))
		}
		return unwrap(strings.Join(comparisons, " || "))
	}
	result := e.temporary()
	e.line("bool %s = false;", result)
	for _, test := range tests {
		e.line("if (!%s) {", result)
		text, value, owned := e.aside(test)
		e.out.WriteString(text)
		e.indent++
		e.line("%s = %s;", result, e.binary(ir.Equal, of, held, value))
		for index := len(owned) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", owned[index])
		}
		e.indent--
		e.line("}")
	}
	return result
}
