package native

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// This named adapter is a shared-dispatch handoff. It uses already evaluated
// objects and shared slot metadata; field reads retain their original guards.
func (e *emitter) viewUntaggedObjectUnion(property ir.Property, object string) {
	if property.ViewType == "" && property.ViewContract > 0 {
		property.ViewType = e.program.ViewContracts[property.ViewContract-1].Name
	}

	e.declarations = append(e.declarations, "#include \"view_callables_contract.h\"")
	name := e.temporary()
	var declarations strings.Builder
	declarations.WriteString("#include \"view_unions_untagged.h\"\n")
	producers := name + "_producers"
	producerRows := []string{}
	producerRegistry := e.unionCallableProducers()
	for _, producer := range producerRegistry {
		producerRows = append(producerRows, e.viewCallableProducerRow(producer.Function, producer.Signature))
	}
	if len(producerRows) == 0 {
		producers = "NULL"
	} else {
		fmt.Fprintf(&declarations, "static const adamic_view_untagged_callable_producer %s[] = {%s};\n", producers, strings.Join(producerRows, ","))
	}
	rows := []string{}
	for i, contract := range e.program.ViewContracts {
		prefix := fmt.Sprintf("%s_%d", name, i)
		allowed, fields, members := "NULL", "NULL", "NULL"
		if len(contract.Allowed) != 0 {
			literals := []string{}
			for _, literal := range contract.Allowed {
				kind, payload := "adamic_view_union_number", ".number = "+strconv.FormatFloat(literal.Number, 'g', -1, 64)
				if literal.Of == ir.Boolean {
					kind = "adamic_view_union_boolean"
					payload = fmt.Sprintf(".boolean = %t", literal.Boolean)
				}
				if literal.Of == ir.String {
					kind = "adamic_view_union_string"
					text := prefix + "_text_" + strconv.Itoa(len(literals))
					fmt.Fprintf(&declarations, "static adamic_string %s = ADAMIC_STRING(%s);\n", text, cString(literal.String))
					payload = ".reference = &" + text
				}
				literals = append(literals, fmt.Sprintf("{.kind=%s,.literal=true,.value={%s}}", kind, payload))
			}
			allowed = prefix + "_allowed"
			fmt.Fprintf(&declarations, "static const adamic_view_union_member %s[] = {%s};\n", allowed, strings.Join(literals, ","))
		}
		if len(contract.Fields) != 0 {
			values := []string{}
			for _, field := range contract.Fields {
				values = append(values, fmt.Sprintf("{%s,%d,%t}", cString(field.Name), field.Contract, field.Optional))
			}
			fields = prefix + "_fields"
			fmt.Fprintf(&declarations, "static const adamic_view_untagged_field %s[] = {%s};\n", fields, strings.Join(values, ","))
		}
		if len(contract.Members) != 0 {
			values := []string{}
			for _, id := range contract.Members {
				values = append(values, strconv.Itoa(int(id)))
			}
			members = prefix + "_members"
			fmt.Fprintf(&declarations, "static const size_t %s[] = {%s};\n", members, strings.Join(values, ","))
		}
		kind := contract.Kind
		if contract.Nominal != "" || contract.Unsupported != "" && contract.Unsupported != "untagged object union" {
			kind = ir.ViewUnknown
		}
		callable := "NULL"
		if kind == ir.ViewCallable {
			callable = e.unionCallableExpected(ir.Property{ViewContract: ir.ViewContractID(i + 1)})
		}
		contractProducers, contractProducerCount := producers, len(producerRows)
		if contract.ProducerCertified {
			allowedProducers := []string{}
			for _, producer := range producerRegistry {
				for _, function := range contract.Functions {
					if producer.Function == function {
						allowedProducers = append(allowedProducers, e.viewCallableProducerRow(function, producer.Signature))
						break
					}
				}
			}
			contractProducers, contractProducerCount = "NULL", len(allowedProducers)
			if len(allowedProducers) > 0 {
				contractProducers = prefix + "_producers"
				fmt.Fprintf(&declarations, "static const adamic_view_untagged_callable_producer %s[] = {%s};\n", contractProducers, strings.Join(allowedProducers, ","))
			}
		}
		rows = append(rows, fmt.Sprintf("{%d,%d,%t,%s,%d,%s,%d,%s,%d,%d,%s,%s,%d,%t,%d}", kind, contract.Of, contract.Undefined, allowed, len(contract.Allowed), fields, len(contract.Fields), members, len(contract.Members), contract.Element, callable, contractProducers, contractProducerCount, contract.FixedTuple, len(contract.Tuple)))
	}
	fmt.Fprintf(&declarations, "static const adamic_view_untagged_contract %s[] = {%s};\n", name, strings.Join(rows, ","))
	e.declarations = append(e.declarations, declarations.String())
	e.line("(void)adamic_view_untagged_plain_select((const adamic_object *)%s, %s, %d, %d, %s, %s);", object, name, len(rows), property.ViewContract, cString(property.View), cString(property.ViewType))
}

// Preserve the existing shared-discriminant dispatch.
func untaggedObjectUnion(contracts []ir.ViewContract, contract ir.ViewContract) bool {
	return len(contract.Members) != 0 && !ir.ViewUnionHasDiscriminant(contracts, contract)
}

func (e *emitter) untaggedCallableUnionExpected(property ir.Property, recorded, fallback string) string {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return fallback
	}
	root := e.program.ViewContracts[id-1]
	if root.Kind != ir.ViewUnion || root.Of != ir.Closure || len(root.Members) == 0 {
		return fallback
	}
	choices := []string{}
	first := "NULL"
	for _, child := range root.Members {
		contract := e.program.ViewContracts[child-1]
		if contract.Kind != ir.ViewCallable || contract.Result == 0 {
			continue
		}
		parameters := make([]ir.Type, len(contract.Parameters))
		result := e.program.ViewContracts[contract.Result-1].Of
		if result == 0 {
			continue
		}
		tests := []string{recorded + " != NULL", fmt.Sprintf("%s->arity == %d", recorded, len(parameters)), fmt.Sprintf("%s->result == %d", recorded, result)}
		known := true
		for i, parameter := range contract.Parameters {
			parameters[i] = e.program.ViewContracts[parameter-1].Of
			known = known && parameters[i] != 0
			tests = append(tests, fmt.Sprintf("%s->parameters != NULL && %s->parameters[%d] == %d", recorded, recorded, i, parameters[i]))
		}
		if !known {
			continue
		}
		expected := e.unionCallableSignature(parameters, result, root.Name)
		if first == "NULL" {
			first = expected
		}
		choices = append(choices, "("+strings.Join(tests, " && ")+") ? "+expected+" : ")
	}
	return "(" + strings.Join(choices, "") + first + ")"
}

func (e *emitter) certifyUntaggedCallableRecorded(property ir.Property, value, recorded, expected string) {
	id := property.ViewContract
	if id <= 0 || int(id) > len(e.program.ViewContracts) {
		return
	}
	root := e.program.ViewContracts[id-1]
	if root.Kind != ir.ViewUnion || root.Of != ir.Closure {
		return
	}
	tests := []string{}
	certified := false
	for _, member := range root.Members {
		contract := e.program.ViewContracts[member-1]
		certified = certified || contract.ProducerCertified
		for _, function := range contract.Functions {
			// Match the producer registry before comparing code identities.
			if !e.program.Functions[function].Closure || e.program.Functions[function].Receiver {
				continue
			}
			tests = append(tests, e.unionClosureCodeIdentity(value, function))
		}
	}
	if !certified {
		return
	}
	allowed := "false"
	if len(tests) > 0 {
		allowed = "(" + strings.Join(tests, " || ") + ")"
	}
	e.line("if (%s != NULL && adamic_view_callable_signatures_match(%s, %s) && !%s) %s = NULL;", recorded, recorded, expected, allowed, recorded)
}

func (e *emitter) viewCallableProducerRow(index int, signature string) string {
	if e.program.ClosureConventionNeeded() && e.program.PackedCountNeeded(index) {
		return fmt.Sprintf("{.counted = true, .counted_code = %s, .signature = %s}", e.functionName(index), signature)
	}
	return fmt.Sprintf("{.code = %s, .signature = %s}", e.functionName(index), signature)
}
