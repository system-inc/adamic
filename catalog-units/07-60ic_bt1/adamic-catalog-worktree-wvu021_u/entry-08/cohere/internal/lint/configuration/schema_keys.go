package configuration

import (
	"reflect"
	"sort"
	"strings"
)

// What a settings file may say, exported for the JSON schema and the reference generated from it
// (internal/settingsschema). The schema is read off these, never written beside them by hand: a key the
// loader learns and the schema lacks, or the reverse, is a test failure there rather than an editor
// that underlines a valid file or completes an invalid one.

// TopLevelKey is one key a settings file may carry.
type TopLevelKey struct {
	Name string

	// Field is the Go type the loader decodes the key into, nil for a key the loader accepts and
	// deliberately does not act on (IgnoredReason says why).
	Field         reflect.Type
	IgnoredReason string
}

// TopLevelKeys returns every key the loader accepts at the top of a settings file, sorted by name: the
// ones rawConfig decodes, typed by its fields, and the ones it ignores on purpose, with the reason.
func TopLevelKeys() []TopLevelKey {
	var keys []TopLevelKey
	for _, field := range jsonFields(reflect.TypeOf(rawConfig{})) {
		keys = append(keys, TopLevelKey{Name: field.Name, Field: field.Type})
	}
	for name, reason := range ignoredTopLevelKeys {
		keys = append(keys, TopLevelKey{Name: name, IgnoredReason: reason})
	}
	sort.Slice(keys, func(left, right int) bool { return keys[left].Name < keys[right].Name })
	return keys
}

// ParsedTopLevelKeyNames is the loader's own list of the keys it acts on, which the schema test holds
// equal to rawConfig's fields: two lists of one thing must not drift.
func ParsedTopLevelKeyNames() []string {
	return sortedKeys(parsedTopLevelKeys)
}

// OverrideField is one key an override block may carry.
type OverrideField struct {
	Name  string
	Field reflect.Type
}

// OverrideFields returns the keys an override block may carry, typed by rawOverride's fields.
func OverrideFields() []OverrideField {
	var fields []OverrideField
	for _, field := range jsonFields(reflect.TypeOf(rawOverride{})) {
		fields = append(fields, OverrideField{Name: field.Name, Field: field.Type})
	}
	return fields
}

// OverrideKeyNames is checkOverrideKeys' own list, held equal to rawOverride's fields by the same test.
func OverrideKeyNames() []string {
	return sortedKeys(overrideKeys)
}

// RuleValueType is the Go type a rule's value decodes from before parseRuleSetting reads it: the
// schema renders a field of this type as a rule setting rather than as free JSON.
var RuleValueType = reflect.TypeOf(rawConfig{}).Field(fieldIndex(reflect.TypeOf(rawConfig{}), "rules")).Type.Elem()

// ExtendsListType is the type `extends` decodes into: one path or rule set, or a list of them. The
// schema renders it as either shape.
var ExtendsListType = reflect.TypeOf(extendsList{})

// SeveritySpellings are the severities parseSeverity accepts, lowercase; it compares case-insensitively.
// The schema test proves parseSeverity accepts exactly these.
var SeveritySpellings = []string{"off", "allow", "warn", "warning", "error", "deny"}

// ParseSeverity is parseSeverity, for the test that holds SeveritySpellings to it.
func ParseSeverity(name string) error {
	_, err := parseSeverity(name)
	return err
}

type jsonField struct {
	Name string
	Type reflect.Type
}

// jsonFields lists a struct's fields by their JSON names, in declaration order.
func jsonFields(structType reflect.Type) []jsonField {
	var fields []jsonField
	for index := 0; index < structType.NumField(); index++ {
		field := structType.Field(index)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		fields = append(fields, jsonField{Name: name, Type: field.Type})
	}
	return fields
}

func fieldIndex(structType reflect.Type, jsonName string) int {
	for index := 0; index < structType.NumField(); index++ {
		if structType.Field(index).Tag.Get("json") == jsonName {
			return index
		}
	}
	panic("configuration: rawConfig has no field " + jsonName)
}

func sortedKeys(set map[string]bool) []string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
