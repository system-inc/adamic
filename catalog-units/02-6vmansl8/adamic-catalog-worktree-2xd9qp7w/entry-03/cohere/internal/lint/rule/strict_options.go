package rule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// UnmarshalOptions decodes a rule's option JSON into a struct, refusing any key the struct does not
// declare and any key spelled in a different case from its declaration.
//
// Go's `json.Unmarshal` does neither. It drops an unknown key without a word and matches keys
// case-insensitively, so `enforceForTsTypes` silently set `enforceForTSTypes` and `ignorePatern`
// silently set nothing. Upstream's schemas refuse both, and a config that loads clean while an
// option does nothing is the defect this whole options layer exists to end (#4n972g9, after
// #27dkbbw closed the same hole one level up, for whole option elements).
//
// `DisallowUnknownFields` covers unknown keys and does not cover case: a key that matches a field
// case-insensitively is not unknown to it. So the keys are also walked against the declared names,
// exactly, after decoding. A type with its own `UnmarshalJSON` is not descended into, because it
// decides its own shape; it is responsible for its own strictness.
//
// Case is enforced only for a field with a `json` tag, because only a tag declares the spelling. An
// untagged field reaches its config key through the case-insensitive match itself (`Ignore` reads
// `ignore`), and deriving upstream's spelling from a Go name would be a guess at every acronym
// (`URLPatterns` is `urlPatterns`, not `uRLPatterns`). Measured 2026-10-02: ahra's and phi's own
// configs reach untagged fields this way, so enforcing case on them would refuse working configs. An
// unknown key is refused either way.
func UnmarshalOptions(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return err
	}
	return checkKeyCase(generic, reflect.TypeOf(target), "")
}

var jsonUnmarshalerType = reflect.TypeFor[json.Unmarshaler]()

// checkKeyCase walks decoded JSON beside the Go type it was decoded into and refuses an object key
// that only matched its field case-insensitively.
func checkKeyCase(value any, valueType reflect.Type, path string) error {
	if valueType == nil {
		return nil
	}
	for valueType.Kind() == reflect.Pointer {
		if valueType.Implements(jsonUnmarshalerType) {
			return nil
		}
		valueType = valueType.Elem()
	}
	if valueType.Implements(jsonUnmarshalerType) || reflect.PointerTo(valueType).Implements(jsonUnmarshalerType) {
		return nil
	}

	switch valueType.Kind() {
	case reflect.Struct:
		object, isObject := value.(map[string]any)
		if !isObject {
			return nil
		}
		fields := declaredFields(valueType)
		for key, member := range object {
			field, exact := fields[key]
			if !exact {
				for declared, candidate := range fields {
					if !strings.EqualFold(declared, key) {
						continue
					}
					if candidate.tagged {
						return fmt.Errorf("option %q is spelled %q: keys are matched exactly, as upstream's schema matches them",
							path+key, path+declared)
					}
					field, exact = candidate, true
				}
				if !exact {
					continue
				}
			}
			if err := checkKeyCase(member, field.Type, path+key+"."); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		list, isList := value.([]any)
		if !isList {
			return nil
		}
		for index, element := range list {
			if err := checkKeyCase(element, valueType.Elem(), fmt.Sprintf("%s%d.", path, index)); err != nil {
				return err
			}
		}
	case reflect.Map:
		object, isObject := value.(map[string]any)
		if !isObject {
			return nil
		}
		for key, member := range object {
			if err := checkKeyCase(member, valueType.Elem(), path+key+"."); err != nil {
				return err
			}
		}
	}
	return nil
}

// declaredField is a struct field and whether a `json` tag declared its key's spelling.
type declaredField struct {
	reflect.StructField
	tagged bool
}

// declaredFields maps each JSON key a struct declares to its field, the way encoding/json names
// them: the tag's name, or the Go field name when the tag gives none, with an untagged embedded
// struct's fields promoted into its parent.
func declaredFields(structType reflect.Type) map[string]declaredField {
	fields := map[string]declaredField{}
	for index := range structType.NumField() {
		field := structType.Field(index)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if field.Anonymous && name == "" {
			embedded := field.Type
			if embedded.Kind() == reflect.Pointer {
				embedded = embedded.Elem()
			}
			if embedded.Kind() == reflect.Struct {
				for promotedName, promoted := range declaredFields(embedded) {
					if _, shadowed := fields[promotedName]; !shadowed {
						fields[promotedName] = promoted
					}
				}
				continue
			}
		}
		if !field.IsExported() {
			continue
		}
		tagged := name != ""
		if !tagged {
			name = field.Name
		}
		fields[name] = declaredField{StructField: field, tagged: tagged}
	}
	return fields
}
