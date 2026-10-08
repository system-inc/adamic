package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// Lane 1 includes this with the shared readiness runtime when array contracts
// reach backend dispatch. check returns the value or a transitive checked view.
// Keeping it outside the array preserves identity and observes mutable aliases.
const viewArraysRuntime = `const adamicViewArray = (value, expression, expected) => {
    if (!Array.isArray(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + (value === null ? "null" : typeof value));
    return value;
};
const adamicViewArrayRead = (array, index, expression, check) => {
    const value = array[index];
    return check(value, expression + "[" + index + "]");
};
`

// Operands and contract checker are supplied by shared dispatch. The helper reads
// only the selected element; neither a field read nor this hook scans the array.
func emitViewArrayRead(array, index, expression, check string) string {
	return fmt.Sprintf("adamicViewArrayRead(%s, %s, %s, %s)", array, index, quote(expression), check)
}

// Presence and readiness share lane 1's state. The kind check deliberately does
// not call the element checker; elements are checked at their own read sites.
func emitViewArrayFieldRead(object, member, expression, expected string) string {
	return fmt.Sprintf("adamicViewArray(adamicReadField(%s, %s, %s, false, false, %s), %s, %s)", object, quote(member), quote(expression), quote(expected), quote(expression), quote(expected))
}

const viewArrayElementsRuntime = `const adamicViewArrayIndex = (array, index, relative, check) => {
    if (relative) { index = Math.trunc(Number(index)) || 0; if (index < 0) index += array.length; }
    if (!Number.isInteger(index) || index < 0 || index >= array.length) return undefined;
    if (!(index in array)) return undefined;
    return check(array[index]);
};
const adamicViewArrayElement = (value, expression, type, expected, allowed, required = false) => {
    if (value === undefined) { if (required) panic("element read failed: " + expression + " expected " + expected + ", found undefined"); return value; }
    const valid = type === 1 || type === 7 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) : type === 5 ? Array.isArray(value) : type === 8 ? adamicTypeOf(value) === "function" : false;
    if (!valid) panic("element read failed: " + expression + " expected " + expected + ", found " + (value === null ? "nullish" : Array.isArray(value) ? "array" : adamicTypeOf(value)));
    if (allowed.length && !allowed.includes(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + typeof value + " " + value);
    return value;
};
`

func (e *emitter) emitViewArrayRead(read ir.ArrayIndex) string {
	if read.Element == ir.Union {
		return e.emitPrimitiveArrayIndex(read)
	}
	array, index := e.value(read.Array), e.value(read.Index)
	allowed := []string{}
	for _, literal := range read.ViewAllowed {
		switch literal.Of {
		case ir.Number:
			allowed = append(allowed, strconv.FormatFloat(literal.Number, 'g', -1, 64))
		case ir.Boolean:
			allowed = append(allowed, strconv.FormatBool(literal.Boolean))
		case ir.String:
			allowed = append(allowed, quote(literal.String))
		}
	}
	checked := fmt.Sprintf("adamicViewArrayIndex(%s, %s, %t, (value) => adamicViewArrayElement(value, %s, %d, %s, [%s], %t))", array, index, read.Relative, quote(read.View), read.Element, quote(read.ViewType), strings.Join(allowed, ", "), !read.UndefinedAllowed)
	if read.Element == ir.Object && read.ViewContract != 0 {
		checked = "((adamicElement) => adamicElement === undefined ? undefined : " + e.viewObjectUnion(ir.Property{View: read.View, ViewContract: read.ViewContract}, "adamicElement") + ")(" + checked + ")"
	}
	return checked
}

const viewArrayOperationsRuntime = `const adamicArrayStorage = new WeakMap();
const adamicArrayStorageValue = (array, storage) => { adamicArrayStorage.set(array, storage === 1 || storage === 2 || storage === 7 ? storage : 10); return array; };
const adamicViewSlice = (array, arguments_) => adamicArrayElementCertificate(adamicArrayStorageValue(array.slice(...arguments_), adamicArrayStorage.get(array)), adamicArrayElementContracts.get(array) || 0);
const adamicArrayWriteCheck = (array, storage) => { const actual = adamicArrayStorage.get(array); if (actual !== (storage === 1 || storage === 2 || storage === 7 ? storage : 10)) panic("element read failed: <array write> expected " + (storage === 7 ? "number" : adamicViewTypeNames[storage] || "uncertified storage") + ", found " + (actual === 7 ? "number" : actual === 10 ? "heap pointers" : adamicViewTypeNames[actual] || "uncertified storage")); if (storage === 4 || storage === 5 || storage === 6 || storage === 8 || storage === 9 || storage === 10) panic("element read failed: <array write> expected " + (adamicViewTypeNames[storage] || "uncertified storage") + ", found uncertified source element contract"); };
const adamicViewMap = (array, callback, check) => array.map((value, index, all) => adamicCall(callback, [check(value), index, all]));
const adamicViewVisit = (array, method, callback, check) => adamicVisit(array, method, new AdamicClosure((self, values) => adamicCall(callback, [check(values[0]), values[1], values[2]]), []));
const adamicViewFind = (array, method, callback, check) => adamicFind(array, method, new AdamicClosure((self, values) => adamicCall(callback, [check(values[0]), values[1], values[2]]), []));
const adamicViewReduce = (array, callback, initial, check) => adamicReduce(array, new AdamicClosure((self, values) => adamicCall(callback, [values[0], check(values[1]), values[2], values[3]]), []), initial);
const adamicViewPop = (array, check) => { if (array.length === 0) return undefined; const index = array.length - 1; const value = index in array ? check(array[index]) : undefined; array.pop(); return value; };
const adamicViewPush = (array, value, storage) => { if (storage === 4) adamicArrayReferenceWrite(array, value); else adamicArrayWriteCheck(array, storage); return array.push(value); };
const adamicViewSetIndex = (array, index, value, storage, holes) => { if (storage === 4) adamicArrayReferenceWrite(array, value); else adamicArrayWriteCheck(array, storage); if (holes) array[index] = value; else adamicSetIndex(array, index, value); };
const adamicViewJoin = (array, separator, check) => array.map(value => check(value)).join(separator);
`

func (e *emitter) value(expression ir.Expression) string {
	if record, ok := expression.(ir.ArrayRecord); ok {
		return e.emitViewArrayRecord(record)
	}
	if properties, ok := expression.(ir.ArrayProperties); ok {
		return e.emitViewArrayProperties(properties)
	}
	if sort, ok := expression.(ir.ArraySort); ok && sort.OptionalComparator {
		return e.emitViewOptionalArraySort(sort)
	}
	if join, ok := expression.(ir.ArrayJoin); ok && join.Stringify {
		return e.emitViewArrayString(join)
	}
	value := e.valueWithoutViewArrays(expression)
	if !ir.HasArrayViews(e.program) {
		return value
	}
	value = e.viewArraySourceCertificate(expression, value)
	storage := ir.Type(0)
	switch expression := expression.(type) {
	case ir.ArrayLiteral:
		storage = expression.Element
	case ir.ArrayHoles:
		storage = expression.Element
	case ir.ArrayMap:
		storage = expression.Result
	case ir.ArrayFrom:
		storage = expression.Element
	case ir.ArrayFill:
		if expression.Array == nil {
			storage = expression.Element
		}
	case ir.ArrayVisit:
		if expression.Method == "filter" {
			storage = expression.Element
		}
	case ir.CodePoints:
		storage = ir.String
	case ir.MapEntries:
		storage = ir.Object
	case ir.StringCall:
		if expression.Method == "split" {
			storage = ir.String
		}
	}
	if storage != 0 {
		return fmt.Sprintf("adamicArrayStorageValue(%s, %d)", value, storage)
	}
	return value
}

func (e *emitter) viewArrayElementCheck(read ir.ArrayViewRead, value string) string {
	literals := []string{}
	for _, literal := range read.ViewAllowed {
		switch literal.Of {
		case ir.String:
			literals = append(literals, quote(literal.String))
		case ir.Number:
			literals = append(literals, strconv.FormatFloat(literal.Number, 'g', -1, 64))
		case ir.Boolean:
			literals = append(literals, strconv.FormatBool(literal.Boolean))
		}
	}
	checked := fmt.Sprintf("adamicViewArrayElement(%s, %s, %d, %s, [%s], %t)", value, quote(read.View), read.Element, quote(read.ViewType), strings.Join(literals, ", "), !read.UndefinedAllowed)
	if read.Element == ir.Object && read.ViewContract != 0 {
		checked = "((adamicElement) => adamicElement === undefined ? undefined : " + e.viewObjectUnion(ir.Property{View: read.View, ViewContract: read.ViewContract}, "adamicElement") + ")(" + checked + ")"
	}
	return checked
}

func (e *emitter) viewArrayChecker(read ir.ArrayViewRead) string {
	return "(adamicElement) => " + e.viewArrayElementCheck(read, "adamicElement")
}

func (e *emitter) emitViewArrayString(join ir.ArrayJoin) string {
	array := e.value(join.Array)
	if join.ViewRead.View == "" {
		return "String(" + array + ")"
	}
	checked := e.viewArrayElementCheck(join.ViewRead, "value")
	return fmt.Sprintf("((array) => array === undefined ? 'undefined' : array.map((value) => %s).join(','))(%s)", checked, array)
}

func (e *emitter) emitViewOptionalArraySort(sort ir.ArraySort) string {
	return fmt.Sprintf("((array, callback) => callback === undefined ? array.sort() : adamicSort(array, callback))(%s, %s)", e.value(sort.Array), e.value(sort.Callback))
}
