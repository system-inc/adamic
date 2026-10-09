package configuration

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/lint/optionschema"
)

// Decoder turns a rule's raw JSON options into the typed value that rule expects.
//
// A rule declares its own options struct, and only the rule's own package knows that type. The
// config layer holds bytes. This function type is the seam: the registry pairs a rule with a
// decoder that knows how to produce its struct, and neither side has to import the other.
//
// It returns an error rather than a zero value, because a rule that silently receives a zero
// options struct is the exact defect this exists to fix. `boundary-no-project-import` was enabled
// and inert for months under the previous gate, declining every file because its LibraryDirectory
// was empty, and a liveness harness reporting `fixtures=54 live=53 dead=1` was the only thing that
// ever noticed.
type Decoder func(raw json.RawMessage) (any, error)

// DecodeInto builds a Decoder for any options struct.
//
// Used as `configuration.DecodeInto[BoundaryNoProjectImportOptions]()`, it decodes the config's JSON into
// that type and reports what went wrong if it cannot, naming the type so the message says which
// rule's options failed rather than only that some JSON did not parse.
func DecodeInto[Options any]() Decoder {
	return func(raw json.RawMessage) (any, error) {
		var decoded Options
		if len(raw) == 0 {
			return decoded, fmt.Errorf("no options were configured")
		}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return decoded, fmt.Errorf("decoding %T: %w", decoded, err)
		}
		return decoded, nil
	}
}

// RuleOptions pairs a rule name with how to decode its options and whether it can run without them.
//
// Exactly one of Decode and DecodeList is set, and which one is the rule's arity. Decode serves a
// rule that takes ONE option element and is handed that element alone. DecodeList serves a rule
// whose upstream schema takes more than one, and is handed every element as a JSON array, which is
// upstream's own `context.options` with the severity removed. The arity has to live here, beside
// the decoder, because this is the only layer that sees both the elements and the rule.
//
// Required is the load-bearing field, and it is the whole lesson of the inert-rule defect. A rule
// that needs an option and does not get one must fail loudly rather than decline quietly: declining
// looks identical to a rule with nothing to report, and that ambiguity hid a dead rule for months.
// A rule whose options only tune it, rather than enable it, sets Required false and runs on
// defaults.
type RuleOptions struct {
	Decode     Decoder
	DecodeList Decoder
	Required   bool
}

// OptionsRegistry maps rule names to their decoders.
type OptionsRegistry map[string]RuleOptions

// Decode produces the typed options for one rule from every option element the config wrote, or
// reports why it cannot.
//
// The outcomes are deliberately distinct rather than collapsed into "nil options":
//
//	no decoder, no elements        the rule takes no options; nil is correct and expected
//	no decoder, any element        an error naming the rule and each element
//	Decode, two or more elements   an error naming the rule and every element after the first
//	DecodeList, any count          the rule's own decoder reads the whole list, upstream's shape
//	decoder, elements present      the typed value
//	decoder, elements absent       an error if the rule requires options, nil if they merely tune it
//
// Collapsing the last case into nil is precisely how a required option goes missing without anyone
// finding out.
//
// # No element is dropped
//
// Every element the config wrote is either handed to the rule's decoder or refused here by name.
// The previous shape kept `tuple[1]`, so `["error", {"object": true},
// {"enforceForRenamedProperties": true}]` loaded clean, ran, and reported 0 on source the second
// element exists to flag, while the one-object spelling reported 1. A refusal costs one edit to the
// config; a silent drop costs an afternoon of somebody wondering why their option does nothing.
func (o OptionsRegistry) Decode(ruleName string, elements []json.RawMessage) (any, error) {
	declared, hasDeclared := o[ruleName]
	if !hasDeclared || (declared.Decode == nil && declared.DecodeList == nil) {
		if len(elements) > 0 {
			return nil, fmt.Errorf(
				"rule %s takes no options, and the config gives it %s, which it would never read. "+
					"Remove the options or write the bare severity", ruleName, describeElements(elements, 0))
		}
		return nil, nil
	}

	if err := refuseNulls(elements); err != nil {
		return nil, fmt.Errorf("rule %s: %w", ruleName, err)
	}

	var raw json.RawMessage
	decode := declared.Decode
	switch {
	case declared.DecodeList != nil:
		decode = declared.DecodeList
		if len(elements) > 0 {
			list, err := json.Marshal(elements)
			if err != nil {
				return nil, fmt.Errorf("rule %s: re-encoding its option elements: %w", ruleName, err)
			}
			raw = list
		}
	case len(elements) > 1:
		return nil, fmt.Errorf(
			"rule %s takes one option element, and the config gives it %d: %s would never be read, "+
				"so the entry is refused rather than half applied",
			ruleName, len(elements), describeElements(elements, 1))
	case len(elements) == 1:
		raw = elements[0]
	}

	decoded, err := decode(raw)
	if err != nil {
		if len(raw) == 0 && !declared.Required {
			// The rule tunes on options it did not get. Defaults are its own business.
			return nil, nil
		}
		return nil, fmt.Errorf("rule %s: %w", ruleName, err)
	}

	// A ported rule's options are also checked against upstream's own schema, so cohere accepts
	// exactly the shapes ESLint accepts (#pd2chkx). A Go decoder reads what its types let it: an
	// unknown enum string fell back to a default, a duplicate stayed a duplicate, and an entry missing
	// a required key decoded and was then dropped, all shapes ESLint refuses. One check here covers
	// every decoder.
	//
	// It runs after the decoder, so a shape both refuse keeps the decoder's message, which knows the
	// rule (an extra element is named by its text, an unknown key by the rule's own words); the
	// schema's refusal is for what the decoder let through. A bare severity is not checked, since no
	// upstream schema refuses an empty list (TestEverySchemaAcceptsTheBareSeverity).
	if len(elements) > 0 {
		if err := optionschema.Check(ruleName, elements); err != nil {
			return nil, fmt.Errorf("rule %s: %w", ruleName, err)
		}
	}
	return decoded, nil
}

// refuseNulls refuses a JSON null anywhere in a rule's option elements, naming the element and the
// path to it.
//
// No ported rule's upstream schema admits null, and ESLint's validator refuses it: measured on
// 2026-10-05 across the 185 option-taking rules from core, typescript-eslint, react, react-hooks,
// next and eslint-comments, plus better-tailwindcss, and none of ESLint 10.8.1's 6,641 corpus option
// rows writes one. Go's decoder reads it differently, as "leave this field alone", so
// `{"allowConstructorFlags": null}` loaded as the default with no error. That is a shape no ESLint
// version accepts, quietly accepted, which is what #pd2chkx exists to end. Refusing it here, the
// one layer every rule's elements pass through, covers every decoder at once, including the ones
// that read their element with json.Unmarshal rather than rule.UnmarshalOptions.
func refuseNulls(elements []json.RawMessage) error {
	for index, element := range elements {
		var value any
		if err := json.Unmarshal(element, &value); err != nil {
			// Malformed JSON is the decoder's to report, in its own words.
			continue
		}
		if path, found := nullPath(value, ""); found {
			at := fmt.Sprintf("element %d", index+1)
			if path != "" {
				at += " at " + path
			}
			return fmt.Errorf("%s is null, which no ESLint version accepts in a rule's options. "+
				"Remove it to take the default", at)
		}
	}
	return nil
}

// nullPath finds the first null in a decoded JSON value, walking object keys in sorted order so the
// same config always names the same path, and returns its dotted path.
func nullPath(value any, path string) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return path, true
	case []any:
		for index, member := range typed {
			if found, isNull := nullPath(member, joinPath(path, strconv.Itoa(index))); isNull {
				return found, true
			}
		}
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if found, isNull := nullPath(typed[key], joinPath(path, key)); isNull {
				return found, true
			}
		}
	}
	return "", false
}

func joinPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// describeElements names the option elements from index `from` on, numbered the way a config
// author counts them (the first element after the severity is element 1), so an error says which
// element it means and what it held.
func describeElements(elements []json.RawMessage, from int) string {
	described := make([]string, 0, len(elements)-from)
	for index := from; index < len(elements); index++ {
		described = append(described, fmt.Sprintf("element %d %s", index+1, truncate(string(elements[index]))))
	}
	return strings.Join(described, ", ")
}
