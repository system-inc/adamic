package react

import (
	"encoding/json"
	"fmt"
)

// NoMethodSetStateOptions configures the three lifecycle set-state rules: no-did-mount-set-state,
// no-did-update-set-state and no-will-update-set-state.
//
// Upstream builds all three from one factory, `makeNoMethodSetStateRule`, with one schema,
// `[{ enum: ["disallow-in-func"] }]`, and reads `context.options[0] || 'allow-in-func'`. So the
// option is a bare string with exactly one accepted value, and the choice it makes is a boolean.
// One options type for the three rules says that the way upstream does.
//
// Before #d21war2 the three took objects in two different shapes, `{"disallowInFunc": true}` and
// `{"mode": "disallow-in-func"}`, and accepted oxc's `"allowed"`; no ESLint version accepts any of
// those, and all of them are now refused.
type NoMethodSetStateOptions struct {
	// DisallowInFunc also reports a `setState` written inside a function nested in the lifecycle
	// method, rather than only one written directly in its body. Off by default.
	DisallowInFunc bool
}

// DecodeNoMethodSetStateOptions reads the option all three rules share.
//
// The option is a bare enum string rather than an object, so the generic helper, which unmarshals
// into a struct, cannot read it. Anything but `"disallow-in-func"` is refused, naming the value: an
// unknown value read as the default would leave an author who misspelled the option believing
// their nested calls are checked.
func DecodeNoMethodSetStateOptions(raw []byte) (any, error) {
	options := NoMethodSetStateOptions{}
	if len(raw) == 0 {
		return options, nil
	}
	var mode string
	if err := json.Unmarshal(raw, &mode); err != nil {
		return options, fmt.Errorf(`expected the mode string "disallow-in-func", got %s`, raw)
	}
	if mode != "disallow-in-func" {
		return options, fmt.Errorf(`mode %q is not "disallow-in-func", the only value upstream accepts`, mode)
	}
	options.DisallowInFunc = true
	return options, nil
}
