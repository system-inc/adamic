package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The three messages. Upstream interpolates the offending names into two of them, and `rule.Message`
// has no interpolation layer, so the names become concatenation at the report site.

func messageConsistentTypeExportsTypeOverValue() rule.Message {
	return rule.Message{
		Id: "typeOverValue",
		Description: "Every name in this export is a type, so the statement emits an import at " +
			"runtime that resolves to nothing. Writing `export type` tells the compiler and any " +
			"bundler that the whole statement can be erased.",
	}
}

func messageConsistentTypeExportsSingleExportIsType(name string) rule.Message {
	return rule.Message{
		Id: "singleExportIsType",
		Description: "The export " + name + " is a type rather than a value, so it is being " +
			"re-exported through a statement that survives to runtime. Move it into an " +
			"`export type` statement so the two kinds stay separable.",
	}
}

func messageConsistentTypeExportsMultipleExportsAreTypes(names string) rule.Message {
	return rule.Message{
		Id: "multipleExportsAreTypes",
		Description: "The exports " + names + " are types rather than values, so they are being " +
			"re-exported through a statement that survives to runtime. Move them into an " +
			"`export type` statement so the two kinds stay separable.",
	}
}

// ConsistentTypeExports requires a type-only export to be written as `export type`.
//
//	valid:   export type { TypeA };
//	valid:   export { valueB };
//	valid:   export { type TypeA, valueB };        an inline specifier is already separated
//	invalid: export { TypeA };                     becomes export type { TypeA };
//	invalid: export { TypeA, valueB };             splits into two statements
//	invalid: export * from './only-types';         becomes export type * from './only-types';
//
// Ported from `@typescript-eslint/consistent-type-exports`, reading the clone at
// `packages/eslint-plugin/src/rules/consistent-type-exports.ts` and measuring every verdict and
// every repair against the installed 8.67.0 build driven over a real TypeScript program with the
// plugin's own multi-file fixtures on disk, since half this rule's judgment is about what another
// file exports.
//
// # The fixer splits a statement, and what the split carries is the thing to watch
//
// A mixed export becomes two statements, so the repair has to decide what text each specifier keeps.
// Upstream re-renders each one from its local and exported names rather than copying its source
// text, which is the shape that lost type information twice in this project. It is safe here
// because a specifier holds only names: there is no annotation, no generic and no modifier inside
// one. Both spellings that carry more than a bare identifier were measured rather than assumed:
//
//	export { TypeA as Renamed, valueC };       becomes export type { TypeA as Renamed };
//	                                           plus    export { valueC };
//	export { TypeA as 'string-name', valueC }; keeps the quoted form verbatim
//
// The second is why the rendering reads a string specifier's RAW text rather than its cooked value:
// re-quoting would change `'a'` into `"a"` and could not represent an escape at all.
//
// # Two arms, and only one has exposure in this tree
//
// The named-export arm judges `export { ... }`. The star arm judges `export * from './x'` by asking
// what the OTHER module exports. Measured with a control: this tree holds 88 named exports and 59
// type exports and ZERO export-star statements, so the star arm finds nothing here and is ported for
// correctness rather than for cleanup.
//
// # The star arm reproduces an upstream hack, deliberately
//
// A module that re-exports only types through its own `export *` puts those symbols in a table the
// checker does not expose. Upstream works around it by calling two different lookups and comparing:
// `getPropertiesOfType` returns everything that was originally a value, while `getPropertyOfType`
// returns nothing for a name that reached the module through a type-only star. Both are reachable
// here and the pair discriminates: probed on a type-only module the property count is zero, and on a
// module exporting one value it is one.
//
// The workaround is reproduced rather than improved because the thing it works around is the same
// here, and a cleaner-looking answer would be a different rule.
//
// # Reporting happens after the whole file, because one source can be exported twice
//
// Upstream gathers per module specifier and reports on `Program:exit`. There is no exit hook here,
// so the gathering and the reporting both happen inside the source-file listener, which fires before
// its children and can walk the tree itself.
var ConsistentTypeExports = rule.Rule{
	Name:             "@typescript-eslint/consistent-type-exports",
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The zero value happens to match upstream's default here, since the one option defaults to
		// false, but the fallback is written anyway: a later option with a true default would
		// otherwise invert silently, and a reader should not have to check the default to know
		// whether nil is handled.
		parsed, decoded := rule.OptionsAs[ConsistentTypeExportsOptions](options)
		if !decoded {
			parsed = DefaultConsistentTypeExportsOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				reportConsistentTypeExports(ctx, parsed)
			},
		}
	},
}

// consistentTypeExportsReport is upstream's `ReportValueExport`.
type consistentTypeExportsReport struct {
	node                 *ast.Node
	inlineTypeSpecifiers []*ast.Node
	typeBasedSpecifiers  []*ast.Node
	valueSpecifiers      []*ast.Node
}

// reportConsistentTypeExports is upstream's two visitors plus its `Program:exit`.
//
// Upstream keys its gathering map by module specifier and never reads the grouping again, so the map
// is not reproduced: the reports are collected in source order and emitted in source order, which is
// the same sequence upstream produces because its map preserves insertion order and each entry holds
// its own reports.
func reportConsistentTypeExports(ctx rule.Context, options ConsistentTypeExportsOptions) {
	var reports []consistentTypeExportsReport

	for _, statement := range ctx.SourceFile.AsNode().Statements() {
		if statement.Kind != ast.KindExportDeclaration {
			continue
		}
		declaration := statement.AsExportDeclaration()

		// `export * from './x'` and `export * as ns from './x'` have no named bindings, or a
		// namespace binding. Upstream handles both in its ExportAllDeclaration visitor.
		if declaration.ExportClause == nil ||
			declaration.ExportClause.Kind == ast.KindNamespaceExport {
			reportConsistentTypeExportsStar(ctx, statement, declaration)
			continue
		}

		if declaration.ExportClause.Kind != ast.KindNamedExports {
			continue
		}

		report, reportable := consistentTypeExportsClassify(ctx, statement, declaration)
		if reportable {
			reports = append(reports, report)
		}
	}

	for _, report := range reports {
		consistentTypeExportsEmit(ctx, report, options)
	}
}

// consistentTypeExportsClassify sorts a named export's specifiers into upstream's three buckets.
func consistentTypeExportsClassify(
	ctx rule.Context,
	statement *ast.Node,
	declaration *ast.ExportDeclaration,
) (consistentTypeExportsReport, bool) {
	report := consistentTypeExportsReport{node: statement}

	// `export type { ... }` cannot hold a specifier this rule objects to, because every name in it
	// is already type-only. Upstream still walks it to collect value specifiers, and then reports
	// only when `exportKind === 'type' && valueSpecifiers.length`, which its own specifier loop
	// makes unreachable: the loop is skipped entirely for a type export, so `valueSpecifiers` is
	// always empty there. Reproduced as an early return with the reasoning recorded rather than as
	// dead bookkeeping.
	if declaration.IsTypeOnly {
		return report, false
	}

	for _, specifier := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
		element := specifier.AsExportSpecifier()

		// An inline `type` specifier is already separated and is carried along so the repair can
		// re-render it without its keyword.
		if element.IsTypeOnly {
			report.inlineTypeSpecifiers = append(report.inlineTypeSpecifiers, specifier)
			continue
		}

		name := element.Name()
		if name == nil {
			continue
		}

		typeBased, resolved := consistentTypeExportsSymbolIsTypeBased(ctx,
			ctx.TypeChecker.GetSymbolAtLocation(name))
		if !resolved {
			// Upstream's `undefined` case: a specifier whose symbol cannot be resolved is left
			// alone rather than guessed at, so an unresolvable name never produces a finding.
			//
			// This branch is UNREACHABLE through this tree's checker and the verdict is recorded
			// rather than the branch deleted. A mutant forcing an unresolved specifier to be
			// bucketed as type-based survives every fixture, including four written for exactly
			// this shape, and the reason is that nothing here answers unresolved.
			//
			// Traced rather than argued. `export { NotDeclaredAnywhere };` still yields a symbol,
			// carrying the alias flag and one declaration, which is the specifier itself.
			// `SkipAlias` then resolves it to a DIFFERENT symbol whose flags include the value bit,
			// so the recursion answers `(false, true)` and the specifier is bucketed as a value.
			// Upstream is silent on the same input and so is this rule, by a different route.
			//
			// The branch is kept because it is upstream's and because the only thing making it
			// unreachable is what the checker chooses to return for an undeclared name. A checker
			// that answered nil there, or a caller reaching this with a genuinely unknown symbol,
			// would need it. That verdict names the two callers that exist today; adding a third
			// voids it.
			continue
		}
		if typeBased {
			report.typeBasedSpecifiers = append(report.typeBasedSpecifiers, specifier)
		} else {
			report.valueSpecifiers = append(report.valueSpecifiers, specifier)
		}
	}

	return report, len(report.typeBasedSpecifiers) > 0
}

// consistentTypeExportsSymbolIsTypeBased is upstream's `isSymbolTypeBased`.
//
// The boolean is the verdict and the second return is upstream's `undefined`, meaning the symbol
// could not be resolved and the specifier must be left alone.
//
// Upstream recurses through `getImmediateAliasedSymbol` one link at a time; `SkipAlias` walks the
// whole chain in one call. The two agree on the answer because every intermediate link is an alias
// and the loop's only exit is a non-alias, which is exactly where `SkipAlias` stops. Measured on a
// re-exported value and a re-exported type through two files: the value resolves to a value and the
// type does not.
func consistentTypeExportsSymbolIsTypeBased(ctx rule.Context, symbol *ast.Symbol) (bool, bool) {
	if symbol == nil {
		return false, false
	}

	// Upstream also declines `checker.isUnknownSymbol(symbol)`, which this tree does not expose:
	// the checker's `unknownSymbol` is an unexported field with no linkname. The nil check above
	// covers the shape that matters, because an unresolvable name gives no symbol at all here
	// rather than the placeholder one. Stated rather than silently dropped; the residual gap is a
	// name that resolves TO the placeholder, which would be read as type-based and reported. No
	// corpus case and no measured shape reaches it, and the failure direction is a false positive
	// on source that already fails to compile.

	// A symbol declared by a type-only import or export is type-based whatever its flags say.
	for _, declaration := range symbol.Declarations {
		if consistentTypeExportsIsTypeOnlyImportOrExport(declaration) {
			return true, true
		}
	}

	if symbol.Flags&ast.SymbolFlagsValue != 0 {
		return false, true
	}

	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		target := shimchecker.SkipAlias(symbol, ctx.TypeChecker)
		if target == nil || target == symbol {
			return true, true
		}
		return consistentTypeExportsSymbolIsTypeBased(ctx, target)
	}

	return true, true
}

// consistentTypeExportsIsTypeOnlyImportOrExport is `ts.isTypeOnlyImportOrExportDeclaration`.
//
// The four shapes that can carry the keyword: a specifier on either side, and a clause on either
// side. `Node.IsTypeOnly()` already dispatches over them, and it is used rather than reaching into
// each declaration type: an import clause spells its own type-only-ness as a phase modifier rather
// than a boolean, which a hand-written switch gets wrong and this does not.
//
// The specifier arms then walk UP as well, because a specifier inside `export type { A }` is
// type-only through its parent rather than through its own flag.
func consistentTypeExportsIsTypeOnlyImportOrExport(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.IsTypeOnly() {
		return true
	}

	switch node.Kind {
	case ast.KindImportSpecifier, ast.KindNamespaceImport:
		return consistentTypeExportsAncestorIsTypeOnly(node, ast.KindImportClause,
			ast.KindImportDeclaration)
	case ast.KindExportSpecifier:
		return consistentTypeExportsAncestorIsTypeOnly(node, ast.KindExportDeclaration,
			ast.KindSourceFile)
	}
	return false
}

// consistentTypeExportsAncestorIsTypeOnly walks up to the declaration that carries the keyword.
//
// The second kind is the ceiling: reaching it means the walk left the construct without finding the
// node that could be type-only, so the answer is no.
func consistentTypeExportsAncestorIsTypeOnly(node *ast.Node, carrier ast.Kind, ceiling ast.Kind) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == carrier {
			return current.IsTypeOnly()
		}
		if current.Kind == ceiling || current.Kind == ast.KindSourceFile {
			return false
		}
	}
	return false
}

// consistentTypeExportsEmit is upstream's `Program:exit` body for one report.
func consistentTypeExportsEmit(
	ctx rule.Context,
	report consistentTypeExportsReport,
	options ConsistentTypeExportsOptions,
) {
	if len(report.valueSpecifiers) == 0 {
		// Every name is a type, so the whole statement becomes `export type`.
		fixes, buildable := consistentTypeExportsInsertTypeFixes(ctx, report)
		if !buildable {
			ctx.ReportNode(report.node, messageConsistentTypeExportsTypeOverValue())
			return
		}
		ctx.ReportNodeWithFixes(report.node, messageConsistentTypeExportsTypeOverValue(), fixes...)
		return
	}

	names := make([]string, 0, len(report.typeBasedSpecifiers))
	for _, specifier := range report.typeBasedSpecifiers {
		local, readable := consistentTypeExportsLocalName(ctx, specifier)
		if !readable {
			continue
		}
		names = append(names, local)
	}

	message := messageConsistentTypeExportsSingleExportIsType(consistentTypeExportsFormatWordList(names))
	if len(names) != 1 {
		message = messageConsistentTypeExportsMultipleExportsAreTypes(
			consistentTypeExportsFormatWordList(names))
	}

	var fixes []rule.Fix
	var buildable bool
	if options.FixMixedExportsWithInlineTypeSpecifier {
		fixes, buildable = consistentTypeExportsInlineSpecifierFixes(ctx, report)
	} else {
		fixes, buildable = consistentTypeExportsSeparateFixes(ctx, report)
	}
	if !buildable {
		ctx.ReportNode(report.node, message)
		return
	}
	ctx.ReportNodeWithFixes(report.node, message, fixes...)
}

// consistentTypeExportsFormatWordList is upstream's `formatWordList`.
//
// One name alone, two joined by ` and `, and more than two joined by commas with ` and ` before the
// last. Upstream's implementation joins the head with commas and appends the tail after ` and `,
// which produces `a, b and c` rather than an Oxford comma.
func consistentTypeExportsFormatWordList(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}

// consistentTypeExportsInsertTypeFixes is upstream's `fixExportInsertType`.
//
// The `type` keyword is inserted after `export`, and any inline `type` already on a specifier is
// removed, since the statement now carries it. Upstream removes from the keyword's start to the
// start of the next token, which takes the whitespace with it.
func consistentTypeExportsInsertTypeFixes(
	ctx rule.Context,
	report consistentTypeExportsReport,
) ([]rule.Fix, bool) {
	sourceText := ctx.SourceFile.Text()
	statementRange := rule.TokenRange(ctx.SourceFile, report.node)

	// The `export` keyword is the statement's first token, so its end is six characters in.
	const exportKeyword = "export"
	if statementRange.Pos()+len(exportKeyword) > len(sourceText) ||
		sourceText[statementRange.Pos():statementRange.Pos()+len(exportKeyword)] != exportKeyword {
		return nil, false
	}
	afterExport := statementRange.Pos() + len(exportKeyword)

	fixes := []rule.Fix{
		rule.ReplaceRange(shimcore.NewTextRange(afterExport, afterExport), " type"),
	}

	for _, specifier := range report.inlineTypeSpecifiers {
		keywordRange, found := consistentTypeExportsInlineTypeKeywordRange(ctx, specifier)
		if !found {
			return nil, false
		}
		fixes = append(fixes, rule.RemoveRange(keywordRange))
	}
	return fixes, true
}

// consistentTypeExportsInlineTypeKeywordRange finds the `type ` an inline specifier carries.
//
// From the specifier's own start through the whitespace that follows the keyword, matching
// upstream's removal from the keyword's start to the next token's start.
func consistentTypeExportsInlineTypeKeywordRange(
	ctx rule.Context,
	specifier *ast.Node,
) (shimcore.TextRange, bool) {
	sourceText := ctx.SourceFile.Text()
	start := rule.TokenRange(ctx.SourceFile, specifier).Pos()

	const typeKeyword = "type"
	if start+len(typeKeyword) > len(sourceText) ||
		sourceText[start:start+len(typeKeyword)] != typeKeyword {
		return shimcore.TextRange{}, false
	}

	end := start + len(typeKeyword)
	for end < len(sourceText) && consistentTypeExportsIsSpace(sourceText[end]) {
		end++
	}
	return shimcore.NewTextRange(start, end), true
}

// consistentTypeExportsIsSpace answers whether a byte is inline whitespace.
func consistentTypeExportsIsSpace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}

// consistentTypeExportsSeparateFixes is upstream's `fixSeparateNamedExports`.
//
// Two edits. The brace interior is rewritten to hold only the value specifiers, and a whole
// `export type { ... }` statement is inserted above. Upstream emits them in the other order and
// ESLint sorts by range; the order here is the same edit either way because the two ranges do not
// overlap.
func consistentTypeExportsSeparateFixes(
	ctx rule.Context,
	report consistentTypeExportsReport,
) ([]rule.Fix, bool) {
	typeSpecifiers := make([]*ast.Node, 0,
		len(report.typeBasedSpecifiers)+len(report.inlineTypeSpecifiers))
	typeSpecifiers = append(typeSpecifiers, report.typeBasedSpecifiers...)
	typeSpecifiers = append(typeSpecifiers, report.inlineTypeSpecifiers...)

	typeNames, typeOk := consistentTypeExportsSpecifierList(ctx, typeSpecifiers)
	valueNames, valueOk := consistentTypeExportsSpecifierList(ctx, report.valueSpecifiers)
	if !typeOk || !valueOk {
		return nil, false
	}

	open, closing, found := consistentTypeExportsBraceRange(ctx, report.node)
	if !found {
		return nil, false
	}

	statementRange := rule.TokenRange(ctx.SourceFile, report.node)
	source, hasSource := consistentTypeExportsModuleSpecifierText(ctx, report.node)

	inserted := "export type { " + typeNames + " }"
	if hasSource {
		inserted += " from '" + source + "'"
	}
	inserted += ";\n"

	return []rule.Fix{
		rule.ReplaceRange(shimcore.NewTextRange(statementRange.Pos(), statementRange.Pos()), inserted),
		rule.ReplaceRange(shimcore.NewTextRange(open+1, closing), " "+valueNames+" "),
	}, true
}

// consistentTypeExportsInlineSpecifierFixes is upstream's `fixAddTypeSpecifierToNamedExports`.
//
// Each type-based specifier gets an inline `type ` rather than being moved, which keeps one
// statement. Upstream returns nothing for a type-only export; that case cannot reach here because
// the classifier declines it earlier, and the guard is kept because it is upstream's.
func consistentTypeExportsInlineSpecifierFixes(
	ctx rule.Context,
	report consistentTypeExportsReport,
) ([]rule.Fix, bool) {
	if report.node.AsExportDeclaration().IsTypeOnly {
		return nil, false
	}

	fixes := make([]rule.Fix, 0, len(report.typeBasedSpecifiers))
	for _, specifier := range report.typeBasedSpecifiers {
		start := rule.TokenRange(ctx.SourceFile, specifier).Pos()
		fixes = append(fixes,
			rule.ReplaceRange(shimcore.NewTextRange(start, start), "type "))
	}
	return fixes, len(fixes) > 0
}

// consistentTypeExportsSpecifierList renders a run of specifiers as `a, b as c`.
func consistentTypeExportsSpecifierList(ctx rule.Context, specifiers []*ast.Node) (string, bool) {
	// BUCKET order rather than source order, which is upstream's and is measured rather than
	// assumed. `fixSeparateNamedExports` concatenates `[...typeBasedSpecifiers,
	// ...inlineTypeSpecifiers]`, so every name the checker found follows in source order and every
	// name already carrying an inline `type` comes after them all.
	//
	// That is visible whenever a statement mixes the two. Measured on the installed build:
	//
	//	export { type Type1, Type2, value1 };
	//	   becomes  export type { Type2, Type1 };  plus  export { value1 };
	//
	// `Type2` leads because it was found by the checker and `Type1` follows because it was already
	// inline, which reverses the order they were written in. Sorting by position here would read as
	// tidier and would disagree with upstream on exactly this shape, so the order is left as the
	// caller built it.
	parts := make([]string, 0, len(specifiers))
	for _, specifier := range specifiers {
		text, readable := consistentTypeExportsSpecifierText(ctx, specifier)
		if !readable {
			return "", false
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, ", "), true
}

// consistentTypeExportsSpecifierText is upstream's `getSpecifierText`.
//
// `local` alone when the two names agree, and `local as exported` when they differ. A string-literal
// name is rendered from its RAW source text rather than its cooked value, which upstream does with
// `.raw`: re-quoting a cooked value would change the quote character and could not represent an
// escape at all.
//
// The inline `type` keyword is deliberately NOT rendered. A specifier moved into an `export type`
// statement no longer needs it, and one left behind was never type-only.
func consistentTypeExportsSpecifierText(ctx rule.Context, specifier *ast.Node) (string, bool) {
	element := specifier.AsExportSpecifier()

	exportedNode := element.Name()
	localNode := element.PropertyName
	if localNode == nil {
		localNode = exportedNode
	}
	if exportedNode == nil || localNode == nil {
		return "", false
	}

	local, localOk := consistentTypeExportsNameText(ctx, localNode)
	exported, exportedOk := consistentTypeExportsNameText(ctx, exportedNode)
	if !localOk || !exportedOk {
		return "", false
	}

	if local == exported {
		return local, true
	}
	return local + " as " + exported, true
}

// consistentTypeExportsLocalName is the name upstream puts in the message.
//
// Upstream reads `specifier.local`, which is the ORIGINAL name rather than the alias, and takes a
// string literal's cooked `value` here where the fixer takes its raw text. Both are reproduced as
// written: the message names what the reader wrote on the left of `as`.
func consistentTypeExportsLocalName(ctx rule.Context, specifier *ast.Node) (string, bool) {
	element := specifier.AsExportSpecifier()
	localNode := element.PropertyName
	if localNode == nil {
		localNode = element.Name()
	}
	if localNode == nil {
		return "", false
	}
	if localNode.Kind == ast.KindStringLiteral {
		return localNode.Text(), true
	}
	return consistentTypeExportsNameText(ctx, localNode)
}

// consistentTypeExportsNameText renders a specifier name as written in the source.
func consistentTypeExportsNameText(ctx rule.Context, node *ast.Node) (string, bool) {
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, node)
	sourceText := ctx.SourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(sourceText) || trimmed.Pos() > trimmed.End() {
		return "", false
	}
	return sourceText[trimmed.Pos():trimmed.End()], true
}

// consistentTypeExportsBraceRange finds the export clause's braces.
func consistentTypeExportsBraceRange(ctx rule.Context, statement *ast.Node) (int, int, bool) {
	clause := statement.AsExportDeclaration().ExportClause
	if clause == nil {
		return 0, 0, false
	}
	clauseRange := rule.TokenRange(ctx.SourceFile, clause)
	sourceText := ctx.SourceFile.Text()

	open := -1
	for offset := clauseRange.Pos(); offset < clauseRange.End() && offset < len(sourceText); offset++ {
		if sourceText[offset] == '{' {
			open = offset
			break
		}
	}
	closing := -1
	for offset := clauseRange.End() - 1; offset >= clauseRange.Pos() && offset < len(sourceText); offset-- {
		if offset < 0 {
			break
		}
		if sourceText[offset] == '}' {
			closing = offset
			break
		}
	}
	if open < 0 || closing < 0 || open >= closing {
		return 0, 0, false
	}
	return open, closing, true
}

// consistentTypeExportsModuleSpecifierText answers a statement's `from` target, cooked.
//
// Upstream reads `node.source.value`, the cooked string, and then re-quotes it with single quotes
// when writing the new statement. Both halves are reproduced, which is why a double-quoted source
// comes back single-quoted: that is upstream's own output rather than a choice made here.
func consistentTypeExportsModuleSpecifierText(ctx rule.Context, statement *ast.Node) (string, bool) {
	specifier := statement.AsExportDeclaration().ModuleSpecifier
	if specifier == nil || specifier.Kind != ast.KindStringLiteral {
		return "", false
	}
	return specifier.Text(), true
}

// reportConsistentTypeExportsStar is upstream's `ExportAllDeclaration` visitor.
//
// It asks whether the module being re-exported has any value export at all, and reports when it does
// not. The question is answered through the module symbol's TYPE rather than through its export
// table, because that is the only route that distinguishes a type-only star re-export; see the note
// on the rule.
func reportConsistentTypeExportsStar(
	ctx rule.Context,
	statement *ast.Node,
	declaration *ast.ExportDeclaration,
) {
	if declaration.IsTypeOnly {
		return
	}
	specifier := declaration.ModuleSpecifier
	if specifier == nil {
		return
	}

	moduleSymbol := ctx.TypeChecker.GetSymbolAtLocation(specifier)
	if moduleSymbol == nil {
		return
	}
	moduleType := ctx.TypeChecker.GetTypeOfSymbol(moduleSymbol)
	if moduleType == nil {
		return
	}

	// Upstream's hack, reproduced. `getPropertiesOfType` lists everything that was originally a
	// value; `getPropertyOfType` answers nil for a name that reached this module through a
	// type-only star re-export. A name present in the first and absent from the second is therefore
	// type-only, and a name present in both is a real exported value.
	for _, property := range shimchecker.Checker_getPropertiesOfType(ctx.TypeChecker, moduleType) {
		if shimchecker.Checker_getPropertyOfType(ctx.TypeChecker, moduleType, property.Name) != nil {
			return
		}
	}

	fix, buildable := consistentTypeExportsStarFix(ctx, statement)
	if !buildable {
		ctx.ReportNode(statement, messageConsistentTypeExportsTypeOverValue())
		return
	}
	ctx.ReportNodeWithFixes(statement, messageConsistentTypeExportsTypeOverValue(), fix)
}

// consistentTypeExportsStarFix inserts `type ` before the asterisk.
//
// The asterisk has to be found as a TOKEN rather than as a byte, and the difference is not
// theoretical: upstream's own corpus writes an export-star with a block comment between `export` and
// `*`, so the first `*` in the statement belongs to `/* comment 2 */`. Scanning bytes put the
// keyword inside the comment and produced `/type * comment 2 */ *`, which is neither valid nor what
// upstream writes.
//
// Comments are skipped rather than the token stream consulted, because a statement this shape holds
// only comments and whitespace before its asterisk. Both comment forms are handled, and an
// unterminated one declines rather than running to the end of the file.
func consistentTypeExportsStarFix(ctx rule.Context, statement *ast.Node) (rule.Fix, bool) {
	sourceText := ctx.SourceFile.Text()
	statementRange := rule.TokenRange(ctx.SourceFile, statement)

	offset := statementRange.Pos()
	end := statementRange.End()
	if end > len(sourceText) {
		end = len(sourceText)
	}

	// Past the `export` keyword itself, so its own bytes cannot be mistaken for anything.
	const exportKeyword = "export"
	if offset+len(exportKeyword) > end ||
		sourceText[offset:offset+len(exportKeyword)] != exportKeyword {
		return rule.Fix{}, false
	}
	offset += len(exportKeyword)

	for offset < end {
		if consistentTypeExportsIsSpace(sourceText[offset]) {
			offset++
			continue
		}
		if sourceText[offset] == '*' {
			return rule.ReplaceRange(shimcore.NewTextRange(offset, offset), "type "), true
		}
		if sourceText[offset] == '/' && offset+1 < end {
			switch sourceText[offset+1] {
			case '/':
				for offset < end && sourceText[offset] != '\n' {
					offset++
				}
				continue
			case '*':
				closing := strings.Index(sourceText[offset+2:end], "*/")
				if closing < 0 {
					return rule.Fix{}, false
				}
				offset += 2 + closing + 2
				continue
			}
		}
		// Anything else before the asterisk is a shape this rule was not handed.
		return rule.Fix{}, false
	}
	return rule.Fix{}, false
}

// ConsistentTypeExportsOptions is the rule's configuration.
type ConsistentTypeExportsOptions struct {
	// FixMixedExportsWithInlineTypeSpecifier changes the repair for a mixed export from splitting
	// the statement to adding an inline `type` to each type specifier. Upstream defaults it to
	// false, so the splitting repair is what a bare configuration gets.
	FixMixedExportsWithInlineTypeSpecifier bool
}

// DefaultConsistentTypeExportsOptions is upstream's `defaultOptions`.
func DefaultConsistentTypeExportsOptions() ConsistentTypeExportsOptions {
	return ConsistentTypeExportsOptions{FixMixedExportsWithInlineTypeSpecifier: false}
}

// consistentTypeExportsRawOptions is the wire shape.
//
// A pointer so an absent key stays distinguishable from an explicit false. The default is false, so
// the two happen to agree today, and the pointer is kept because the distinction is what the
// decoder exists to preserve and a later default change would otherwise be silent.
type consistentTypeExportsRawOptions struct {
	FixMixedExportsWithInlineTypeSpecifier *bool `json:"fixMixedExportsWithInlineTypeSpecifier"`
}

// DecodeConsistentTypeExportsOptions maps the wire key onto the rule's setting.
func DecodeConsistentTypeExportsOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[consistentTypeExportsRawOptions]()(raw)
	if err != nil {
		return DefaultConsistentTypeExportsOptions(), err
	}

	wire, _ := decoded.(consistentTypeExportsRawOptions)
	options := DefaultConsistentTypeExportsOptions()
	if wire.FixMixedExportsWithInlineTypeSpecifier != nil {
		options.FixMixedExportsWithInlineTypeSpecifier = *wire.FixMixedExportsWithInlineTypeSpecifier
	}
	return options, nil
}
