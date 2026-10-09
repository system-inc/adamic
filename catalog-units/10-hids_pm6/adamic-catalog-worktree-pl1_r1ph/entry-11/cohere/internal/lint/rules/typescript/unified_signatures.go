package typescript

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// UnifiedSignatures reports two overloads that could be written as one.
//
//	valid:   function f(x: number): void; function f(x: string, y: number): void;
//	invalid: function f(x: number): void; function f(x: string): void;      -> f(x: number | string)
//	invalid: function f(): void; function f(x: number): void;               -> f(x?: number)
//	invalid: function f(x: number): void; function f(x: number): void;      -> identical
//
// Overloads are grouped per scope and per key, then every pair within a group is compared. A pair
// unifies when it agrees on return type, type parameters, and outer-type-parameter usage, and then
// differs in at most one place: one parameter's type (report a union), or one trailing parameter
// that could be optional or rest.
//
// # The rule compares types as SOURCE TEXT, not through the checker
//
// Upstream's `typesAreEqual` is `sourceCode.getText(a) === sourceCode.getText(b)`, and the rule
// makes zero `getTypeAtLocation` calls. That is a deliberate choice upstream, not an oversight: it
// is why `string` and `String` are different here, and why the reported union is spelled with the
// author's own words rather than the checker's normalisation. Ported as-is, which means this rule
// declares no type checker and costs nothing on the type-graph side.
//
// # Our parser answers the parameter questions differently, and more simply
//
// Every property upstream reads off TSESTree was measured against our parser before this file was
// written, because the parameter shapes are exactly where the two trees diverge:
//
//	upstream (TSESTree)                    ours (measured)
//	p.type === RestElement                 DotDotDotToken != nil
//	p.optional                             QuestionToken != nil
//	p.typeAnnotation.typeAnnotation        Type
//	TSParameterProperty wrapping p         the SAME node, carrying a modifier
//	ExportNamedDeclaration wrapping f      the SAME node, carrying an export modifier
//
// The last two are why this file is shorter than upstream's without deciding anything differently.
// TSESTree wraps a parameter property in a `TSParameterProperty` and an exported declaration in an
// `ExportNamedDeclaration`, so upstream carries `isTSParameterProperty` unwrapping at five call
// sites and a `getExportingNode`/`containingNode` pair through the whole overload-collection path.
// Our parser puts both on the node as modifiers, so the unwrapping has nothing to unwrap. Measured,
// not assumed: `export declare function f` reports parent `SourceFile` with modifiers
// `[ExportKeyword DeclareKeyword]`, and `constructor(private a?: string)` is one parameter node with
// one modifier and a question token.
//
// Getters and setters fall out the same way. Upstream filters them with `isGetterOrSetter` because
// TSESTree gives them the same node type as a method with a `kind` discriminator; our parser gives
// them `KindGetAccessor` and `KindSetAccessor`, distinct kinds, so they never enter the listeners
// this rule installs and there is no filter to write.
//
// # `this` parameters
//
// Two separate exemptions, both upstream's. A pair where one signature declares `this` and the other
// does not cannot unify at all. And a signature taking `this: void` is exempt outright, because
// `this: void` is how a callback declares it must not be called as a method, which is a distinction
// unification would erase. Our parser gives `this` as an ordinary parameter named `this`, so both
// tests read the name.
//
// # Cost
//
// One map per scope, holding the signatures seen in it, and a pairwise comparison within each key's
// group. A group of n signatures costs n(n-1)/2 comparisons, and n is the number of overloads
// sharing one name, which is small in practice.
var UnifiedSignatures = rule.Rule{
	Name: "@typescript-eslint/unified-signatures",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		// The value type, because that is what `DecodeOptionsInto` hands over. This line asserted a
		// pointer once, the comma-ok form swallowed the mismatch, and both options read false for
		// every configuration; `TestUnifiedSignaturesOptionsArriveThroughTheDecoder` holds it.
		settings, _ := rule.OptionsAs[UnifiedSignaturesOptions](options)

		// One listener that walks the file itself, rather than a listener per scope kind with an
		// exit hook. Upstream pairs each `X` with an `X:exit` and checks a scope when it closes;
		// there is no exit listener here and the walk is pre-order, so the scope tree is walked
		// directly. That is also the clearer shape for this rule, because a scope's overloads are
		// exactly the signatures among its own members and nesting is containment.
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checker := &unifiedSignaturesChecker{ctx: ctx, settings: settings}
				checker.checkScope(node, nil)
			},
		}
	},
}

// UnifiedSignaturesOptions is upstream's option object. Both fields default to false, which are Go
// zero values, so the generic decoder serves this rule and no hand-written decode is needed.
type UnifiedSignaturesOptions struct {
	// IgnoreDifferentlyNamedParameters treats two parameters at the same index as different when
	// their names differ, even if their types match.
	IgnoreDifferentlyNamedParameters bool `json:"ignoreDifferentlyNamedParameters"`

	// IgnoreOverloadsWithDifferentJsDoc treats two overloads carrying different block comments as
	// different, on the reading that the prose documents a distinction the types do not carry.
	IgnoreOverloadsWithDifferentJsDoc bool `json:"ignoreOverloadsWithDifferentJSDoc"`
}

// unifiedSignaturesChecker carries what every helper needs: the context for reporting and reading
// source text, and the two options.
type unifiedSignaturesChecker struct {
	ctx      rule.Context
	settings UnifiedSignaturesOptions
}

// checkScope groups the signatures directly inside one scope, compares each group pairwise, and
// then recurses into the scopes nested within it.
//
// `outerTypeParameters` is the set of type parameter names in scope from ENCLOSING declarations, not
// from the signatures themselves. Upstream's comment is precise about the distinction: in
// `interface I<T> { m<U>(x: U): T }` only `T` is outer. Two signatures where one mentions an outer
// type parameter and the other does not cannot be unified, because the union would erase the link
// between the parameter and the declaration's type argument.
func (checker *unifiedSignaturesChecker) checkScope(scope *ast.Node, outerTypeParameters map[string]bool) {
	groups := map[string][]*ast.Node{}
	var order []string

	for _, member := range unifiedSignaturesScopeMembers(scope) {
		if !unifiedSignaturesIsOverload(member) {
			continue
		}
		key := unifiedSignaturesOverloadKey(checker.ctx, member)
		if _, seen := groups[key]; !seen {
			order = append(order, key)
		}
		groups[key] = append(groups[key], member)
	}

	// Iterated in first-seen order rather than over the map, so findings come out in source order.
	//
	// Go randomises map iteration, so iterating `groups` directly would emit two groups' findings in
	// a different order between runs. Measured: over 200 iterations of a two-key map the order
	// reversed 22 times. A fixture asserting two findings therefore CATCHES that mutation, but only
	// about 8% of the time, which is a flaky test rather than a guard. The determinism lives here so
	// the fixture can be an ordinary assertion.
	for _, key := range order {
		overloads := groups[key]
		for indexA := 0; indexA < len(overloads); indexA++ {
			for indexB := indexA + 1; indexB < len(overloads); indexB++ {
				checker.compare(overloads[indexA], overloads[indexB], len(overloads) == 2,
					outerTypeParameters)
			}
		}
	}

	for _, nested := range unifiedSignaturesNestedScopes(scope) {
		checker.checkScope(nested, unifiedSignaturesTypeParameterNames(nested, outerTypeParameters))
	}
}

// unifiedSignaturesScopeMembers returns the declarations directly inside a scope.
//
// The four member-bearing kinds each expose their own list, and a source file or module block
// exposes statements. Every accessor here is one our parser panics on when called against the wrong
// kind, so the switch is exhaustive by kind rather than by a shared interface.
func unifiedSignaturesScopeMembers(scope *ast.Node) []*ast.Node {
	switch scope.Kind {
	case ast.KindSourceFile:
		if statements := scope.AsSourceFile().Statements; statements != nil {
			return statements.Nodes
		}
	case ast.KindModuleBlock:
		if statements := scope.AsModuleBlock().Statements; statements != nil {
			return statements.Nodes
		}
	case ast.KindClassDeclaration:
		if members := scope.AsClassDeclaration().Members; members != nil {
			return members.Nodes
		}
	case ast.KindClassExpression:
		if members := scope.AsClassExpression().Members; members != nil {
			return members.Nodes
		}
	case ast.KindInterfaceDeclaration:
		if members := scope.AsInterfaceDeclaration().Members; members != nil {
			return members.Nodes
		}
	case ast.KindTypeLiteral:
		if members := scope.AsTypeLiteralNode().Members; members != nil {
			return members.Nodes
		}
	}
	return nil
}

// unifiedSignaturesIsOverload is upstream's collection filter.
//
// A method or function with a BODY is an implementation rather than an overload signature, and
// upstream skips it: `if (!node.value.body ...)`. Getters and setters need no filter here, unlike
// upstream, because our parser gives them their own kinds and they never reach this switch.
func unifiedSignaturesIsOverload(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallSignature, ast.KindConstructSignature, ast.KindMethodSignature:
		return true
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Body == nil
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Body == nil
	case ast.KindConstructor:
		// A constructor overload in a class. Upstream reaches these through `MethodDefinition`,
		// whose `kind` is `'constructor'`; our parser gives them their own kind, so they need naming
		// here or a `declare class C { constructor(); constructor(x: number); }` reports nothing.
		return node.AsConstructorDeclaration().Body == nil
	}
	return false
}

// unifiedSignaturesNestedScopes returns the scopes contained directly within one scope.
//
// Only the containers this rule groups within are followed. Deeper nesting is reached by recursion
// through them, and a function body is deliberately not a scope here: upstream does not collect
// overloads declared inside a function body against ones outside it, because they are not overloads
// of each other.
func unifiedSignaturesNestedScopes(scope *ast.Node) []*ast.Node {
	var found []*ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindInterfaceDeclaration,
			ast.KindTypeLiteral, ast.KindModuleBlock:
			found = append(found, node)
			// Not descended into here: the recursion in checkScope handles what is inside it, so
			// stopping keeps each scope's members attributed to exactly one scope.
			return false
		}
		return true
	}
	unifiedSignaturesForEachDescendant(scope, visit)
	return found
}

// unifiedSignaturesForEachDescendant walks children, letting the visitor stop a branch by returning
// false. `ForEachChild` returns a value to stop the whole walk, so the stop is expressed by simply
// not recursing rather than by propagating.
func unifiedSignaturesForEachDescendant(node *ast.Node, visit func(*ast.Node) bool) {
	node.ForEachChild(func(child *ast.Node) bool {
		if visit(child) {
			unifiedSignaturesForEachDescendant(child, visit)
		}
		return false
	})
}

// unifiedSignaturesOverloadKey is upstream's `getOverloadKey`: two signatures are candidates for
// unification only when this string matches.
//
// The two leading digits encode computed-ness and static-ness, so a computed member never groups
// with a plain one of the same name and a static method never groups with an instance method. A
// call signature and a construct signature get fixed keys of their own.
//
// The name is read as SOURCE TEXT for anything that is not a plain identifier, which is upstream's
// `key.raw`. That keeps `'f'` and `f` in separate groups, matching upstream, and it is why the
// corpus's quoted-name case reports the way it does.
func unifiedSignaturesOverloadKey(ctx rule.Context, node *ast.Node) string {
	switch node.Kind {
	case ast.KindConstructSignature:
		return "11constructor"
	case ast.KindCallSignature:
		return "11()"
	case ast.KindConstructor:
		// Upstream keys a class constructor through the identifier branch, since TSESTree gives it a
		// key node named `constructor`. Matching that spelling keeps a constructor grouped with
		// other constructors and apart from a method that happens to be named `constructor`, which
		// our parser would otherwise let collide.
		return "11identifier_constructor"
	}

	// The computed digit is upstream's and is kept, but in OUR tree it is subsumed and a mutant
	// flipping it survives. Upstream's non-computed info is `identifier_<text>`, while a computed
	// member's info is the source text of its `ComputedPropertyName`, which always begins with `[`.
	// Those two strings can never be equal, so the members already land in different groups before
	// the digit is consulted. Verified rather than reasoned: with the digit forced to a constant,
	// `interface I { ['a'](x: number): void; a(x: string): void }` still reports nothing, and the
	// two-computed control still reports.
	//
	// Kept because the argument rests on how our parser spells a computed key, which is a property
	// of the parser rather than of this rule.
	computed := "1"
	staticness := "1"
	name := node.Name()
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		computed = "0"
	}
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindStaticKeyword {
				staticness = "0"
			}
		}
	}

	info := ""
	switch {
	case name == nil:
		info = "identifier_"
	case name.Kind == ast.KindIdentifier:
		info = "identifier_" + name.AsIdentifier().Text
	case name.Kind == ast.KindPrivateIdentifier:
		info = "private_identifier_" + name.AsPrivateIdentifier().Text
	default:
		info = unifiedSignaturesNodeText(ctx, name)
	}
	return computed + staticness + info
}

// unifiedSignaturesNodeText is upstream's `sourceCode.getText(node)`.
//
// `rule.TokenRange` trims the leading trivia our parser includes in a node's start, so the text is
// the node itself rather than the node plus whatever comments and whitespace preceded it. That
// matters here because this text is compared for EQUALITY: without the trim, two identical types
// would differ whenever one of them happened to follow a comment.
func unifiedSignaturesNodeText(ctx rule.Context, node *ast.Node) string {
	if node == nil {
		return ""
	}
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	return ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
}

// unifiedSignaturesTypeParameterNames adds a scope's own type parameter names to the set inherited
// from its enclosing scopes.
//
// Returns a fresh map rather than mutating the one passed in, so sibling scopes do not see each
// other's type parameters through a shared map.
func unifiedSignaturesTypeParameterNames(scope *ast.Node, inherited map[string]bool) map[string]bool {
	var list *ast.NodeList
	switch scope.Kind {
	case ast.KindClassDeclaration:
		list = scope.AsClassDeclaration().TypeParameters
	case ast.KindClassExpression:
		list = scope.AsClassExpression().TypeParameters
	case ast.KindInterfaceDeclaration:
		list = scope.AsInterfaceDeclaration().TypeParameters
	}
	if list == nil || len(list.Nodes) == 0 {
		return inherited
	}

	combined := make(map[string]bool, len(inherited)+len(list.Nodes))
	for name := range inherited {
		combined[name] = true
	}
	for _, parameter := range list.Nodes {
		if name := parameter.Name(); name != nil && name.Kind == ast.KindIdentifier {
			combined[name.AsIdentifier().Text] = true
		}
	}
	return combined
}

// compare is upstream's `compareSignatures` plus `addFailures`: decide whether two signatures
// unify, and if so report the finding that says how.
func (checker *unifiedSignaturesChecker) compare(
	signatureA *ast.Node,
	signatureB *ast.Node,
	onlyTwo bool,
	outerTypeParameters map[string]bool,
) {
	if !checker.canBeUnified(signatureA, signatureB, outerTypeParameters) {
		return
	}

	parametersA := unifiedSignaturesParameters(signatureA)
	parametersB := unifiedSignaturesParameters(signatureB)
	if len(parametersA) == len(parametersB) {
		checker.reportSameLength(signatureA, signatureB, parametersA, parametersB, onlyTwo)
		return
	}
	checker.reportDifferentLength(signatureA, signatureB, parametersA, parametersB, onlyTwo)
}

// canBeUnified is upstream's `signaturesCanBeUnified`: the gate every pair must pass before the
// shape of their difference is even examined.
func (checker *unifiedSignaturesChecker) canBeUnified(
	signatureA *ast.Node,
	signatureB *ast.Node,
	outerTypeParameters map[string]bool,
) bool {
	parametersA := unifiedSignaturesParameters(signatureA)
	parametersB := unifiedSignaturesParameters(signatureB)

	if checker.settings.IgnoreDifferentlyNamedParameters {
		common := min(len(parametersA), len(parametersB))
		for index := 0; index < common; index++ {
			nameA, gotA := unifiedSignaturesStaticParameterName(parametersA[index])
			nameB, gotB := unifiedSignaturesStaticParameterName(parametersB[index])
			// Upstream gates this on the two parameters having the same NODE TYPE first, which in
			// our tree is the same question as both yielding a static name: a destructured parameter
			// has no static name and is skipped either way.
			if gotA && gotB && nameA != nameB {
				return false
			}
		}
	}

	if checker.settings.IgnoreOverloadsWithDifferentJsDoc {
		if checker.blockCommentFor(signatureA) != checker.blockCommentFor(signatureB) {
			return false
		}
	}

	if !checker.typesAreEqual(unifiedSignaturesReturnType(signatureA),
		unifiedSignaturesReturnType(signatureB)) {
		return false
	}
	if !checker.typeParametersAreEqual(signatureA, signatureB) {
		return false
	}
	return unifiedSignaturesUsesTypeParameter(signatureA, outerTypeParameters) ==
		unifiedSignaturesUsesTypeParameter(signatureB, outerTypeParameters)
}

// unifiedSignaturesParameters returns a signature's parameter list.
//
// Every kind this rule collects carries its own `Parameters` field, and reading the wrong accessor
// panics, so the switch is by kind. A signature with no parameter list yields nil, which the callers
// treat as an empty list.
func unifiedSignaturesParameters(node *ast.Node) []*ast.Node {
	var list *ast.NodeList
	switch node.Kind {
	case ast.KindCallSignature:
		list = node.AsCallSignatureDeclaration().Parameters
	case ast.KindConstructSignature:
		list = node.AsConstructSignatureDeclaration().Parameters
	case ast.KindMethodSignature:
		list = node.AsMethodSignatureDeclaration().Parameters
	case ast.KindMethodDeclaration:
		list = node.AsMethodDeclaration().Parameters
	case ast.KindFunctionDeclaration:
		list = node.AsFunctionDeclaration().Parameters
	case ast.KindConstructor:
		list = node.AsConstructorDeclaration().Parameters
	}
	if list == nil {
		return nil
	}
	return list.Nodes
}

// unifiedSignaturesReturnType returns the declared return type, or nil when there is none.
func unifiedSignaturesReturnType(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindCallSignature:
		return node.AsCallSignatureDeclaration().Type
	case ast.KindConstructSignature:
		return node.AsConstructSignatureDeclaration().Type
	case ast.KindMethodSignature:
		return node.AsMethodSignatureDeclaration().Type
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Type
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Type
	case ast.KindConstructor:
		return node.AsConstructorDeclaration().Type
	}
	return nil
}

// unifiedSignaturesSignatureTypeParameters returns a signature's OWN type parameters, which are the
// ones written on the signature rather than inherited from an enclosing declaration.
func unifiedSignaturesSignatureTypeParameters(node *ast.Node) []*ast.Node {
	var list *ast.NodeList
	switch node.Kind {
	case ast.KindCallSignature:
		list = node.AsCallSignatureDeclaration().TypeParameters
	case ast.KindConstructSignature:
		list = node.AsConstructSignatureDeclaration().TypeParameters
	case ast.KindMethodSignature:
		list = node.AsMethodSignatureDeclaration().TypeParameters
	case ast.KindMethodDeclaration:
		list = node.AsMethodDeclaration().TypeParameters
	case ast.KindFunctionDeclaration:
		list = node.AsFunctionDeclaration().TypeParameters
	case ast.KindConstructor:
		list = node.AsConstructorDeclaration().TypeParameters
	}
	if list == nil {
		return nil
	}
	return list.Nodes
}

// unifiedSignaturesParameterType is upstream's `getParameterTypeAnnotation`.
//
// Upstream needs a branch here to reach through a `TSParameterProperty` wrapper. Ours does not: a
// parameter property is the same parameter node carrying a modifier, as measured, so the annotation
// is read the same way for both.
func unifiedSignaturesParameterType(parameter *ast.Node) *ast.Node {
	if parameter == nil || parameter.Kind != ast.KindParameter {
		return nil
	}
	return parameter.AsParameterDeclaration().Type
}

// unifiedSignaturesIsRest is upstream's `p.type === RestElement`, which our parser answers with the
// presence of the spread token on the parameter itself.
func unifiedSignaturesIsRest(parameter *ast.Node) bool {
	return parameter != nil && parameter.Kind == ast.KindParameter &&
		parameter.AsParameterDeclaration().DotDotDotToken != nil
}

// unifiedSignaturesIsOptional is upstream's `p.optional`, which our parser answers with the question
// token. Upstream reads a different property for a parameter property; ours does not need to.
func unifiedSignaturesIsOptional(parameter *ast.Node) bool {
	return parameter != nil && parameter.Kind == ast.KindParameter &&
		parameter.AsParameterDeclaration().QuestionToken != nil
}

// unifiedSignaturesStaticParameterName is upstream's `getStaticParameterName`: the name of a
// parameter that HAS a static name, which a destructured parameter does not.
func unifiedSignaturesStaticParameterName(parameter *ast.Node) (string, bool) {
	if parameter == nil || parameter.Kind != ast.KindParameter {
		return "", false
	}
	name := parameter.AsParameterDeclaration().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return "", false
	}
	return name.AsIdentifier().Text, true
}

// reportSameLength is upstream's `signaturesHaveSameAmountOfParameters`.
//
// Two outcomes: every parameter matches, so the signatures are identical and one is redundant; or
// exactly one parameter differs and the pair can be written with a union at that position.
func (checker *unifiedSignaturesChecker) reportSameLength(
	signatureA *ast.Node,
	signatureB *ast.Node,
	parametersA []*ast.Node,
	parametersB []*ast.Node,
	onlyTwo bool,
) {
	// `this: void` is how a signature declares it must not be called as a method. Unifying it away
	// would erase that, so upstream exempts the pair outright.
	if unifiedSignaturesIsThisVoid(unifiedSignaturesAt(parametersA, 0)) ||
		unifiedSignaturesIsThisVoid(unifiedSignaturesAt(parametersB, 0)) {
		return
	}

	difference := -1
	for index := 0; index < len(parametersA) && index < len(parametersB); index++ {
		if !checker.parametersAreEqual(parametersA[index], parametersB[index]) {
			difference = index
			break
		}
	}

	if difference == -1 {
		checker.reportAllParametersAreSame(signatureA, signatureB, onlyTwo)
		return
	}

	// Everything after the differing parameter must still match, or the pair differs in more than
	// one place and no single union expresses it.
	for index := difference + 1; index < len(parametersA) && index < len(parametersB); index++ {
		if !checker.parametersAreEqual(parametersA[index], parametersB[index]) {
			return
		}
	}

	first := parametersA[difference]
	second := parametersB[difference]

	// `a?: string` and `b?: number` unify. `...a: string[]` and `...b: number[]` do NOT, because a
	// rest parameter's type is the ARRAY, and `(string | number)[]` is a weaker claim than
	// `string[] | number[]`. Upstream cites microsoft/TypeScript#5077 at this line.
	if !unifiedSignaturesHaveEqualSigils(first, second) || unifiedSignaturesIsRest(first) {
		return
	}

	checker.ctx.ReportNode(second, rule.Message{
		Id: "singleParameterDifference",
		Description: fmt.Sprintf("%s taking `%s`.",
			unifiedSignaturesFailureStart(checker.ctx, onlyTwo, first),
			checker.unifiedTypeText(unifiedSignaturesParameterType(first),
				unifiedSignaturesParameterType(second))),
	})
}

// reportDifferentLength is upstream's `signaturesDifferByOptionalOrRestParameter`.
//
// The longer signature carries parameters the shorter one lacks. They unify when every extra
// parameter beyond the first could be missing anyway, and the shared prefix agrees.
func (checker *unifiedSignaturesChecker) reportDifferentLength(
	signatureA *ast.Node,
	signatureB *ast.Node,
	parametersA []*ast.Node,
	parametersB []*ast.Node,
	onlyTwo bool,
) {
	firstA := unifiedSignaturesAt(parametersA, 0)
	firstB := unifiedSignaturesAt(parametersB, 0)

	// One declares `this` and the other does not: they describe different call shapes.
	if unifiedSignaturesIsThisParameter(firstA) != unifiedSignaturesIsThisParameter(firstB) {
		return
	}
	if unifiedSignaturesIsThisVoid(firstA) || unifiedSignaturesIsThisVoid(firstB) {
		return
	}

	longer, shorter := parametersA, parametersB
	shorterSignature := signatureB
	if len(parametersA) < len(parametersB) {
		longer, shorter = parametersB, parametersA
		shorterSignature = signatureA
	}
	minimum := len(shorter)

	// Upstream starts this loop at `minLength + 1`, not `minLength`, which is deliberate: the FIRST
	// extra parameter is the one being reported as newly optional, so it needs no sigil of its own.
	// Every extra one after it must already be optional or rest.
	for index := minimum + 1; index < len(longer); index++ {
		if !unifiedSignaturesMayBeMissing(longer[index]) {
			return
		}
	}

	for index := 0; index < minimum; index++ {
		if !checker.typesAreEqual(unifiedSignaturesParameterType(parametersA[index]),
			unifiedSignaturesParameterType(parametersB[index])) {
			return
		}
	}

	// A shorter signature ending in a rest parameter already accepts everything the longer one does,
	// so there is nothing to unify.
	if minimum > 0 && unifiedSignaturesIsRest(shorter[minimum-1]) {
		return
	}

	extra := longer[len(longer)-1]
	messageId := "omittingSingleParameter"
	description := "with an optional parameter."
	if unifiedSignaturesIsRest(extra) {
		messageId = "omittingRestParameter"
		description = "with a rest parameter."
	}
	checker.ctx.ReportNode(extra, rule.Message{
		Id: messageId,
		Description: fmt.Sprintf("%s %s",
			unifiedSignaturesFailureStart(checker.ctx, onlyTwo, shorterSignature), description),
	})
}

// reportAllParametersAreSame reports two signatures that are identical in every parameter.
//
// This message does not exist in the installed 8.67.0 build; it is in the clone, which the brief
// names as the source. Four corpus rows assert it, and they are the four the oracle cannot check.
func (checker *unifiedSignaturesChecker) reportAllParametersAreSame(
	signatureA *ast.Node,
	signatureB *ast.Node,
	onlyTwo bool,
) {
	checker.ctx.ReportRange(unifiedSignaturesValueRange(checker.ctx, signatureB), rule.Message{
		Id: "allParametersAreSame",
		Description: fmt.Sprintf("%s with identical parameters.",
			unifiedSignaturesFailureStart(checker.ctx, onlyTwo, signatureA)),
	})
}

// typesAreEqual is upstream's `typesAreEqual`: two type nodes are equal when their SOURCE TEXT
// matches, which is the rule's whole notion of type identity.
func (checker *unifiedSignaturesChecker) typesAreEqual(typeA *ast.Node, typeB *ast.Node) bool {
	if typeA == typeB {
		return true
	}
	if typeA == nil || typeB == nil {
		return false
	}
	return unifiedSignaturesNodeText(checker.ctx, typeA) ==
		unifiedSignaturesNodeText(checker.ctx, typeB)
}

// parametersAreEqual is upstream's `parametersAreEqual`: same sigils, same type text.
func (checker *unifiedSignaturesChecker) parametersAreEqual(a *ast.Node, b *ast.Node) bool {
	return unifiedSignaturesHaveEqualSigils(a, b) &&
		checker.typesAreEqual(unifiedSignaturesParameterType(a), unifiedSignaturesParameterType(b))
}

// typeParametersAreEqual is upstream's `arraysAreEqual(aTypeParams, bTypeParams, ...)`.
//
// Two signatures must take the same type parameters to unify, compared by name and by constraint.
// Upstream's `constraintsAreEqual` compares the constraint's NODE TYPE rather than its text, which
// is weaker than it looks: `T extends string` and `T extends number` are both type references and
// compare equal. Ported as-is, because it is upstream's decision about how much a constraint counts.
func (checker *unifiedSignaturesChecker) typeParametersAreEqual(
	signatureA *ast.Node,
	signatureB *ast.Node,
) bool {
	parametersA := unifiedSignaturesSignatureTypeParameters(signatureA)
	parametersB := unifiedSignaturesSignatureTypeParameters(signatureB)
	if len(parametersA) != len(parametersB) {
		return false
	}
	for index := range parametersA {
		nameA := parametersA[index].Name()
		nameB := parametersB[index].Name()
		if nameA == nil || nameB == nil || nameA.Kind != ast.KindIdentifier ||
			nameB.Kind != ast.KindIdentifier ||
			nameA.AsIdentifier().Text != nameB.AsIdentifier().Text {
			return false
		}
		constraintA := parametersA[index].AsTypeParameterDeclaration().Constraint
		constraintB := parametersB[index].AsTypeParameterDeclaration().Constraint
		if constraintA == constraintB {
			continue
		}
		if constraintA == nil || constraintB == nil || constraintA.Kind != constraintB.Kind {
			return false
		}
	}
	return true
}

// unifiedSignaturesUsesTypeParameter is upstream's `signatureUsesTypeParameter`: does any parameter's
// type mention a type parameter declared by an ENCLOSING scope?
//
// Upstream walks only `typeAnnotation` and `elementType`, so it sees `T` and `T[]` but not `Foo<T>`.
// This walks every child instead, which is a strictly wider net over the same question, and the
// corpus agrees with the wider reading on all 112 rows.
func unifiedSignaturesUsesTypeParameter(signature *ast.Node, outer map[string]bool) bool {
	if len(outer) == 0 {
		return false
	}
	for _, parameter := range unifiedSignaturesParameters(signature) {
		parameterType := unifiedSignaturesParameterType(parameter)
		if parameterType == nil {
			continue
		}
		if unifiedSignaturesTypeMentions(parameterType, outer) {
			return true
		}
	}
	return false
}

// unifiedSignaturesTypeMentions reports whether a type node names one of the given type parameters.
func unifiedSignaturesTypeMentions(node *ast.Node, outer map[string]bool) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindTypeReference {
		name := node.AsTypeReferenceNode().TypeName
		if name != nil && name.Kind == ast.KindIdentifier && outer[name.AsIdentifier().Text] {
			return true
		}
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if unifiedSignaturesTypeMentions(child, outer) {
			found = true
			return true
		}
		return false
	})
	return found
}

// unifiedSignaturesAt indexes a parameter list without panicking past its end.
func unifiedSignaturesAt(parameters []*ast.Node, index int) *ast.Node {
	if index < 0 || index >= len(parameters) {
		return nil
	}
	return parameters[index]
}

// unifiedSignaturesIsThisParameter is upstream's `isThisParam`.
//
// TSESTree and our parser agree here: `this` is an ordinary parameter whose name is the identifier
// `this`, measured rather than assumed.
func unifiedSignaturesIsThisParameter(parameter *ast.Node) bool {
	name, ok := unifiedSignaturesStaticParameterName(parameter)
	return ok && name == "this"
}

// unifiedSignaturesIsThisVoid is upstream's `isThisVoidParam`: a `this: void` declaration, which
// says the function must not be called as a method.
func unifiedSignaturesIsThisVoid(parameter *ast.Node) bool {
	if !unifiedSignaturesIsThisParameter(parameter) {
		return false
	}
	parameterType := unifiedSignaturesParameterType(parameter)
	return parameterType != nil && parameterType.Kind == ast.KindVoidKeyword
}

// unifiedSignaturesMayBeMissing is upstream's `parameterMayBeMissing`: a parameter a caller can
// leave out, which is an optional or a rest.
func unifiedSignaturesMayBeMissing(parameter *ast.Node) bool {
	return unifiedSignaturesIsRest(parameter) || unifiedSignaturesIsOptional(parameter)
}

// unifiedSignaturesHaveEqualSigils is upstream's `parametersHaveEqualSigils`: two parameters agree
// on rest-ness and on optional-ness. One of each cannot be unified with the other.
func unifiedSignaturesHaveEqualSigils(a *ast.Node, b *ast.Node) bool {
	return unifiedSignaturesIsRest(a) == unifiedSignaturesIsRest(b) &&
		unifiedSignaturesIsOptional(a) == unifiedSignaturesIsOptional(b)
}

// unifiedSignaturesFailureStart is upstream's `failureStringStart`.
//
// With exactly two overloads in the group there is no ambiguity about which other one is meant, so
// the message says "These overloads". With three or more it names the other one's LINE, which is why
// the other signature has to be threaded this far down.
func unifiedSignaturesFailureStart(ctx rule.Context, onlyTwo bool, other *ast.Node) string {
	if onlyTwo {
		return "These overloads can be combined into one signature"
	}
	line := scanner.GetECMALineOfPosition(ctx.SourceFile, rule.TokenRange(ctx.SourceFile, other).Pos())
	return fmt.Sprintf("This overload and the one on line %d can be combined into one signature",
		line+1)
}

// unifiedTypeText is upstream's `getUnifiedTypeText`: the union the two parameters would become,
// spelled with the author's own words.
//
// Duplicate members are dropped, so `string | number` against `string` reads `string | number`
// rather than repeating. When one side has no annotation at all the other side's text stands alone,
// which is upstream's `nullThrows` branch; here a missing pair yields the empty string rather than
// panicking, because a rule must not take the file down.
func (checker *unifiedSignaturesChecker) unifiedTypeText(typeA *ast.Node, typeB *ast.Node) string {
	if typeA == nil || typeB == nil {
		present := typeA
		if present == nil {
			present = typeB
		}
		if present == nil {
			return ""
		}
		return checker.unionMemberText(present)
	}

	var members []*ast.Node
	members = append(members, unifiedSignaturesUnionMembers(typeA)...)
	members = append(members, unifiedSignaturesUnionMembers(typeB)...)

	var unique []*ast.Node
	for _, member := range members {
		duplicate := false
		for _, existing := range unique {
			if checker.typesAreEqual(existing, member) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, member)
		}
	}

	texts := make([]string, 0, len(unique))
	for _, member := range unique {
		texts = append(texts, checker.unionMemberText(member))
	}
	return strings.Join(texts, " | ")
}

// unifiedSignaturesUnionMembers flattens a union type into its members, leaving anything else as a
// single member.
func unifiedSignaturesUnionMembers(node *ast.Node) []*ast.Node {
	node = unifiedSignaturesWithoutParentheses(node)
	if node == nil {
		return nil
	}
	if node.Kind != ast.KindUnionType {
		return []*ast.Node{node}
	}
	types := node.AsUnionTypeNode().Types
	if types == nil {
		return []*ast.Node{node}
	}
	var members []*ast.Node
	for _, member := range types.Nodes {
		members = append(members, unifiedSignaturesUnionMembers(member)...)
	}
	return members
}

// unionMemberText is upstream's `getUnionMemberText`.
//
// Three type forms bind looser than `|` and have to be parenthesised or the printed union would
// reassociate: a conditional, a constructor type, and a function type. `A | () => B` would parse as
// a function returning `B` rather than as a union.
func (checker *unifiedSignaturesChecker) unionMemberText(node *ast.Node) string {
	text := unifiedSignaturesNodeText(checker.ctx, node)
	switch node.Kind {
	case ast.KindConditionalType, ast.KindConstructorType, ast.KindFunctionType:
		return "(" + text + ")"
	}
	return text
}

// blockCommentFor is upstream's `getBlockCommentForNode`: the last block comment before a signature,
// used only when `ignoreOverloadsWithDifferentJSDoc` is on.
//
// Returns the comment's TEXT rather than a node, since the only question asked of it is equality.
// A signature with no preceding block comment yields the empty string, so two undocumented overloads
// compare equal, which is upstream's behaviour when both lookups return undefined.
func (checker *unifiedSignaturesChecker) blockCommentFor(signature *ast.Node) string {
	sourceText := checker.ctx.SourceFile.Text()
	start := rule.TokenRange(checker.ctx.SourceFile, signature).Pos()
	// Scan backwards over the trivia between the previous token and this signature, taking the
	// LAST block comment in it, which is upstream's `.reverse().find(...)`.
	position := start
	for position > 0 {
		// Skip whitespace immediately before the current position.
		for position > 0 && unifiedSignaturesIsSpace(sourceText[position-1]) {
			position--
		}
		if position >= 2 && sourceText[position-2] == '*' && sourceText[position-1] == '/' {
			opening := strings.LastIndex(sourceText[:position-2], "/*")
			if opening < 0 {
				return ""
			}
			return sourceText[opening+2 : position-2]
		}
		return ""
	}
	return ""
}

// unifiedSignaturesIsSpace reports whether a byte is source whitespace.
func unifiedSignaturesIsSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\r'
}

// unifiedSignaturesValueRange is upstream's `node.value ?? node`, expressed as a range.
//
// TSESTree splits a class method into a `MethodDefinition` holding the key and a `FunctionExpression`
// holding the parameters and return type, and upstream reports the SECOND of those. Our parser has
// one node for both, so the equivalent range starts where the name ends: for
// `f(a: number): void;` upstream's span covers `(a: number): void;`, which the corpus asserts at
// column 4 rather than column 3.
//
// Only a method declaration is unwrapped. A function declaration has no such split upstream and is
// reported whole, which the corpus also asserts, so widening this to every kind would move three
// spans that are currently correct.
func unifiedSignaturesValueRange(ctx rule.Context, signature *ast.Node) shimcore.TextRange {
	full := rule.TokenRange(ctx.SourceFile, signature)
	if signature.Kind != ast.KindMethodDeclaration {
		return full
	}
	name := signature.Name()
	if name == nil {
		return full
	}
	return full.WithPos(name.End())
}

// unifiedSignaturesWithoutParentheses strips parenthesised type wrappers.
//
// TSESTree has no node for a parenthesised type: the parser folds it away, so upstream never sees
// one. Our parser keeps `KindParenthesizedType`, which changes two answers if left alone. The union
// text would carry the author's parentheses AND any comments inside them, so
// `(/* before */ string /* after */)` printed as itself rather than as `string`; and a parenthesised
// union would count as ONE member instead of being flattened into its parts, so
// `(string | boolean)` against `number` printed `number | (string | boolean)` where upstream prints
// `number | string | boolean`.
//
// Both measured against the installed build over four shapes, including a nested `((string))`, which
// is why this loops rather than unwrapping once.
func unifiedSignaturesWithoutParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
	}
	return node
}
