package react

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CompilerRuleOptions is the option the React Compiler rules take, which carries nothing.
//
// Upstream builds all fifteen from one factory with one schema, `[{ type: "object",
// additionalProperties: true }]`, and hands the object to the React Compiler as its configuration
// (`environment` and the like). cohere 1.0 carries no compiler configuration (ruled on #e06zm4b), so
// the only object accepted is the empty one, which ESLint 9 and later write into resolved configs as
// a default and which configures nothing.
type CompilerRuleOptions struct{}

// DecodeCompilerRuleOptions reads the one option the fifteen React Compiler rules share.
//
// An empty object is accepted and honoured, since it sets nothing. An object with keys is refused,
// naming them: it is compiler configuration, and accepting it would be an option read and ignored.
func DecodeCompilerRuleOptions(raw []byte) (any, error) {
	options := CompilerRuleOptions{}
	if len(raw) == 0 {
		return options, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return options, fmt.Errorf("expected an options object, got %s", raw)
	}
	if len(object) == 0 {
		return options, nil
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return options, fmt.Errorf("takes only {}: its options object is React Compiler configuration, "+
		"which cohere does not carry, so %s would be accepted and never honoured", strings.Join(keys, ", "))
}
