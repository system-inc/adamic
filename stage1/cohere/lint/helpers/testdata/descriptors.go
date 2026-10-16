package main

import (
	"encoding"
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"reflect"
	"strconv"
	"strings"
)

var targets = map[string]reflect.Type{}
var descriptorBytes []byte

// custom states, by hand, what a type with its own decoder accepts, read from that decoder at the pin.
// A self-decoding type not listed here stays unsupported. The corpus checks each claim against Go.
var custom = map[reflect.Type]map[string]any{
	// LegacySelector.UnmarshalJSON: a name, or exactly a name and its matchers. The matchers go through
	// rule.UnmarshalOptions into legacyMatcher, then each match must be one of three words, so a null
	// matcher or one without its match is refused.
	reflect.TypeFor[tailwind.LegacySelector](): {"kind": "anyOf", "options": []any{
		map[string]any{"kind": "string"},
		map[string]any{"kind": "tuple", "items": []any{
			map[string]any{"kind": "string"},
			map[string]any{"kind": "array", "item": map[string]any{"kind": "object", "nonNull": true, "fields": map[string]any{
				"match":       map[string]any{"tagged": true, "required": true, "shape": map[string]any{"kind": "string", "nonNull": true, "enum": []any{"strings", "objectKeys", "objectValues"}}},
				"pathPattern": map[string]any{"tagged": true, "shape": map[string]any{"kind": "string"}},
			}}},
		}},
	}},
	// SelectorTarget.UnmarshalJSON: one of three words, or a whole number no larger than 1<<31 either way.
	reflect.TypeFor[tailwind.SelectorTarget](): {"kind": "anyOf", "options": []any{
		map[string]any{"kind": "string", "enum": []any{"all", "first", "last"}},
		map[string]any{"kind": "number", "whole": true, "magnitude": "2147483648"},
	}},
}

// describer walks one rule's target. A struct reached again inside itself is a ref to its definition,
// so a recursive type is described whole instead of cut off at the depth limit.
type describer struct {
	open        map[reflect.Type]bool
	recursive   map[reflect.Type]bool
	definitions map[string]any
}

func (d *describer) shape(t reflect.Type, depth int) map[string]any {
	unsupported := map[string]any{"kind": "unsupported"}
	if depth > 20 || t == nil {
		return unsupported
	}
	// Dereferencing first changes no verdict below (a pointer's method set holds its element's), and
	// lets a custom type behind a pointer be found.
	if t.Kind() == reflect.Pointer {
		return d.shape(t.Elem(), depth+1)
	}
	if described, ok := custom[t]; ok {
		return described
	}
	if t.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) || reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) || t.Implements(reflect.TypeFor[json.Unmarshaler]()) || reflect.PointerTo(t).Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return unsupported
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
		return map[string]any{"kind": "array", "item": d.shape(t.Elem(), depth+1)}
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return map[string]any{"kind": "object", "item": d.shape(t.Elem(), depth+1)}
		}
	case reflect.Struct:
		if d.open[t] {
			d.recursive[t] = true
			return map[string]any{"kind": "ref", "name": t.String()}
		}
		d.open[t] = true
		defer delete(d.open, t)
		fields := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if f.Anonymous && name == "" {
				s := d.shape(f.Type, depth+1)
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
			fields[name] = map[string]any{"tagged": tagged, "shape": d.shape(f.Type, depth+1)}
		}
		described := map[string]any{"kind": "object", "fields": fields}
		if d.recursive[t] {
			d.definitions[t.String()] = described
		}
		return described
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
				d := describer{open: map[reflect.Type]bool{}, recursive: map[reflect.Type]bool{}, definitions: map[string]any{}}
				described := d.shape(target, 0)
				if len(d.definitions) > 0 {
					// A copy, so a target that is itself a definition does not come to contain itself.
					root := map[string]any{"definitions": d.definitions}
					for key, value := range described {
						root[key] = value
					}
					described = root
				}
				result[registration.Rule.Name] = described
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
