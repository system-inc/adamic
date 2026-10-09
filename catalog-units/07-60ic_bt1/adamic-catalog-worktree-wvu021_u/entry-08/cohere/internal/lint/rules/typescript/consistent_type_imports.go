package typescript

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConsistentTypeImportsPrefer selects which direction the rule enforces.
type ConsistentTypeImportsPrefer string

const (
	// ConsistentTypeImportsPreferTypeImports requires `import type` wherever every use is a type.
	ConsistentTypeImportsPreferTypeImports ConsistentTypeImportsPrefer = "type-imports"
	// ConsistentTypeImportsPreferNoTypeImports forbids the `type` keyword on imports entirely.
	ConsistentTypeImportsPreferNoTypeImports ConsistentTypeImportsPrefer = "no-type-imports"
)

// ConsistentTypeImportsFixStyle selects where the repair places the `type` keyword. It changes the
// repair only, never what reports.
type ConsistentTypeImportsFixStyle string

const (
	// ConsistentTypeImportsFixStyleSeparate writes `import type { A } from 'm'`.
	ConsistentTypeImportsFixStyleSeparate ConsistentTypeImportsFixStyle = "separate-type-imports"
	// ConsistentTypeImportsFixStyleInline writes `import { type A } from 'm'`.
	ConsistentTypeImportsFixStyleInline ConsistentTypeImportsFixStyle = "inline-type-imports"
)

// ConsistentTypeImportsOptions configures the rule.
//
// The authoritative surface is oxc's `ConsistentTypeImportsConfig`, which carries exactly these
// three fields under `serde(rename_all = "camelCase", deny_unknown_fields)`. Read from that struct
// rather than from an inventory column: the inventory names two options and there are three, and
// the third one is the only one whose default is `true`.
type ConsistentTypeImportsOptions struct {
	// DisallowTypeAnnotations forbids `import()` written inside a type, as in `type T = import('m')`.
	//
	// **Defaults to true**, which is the field most likely to be got wrong, because every other
	// boolean option in this tree defaults to false and a Go zero value would silently turn this
	// arm off. Three of the corpus's seventy-nine diagnostics come from this arm alone, and one
	// corpus pass case exists only to pin the `false` setting.
	DisallowTypeAnnotations bool

	// FixStyle selects where the repair puts the keyword: a separate `import type` declaration, or
	// `type` inline on each specifier.
	FixStyle ConsistentTypeImportsFixStyle

	// Prefer selects the direction. `type-imports` is the default and is what the live config sets.
	Prefer ConsistentTypeImportsPrefer
}

// DefaultConsistentTypeImportsOptions is what the rule uses when the configuration names nothing.
//
// Every field here is non-zero, so this cannot be spelled as a zero-value struct. That matters more
// than it reads: a rule configured as a bare `"error"` receives nil options, and a decoder built
// from `rule.DecodeOptionsInto` would hand the rule three empty strings and a false, matching no arm
// and reporting nothing while looking correctly wired.
func DefaultConsistentTypeImportsOptions() ConsistentTypeImportsOptions {
	return ConsistentTypeImportsOptions{
		DisallowTypeAnnotations: true,
		FixStyle:                ConsistentTypeImportsFixStyleSeparate,
		Prefer:                  ConsistentTypeImportsPreferTypeImports,
	}
}

var messageConsistentTypeImportsTypeOverValue = rule.Message{
	Id: "typeOverValue",
	Description: "Every name this declaration imports is used only in a type position, so nothing " +
		"it brings in survives compilation. Written as `import type`, the statement is erased " +
		"outright rather than leaving a module request that a bundler has to keep, and a reader " +
		"can see at a glance that it cannot affect runtime behavior.",
}

var messageConsistentTypeImportsAvoidImportType = rule.Message{
	Id: "avoidImportType",
	Description: "This configuration asks for value imports throughout, and the `type` keyword " +
		"here opts one declaration out of that. Drop it so every import in the codebase reads " +
		"the same way.",
}

var messageConsistentTypeImportsNoImportTypeAnnotations = rule.Message{
	Id: "noImportTypeAnnotations",
	Description: "An `import()` written inside a type annotation hides a module dependency where " +
		"nothing looking at the import list will find it. Move it to a real `import type` " +
		"declaration at the top of the file and refer to the name.",
}

// consistentTypeImportsSomeAreOnlyTypes builds the per-declaration message naming the offenders.
//
// The names are interpolated, so a fixture asserting only the message id cannot see this line at
// all. The rendered text is asserted by equality in the fixtures for exactly that reason.
func consistentTypeImportsSomeAreOnlyTypes(names []string) rule.Message {
	return rule.Message{
		Id: "someImportsAreOnlyTypes",
		Description: fmt.Sprintf("Imports %s are only used as type.",
			consistentTypeImportsWordList(names)),
	}
}

// ConsistentTypeImports flags imports whose every use is a type, and `import type` under the
// inverted setting.
//
//	valid:   import type { Foo } from 'foo'; let foo: Foo;
//	valid:   import { A, B } from 'foo'; const foo: A = B();
//	valid:   import { type A, B } from 'foo'; type T = A; const b = B;
//	invalid: import { Foo } from 'foo'; let foo: Foo;
//	invalid: import { A, B } from 'foo'; const foo: A = B();   // A only
//	invalid: type T = import('foo');
//
// Ported from oxc's `consistent_type_imports`, which is what the gate runs.
//
// # The whole rule is one question asked per imported name
//
// For each name a declaration binds: does it have at least one reference, and is every one of those
// references a type reference? A name with no references at all answers no, deliberately, so an
// entirely unused import is silent here and left to the unused-import rules. That asymmetry is
// upstream's `is_only_has_type_references`, whose first act is to return false on an empty iterator,
// and it is pinned by the corpus's `import Foo from 'foo';` pass case.
//
// # What counts as a type reference is not `IsPartOfTypeNode`, and that was measured
//
// The obvious substrate for this is `ast.IsPartOfTypeNode`, and reaching for it alone loses three
// of the corpus's own failing shapes. Probed directly against the parser, `typeof Foo` inside a
// type alias gives a parent of `KindTypeQuery` with `IsPartOfTypeNode` answering **false**, and the
// left side of a qualified name such as `foo.Bar` gives `KindQualifiedName` with the same false. Both
// report upstream. So the predicate here is a walk that names those two kinds explicitly, and the
// corpus's `// TSTypeQuery` and `// TSQualifiedName` comments are upstream telling you the same
// thing.
//
// The opposite error is available too, and the parser hands us the discrimination for free. Probed
// on the release binary: `class C implements A {}` **reports** and `class C extends A {}` is
// **silent**. In this AST the `implements` name parses to `KindTypeReference` and the `extends` name
// to `KindExpressionWithTypeArguments`, so treating the latter as a type position would report a
// superclass as type-only and propose erasing an import the emitted code needs. There is no
// heritage-clause token to inspect; the parse shape already decided it.
//
// # Three sites resolve to the import through a different symbol
//
// `export { X }`, `export type { X }` and `export default X` put an identifier in the file that
// `GetSymbolAtLocation` resolves to the *export's* own symbol rather than the import's, so a rule
// matching on symbol identity sees no reference there at all and goes silent on six corpus cases.
// The recovery is `GetExportSpecifierLocalTargetSymbol`, which resolves an export specifier back to
// what it re-exports.
//
// Whether such a reference counts as a type is then decided by the export's own `type` keyword and
// not by position, which is the half a reading of the AST cannot supply. Measured on the release
// binary, all four with the same import:
//
//	export { Type };        SILENT   — a value export
//	export default Type;    SILENT   — a value export
//	export = Type;          SILENT   — a value export
//	export type { Type };   REPORTS  — a type-only export
//	export { type Type };   REPORTS  — likewise, per specifier
//
// Upstream reaches the same answers by a route worth recording, because it looks like it should
// give different ones. Its semantic builder marks all of these `Read | Type`, genuinely ambiguous,
// and then collapses the pair to a plain value reference during resolution whenever the symbol is
// value-capable, which an ordinary import alias always is. The ambiguity is real in the flags and
// resolved away before any rule sees it, so reproducing the flags rather than the outcome would be
// reproducing a mechanism instead of a decision.
//
// # JSX makes the JSX factory a value use
//
// In a file with JSX, the binding named for the JSX factory (`React` unless `jsxFactory` says
// otherwise) is a value use, under any import form, because ESLint's scope manager makes every JSX
// element a reference to it. In a file without JSX it is judged like any other name, so
// `import React from 'react'` used only as `React.FC` reports. This was a skip by local name, taken
// from oxc, until phi web found the gap; see `consistentTypeImportsJsxValueNames`.
//
// # The repair is upstream's, delivered as one edit
//
// Upstream's fixer yields several primitive edits per report (an insertion before the declaration,
// ` type` after `import`, a removal per moved specifier), and our engine refuses two insertions at
// one point, measured at zero applied and 1024 rejections when those were proposed side by side. So
// consistent_type_imports_fix.go computes the same primitive edits, applies them to the declaration's
// own text, and proposes the result as one replacement. Against upstream's corpus that reproduces
// its repair byte for byte, spacing included, on 69 of its 70 cases.
//
// The seventieth is the one deliberate divergence: upstream merges moved names into an `import type`
// that already exists elsewhere in the file, which is a second statement and so a second edit. Here
// they go into a new declaration beside the reported one. The repair moves only names the report
// found type-only, so a class decorator metadata serializes is never moved, and a declaration
// carrying import attributes, which a type-only import cannot have, reports with no repair.
var ConsistentTypeImports = rule.Rule{
	Name: "@typescript-eslint/consistent-type-imports",

	// Every part of the rule resolves a name to a declaration. There is no syntactic route to
	// "which import does this identifier bind to", and the corpus turns on exactly that: several
	// failing cases are textually near-identical to passing ones and differ only in what a name
	// resolves to.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// The compiler options.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[ConsistentTypeImportsOptions](options)
		if !ok {
			// A rule configured as a bare `"error"` is handed nil, and every default here is
			// non-zero. Without this line the rule registers, passes its fixtures through the
			// decoder, and reports nothing on the real tree.
			settings = DefaultConsistentTypeImportsOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				sourceFile := node.AsSourceFile()
				if sourceFile == nil {
					return
				}

				declarations := imports.Declarations(sourceFile)
				judgesValueImports := settings.Prefer != ConsistentTypeImportsPreferNoTypeImports &&
					consistentTypeImportsHasValueImport(declarations)
				facts := consistentTypeImportsScan(ctx, sourceFile, settings.DisallowTypeAnnotations,
					judgesValueImports, declarations)

				switch settings.Prefer {
				case ConsistentTypeImportsPreferNoTypeImports:
					reportTypeKeywordsToRemove(ctx, declarations)
				default:
					if judgesValueImports {
						reportImportsUsedOnlyAsTypes(ctx, sourceFile, declarations, facts, settings.FixStyle)
					}
				}
			},
		}
	},
}

// consistentTypeImportsFacts is what the type-imports judgment reads about a file, gathered in one
// walk. Nil maps when the file has no value import to judge.
type consistentTypeImportsFacts struct {
	// byText indexes the identifiers spelled like an imported local name. See
	// consistentTypeImportsLocalNames.
	byText map[string][]*ast.Node
	// metadataRoots are the identifiers decorator metadata emits as runtime references. See
	// markDecoratorMetadataRoots.
	metadataRoots map[*ast.Node]bool
	// jsxValueNames are the bindings this file's JSX uses as values. See
	// consistentTypeImportsJsxValueNames.
	jsxValueNames map[string]bool
}

// consistentTypeImportsScan walks the file once, reporting every `import('m')` written inside a type
// when reportImportTypes is set, and gathering the facts the type-imports judgment reads when
// gatherFacts is set.
//
// One walk rather than four. Each of those questions used to walk the whole file on its own, and
// together they made this rule 153ms of rule CPU on a cold ahra run, among the costliest rules that
// visit once per file (#fcac58b). The questions are independent per node, so one visit answers all
// of them.
//
// A whole-file walk rather than a listener on `KindImportType`, because the rule already owns a
// `KindSourceFile` listener for the parts that must gather before they judge, and splitting one
// judgment across two anchors makes the ordering of findings depend on the walk rather than on the
// source. The `import()` findings come out in source order, ahead of the declarations' findings,
// which is what the corpus asserts.
func consistentTypeImportsScan(
	ctx rule.Context,
	sourceFile *ast.SourceFile,
	reportImportTypes bool,
	gatherFacts bool,
	declarations []*ast.Node,
) consistentTypeImportsFacts {
	if !reportImportTypes && !gatherFacts {
		return consistentTypeImportsFacts{}
	}

	var wanted map[string]bool
	var byText map[string][]*ast.Node
	var metadataRoots map[*ast.Node]bool
	var metadata decoratorMetadataSettings
	if gatherFacts {
		wanted = consistentTypeImportsLocalNames(declarations)
		byText = make(map[string][]*ast.Node, len(wanted))
		metadataRoots = map[*ast.Node]bool{}
		metadata = decoratorMetadataSettingsOf(ctx)
	}
	hasJsx := false
	hasFragment := false

	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil {
			return
		}
		if reportImportTypes && current.Kind == ast.KindImportType {
			ctx.ReportNode(current, messageConsistentTypeImportsNoImportTypeAnnotations)
		}
		if gatherFacts {
			switch current.Kind {
			case ast.KindIdentifier:
				if text := current.Text(); wanted[text] {
					byText[text] = append(byText[text], current)
				}
			case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
				hasJsx = true
			case ast.KindJsxFragment:
				hasJsx = true
				hasFragment = true
			}
			if metadata.enabled {
				markDecoratorMetadataRoots(current, metadataRoots, metadata)
			}
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())

	if !gatherFacts {
		return consistentTypeImportsFacts{}
	}
	return consistentTypeImportsFacts{
		byText:        byText,
		metadataRoots: metadataRoots,
		jsxValueNames: consistentTypeImportsJsxValueNames(ctx, hasJsx, hasFragment),
	}
}

// consistentTypeImportsHasValueImport reports whether any declaration could report under
// `prefer: type-imports`: one that binds something and is not already `import type`. A file with
// none pays nothing for the facts.
func consistentTypeImportsHasValueImport(declarations []*ast.Node) bool {
	for _, statement := range declarations {
		clause := statement.AsImportDeclaration().ImportClause
		if clause != nil && !clause.AsImportClause().IsTypeOnly() {
			return true
		}
	}
	return false
}

// reportTypeKeywordsToRemove implements `prefer: no-type-imports`.
//
// Two anchors, and they are separate findings rather than one per declaration: a whole
// `import type { A } from 'm'` reports against the declaration, while an inline `import { type A }`
// reports against the specifier. The corpus pins both, and one three-declaration case produces
// exactly three findings.
//
// Nothing about references is consulted here. The setting says the keyword is unwanted wherever it
// appears, so a type-only import of a name used only as a type still reports.
//
// declarations is every import declaration, including those inside an ambient module body, the way
// upstream visits ImportDeclaration wherever it sits.
func reportTypeKeywordsToRemove(ctx rule.Context, declarations []*ast.Node) {
	for _, statement := range declarations {
		clause := statement.AsImportDeclaration().ImportClause
		if clause == nil {
			continue
		}
		if clause.AsImportClause().IsTypeOnly() {
			reportConsistentTypeImportsWithFix(ctx, statement, messageConsistentTypeImportsAvoidImportType,
				func() (rule.Fix, bool) { return consistentTypeImportsRemoveTypeFix(ctx.SourceFile, statement, nil) })
			continue
		}
		// An inline `type` on a specifier is only reachable when the declaration itself is not
		// type-only: `import type { type A }` is a syntax error, so the two forms never coexist.
		for _, named := range imports.BindingsOf(statement).Named {
			if named.Kind == ast.KindImportSpecifier && named.AsImportSpecifier().IsTypeOnly {
				reportConsistentTypeImportsWithFix(ctx, named, messageConsistentTypeImportsAvoidImportType,
					func() (rule.Fix, bool) { return consistentTypeImportsRemoveTypeFix(ctx.SourceFile, statement, named) })
			}
		}
	}
}

// reportImportsUsedOnlyAsTypes implements `prefer: type-imports`, which is the default and the
// setting the live config uses.
//
// One finding per declaration at most. Which of the two messages it carries depends on whether
// *every* specifier was found type-only, and the count is against the specifier list rather than
// against the named specifiers: `import A, {} from 'foo'` has one specifier and one type-only name,
// so it takes the whole-declaration message, which is what the corpus's `import A, {} from 'foo'`
// case asserts.
//
// declarations is every import declaration, including those inside an ambient module body
// (#7g5r6vt), and facts is consistentTypeImportsScan's answer for the file.
func reportImportsUsedOnlyAsTypes(
	ctx rule.Context,
	sourceFile *ast.SourceFile,
	declarations []*ast.Node,
	facts consistentTypeImportsFacts,
	fixStyle ConsistentTypeImportsFixStyle,
) {
	for _, statement := range declarations {
		declaration := statement.AsImportDeclaration()
		clause := declaration.ImportClause
		if clause == nil {
			// `import 'm'` binds nothing. The corpus writes one as the first statement of a
			// two-import case precisely to check it is stepped over rather than tripped on.
			continue
		}
		if clause.AsImportClause().IsTypeOnly() {
			// Already what the rule is asking for.
			continue
		}

		bindings := imports.BindingsOf(statement)
		specifierCount := 0
		typeOnlyNames := []string{}
		// The specifier nodes behind typeOnlyNames, which is what the repair moves.
		typeOnlySpecifiers := map[*ast.Node]bool{}

		consider := func(specifier *ast.Node, local *ast.Node, alreadyTypeOnly bool) {
			if local == nil {
				return
			}
			specifierCount++
			if facts.jsxValueNames[local.Text()] {
				// Every JSX element in the file is a value reference to the JSX factory, so the
				// binding it names is a value use. See consistentTypeImportsJsxValueNames.
				return
			}
			if alreadyTypeOnly {
				return
			}
			if isReferencedOnlyAsType(ctx, facts.byText, facts.metadataRoots, local) {
				typeOnlyNames = append(typeOnlyNames, local.Text())
				typeOnlySpecifiers[specifier] = true
			}
		}

		consider(bindings.Default, bindings.Default, false)
		if bindings.Namespace != nil {
			consider(bindings.Namespace, bindings.Namespace.Name(), false)
		}
		for _, named := range bindings.Named {
			if named.Kind != ast.KindImportSpecifier {
				continue
			}
			consider(named, named.Name(), named.AsImportSpecifier().IsTypeOnly)
		}
		fix := func() (rule.Fix, bool) {
			return consistentTypeImportsTypeImportFix(sourceFile, statement, typeOnlySpecifiers, fixStyle)
		}

		if len(typeOnlyNames) == 0 {
			continue
		}

		if len(typeOnlyNames) == specifierCount {
			// `import type {} from 'm' with { type: 'json' }` is not valid TypeScript, so upstream
			// withholds the whole-declaration finding when an attributes clause is present rather
			// than proposing something that cannot be written. Probed: an import with an `assert`
			// clause is silent even when every name is type-only.
			if declaration.Attributes == nil {
				reportConsistentTypeImportsWithFix(ctx, statement, messageConsistentTypeImportsTypeOverValue, fix)
			}
			continue
		}

		reportConsistentTypeImportsWithFix(ctx, statement, consistentTypeImportsSomeAreOnlyTypes(typeOnlyNames), fix)
	}
}

// reportConsistentTypeImportsWithFix reports with the repair when one can be built, and bare when the
// fixer declines, so a declaration it cannot rewrite safely is still named.
func reportConsistentTypeImportsWithFix(ctx rule.Context, node *ast.Node, message rule.Message, fix func() (rule.Fix, bool)) {
	if proposed, ok := fix(); ok {
		ctx.ReportNodeWithFixes(node, message, proposed)
		return
	}
	ctx.ReportNode(node, message)
}

// consistentTypeImportsJsxValueNames names the bindings JSX in this file uses as values.
//
// ESLint's scope manager makes every JSX element and fragment a value reference to the JSX pragma,
// and every fragment one to the fragment factory too. The pragma is the first identifier of the
// program's `jsxFactory`, `React` when none is set, whatever the `jsx` mode; the fragment name is the
// first identifier of `jsxFragmentFactory`, and nothing when none is set. Read from
// @typescript-eslint/parser and scope-manager as installed, which is the gate this rule is held to.
//
// This replaced a skip by local name, which was oxc's: a default or namespace import spelled `React`
// was never judged at all, so `import React from 'react'` used only in types in a file with no JSX
// was silent where ESLint reports it, a parity gap phi web found. The binding is a value use exactly
// when the file has JSX, under any import form, and a renamed import is not the pragma.
//
// hasJsx and hasFragment are whether the file holds any JSX element or fragment, and any fragment,
// from consistentTypeImportsScan's walk.
func consistentTypeImportsJsxValueNames(ctx rule.Context, hasJsx bool, hasFragment bool) map[string]bool {
	if !hasJsx {
		return nil
	}

	firstIdentifier := func(factory string) string {
		return text.TrimWhitespace(strings.SplitN(factory, ".", 2)[0])
	}
	names := map[string]bool{"React": true}
	if ctx.Program != nil {
		options := ctx.Program.Options()
		if options.JsxFactory != "" {
			names = map[string]bool{firstIdentifier(options.JsxFactory): true}
		}
		if hasFragment && options.JsxFragmentFactory != "" {
			names[firstIdentifier(options.JsxFragmentFactory)] = true
		}
	}
	return names
}

// consistentTypeImportsLocalNames is every local name a value import declaration binds, which is
// every name `isReferencedOnlyAsType` can be asked about. A declaration already written
// `import type` is never judged, so its names are left out.
//
// The scan indexes identifiers by text rather than walking the file per imported name, because the
// naive shape re-walks the whole tree for every specifier in every import statement: measured on the
// ahra tree at 3,407 files, the per-name walk cost 667ms and 2.9% of all rule time. Keyed by text
// because an identifier spelled differently cannot resolve to this import, so the text is an exact
// pre-filter on a question that would otherwise cost a checker call per identifier; symbol identity
// below already implies it. Only these texts are kept: indexing every identifier cost 37 MB in 477K
// objects on a cold ahra run (#942rdnn), nearly all of it slices for names no import binds.
func consistentTypeImportsLocalNames(declarations []*ast.Node) map[string]bool {
	names := map[string]bool{}
	for _, statement := range declarations {
		clause := statement.AsImportDeclaration().ImportClause
		if clause == nil || clause.AsImportClause().IsTypeOnly() {
			continue
		}
		bindings := imports.BindingsOf(statement)
		if bindings.Default != nil {
			names[bindings.Default.Text()] = true
		}
		if bindings.Namespace != nil {
			if name := bindings.Namespace.Name(); name != nil {
				names[name.Text()] = true
			}
		}
		for _, named := range bindings.Named {
			if name := named.Name(); name != nil {
				names[name.Text()] = true
			}
		}
	}
	return names
}

// isReferencedOnlyAsType answers the rule's single question for one imported name.
//
// False when the name has no references at all, which is upstream's own first line and not an
// accident of the loop: an unused import is a different rule's finding.
func isReferencedOnlyAsType(
	ctx rule.Context,
	byText map[string][]*ast.Node,
	metadataRoots map[*ast.Node]bool,
	local *ast.Node,
) bool {
	target := ctx.TypeChecker.GetSymbolAtLocation(local)
	if target == nil {
		return false
	}

	found := false
	for _, candidate := range byText[local.Text()] {
		if candidate == local || !resolvesToImport(ctx, candidate, target) {
			continue
		}
		found = true
		if !isTypeReference(candidate) {
			// One value use settles it, and the remaining candidates cannot change the answer.
			return false
		}
		if metadataRoots[candidate] && isRuntimeValueImport(ctx, target) {
			// A type position that decorator metadata emits as a runtime reference. See
			// `decoratorMetadataRoots`.
			return false
		}
	}
	return found
}

// markDecoratorMetadataRoots names the identifiers that decorator metadata turns into runtime references,
// for one node of consistentTypeImportsScan's walk.
//
// Under `emitDecoratorMetadata`, TypeScript serializes the types of a decorated declaration into
// `design:paramtypes`, `design:type` and `design:returntype`, and a type that names a class is
// emitted as a reference to that class's value. Rewriting its import to `import type` erases the
// binding, the metadata silently becomes `Object`, and nothing fails until a framework reads the
// metadata at run time. Measured on api-phi-health, whose Base GraphQL layer reads it: 1,295
// findings where ESLint reports none.
//
// typescript-eslint answers this by skipping every file that contains a decorator. This mirrors
// what the compiler itself keeps instead, `markDecoratorAliasReferenced` and
// `getEntityNameForDecoratorMetadata` in the vendored checker, so the rest of such a file still
// reports. The answer is one identifier per serialized type: the first identifier of the entity
// name the checker extracts, which is the only part of the type the emitted code references.
//
// The walk asks only when the option is on, which is the whole gate: without metadata no type position
// reaches run time.
func markDecoratorMetadataRoots(current *ast.Node, roots map[*ast.Node]bool, metadata decoratorMetadataSettings) {
	if !ast.CanHaveDecorators(current) || !ast.HasDecorators(current) || current.Modifiers() == nil ||
		!ast.NodeCanBeDecorated(metadata.legacyDecorators, current, current.Parent, current.Parent.Parent) {
		return
	}
	mark := func(typeNode *ast.Node) {
		entityName := decoratorMetadataEntityName(typeNode, metadata.strictNullChecks)
		if entityName != nil && ast.IsEntityName(entityName) {
			roots[ast.GetFirstIdentifier(entityName)] = true
		}
	}
	switch current.Kind {
	case ast.KindClassDeclaration:
		if constructor := ast.GetFirstConstructorWithBody(current); constructor != nil {
			for _, parameter := range constructor.Parameters() {
				mark(decoratorMetadataParameterType(parameter))
			}
		}
	case ast.KindGetAccessor, ast.KindSetAccessor:
		annotation := decoratorMetadataAccessorType(current)
		if annotation == nil {
			annotation = decoratorMetadataAccessorType(decoratorMetadataOtherAccessor(current))
		}
		mark(annotation)
	case ast.KindMethodDeclaration:
		for _, parameter := range current.Parameters() {
			mark(decoratorMetadataParameterType(parameter))
		}
		mark(current.Type())
	case ast.KindPropertyDeclaration:
		mark(current.Type())
	case ast.KindParameter:
		// A decorated parameter serializes its whole signature, siblings and return included.
		signature := current.Parent
		for _, parameter := range signature.Parameters() {
			mark(decoratorMetadataParameterType(parameter))
		}
		mark(signature.Type())
	}
}

// decoratorMetadataSettings is what markDecoratorMetadataRoots reads from the compiler options,
// read once per file.
type decoratorMetadataSettings struct {
	enabled          bool
	legacyDecorators bool
	strictNullChecks bool
}

// decoratorMetadataSettingsOf reads the options that decide whether and how metadata is emitted.
func decoratorMetadataSettingsOf(ctx rule.Context) decoratorMetadataSettings {
	options := ctx.Program.Options()
	return decoratorMetadataSettings{
		enabled:          options.EmitDecoratorMetadata.IsTrue(),
		legacyDecorators: options.ExperimentalDecorators.IsTrue(),
		strictNullChecks: type_checking.IsStrictCompilerOptionEnabled(options, options.StrictNullChecks),
	}
}

// decoratorMetadataEntityName is the checker's `getEntityNameForDecoratorMetadata`.
//
// A union or intersection serializes to one reference only when every member names the same
// identifier, after `never` and, without strict null checks, `null` and `undefined` are elided. With
// strict null checks on, `Foo | null` yields no entity name at all, so the import is not kept by
// metadata and the rule still reports it, which is what the compiler would emit.
func decoratorMetadataEntityName(node *ast.Node, strictNullChecks bool) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindIntersectionType:
		return decoratorMetadataCommonEntityName(node.AsIntersectionTypeNode().Types.Nodes, strictNullChecks)
	case ast.KindUnionType:
		return decoratorMetadataCommonEntityName(node.AsUnionTypeNode().Types.Nodes, strictNullChecks)
	case ast.KindConditionalType:
		conditional := node.AsConditionalTypeNode()
		return decoratorMetadataCommonEntityName(
			[]*ast.Node{conditional.TrueType, conditional.FalseType}, strictNullChecks)
	case ast.KindParenthesizedType:
		return decoratorMetadataEntityName(node.AsParenthesizedTypeNode().Type, strictNullChecks)
	case ast.KindNamedTupleMember:
		return decoratorMetadataEntityName(node.AsNamedTupleMember().Type, strictNullChecks)
	case ast.KindTypeReference:
		return node.AsTypeReferenceNode().TypeName
	}
	return nil
}

// decoratorMetadataCommonEntityName is the checker's `getEntityNameForDecoratorMetadataFromTypeList`.
func decoratorMetadataCommonEntityName(typeNodes []*ast.Node, strictNullChecks bool) *ast.Node {
	var common *ast.Node
	for _, typeNode := range typeNodes {
		if typeNode.Kind == ast.KindNeverKeyword {
			continue
		}
		if !strictNullChecks && (typeNode.Kind == ast.KindUndefinedKeyword ||
			typeNode.Kind == ast.KindLiteralType && typeNode.AsLiteralTypeNode().Literal.Kind == ast.KindNullKeyword) {
			continue
		}
		individual := decoratorMetadataEntityName(typeNode, strictNullChecks)
		if individual == nil {
			return nil
		}
		if common == nil {
			common = individual
			continue
		}
		if !ast.IsIdentifier(common) || !ast.IsIdentifier(individual) || common.Text() != individual.Text() {
			return nil
		}
	}
	return common
}

// decoratorMetadataParameterType is the checker's `getParameterTypeNodeForDecoratorCheck`: a rest
// parameter serializes its element type.
func decoratorMetadataParameterType(parameter *ast.Node) *ast.Node {
	typeNode := parameter.Type()
	if parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return ast.GetRestParameterElementType(typeNode)
	}
	return typeNode
}

// decoratorMetadataAccessorType is the checker's `getAnnotatedAccessorTypeNode`: a getter's return
// annotation, or a setter's parameter annotation.
func decoratorMetadataAccessorType(accessor *ast.Node) *ast.Node {
	if accessor == nil {
		return nil
	}
	if accessor.Kind == ast.KindGetAccessor {
		return accessor.Type()
	}
	for _, parameter := range accessor.Parameters() {
		if !ast.IsThisParameter(parameter) {
			return parameter.Type()
		}
	}
	return nil
}

// decoratorMetadataOtherAccessor finds the accessor of the opposite kind for the same member.
//
// The checker reaches it through the member's symbol. The same member is the same name with the
// same staticness in the same class body, which is what that symbol groups, and asking the AST keeps
// a type-graph call out of a walk that otherwise needs none.
func decoratorMetadataOtherAccessor(accessor *ast.Node) *ast.Node {
	otherKind := ast.KindSetAccessor
	if accessor.Kind == ast.KindSetAccessor {
		otherKind = ast.KindGetAccessor
	}
	name := accessor.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return nil
	}
	for _, member := range accessor.Parent.Members() {
		if member.Kind != otherKind || ast.IsStatic(member) != ast.IsStatic(accessor) {
			continue
		}
		if memberName := member.Name(); memberName != nil && memberName.Kind == ast.KindIdentifier &&
			memberName.Text() == name.Text() {
			return member
		}
	}
	return nil
}

// isRuntimeValueImport reports whether an import binding names something that exists at run time.
//
// The checker keeps an import for metadata only when `symbolIsValue` holds of it and its target is
// not a const enum, whose uses are inlined. An interface or type alias is serialized as `Object`
// and needs no binding; under `isolatedModules` the compiler goes further and requires those to be
// imported with `import type` (TS1272), so continuing to report them is exactly what tsc asks for.
func isRuntimeValueImport(ctx rule.Context, binding *ast.Symbol) bool {
	resolved := ctx.TypeChecker.GetAliasedSymbol(binding)
	if resolved == nil {
		return false
	}
	return resolved.Flags&ast.SymbolFlagsValue != 0 && resolved.Flags&ast.SymbolFlagsConstEnum == 0
}

// resolvesToImport reports whether an identifier binds to the import's symbol.
//
// Two accessors rather than one, and the second is not a fallback for a failure: an identifier
// inside an export specifier resolves through `GetSymbolAtLocation` to the export's own symbol,
// which is a perfectly good symbol that simply is not this one. Asking only the first accessor
// answers "no reference here" for every `export { X }`, `export type { X }` and `export default X`
// in the file, which is silence on six corpus cases rather than a crash.
func resolvesToImport(ctx rule.Context, identifier *ast.Node, target *ast.Symbol) bool {
	if ctx.TypeChecker.GetSymbolAtLocation(identifier) == target {
		return true
	}
	return ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol(identifier) == target
}

// isTypeReference reports whether one identifier's occurrence is a type reference.
//
// See the rule's doc comment for what was measured here. The short form: `IsPartOfTypeNode` alone
// misses `typeof X` and the left of a qualified name, an export specifier's type-ness comes from
// the export's keyword rather than from its position, and a class `extends` clause must stay a
// value reference.
func isTypeReference(identifier *ast.Node) bool {
	if ast.IsPartOfTypeNode(identifier) {
		return true
	}

	for current := identifier; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindTypeQuery, ast.KindQualifiedName:
			// `type T = typeof X` and the `foo` of `type T = foo.Bar`. Both answer false to
			// IsPartOfTypeNode and both report upstream.
			return true

		case ast.KindExportSpecifier:
			// `export { type X }` marks the specifier; `export type { X }` marks the declaration
			// two levels up. Either one makes this a type reference and neither is visible from
			// the identifier's own position.
			if current.AsExportSpecifier().IsTypeOnly {
				return true
			}
			return isTypeOnlyExportDeclaration(current)

		case ast.KindExportAssignment:
			// `export = X` and `export default X`. Both are value exports, measured.
			return false

		case ast.KindExpressionWithTypeArguments:
			// A class `extends` clause. The emitted code needs the binding, so this is a value use
			// even though it sits inside a heritage clause.
			return false
		}

		if ast.IsStatement(current) {
			// Nothing above a statement can turn an expression into a type.
			return false
		}
	}
	return false
}

// isTypeOnlyExportDeclaration reports whether the export declaration containing a specifier carries
// the `type` keyword, as `export type { X }` does.
func isTypeOnlyExportDeclaration(specifier *ast.Node) bool {
	for current := specifier.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindExportDeclaration {
			return current.AsExportDeclaration().IsTypeOnly
		}
		if ast.IsStatement(current) && current.Kind != ast.KindExportDeclaration {
			return false
		}
	}
	return false
}

// consistentTypeImportsWordList renders names the way upstream's message does.
//
// One name is bare, two are joined with `and` and no comma, and three or more take an Oxford comma
// before the final `and`. The three-name form is what a reimplementation gets wrong, and the corpus
// asserts it directly with `Imports A, C, and D are only used as type.`
func consistentTypeImportsWordList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		return strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
	}
}

// DecodeConsistentTypeImportsOptions maps the wire keys onto the settings the rule reads.
//
// Hand written rather than `rule.DecodeOptionsInto` because every default is non-zero, so a
// zero-value struct is not a valid configuration of this rule. An unrecognized enum value falls
// back to the default rather than disabling the arm: upstream refuses such a configuration outright
// and there is no error channel that reaches a user here, so the choice is between the documented
// behavior and silence, and silence would look like the rule working.
func DecodeConsistentTypeImportsOptions(raw []byte) (any, error) {
	options := DefaultConsistentTypeImportsOptions()

	decoded, err := rule.DecodeOptionsInto[consistentTypeImportsRawOptions]()(raw)
	if err != nil {
		return options, err
	}
	wire, _ := decoded.(consistentTypeImportsRawOptions)

	if wire.DisallowTypeAnnotations != nil {
		options.DisallowTypeAnnotations = *wire.DisallowTypeAnnotations
	}
	if wire.Prefer != nil {
		switch ConsistentTypeImportsPrefer(*wire.Prefer) {
		case ConsistentTypeImportsPreferTypeImports:
			options.Prefer = ConsistentTypeImportsPreferTypeImports
		case ConsistentTypeImportsPreferNoTypeImports:
			options.Prefer = ConsistentTypeImportsPreferNoTypeImports
		}
	}
	if wire.FixStyle != nil {
		switch ConsistentTypeImportsFixStyle(*wire.FixStyle) {
		case ConsistentTypeImportsFixStyleSeparate:
			options.FixStyle = ConsistentTypeImportsFixStyleSeparate
		case ConsistentTypeImportsFixStyleInline:
			options.FixStyle = ConsistentTypeImportsFixStyleInline
		}
	}

	return options, nil
}

// consistentTypeImportsRawOptions is the wire shape, with pointers so an absent key is
// distinguishable from a key set to the zero value. That distinction is the whole reason this
// struct exists separately: `disallowTypeAnnotations: false` and an absent key mean opposite things.
type consistentTypeImportsRawOptions struct {
	DisallowTypeAnnotations *bool   `json:"disallowTypeAnnotations"`
	FixStyle                *string `json:"fixStyle"`
	Prefer                  *string `json:"prefer"`
}
