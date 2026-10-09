package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { checkedViewReadAdapter = readUnionView }

func readUnionView(e *emitter, property ir.Property) (string, bool) {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return "", false
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewUnion && !contract.FixedTuple && !(contract.Kind == ir.ViewCallable && contract.ProducerCertified) {
		return "", false
	}
	if contract.Unsupported != "" || property.Optional || property.Absent || property.Method {
		panic("compiler bug: unavailable checked union read")
	}
	raw := fmt.Sprintf("adamicReadField(%s,%s,%s,false,false,%s)", e.value(property.Object), quote(property.Name), quote(property.View), quote(contract.Name))
	if contract.Of == ir.Closure {
		recorded := e.unionCallableRecorded("v")
		expected := e.untaggedCallableUnionExpected(property, recorded, e.unionCallableExpected(property))
		recorded = e.untaggedCallableRecorded(property, recorded, expected)
		return "((v)=>{" + unionCallableShapeRuntime + ";return " + emitUnionCallableShape("v", recorded, expected, property.View, false) + ";})(" + raw + ")", true
	}
	checked := e.viewUntaggedObjectUnion(property, raw)
	if property.Of != ir.Union {
		test := map[ir.Type]string{ir.Object: "v!==null && typeof v==='object' && !Array.isArray(v) && !(v instanceof Map)", ir.Array: "Array.isArray(v)", ir.String: "typeof v==='string'", ir.Closure: "adamicTypeOf(v)==='function'"}[property.Of]
		if contract.FixedTuple {
			test = "Array.isArray(v)"
		}
		if contract.Undefined && property.Of == contract.Of {
			test = "v===undefined || (" + test + ")"
		}
		if test == "" {
			panic("compiler bug: unavailable narrowed union read")
		}
		checked = "((v)=>{if (!(" + test + ")) panic('field read failed: '+" + quote(property.View) + "+' narrowed read has a different member');return v;})(" + checked + ")"
	}
	return checked, true
}

func viewFieldRepresentation(value ir.Expression) int {
	switch value.(type) {
	case ir.Null:
		return int(ir.NullRepresentation)
	case ir.Undefined:
		if value.Type().IsReference() {
			return int(ir.UndefinedRepresentation)
		}
	}
	return int(value.Type())
}
