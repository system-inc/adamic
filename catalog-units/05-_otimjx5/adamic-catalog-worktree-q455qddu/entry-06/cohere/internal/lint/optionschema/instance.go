package optionschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Violation is the first place an option list departs from upstream's schema: where, as a path into the
// list ESLint validates (`[0].allow[1].from`), and what the schema expected there.
type Violation struct {
	Path     string
	Expected string
}

// Error names the element the way a config author counts them, the first after the severity being
// element 1, and the path inside it: `element 1 at allow[1].from: upstream's schema expects ...`.
func (violation *Violation) Error() string {
	where := violation.Path
	if rest, isElement := strings.CutPrefix(violation.Path, "["); isElement {
		if closing := strings.Index(rest, "]"); closing > 0 {
			if index, err := strconv.Atoi(rest[:closing]); err == nil {
				where = "element " + strconv.Itoa(index+1)
				if inner := strings.TrimPrefix(rest[closing+1:], "."); inner != "" {
					where += " at " + inner
				}
			}
		}
	}
	return fmt.Sprintf("%s: upstream's schema, which ESLint validates against, expects %s", where, violation.Expected)
}

// Extension is a shape cohere accepts on purpose beyond upstream's schema: one key of one option
// element, removed before the element is checked and left for the rule's own decoder to read.
type Extension struct {
	Rule string
	// Element counts as a config author does: 1 is the first element after the severity.
	Element int
	Key     string
	Reason  string
}

// Extensions are every shape cohere accepts that upstream's schema refuses. Each is a decision, so the
// list's length is asserted exactly (TestExtensionsAreExactlyTheDeclaredOnes) and adding one means
// changing that number on purpose (#pd2chkx condition 6).
var Extensions = []Extension{
	{
		Rule:    "boundaries/dependencies",
		Element: 1,
		Key:     "elements",
		Reason: "upstream reads its element descriptors from settings[\"boundaries/elements\"] beside the rule, " +
			"and cohere carries no boundaries settings, so they are written in the rule's own options",
	},
}

// Check validates a rule's option elements, the list after the severity, against upstream's schema. A
// rule with no schema here, either not a port or one whose upstream opts out, passes unchecked.
func Check(ruleName string, elements []json.RawMessage) error {
	schema, found, err := For(ruleName)
	if err != nil || !found {
		return err
	}
	return schema.Validate(withoutExtensions(ruleName, elements))
}

// withoutExtensions removes each declared extension key from its element, leaving the rest to be
// checked as upstream wrote its schema.
func withoutExtensions(ruleName string, elements []json.RawMessage) []json.RawMessage {
	var stripped []json.RawMessage
	for _, extension := range Extensions {
		if extension.Rule != ruleName || extension.Element < 1 || extension.Element > len(elements) {
			continue
		}
		var object map[string]json.RawMessage
		if json.Unmarshal(elements[extension.Element-1], &object) != nil {
			continue
		}
		if _, present := object[extension.Key]; !present {
			continue
		}
		delete(object, extension.Key)
		encoded, err := json.Marshal(object)
		if err != nil {
			continue
		}
		if stripped == nil {
			stripped = append([]json.RawMessage(nil), elements...)
		}
		stripped[extension.Element-1] = encoded
	}
	if stripped == nil {
		return elements
	}
	return stripped
}

// Validate checks an option list against the schema and returns its first Violation, or nil.
func (schema *Schema) Validate(elements []json.RawMessage) error {
	list := make([]any, len(elements))
	for index, element := range elements {
		decoder := json.NewDecoder(bytes.NewReader(element))
		decoder.UseNumber()
		if err := decoder.Decode(&list[index]); err != nil {
			return fmt.Errorf("element %d is not JSON: %w", index+1, err)
		}
	}
	if violation := validate(schema.root, list, ""); violation != nil {
		return violation
	}
	return nil
}

func validate(schema *node, value any, path string) *Violation {
	if schema.referenceTarget != nil {
		return validate(schema.referenceTarget, value, path)
	}
	if schema.always {
		return nil
	}
	if schema.never {
		return &Violation{Path: pathOrRoot(path), Expected: "nothing at all here"}
	}

	if len(schema.types) > 0 && !hasAnyType(value, schema.types) {
		return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("%s, got %s", typeList(schema.types), describe(value))}
	}
	if schema.hasConst && !equal(value, schema.constant) {
		return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("%s, got %s", encode(schema.constant), describe(value))}
	}
	if schema.enumValues != nil {
		matched := false
		for _, candidate := range schema.enumValues {
			if equal(value, candidate) {
				matched = true
				break
			}
		}
		if !matched {
			listed := make([]string, len(schema.enumValues))
			for index, candidate := range schema.enumValues {
				listed[index] = encode(candidate)
			}
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("one of %s, got %s", strings.Join(listed, ", "), describe(value))}
		}
	}

	switch typed := value.(type) {
	case json.Number:
		number, _ := typed.Float64()
		if schema.minimum != nil && number < *schema.minimum {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("a number of at least %s, got %s", formatNumber(*schema.minimum), typed)}
		}
		if schema.maximum != nil && number > *schema.maximum {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("a number of at most %s, got %s", formatNumber(*schema.maximum), typed)}
		}
	case string:
		length := utf8.RuneCountInString(typed)
		if schema.minLength != nil && length < *schema.minLength {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("a string of at least %d characters, got %s", *schema.minLength, describe(value))}
		}
		if schema.maxLength != nil && length > *schema.maxLength {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("a string of at most %d characters, got %s", *schema.maxLength, describe(value))}
		}
		// A match that runs past esregexp's time bound counts as a match: this check may only refuse
		// what ajv refuses, and a refusal stops the whole run.
		if schema.pattern != nil && !schema.pattern.TestOrTimeout(typed) {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("a string matching /%s/, got %s", schema.patternSource, describe(value))}
		}
	case []any:
		if violation := validateArray(schema, typed, path); violation != nil {
			return violation
		}
	case map[string]any:
		if violation := validateObject(schema, typed, path); violation != nil {
			return violation
		}
	}

	for _, branch := range schema.allOf {
		if violation := validate(branch, value, path); violation != nil {
			return violation
		}
	}
	if len(schema.anyOf) > 0 {
		var first *Violation
		matched := false
		for _, branch := range schema.anyOf {
			violation := validate(branch, value, path)
			if violation == nil {
				matched = true
				break
			}
			if first == nil {
				first = violation
			}
		}
		if !matched {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("one of %d shapes, and matches none (the first: %s)", len(schema.anyOf), first.Error())}
		}
	}
	if len(schema.oneOf) > 0 {
		matches := 0
		var first *Violation
		for _, branch := range schema.oneOf {
			violation := validate(branch, value, path)
			if violation == nil {
				matches++
			} else if first == nil {
				first = violation
			}
		}
		switch {
		case matches == 0:
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("exactly one of %d shapes, and matches none (the first: %s)", len(schema.oneOf), first.Error())}
		case matches > 1:
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("exactly one of %d shapes, and matches %d of them", len(schema.oneOf), matches)}
		}
	}
	if schema.not != nil && validate(schema.not, value, path) == nil {
		return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("anything but %s", describe(value))}
	}
	return nil
}

func validateArray(schema *node, list []any, path string) *Violation {
	if schema.minItems != nil && len(list) < *schema.minItems {
		return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("at least %d items, got %d", *schema.minItems, len(list))}
	}
	if schema.maxItems != nil && len(list) > *schema.maxItems {
		return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("at most %d items, got %d", *schema.maxItems, len(list))}
	}
	if schema.uniqueItems {
		for later := 1; later < len(list); later++ {
			for earlier := 0; earlier < later; earlier++ {
				if equal(list[earlier], list[later]) {
					return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("unique items, and items %d and %d are both %s", earlier, later, encode(list[later]))}
				}
			}
		}
	}
	for index, item := range list {
		itemPath := path + "[" + strconv.Itoa(index) + "]"
		var itemSchema *node
		switch {
		case schema.tupleItems != nil && index < len(schema.tupleItems):
			itemSchema = schema.tupleItems[index]
		case schema.tupleItems != nil && schema.noAdditionalItems:
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("at most %d items, got %d", len(schema.tupleItems), len(list))}
		case schema.tupleItems != nil:
			itemSchema = schema.additionalItems
		default:
			itemSchema = schema.items
		}
		if itemSchema == nil {
			continue
		}
		if violation := validate(itemSchema, item, itemPath); violation != nil {
			return violation
		}
	}
	return nil
}

func validateObject(schema *node, object map[string]any, path string) *Violation {
	for _, name := range schema.required {
		if _, present := object[name]; !present {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("the key %q, which is required", name)}
		}
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		keyPath := path + "." + key
		declared := false
		if property, found := schema.properties[key]; found {
			declared = true
			if violation := validate(property, object[key], keyPath); violation != nil {
				return violation
			}
		}
		for _, patterned := range schema.patternProperties {
			if patterned.pattern.TestOrTimeout(key) {
				declared = true
				if violation := validate(patterned.schema, object[key], keyPath); violation != nil {
					return violation
				}
			}
		}
		if declared {
			continue
		}
		if schema.noAdditional {
			return &Violation{Path: pathOrRoot(path), Expected: fmt.Sprintf("only its declared keys, and %q is not one", key)}
		}
		if schema.additionalProperties != nil {
			if violation := validate(schema.additionalProperties, object[key], keyPath); violation != nil {
				return violation
			}
		}
	}
	return nil
}

// hasAnyType is ajv's type check: "integer" is a number with no fractional part.
func hasAnyType(value any, types []string) bool {
	for _, name := range types {
		switch name {
		case "null":
			if value == nil {
				return true
			}
		case "boolean":
			if _, is := value.(bool); is {
				return true
			}
		case "string":
			if _, is := value.(string); is {
				return true
			}
		case "number":
			if _, is := value.(json.Number); is {
				return true
			}
		case "integer":
			if number, is := value.(json.Number); is {
				parsed, err := number.Float64()
				if err == nil && !math.IsInf(parsed, 0) && parsed == math.Trunc(parsed) {
					return true
				}
			}
		case "array":
			if _, is := value.([]any); is {
				return true
			}
		case "object":
			if _, is := value.(map[string]any); is {
				return true
			}
		}
	}
	return false
}

// equal is ajv's deep equality (fast-deep-equal): numbers by value, objects by key set and values.
func equal(left any, right any) bool {
	switch typedLeft := left.(type) {
	case nil:
		return right == nil
	case bool:
		typedRight, is := right.(bool)
		return is && typedLeft == typedRight
	case string:
		typedRight, is := right.(string)
		return is && typedLeft == typedRight
	case json.Number:
		typedRight, is := right.(json.Number)
		if !is {
			return false
		}
		leftValue, leftErr := typedLeft.Float64()
		rightValue, rightErr := typedRight.Float64()
		return leftErr == nil && rightErr == nil && leftValue == rightValue
	case []any:
		typedRight, is := right.([]any)
		if !is || len(typedLeft) != len(typedRight) {
			return false
		}
		for index := range typedLeft {
			if !equal(typedLeft[index], typedRight[index]) {
				return false
			}
		}
		return true
	case map[string]any:
		typedRight, is := right.(map[string]any)
		if !is || len(typedLeft) != len(typedRight) {
			return false
		}
		for key, value := range typedLeft {
			other, present := typedRight[key]
			if !present || !equal(value, other) {
				return false
			}
		}
		return true
	}
	return false
}

func count(value any) (*int, error) {
	number, isNumber := value.(json.Number)
	if !isNumber {
		return nil, fmt.Errorf("must be a number")
	}
	parsed, err := strconv.Atoi(string(number))
	if err != nil || parsed < 0 {
		return nil, fmt.Errorf("must be a non-negative integer")
	}
	return &parsed, nil
}

func number(value any) (*float64, error) {
	typed, isNumber := value.(json.Number)
	if !isNumber {
		return nil, fmt.Errorf("must be a number")
	}
	parsed, err := typed.Float64()
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func escapePointer(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "~", "~0"), "/", "~1")
}

// pointerTarget follows a JSON pointer reference (`#/items/0/$defs/name`) through the document.
func pointerTarget(document any, reference string) (any, error) {
	if reference == "#" {
		return document, nil
	}
	rest, found := strings.CutPrefix(reference, "#/")
	if !found {
		return nil, fmt.Errorf("$ref %q is not a pointer into this schema", reference)
	}
	current := document
	for _, part := range strings.Split(rest, "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			next, present := typed[part]
			if !present {
				return nil, fmt.Errorf("$ref %q names %q, which is not there", reference, part)
			}
			current = next
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, fmt.Errorf("$ref %q names item %q, which is not there", reference, part)
			}
			current = typed[index]
		default:
			return nil, fmt.Errorf("$ref %q runs past a value at %q", reference, part)
		}
	}
	return current, nil
}

func pathOrRoot(path string) string {
	if path == "" {
		return "the option list"
	}
	return path
}

func typeList(types []string) string {
	if len(types) == 1 {
		return "a value of type " + types[0]
	}
	return "a value of type " + strings.Join(types[:len(types)-1], ", ") + " or " + types[len(types)-1]
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func encode(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(encoded)
}

func describe(value any) string {
	encoded := encode(value)
	if len(encoded) > 60 {
		encoded = encoded[:57] + "..."
	}
	return encoded
}
