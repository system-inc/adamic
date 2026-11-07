package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Each case expression runs only if the preceding cases did not match, including
// cases grouped into one body. Its temporaries are released after the comparison.
func (nativeEmitter *emitter) switchTests(of ir.Type, held string, tests []ir.Expression) string {
	constant := true
	for _, testCase := range tests {
		switch testCase.(type) {
		case ir.NumberConstant, ir.StringConstant, ir.BooleanConstant:
		default:
			constant = false
		}
	}
	if constant {
		comparisons := []string{}
		for _, testCase := range tests {
			comparisons = append(comparisons, nativeEmitter.binary(ir.Equal, of, held, nativeEmitter.value(testCase)))
		}
		return unwrap(strings.Join(comparisons, " || "))
	}
	result := nativeEmitter.temporary()
	nativeEmitter.line("bool %s = false;", result)
	for _, testCase := range tests {
		nativeEmitter.line("if (!%s) {", result)
		text, value, owned := nativeEmitter.aside(testCase)
		nativeEmitter.out.WriteString(text)
		nativeEmitter.indent++
		nativeEmitter.line("%s = %s;", result, nativeEmitter.binary(ir.Equal, of, held, value))
		for index := len(owned) - 1; index >= 0; index-- {
			nativeEmitter.line("adamic_release(%s);", owned[index])
		}
		nativeEmitter.indent--
		nativeEmitter.line("}")
	}
	return result
}
