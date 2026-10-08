package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"sort"
	"strings"
)

func (e *emitter) reflectionDeclarations(builder *strings.Builder) {
	names := []string{}
	for _, name := range e.shapes {
		names = append(names, name)
	}
	sort.Strings(names)
	builder.WriteString("static const adamic_reflection_layout *adamic_reflection_shape_types(const adamic_shape *shape) {\n")
	for _, name := range names {
		// Empty shapes have no values to classify.
		builder.WriteString(fmt.Sprintf(" if (shape == &%s) {\n", name))
		found := false
		for _, declaration := range e.declarations {
			if strings.Contains(declaration, "int "+name+"_reflection_types[]") {
				found = true
				break
			}
		}
		if found {
			fmt.Fprintf(builder, "  return &%s_reflection_layout;\n", name)
		} else {
			builder.WriteString("  return NULL;\n")
		}
		builder.WriteString(" }\n")
	}
	builder.WriteString(" return NULL;\n}\n")
}

func (e *emitter) reflectionMembers(members []ir.ReflectionMember) string {
	if len(members) == 0 {
		return "NULL"
	}
	values := []string{}
	for _, member := range members {
		values = append(values, fmt.Sprintf("{%d, %t, %s, %d, %s, %t}", member.Kind, member.Literal, cString(member.Text), len(member.Text), cNumber(member.Number), member.Boolean))
	}
	name := e.temporary() + "_members"
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_reflection_member %s[] = {%s};", name, strings.Join(values, ", ")))
	return name
}

func (e *emitter) checkedObjectCall(call ir.ObjectCall, arguments []string) string {
	proof := call.Reflection
	if call.Method == "entries" {
		return e.own(ir.Array, fmt.Sprintf("adamic_checked_entries(%s, %d, %d, %s, adamic_reflection_shape_types, %t, %s)", arguments[0], call.Element, len(proof.Members), e.reflectionMembers(proof.Members), proof.OwnPropertyOrder, cString(proof.Message)))
	}
	for _, source := range arguments[1:] {
		e.line("adamic_reflection_assign(%s, %s, adamic_reflection_shape_types, %t);", arguments[0], source, proof.OwnPropertyOrder)
	}
	return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
}
