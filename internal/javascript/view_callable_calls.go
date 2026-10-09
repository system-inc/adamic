package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) viewCallableDomain(id ir.ViewContractID, value string) string {
	c := e.program.ViewContracts[id-1]
	result := "false"
	switch c.Kind {
	case ir.ViewUnknown:
		if c.Name == "unknown" && c.Of == ir.Union {
			result = "true"
		} else if c.Name == "object" && c.Of == ir.Union {
			result = value + " !== null && (typeof " + value + " === 'object' || adamicTypeOf(" + value + ") === 'function')"
		}
	case ir.ViewUndefined:
		result = value + " === undefined"
	case ir.ViewNull:
		result = value + " === null"
	case ir.ViewUnion:
		tests := []string{}
		for _, child := range c.Members {
			tests = append(tests, "("+e.viewCallableDomain(child, value)+")")
		}
		result = strings.Join(tests, " || ")
	case ir.ViewScalar:
		kind := map[ir.Type]string{ir.Number: "number", ir.MaybeNumber: "number", ir.Boolean: "boolean", ir.MaybeBoolean: "boolean", ir.String: "string"}[c.Of]
		result = "typeof " + value + " === " + quote(kind)
		if len(c.Allowed) > 0 {
			tests := []string{}
			for _, lit := range c.Allowed {
				literal := strconv.FormatFloat(lit.Number, 'g', -1, 64)
				if lit.Of == ir.Boolean {
					literal = fmt.Sprint(lit.Boolean)
				}
				if lit.Of == ir.String {
					literal = quote(lit.String)
				}
				tests = append(tests, value+" === "+literal)
			}
			result += " && (" + strings.Join(tests, " || ") + ")"
		}
	case ir.ViewObject:
		tests := []string{value + " !== null", "typeof " + value + " === 'object'", "!Array.isArray(" + value + ")", "!(" + value + " instanceof Map)"}
		for _, field := range c.Fields {
			check := e.viewCallableDomain(field.Contract, "slot.value")
			absent := "false"
			if field.Optional {
				absent = "!Object.hasOwn(" + value + ", " + quote(field.Name) + ")"
			}
			tests = append(tests, "((slot) => slot === undefined ? "+absent+" : Object.hasOwn(slot, 'value') && !adamicFieldReadiness.get("+value+")?.has("+quote(field.Name)+") && ("+check+"))(Object.getOwnPropertyDescriptor("+value+", "+quote(field.Name)+"))")
		}
		result = strings.Join(tests, " && ")
	case ir.ViewArray:
		result = "Array.isArray(" + value + ") && " + value + ".every(element => " + e.viewCallableDomain(c.Element, "element") + ")"
	}
	if c.Undefined {
		result = value + " === undefined || (" + result + ")"
	}
	return "(" + result + ")"
}

// Passing an already formed argument array guarantees every argument expression
// runs before checks. The callee snapshot precedes the argument array.
func (e *emitter) viewCallableInvoke(call ir.CallClosure, p ir.Property) string {
	target := e.program.ViewContracts[call.CallContract-1]
	var b strings.Builder
	b.WriteString("((prepared, arguments_) => { const object = prepared.object; const value = adamicViewAdapterUnderlying(prepared.value); const code = value instanceof AdamicClosure ? value.code : value;\n")
	for index, f := range e.program.Functions {
		method := !f.Closure && f.MethodName != ""
		if !f.Closure && !method || f.CallableResultName == "" || f.RestElement != 0 {
			continue
		}
		offset := 0
		if f.Receiver || method {
			offset = 1
		}
		if len(f.Parameters)-offset != len(f.CallableParameters) {
			continue
		}
		valid := true
		for _, id := range f.CallableParameters {
			valid = valid && id != 0
		}
		if !valid {
			continue
		}
		fmt.Fprintf(&b, "if (code === %s) {\n", functionName(e.program, index))
		if f.CallableReceiver != 0 {
			message := fmt.Sprintf("callable call failed: %s at %s thisArg expected producer %s, view %s", p.View, call.CallWhere, e.program.ViewContracts[f.CallableReceiver-1].Name, p.ViewType)
			fmt.Fprintf(&b, "if (!%s) panic(%s);\n", e.viewCallableDomain(f.CallableReceiver, "object"), quote(message))
		}

		for i, id := range f.CallableParameters {
			actual := e.program.ViewContracts[id-1].Name
			declared := "undefined"
			if i < len(target.Parameters) {
				declared = e.program.ViewContracts[target.Parameters[i]-1].Name
			}
			message := fmt.Sprintf("callable call failed: %s at %s argument %d expected producer %s, view %s", p.View, call.CallWhere, i+1, actual, declared)
			guard := ""
			if call.CheckBound {
				guard = fmt.Sprintf("arguments_.length > %d && ", i)
			}
			fmt.Fprintf(&b, "if (%s!%s) panic(%s);\n", guard, e.viewCallableDomain(id, fmt.Sprintf("arguments_[%d]", i)), quote(message))
		}
		if call.CheckBound {
			b.WriteString("return undefined; }\n")
			continue
		}
		invocation := "adamicCall(value, arguments_)"
		if f.Receiver {
			invocation = "adamicCall(value, [object, ...arguments_])"
		}
		if method {
			invocation = "code(object, ...arguments_)"
			if f.ArgumentsCount != 0 {
				invocation = "adamicDirect(code, [object, ...arguments_])"
			}
		}
		fmt.Fprintf(&b, "const result = %s;\n", invocation)
		if call.Returns != 0 {
			message := fmt.Sprintf("callable call failed: %s at %s result expected view %s, producer %s", p.View, call.CallWhere, e.program.ViewContracts[target.Result-1].Name, f.CallableResultName)
			fmt.Fprintf(&b, "if (!%s) panic(%s);\n", e.viewCallableDomain(target.Result, "result"), quote(message))
		}
		b.WriteString("return result; }\n")
	}
	b.WriteString("panic(" + quote("callable call failed: "+p.View+" at "+call.CallWhere+" has no checkable producer signature") + "); })")
	return b.String()
}

func (e *emitter) emitDirectViewCallable(call ir.CallClosure, p ir.Property) string {
	raw := "adamicViewCallablePreparedRead(object, " + quote(p.Name) + ", " + quote(p.View) + ", true, false, true, " + quote(p.ViewType) + ")"
	invocation := e.viewCallableInvoke(call, p) + "({object, value: " + raw + "}, [" + e.callValues(call.Arguments, call.Spread) + "])"
	if p.Optional {
		invocation = "object === undefined ? undefined : " + invocation
	}
	return "((object) => " + invocation + ")(" + e.value(p.Object) + ")"
}
