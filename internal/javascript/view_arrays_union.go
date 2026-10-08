package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) arrayPrimitiveUnionRead(id ir.ViewContractID, value, expression, expected string) string {
	test := "false"
	if ir.PrimitiveArrayContract(e.program, id) {
		test = e.arrayPrimitiveUnionTest(id, "v")
	}
	return fmt.Sprintf("((v) => { if (!(%s)) panic(%s + (v === null ? 'null' : typeof v)); return v; })(%s)", test, quote("cast failed: element read failed: "+expression+" expected "+expected+", found "), value)
}

func (e *emitter) arrayPrimitiveUnionTest(id ir.ViewContractID, value string) string {
	c := e.program.ViewContracts[id-1]
	tests := []string{}
	if c.Null || c.Kind == ir.ViewNull {
		tests = append(tests, value+" === null")
	}
	if c.Undefined || c.Kind == ir.ViewUndefined {
		tests = append(tests, value+" === undefined")
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
		test := "typeof " + value + " === " + quote(map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[of])
		literals := []string{}
		for _, literal := range c.Allowed {
			constant := quote(literal.String)
			if of == ir.Number {
				constant = strconv.FormatFloat(literal.Number, 'g', -1, 64)
			}
			if of == ir.Boolean {
				constant = strconv.FormatBool(literal.Boolean)
			}
			literals = append(literals, value+" === "+constant)
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
