// Package registry is the list of every rule cohere knows.
//
// Rules are compiled in rather than loaded, which is what makes them free to run: a rule walks the
// same AST the parser already built, in the same address space, so the five hundredth rule costs
// what the hundredth does. It also means a rule present in the source cannot be silently absent at
// runtime, which a configuration-driven plugin system cannot promise. During the migration this
// tool replaces, three configurations ran successfully having loaded zero plugins.
//
// The catalog is assembled rather than written down. Each rule package registers its own rules in
// its own `init`, and this package's only job is to import every one of them and read the result.
// Before that, adding a rule meant appending a line to a sorted list here and often a second entry
// to an options map, so every port edited this file and two ports landing at once conflicted on a
// file neither was really changing.
//
// The blank imports below are therefore load-bearing rather than incidental. A rule package that
// nothing imports contributes nothing, so dropping one of these lines does not fail to compile, it
// makes those rules quietly stop existing. `TestParityAgainstInventory` is what turns that into a
// visible number, and it is the reason this indirection is safe to rely on.
package registry

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"

	_ "github.com/system-inc/cohere/internal/lint/rules/adamic"
	_ "github.com/system-inc/cohere/internal/lint/rules/base"
	_ "github.com/system-inc/cohere/internal/lint/rules/boundaries"
	_ "github.com/system-inc/cohere/internal/lint/rules/core"
	_ "github.com/system-inc/cohere/internal/lint/rules/next"
	_ "github.com/system-inc/cohere/internal/lint/rules/nexus"
	_ "github.com/system-inc/cohere/internal/lint/rules/react"
	_ "github.com/system-inc/cohere/internal/lint/rules/structure"
	_ "github.com/system-inc/cohere/internal/lint/rules/tailwind"
	_ "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

// All returns every rule, in a stable order.
//
// Order is stable so that two runs over the same tree report findings in the same sequence, which
// is what makes a diff between them meaningful. `rule.Registered` sorts by name, so the order no
// longer depends on how a human kept a list tidy, nor on the order Go happens to run `init`.
func All() []rule.Rule {
	registrations := rule.Registered()
	rules := make([]rule.Rule, 0, len(registrations))
	for _, registration := range registrations {
		rules = append(rules, registration.Rule)
	}
	return rules
}

// Names returns every registered rule's name, in the order All returns them.
func Names() []string {
	rules := All()
	names := make([]string, 0, len(rules))
	for _, registeredRule := range rules {
		names = append(names, registeredRule.Name)
	}
	return names
}

// Count is how many rules exist, for the coverage line.
func Count() int {
	return len(All())
}

// Options says how to decode each rule's configuration, and which rules cannot run without it.
//
// A rule declares its own options struct and only its own package knows that type, while the config
// layer holds JSON. A rule's registration carries the decoder across that gap, and this function
// only reshapes what the packages already declared.
//
// Required is the field that matters. `boundary-no-project-import` was enabled and inert for months
// under the gate cohere replaces: it declines every file when LibraryDirectory is empty, which is
// correct behavior for a misconfigured guard and indistinguishable from a rule with nothing to
// report. A liveness harness reporting `fixtures=54 live=53 dead=1` was the only thing that ever
// caught it. Marking it required turns that silence into a failure.
//
// A rule that registers no decoder takes no options, which is the common case, and the config layer
// refuses any option written for it. Which decoder a rule registers is its arity: Decode for one
// option element, DecodeOptionList for upstream's multi-element `context.options`.
//
// This form knows no base, so a rule whose options name a place on disk can only take an absolute
// path from it and refuses a relative or absent one by name. The command uses OptionsAt.
func Options() configuration.OptionsRegistry {
	return OptionsAt(rule.OptionsBase{})
}

// OptionsAt is Options for a run that knows where its config lives and which project it checks, so a
// rule registered with DecodeAt can anchor a relative path, or default an absent one, rather than
// matching nothing.
func OptionsAt(base rule.OptionsBase) configuration.OptionsRegistry {
	options := configuration.OptionsRegistry{}
	for _, registration := range rule.Registered() {
		switch {
		case registration.DecodeAt != nil:
			decodeAt := registration.DecodeAt
			options[registration.Rule.Name] = configuration.RuleOptions{
				Decode:   func(raw json.RawMessage) (any, error) { return decodeAt(raw, base) },
				Required: registration.RequiresOptions,
			}
		case registration.DecodeOptionList != nil:
			decodeList := registration.DecodeOptionList
			options[registration.Rule.Name] = configuration.RuleOptions{
				DecodeList: func(raw json.RawMessage) (any, error) { return decodeList(raw) },
				Required:   registration.RequiresOptions,
			}
		case registration.Decode != nil:
			decode := registration.Decode
			options[registration.Rule.Name] = configuration.RuleOptions{
				Decode:   func(raw json.RawMessage) (any, error) { return decode(raw) },
				Required: registration.RequiresOptions,
			}
		}
	}
	return options
}
