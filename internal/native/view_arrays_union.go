package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// The physical adapter only supplies a boxed snapshot. Membership is checked at
// the read, including finite literals and distinct null/undefined members.
func (e *emitter) arrayPrimitiveUnionRead(id ir.ViewContractID, value, expression, expected string) {
	if ir.MixedArrayContract(e.program, id) {
		e.viewUntaggedObjectUnion(ir.Property{View: expression, ViewType: expected, ViewContract: id}, value)
		return
	}
	test := "false"
	if ir.PrimitiveArrayContract(e.program, id) {
		test = e.arrayPrimitiveUnionTest(id, value)
	}
	e.declarations = append(e.declarations, "#include \"view_nullish.h\"")
	e.line("if (!(%s)) adamic_nullish_failure(%s, %s, %s);", test, cString(expression), cString(expected), value)
}

func (e *emitter) arrayPrimitiveUnionTest(id ir.ViewContractID, value string) string {
	c := e.program.ViewContracts[id-1]
	tests := []string{}
	if c.Null || c.Kind == ir.ViewNull {
		tests = append(tests, value+" == &adamic_null")
	}
	if c.Undefined || c.Kind == ir.ViewUndefined {
		tests = append(tests, value+" == NULL")
	}
	switch c.Kind {
	case ir.ViewNullable:
		tests = append(tests, e.arrayPrimitiveUnionTest(c.Element, value))
	case ir.ViewUnion:
		for _, member := range c.Members {
			tests = append(tests, e.arrayPrimitiveUnionTest(member, value))
		}
	case ir.ViewScalar:
		of := c.Of.Present()
		kind := map[ir.Type]string{ir.Number: "adamic_kind_number", ir.Boolean: "adamic_kind_boolean", ir.String: "adamic_kind_string"}[of]
		test := value + " != NULL && ((adamic_heap *)" + value + ")->kind == " + kind
		literals := []string{}
		for _, literal := range c.Allowed {
			switch of {
			case ir.Number:
				literals = append(literals, fmt.Sprintf("((adamic_number_box *)%s)->number == %s", value, cNumber(literal.Number)))
			case ir.Boolean:
				literals = append(literals, fmt.Sprintf("((adamic_boolean_box *)%s)->boolean == %t", value, literal.Boolean))
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
				literals = append(literals, fmt.Sprintf("adamic_string_equal((adamic_string *)%s,&%s)", value, name))
			}
		}
		if len(literals) > 0 {
			test += " && (" + strings.Join(literals, " || ") + ")"
		}
		tests = append(tests, "("+test+")")
	}
	if len(tests) == 0 {
		return "false"
	}
	return "(" + strings.Join(tests, " || ") + ")"
}
