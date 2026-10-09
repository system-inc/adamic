package boundaries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// DependenciesOptions is the subset of upstream's options this port decides, plus the element
// descriptors upstream reads from settings.
//
// # Why the elements live in the options
//
// Upstream reads its elements from `settings["boundaries/elements"]`, beside the rule rather than in
// it, and Base sets different elements in each of the two config blocks that turn the rule on.
// cohere's config has no per-plugin settings, and its overrides carry rules alone, so the only place
// a block-scoped set of elements can travel is inside the rule entry that block writes. That is the
// one deliberate difference from upstream's option schema, and it is an addition rather than a
// reinterpretation: every key upstream writes means here what it means there.
type DependenciesOptions struct {
	// Default decides a dependency no policy matched. Upstream treats anything but "allow" as
	// disallow, so an absent default disallows.
	Default string `json:"default"`

	// Message replaces the default message for every violation a policy does not word itself.
	Message string `json:"message"`

	// Policies are evaluated in order, and the last one to match decides, the way upstream does it.
	Policies []dependencyPolicy `json:"policies"`

	// Elements are upstream's `boundaries/elements` setting: which folders are which layer.
	Elements []elementDescriptor `json:"elements"`

	// CheckAllOrigins, CheckUnknownLocals and CheckInternals are upstream's three widening flags. Each
	// defaults to false, and false is the only value this port implements, so true is refused.
	CheckAllOrigins    *bool `json:"checkAllOrigins"`
	CheckUnknownLocals *bool `json:"checkUnknownLocals"`
	CheckInternals     *bool `json:"checkInternals"`

	// Root is the directory element patterns are relative to, upstream's `rootPath`, which defaults to
	// the directory ESLint runs in. Settled by the decoder from where the config was written, and
	// never read from the config itself.
	Root string `json:"-"`
}

// elementDescriptor is one entry of upstream's elements setting, in its default `folder` mode.
type elementDescriptor struct {
	Type    string    `json:"type"`
	Pattern stringSet `json:"pattern"`

	// Capture names the values a pattern's wildcards capture. Accepted and unused: captured values
	// only matter to a selector that matches on them or a message template that prints them, and this
	// port refuses both, so no decision here can depend on a capture.
	Capture []string `json:"capture"`

	// Mode is accepted only as `folder`, upstream's default. `file` and `full` decide membership
	// differently and are refused rather than read as folder.
	Mode string `json:"mode"`
}

// dependencyPolicy is one entry of `policies`: an outer from and to, narrowed by its allow or disallow.
type dependencyPolicy struct {
	From     entitySelectors     `json:"from"`
	To       entitySelectors     `json:"to"`
	Allow    dependencySelectors `json:"allow"`
	Disallow dependencySelectors `json:"disallow"`
	Message  string              `json:"message"`
}

// dependencySelector is an allow or disallow entry. Upstream's legacy string form is refused.
type dependencySelector struct {
	From entitySelectors `json:"from"`
	To   entitySelectors `json:"to"`
}

// entitySelector picks the element on one end of a dependency. Upstream also selects by `file` and
// `module`; both are refused.
type entitySelector struct {
	Element *elementSelector `json:"element"`
}

// elementSelector matches an element by its type.
type elementSelector struct {
	Type  string      `json:"type"`
	Types *typesQuery `json:"types"`
}

// typesQuery is upstream's array query over an element's types. Only `anyOf` is ported.
type typesQuery struct {
	AnyOf []string `json:"anyOf"`
}

// stringSet is upstream's string-or-array-of-strings.
type stringSet []string

func (set *stringSet) UnmarshalJSON(raw []byte) error {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		*set = stringSet{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return fmt.Errorf("expected a string or an array of strings, got %s", raw)
	}
	*set = many
	return nil
}

// entitySelectors is upstream's selector-or-array-of-selectors for `from` and `to`.
type entitySelectors []entitySelector

func (selectors *entitySelectors) UnmarshalJSON(raw []byte) error {
	return decodeOneOrMany(raw, (*[]entitySelector)(selectors), "an element selector such as {\"element\": {\"type\": \"api\"}}")
}

// dependencySelectors is upstream's selector-or-array-of-selectors for `allow` and `disallow`.
type dependencySelectors []dependencySelector

func (selectors *dependencySelectors) UnmarshalJSON(raw []byte) error {
	return decodeOneOrMany(raw, (*[]dependencySelector)(selectors), "a dependency selector such as {\"to\": {\"element\": {\"type\": \"api\"}}}")
}

// decodeOneOrMany reads one selector or an array of them, each through rule.UnmarshalOptions.
//
// Every level is decoded strictly here rather than through json.Unmarshal, because a custom
// UnmarshalJSON that calls json.Unmarshal drops the outer decoder's strictness for everything beneath
// it. A selector key this port does not implement, `captured` or `file` or `parent`, would then load
// clean and be ignored, which is a policy narrower than the one written with nothing saying so.
func decodeOneOrMany[Item any](raw []byte, into *[]Item, shape string) error {
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var many []json.RawMessage
		if err := json.Unmarshal(trimmed, &many); err != nil {
			return err
		}
		for _, element := range many {
			var item Item
			if err := rule.UnmarshalOptions(element, &item); err != nil {
				return fmt.Errorf("%w; this port reads %s, and upstream's legacy string selectors are not ported", err, shape)
			}
			*into = append(*into, item)
		}
		return nil
	}
	var item Item
	if err := rule.UnmarshalOptions(trimmed, &item); err != nil {
		return fmt.Errorf("%w; this port reads %s, and upstream's legacy string selectors are not ported", err, shape)
	}
	*into = append(*into, item)
	return nil
}

// decodeDependenciesOptions reads the options and refuses every shape this port does not decide.
//
// A refusal names the key. The alternative, loading a config whose policy this port only partly
// understands, would enforce a narrower boundary than the one written and say nothing about it.
func decodeDependenciesOptions(raw []byte, base rule.OptionsBase) (any, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("boundaries/dependencies needs its options: the elements that say which " +
			"folder is which layer, and the policies between them. With neither, no file is an element " +
			"and the rule would decide nothing")
	}

	var settings DependenciesOptions
	if err := rule.UnmarshalOptions(raw, &settings); err != nil {
		return nil, fmt.Errorf("boundaries/dependencies: %w", err)
	}

	if settings.Default != "" && settings.Default != "allow" && settings.Default != "disallow" {
		return nil, fmt.Errorf("boundaries/dependencies: default is %q, and upstream accepts only \"allow\" or \"disallow\"", settings.Default)
	}
	for name, flag := range map[string]*bool{
		"checkAllOrigins":    settings.CheckAllOrigins,
		"checkUnknownLocals": settings.CheckUnknownLocals,
		"checkInternals":     settings.CheckInternals,
	} {
		if flag != nil && *flag {
			return nil, fmt.Errorf("boundaries/dependencies: %s is true, and this port implements only its "+
				"default, false. Reading it as false would check fewer dependencies than the config asks for", name)
		}
	}

	if len(settings.Elements) == 0 {
		return nil, fmt.Errorf("boundaries/dependencies has no elements, so no file would belong to a layer " +
			"and the rule would report nothing. Write them under \"elements\", the shape of upstream's " +
			"boundaries/elements setting")
	}
	for index, descriptor := range settings.Elements {
		if descriptor.Type == "" {
			return nil, fmt.Errorf("boundaries/dependencies: element %d has no type", index)
		}
		if len(descriptor.Pattern) == 0 {
			return nil, fmt.Errorf("boundaries/dependencies: element %q has no pattern", descriptor.Type)
		}
		for _, pattern := range descriptor.Pattern {
			if strings.ContainsAny(pattern, "{}()") {
				return nil, fmt.Errorf("boundaries/dependencies: element %q has pattern %q, and this port matches "+
					"`*`, `?`, `[...]` and `**` but not micromatch's brace or extglob syntax", descriptor.Type, pattern)
			}
		}
		if descriptor.Mode != "" && descriptor.Mode != "folder" {
			return nil, fmt.Errorf("boundaries/dependencies: element %q has mode %q, and this port implements "+
				"only the default folder mode", descriptor.Type, descriptor.Mode)
		}
	}

	for index, policy := range settings.Policies {
		if len(policy.Allow) == 0 && len(policy.Disallow) == 0 {
			return nil, fmt.Errorf("boundaries/dependencies: policy %d has neither allow nor disallow, so it "+
				"could never decide a dependency", index)
		}
		if err := refuseTemplate(policy.Message); err != nil {
			return nil, fmt.Errorf("boundaries/dependencies: policy %d: %w", index, err)
		}
		for _, selectors := range []entitySelectors{policy.From, policy.To} {
			if err := checkEntitySelectors(selectors); err != nil {
				return nil, fmt.Errorf("boundaries/dependencies: policy %d: %w", index, err)
			}
		}
		for _, entries := range []dependencySelectors{policy.Allow, policy.Disallow} {
			for _, entry := range entries {
				for _, selectors := range []entitySelectors{entry.From, entry.To} {
					if err := checkEntitySelectors(selectors); err != nil {
						return nil, fmt.Errorf("boundaries/dependencies: policy %d: %w", index, err)
					}
				}
			}
		}
	}
	if err := refuseTemplate(settings.Message); err != nil {
		return nil, fmt.Errorf("boundaries/dependencies: %w", err)
	}

	// Anchored where the config was written, which is where upstream's rootPath sits for every project
	// this runs on: ESLint resolves it from the directory it runs in, and Base runs it from the
	// directory its config lives in.
	settings.Root = base.ConfigDirectory
	if settings.Root == "" {
		settings.Root = base.ProjectRoot
	}
	return settings, nil
}

// checkEntitySelectors refuses a selector that would match by something other than an exact type.
func checkEntitySelectors(selectors entitySelectors) error {
	for _, selector := range selectors {
		if selector.Element == nil {
			return fmt.Errorf("a selector with no element matches every entity upstream knows, by file and " +
				"module as well as element, and only element selectors are ported")
		}
		if selector.Element.Type == "" && selector.Element.Types == nil {
			return fmt.Errorf("an element selector with neither type nor types matches every element")
		}
		names := []string{selector.Element.Type}
		if selector.Element.Types != nil {
			if len(selector.Element.Types.AnyOf) == 0 {
				return fmt.Errorf("types has an empty anyOf, which upstream reads as matching nothing")
			}
			names = append(names, selector.Element.Types.AnyOf...)
		}
		for _, name := range names {
			if strings.ContainsAny(name, "*?[]{}!()+@") {
				return fmt.Errorf("element type %q is a micromatch pattern, and this port matches types "+
					"exactly", name)
			}
		}
	}
	return nil
}

// refuseTemplate refuses a message that upstream would render as a template.
func refuseTemplate(message string) error {
	if strings.Contains(message, "{{") || strings.Contains(message, "${") {
		return fmt.Errorf("the message %q is a template, and this port prints messages verbatim rather "+
			"than rendering upstream's handlebars or legacy placeholders", message)
	}
	return nil
}

// element is a file's place in the element layout: its type, and the folder that is the element.
type element struct {
	Type string
	Path string
}

// describeElement finds which element a root-relative file path belongs to, as @boundaries/elements
// does in folder mode with one type per element, upstream's default.
//
// The walk starts at the file's own name and grows leftwards one segment at a time. At each length,
// every descriptor is tried in order against `pattern/**/*`, and the first to match is the element.
// So the shortest matching suffix wins, which is why upstream documents that a pattern need not be
// written from the root: `source/*` also claims `libraries/base/source/foundation`, because that path
// ends in a suffix the pattern matches. Measured with the installed build rather than taken from its
// documentation, because Base's project boundary depends on it: under that block, Base's own
// foundation files are `project-source` elements and its modules are `project-module` ones.
func describeElement(relativePath string, descriptors []elementDescriptor) (element, bool) {
	segments := strings.Split(relativePath, "/")
	for start := len(segments) - 1; start >= 0; start-- {
		suffix := segments[start:]
		for _, descriptor := range descriptors {
			for _, pattern := range descriptor.Pattern {
				if !globMatches(pattern+"/**/*", suffix) {
					continue
				}
				return element{Type: descriptor.Type, Path: elementPath(pattern, suffix, segments)}, true
			}
		}
	}
	return element{}, false
}

// elementPath is the folder an element is, which is what decides whether two files share one.
//
// Upstream finds the shortest prefix of the matched suffix that the bare pattern matches, then cuts
// the full path at the first place that prefix appears as text. The cut is by text rather than by
// segment, and it is reproduced as text: a path that repeats the matched prefix earlier on would
// resolve to the earlier occurrence upstream, and so it does here.
func elementPath(pattern string, suffix []string, segments []string) string {
	full := strings.Join(segments, "/")
	for length := 1; length <= len(suffix); length++ {
		if !globMatches(pattern, suffix[:length]) {
			continue
		}
		matched := strings.Join(suffix[:length], "/")
		return full[:strings.Index(full, matched)] + matched
	}
	return full
}

// globMatches is micromatch's answer for a slash-separated path, with its default handling of dots.
//
// Element patterns go through micromatch with its defaults, where neither `*` nor `**` matches a
// segment that begins with a dot unless the pattern spells the dot. cohere's config glob matches dot
// segments (config files are matched with `dot: true`), so reusing it would put
// `workers/api/.wrangler/state.ts` inside the `workers/*` element, a membership upstream does not
// grant. It is also not imported for a second reason: a rule package stays a leaf of the build graph
// (`TestRulePackagesStayLeaves`), and the config package is not one.
func globMatches(pattern string, path []string) bool {
	return matchGlobSegments(strings.Split(strings.Trim(pattern, "/"), "/"), path)
}

func matchGlobSegments(pattern []string, segments []string) bool {
	if len(pattern) == 0 {
		return len(segments) == 0
	}
	if pattern[0] == "**" {
		if matchGlobSegments(pattern[1:], segments) {
			return true
		}
		if len(segments) == 0 || strings.HasPrefix(segments[0], ".") {
			return false
		}
		return matchGlobSegments(pattern, segments[1:])
	}
	if len(segments) == 0 {
		return false
	}
	if strings.HasPrefix(segments[0], ".") && !strings.HasPrefix(pattern[0], ".") {
		return false
	}
	// One segment, so `*` and `?` cannot cross a slash. Brace and extglob syntax, which path.Match
	// does not read, is refused when the options are decoded.
	if matched, err := path.Match(pattern[0], segments[0]); err != nil || !matched {
		return false
	}
	return matchGlobSegments(pattern[1:], segments[1:])
}

// matches reports whether one element satisfies an element selector.
func (selector *elementSelector) matches(subject element) bool {
	if selector.Type != "" && selector.Type != subject.Type {
		return false
	}
	if selector.Types != nil {
		found := false
		for _, name := range selector.Types.AnyOf {
			if name == subject.Type {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// mergeEntitySelectors composes a policy's outer selector with one entry's, as upstream does.
//
// Every outer selector is paired with every entry selector, and within a pair the entry's fields win
// over the outer's. Either side absent leaves the other as it is.
func mergeEntitySelectors(outer entitySelectors, entry entitySelectors) entitySelectors {
	if len(entry) == 0 {
		return outer
	}
	if len(outer) == 0 {
		return entry
	}
	merged := entitySelectors{}
	for _, outerSelector := range outer {
		for _, entrySelector := range entry {
			combined := *outerSelector.Element
			if entrySelector.Element.Type != "" {
				combined.Type = entrySelector.Element.Type
			}
			if entrySelector.Element.Types != nil {
				combined.Types = entrySelector.Element.Types
			}
			merged = append(merged, entitySelector{Element: &combined})
		}
	}
	return merged
}

// selectorsMatch reports whether any selector matches, and an absent side matches anything.
func selectorsMatch(selectors entitySelectors, subject element) bool {
	if len(selectors) == 0 {
		return true
	}
	for _, selector := range selectors {
		if selector.Element.matches(subject) {
			return true
		}
	}
	return false
}

// entryMatches reports whether one allow or disallow entry, composed with its policy, matches.
func (policy *dependencyPolicy) entryMatches(entry dependencySelector, from element, to element) bool {
	return selectorsMatch(mergeEntitySelectors(policy.From, entry.From), from) &&
		selectorsMatch(mergeEntitySelectors(policy.To, entry.To), to)
}

// decide evaluates every policy and reports whether the dependency is refused, and by which policy.
//
// Upstream's order, reproduced exactly: every policy is evaluated and the last one that matches
// decides. Within one policy, disallow is checked first and allow is not checked at all once
// disallow matched. A dependency no policy matched falls to the default. The returned index is -1
// when no policy matched.
func (settings DependenciesOptions) decide(from element, to element) (refused bool, policyIndex int) {
	allowed := false
	matched := -1
	for index := range settings.Policies {
		policy := &settings.Policies[index]
		disallowMatched := false
		for _, entry := range policy.Disallow {
			if policy.entryMatches(entry, from, to) {
				allowed, matched, disallowMatched = false, index, true
				break
			}
		}
		if disallowMatched {
			continue
		}
		for _, entry := range policy.Allow {
			if policy.entryMatches(entry, from, to) {
				allowed, matched = true, index
				break
			}
		}
	}
	if matched < 0 {
		return settings.Default != "allow", -1
	}
	return !allowed, matched
}

// messageFor words a refusal: the policy's own message, else the rule's, else one naming both layers.
func (settings DependenciesOptions) messageFor(policyIndex int, from element, to element) rule.Message {
	if policyIndex >= 0 && settings.Policies[policyIndex].Message != "" {
		return rule.Message{Id: "policyMessage", Description: settings.Policies[policyIndex].Message}
	}
	if settings.Message != "" {
		return rule.Message{Id: "policyMessage", Description: settings.Message}
	}
	if policyIndex < 0 {
		return rule.Message{
			Id: "noPolicyAllows",
			Description: "Nothing in this boundary lets a file in the \"" + from.Type + "\" layer import " +
				"from the \"" + to.Type + "\" layer, and the boundary refuses what it does not allow. A " +
				"layer that reaches a layer it was not given is how the direction the layering promises " +
				"stops being true. Move the shared code where both may reach it, or change the policy.",
		}
	}
	return rule.Message{
		Id: "policyDisallows",
		Description: "A policy in this boundary forbids a file in the \"" + from.Type + "\" layer from " +
			"importing the \"" + to.Type + "\" layer. Move the shared code where both may reach it, or " +
			"change the policy.",
	}
}

// Dependencies refuses an import that crosses from one element into another its policies do not allow.
//
//	elements: api = libraries/base/source/api/**, foundation = libraries/base/source/foundation/**
//	policies: from foundation allow to api; default disallow
//	valid:   foundation/base/Handler.ts imports '@base/source/api/rpc/client/RpcClient'
//	valid:   api/rpc/client/RpcClient.ts imports './driver/RpcClientDriver'      same element
//	valid:   api/rpc/client/RpcClient.ts imports 'zod'                           not a local file
//	valid:   api/rpc/client/RpcClient.ts imports '@nexus/source/time/DueDate'    no element
//	invalid: api/rpc/client/RpcClient.ts imports '@base/source/foundation/base/Base'
//
// # What is checked, all measured on the installed build
//
// A file that is no element is never checked, and neither is a dependency that resolves to a package,
// to nothing, to a local file in no element, or to a file in the importing file's own element. What
// is left is decided by the policies, last match winning, and then by the default.
//
// A value or type import, a side-effect import, `export * from`, `export { } from` and their
// type-only forms, and `import('…')` with a string argument are each a dependency, as upstream finds
// them. `import x = require('…')` and an `import('…')` type are not, because upstream's selectors do
// not reach them either. The finding lands on the specifier string, where upstream puts it.
//
// # The one divergence, and its size
//
// Upstream also reads `require('…')` called by that bare name. This port does not. The program
// resolves only the specifiers TypeScript collects, and in a TypeScript file it collects no `require`
// call, so the dependency would resolve to nothing; in a JavaScript file it does, but the typed test
// harness builds a TypeScript-only program, so a `require` branch could be written and never shown to
// work, and an unproven branch is not shipped. Measured on api-phi-health: both of Base's blocks lint
// only `.ts`, `.mts` and `.tsx`, and those files hold no `require` call by that bare name (the six
// `require(` matches are `request.require(key)`, which upstream's `callee.name=require` skips too),
// so the divergence reaches nothing there today.
var Dependencies = rule.Rule{
	Name: "boundaries/dependencies",
	// The project root, and where this file's imports resolve.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsModuleResolution,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil || ctx.Program == nil {
			return nil
		}
		settings, configured := rule.OptionsAs[DependenciesOptions](options)
		if !configured || len(settings.Elements) == 0 {
			// The decoder refuses an empty element list, so this is a hand-built Context's case, and a
			// skip rather than a silence if it is ever reached in a run (#pa7k7zv).
			ctx.Skip("no elements are configured")
			return nil
		}

		root := settings.Root
		if root == "" {
			// A test hands the struct straight in, with no config to anchor it. The program's
			// directory is where its tsconfig sits, which is upstream's rootPath for every project
			// this runs on.
			root = ctx.Program.GetCurrentDirectory()
		}
		// The resolver returns a package's file by its real path and a project file by the path it was
		// reached through, so under a root that is itself reached through a link the two disagree:
		// measured in the test harness, `./b` resolved under /tmp and a package under /private/tmp.
		// Measuring only against the root as written would put every package outside the project,
		// where it is skipped for the wrong reason and the external check is never asked.
		roots := []string{root}
		if realRoot, err := filepath.EvalSymlinks(root); err == nil && realRoot != root {
			roots = append(roots, realRoot)
		}
		relative := func(fileName string) (string, bool) {
			for _, candidate := range roots {
				path, err := filepath.Rel(candidate, fileName)
				// Slashed before the test rather than after it: on Windows Rel answers `..\x`, which the
				// `../` test let through, so every file outside the root was described as inside it.
				path = filepath.ToSlash(path)
				if err == nil && path != ".." && !strings.HasPrefix(path, "../") {
					return path, true
				}
			}
			return "", false
		}

		fromPath, inside := relative(ctx.SourceFile.FileName().AsString())
		if !inside {
			return nil
		}
		from, isElement := describeElement(fromPath, settings.Elements)
		if !isElement {
			return nil
		}

		check := func(specifier *ast.Node) {
			if specifier == nil || specifier.Kind != ast.KindStringLiteral {
				return
			}
			resolved := ctx.Program.ResolveModule(ctx.SourceFile, specifier)
			if !resolved.IsResolved() || resolved.IsExternalLibraryImport {
				return
			}
			toPath, inside := relative(resolved.ResolvedFileName)
			if !inside {
				return
			}
			to, isElement := describeElement(toPath, settings.Elements)
			if !isElement || to.Path == from.Path {
				return
			}
			if refused, policyIndex := settings.decide(from, to); refused {
				ctx.ReportNode(specifier, settings.messageFor(policyIndex, from, to))
			}
		}

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				check(node.AsImportDeclaration().ModuleSpecifier)
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				check(node.AsExportDeclaration().ModuleSpecifier)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !ast.IsImportCall(node) || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				check(call.Arguments.Nodes[0])
			},
		}
	},
}
