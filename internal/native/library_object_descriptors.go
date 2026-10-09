package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) objectDescriptorCall(call ir.ObjectCall) (string, bool) {
	switch call.Method {
	case "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "defineProperty", "defineProperties", "propertyIsEnumerable", "descriptorFlag", "descriptorValue", "descriptorAbsent":
	default:
		return "", false
	}
	e.declarations = append(e.declarations, "#include \"library_object_descriptors.h\"")
	args := make([]string, len(call.Arguments))
	for i, a := range call.Arguments {
		args[i] = e.value(a)
	}
	if call.Method == "propertyIsEnumerable" {
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_property_enumerable(%s, %s)", args[0], args[1])), true
	}
	if call.Method == "descriptorFlag" {
		return e.snapshot(ir.MaybeBoolean, fmt.Sprintf("adamic_descriptor_flag(%s, %s)", args[0], args[1])), true
	}
	if call.Method == "descriptorAbsent" {
		return e.own(ir.Closure, "NULL"), true
	}
	if call.Method == "descriptorValue" {
		if call.Returns == ir.MaybeNumber || call.Returns == ir.Number {
			v := e.snapshot(ir.MaybeNumber, fmt.Sprintf("adamic_descriptor_number(%s)", args[0]))
			if call.Returns == ir.Number {
				v = e.snapshot(ir.Number, v+".number")
			}
			return v, true
		}
		if call.Returns == ir.MaybeBoolean || call.Returns == ir.Boolean {
			v := e.snapshot(ir.MaybeBoolean, fmt.Sprintf("adamic_descriptor_boolean(%s)", args[0]))
			if call.Returns == ir.Boolean {
				v = e.snapshot(ir.Boolean, v+".boolean")
			}
			return v, true
		}
		return e.own(call.Returns, fmt.Sprintf("(%s)adamic_descriptor_value(%s)", cType(call.Returns), args[0])), true
	}
	e.temporaries++
	table := fmt.Sprintf("adamic_descriptor_fields_%d", e.temporaries)
	fields := []string{}
	names := []string{}
	types := []ir.Type{}
	for _, f := range call.DescriptorFields {
		fields = append(fields, fmt.Sprintf("{%s, %d}", cString(f.Name), f.Of))
		names = append(names, f.Name)
		types = append(types, ir.Object)
	}
	metadata := "NULL"
	if len(fields) > 0 {
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_descriptor_field %s[] = {%s};", table, strings.Join(fields, ", ")))
		metadata = table
	}
	suffix := fmt.Sprintf("%s, %d", metadata, len(fields))
	code := ""
	switch call.Method {
	case "getOwnPropertyDescriptor":
		code = fmt.Sprintf("adamic_object_descriptor(%s, %s, %s)", args[0], args[1], suffix)
	case "getOwnPropertyDescriptors":
		code = fmt.Sprintf("adamic_object_descriptors(%s, &%s, %s)", args[0], e.shapeOf(names, types), suffix)
	case "defineProperty":
		code = fmt.Sprintf("adamic_object_define_property(%s, %s, %s, %s)", args[0], args[1], args[2], suffix)
	case "defineProperties":
		code = fmt.Sprintf("adamic_object_define_properties(%s, %s, %s)", args[0], args[1], suffix)
	}
	result := e.own(ir.Object, code)
	if call.Method == "defineProperty" || call.Method == "defineProperties" {
		e.checkThrown()
	}
	return result, true
}
