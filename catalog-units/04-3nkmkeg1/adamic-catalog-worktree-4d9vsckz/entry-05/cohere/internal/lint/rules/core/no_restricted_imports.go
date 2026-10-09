package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoRestrictedImportsPath is one entry in the `paths` list: a module named exactly.
//
// Upstream's schema accepts a bare string here as shorthand for `{name: "<string>"}`, and the
// decoder below expands that rather than the rule carrying two shapes.
type NoRestrictedImportsPath struct {
	// Name is the module specifier, compared for EXACT equality after trimming. Not a pattern:
	// `paths` is the literal-name list and `patterns` is the pattern list.
	Name string `json:"name"`

	// Message is appended to the finding, and is where a project says what to import instead.
	Message string `json:"message"`

	// ImportNames restricts only these bindings from the module, leaving the rest importable.
	ImportNames []string `json:"importNames"`

	// AllowImportNames is the inverse: only these bindings may be imported, and everything else
	// reports. Upstream's schema forbids pairing it with ImportNames.
	AllowImportNames []string `json:"allowImportNames"`

	// AllowTypeImports exempts `import type` from this entry.
	AllowTypeImports bool `json:"allowTypeImports"`
}

// NoRestrictedImportsPattern is one entry in the `patterns` list: a module matched by shape.
//
// Exactly one of Group and Regex is set, which upstream expresses as a schema `oneOf` and the
// decoder here checks directly.
type NoRestrictedImportsPattern struct {
	// Group is a list of gitignore-style patterns, matched in order with the last match winning so
	// a `!` entry can undo an earlier one. See noRestrictedImportsCompileGroup.
	Group []string `json:"group"`

	// Regex is a JavaScript regular expression matched against the whole specifier.
	Regex string `json:"regex"`

	// Message is appended to the finding.
	Message string `json:"message"`

	// CaseSensitive makes both Group and Regex case-sensitive. Off by default, which is the
	// opposite of what a reader usually assumes about a path comparison and is upstream's choice.
	CaseSensitive bool `json:"caseSensitive"`

	// ImportNames restricts only these bindings from a matched module.
	ImportNames []string `json:"importNames"`

	// ImportNamePattern restricts every binding whose name matches this regular expression.
	ImportNamePattern string `json:"importNamePattern"`

	// AllowImportNames permits only these bindings from a matched module.
	AllowImportNames []string `json:"allowImportNames"`

	// AllowImportNamePattern permits only bindings matching this regular expression.
	AllowImportNamePattern string `json:"allowImportNamePattern"`

	// AllowTypeImports exempts `import type` from this entry.
	AllowTypeImports bool `json:"allowTypeImports"`
}

// NoRestrictedImportsOptions is the rule's whole configuration.
//
// Upstream's options are polymorphic at the top level: either a bare array of paths (strings or
// objects), or a one-element array holding `{paths, patterns}`. Both collapse to this, and the
// decoder is where that happens.
type NoRestrictedImportsOptions struct {
	// Paths restricts modules named exactly.
	Paths []NoRestrictedImportsPath `json:"paths"`

	// Patterns restricts modules matched by glob or regular expression.
	Patterns []NoRestrictedImportsPattern `json:"patterns"`
}

// noRestrictedImportsRawPattern is the wire shape of a `patterns` entry.
//
// Separate from the resolved type because a `patterns` value may also be a bare array of STRINGS,
// which upstream folds into a single entry with all of them as its group.
type noRestrictedImportsRawPattern = NoRestrictedImportsPattern

// DecodeNoRestrictedImportsOptions reads upstream's option list off the config.
//
// Hand-rolled because upstream's schema is an `anyOf` of two LISTS that no single struct can accept,
// and because three of upstream's schema constraints are cross-field rules with no struct-tag
// spelling. The rule registers with `DecodeOptionList`, so the list arrives whole:
//
//	["foo", {name: "bar", ...}]           every element a restricted path, a string or an object
//	[{paths: [...], patterns: [...]}]     exactly one element, the object form
//	[{patterns: ["foo/*", "bar"]}]        patterns as bare strings, folded into one group entry
//
// The object form is recognised the way upstream recognises it, by the first element being an
// object with a `paths` or `patterns` key. It takes no second element and no other key, so either is
// refused rather than dropped.
//
// The config layer used to keep only the first element after the severity, so upstream's variadic
// `["error", "foo", "bar"]` arrived here as the bare string "foo" and restricted one module of two.
func DecodeNoRestrictedImportsOptions(list []byte) (any, error) {
	var options NoRestrictedImportsOptions
	if len(list) == 0 {
		return options, nil
	}

	var elements []json.RawMessage
	if err := json.Unmarshal(list, &elements); err != nil {
		return options, fmt.Errorf("no-restricted-imports: the option list is not a JSON array: %w", err)
	}
	if len(elements) == 0 {
		return options, nil
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(elements[0], &probe); err == nil {
		// `[{}]`, the object form with nothing in it, restricts nothing. Upstream's schema accepts it
		// through the object branch, since a path object requires a name, and `[{}, "foo"]` fails
		// both branches. Kept divergence, ruled on #e06zm4b (#d21war2): upstream's runtime finds no
		// `paths` or `patterns` key, reads `{}` as a path entry, and keys it by its missing name, so
		// ESLint 10.8.1's Linter reports `import b from "undefined"` under `[{}]` and nothing else.
		// That is a missing key read as the name "undefined", not a restriction anyone wrote.
		if len(probe) == 0 && len(elements) == 1 {
			return options, nil
		}
		_, hasPaths := probe["paths"]
		_, hasPatterns := probe["patterns"]
		if hasPaths || hasPatterns {
			if len(elements) > 1 {
				return options, fmt.Errorf(
					"no-restricted-imports: the {paths, patterns} object takes no second element, "+
						"so %s would never be read", elements[1])
			}
			for key := range probe {
				if key != "paths" && key != "patterns" {
					return options, fmt.Errorf(
						"no-restricted-imports: the {paths, patterns} object has no key %q", key)
				}
			}
			return noRestrictedImportsDecodeObjectForm(probe)
		}
	}

	// Every element is a restricted path, a string or a path object.
	paths, err := noRestrictedImportsDecodePaths(list)
	if err != nil {
		return options, err
	}
	options.Paths = paths
	return options, noRestrictedImportsValidate(&options)
}

// noRestrictedImportsDecodeObjectForm reads the `{paths, patterns}` shape.
func noRestrictedImportsDecodeObjectForm(
	probe map[string]json.RawMessage,
) (NoRestrictedImportsOptions, error) {
	var options NoRestrictedImportsOptions

	if rawPaths, present := probe["paths"]; present {
		paths, err := noRestrictedImportsDecodePaths(rawPaths)
		if err != nil {
			return options, err
		}
		options.Paths = paths
	}

	if rawPatterns, present := probe["patterns"]; present {
		// `patterns` may be a bare array of strings, which upstream folds into ONE entry whose
		// group is all of them. That folding is observable rather than cosmetic: a single group
		// evaluates its `!` entries against each other, where separate entries would not.
		if string(rawPatterns) == "null" {
			rawPatterns = nil
		}
		var asStrings []string
		if len(rawPatterns) == 0 {
			asStrings = nil
		} else if err := json.Unmarshal(rawPatterns, &asStrings); err == nil {
			if len(asStrings) > 0 {
				options.Patterns = []NoRestrictedImportsPattern{{Group: asStrings}}
			}
		} else {
			var asObjects []noRestrictedImportsRawPattern
			if err := rule.UnmarshalOptions(rawPatterns, &asObjects); err != nil {
				return options, err
			}
			options.Patterns = asObjects
		}
	}

	return options, noRestrictedImportsValidate(&options)
}

// noRestrictedImportsDecodePaths reads a bare string, or an array whose members are strings or
// objects.
func noRestrictedImportsDecodePaths(raw json.RawMessage) ([]NoRestrictedImportsPath, error) {
	// A JSON `null` unmarshals into a string as `""` without error, so a key written but left null
	// would otherwise decode as one path entry with an empty name. That is not hypothetical: it is
	// what Go's own encoder produces for a nil slice, so every configuration round-tripped through
	// this struct hits it.
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []NoRestrictedImportsPath{{Name: single}}, nil
	}

	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}

	paths := make([]NoRestrictedImportsPath, 0, len(entries))
	for _, entry := range entries {
		var name string
		if err := json.Unmarshal(entry, &name); err == nil {
			paths = append(paths, NoRestrictedImportsPath{Name: name})
			continue
		}
		var object NoRestrictedImportsPath
		if err := rule.UnmarshalOptions(entry, &object); err != nil {
			return nil, err
		}
		paths = append(paths, object)
	}
	return paths, nil
}

// noRestrictedImportsValidate enforces the schema constraints upstream expresses declaratively.
//
// Every one of these would otherwise fail SILENTLY. A path with no name matches nothing and reads as
// a clean tree; a pattern with neither a group nor a regex likewise; and a pattern pairing an
// allow-list with a restrict-list asks the rule to both permit and forbid the same name, where the
// answer depends on which branch is reached first rather than on anything the author decided.
func noRestrictedImportsValidate(options *NoRestrictedImportsOptions) error {
	for index, path := range options.Paths {
		if path.Name == "" {
			return fmt.Errorf(
				"no-restricted-imports: paths entry %d has no name, so it can never match a module",
				index)
		}
		if len(path.ImportNames) > 0 && len(path.AllowImportNames) > 0 {
			return fmt.Errorf(
				"no-restricted-imports: paths entry %d pairs importNames with allowImportNames, "+
					"which restricts some names and permits only some others at the same time",
				index)
		}
	}

	for index, pattern := range options.Patterns {
		hasGroup := len(pattern.Group) > 0
		hasRegex := pattern.Regex != ""
		if hasGroup == hasRegex {
			return fmt.Errorf(
				"no-restricted-imports: patterns entry %d must carry exactly one of group and "+
					"regex, and carries %s", index,
				map[bool]string{true: "both", false: "neither"}[hasGroup])
		}
		// Upstream's schema forbids each of these five pairings for the same reason: each names a
		// restrict-list and an allow-list for the same question.
		if len(pattern.ImportNames) > 0 && len(pattern.AllowImportNames) > 0 {
			return fmt.Errorf("no-restricted-imports: patterns entry %d pairs importNames with "+
				"allowImportNames", index)
		}
		if pattern.ImportNamePattern != "" && pattern.AllowImportNamePattern != "" {
			return fmt.Errorf("no-restricted-imports: patterns entry %d pairs importNamePattern "+
				"with allowImportNamePattern", index)
		}
		if len(pattern.ImportNames) > 0 && pattern.AllowImportNamePattern != "" {
			return fmt.Errorf("no-restricted-imports: patterns entry %d pairs importNames with "+
				"allowImportNamePattern", index)
		}
		if pattern.ImportNamePattern != "" && len(pattern.AllowImportNames) > 0 {
			return fmt.Errorf("no-restricted-imports: patterns entry %d pairs importNamePattern "+
				"with allowImportNames", index)
		}
		if len(pattern.AllowImportNames) > 0 && pattern.AllowImportNamePattern != "" {
			return fmt.Errorf("no-restricted-imports: patterns entry %d pairs allowImportNames "+
				"with allowImportNamePattern", index)
		}

		// Every regular expression is compiled here so a misspelling fails at configuration time
		// rather than matching nothing at lint time.
		flags := "iu"
		if pattern.CaseSensitive {
			flags = "u"
		}
		if hasRegex {
			if _, err := esregexp.Compile(pattern.Regex, flags); err != nil {
				return fmt.Errorf("no-restricted-imports: patterns entry %d has an invalid regex "+
					"%q: %w", index, pattern.Regex, err)
			}
		}
		for _, source := range []string{pattern.ImportNamePattern, pattern.AllowImportNamePattern} {
			if source == "" {
				continue
			}
			if _, err := esregexp.Compile(source, "u"); err != nil {
				return fmt.Errorf("no-restricted-imports: patterns entry %d has an invalid import "+
					"name pattern %q: %w", index, source, err)
			}
		}
		if hasGroup {
			if _, ok := noRestrictedImportsCompileGroup(pattern.Group, pattern.CaseSensitive); !ok {
				return fmt.Errorf("no-restricted-imports: patterns entry %d has a group this "+
					"matcher cannot compile: %v", index, pattern.Group)
			}
		}
	}
	return nil
}

// NoRestrictedImports reports an import of a module a project has banned.
//
//	valid:   import os from 'os';                    with nothing configured
//	valid:   import { b } from 'mod';                with paths [{name:'mod', importNames:['a']}]
//	invalid: import fs from 'fs';                    with paths ['fs']
//	invalid: import { a } from 'mod';                with paths [{name:'mod', importNames:['a']}]
//	invalid: import { a } from 'foo/bar';            with patterns ['foo/*']
//	invalid: export { a } from 'fs';                 a re-export is an import
//	invalid: import * as ns from 'mod';              with any importNames restriction on 'mod'
//
// # This rule enforces nothing until somebody configures it
//
// With no `paths` and no `patterns` upstream returns an empty visitor object and the rule never runs.
// That is worth saying plainly because it makes an audit's violation count structurally zero rather
// than evidence of a clean tree, which is exactly what the audit for this rule recorded. Measured
// through the installed ESLint 10.8.1: `import fs from 'fs';` with no options reports nothing.
//
// So it is registered here and NOT enabled. Enabling it means choosing which modules a project bans,
// which is a decision about this codebase rather than a porting question.
//
// # A namespace import is reported once for the whole module, at the `*`
//
// `import * as ns from 'mod'` binds every export at once, so a restriction naming any import name at
// all applies to it and there is no per-name finding to give. Upstream reports at the SPECIFIER's
// location under the `everything` family of messages, and the corpus pins that: `import
// AllowedObject, * as AllowedObjectTwo from "foo"` reports at column 23, the `*`, not at column 1.
//
// `export * from 'foo'` is the same shape with no specifier node at all. Upstream synthesises one
// from `sourceCode.getFirstToken(node, 1)`, the token after `export`, which is the `*`. Reproduced
// here by reporting the `*` token's range, measured at column 8 across ten corpus cases.
//
// # A re-export is an import for this rule's purposes, and a type-only one is still a module read
//
// `export { a } from 'mod'` and `export * from 'mod'` both name a module the file depends on, so
// both go through the same check as an `import`. `export const a = 1` names no module and is not
// visited at all, which is why the listener is on the export declaration's MODULE SPECIFIER being
// present rather than on the statement.
//
// `allowTypeImports` is the one option that reads the type-only marker, and the marker lives in two
// places here rather than one. `import type { A } from 'm'` sets `PhaseModifier` on the import
// CLAUSE to `KindTypeKeyword`; `import { type A } from 'm'` sets `IsTypeOnly` on the SPECIFIER.
// Upstream's `isTypeOnlyImport` is the union of both -- the declaration is type-only, OR it has
// specifiers and every one of them is. Measured directly rather than read off the field names,
// because an earlier reading looked for `IsTypeOnly` on the clause, where this version of the
// compiler does not put it.
//
// # `import foo = require('bar')` is checked, and upstream's corpus does not reach it
//
// Upstream added a `TSImportEqualsDeclaration` arm for this TypeScript form and its own corpus has
// no case for it, so this arm is ported from the source rather than from a measured verdict, and its
// fixtures are written rather than imported. That distinction is stated here because everything else
// in this rule's tests came out of upstream's corpus and was replayed against the installed build.
var NoRestrictedImports = rule.Rule{
	Name: "no-restricted-imports",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, configured := rule.OptionsAs[NoRestrictedImportsOptions](options)
		if !configured || (len(settings.Paths) == 0 && len(settings.Patterns) == 0) {
			// Upstream's `if (Object.keys(restrictedPaths).length === 0 &&
			// restrictedPatternGroups.length === 0) return {}`.
			return rule.Listeners{}
		}

		checker, ok := newNoRestrictedImportsChecker(ctx, settings)
		if !ok {
			// A configuration that reached the rule without passing the decoder. Reporting nothing
			// is the only honest answer, and it cannot happen through the config layer.
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindImportDeclaration:       checker.checkImportDeclaration,
			ast.KindExportDeclaration:       checker.checkExportDeclaration,
			ast.KindImportEqualsDeclaration: checker.checkImportEquals,
		}
	},
}

// noRestrictedImportsCompiledPattern is one `patterns` entry with its regular expressions built.
type noRestrictedImportsCompiledPattern struct {
	entry NoRestrictedImportsPattern

	// matcher is set for a `group` entry, regularExpression for a `regex` one. Exactly one is
	// non-nil, which the decoder guarantees.
	matcher            *noRestrictedImportsMatcher
	regularExpression  *esregexp.RegExp
	importNamePattern  *esregexp.RegExp
	allowedNamePattern *esregexp.RegExp
}

// noRestrictedImportsChecker carries the compiled configuration across the listeners.
type noRestrictedImportsChecker struct {
	ctx rule.Context

	// paths is grouped by module name, because upstream groups them and a module named twice
	// reports twice. `import { bar } from 'mod'` with `mod` listed both bare and with importNames
	// produces two findings, which the corpus tests.
	paths    map[string][]NoRestrictedImportsPath
	patterns []noRestrictedImportsCompiledPattern
}

// newNoRestrictedImportsChecker compiles the configuration once per file.
func newNoRestrictedImportsChecker(
	ctx rule.Context,
	settings NoRestrictedImportsOptions,
) (*noRestrictedImportsChecker, bool) {
	checker := &noRestrictedImportsChecker{ctx: ctx, paths: map[string][]NoRestrictedImportsPath{}}

	for _, path := range settings.Paths {
		checker.paths[path.Name] = append(checker.paths[path.Name], path)
	}

	for _, entry := range settings.Patterns {
		compiled := noRestrictedImportsCompiledPattern{entry: entry}
		flags := "iu"
		if entry.CaseSensitive {
			flags = "u"
		}
		if len(entry.Group) > 0 {
			matcher, ok := noRestrictedImportsCompileGroup(entry.Group, entry.CaseSensitive)
			if !ok {
				return nil, false
			}
			compiled.matcher = matcher
		}
		if entry.Regex != "" {
			expression, err := esregexp.Compile(entry.Regex, flags)
			if err != nil {
				return nil, false
			}
			compiled.regularExpression = expression
		}
		// The import-name patterns are always `u`, never `iu`. `caseSensitive` governs the module
		// path only, which upstream expresses by building these two with a bare "u" regardless.
		if entry.ImportNamePattern != "" {
			expression, err := esregexp.Compile(entry.ImportNamePattern, "u")
			if err != nil {
				return nil, false
			}
			compiled.importNamePattern = expression
		}
		if entry.AllowImportNamePattern != "" {
			expression, err := esregexp.Compile(entry.AllowImportNamePattern, "u")
			if err != nil {
				return nil, false
			}
			compiled.allowedNamePattern = expression
		}
		checker.patterns = append(checker.patterns, compiled)
	}
	return checker, true
}

// noRestrictedImportsBinding is one imported name and where to report it.
//
// The range is carried rather than the node because a namespace re-export (`export * from 'm'`) has
// no specifier node to point at, and its finding is anchored on the `*` token instead.
type noRestrictedImportsBinding struct {
	// name is `default` for a default import, `*` for a namespace one, and the imported name
	// otherwise -- the name the MODULE exports, not the local alias.
	name string

	// reportRange is where the finding is anchored.
	reportRange core.TextRange

	// typeOnly marks a specifier written `import { type A }`, which `allowTypeImports` exempts.
	typeOnly bool
}

// checkImportDeclaration handles `import ... from '...'` in all its forms.
func (c *noRestrictedImportsChecker) checkImportDeclaration(node *ast.Node) {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return
	}
	source, ok := noRestrictedImportsSpecifierText(declaration.ModuleSpecifier)
	if !ok {
		return
	}

	// `imports.BindingsOf` decides which of the three shapes this statement carries, which is the
	// question upstream answers by filtering one flat `specifiers` array. The three live in three
	// different places in this AST, and a rule reading only one of them pins half the forms and
	// looks complete -- a mistake this tree has made and corrected once already.
	declared := imports.BindingsOf(node)

	bindings := []noRestrictedImportsBinding{}
	if declared.Default != nil {
		bindings = append(bindings, noRestrictedImportsBinding{
			name:        "default",
			reportRange: rule.TokenRange(c.ctx.SourceFile, declared.Default),
		})
	}
	if declared.Namespace != nil {
		bindings = append(bindings, noRestrictedImportsBinding{
			name:        "*",
			reportRange: rule.TokenRange(c.ctx.SourceFile, declared.Namespace),
		})
	}
	for _, specifier := range declared.Named {
		bindings = append(bindings, noRestrictedImportsBinding{
			name:        noRestrictedImportsImportedName(specifier),
			reportRange: rule.TokenRange(c.ctx.SourceFile, specifier),
			typeOnly:    specifier.AsImportSpecifier().IsTypeOnly,
		})
	}

	c.check(node, source, bindings, noRestrictedImportsDeclarationIsTypeOnly(node))
}

// checkExportDeclaration handles a re-export, which names a module exactly as an import does.
//
// An export with no module specifier (`export { a };`) names nothing and is skipped, which is
// upstream's `if (node.source)` guard.
func (c *noRestrictedImportsChecker) checkExportDeclaration(node *ast.Node) {
	declaration := node.AsExportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return
	}
	source, ok := noRestrictedImportsSpecifierText(declaration.ModuleSpecifier)
	if !ok {
		return
	}

	bindings := []noRestrictedImportsBinding{}
	switch {
	case declaration.ExportClause == nil:
		// `export * from 'm'`. Upstream reports at `sourceCode.getFirstToken(node, 1)`, the token
		// after `export`, which is the `*`.
		if star := noRestrictedImportsStarToken(c.ctx, node); star != nil {
			bindings = append(bindings, noRestrictedImportsBinding{name: "*", reportRange: *star})
		}

	case declaration.ExportClause.Kind == ast.KindNamespaceExport:
		// `export * as ns from 'm'`. Upstream's ExportAllDeclaration arm reaches this with the same
		// `*` token, because it reads the token rather than the clause.
		if star := noRestrictedImportsStarToken(c.ctx, node); star != nil {
			bindings = append(bindings, noRestrictedImportsBinding{name: "*", reportRange: *star})
		}

	case declaration.ExportClause.Kind == ast.KindNamedExports:
		if elements := declaration.ExportClause.AsNamedExports().Elements; elements != nil {
			for _, specifier := range elements.Nodes {
				bindings = append(bindings, noRestrictedImportsBinding{
					name:        noRestrictedImportsExportedLocalName(specifier),
					reportRange: rule.TokenRange(c.ctx.SourceFile, specifier),
					typeOnly:    specifier.AsExportSpecifier().IsTypeOnly,
				})
			}
		}
	}

	c.check(node, source, bindings, noRestrictedImportsExportIsTypeOnly(node))
}

// checkImportEquals handles `import foo = require('bar')`.
//
// Upstream passes an EMPTY import-name map for this form, so only a whole-module restriction can
// report on it and an `importNames` entry never does. Ported from upstream's source rather than
// from a measured verdict: its corpus has no case for this arm.
func (c *noRestrictedImportsChecker) checkImportEquals(node *ast.Node) {
	declaration := node.AsImportEqualsDeclaration()
	if declaration == nil || declaration.ModuleReference == nil {
		return
	}
	if declaration.ModuleReference.Kind != ast.KindExternalModuleReference {
		return
	}
	expression := declaration.ModuleReference.AsExternalModuleReference().Expression
	source, ok := noRestrictedImportsSpecifierText(expression)
	if !ok {
		return
	}
	c.check(node, source, nil, declaration.IsTypeOnly)
}

// check is upstream's `checkNode`: the path list first, then every matching pattern group.
func (c *noRestrictedImportsChecker) check(
	node *ast.Node,
	source string,
	bindings []noRestrictedImportsBinding,
	declarationIsTypeOnly bool,
) {
	// Upstream trims the specifier before comparing, so `import x from ' fs '` is restricted by
	// `fs`. The trim applies to both the path list and the pattern list.
	source = text.TrimWhitespace(source)

	c.checkPaths(node, source, bindings, declarationIsTypeOnly)

	for _, pattern := range c.patterns {
		if !c.patternMatches(pattern, source) {
			continue
		}
		c.reportForPattern(node, pattern, source, bindings, declarationIsTypeOnly)
	}
}

// patternMatches answers upstream's `isRestrictedPattern`.
func (c *noRestrictedImportsChecker) patternMatches(
	pattern noRestrictedImportsCompiledPattern,
	source string,
) bool {
	if pattern.regularExpression != nil {
		return pattern.regularExpression.Test(source)
	}
	return pattern.matcher.Ignores(source)
}

// checkPaths is upstream's `checkRestrictedPathAndReport`.
func (c *noRestrictedImportsChecker) checkPaths(
	node *ast.Node,
	source string,
	bindings []noRestrictedImportsBinding,
	declarationIsTypeOnly bool,
) {
	entries, restricted := c.paths[source]
	if !restricted {
		return
	}

	for _, entry := range entries {
		if entry.AllowTypeImports && declarationIsTypeOnly {
			continue
		}

		if len(entry.ImportNames) == 0 && len(entry.AllowImportNames) == 0 {
			// The whole module is banned, so the finding is on the statement rather than on any
			// one binding.
			c.report(node, noRestrictedImportsPathMessage(source, entry.Message))
			continue
		}

		for _, binding := range bindings {
			if binding.name == "*" {
				if len(entry.ImportNames) > 0 {
					c.reportRange(binding.reportRange, noRestrictedImportsEverythingMessage(
						source, entry.ImportNames, entry.Message))
				} else if len(entry.AllowImportNames) > 0 {
					c.reportRange(binding.reportRange,
						noRestrictedImportsEverythingWithAllowMessage(
							source, entry.AllowImportNames, entry.Message))
				}
				continue
			}

			if entry.AllowTypeImports && binding.typeOnly {
				continue
			}

			if noRestrictedImportsContains(entry.ImportNames, binding.name) {
				c.reportRange(binding.reportRange,
					noRestrictedImportsImportNameMessage(source, binding.name, entry.Message))
			}
			if len(entry.AllowImportNames) > 0 &&
				!noRestrictedImportsContains(entry.AllowImportNames, binding.name) {
				c.reportRange(binding.reportRange, noRestrictedImportsAllowedImportNameMessage(
					source, binding.name, entry.AllowImportNames, entry.Message))
			}
		}
	}
}

// reportForPattern is upstream's `reportPathForPatterns`.
//
// The four name-restriction fields are mutually exclusive, which the decoder enforces, so the arms
// below never both apply. Their ORDER still matters for the namespace case, where upstream picks
// among four messages by which field is set.
func (c *noRestrictedImportsChecker) reportForPattern(
	node *ast.Node,
	pattern noRestrictedImportsCompiledPattern,
	source string,
	bindings []noRestrictedImportsBinding,
	declarationIsTypeOnly bool,
) {
	entry := pattern.entry
	if entry.AllowTypeImports && declarationIsTypeOnly {
		return
	}

	restrictsNames := len(entry.ImportNames) > 0 || pattern.importNamePattern != nil
	allowsNames := len(entry.AllowImportNames) > 0 || pattern.allowedNamePattern != nil
	if !restrictsNames && !allowsNames {
		c.report(node, noRestrictedImportsPatternMessage(source, entry.Message))
		return
	}

	for _, binding := range bindings {
		if binding.name == "*" {
			switch {
			case len(entry.ImportNames) > 0:
				c.reportRange(binding.reportRange, noRestrictedImportsPatternEverythingMessage(
					source, entry.ImportNames, entry.Message))
			case len(entry.AllowImportNames) > 0:
				c.reportRange(binding.reportRange,
					noRestrictedImportsEverythingWithAllowMessage(
						source, entry.AllowImportNames, entry.Message))
			case pattern.allowedNamePattern != nil:
				c.reportRange(binding.reportRange,
					noRestrictedImportsEverythingWithAllowedPatternMessage(
						source, entry.AllowImportNamePattern, entry.Message))
			default:
				c.reportRange(binding.reportRange,
					noRestrictedImportsPatternEverythingWithRegexMessage(
						source, entry.ImportNamePattern, entry.Message))
			}
			continue
		}

		if entry.AllowTypeImports && binding.typeOnly {
			continue
		}

		if noRestrictedImportsContains(entry.ImportNames, binding.name) ||
			(pattern.importNamePattern != nil && pattern.importNamePattern.Test(binding.name)) {
			c.reportRange(binding.reportRange, noRestrictedImportsPatternAndImportNameMessage(
				source, binding.name, entry.Message))
		}

		// An `else if` upstream, not two independent tests: a name permitted by neither list is
		// reported once, under whichever list was configured.
		if len(entry.AllowImportNames) > 0 &&
			!noRestrictedImportsContains(entry.AllowImportNames, binding.name) {
			c.reportRange(binding.reportRange, noRestrictedImportsAllowedImportNameMessage(
				source, binding.name, entry.AllowImportNames, entry.Message))
		} else if pattern.allowedNamePattern != nil &&
			!pattern.allowedNamePattern.Test(binding.name) {
			c.reportRange(binding.reportRange, noRestrictedImportsAllowedNamePatternMessage(
				source, binding.name, entry.AllowImportNamePattern, entry.Message))
		}
	}
}

// report anchors a finding on a whole statement.
func (c *noRestrictedImportsChecker) report(node *ast.Node, message rule.Message) {
	c.ctx.ReportNode(node, message)
}

// reportRange anchors a finding on a span, which a namespace re-export needs because it has no
// specifier node.
func (c *noRestrictedImportsChecker) reportRange(span core.TextRange, message rule.Message) {
	c.ctx.ReportRange(span, message)
}

// noRestrictedImportsContains answers whether a name is in a list.
func noRestrictedImportsContains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

// noRestrictedImportsSpecifierText reads a module specifier's cooked string value.
func noRestrictedImportsSpecifierText(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindStringLiteral {
		return "", false
	}
	return node.Text(), true
}

// noRestrictedImportsImportedName reads the name the MODULE exports for an import specifier.
//
// `import { a as b }` imports `a` and binds `b`, and this rule restricts what a module exports, so
// the answer is `a`. On an `ImportSpecifier` here `PropertyName` is the imported name when an alias
// is present and nil otherwise, which is the mirror of the export specifier's layout.
func noRestrictedImportsImportedName(specifier *ast.Node) string {
	imported := specifier.AsImportSpecifier()
	if imported.PropertyName != nil {
		return noRestrictedExportsNameText(imported.PropertyName)
	}
	return noRestrictedExportsNameText(imported.Name())
}

// noRestrictedImportsExportedLocalName reads the name a re-export takes FROM the module.
//
// `export { a as b } from 'm'` reads `a` out of the module and publishes `b`, and the restriction is
// about what is taken, so the answer is `a`.
func noRestrictedImportsExportedLocalName(specifier *ast.Node) string {
	exported := specifier.AsExportSpecifier()
	if exported.PropertyName != nil {
		return noRestrictedExportsNameText(exported.PropertyName)
	}
	return noRestrictedExportsNameText(exported.Name())
}

// noRestrictedImportsDeclarationIsTypeOnly is upstream's `isTypeOnlyImport`.
//
// Two questions, either of which is enough: the declaration itself is `import type`, or it has
// specifiers and EVERY one of them is `type`-marked. The second is why `import { type A, B }` is not
// type-only while `import { type A, type B }` is.
//
// The declaration-level marker is `PhaseModifier == KindTypeKeyword` on the import clause in this
// version of the compiler, not an `IsTypeOnly` field. Measured, because the natural guess is wrong.
func noRestrictedImportsDeclarationIsTypeOnly(node *ast.Node) bool {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause.PhaseModifier == ast.KindTypeKeyword {
		return true
	}

	// The named specifiers again through the shared reader, for the reason given at the call site.
	// A default or namespace import is not `type`-markable per specifier, so an all-type list means
	// a list of NAMED specifiers and nothing else -- `import def, { type A } from 'm'` is not
	// type-only, and reading only the named ones would say it was.
	declared := imports.BindingsOf(node)
	if declared.Default != nil || declared.Namespace != nil || len(declared.Named) == 0 {
		return false
	}
	for _, specifier := range declared.Named {
		if !specifier.AsImportSpecifier().IsTypeOnly {
			return false
		}
	}
	return true
}

// noRestrictedImportsExportIsTypeOnly is upstream's `isTypeOnlyExport`, the re-export mirror.
func noRestrictedImportsExportIsTypeOnly(node *ast.Node) bool {
	declaration := node.AsExportDeclaration()
	if declaration == nil {
		return false
	}
	if declaration.IsTypeOnly {
		return true
	}
	if declaration.ExportClause == nil || declaration.ExportClause.Kind != ast.KindNamedExports {
		return false
	}
	elements := declaration.ExportClause.AsNamedExports().Elements
	if elements == nil || len(elements.Nodes) == 0 {
		return false
	}
	for _, specifier := range elements.Nodes {
		if !specifier.AsExportSpecifier().IsTypeOnly {
			return false
		}
	}
	return true
}

// noRestrictedImportsStarToken finds the `*` in `export * from 'm'` and `export * as ns from 'm'`.
//
// Upstream reads `sourceCode.getFirstToken(node, 1)`, the second token of the statement. There is no
// token list here, so the `*` is found by scanning the statement's own source text from its start,
// which cannot run past the specifier because a `*` is the only thing that can follow `export` in
// these two forms.
func noRestrictedImportsStarToken(ctx rule.Context, node *ast.Node) *core.TextRange {
	if ctx.SourceFile == nil {
		return nil
	}
	statement := rule.TokenRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()
	if statement.Pos() < 0 || statement.End() > len(text) {
		return nil
	}
	offset := strings.IndexByte(text[statement.Pos():statement.End()], '*')
	if offset < 0 {
		return nil
	}
	span := core.NewTextRange(statement.Pos()+offset, statement.Pos()+offset+1)
	return &span
}
