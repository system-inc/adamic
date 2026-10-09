package structure

import "strings"

// groupOrder is the canonical sequence of import groups, and it is the whole specification of the
// rule's ordering: a file is organized when its imports appear in this order, under these headers.
//
// Copied from `groups` in the original rather than reasoned about. The sequence is not alphabetical
// and not derivable; it is a reading order somebody chose, running from what the file is configured
// by, through what it borrows, to what it draws with.
var groupOrder = []string{
	"Project",
	"Node",
	"Frameworks",
	"Third-party",
	"Nexus",
	"Appearance",
	"Theme",
	"Localization",
	"Types",
	"Constants",
	"Providers",
	"Context",
	"Services",
	"Data",
	"Shared State",
	"APIs",
	"Hooks",
	"Layouts",
	"Components",
	"Local Components",
	"Animations",
	"Assets",
	"Utilities",
}

// frameworkExact and frameworkPrefixes are the packages that count as the framework itself.
//
// `next-intl` is deliberately absent and lands in Third-party, which was measured rather than
// assumed: the original's own header comment says "react, next, next-intl, etc.", and that comment
// is wrong about its own code. Only `next` exactly and `next/` as a prefix match.
var frameworkExact = map[string]bool{"react": true, "react-dom": true, "next": true}

var frameworkPrefixes = []string{"react/", "react-dom/", "next/"}

// nodeBuiltinModules are the bare specifiers that name a Node built-in without the `node:` prefix.
var nodeBuiltinModules = map[string]bool{
	"assert": true, "async_hooks": true, "buffer": true, "child_process": true, "cluster": true,
	"console": true, "constants": true, "crypto": true, "dgram": true, "diagnostics_channel": true,
	"dns": true, "domain": true, "events": true, "fs": true, "http": true, "http2": true,
	"https": true, "inspector": true, "module": true, "net": true, "os": true, "path": true,
	"perf_hooks": true, "process": true, "punycode": true, "querystring": true, "readline": true,
	"repl": true, "stream": true, "string_decoder": true, "sys": true, "timers": true, "tls": true,
	"trace_events": true, "tty": true, "url": true, "util": true, "v8": true, "vm": true,
	"wasi": true, "worker_threads": true, "zlib": true,
}

// animationPackages and assetPackages are the third-party libraries that get their own group.
var animationPackages = []string{"motion/", "motion"}

var assetPackages = []string{"@phosphor-icons/"}

// pathSegmentMatcher maps a directory name inside a first-party path to the group it names.
type pathSegmentMatcher struct {
	segment string
	group   string
}

// pathSegmentMatchers are checked in order, so an earlier entry wins over a later one on a path
// holding both. The order is the original's and it is observable: `/api/hooks/x` is APIs, not Hooks.
var pathSegmentMatchers = []pathSegmentMatcher{
	{"/appearance/", "Appearance"},
	{"/theme/", "Theme"},
	{"/localization/", "Localization"},
	{"/_localization/", "Localization"},
	{"/_translations", "Localization"},
	{"/locales/", "Localization"},
	{"/context/", "Context"},
	{"/services/", "Services"},
	{"/data/", "Data"},
	{"/shared-state/", "Shared State"},
	{"/api/", "APIs"},
	{"/hooks/", "Hooks"},
	{"/animations/", "Animations"},
	{"/components/", "Components"},
	{"/modules/", "Components"},
	{"/layouts/", "Layouts"},
	{"/pages/", "Components"},
	{"/utilities/", "Utilities"},
}

// specificGroups are the groups a path match is trusted for outright.
//
// The two groups NOT in this set are the ones a path can name generically: Components and Local
// Components. For those the specifier names get a second say, so `import { useThing } from
// '@structure/source/components/Button'` is Hooks rather than Components. Measured both ways.
var specificGroups = map[string]bool{
	"Project": true, "Node": true, "Frameworks": true, "Nexus": true, "Appearance": true,
	"Theme": true, "Localization": true, "Types": true, "Constants": true, "Providers": true,
	"Context": true, "Services": true, "Data": true, "Shared State": true, "APIs": true,
	"Hooks": true, "Layouts": true, "Animations": true, "Assets": true, "Utilities": true,
}

// firstPartyAliasPrefixes are the module aliases whose target is a path inside the project.
//
// The original discovers these from the nearest tsconfig at startup and falls back to this pair.
// **We use the pair directly, and that is fidelity rather than a shortcut.** Measured on the ahra
// tree: `TypeScriptConfiguration.jsonc` declares exactly `@structure/*`, `@nexus/*` and
// `@project/*`, and the original's reader filters `@nexus/` out on the way through because Nexus
// has its own group. So discovery on this tree returns `@structure/` and `@project/`, which is the
// fallback, and the two paths agree.
//
// Reproducing the filesystem walk would be the workaround the brief names: the original reads
// `process.cwd()` because ESLint hands it one file and no project. If a consuming project ever adds
// a fourth alias this needs the tsconfig, and the tell will be first-party imports classifying as
// Third-party.
var firstPartyAliasPrefixes = []string{"@structure/", "@project/"}

func isFirstPartyAliasSource(source string) bool {
	for _, prefix := range firstPartyAliasPrefixes {
		if strings.HasPrefix(source, prefix) {
			return true
		}
	}
	return false
}

// isHookSpecifierName reports a name shaped like a React hook: `use` then an ASCII capital.
//
// **Deliberately not `react.IsHookName`, and the difference is measurable.** The shelf function
// tests the fourth rune with `unicode.IsUpper`, and the original here is the regex `/^use[A-Z]/`,
// which is ASCII-only. Probed on the live rule: `import { useEthing } from './Thing'` spelled with
// an accented capital E classifies as Local Components upstream, and the shelf predicate would
// answer true and move it to Hooks. That is one import in a group it does not belong in, silently.
func isHookSpecifierName(name string) bool {
	if !strings.HasPrefix(name, "use") || len(name) == 3 {
		return false
	}
	character := name[3]
	return character >= 'A' && character <= 'Z'
}

// isProviderSpecifierName reports a name ending in `Provider` or `Providers`.
//
// The original's regex is `/Providers?$/`, which is unanchored at the front, so `MyProviders`
// matches and so does the bare word. Probed: all three of `Provider`, `Providers` and `MyProviders`
// classify as Providers.
func isProviderSpecifierName(name string) bool {
	return strings.HasSuffix(name, "Provider") || strings.HasSuffix(name, "Providers")
}

// matchesProviderPath reports the original's `/Providers?(?:\.|\b)/` against a path.
//
// The alternation is `.` or a word boundary, and after `Provider` or `Providers` every character
// that is not a word character is a boundary, as is the end of the string. So the pattern reduces to
// "contains `Provider` not immediately followed by a word character", with the one exception that a
// following `s` is absorbed by the optional plural. `Providers` matches, `ProviderThing` does not.
func matchesProviderPath(source string) bool {
	return containsWordEndingAt(source, "Provider")
}

// matchesConstantsPath reports the original's `/(?:Constants|Config|Enums)(?:\.|\b)/`.
func matchesConstantsPath(source string) bool {
	return containsWordEndingAt(source, "Constants") ||
		containsWordEndingAt(source, "Config") ||
		containsWordEndingAt(source, "Enums")
}

// containsWordEndingAt reports whether needle appears in source with a word boundary after it.
//
// This is the shared shape of the two regexes above. JavaScript's `\b` sits between a word character
// (letter, digit, underscore) and anything else, and `Providers?` already consumes a trailing `s`,
// so the question at each match site is whether the next character is a word character. The optional
// plural is handled by the caller passing the singular and this accepting one `s` after it.
func containsWordEndingAt(source string, needle string) bool {
	for index := 0; index+len(needle) <= len(source); index++ {
		if source[index:index+len(needle)] != needle {
			continue
		}
		after := index + len(needle)
		// The optional plural in `Providers?`. `Config` and `Enums` have no plural in the original's
		// pattern, so passing them through here is harmless: an `s` following them would be a word
		// character and decline anyway, which is the same answer.
		if needle == "Provider" && after < len(source) && source[after] == 's' {
			after++
		}
		if after >= len(source) {
			return true
		}
		if !isWordCharacter(source[after]) {
			return true
		}
	}
	return false
}

func isWordCharacter(character byte) bool {
	return character == '_' ||
		(character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9')
}

// containsHookPath reports the original's `/\/use[A-Z]/`: a path segment starting a hook name.
func containsHookPath(source string) bool {
	for index := 0; index+4 < len(source)+1; index++ {
		if index+4 > len(source) {
			break
		}
		if source[index] != '/' {
			continue
		}
		if index+4 <= len(source) && source[index+1:index+4] == "use" {
			if index+4 < len(source) && source[index+4] >= 'A' && source[index+4] <= 'Z' {
				return true
			}
		}
	}
	return false
}

// classifyImport names the group one import belongs to.
//
// The order of the tests below is the original's and every branch of it was pinned against the live
// rule. It is not a decision tree anybody would derive twice the same way, so it is transcribed
// rather than restructured: `@nexus/` is answered before the relative split, relative paths get
// their own smaller ladder, and the first-party ladder puts type-only above every path segment,
// which is why `import type` from a `/components/` path is Types rather than Components.
func classifyImport(source string, isTypeOnly bool, localNames []string) string {
	// Nexus wins over everything, including the relative split below it. A relative path that
	// happens to run through a `/nexus/` directory is a Nexus import.
	if strings.HasPrefix(source, "@nexus/") || strings.Contains(source, "/nexus/") {
		return "Nexus"
	}

	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		return classifyRelativeImport(source, isTypeOnly, localNames)
	}

	if strings.HasPrefix(source, "node:") {
		return "Node"
	}
	baseSpecifier := source
	if slash := strings.Index(source, "/"); slash >= 0 {
		baseSpecifier = source[:slash]
	}
	if nodeBuiltinModules[baseSpecifier] {
		return "Node"
	}

	if frameworkExact[source] {
		return "Frameworks"
	}
	for _, prefix := range frameworkPrefixes {
		if strings.HasPrefix(source, prefix) {
			return "Frameworks"
		}
	}

	if strings.HasPrefix(source, "@structure/assets/") || strings.HasSuffix(source, ".svg") {
		return "Assets"
	}
	for _, packagePrefix := range assetPackages {
		if strings.HasPrefix(source, packagePrefix) {
			return "Assets"
		}
	}
	for _, packagePrefix := range animationPackages {
		if source == packagePrefix || strings.HasPrefix(source, packagePrefix) {
			return "Animations"
		}
	}

	if isFirstPartyAliasSource(source) {
		return classifyFirstPartyImport(source, isTypeOnly, localNames)
	}

	if isTypeOnly {
		return "Types"
	}
	// A third-party package whose specifiers are hooks is Hooks, which is how a query library lands
	// beside the project's own hooks rather than in the middle of the third-party block.
	if anyHookName(localNames) {
		return "Hooks"
	}
	return "Third-party"
}

// classifyRelativeImport is the ladder for `./` and `../` paths.
//
// Note what is NOT here that is in the first-party ladder: there is no `/components/` segment and no
// `/app/` fallback, so a relative path through a components directory falls all the way to Local
// Components. And note the `Context` test is a bare substring rather than a path segment, so
// `./MyContext` is Context while `@structure/source/MyContext` is not.
func classifyRelativeImport(source string, isTypeOnly bool, localNames []string) string {
	if strings.Contains(source, "/utilities/") {
		return "Utilities"
	}
	if strings.Contains(source, "/api/") {
		return "APIs"
	}
	if strings.Contains(source, "/hooks/") || containsHookPath(source) {
		return "Hooks"
	}
	if strings.Contains(source, "/data/") {
		return "Data"
	}
	if strings.Contains(source, "/appearance/") {
		return "Appearance"
	}
	if strings.Contains(source, "/theme/") {
		return "Theme"
	}
	if strings.Contains(source, "/localization/") || strings.Contains(source, "/_localization/") ||
		strings.Contains(source, "/_translations") || strings.Contains(source, "/locales/") {
		return "Localization"
	}
	if strings.Contains(source, "/layouts/") {
		return "Layouts"
	}
	if strings.Contains(source, "/context/") || strings.Contains(source, "Context") {
		return "Context"
	}
	if strings.Contains(source, "/services/") {
		return "Services"
	}
	if strings.Contains(source, "/shared-state/") {
		return "Shared State"
	}
	if strings.Contains(source, "/animations/") {
		return "Animations"
	}
	if strings.HasSuffix(source, ".svg") {
		return "Assets"
	}
	if anyHookName(localNames) {
		return "Hooks"
	}
	if anyProviderName(localNames) || matchesProviderPath(source) {
		return "Providers"
	}
	// Type-only sits BELOW the hook and provider tests here and ABOVE every path segment in the
	// first-party ladder. That asymmetry is the original's and it is observable: a type-only
	// relative import of a hook is Hooks, while a type-only first-party import of one is Types.
	if isTypeOnly {
		return "Types"
	}
	if matchesConstantsPath(source) {
		return "Constants"
	}
	return "Local Components"
}

// classifyFirstPartyImport is the ladder for `@structure/` and `@project/` paths.
func classifyFirstPartyImport(source string, isTypeOnly bool, localNames []string) string {
	if isTypeOnly {
		return "Types"
	}
	if matchesConstantsPath(source) {
		return "Constants"
	}
	if anyProviderName(localNames) || matchesProviderPath(source) {
		return "Providers"
	}

	for _, matcher := range pathSegmentMatchers {
		if !strings.Contains(source, matcher.segment) {
			continue
		}
		if specificGroups[matcher.group] {
			return matcher.group
		}
		// Only Components reaches here, since it is the one matched group outside specificGroups.
		if anyHookName(localNames) {
			return "Hooks"
		}
		return matcher.group
	}

	// **This whole branch is subsumed and no fixture can catch a mutation that removes it.**
	// Kept because it is the original's, and because reducing it would make this file disagree with
	// the source it ports for a reader diffing the two.
	//
	// Scored as a survivor and then resolved by enumeration rather than by adding a test. Three
	// routes leave this branch and every one of them is already answered below it. The hook arm
	// matches the hook fallback. The provider arm is unreachable outright: the provider guard above
	// returns before anything reaches here, so `anyProviderName` is false by the time it is asked.
	// The Components arm matches the final return, and the one input the two could differ on is
	// `@project/ProjectSettings`, which the branch would call Components while the fallthrough calls
	// Project. That input cannot arrive: it holds no `/app/` segment and does not end in `/app`, so
	// `isApp` is false for it and the branch declines. Enumerated over every reachable combination
	// and the two spellings disagreed on nothing.
	if strings.Contains(source, "/app/") || strings.HasSuffix(source, "/app") {
		if anyHookName(localNames) {
			return "Hooks"
		}
		if anyProviderName(localNames) {
			return "Providers"
		}
		return "Components"
	}

	if source == "@project/ProjectSettings" {
		return "Project"
	}

	if anyHookName(localNames) {
		return "Hooks"
	}
	return "Components"
}

func anyHookName(localNames []string) bool {
	for _, name := range localNames {
		if isHookSpecifierName(name) {
			return true
		}
	}
	return false
}

func anyProviderName(localNames []string) bool {
	for _, name := range localNames {
		if isProviderSpecifierName(name) {
			return true
		}
	}
	return false
}
