package rule

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// Registration is one rule plus everything the tool needs to know about it.
//
// A rule's options decoder lives here rather than on Rule itself, because Rule is the type a rule
// author writes and a decoder is the config layer's business. Keeping them in one Registration
// means a rule is declared in exactly one place while neither package has to import the other: the
// rule package builds the decoder with a generic helper and hands it over as an opaque function.
type Registration struct {
	Rule Rule
	// Decode turns this rule's single option element into its own options struct, or nil if the
	// rule takes no options, which is the common case. A config giving a rule with Decode more than
	// one option element is refused by name, never truncated.
	Decode func(raw []byte) (any, error)
	// DecodeOptionList is Decode for a rule whose upstream schema takes more than one option
	// element: `eqeqeq`'s `["error", "always", {"null": "ignore"}]`, or `consistent-this`'s
	// variadic `["error", "self", "vm"]`. It is handed every element after the severity as one JSON
	// array, which is upstream's `context.options`, or nil for a bare severity. Set at most one of
	// Decode and DecodeOptionList; which one is set is how the config layer learns the rule's arity.
	DecodeOptionList func(list []byte) (any, error)
	// DecodeAt is Decode for a rule whose options name a place on disk, so decoding needs to know
	// where the config was written and which project is being checked. It takes one option element,
	// as Decode does, and is handed nil for a bare severity so it can fill in a default from the base.
	// Set at most one of Decode, DecodeOptionList and DecodeAt.
	//
	// It exists because a path in a committed config is read on every machine and in every worktree.
	// An absolute one is right on exactly one of them, and a rule whose root matches none of the files
	// it is handed declines them all, which is the inert-rule defect with a path in front of it.
	DecodeAt func(raw []byte, base OptionsBase) (any, error)
	// RequiresOptions is whether the rule declines every file without its options.
	//
	// The load-bearing field, and the whole lesson of the inert-rule defect.
	// `boundary-no-project-import` was enabled and inert for months, declining every file because
	// its configured directory was empty, which is correct behavior for a misconfigured guard and
	// indistinguishable from a rule with nothing to report. Marking it required turns that silence
	// into a failure.
	RequiresOptions bool
}

// OptionsBase is what a decoder registered with DecodeAt knows beyond its own option element: where
// the config's paths are anchored, and the project's per-plugin settings.
//
// Either path may be empty, which means the caller does not know it. A decoder that needs one it
// was not given fails rather than guessing, because guessing is how a root ends up matching nothing.
type OptionsBase struct {
	// ConfigDirectory is the absolute directory of the config file the options were written in. A
	// relative path in an option resolves against it, the way a relative path in a tsconfig does.
	ConfigDirectory string

	// ProjectRoot is the absolute root of the project being checked, as cohere discovered it.
	ProjectRoot string

	// Settings is the project's own `settings`, keyed by plugin namespace, each value as written. A
	// decoder reads only its own plugin's, and only a namespace some package registered with
	// RegisterSettings is ever here: the run refuses any other before a rule decodes (see
	// registry.CheckSettings).
	Settings map[string]json.RawMessage
}

// SettingsRegistration is how a rule package says it reads one plugin's `settings`, the way ESLint
// hands every rule `context.settings` and a plugin reads its own namespace from it.
type SettingsRegistration struct {
	// Namespaces are the keys under `settings` this plugin reads, in the order it prefers them when
	// more than one is written. better-tailwindcss reads two spellings of its own name.
	Namespaces []string

	// Split hands back the part of one namespace's settings each rule reads, keyed by rule name, or
	// an error naming every key no rule of the plugin reads. A key nothing reads is refused rather
	// than dropped, for the reason an unknown option key is: it loads clean and does nothing.
	Split func(raw []byte) (map[string]json.RawMessage, error)
}

var (
	settingsMutex      sync.Mutex
	settingsRegistered []SettingsRegistration
)

// RegisterSettings records that a rule package reads a plugin's settings.
func RegisterSettings(registration SettingsRegistration) {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()
	settingsRegistered = append(settingsRegistered, registration)
}

// RegisteredSettings returns every settings registration, for the check that refuses a namespace no
// package reads.
func RegisteredSettings() []SettingsRegistration {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()
	return append([]SettingsRegistration(nil), settingsRegistered...)
}

// OptionKeys returns every key an options struct declares, the way UnmarshalOptions reads them: the
// `json` tag's name, or the Go field name for an untagged field, with an embedded struct's keys
// promoted. Sorted.
func OptionKeys(target any) []string {
	targetType := reflect.TypeOf(target)
	for targetType != nil && targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}
	if targetType == nil || targetType.Kind() != reflect.Struct {
		return nil
	}
	keys := make([]string, 0, targetType.NumField())
	for key := range declaredFields(targetType) {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// registered is every rule any linked package has registered, keyed by name.
//
// Guarded by a mutex because `init` functions across packages are ordered but a test may register
// from a goroutine, and a map written concurrently crashes the process rather than corrupting
// quietly. The cost is paid once per rule at startup.
var (
	registeredMutex sync.Mutex
	registered      = map[string]Registration{}
)

// Register adds a rule to the catalog, and panics if the name is already taken.
//
// Called from each rule package's own `init`, which is what removes the shared file every port used
// to edit. Before this, adding a rule meant appending one line to a single sorted list in another
// package, so two ports landing at once conflicted on a file neither of them was really changing.
// Now a rule is two new files and no shared edit, which is the difference between porting a few
// rules at a time and porting a family at once.
//
// A duplicate name panics rather than overwriting. Two rules answering to one name is not a
// resolvable state: the config addresses rules by name, so whichever one lost would be silently
// unreachable while its tests kept passing, which is the same shape as the inert-rule defect this
// tool exists to catch. Failing at startup makes it a crash in the first second rather than a gap
// nobody measures.
func Register(registrations ...Registration) {
	registeredMutex.Lock()
	defer registeredMutex.Unlock()

	for _, registration := range registrations {
		name := registration.Rule.Name
		if name == "" {
			panic("rule.Register: a rule was registered with an empty name, so the config could never address it")
		}
		decoders := 0
		for _, set := range []bool{registration.Decode != nil, registration.DecodeOptionList != nil, registration.DecodeAt != nil} {
			if set {
				decoders++
			}
		}
		if decoders > 1 {
			panic(fmt.Sprintf(
				"rule.Register: %q sets more than one of Decode, DecodeOptionList and DecodeAt; the "+
					"config layer reads which one is set as the rule's option arity, so two is a contradiction",
				name,
			))
		}
		if existing, taken := registered[name]; taken {
			panic(fmt.Sprintf(
				"rule.Register: two rules answer to %q (%T and %T); the config addresses rules by "+
					"name, so one of them would be permanently unreachable while its tests passed",
				name, existing.Rule.Run, registration.Rule.Run,
			))
		}
		registered[name] = registration
	}
}

// Registered returns every registered rule, in a stable order.
//
// Order is by name, and it is stable so that two runs over the same tree report findings in the
// same sequence, which is what makes a diff between them meaningful. Sorting here rather than
// relying on registration order also removes the last thing that depended on `init` ordering across
// packages, which Go defines but which no reader should have to reason about.
func Registered() []Registration {
	registeredMutex.Lock()
	defer registeredMutex.Unlock()

	names := make([]string, 0, len(registered))
	for name := range registered {
		names = append(names, name)
	}
	sort.Strings(names)

	ordered := make([]Registration, 0, len(names))
	for _, name := range names {
		ordered = append(ordered, registered[name])
	}
	return ordered
}

// OptionElements splits the list a DecodeOptionList decoder is handed into its elements, refusing
// more than `maximum` by naming each extra one.
//
// The refusal is the point. A rule whose upstream schema has two elements and is handed three would
// otherwise read two and drop the third, which is the defect this whole arity mechanism exists to
// end, moved one layer in. Nil or empty input is no elements, which is a bare severity.
//
// The message leaves out the rule's name because the config layer prefixes every decoder error with
// it.
func OptionElements(list []byte, maximum int) ([]json.RawMessage, error) {
	if len(list) == 0 {
		return nil, nil
	}
	var elements []json.RawMessage
	if err := json.Unmarshal(list, &elements); err != nil {
		return nil, fmt.Errorf("the option list is not a JSON array: %w", err)
	}
	if len(elements) <= maximum {
		return elements, nil
	}
	extras := make([]string, 0, len(elements)-maximum)
	for index := maximum; index < len(elements); index++ {
		extras = append(extras, fmt.Sprintf("element %d %s", index+1, elements[index]))
	}
	return nil, fmt.Errorf("takes at most %d option elements, and the config gives it %d: %s would never be read",
		maximum, len(elements), strings.Join(extras, ", "))
}

// DecodeOptionsInto builds the decoder a Registration carries, for any options struct.
//
// Used as `rule.DecodeOptionsInto[NoConcatenatedClassesOptions]()`, it lives here rather than in
// `internal/config` for a measured reason: a rule package that imports the config layer stops being
// a leaf of the import graph, and editing a rule then rebuilds in about 8.5s instead of about 1.8s.
// `TestRulePackagesStayLeaves` asserts that, and it caught this file importing config the first
// time registration moved into the rule packages.
//
// Nothing here is config-specific. It decodes JSON into a struct, which is what the rule declared
// and what only the rule's own package knows the type of. The config layer's job is deciding
// whether a rule that got no options may still run, and that judgment stays there.
func DecodeOptionsInto[Options any]() func(raw []byte) (any, error) {
	return func(raw []byte) (any, error) {
		var decoded Options
		if len(raw) == 0 {
			return decoded, fmt.Errorf("no options were configured")
		}
		if err := UnmarshalOptions(raw, &decoded); err != nil {
			return decoded, fmt.Errorf("decoding %T: %w", decoded, err)
		}
		return decoded, nil
	}
}
