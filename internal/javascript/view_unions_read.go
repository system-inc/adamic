package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func init() { checkedViewReadAdapter = readUnionViewAt }

func readUnionViewAt(emitter *emitter, property ir.Property, object string) (string, bool) {
	id := property.ViewContract
	if id <= 0 || int(id) > len(emitter.program.ViewContracts) {
		return "", false
	}
	contract := emitter.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewUnion &&
		!contract.FixedTuple &&
		!(contract.Kind == ir.ViewCallable && contract.ProducerCertified) {
		return "", false
	}
	// An absent optional field reads undefined, so its declared union has to hold undefined. A
	// callable's recorded signature has no undefined to stand for.
	if contract.Unsupported != "" ||
		property.Absent && !contract.Undefined ||
		(property.Optional || property.Absent) && contract.Of == ir.Closure ||
		property.Method {
		panic("compiler bug: unavailable checked union read")
	}
	if object == "" {
		object = emitter.value(property.Object)
	}
	if property.Optional {
		return throughOptionalReceiver(emitter, property, object, readUnionViewAt), true
	}
	raw := fmt.Sprintf(
		"adamicReadField(%s,%s,%s,false,%t,%s)",
		object,
		quote(property.Name),
		quote(property.View),
		property.Absent,
		quote(contract.Name),
	)
	if contract.Of == ir.Closure {
		recorded := emitter.unionCallableRecorded("v")
		expected := emitter.untaggedCallableUnionExpected(property, recorded, emitter.unionCallableExpected(property))
		recorded = emitter.untaggedCallableRecorded(property, recorded, expected)
		return "((v)=>{" +
			unionCallableShapeRuntime +
			";return " +
			emitUnionCallableShape("v", recorded, expected, property.View, false) +
			";})(" +
			raw +
			")", true
	}
	checked := emitter.viewUntaggedObjectUnion(property, raw)
	if property.Of != ir.Union {
		test := map[ir.Type]string{
			ir.Number:       "typeof v==='number'",
			ir.Boolean:      "typeof v==='boolean'",
			ir.MaybeNumber:  "typeof v==='number'",
			ir.MaybeBoolean: "typeof v==='boolean'",
			ir.Object:       "adamicViewObject(v)",
			ir.Array:        "Array.isArray(v)",
			ir.String:       "typeof v==='string'",
			ir.Closure:      "adamicTypeOf(v)==='function'",
		}[property.Of]
		if contract.FixedTuple {
			test = "Array.isArray(v)"
		}
		if contract.Null && property.Of == contract.Of {
			test = "v===null || (" + test + ")"
		}
		if contract.Undefined && property.Of == contract.Of {
			test = "v===undefined || (" + test + ")"
		}
		if test == "" {
			panic("compiler bug: unavailable narrowed union read")
		}
		checked = "((v)=>{if (!(" +
			test +
			")) panic('field read failed: '+" +
			quote(property.View) +
			"+' narrowed read has a different member');return v;})(" +
			checked +
			")"
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
