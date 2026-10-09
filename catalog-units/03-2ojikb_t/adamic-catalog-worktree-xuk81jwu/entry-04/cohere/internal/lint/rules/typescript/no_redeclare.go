package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoRedeclare = rule.Message{
	Id: "redeclared",
	Description: "This name is already declared in the same scope. The second declaration does not " +
		"create a second binding, it overwrites what the name reaches, so every use of the name " +
		"after this point gets whatever the last declaration bound and the earlier one becomes " +
		"unreachable without ever being deleted. The two spellings that look intentional are the " +
		"dangerous ones: a value and a type sharing a name compile fine and diverge later, and a " +
		"second class silently replaces the first for every `new`, every `extends`, and every " +
		"`instanceof` in the file. Rename one of them, or delete the declaration that is dead.",
}

var messageNoRedeclareAsBuiltin = rule.Message{
	Id: "redeclaredAsBuiltin",
	Description: "This name is already declared by the standard library this program compiles against, " +
		"and this file is a script, so its top-level declarations land in the same global scope " +
		"the library's do. The declaration does not make a new name, it collides with the " +
		"built-in one: a type merges into it or conflicts with it, and a value replaces it for " +
		"every script on the page that reads the global. Rename it, or make the file a module " +
		"with an import or export so its declarations stop being globals.",
}

// NoRedeclareOptions tunes whether a script may redeclare the standard library's globals, and
// whether TypeScript's legitimate declaration merges are exempt.
//
// The wire fields are pointers because both defaults are TRUE. A plain bool cannot tell an absent
// key from an explicit `false`, and a zero-valued struct would turn every merge exemption off, which
// silently converts thirteen of upstream's clean cases into findings.
type NoRedeclareOptions struct {
	BuiltinGlobals         bool
	IgnoreDeclarationMerge bool
}

// noRedeclareWire is the shape the config layer actually delivers, before defaults are applied.
type noRedeclareWire struct {
	BuiltinGlobals         *bool `json:"builtinGlobals"`
	IgnoreDeclarationMerge *bool `json:"ignoreDeclarationMerge"`
}

// DefaultNoRedeclareOptions is the configuration upstream applies when the rule is written bare.
func DefaultNoRedeclareOptions() NoRedeclareOptions {
	return NoRedeclareOptions{BuiltinGlobals: true, IgnoreDeclarationMerge: true}
}

// DecodeNoRedeclareOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto`, because that helper reports an error on empty
// input and the config layer turns the error into nil, so a rule reached through it can never see
// the difference between "no options were written" and "the option was written as false". For an
// option defaulting to false that distinction is invisible. This one defaults to true, so the
// generic helper's zero value inverts the rule rather than merely weakening it.
func DecodeNoRedeclareOptions(raw []byte) (any, error) {
	options := DefaultNoRedeclareOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var wire noRedeclareWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	if wire.BuiltinGlobals != nil {
		options.BuiltinGlobals = *wire.BuiltinGlobals
	}
	if wire.IgnoreDeclarationMerge != nil {
		options.IgnoreDeclarationMerge = *wire.IgnoreDeclarationMerge
	}
	return options, nil
}

// NoRedeclare flags a name declared more than once in the same scope.
//
//	valid:   var a = 3; var b = function () { var a = 10; };
//	valid:   if (true) { let b = 2; } else { let b = 3; }
//	valid:   function a(): string; function a(): number; function a() {}
//	valid:   interface A {} class A {} namespace A {}
//	valid:   function A<T>() {} interface B<T> {} type C<T> = Array<T>; class D<T> {}
//	invalid: var a = 3; var a = 10;
//	invalid: var a; function a() {}
//	invalid: type T = 1; type T = 2;
//	invalid: type something = string; const something = 2;
//	invalid: class A {} class A {} namespace A {}
//	invalid: var Object = 0;                 in a script, redeclaredAsBuiltin
//	valid:   var Object = 0; export {};      a module's top level is not the global scope
//
// A second declaration of a name does not add a binding, it replaces what the name reaches. In
// plain JavaScript that is a typo with no diagnostic; in TypeScript it is worse, because the
// language deliberately allows some of these pairs to MERGE into one thing and the merging ones are
// spelled exactly like the mistaken ones. `class A {} interface A {}` is one type with extra
// members and is correct. `class A {} class A {}` is one class silently replacing another and is
// not. The whole judgment this rule makes is telling those apart.
//
// # Why this reads the checker
//
// Upstream is a thin caller over `eslint-scope`: it asks `sourceCode.getScope(node)` for a scope,
// walks `scope.variables`, and reports any variable holding more than one declaration identifier.
// The scope partition IS the rule. Without one, `function f1() { var x = 1; } function f2() { var x
// = 2; }` and `var x = 1; var x = 2;` are the same two declarations of the same name and no
// syntactic test separates them.
//
// The equivalent partition our tree already has is TypeScript's own binder. `ast.IsLocalsContainer`
// names the nodes that own a scope and `ast.GetLocals` hands back that scope's symbol table, so
// walking from a declaration to its nearest locals container reproduces exactly the grouping
// upstream's scope manager builds. Measured on ten shapes against the installed rule, including the
// three that must produce no group at all: two function bodies, the arms of an if/else, and four
// generic parameters all named `T` across a function, an interface, a type alias and a class.
//
// # Why the obvious predicate does not work, which is the part worth reading
//
// The reflex is to ask the checker how many declarations a name has and report when it holds more
// than one. That answers wrongly in BOTH directions and a port built on it fails quietly:
//
//	class E {} class E {}                      symbol.Declarations is 1, and this REPORTS
//	interface I {} interface I {}              symbol.Declarations is 2, and this is CLEAN
//
// The two duplicate classes do not share a symbol at all. They are two symbols competing for one
// name, and the binder keeps one of them in the scope table and drops the other, so the count
// collapses to 1 on exactly the input that must report. The legitimately merging pair does share
// one symbol carrying both declarations, so it counts 2 on exactly the input that must not. A
// previous attempt at this rule measured that pair of facts, concluded the substrate was absent,
// and stopped. The measurement was right and the conclusion did not follow: the count is the wrong
// question, not a missing answer.
//
// So this rule never asks the checker for a count. It groups the declaration NODES by scope and
// name from the tree, which cannot lose one, and then applies upstream's merge arithmetic to the
// kinds in the group. The checker is still needed for the scope partition itself, which is what
// `ast.GetLocals` reads.
//
// # The second wrong predicate, recorded because it is the more convincing one
//
// Having found that the count collapses for duplicate classes, the natural next move is to compare
// SYMBOL IDENTITY instead: two duplicate classes resolve to two distinct symbols while a merging
// pair resolves to one shared symbol, which reads as exactly the discrimination this rule wants. It
// is not, and it fails in both directions. Measured over the same 23 shapes, identity disagrees
// with upstream on 7:
//
//	var a = 1; var a = 2;                      ONE shared symbol, and this REPORTS
//	function F() {} function F() {}            ONE shared symbol, and this REPORTS
//	enum H {A} enum H {B}                      ONE shared symbol, and this REPORTS
//	var d = 1; type d = string;                ONE shared symbol, and this REPORTS
//	function f1() { var x = 1; }
//	function f2() { var x = 2; }               TWO symbols, and this is CLEAN
//
// Identity is right for the class and interface merge family and wrong for the plain duplicates,
// which are the most common shapes the rule has to catch. Nothing in this file uses it: the two
// false positives are answered by the scope partition, since a name in two non-overlapping scopes
// is two groups of one rather than one group of two, and the five false negatives are answered by
// the kind arithmetic, since a var beside a var belongs to no merge set. Recorded rather than left
// out because identity is the predicate that looks right, and a later reader simplifying this rule
// toward it would pass the merge fixtures and break `var a = 1; var a = 2;`.
// `TestPartitionAndKindSeparatesAllTwentyThree` asserts all 23 so that edit fails loudly.
//
// # The merge arithmetic, which is upstream's and reproduced rather than reasoned about
//
// With `ignoreDeclarationMerge` on, a group is exempt when it is all interfaces, all namespaces, or
// drawn from one of three merge sets with at most one member from the set's "primary" kind: class
// with interface and namespace, function with namespace, enum with namespace. Past one primary the
// exemption lapses and only the PRIMARY declarations report, which is why `class A {} class A {}
// namespace A {}` reports once rather than twice: the namespace is still a legitimate merge partner
// and is not what went wrong.
//
// Overload signatures are filtered before any of this. Upstream drops `TSDeclareFunction`, its node
// for a function declaration with no body, so `function a(): string; function a(): number; function
// a() {}` is one implementation and two signatures rather than three declarations. Our parser gives
// all three `KindFunctionDeclaration` and separates them by whether a body is present.
//
// # A `var` binds into its function, not its block
//
// A block-scoped declaration groups by its nearest locals container, and a `var` does not: it hoists
// to the enclosing function, namespace, static block or file, which is where upstream's scope manager
// puts it. So `var a; if (x) { var a; }` is one name declared twice, and a `var Object` inside a
// top-level block of a script is a global. The key a `var` gets is the one a `let` written directly
// in that function's body gets, so `var a; let a;` in one body still meets.
//
// Some scopes upstream never looks inside: a namespace body, a class static block and a catch clause.
// Its listeners find scopes by the node that owns them, and none of its eight is the owner of those,
// so a duplicate there is clean to it and to this port. See upstreamVisitsScope.
//
// A destructured name is a declaration too: `var { a = 0, b: Object = 0 } = {};` declares `a` and
// `Object`, and each reports at its own identifier. A parameter's pattern is not walked, since a
// parameter belongs to the function's own scope, as above.
//
// # builtinGlobals: a script's top level redeclaring the standard library
//
// With `builtinGlobals` on, the default, upstream yields a "builtin" declaration first for any name
// in the GLOBAL scope that ESLint knows as a read-only global, and then every declaration of the name
// in the file after it. Two sources make one, and the replay measured both:
//
//   - ESLint's own ECMAScript globals for `ecmaVersion: "latest"`, which its Linter declares in every
//     run whatever the parser: `NaN`, `parseInt`, `toString` and the rest of eslintLatestGlobals.
//   - The TypeScript lib, which typescript-eslint's scope manager seeds the global scope with, as the
//     parser derived it from the program's compiler options: every name a lib file declares as a TYPE
//     (an interface, type alias, class, enum or namespace), so `type NodeListOf = 1;` under lib dom
//     reports. A lib `declare var` alone is a value with no type and the lib generator leaves it out,
//     which is why `var top = 0;` and `var self = 1;` are clean.
//
// Two things decide it, and both are read here rather than configured:
//
//   - The scope is the global scope only when the file is a script. A module's top level is its own
//     scope, so the same declaration there is clean. Script means TypeScript's notion, no import or
//     export (or every file under `moduleDetection: force`), the test `no-implicit-globals` uses. The
//     ESLint twins parse every file as `sourceType: "module"`, where the option never fires; ruled on
//     #e1zk9s0 as the twin's simplification, and a script in house code gets the faithful answer.
//   - The lib half is the program's own lib, read as the global symbol of the name and whether any
//     of its declarations is a type-kind declaration in a default library file. The checker merges
//     the lib files first, and on a conflicting user declaration `mergeSymbol` keeps the lib's
//     symbol, so the lib's declarations are still on the global when the user's is not.
//
// The arithmetic is upstream's yield order. The builtin is first, so EVERY surviving syntax
// declaration reports, each as `redeclaredAsBuiltin`, rather than every one after the first. The
// merge exemption still runs first: two interfaces augmenting a lib interface merge and nothing
// reports, while a lone augmenting `interface Window {}` in a script reports, because upstream only
// applies the exemption to more than one declaration.
//
// What stays out, pinned by name in the upstream replay rather than dropped: ESLint's `globals`
// configuration and `/*global*/` comment directives (cohere carries no globals configuration, by the
// 1.0 contract), and `ecmaFeatures.globalReturn`. Upstream's third message, `redeclaredBySyntax`, is
// reachable only from a `/*global*/` directive, so it is not declared here.
//
// No fix. The repair is a rename or a deletion, and choosing which declaration is the dead one, and
// what to call the survivor, is exactly what the rule cannot know.
var NoRedeclare = rule.Rule{
	Name: "@typescript-eslint/no-redeclare",

	// See the doc above: the scope partition comes from the binder, which the program builds.
	NeedsTypeChecker: true,
	// Contents, not Shapes: builtinGlobals reads a global symbol's declarations, and the rule walks its
	// own file's bodies, which is the pair TestRulesClaimShapesOnlyWhereTheScanAllowsIt refuses. The
	// declarations it reads are the lib's, by kind and file only, but the scan cannot tell them from an
	// imported one, and a refused claim costs re-runs where a wrong one costs a stale verdict.
	//
	// builtinGlobals asks whether a global's declarations live in a lib file.
	ProgramReads: rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoRedeclareOptions](options)
		if !ok {
			// A rule configured as a bare "error" is handed nil, and `options.(T)` on nil yields the
			// zero value, whose false `IgnoreDeclarationMerge` inverts this rule rather than
			// weakening it. Falling back to the real defaults is what keeps the bare configuration,
			// which is how this rule is actually written down, from reporting every legitimate
			// declaration merge in the tree.
			settings = DefaultNoRedeclareOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// The engine hands every rule a nil checker when the program could not be built.
				// This rule reads the binder's scope tables, which only exist on a bound program,
				// so without one there is nothing to group and the honest answer is silence.
				if ctx.TypeChecker == nil {
					return
				}
				reportRedeclarationsIn(ctx, node, settings)
			},
		}
	},
}

// declarationGroup is every declaration of one name inside one scope, in source order.
type declarationGroup struct {
	declarations []*ast.Node
}

// reportRedeclarationsIn walks the file once, partitions every declaration by the scope it binds
// into, and reports the groups upstream would.
//
// One pass over the file rather than a listener per scope-bearing kind. Upstream registers eight
// listeners and asks each for its scope, which is the shape ESLint's API forces; we have the whole
// tree in hand, so the same partition falls out of a single walk that asks each declaration which
// scope owns it. Fidelity is to which inputs report, not to how upstream obtained its scopes.
func reportRedeclarationsIn(ctx rule.Context, sourceFile *ast.Node, settings NoRedeclareOptions) {
	// Keyed by the scope-owning node and the name, so two scopes holding the same name never meet.
	groups := map[scopedName]*declarationGroup{}
	var order []scopedName

	var walk func(*ast.Node)
	walk = func(current *ast.Node) {
		if current == nil {
			return
		}
		if name := redeclarableName(current); name != "" {
			container := nearestScopeContainer(current)
			if bindsIntoFunctionScope(current) {
				container = varScopeContainer(current)
			}
			key := scopedName{container: container, name: name}
			group, seen := groups[key]
			if !seen {
				group = &declarationGroup{}
				groups[key] = group
				order = append(order, key)
			}
			group.declarations = append(group.declarations, current)
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile)

	// Only a script's top level is the global scope, so only its groups can meet a builtin.
	file := sourceFile.AsSourceFile()
	checksBuiltins := settings.BuiltinGlobals && !ast.IsExternalModule(file) && ctx.Program != nil

	// Iterate the recorded order rather than the map. Go randomises map iteration, so reporting
	// straight from it would emit findings in a different order on every run, and the harness
	// asserts findings in order.
	for _, key := range order {
		if !upstreamVisitsScope(key.container) {
			continue
		}
		declarations := survivingDeclarations(groups[key].declarations, settings)
		message := messageNoRedeclare
		if checksBuiltins && key.container == sourceFile && len(declarations) > 0 && isBuiltinGlobal(ctx.TypeChecker, ctx.Program, key.name) {
			// The builtin is the first declaration, so every one in the file redeclares it.
			message = messageNoRedeclareAsBuiltin
		} else if len(declarations) < 2 {
			continue
		} else {
			// The first declaration is the one being redeclared, so it is not itself a finding.
			declarations = declarations[1:]
		}
		for _, declaration := range declarations {
			name := declaration.Name()
			if name == nil {
				continue
			}
			ctx.ReportNode(name, message)
		}
	}
}

// upstreamVisitsScope reports whether upstream ever examines the scope a container stands for.
//
// It finds its scopes from eight listeners: the program, functions, arrow functions, blocks, `for`,
// `for in`, `for of` and `switch`. A namespace body, a class static block and a catch clause own a
// scope that none of those nodes is the block of, so upstream never walks their variables, and two
// `var a` in one namespace, or two `let a` in one static block, are clean to it. Measured against
// the installed rule, which the edge rows record.
func upstreamVisitsScope(container *ast.Node) bool {
	if container == nil {
		return true
	}
	switch container.Kind {
	case ast.KindModuleDeclaration, ast.KindClassStaticBlockDeclaration, ast.KindCatchClause:
		return false
	case ast.KindBlock:
		return container.Parent == nil || container.Parent.Kind != ast.KindClassStaticBlockDeclaration
	}
	return true
}

// scopedName identifies one name inside one scope.
type scopedName struct {
	container *ast.Node
	name      string
}

// survivingDeclarations applies upstream's merge arithmetic to one group and returns the syntax
// declarations upstream yields for it, in source order. With no builtin ahead of them every one after
// the first reports; with one, every one does.
func survivingDeclarations(declarations []*ast.Node, settings NoRedeclareOptions) []*ast.Node {
	// Overload signatures are not declarations for this rule's purposes. Upstream filters its
	// `TSDeclareFunction` nodes before counting, so a function with two signatures and one body is
	// a single declaration rather than three.
	var considered []*ast.Node
	for _, declaration := range declarations {
		if isFunctionOverloadSignature(declaration) {
			continue
		}
		considered = append(considered, declaration)
	}

	// A discrimination since builtinGlobals, where it used to be only a cost guard. Upstream applies
	// the exemption to more than one declaration only, so a lone `interface Window {}` in a script is
	// yielded and reports against the lib's, where the merge sets would call one interface exempt.
	if settings.IgnoreDeclarationMerge && len(considered) > 1 {
		if exempt, primaries := mergeExemption(considered); exempt {
			// Every declaration merges legitimately and nothing is yielded, or more than one
			// declaration of the merge set's primary kind is, and the legitimate merge partners
			// beside them are not.
			return primaries
		}
	}

	return considered
}

// mergeExemption reports whether a group is one of TypeScript's legitimate declaration merges, and
// when the group holds more than one declaration of the merging kind, which ones those are.
//
// The three merge sets and their primary kinds are upstream's, at
// `typescript-eslint/packages/eslint-plugin/src/rules/no-redeclare.ts`. A group qualifies when every
// declaration in it is drawn from one set. If the set's primary kind appears once, the merge is safe
// and nothing reports. If it appears more than once, the extra copies report and the other members
// of the set do not, which is what makes `class A {} class A {} namespace A {}` one finding.
func mergeExemption(declarations []*ast.Node) (bool, []*ast.Node) {
	// Upstream tests all-interfaces and all-namespaces before the three merge sets, and that test is
	// deliberately absent here because the sets already answer it. An all-interface group is drawn
	// from the class set, whose members are class, interface and namespace, and supplies none of its
	// primary, so it takes the `len(primaries) <= 1` exit and comes back exempt. Namespaces are in
	// all three sets and do the same. Measured rather than argued: a mutant neutralising an explicit
	// arm for these two survived the whole corpus, which is what identified it as subsumed rather
	// than as untested, and upstream is silent on three interfaces, three namespaces, and the two
	// mixed, all of which this reaches through the sets.
	for _, set := range mergeSets {
		drawnFromSet := true
		var primaries []*ast.Node
		for _, declaration := range declarations {
			if !set.members[declaration.Kind] {
				drawnFromSet = false
				break
			}
			if declaration.Kind == set.primary {
				primaries = append(primaries, declaration)
			}
		}
		if !drawnFromSet {
			continue
		}
		if len(primaries) <= 1 {
			return true, nil
		}
		return true, primaries
	}

	return false, nil
}

// mergeSet is one of TypeScript's legitimate declaration merges: the kinds that may share a name,
// and the one of them that may only appear once.
type mergeSet struct {
	primary ast.Kind
	members map[ast.Kind]bool
}

// mergeSets is upstream's three sets, in upstream's order.
//
// Order is load-bearing rather than incidental. A group of one class and one namespace is drawn
// from the class set, and a group of one function and one namespace is drawn from the function set;
// nothing is drawn from two, because each set's members differ past the shared namespace. Reading
// them in a fixed order keeps the answer stable regardless.
var mergeSets = []mergeSet{
	{
		primary: ast.KindClassDeclaration,
		members: map[ast.Kind]bool{
			ast.KindClassDeclaration:     true,
			ast.KindInterfaceDeclaration: true,
			ast.KindModuleDeclaration:    true,
		},
	},
	{
		primary: ast.KindFunctionDeclaration,
		members: map[ast.Kind]bool{
			ast.KindFunctionDeclaration: true,
			ast.KindModuleDeclaration:   true,
		},
	},
	{
		primary: ast.KindEnumDeclaration,
		members: map[ast.Kind]bool{
			ast.KindEnumDeclaration:   true,
			ast.KindModuleDeclaration: true,
		},
	},
}

// redeclarableName returns the text of a declaration's name when the node binds that name into an
// enclosing scope, and "" otherwise.
//
// The kinds are the ones that can collide. A parameter and a class member are deliberately absent:
// a parameter belongs to the function's own scope where upstream's scope manager puts it too, and a
// member belongs to the class rather than to any scope this rule partitions. A binding element counts
// when its pattern belongs to a variable declaration, since upstream reports `var { a = 0 } = {};`
// against an earlier `var a;`, and not when it belongs to a parameter.
func redeclarableName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration, ast.KindModuleDeclaration, ast.KindFunctionDeclaration,
		ast.KindVariableDeclaration:
	case ast.KindBindingElement:
		if bindingRoot(node).Kind != ast.KindVariableDeclaration {
			return ""
		}
	default:
		return ""
	}
	name := node.Name()
	if name != nil && ast.IsIdentifier(name) {
		return name.Text()
	}
	return ""
}

// bindingRoot walks up from a binding element through its patterns to the declaration that owns
// them: a variable declaration or a parameter.
func bindingRoot(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil && (current.Kind == ast.KindBindingElement || ast.IsBindingPattern(current)) {
		current = current.Parent
	}
	return current
}

// bindsIntoFunctionScope reports whether a declaration is a `var`, or a name destructured by one,
// which hoists out of the blocks around it.
//
// The test is the declaration's own flags, read through its list and patterns: a `let`, `const` or
// `using` is block-scoped, and so is a catch clause's variable, which TypeScript spells as a variable
// declaration too.
func bindsIntoFunctionScope(node *ast.Node) bool {
	if node.Kind != ast.KindVariableDeclaration && node.Kind != ast.KindBindingElement {
		return false
	}
	return !ast.IsBlockOrCatchScoped(node)
}

// varScopeContainer returns the scope a `var` binds into: the nearest function, static block,
// namespace or file.
//
// The key is the one a block-scoped declaration written directly in that scope's body gets from
// nearestScopeContainer, so the two kinds still meet: a function's body block rather than the
// function, and a namespace itself, since its module block owns no locals.
func varScopeContainer(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch {
		case current.Kind == ast.KindSourceFile, current.Kind == ast.KindModuleDeclaration:
			return current
		case ast.IsFunctionLike(current), current.Kind == ast.KindClassStaticBlockDeclaration:
			if body := current.Body(); body != nil && ast.IsLocalsContainer(body) {
				return body
			}
			return current
		}
	}
	return nil
}

// isFunctionOverloadSignature reports whether a function declaration is a signature with no body.
//
// Upstream's parser gives these their own node type and filters on it. Ours gives every form
// `KindFunctionDeclaration` and distinguishes them by the body, so the test is the body rather than
// the kind. An ambient `declare function` and a plain overload signature both land here, which
// matches upstream: both are `TSDeclareFunction` to it.
//
// Named for functions rather than sharing `isOverloadSignature` with `no-dupe-class-members`, which
// asks the same-sounding question about a class MEMBER and answers for method and accessor kinds
// this rule never sees. One name over two domains would have to widen to cover both and would then
// answer for kinds neither caller wants.
func isFunctionOverloadSignature(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration && node.Body() == nil
}

// nearestScopeContainer walks up to the first ancestor that owns a scope, which is the scope the
// declaration binds into.
//
// A nil container means the declaration sits outside any locals-bearing node, which the tree should
// not produce because the source file itself owns one. It is returned rather than guarded so that
// two such declarations would still group together under the same nil key rather than being
// silently dropped.
func nearestScopeContainer(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		// Only the container test. This once also asked `ast.GetLocals(current) != nil`, which was
		// always true, since GetLocals creates a missing table, and the creating was a write to a
		// node other workers read, a data race.
		if ast.IsLocalsContainer(current) {
			return current
		}
	}
	return nil
}
