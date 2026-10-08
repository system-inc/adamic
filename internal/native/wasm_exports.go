package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// ABIType is the version 1 schema. Fields remain in declaration order.
type ABIType struct {
	Kind   string     `json:"kind"`
	Fields []ABIField `json:"fields,omitempty"`
}
type ABIField struct {
	Name string  `json:"name"`
	Type ABIType `json:"type"`
}
type ABIExport struct {
	Name       string     `json:"name"`
	Parameters []ABIField `json:"parameters"`
	Returns    ABIType    `json:"returns"`
	Function   int        `json:"-"`
}
type ABITable struct {
	Version int         `json:"version"`
	Exports []ABIExport `json:"exports"`
}

func (t ABIType) representation() ir.Type {
	switch t.Kind {
	case "number":
		return ir.Number
	case "boolean":
		return ir.Boolean
	case "string":
		return ir.String
	case "number[]", "string[]":
		return ir.Array
	case "record":
		return ir.Object
	}
	return 0
}

// WASIExports appends typed boundaries to the normal emitter, retaining module globals.
func WASIExports(program *ir.Program, exports []ABIExport) (string, error) {
	for _, export := range exports {
		if export.Function < 0 || export.Function >= len(program.Functions) {
			return "", fmt.Errorf("invalid ABI function %s", export.Name)
		}
		f := program.Functions[export.Function]
		if f.Closure || f.Returns != export.Returns.representation() || len(f.Parameters) != len(export.Parameters) {
			return "", fmt.Errorf("ABI representation mismatch for %s", export.Name)
		}
		for i, p := range f.Parameters {
			if program.Locals[p].Type != export.Parameters[i].Type.representation() {
				return "", fmt.Errorf("ABI parameter mismatch for %s", export.Name)
			}
		}
	}
	// Include host-created layouts in the emitter's uniform-field-offset proof.
	// The witnesses are unreachable and allocate nothing at module initialization.
	copyProgram := *program
	copyProgram.Functions = append([]ir.Function(nil), program.Functions...)
	var witness func(ABIType) ir.Expression
	witness = func(t ABIType) ir.Expression {
		switch t.Kind {
		case "number":
			return ir.NumberConstant{}
		case "boolean":
			return ir.BooleanConstant{}
		case "record":
			fields := []ir.Field{}
			for _, f := range t.Fields {
				fields = append(fields, ir.Field{Name: f.Name, Value: witness(f.Type)})
			}
			literal := ir.ObjectLiteral{Fields: fields}
			copyProgram.Functions = append(copyProgram.Functions, ir.Function{Name: "abi_layout", Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: literal}}})
			return literal
		default:
			return ir.Undefined{Of: t.representation()}
		}
	}
	for _, export := range exports {
		for _, p := range export.Parameters {
			witness(p.Type)
		}
		witness(export.Returns)
	}
	program = &copyProgram
	source := cProgram(program, len(program.Functions))
	_, lending := planElementBorrows(program)
	e := &emitter{program: program, reuse: planReuse(program, lending)}
	var declarations, bodies strings.Builder
	bodies.WriteString(abiSupport)
	serial := 0
	var codec func(ABIType) string
	codec = func(t ABIType) string {
		serial++
		name := fmt.Sprintf("adamic_abi_type_%d", serial)
		children := []string{}
		for _, field := range t.Fields {
			children = append(children, codec(field.Type))
		}
		ct := cType(t.representation())
		fmt.Fprintf(&declarations, "static %s %s_read(adamic_abi_buffer *b) {\n", ct, name)
		switch t.Kind {
		case "number":
			declarations.WriteString(" return adamic_abi_number_read(b);\n")
		case "boolean":
			declarations.WriteString(" return adamic_abi_boolean_read(b);\n")
		case "string":
			declarations.WriteString(" return adamic_abi_string_read(b);\n")
		case "number[]", "string[]":
			refs := t.Kind == "string[]"
			fmt.Fprintf(&declarations, " uint32_t count = adamic_abi_u32_read(b);\n adamic_array *v = adamic_array_new(count, %t);\n", refs)
			if !refs {
				declarations.WriteString(" adamic_abi_align(b);\n")
			}
			read, member := "adamic_abi_number_read(b)", "number"
			if refs {
				read, member = "adamic_abi_string_read(b)", "reference"
			}
			fmt.Fprintf(&declarations, " for (uint32_t i=0; i<count; i++) adamic_array_push(v, (adamic_value){.%s = %s});\n return v;\n", member, read)
		case "record":
			fmt.Fprintf(&declarations, " static const char *const names[%d] = {", max(1, len(t.Fields)))
			for i, f := range t.Fields {
				if i > 0 {
					declarations.WriteString(",")
				}
				declarations.WriteString(cString(f.Name))
			}
			if len(t.Fields) == 0 {
				declarations.WriteString("NULL")
			}
			declarations.WriteString("};\n")
			fmt.Fprintf(&declarations, " static const bool refs[%d] = {", max(1, len(t.Fields)))
			for i, f := range t.Fields {
				if i > 0 {
					declarations.WriteString(",")
				}
				fmt.Fprintf(&declarations, "%t", f.Type.representation().IsReference())
			}
			if len(t.Fields) == 0 {
				declarations.WriteString("false")
			}
			declarations.WriteString("};\n")
			fmt.Fprintf(&declarations, " static const adamic_shape shape = {%d,names,refs,NULL};\n adamic_object *v = adamic_object_new(&shape);\n", len(t.Fields))
			for i, f := range t.Fields {
				fmt.Fprintf(&declarations, " v->slots[%d].%s = %s_read(b);\n", i, member(f.Type.representation()), children[i])
			}
			declarations.WriteString(" return v;\n")
		}
		declarations.WriteString("}\n")
		fmt.Fprintf(&declarations, "static void %s_write(adamic_abi_buffer *b, %s v) {\n", name, ct)
		switch t.Kind {
		case "number":
			declarations.WriteString(" adamic_abi_number_write(b,v);\n")
		case "boolean":
			declarations.WriteString(" adamic_abi_boolean_write(b,v);\n")
		case "string":
			declarations.WriteString(" adamic_abi_string_write(b,v);\n")
		case "number[]", "string[]":
			declarations.WriteString(" adamic_abi_u32_write(b,(uint32_t)v->length);\n")
			fn, mem := "adamic_abi_number_write", "number"
			if t.Kind == "string[]" {
				fn, mem = "adamic_abi_string_write", "reference"
			} else {
				declarations.WriteString(" adamic_abi_align(b);\n")
			}
			cast := ""
			if t.Kind == "string[]" {
				cast = "(adamic_string *)"
			}
			fmt.Fprintf(&declarations, " for (size_t i=0; i<v->length; i++) %s(b,%sv->elements[i].%s);\n", fn, cast, mem)
		case "record":
			for i, f := range t.Fields {
				fmt.Fprintf(&declarations, " static adamic_slot_cache cache_%d;\n", i)
				cast := ""
				if f.Type.representation().IsReference() {
					cast = "(" + cType(f.Type.representation()) + ")"
				}
				fmt.Fprintf(&declarations, " %s_write(b, %sadamic_object_field(v,%s,&cache_%d)->%s);\n", children[i], cast, cString(f.Name), i, member(f.Type.representation()))
			}
		}
		declarations.WriteString("}\n")
		return name
	}
	var wrappers strings.Builder
	for exportIndex, export := range exports {
		f := program.Functions[export.Function]
		args, params := []string{}, []string{}
		var body strings.Builder
		for i, p := range export.Parameters {
			name := fmt.Sprintf("p%d", i)
			t := p.Type
			switch t.Kind {
			case "number":
				params = append(params, "double "+name)
				args = append(args, name)
			case "boolean":
				params = append(params, "int32_t "+name)
				args = append(args, "("+name+" != 0)")
			default:
				params = append(params, "const unsigned char *"+name, "uint32_t "+name+"_length")
				value := name + "_value"
				args = append(args, value)
				if t.Kind == "string" {
					fmt.Fprintf(&body, " adamic_string *%s = adamic_abi_wtf8(%s,%s_length);\n", value, name, name)
				} else if t.Kind == "number[]" {
					fmt.Fprintf(&body, " adamic_array *%s = adamic_array_new(%s_length,false);\n for (uint32_t i=0; i<%s_length; i++) { double v; memcpy(&v,%s+8*i,8); adamic_array_push(%s,(adamic_value){.number=v}); }\n", value, name, name, name, value)
				} else {
					coder := codec(t)
					fmt.Fprintf(&body, " adamic_abi_buffer %s_buffer = {%s,%s_length,0,false};\n %s %s = %s_read(&%s_buffer);\n if (%s_buffer.offset != %s_length) adamic_panic_c(\"ABI trailing input bytes\");\n", name, name, name, cType(t.representation()), value, coder, name, name, name)
				}
				if e.reuse.consumed[f.Parameters[i]] {
					args[len(args)-1] = "adamic_retain(" + value + ")"
				}
			}
		}
		ret := "uint32_t"
		if export.Returns.Kind == "void" {
			ret = "void"
		} else if export.Returns.Kind == "number" {
			ret = "double"
		} else if export.Returns.Kind == "boolean" {
			ret = "int32_t"
		}
		if len(params) == 0 {
			params = append(params, "void")
		}
		fmt.Fprintf(&wrappers, "__attribute__((export_name(%s))) %s adamic_abi_export_%d(%s) {\n", cString("adamic_export_"+export.Name), ret, exportIndex, strings.Join(params, ", "))
		wrappers.WriteString(body.String())
		call := fmt.Sprintf("%s(%s)", e.functionName(export.Function), strings.Join(args, ", "))
		if export.Returns.Kind == "void" {
			fmt.Fprintf(&wrappers, " %s;\n", call)
		} else {
			fmt.Fprintf(&wrappers, " %s result = %s;\n", cType(f.Returns), call)
		}
		for i, p := range export.Parameters {
			if p.Type.representation().IsReference() {
				fmt.Fprintf(&wrappers, " adamic_release(p%d_value);\n", i)
			}
		}
		if f.MayThrow {
			wrappers.WriteString(" if (adamic_thrown != NULL) adamic_uncaught();\n")
		}
		wrappers.WriteString(" adamic_output_flush();\n")
		switch export.Returns.Kind {
		case "void":
		case "number", "boolean":
			wrappers.WriteString(" return result;\n")
		default:
			wrappers.WriteString(" adamic_abi_buffer encoded = {NULL,0,0,true};\n")
			if export.Returns.Kind == "string" {
				wrappers.WriteString(" adamic_abi_write(&encoded,result->bytes,result->length);\n")
			} else {
				fmt.Fprintf(&wrappers, " %s_write(&encoded,result);\n", codec(export.Returns))
			}
			wrappers.WriteString(" adamic_release(result);\n return adamic_abi_result(&encoded);\n")
		}
		wrappers.WriteString("}\n")
	}
	// Preserve the earlier request host API for the ordinary string handler.
	for _, export := range exports {
		if export.Name == "handleRequest" && export.Returns.Kind == "string" && len(export.Parameters) == 1 && export.Parameters[0].Type.Kind == "string" {
			legacy := e.requestABI(export.Function)
			start := strings.Index(legacy, "// Borrowed UTF-8 input")
			end := strings.Index(legacy, "#ifdef ADAMIC_COUNT")
			wrappers.WriteString(legacy[start:end])
		}
	}
	return source + bodies.String() + declarations.String() + wrappers.String(), nil
}

const abiSupport = `
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
static void adamic_panic_c(const char *message) { adamic_panic(message,strlen(message)); }
typedef struct { const unsigned char *bytes; size_t length, offset; bool writing; } adamic_abi_buffer;
static void adamic_abi_need(adamic_abi_buffer *b, size_t n) {
 if (n > SIZE_MAX-b->offset) adamic_panic_c("ABI size overflow");
 size_t end = b->offset+n;
 if (end <= b->length) return;
 if (!b->writing) adamic_panic_c("ABI truncated input");
 size_t capacity = b->length > SIZE_MAX/2 ? end : b->length*2;
 if (capacity < end) capacity=end;
 void *next = realloc((void *)b->bytes,capacity);
 if (next == NULL) adamic_panic_c("ABI out of memory");
 b->bytes=next; b->length=capacity;
}
static void adamic_abi_write(adamic_abi_buffer *b,const void *bytes,size_t n) {
 adamic_abi_need(b,n); if (n != 0) memcpy((unsigned char *)b->bytes+b->offset,bytes,n); b->offset+=n;
}
static void adamic_abi_align(adamic_abi_buffer *b) {
 size_t pad=(8-(b->offset%8))%8; adamic_abi_need(b,pad);
 if (b->writing && pad != 0) memset((unsigned char *)b->bytes+b->offset,0,pad);
 b->offset+=pad;
}
static uint32_t adamic_abi_u32_read(adamic_abi_buffer *b) {
 adamic_abi_need(b,4); uint32_t v; memcpy(&v,b->bytes+b->offset,4); b->offset+=4; return v;
}
static void adamic_abi_u32_write(adamic_abi_buffer *b,uint32_t v) { adamic_abi_write(b,&v,4); }
static double adamic_abi_number_read(adamic_abi_buffer *b) {
 adamic_abi_align(b); adamic_abi_need(b,8); double v; memcpy(&v,b->bytes+b->offset,8); b->offset+=8; return v;
}
static void adamic_abi_number_write(adamic_abi_buffer *b,double v) { adamic_abi_align(b); adamic_abi_write(b,&v,8); }
static bool adamic_abi_boolean_read(adamic_abi_buffer *b) {
 adamic_abi_need(b,1); unsigned char v=b->bytes[b->offset++]; if (v>1) adamic_panic_c("ABI invalid boolean"); return v != 0;
}
static void adamic_abi_boolean_write(adamic_abi_buffer *b,bool v) { unsigned char byte=v ? 1 : 0; adamic_abi_write(b,&byte,1); }
// Accept canonical UTF-8 and WTF-8 surrogate code points. Never store malformed bytes.
static adamic_string *adamic_abi_wtf8(const unsigned char *bytes,size_t length) {
 for (size_t i=0; i<length;) {
  unsigned char lead=bytes[i++]; if (lead<128) continue;
  unsigned point; size_t more;
  if (lead>=0xc2 && lead<=0xdf) { point=lead&31; more=1; }
  else if (lead>=0xe0 && lead<=0xef) { point=lead&15; more=2; }
  else if (lead>=0xf0 && lead<=0xf4) { point=lead&7; more=3; }
  else { adamic_panic_c("ABI invalid UTF-8"); return NULL; }
  if (more>length-i) adamic_panic_c("ABI invalid UTF-8");
  for (size_t j=0; j<more; j++) { unsigned char c=bytes[i++]; if ((c&0xc0)!=0x80) adamic_panic_c("ABI invalid UTF-8"); point=(point<<6)|(c&63); }
  if ((more==1 && point<128)||(more==2 && point<2048)||(more==3 && point<65536)||point>0x10ffff) adamic_panic_c("ABI invalid UTF-8");
 }
 adamic_string *v=adamic_string_allocate(length); if (length != 0) memcpy((char *)v->bytes,bytes,length); return v;
}
static adamic_string *adamic_abi_string_read(adamic_abi_buffer *b) {
 uint32_t n=adamic_abi_u32_read(b); adamic_abi_need(b,n); adamic_string *v=adamic_abi_wtf8(b->bytes+b->offset,n); b->offset+=n; return v;
}
static void adamic_abi_string_write(adamic_abi_buffer *b,adamic_string *v) {
 if (v->length>UINT32_MAX) adamic_panic_c("ABI string too large");
 adamic_abi_u32_write(b,(uint32_t)v->length); adamic_abi_write(b,v->bytes,v->length);
}
typedef struct { unsigned char *bytes; uint32_t length; } adamic_abi_result_buffer;
__attribute__((export_name("adamic_alloc"))) void *adamic_alloc(uint32_t size) {
 void *p=malloc(size == 0 ? 1 : size); if (p == NULL) adamic_panic_c("ABI out of memory"); return p;
}
__attribute__((export_name("adamic_free"))) void adamic_free(void *p) { free(p); }
static uint32_t adamic_abi_result(adamic_abi_buffer *b) {
 if (b->offset>UINT32_MAX) adamic_panic_c("ABI result too large");
 adamic_abi_result_buffer *r=adamic_alloc(sizeof *r); r->bytes=(unsigned char *)b->bytes; r->length=(uint32_t)b->offset; return (uint32_t)(uintptr_t)r;
}
__attribute__((export_name("adamic_result_bytes"))) const unsigned char *adamic_result_bytes(adamic_abi_result_buffer *r) { return r->bytes; }
__attribute__((export_name("adamic_result_length"))) uint32_t adamic_result_length(adamic_abi_result_buffer *r) { return r->length; }
__attribute__((export_name("adamic_result_release"))) void adamic_result_release(adamic_abi_result_buffer *r) { free(r->bytes); free(r); }
`
