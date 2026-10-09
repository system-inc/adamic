// Theme resolution: the store that turns a repository's `@theme` blocks into the namespace lookup
// every utility consults.
//
// Ported from `src/theme.ts` at Tailwind 4.3.3, read at the pinned tag rather than from the
// minified bundle, so a disagreement is a finding rather than version skew.
//
// # Why this component carries the argument for the whole port
//
// The generated tables cohere shipped before this were per-repository, not per-Tailwind-release:
// running the generator against two repositories on the same Tailwind 4.3.3 produced tables that
// differed, because each project's own `@theme` entries were baked in as though they were framework
// facts. Measured on the corpus this file is tested against, ahra resolves 744 theme entries and
// www-connected-app resolves 748, on the same engine. This is the code that reads the repository in
// front of us instead of shipping one repository's tokens as everyone's.
//
// # Insertion order is part of the contract, not an implementation detail
//
// Upstream stores entries in a JavaScript `Map`, and three of its methods return values in
// insertion order: `entries`, `keysInNamespaces`, and `namespace`. A Go `map` has no order, and its
// iteration is deliberately randomized per run, so a port backed by a plain map would return a
// different answer on every execution while passing any test that sorted before comparing.
//
// So this stores the order explicitly, in `keyOrder`, alongside the map. Two behaviours of
// `Map.set` are reproduced exactly and both are observable:
//
//   - Overwriting an existing key keeps its original position. `--color-a: red; --color-b: blue;
//     --color-a: green` yields `a, b` with a's value updated, not `b, a`. An append-on-every-write
//     port passes every single-assignment test and reorders every redefinition, which is what a
//     repository's own `@theme` does to the framework defaults it overrides on nearly every key.
//   - Deleting a key and re-adding it puts it at the end, because the position is genuinely gone.
//
// Deleting blanks the key's slot rather than removing it, so a delete stays O(1) amortized. Every
// read skips blank slots, and `compactKeyOrder` reclaims them once the dead outnumber the live,
// which bounds the slice at twice the live size rather than at the total number of writes. Blanking
// rather than tombstoning the key name is load-bearing: a slot still holding its key would be
// indistinguishable from a live one after that key was re-added, and every ordered read would yield
// it twice. No corpus stylesheet deletes and re-adds a key, so the corpus cannot catch that; the
// unit test in theme_test.go is what does.
//
// # What was deliberately not ported, and why
//
// `keyframes` and its two methods. Upstream collects `@keyframes` rules found inside `@theme` so it
// can re-emit them after the `@theme` rule is removed from the output. cohere emits no CSS, and the
// reading of a candidate, the `{order, count}` pair the class-order comparator sorts on, never
// consults them. Adding a set that nothing reads would be untested surface that reads as supported.
//
// `src` on each entry, the `SourceLocation` upstream carries so it can build source maps. cohere
// reports on the class literal in a user's source file, never on generated CSS.
//
// `markUsedVariable` and `ThemeOptionUsed` are ported, because the bit is observable through
// `GetOptions` and a port that dropped it would silently answer a different bitfield than the
// engine. Nothing in cohere calls it yet; see its doc comment.
package tailwind

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// ThemeOptions is the bitfield upstream stores next to every theme value.
//
// The values are upstream's `const enum ThemeOptions` and the numbers are load-bearing rather than
// arbitrary: they are what `GetOptions` returns and what the fixtures compare against, so renaming
// them is free and renumbering them is not.
type ThemeOptions int

const (
	// ThemeOptionNone is the absence of every flag.
	ThemeOptionNone ThemeOptions = 0
	// ThemeOptionInline makes Resolve return the literal value instead of a `var()` reference.
	ThemeOptionInline ThemeOptions = 1 << 0
	// ThemeOptionReference marks a value that emits no CSS variable, so a `var()` built from it
	// gets the literal value as a fallback.
	ThemeOptionReference ThemeOptions = 1 << 1
	// ThemeOptionDefault marks a value that yields to any non-default value for the same key.
	ThemeOptionDefault ThemeOptions = 1 << 2
	// ThemeOptionStatic marks a value that is emitted whether or not it is used.
	ThemeOptionStatic ThemeOptions = 1 << 3
	// ThemeOptionUsed marks a value some candidate has referenced. See MarkUsedVariable.
	ThemeOptionUsed ThemeOptions = 1 << 4
)

// ignoredThemeKeys is upstream's `ignoredThemeKeyMap`: for a namespace, the keys that live under it
// lexically but belong to a different namespace.
//
// `--font-weight-bold` starts with `--font-`, so a naive prefix match would report it as the key
// `weight-bold` in the `--font` namespace. It is not; it is `bold` in `--font-weight`. Measured on
// this repository's theme, `--color` holds 567 keys and `keysInNamespaces(['--color'])` returns 558
// for the same reason via a different namespace, and the gap is invisible to any check that only
// counted one of them.
//
// Upstream's note that this may one day be replaced by letting the computer sort out overlapping
// prefixes is carried here as a reminder that the list is a fact about Tailwind's own namespace
// choices, not a general rule that can be derived.
var ignoredThemeKeys = map[string][]string{
	"--font":        {"--font-weight", "--font-size"},
	"--inset":       {"--inset-shadow", "--inset-ring"},
	"--text":        {"--text-color", "--text-decoration-color", "--text-decoration-thickness", "--text-indent", "--text-shadow", "--text-underline-offset"},
	"--grid-column": {"--grid-column-start", "--grid-column-end"},
	"--grid-row":    {"--grid-row-start", "--grid-row-end"},
}

// isIgnoredThemeKey reports whether themeKey belongs to a different namespace than the one asking.
//
// The match is exact equality or a `-` boundary, never a bare prefix. `--font-weightless` is not
// ignored by `--font-weight` because `weightless` continues the word rather than starting a new
// segment, and a port using `strings.HasPrefix(themeKey, ignored)` alone would drop it from the
// `--font` namespace where the engine keeps it.
func isIgnoredThemeKey(themeKey, namespace string) bool {
	for _, ignoredThemeKey := range ignoredThemeKeys[namespace] {
		if themeKey == ignoredThemeKey || strings.HasPrefix(themeKey, ignoredThemeKey+"-") {
			return true
		}
	}
	return false
}

// themeValue is one stored entry.
type themeValue struct {
	value   string
	options ThemeOptions
}

// Theme is the resolved theme: every `--custom-property` the stylesheet's `@theme` blocks declared,
// in the order they were declared, with the options each was declared under.
//
// The zero value is not usable; build one with NewTheme. A Theme is not safe for concurrent
// mutation, and is safe for concurrent reads once built. That split is the intended usage: the
// design system is built once per run and shared read-only across the worker pool.
type Theme struct {
	// Prefix is the `prefix(...)` of an `@theme` block, without dashes, or the empty string when
	// none was given.
	//
	// Upstream this is `string | null` and the distinction between null and empty never arises,
	// because `parseThemeOptions` only ever sets it from `prefix(x)` where x is non-empty and
	// validated against `/^[a-z]+$/`. The empty string therefore means "no prefix" unambiguously.
	Prefix string

	values map[string]themeValue
	// keyOrder is insertion order. It may contain keys absent from values; see the note on
	// ordering in the package comment.
	keyOrder []string
	// deadKeys counts entries in keyOrder that are no longer in values, so compaction can be
	// amortized rather than run on every delete.
	deadKeys int
}

// NewTheme returns an empty Theme.
func NewTheme() *Theme {
	return &Theme{values: make(map[string]themeValue)}
}

// Size is the number of entries, upstream's `get size()`.
func (theme *Theme) Size() int {
	return len(theme.values)
}

// Add records a theme entry, and is upstream's `add`.
//
// Four behaviours in one method, in upstream's order, because the order matters:
//
//   - A key ending `-*` is a clear directive rather than an entry. `--*` clears everything; anything
//     else clears that namespace. The value must be `initial`, and is an error otherwise.
//   - A ThemeOptionDefault value loses to an existing non-default value for the same key. This is
//     how `@theme default` in the framework's own `theme.css` yields to a repository's `@theme`.
//   - The literal value `initial` deletes the key rather than storing the string.
//   - Otherwise the entry is stored, keeping its position if the key already existed.
//
// The clear branch falls through rather than returning, which is upstream's behaviour and not an
// oversight: after `--color-*: initial` clears the namespace, the same call continues to the store
// below and would write the key `--color-*` itself if the value were not `initial`. It always is,
// because the guard above rejects anything else, so the fallthrough reaches the delete branch and
// the directive leaves no entry behind. Restructuring it into an early return gives the same
// answer today and diverges the moment upstream relaxes that guard.
//
// An invalid clear directive returns an error rather than panicking, since the input is a
// repository's stylesheet and a malformed one must surface as a diagnostic rather than take down
// the linter.
func (theme *Theme) Add(key, value string, options ThemeOptions) error {
	if strings.HasSuffix(key, "-*") {
		if value != "initial" {
			return &CSSSyntaxError{Message: "Invalid theme value `" + value + "` for namespace `" + key + "`"}
		}
		if key == "--*" {
			theme.clearAll()
		} else {
			// `--${key}-*: initial` clears *all* theme values in the namespace, so no option bits
			// are required of the entries it removes.
			theme.ClearNamespace(strings.TrimSuffix(key, "-*"), ThemeOptionNone)
		}
	}

	if options&ThemeOptionDefault != 0 {
		if existing, ok := theme.values[key]; ok && existing.options&ThemeOptionDefault == 0 {
			return nil
		}
	}

	if value == "initial" {
		theme.delete(key)
		return nil
	}

	if _, ok := theme.values[key]; !ok {
		theme.keyOrder = append(theme.keyOrder, key)
	}
	theme.values[key] = themeValue{value: value, options: options}
	return nil
}

// delete removes a key and marks its slot in keyOrder dead, so that readers skip it and
// compactKeyOrder can reclaim it.
//
// The slot is blanked rather than left holding the key name. A tombstone that still carried its key
// would be indistinguishable from a live slot the moment that key was re-added, and every ordered
// read would then yield the key twice: once at the dead position and once at the new one. The empty
// string is safe as the dead marker because a theme key always begins `--`, so no live key can
// collide with it.
func (theme *Theme) delete(key string) {
	if _, ok := theme.values[key]; !ok {
		return
	}
	delete(theme.values, key)

	for index := len(theme.keyOrder) - 1; index >= 0; index-- {
		if theme.keyOrder[index] == key {
			theme.keyOrder[index] = ""
			break
		}
	}

	theme.deadKeys++
	theme.compactKeyOrder()
}

// clearAll empties the theme, upstream's `this.values.clear()`.
func (theme *Theme) clearAll() {
	theme.values = make(map[string]themeValue)
	theme.keyOrder = nil
	theme.deadKeys = 0
}

// compactKeyOrder drops keys that are no longer present, once they outnumber the live ones.
//
// Without a bound, a stylesheet that redefines and clears repeatedly would grow keyOrder without
// limit; compacting on every delete would make clearing an n-key namespace quadratic. Compacting
// when the dead outnumber the live keeps keyOrder within twice the live size and the amortized cost
// of a delete constant.
func (theme *Theme) compactKeyOrder() {
	if theme.deadKeys <= len(theme.values) {
		return
	}
	compacted := make([]string, 0, len(theme.values))
	for _, key := range theme.keyOrder {
		if key == "" {
			continue
		}
		if _, ok := theme.values[key]; ok {
			compacted = append(compacted, key)
		}
	}
	theme.keyOrder = compacted
	theme.deadKeys = 0
}

// liveKeys iterates the present keys in insertion order.
//
// Every ordered read goes through here rather than ranging keyOrder directly, so that the
// possibility of a stale key is handled in exactly one place.
func (theme *Theme) liveKeys(visit func(key string, value themeValue) bool) {
	for _, key := range theme.keyOrder {
		if key == "" {
			continue
		}
		value, ok := theme.values[key]
		if !ok {
			continue
		}
		if !visit(key, value) {
			return
		}
	}
}

// ThemeEntry is one entry of Entries.
type ThemeEntry struct {
	// Key is the custom property name, prefixed if the theme carries a prefix.
	Key string
	// Value is the declared value.
	Value string
	// Options is the bitfield the entry was declared under.
	Options ThemeOptions
}

// Entries returns every entry in insertion order, upstream's `entries()`.
//
// Keys come back prefixed, matching upstream, which maps `prefixKey` over the pairs on the way out.
// The stored keys are unprefixed; see PrefixKey.
func (theme *Theme) Entries() []ThemeEntry {
	entries := make([]ThemeEntry, 0, len(theme.values))
	theme.liveKeys(func(key string, value themeValue) bool {
		entries = append(entries, ThemeEntry{
			Key:     theme.PrefixKey(key),
			Value:   value.value,
			Options: value.options,
		})
		return true
	})
	return entries
}

// KeysInNamespaces returns the suffix of every key belonging to one of the given namespaces, in
// insertion order, and is upstream's `keysInNamespaces`.
//
// Three filters, all of which change the count on a real theme:
//
//   - The key must start with `namespace-`. The namespace itself, exactly, is not included: the
//     key `--spacing` contributes nothing to the `--spacing` namespace here, though Namespace does
//     report it under a null key.
//   - A key containing `--` at or after index 2 is skipped, which drops every sub-variable.
//     `--text-sm--line-height` is not a key of the `--text` namespace even though it starts with
//     `--text-`.
//   - An ignored key is skipped. See isIgnoredThemeKey.
func (theme *Theme) KeysInNamespaces(themeKeys []string) []string {
	var keys []string
	for _, namespace := range themeKeys {
		prefix := namespace + "-"
		theme.liveKeys(func(key string, _ themeValue) bool {
			if !strings.HasPrefix(key, prefix) {
				return true
			}
			// Upstream is `key.indexOf('--', 2) !== -1`: the search starts past the key's own
			// leading `--` so that every key is not trivially its own sub-variable.
			if strings.Contains(key[2:], "--") {
				return true
			}
			if isIgnoredThemeKey(key, namespace) {
				return true
			}
			keys = append(keys, key[len(prefix):])
			return true
		})
	}
	return keys
}

// Get returns the value of the first key present, or false when none is, and is upstream's `get`.
//
// Unlike the Resolve family this takes fully-qualified keys rather than a namespace and a candidate
// value, and applies no ignored-key filtering: it is a direct lookup.
func (theme *Theme) Get(themeKeys []string) (string, bool) {
	for _, key := range themeKeys {
		if value, ok := theme.values[key]; ok {
			return value.value, true
		}
	}
	return "", false
}

// HasDefault reports whether the key was declared with `@theme default`.
func (theme *Theme) HasDefault(key string) bool {
	return theme.GetOptions(key)&ThemeOptionDefault == ThemeOptionDefault
}

// GetOptions returns the bitfield a key was declared under, or ThemeOptionNone when absent.
//
// The key is unprefixed and unescaped first, so this accepts the key as it appears in emitted CSS
// rather than as it is stored. That asymmetry is upstream's: Entries hands out prefixed keys, and
// this takes them back.
func (theme *Theme) GetOptions(key string) ThemeOptions {
	key = unescapeCSSIdentifier(theme.unprefixKey(key))
	if value, ok := theme.values[key]; ok {
		return value.options
	}
	return ThemeOptionNone
}

// PrefixKey applies the theme's prefix to a stored key, upstream's `prefixKey`.
func (theme *Theme) PrefixKey(key string) string {
	if theme.Prefix == "" {
		return key
	}
	return "--" + theme.Prefix + "-" + key[2:]
}

// unprefixKey removes the theme's prefix from a key, upstream's private `#unprefixKey`.
//
// Upstream slices at `3 + prefix.length` unconditionally, without checking that the key actually
// carries the prefix, and this does the same. On a key that lacks it the result is meaningless in
// both implementations; reproducing the behaviour keeps the port faithful where a defensive check
// would make it answer differently than the engine on malformed input.
func (theme *Theme) unprefixKey(key string) string {
	if theme.Prefix == "" {
		return key
	}
	cut := 3 + len(theme.Prefix)
	if cut > len(key) {
		return "--"
	}
	return "--" + key[cut:]
}

// ClearNamespace deletes every key under a namespace, and is upstream's `clearNamespace`.
//
// clearOptions is a filter rather than a mask to remove: when it is not ThemeOptionNone, only
// entries carrying *all* of its bits are deleted. Keys belonging to an ignored sub-namespace
// survive, which is why `--font-*: initial` leaves `--font-weight-*` standing.
//
// The namespace match here is a bare prefix, not a `-` boundary, matching upstream. That is
// observable: `--color-*: initial` also clears `--colorful-b`, because the directive strips the
// `-*` and tests `strings.HasPrefix(key, "--color")`. It looks like a bug and it is the engine's
// behaviour, so the port reproduces it rather than improving on it.
func (theme *Theme) ClearNamespace(namespace string, clearOptions ThemeOptions) {
	ignored := ignoredThemeKeys[namespace]

	var doomed []string
	theme.liveKeys(func(key string, value themeValue) bool {
		if !strings.HasPrefix(key, namespace) {
			return true
		}
		if clearOptions != ThemeOptionNone && value.options&clearOptions != clearOptions {
			return true
		}
		for _, ignoredNamespace := range ignored {
			if strings.HasPrefix(key, ignoredNamespace) {
				return true
			}
		}
		doomed = append(doomed, key)
		return true
	})

	// Collected first, then deleted, because liveKeys reads the map it would be mutating.
	for _, key := range doomed {
		theme.delete(key)
	}
}

// resolveKey finds the first theme key that a candidate value resolves to under the given
// namespaces, and is upstream's private `#resolveKey`.
//
// candidateValuePresent distinguishes "resolve `red-500` under `--color`" from "resolve the
// namespace `--color` itself", which upstream expresses as `string | null`. The two ask different
// questions and a port collapsing null onto the empty string would turn the second into a lookup
// for the key `--color-`.
//
// The dot-to-underscore retry is the subtle branch. A candidate value containing `.` is looked up
// again with every `.` replaced by `_`, because `--spacing-1_5` is how a theme spells a key for the
// candidate `1.5`. The retry runs only when the exact key missed and only when the value contains a
// dot, so it costs nothing on the overwhelming majority of lookups.
func (theme *Theme) resolveKey(candidateValue string, candidateValuePresent bool, themeKeys []string) (string, bool) {
	for _, namespace := range themeKeys {
		themeKey := namespace
		if candidateValuePresent {
			themeKey = namespace + "-" + candidateValue
		}

		if _, ok := theme.values[themeKey]; !ok {
			if candidateValuePresent && strings.Contains(candidateValue, ".") {
				themeKey = namespace + "-" + strings.ReplaceAll(candidateValue, ".", "_")
				if _, ok := theme.values[themeKey]; !ok {
					continue
				}
			} else {
				continue
			}
		}

		if isIgnoredThemeKey(themeKey, namespace) {
			continue
		}

		return themeKey, true
	}

	return "", false
}

// variableReference builds the `var(...)` a resolved key emits, and is upstream's private `#var`.
//
// A ThemeOptionReference value gets its own literal value as the `var()` fallback. Upstream's
// reasoning: a `@theme reference` block emits no CSS variable, so a stylesheet using `@apply` under
// `@reference` cannot assume the browser will have one, and without the fallback the declaration
// would resolve to nothing at runtime.
func (theme *Theme) variableReference(themeKey string) (string, bool) {
	value, ok := theme.values[themeKey]
	if !ok {
		return "", false
	}

	reference := "var(" + escapeCSSIdentifier(theme.PrefixKey(themeKey))
	if value.options&ThemeOptionReference != 0 && value.value != "" {
		reference += ", " + value.value
	}
	return reference + ")", true
}

// MarkUsedVariable sets ThemeOptionUsed on a key and reports whether this call is what set it.
//
// Nothing in cohere calls this. It is ported because the bit it sets is readable through
// GetOptions, so a Theme that silently lacked it would answer a different bitfield than the engine
// for any consumer that did mark usage, and because the `!isUsed` return is the kind of
// first-caller-wins signal that is easy to get backwards when written later from memory. The key is
// unprefixed and unescaped first, matching GetOptions.
func (theme *Theme) MarkUsedVariable(themeKey string) bool {
	key := unescapeCSSIdentifier(theme.unprefixKey(themeKey))
	value, ok := theme.values[key]
	if !ok {
		return false
	}
	wasUsed := value.options&ThemeOptionUsed != 0
	value.options |= ThemeOptionUsed
	theme.values[key] = value
	return !wasUsed
}

// Resolve returns what a candidate value resolves to as a CSS value, and is upstream's `resolve`.
//
// The result is a `var()` reference unless ThemeOptionInline is set, either on the entry or by the
// caller, in which case it is the literal value. Upstream ORs the two option sets before testing,
// so a caller passing inline gets the literal even for an entry that did not declare it.
//
// candidateValuePresent has the meaning documented on resolveKey.
func (theme *Theme) Resolve(candidateValue string, candidateValuePresent bool, themeKeys []string, options ThemeOptions) (string, bool) {
	themeKey, ok := theme.resolveKey(candidateValue, candidateValuePresent, themeKeys)
	if !ok {
		return "", false
	}

	value := theme.values[themeKey]
	if (options|value.options)&ThemeOptionInline != 0 {
		return value.value, true
	}

	return theme.variableReference(themeKey)
}

// ResolveValue returns the literal declared value a candidate resolves to, and is upstream's
// `resolveValue`. Unlike Resolve it never builds a `var()` and never consults the options.
//
// This is the lookup the `@utility` evaluator makes for `--value(--percentage-*, [*])`, and the
// false return is what makes utility arity value-dependent: a declaration whose `--value()` fails
// to resolve is dropped, so `fade-in-0` emits two declarations and `fade-in-1` emits one, because
// `0` is a key of `--percentage` in this repository's theme and `1` is not.
func (theme *Theme) ResolveValue(candidateValue string, candidateValuePresent bool, themeKeys []string) (string, bool) {
	themeKey, ok := theme.resolveKey(candidateValue, candidateValuePresent, themeKeys)
	if !ok {
		return "", false
	}
	return theme.values[themeKey].value, true
}

// ResolveWith resolves a candidate value and, alongside it, the sub-variables hanging off the same
// key. It is upstream's `resolveWith`.
//
// This is the `-*--nested` suffix form: `--text-sm` carries `--text-sm--line-height`, so `text-sm`
// emits both a font-size and a line-height. nestedKeys are suffixes including their leading `--`,
// and the returned map is keyed by those same suffixes. A nested key that does not exist is
// omitted rather than mapped to an empty value, which is what lets a caller distinguish "no
// line-height declared" from "line-height declared empty".
//
// Each nested value independently honours ThemeOptionInline, so a theme can inline a line-height
// while leaving the size a `var()`.
func (theme *Theme) ResolveWith(candidateValue string, themeKeys []string, nestedKeys []string) (string, map[string]string, bool) {
	themeKey, ok := theme.resolveKey(candidateValue, true, themeKeys)
	if !ok {
		return "", nil, false
	}

	extra := make(map[string]string, len(nestedKeys))
	for _, name := range nestedKeys {
		nestedKey := themeKey + name
		nestedValue, ok := theme.values[nestedKey]
		if !ok {
			continue
		}
		if nestedValue.options&ThemeOptionInline != 0 {
			extra[name] = nestedValue.value
			continue
		}
		if reference, ok := theme.variableReference(nestedKey); ok {
			extra[name] = reference
		}
	}

	value := theme.values[themeKey]
	if value.options&ThemeOptionInline != 0 {
		return value.value, extra, true
	}

	reference, _ := theme.variableReference(themeKey)
	return reference, extra, true
}

// NamespaceEntry is one entry of Namespace.
type NamespaceEntry struct {
	// Key is the part of the theme key after the namespace. It keeps a leading `--` for a
	// sub-variable, so `--text-sm--line-height` appears under `--text` as `sm--line-height`.
	Key string
	// KeyIsNull marks the entry for the namespace itself, which upstream keys as `null`.
	// The key `--spacing` appears in the `--spacing` namespace with this set and Key empty.
	KeyIsNull bool
	// Value is the declared value.
	Value string
}

// Namespace returns every entry under a namespace, in insertion order, and is upstream's
// `namespace`.
//
// This is the broader of the two namespace reads and the differences from KeysInNamespaces are all
// load-bearing:
//
//   - The namespace itself is included, under a null key rather than an empty one. `--spacing:
//     0.25rem` is reachable here and invisible to KeysInNamespaces.
//   - Sub-variables are included, keeping their `--` prefix.
//   - No ignored-key filtering happens at all, so `--font-weight-bold` does appear under `--font`
//     here while being correctly absent from KeysInNamespaces([`--font`]).
//
// The sub-variable branch tests `namespace + "--"` before the plain `namespace + "-"` prefix,
// because the second would match the first and slice one character too many.
func (theme *Theme) Namespace(namespace string) []NamespaceEntry {
	var entries []NamespaceEntry
	prefix := namespace + "-"

	theme.liveKeys(func(key string, value themeValue) bool {
		switch {
		case key == namespace:
			entries = append(entries, NamespaceEntry{KeyIsNull: true, Value: value.value})
		case strings.HasPrefix(key, prefix+"-"):
			// Preserve the `--` prefix for sub-variables, e.g. `--font-size-sm--line-height`.
			entries = append(entries, NamespaceEntry{Key: key[len(namespace):], Value: value.value})
		case strings.HasPrefix(key, prefix):
			entries = append(entries, NamespaceEntry{Key: key[len(prefix):], Value: value.value})
		}
		return true
	})

	return entries
}

// escapeCSSIdentifier serializes a string as a CSS identifier, and is a port of `CSS.escape` from
// `src/utils/escape.ts`.
//
// Ported rather than approximated because it runs on every `var()` this file builds, and the escape
// rules are not the ones a reasonable guess produces: a lone `-` is escaped, a leading digit
// becomes a hex escape with a trailing space rather than a backslash, and a digit in second
// position is escaped only when the first character is `-`.
//
// Upstream indexes UTF-16 code units. This indexes bytes, which agrees because the only branch
// reachable for a non-ASCII byte is the `>= 0x80` passthrough, and every byte of a multi-byte UTF-8
// sequence has its high bit set. So a multi-byte character is copied through byte by byte and
// reassembles unchanged, exactly as upstream copies it code unit by code unit.
func escapeCSSIdentifier(value string) string {
	if len(value) == 1 && value[0] == '-' {
		return "\\-"
	}

	var builder strings.Builder
	builder.Grow(len(value))

	for index := 0; index < len(value); index++ {
		character := value[index]

		if character == 0x00 {
			builder.WriteRune('�')
			continue
		}

		if (character >= 0x01 && character <= 0x1F) || character == 0x7F ||
			(index == 0 && character >= '0' && character <= '9') ||
			(index == 1 && character >= '0' && character <= '9' && value[0] == '-') {
			builder.WriteByte('\\')
			builder.WriteString(strconv.FormatUint(uint64(character), 16))
			builder.WriteByte(' ')
			continue
		}

		if character >= 0x80 || character == '-' || character == '_' ||
			(character >= '0' && character <= '9') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= 'a' && character <= 'z') {
			builder.WriteByte(character)
			continue
		}

		builder.WriteByte('\\')
		builder.WriteByte(character)
	}

	return builder.String()
}

// unescapeCSSIdentifier reverses escapeCSSIdentifier, and is a port of `unescape` from
// `src/utils/escape.ts`.
//
// Upstream is a regex, `/\\([\dA-Fa-f]{1,6}[\t\n\f\r ]?|[\S\s])/g`, and this is the same grammar as
// a scan: a backslash begins either up to six hex digits with one optional trailing whitespace
// character, or exactly one other character taken literally. A hex escape naming a NUL, a surrogate
// or a value past the Unicode maximum becomes U+FFFD.
//
// The one place bytes and code units genuinely differ is here, because a hex escape can name a
// character outside ASCII, so the decoded code point is written as a rune. The literal branch takes
// one *rune* rather than one byte for the same reason: upstream's `[\S\s]` matches a whole code
// unit, and taking a byte would split a multi-byte character after a backslash. Everything else
// scans bytes, since the hex digits and whitespace terminators are all ASCII.
func unescapeCSSIdentifier(escaped string) string {
	if !strings.Contains(escaped, "\\") {
		return escaped
	}

	var builder strings.Builder
	builder.Grow(len(escaped))

	for index := 0; index < len(escaped); {
		if escaped[index] != '\\' {
			builder.WriteByte(escaped[index])
			index++
			continue
		}

		// A trailing backslash matches nothing in the regex and is therefore left as-is.
		if index+1 >= len(escaped) {
			builder.WriteByte('\\')
			index++
			continue
		}

		digitCount := 0
		for digitCount < 6 && index+1+digitCount < len(escaped) && isHexDigit(escaped[index+1+digitCount]) {
			digitCount++
		}

		if digitCount == 0 {
			// `[\S\s]`: exactly one code unit, taken literally. Decoding a rune keeps a multi-byte
			// character whole.
			_, size := utf8.DecodeRuneInString(escaped[index+1:])
			builder.WriteString(escaped[index+1 : index+1+size])
			index += 1 + size
			continue
		}

		codePoint, err := strconv.ParseUint(escaped[index+1:index+1+digitCount], 16, 32)
		consumed := 1 + digitCount
		// One optional whitespace character terminates the escape and is consumed with it.
		if index+consumed < len(escaped) && isEscapeTerminator(escaped[index+consumed]) {
			consumed++
		}
		index += consumed

		if err != nil || codePoint == 0 || codePoint > 0x10FFFF || (codePoint >= 0xD800 && codePoint <= 0xDFFF) {
			builder.WriteRune('�')
			continue
		}
		builder.WriteRune(rune(codePoint))
	}

	return builder.String()
}

func isHexDigit(character byte) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'f') ||
		(character >= 'A' && character <= 'F')
}

// isEscapeTerminator reports whether a byte is one of the whitespace characters that may terminate
// a hex escape: upstream's `[\t\n\f\r ]`.
func isEscapeTerminator(character byte) bool {
	switch character {
	case '\t', '\n', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}
