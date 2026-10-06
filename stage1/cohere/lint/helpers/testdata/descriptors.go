package main

import (
	"encoding"
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"reflect"
	"strconv"
	"strings"
)

var targets = map[string]reflect.Type{}
var descriptorBytes []byte

func shape(t reflect.Type, depth int) map[string]any {
	unsupported := map[string]any{"kind": "unsupported"}
	if depth > 20 || t == nil {
		return unsupported
	}
	if t.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) || reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) || t.Implements(reflect.TypeFor[json.Unmarshaler]()) || reflect.PointerTo(t).Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return unsupported
	}
	if t.Kind() == reflect.Pointer {
		return shape(t.Elem(), depth+1)
	}
	switch t.Kind() {
	case reflect.Bool:
		return map[string]any{"kind": "boolean"}
	case reflect.String:
		return map[string]any{"kind": "string"}
	case reflect.Float64:
		return map[string]any{"kind": "number"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bits := t.Bits()
		minimum := strconv.FormatInt(-1<<(bits-1), 10)
		maximum := strconv.FormatUint((uint64(1)<<(bits-1))-1, 10)
		return map[string]any{"kind": "number", "integer": true, "rawMinimum": minimum, "rawMaximum": maximum}

	case reflect.Slice:
		return map[string]any{"kind": "array", "item": shape(t.Elem(), depth+1)}
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return map[string]any{"kind": "object", "item": shape(t.Elem(), depth+1)}
		}
	case reflect.Struct:
		fields := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if f.Anonymous && name == "" {
				s := shape(f.Type, depth+1)
				if promoted, ok := s["fields"].(map[string]any); ok {
					for k, v := range promoted {
						fields[k] = v
					}
				}
				continue
			}
			if !f.IsExported() {
				continue
			}
			tagged := name != ""
			if name == "" {
				name = f.Name
			}
			fields[name] = map[string]any{"tagged": tagged, "shape": shape(f.Type, depth+1)}
		}
		return map[string]any{"kind": "object", "fields": fields}
	}
	return unsupported
}
func descriptors() []byte {
	if descriptorBytes != nil {
		return descriptorBytes
	}
	_ = registry.Count()
	result := map[string]any{}
	for _, registration := range rule.Registered() {
		for _, raw := range []string{"{}", "null", "[]", "\"always\"", "1"} {
			var value any
			if registration.Decode != nil {
				value, _ = registration.Decode([]byte(raw))
			} else if registration.DecodeOptionList != nil {
				value, _ = registration.DecodeOptionList([]byte(raw))
			} else if registration.DecodeAt != nil {
				value, _ = registration.DecodeAt([]byte(raw), rule.OptionsBase{})
			}
			if value != nil {
				target := reflect.TypeOf(value)
				targets[registration.Rule.Name] = target
				result[registration.Rule.Name] = shape(target, 0)
				break
			}
		}
	}
	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	descriptorBytes = data
	return data
}
func decode(c Case) string {
	descriptors()
	target := targets[c.Rule]
	if target == nil {
		panic("missing option target " + c.Rule)
	}
	if rule.UnmarshalOptions([]byte(c.Input), reflect.New(target).Interface()) == nil {
		return "valid"
	}
	return "invalid"
}
