package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

func (e *emitter) viewCallableDomain(id ir.ViewContractID) string {
	c := e.program.ViewContracts[id-1]
	name := e.temporary()
	var b strings.Builder
	fmt.Fprintf(&b, "static bool %s(adamic_view_union_value value) {\n", name)
	b.WriteString("(void)value;\n")
	if c.Undefined {
		b.WriteString("if (value.kind == adamic_view_union_undefined) return true;\n")
	}
	switch c.Kind {
	case ir.ViewUnknown:
		if c.Name == "unknown" && c.Of == ir.Union {
			b.WriteString("return true;\n")
		} else if c.Name == "object" && c.Of == ir.Union {
			b.WriteString("return value.kind == adamic_view_union_object || value.kind == adamic_view_union_array || value.kind == adamic_view_union_map || value.kind == adamic_view_union_function;\n")
		} else {
			b.WriteString("return false;\n")
		}
	case ir.ViewUndefined:
		b.WriteString("return value.kind == adamic_view_union_undefined;\n")
	case ir.ViewNull:
		b.WriteString("return value.kind == adamic_view_union_null;\n")
	case ir.ViewUnion:
		tests := []string{}
		for _, child := range c.Members {
			tests = append(tests, e.viewCallableDomain(child)+"(value)")
		}
		fmt.Fprintf(&b, "return %s;\n", strings.Join(tests, " || "))
	case ir.ViewScalar:
		kind := map[ir.Type]string{ir.Number: "number", ir.MaybeNumber: "number", ir.Boolean: "boolean", ir.MaybeBoolean: "boolean", ir.String: "string"}[c.Of]
		fmt.Fprintf(&b, "if (value.kind != adamic_view_union_%s) return false;\n", kind)
		tests := []string{}
		for _, lit := range c.Allowed {
			switch lit.Of {
			case ir.Number:
				tests = append(tests, "value.payload.number == "+strconv.FormatFloat(lit.Number, 'g', -1, 64))
			case ir.Boolean:
				tests = append(tests, fmt.Sprintf("value.payload.boolean == %t", lit.Boolean))
			case ir.String:
				text := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", text, cString(lit.String)))
				tests = append(tests, fmt.Sprintf("adamic_string_equal(value.payload.reference, &%s)", text))
			}
		}
		if len(tests) == 0 {
			tests = append(tests, "true")
		}
		fmt.Fprintf(&b, "return %s;\n", strings.Join(tests, " || "))
	case ir.ViewObject:
		b.WriteString("if (value.kind != adamic_view_union_object || value.payload.reference == NULL) return false;\n")
		for _, field := range c.Fields {
			child := e.viewCallableDomain(field.Contract)
			fmt.Fprintf(&b, "{ adamic_view_union_value slot; if (!adamic_view_untagged_plain_slot(NULL, &value, %s, &slot)) {", cString(field.Name))
			if !field.Optional {
				b.WriteString("return false;")
			} else {
				fmt.Fprintf(&b, "if (adamic_has_property(value.payload.reference, %s)) return false;", cString(field.Name))
			}
			fmt.Fprintf(&b, "} else if (!%s(slot)) return false; }\n", child)
		}
		b.WriteString("return true;\n")
	case ir.ViewArray:
		child := e.viewCallableDomain(c.Element)
		b.WriteString("if (value.kind != adamic_view_union_array || value.payload.reference == NULL) return false;\nconst adamic_array *array = value.payload.reference;\nsize_t slots = array->sparse == NULL ? array->length : array->sparse->count;\nfor (size_t i=0; i<slots; i++) { double index=(double)i; if(array->sparse != NULL) { const adamic_map_entry *entry=&array->sparse->entries[i]; if(entry->deleted) continue; index=entry->key.number; } adamic_value *slot=adamic_array_holes_at(array,index); if(slot==NULL) continue; adamic_view_union_value actual={adamic_view_union_unknown,*slot};\n")
		b.WriteString("if(array->references) actual=adamic_view_union_heap(slot->reference); else if(array->element_kind==adamic_rep_number) actual.kind=adamic_view_union_number; else if(array->element_kind==adamic_rep_boolean) actual.kind=adamic_view_union_boolean; else if(array->element_kind==adamic_rep_maybe_number) { adamic_maybe_number n=adamic_maybe_number_unpack(slot->number); actual.kind=n.present?adamic_view_union_number:adamic_view_union_undefined; actual.payload.number=n.number; } else if(array->element_kind==adamic_rep_maybe_boolean) { adamic_maybe_boolean n=adamic_maybe_boolean_unpack(slot->maybe_boolean); actual.kind=n.present?adamic_view_union_boolean:adamic_view_union_undefined; actual.payload.boolean=n.boolean; }\n")
		fmt.Fprintf(&b, "if(!%s(actual)) return false; } return true;\n", child)
	default:
		b.WriteString("return false;\n")
	}
	b.WriteString("}\n")
	e.declarations = append(e.declarations, b.String())
	return name
}

func viewCallableSnapshot(of ir.Type, slot string) string {
	switch of {
	case 0:
		return "(adamic_view_union_value){adamic_view_union_undefined,{.reference=NULL}}"
	case ir.Number:
		return "(adamic_view_union_value){adamic_view_union_number," + slot + "}"
	case ir.Boolean:
		return "(adamic_view_union_value){adamic_view_union_boolean," + slot + "}"
	case ir.MaybeNumber:
		n := "adamic_maybe_number_unpack(" + slot + ".number)"
		return "(adamic_view_union_value){" + n + ".present?adamic_view_union_number:adamic_view_union_undefined,{.number=" + n + ".number}}"
	case ir.MaybeBoolean:
		n := "adamic_maybe_boolean_unpack(" + slot + ".maybe_boolean)"
		return "(adamic_view_union_value){" + n + ".present?adamic_view_union_boolean:adamic_view_union_undefined,{.boolean=" + n + ".boolean}}"
	}
	return "adamic_view_union_heap(" + slot + ".reference)"
}

// Typed nullable references use NULL for null; boxed unions use adamic_null.
func (e *emitter) viewCallableDomainHas(id ir.ViewContractID, kind ir.ViewKind) bool {
	if id == 0 {
		return false
	}
	c := e.program.ViewContracts[id-1]
	if c.Kind == kind || kind == ir.ViewUndefined && c.Undefined {
		return true
	}
	if c.Kind == ir.ViewUnion {
		for _, child := range c.Members {
			if e.viewCallableDomainHas(child, kind) {
				return true
			}
		}
	}
	return false
}

func (e *emitter) viewCallableDomainSnapshot(of ir.Type, slot string, id ir.ViewContractID) string {
	value := viewCallableSnapshot(of, slot)
	if of.IsReference() && of != ir.Union && e.viewCallableDomainHas(id, ir.ViewNull) && !e.viewCallableDomainHas(id, ir.ViewUndefined) {
		return "(" + slot + ".reference == NULL ? (adamic_view_union_value){adamic_view_union_null,{.reference=&adamic_null}} : " + value + ")"
	}
	return value
}

func viewCallableConverted(from, to ir.Type, slot, snapshot string) (string, bool) {
	if from == to {
		return slot + "." + member(to), false
	}
	switch to {
	case ir.Number:
		return snapshot + ".payload.number", false
	case ir.Boolean:
		return snapshot + ".payload.boolean", false
	case ir.MaybeNumber:
		return "adamic_maybe_number_pack((adamic_maybe_number){" + snapshot + ".kind!=adamic_view_union_undefined," + snapshot + ".payload.number})", false
	case ir.MaybeBoolean:
		return "adamic_maybe_boolean_pack((adamic_maybe_boolean){" + snapshot + ".kind!=adamic_view_union_undefined," + snapshot + ".payload.boolean})", false
	case ir.Union:
		if from == 0 {
			return "NULL", false
		}
		value, fresh := converted(from, to, unslotted(from, slot+"."+member(from)))
		return "(" + snapshot + ".kind == adamic_view_union_null ? &adamic_null : " + value + ")", fresh
	default:
		return "(" + snapshot + ".kind == adamic_view_union_null ? NULL : " + snapshot + ".payload.reference)", false
	}
}

// The field is snapshotted before arguments. All argument expressions have run
// before this dispatch tests values or invokes producer code.
func (e *emitter) directViewCallableInvoke(call ir.CallClosure, property ir.Property) string {
	e.declarations = append(e.declarations, "#include \"view_unions_untagged.h\"")
	name := e.temporary()
	methodType := "adamic_method"
	if e.program.ClosureConventionNeeded() {
		methodType = "adamic_method_entry"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "static adamic_value %s(adamic_closure *self, adamic_object *receiver, %s method, adamic_value *arguments, size_t count) {\n(void)self; (void)receiver; (void)method; (void)arguments; (void)count;\n", name, methodType)
	// Re-viewing retains only the new view contract, not another adapter layer.
	fmt.Fprintln(&b, "self = adamic_view_adapter_underlying(self);")
	target := e.program.ViewContracts[call.CallContract-1]
	resultCheck := ""
	if call.Returns != 0 {
		resultCheck = e.viewCallableDomain(target.Result)
	}
	for index, f := range e.program.Functions {
		methodProducer := !f.Closure && f.MethodName != "" && e.dispatchable(index)
		if !f.Closure && !methodProducer || f.CallableResultName == "" || f.RestElement != 0 {
			continue
		}
		valid := true
		for _, id := range f.CallableParameters {
			valid = valid && id != 0
		}
		if !valid {
			continue
		}
		offset, receiverParameter := 0, 0
		if f.Receiver || methodProducer {
			receiverParameter = 1
		}
		if f.Receiver && !methodProducer {
			offset = 1
		}
		if len(f.Parameters)-receiverParameter != len(f.CallableParameters) {
			continue
		}
		identity := "self != NULL && " + e.unionClosureCodeIdentity("self", index)
		if methodProducer {
			thunk := e.methodThunk(index)
			identity = "self == NULL && method == " + thunk
			if e.program.ClosureConventionNeeded() {
				identity = "self == NULL && !method.counted && method.code == " + thunk
				if e.program.PackedCountNeeded(index) {
					identity = "self == NULL && method.counted && method.counted_code == " + thunk
				}
			}
		}
		fmt.Fprintf(&b, "if (%s) {\n", identity)
		if f.CallableReceiver != 0 {
			domain := e.viewCallableDomain(f.CallableReceiver)
			message := fmt.Sprintf("callable call failed: %s at %s thisArg expected producer %s, view %s", property.View, call.CallWhere, e.program.ViewContracts[f.CallableReceiver-1].Name, property.ViewType)
			fmt.Fprintf(&b, "if (!%s(adamic_view_union_heap((const adamic_heap *)receiver))) adamic_panic(%s, sizeof %s - 1);\n", domain, cString(message), cString(message))
		}

		n := max(1, len(f.Parameters), len(call.Arguments)+offset, e.program.FixedArgumentSlots+offset)
		fmt.Fprintf(&b, "adamic_value adapted[%d] = {{.reference=NULL}};\n", n)
		if offset != 0 {
			b.WriteString("adapted[0].reference=receiver;\n")
		}
		releases := []string{}
		for i, id := range f.CallableParameters {
			from := ir.Type(0)
			slot := "((adamic_value){.reference=NULL})"
			if i < len(call.Arguments) {
				from = call.Arguments[i].Type()
				slot = fmt.Sprintf("arguments[%d]", i)
			}
			snapshot := fmt.Sprintf("argument_%d", i)
			sourceDomain := ir.ViewContractID(0)
			if i < len(target.Parameters) {
				sourceDomain = target.Parameters[i]
			}
			fmt.Fprintf(&b, "adamic_view_union_value %s = count > %d ? %s : (adamic_view_union_value){adamic_view_union_undefined,{.reference=NULL}};\n", snapshot, i, e.viewCallableDomainSnapshot(from, slot, sourceDomain))
			domain := e.viewCallableDomain(id)
			actual := e.program.ViewContracts[id-1].Name
			declared := "undefined"
			if i < len(target.Parameters) {
				declared = e.program.ViewContracts[target.Parameters[i]-1].Name
			}
			message := fmt.Sprintf("callable call failed: %s at %s argument %d expected producer %s, view %s", property.View, call.CallWhere, i+1, actual, declared)
			fmt.Fprintf(&b, "if (!%s(%s)) adamic_panic(%s, sizeof %s - 1);\n", domain, snapshot, cString(message), cString(message))
			to := e.program.Locals[f.Parameters[i+receiverParameter]].Type
			value, fresh := viewCallableConverted(from, to, slot, snapshot)
			fmt.Fprintf(&b, "adapted[%d].%s=%s;\n", i+offset, member(to), value)
			if fresh {
				releases = append(releases, fmt.Sprintf("adamic_release(adapted[%d].reference);\n", i+offset))
			}
		}
		invocation := ""
		if methodProducer {
			invocation = fmt.Sprintf("%s(receiver, adapted)", e.methodThunk(index))
			if e.program.PackedCountNeeded(index) {
				invocation = fmt.Sprintf("%s(receiver, adapted, count)", e.methodThunk(index))
			}
		} else {
			invocation = fmt.Sprintf("%s(self, adapted)", e.functionName(index))
			if e.program.PackedCountNeeded(index) {
				invocation = fmt.Sprintf("%s(self, adapted, count + %d)", e.functionName(index), offset)
			}
		}
		fmt.Fprintf(&b, "adamic_value result=%s;\n", invocation)
		for _, release := range releases {
			b.WriteString(release)
		}
		b.WriteString("if (adamic_thrown != NULL) return (adamic_value){.reference=NULL};\n")
		if call.Returns == 0 {
			if f.Returns.IsReference() {
				b.WriteString("adamic_release(result.reference);\n")
			}
			b.WriteString("(void)result; return (adamic_value){.reference=NULL};\n")
		} else {
			returnedSnapshot := e.viewCallableDomainSnapshot(f.Returns, "result", f.CallableResult)
			if f.CallableResultNull && f.Returns.IsReference() && f.Returns != ir.Union {
				returnedSnapshot = "(result.reference == NULL ? (adamic_view_union_value){adamic_view_union_null,{.reference=&adamic_null}} : " + returnedSnapshot + ")"
			}
			fmt.Fprintf(&b, "adamic_view_union_value returned=%s;\n", returnedSnapshot)
			message := fmt.Sprintf("callable call failed: %s at %s result expected view %s, producer %s", property.View, call.CallWhere, e.program.ViewContracts[target.Result-1].Name, f.CallableResultName)
			fmt.Fprintf(&b, "if (!%s(returned)) adamic_panic(%s, sizeof %s - 1);\n", resultCheck, cString(message), cString(message))
			value, _ := viewCallableConverted(f.Returns, call.Returns, "result", "returned")
			fmt.Fprintf(&b, "adamic_value converted_result={.%s=%s};\n", member(call.Returns), value)
			if f.Returns.IsReference() && !call.Returns.IsReference() {
				b.WriteString("adamic_release(result.reference);\n")
			}
			b.WriteString("return converted_result;\n")
		}
		b.WriteString("}\n")
	}
	message := "callable call failed: " + property.View + " at " + call.CallWhere + " has no checkable producer signature"
	fmt.Fprintf(&b, "adamic_panic(%s,sizeof %s - 1);\n}\n", cString(message), cString(message))
	e.declarations = append(e.declarations, b.String())
	return name
}

func (e *emitter) emitDirectViewCallable(call ir.CallClosure, receiver string) string {
	e.declarations = append(e.declarations, "#include \"view_callables.h\"")
	p := call.Closure.(ir.Property)
	method := e.temporary()
	if e.program.ClosureConventionNeeded() {
		e.line("adamic_method_entry %s={0};", method)
	} else {
		e.line("adamic_method %s=NULL;", method)
	}
	candidate := e.temporary()
	e.line("adamic_view_union_value %s=adamic_view_callable_candidate(%s,%s,&%s,&%s,%s,%s);", candidate, receiver, cString(p.Name), e.cache(), method, cString(p.View), cString(p.ViewType))
	closure := e.own(ir.Closure, fmt.Sprintf("adamic_retain(%s.kind == adamic_view_union_function ? %s.payload.reference : NULL)", candidate, candidate))
	packed, count := e.closureArguments(call)
	invoke := e.directViewCallableInvoke(call, p)
	result := e.temporary()
	e.line("adamic_value %s=%s(%s,%s,%s,%s,%s);", result, invoke, closure, receiver, method, packed, count)
	e.closureThrown()
	if call.Returns == 0 {
		e.line("(void)%s;", result)
		return "0"
	}
	if call.Returns.IsReference() {
		return e.own(call.Returns, fmt.Sprintf("(%s)%s.reference", cType(call.Returns), result))
	}
	return e.snapshot(call.Returns, unslotted(call.Returns, result+"."+member(call.Returns)))
}
