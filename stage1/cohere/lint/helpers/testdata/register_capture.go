package rule

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// EntryOf is a registration as Register would hold it, for a test that hands rules to the walk without
// registering them in the process-wide catalog.
func EntryOf[Options any](registration Registration[Options]) (Entry, error) {
	return registration.entry()
}

// runPackage is the import path of the package whose code run is, read off the function itself, so the docs
// tell a bare house rule from an ESLint core port by where its Run lives rather than by a list. It is read from
// the author's Run before Bind wraps it, since the wrapper lives in this package.
func runPackage[Options any](run func(ruleContext Context, options Given[Options]) Listeners) string {
	function := runtime.FuncForPC(reflect.ValueOf(run).Pointer())
	if function == nil {
		return ""
	}
	// A function's name is its import path, then a dot, then the symbol and any closure suffixes. The
	// path's last segment may hold dots of its own only in a domain, which comes before the last slash.
	name := function.Name()
	slash := strings.LastIndex(name, "/")
	dot := strings.Index(name[slash+1:], ".")
	if dot < 0 {
		return name
	}
	return name[:slash+1+dot]
}

// Bind is the rule with its options decided: Run and ProgramFingerprint closed over options, so the walk hands
// a file to Bound.Run and never sees the options' type. The framework binds through an Entry; a test harness
// holding the Rule itself binds it here.
func Bind[Options any](subject Rule[Options], options Given[Options]) Bound {
	bound := Bound{
		Traits: traitsOf(subject),
		Run: func(ruleContext Context) Listeners {
			return subject.Run(ruleContext, options)
		},
	}
	if subject.ProgramFingerprint != nil {
		bound.ProgramFingerprint = func(program Program) [sha256.Size]byte {
			return subject.ProgramFingerprint(program, options)
		}
	}
	return bound
}

// traitsOf copies what a rule declares about itself into the Traits the framework reads.
func traitsOf[Options any](subject Rule[Options]) Traits {
	return Traits{
		Name:                    subject.Name,
		NoListener:              subject.NoListener,
		NeedsTypeChecker:        subject.NeedsTypeChecker,
		ProgramReads:            subject.ProgramReads,
		TypeReach:               subject.TypeReach,
		ResolvesReactValueTypes: subject.ResolvesReactValueTypes,
		HasProgramFingerprint:   subject.ProgramFingerprint != nil,
	}
}

// Registrant is a Registration of any options type, which is what Register takes, so one call registers rules
// whose options differ. Only Registration implements it.
type Registrant interface {
	entry() (Entry, error)
}

// Entry is a registered rule as the framework holds it, whatever its options type.
type Entry struct {
	Traits
	Arity           OptionArity
	RequiresOptions bool

	// Unconfigured is the rule as a bare severity runs it, handed no options.
	Unconfigured Bound

	// RunPackage is the import path of the package the rule's own Run is written in, which the docs read to
	// tell a bare house rule from an ESLint core port.
	RunPackage string

	// Bind decodes the rule's options and returns the rule ready to run with them. raw is the one option
	// element for OptionArityOne and OptionArityAt (nil for a bare severity under OptionArityAt, so the
	// decoder can fill a default from base), and the JSON array of every element for OptionArityList. base is
	// read only under OptionArityAt. Nil under OptionArityNone.
	Bind func(raw []byte, base OptionsBase) (Bound, error)
}

// Bound is a rule with its options decided, ready to run on files.
type Bound struct {
	Traits

	// Run is the rule's Run with its options already handed in.
	Run func(ruleContext Context) Listeners

	// ProgramFingerprint is the rule's ProgramFingerprint with its options already handed in, and nil when the
	// rule declares none (Traits.HasProgramFingerprint is false).
	ProgramFingerprint func(program Program) [sha256.Size]byte
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
func OptionKeys[Options any]() []string {
	targetType := reflect.TypeFor[Options]()
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
	registered      = map[string]Entry{}
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
func Register(registrations ...Registrant) {
	AdamicRegistrations = append(AdamicRegistrations, registrations...)
	registeredMutex.Lock()
	defer registeredMutex.Unlock()

	for _, registration := range registrations {
		entry, entryError := registration.entry()
		if entryError != nil {
			// Invariant: a registration sets at most one decoder, and Register runs from package init, so a
			// contradiction stops the program the moment it starts.
			panic("rule.Register: " + entryError.Error())
		}
		if entry.Name == "" {
			// Invariant: every rule names itself, and Register runs from package init, so a nameless rule
			// stops the program the moment it starts, as the comment above explains.
			panic("rule.Register: a rule was registered with an empty name, so the config could never address it")
		}
		if _, taken := registered[entry.Name]; taken {
			// Invariant: rule names are unique, and Register runs from package init, so a duplicate stops the
			// program the moment it starts, as the comment above explains.
			panic(fmt.Sprintf(
				"rule.Register: two rules answer to %q; the config addresses rules by name, so one of them "+
					"would be permanently unreachable while its tests passed",
				entry.Name,
			))
		}
		registered[entry.Name] = entry
	}
}

// Registered returns every registered rule, in a stable order.
//
// Order is by name, and it is stable so that two runs over the same tree report findings in the
// same sequence, which is what makes a diff between them meaningful. Sorting here rather than
// relying on registration order also removes the last thing that depended on `init` ordering across
// packages, which Go defines but which no reader should have to reason about.
func Registered() []Entry {
	registeredMutex.Lock()
	defer registeredMutex.Unlock()

	names := make([]string, 0, len(registered))
	for name := range registered {
		names = append(names, name)
	}
	sort.Strings(names)

	ordered := make([]Entry, 0, len(names))
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
	if decodeError := json.Unmarshal(list, &elements); decodeError != nil {
		return nil, fmt.Errorf("the option list is not a JSON array: %w", decodeError)
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
//
// Its errors name no Go type: the config layer puts the rule's name in front of them
// (configuration.OptionsRegistry.Decode), and a user who misconfigured a rule reads that name and what
// was wrong with the options they wrote, not how cohere spells the struct it decodes into.
func DecodeOptionsInto[Options any]() func(raw []byte) (Options, error) {
	return func(raw []byte) (Options, error) {
		var decoded Options
		if len(raw) == 0 {
			return decoded, fmt.Errorf("no options were configured")
		}
		if unmarshalOptionsError := UnmarshalOptions(raw, &decoded); unmarshalOptionsError != nil {
			return decoded, fmt.Errorf("reading its options: %w", unmarshalOptionsError)
		}
		return decoded, nil
	}
}
