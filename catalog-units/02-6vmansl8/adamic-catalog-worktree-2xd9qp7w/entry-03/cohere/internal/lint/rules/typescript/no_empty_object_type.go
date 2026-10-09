package typescript

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageEmptyInterface = rule.Message{
	Id: "noEmptyInterface",
	Description: "An interface with no members allows any non-nullish value, so it constrains " +
		"nothing: every object, every string, every number satisfies it. It reads like a type and " +
		"behaves like `unknown` with better manners. Give it members, or use the type it actually " +
		"means.",
}

var messageEmptyInterfaceWithSuper = rule.Message{
	Id: "noEmptyInterfaceWithSuper",
	Description: "An interface that declares no members and extends one type is that type under a " +
		"second name: every value of the supertype satisfies it, and it asks nothing more. It reads " +
		"like a new shape and adds none. Use a type alias for the supertype, or give the interface " +
		"members of its own.",
}

var messageEmptyObject = rule.Message{
	Id: "noEmptyObject",
	Description: "`{}` does not mean an empty object. It means any value that is not null or " +
		"undefined, so a string and a number both satisfy it. That is almost never what the author " +
		"intended, and the two readings are impossible to tell apart at a glance. Use `object` for " +
		"any non-primitive, or `unknown` for genuinely any value.",
}

// NoEmptyObjectTypeOptions are upstream's three options, under upstream's names.
//
// AllowInterfaces is "always", "never" or "with-single-extends", and AllowObjectTypes "always" or
// "never"; the config layer checks both against upstream's schema before the decoder sees them.
// AllowWithName is a JavaScript regular expression, compiled with the `u` flag as upstream's
// `new RegExp(allowWithName, 'u')` is, and tested as a search: `Base` exempts `XBaseY`.
type NoEmptyObjectTypeOptions struct {
	AllowInterfaces  string `json:"allowInterfaces"`
	AllowObjectTypes string `json:"allowObjectTypes"`
	AllowWithName    string `json:"allowWithName,omitempty"`

	allowWithNameTester *esregexp.RegExp
}

// DefaultNoEmptyObjectTypeOptions is the configuration upstream applies when the rule is written bare.
func DefaultNoEmptyObjectTypeOptions() NoEmptyObjectTypeOptions {
	return NoEmptyObjectTypeOptions{AllowInterfaces: "never", AllowObjectTypes: "never"}
}

// noEmptyObjectTypeWire is the shape the config layer delivers, before defaults are applied.
type noEmptyObjectTypeWire struct {
	AllowInterfaces  *string `json:"allowInterfaces"`
	AllowObjectTypes *string `json:"allowObjectTypes"`
	AllowWithName    *string `json:"allowWithName"`
}

// DecodeNoEmptyObjectTypeOptions reads this rule's configuration from the config layer.
//
// An allowWithName JavaScript will not compile is an error, as upstream's `new RegExp` throws when
// the rule is created. An empty one is no pattern at all, as upstream's falsy test reads it.
func DecodeNoEmptyObjectTypeOptions(raw []byte) (any, error) {
	options := DefaultNoEmptyObjectTypeOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var wire noEmptyObjectTypeWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	if wire.AllowInterfaces != nil {
		switch *wire.AllowInterfaces {
		case "always", "never", "with-single-extends":
			options.AllowInterfaces = *wire.AllowInterfaces
		default:
			return options, fmt.Errorf("allowInterfaces is %q, and upstream takes always, never or with-single-extends", *wire.AllowInterfaces)
		}
	}
	if wire.AllowObjectTypes != nil {
		switch *wire.AllowObjectTypes {
		case "always", "never":
			options.AllowObjectTypes = *wire.AllowObjectTypes
		default:
			return options, fmt.Errorf("allowObjectTypes is %q, and upstream takes always or never", *wire.AllowObjectTypes)
		}
	}
	if wire.AllowWithName != nil && *wire.AllowWithName != "" {
		tester, err := esregexp.Compile(*wire.AllowWithName, "u")
		if err != nil {
			return options, fmt.Errorf("allowWithName %q is not a JavaScript regular expression under the u flag: %w", *wire.AllowWithName, err)
		}
		options.AllowWithName = *wire.AllowWithName
		options.allowWithNameTester = tester
	}
	return options, nil
}

// NoEmptyObjectType flags an interface with no members and the `{}` type literal.
//
//	valid:   interface Thing { name: string }
//	valid:   interface Both extends Base, Derived {}
//	valid:   type Thing = Base & {}
//	invalid: interface Thing {}                   noEmptyInterface
//	invalid: interface Thing extends Base {}      noEmptyInterfaceWithSuper
//	invalid: type Thing = {}                      noEmptyObject
//	invalid: let value: {}                        noEmptyObject
//
// Ported from typescript-eslint 8.71.0, whose rows and edge rows TestNoEmptyObjectTypeUpstreamCorpus
// replays; no_empty_object_type_corpus_data_test.go says how. It began as a port of oxc's rule, which
// diverged from upstream in how a finding looks though never in what it reports: the whole declaration
// as the span, one message for both interface shapes, no interface suggestions, and its own ids.
// Ruled on #e1zk9s0 to take upstream's.
//
// Two boundaries decide what reports. An interface extending two or more types is allowed even when
// empty, because that is how a reader names an intersection of interfaces; one extending exactly one is
// the supertype under a second name. And `{}` directly inside an intersection is allowed: `T & {}` is
// the idiom that strips null and undefined from T, and the empty half is doing real work there.
//
// The repairs are suggestions, never fixes, since `object` and `unknown` mean different things and
// only the author knows which was meant. An empty interface is rewritten as a type alias of `object`
// or `unknown`, or of its one supertype. Upstream withholds that rewrite in two places, and so does
// this: from `export default interface`, which a type alias cannot spell, and from an interface that
// merges with a class or another interface of its name, where the alias would collide. The report
// itself still stands in both.
//
// "Merges" is upstream's scope question, a same-named class or interface among the declarations of
// the scope the interface sits in, so it is read from syntax: the statement list holding the
// interface. Not from the checker, whose merged symbol also holds the lib's declarations and other
// files', which upstream never sees. An interface with type parameters has a scope of its own to
// upstream, the type scope its parameters live in, so its question finds nothing and it is always
// offered the rewrite; kept as upstream's, and measured.
var NoEmptyObjectType = rule.Rule{
	Name: "@typescript-eslint/no-empty-object-type",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoEmptyObjectTypeOptions](options)
		if !ok {
			settings = DefaultNoEmptyObjectTypeOptions()
		}

		listeners := rule.Listeners{}
		if settings.AllowInterfaces != "always" {
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				checkEmptyInterface(ctx, settings, node)
			}
		}
		if settings.AllowObjectTypes != "always" {
			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				checkEmptyTypeLiteral(ctx, settings, node)
			}
		}
		return listeners
	},
}

// checkEmptyInterface reports an interface with no members, unless its extends clause or its name
// exempts it.
func checkEmptyInterface(ctx rule.Context, settings NoEmptyObjectTypeOptions, node *ast.Node) {
	declaration := node.AsInterfaceDeclaration()
	name := declaration.Name()
	if name == nil {
		return
	}
	if settings.allowWithNameTester != nil && settings.allowWithNameTester.Test(name.Text()) {
		return
	}
	if declaration.Members != nil && len(declaration.Members.Nodes) != 0 {
		return
	}
	extended := heritageTypes(declaration.HeritageClauses)
	if len(extended) > 1 || (len(extended) == 1 && settings.AllowInterfaces == "with-single-extends") {
		return
	}

	shouldSuggest := !ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) && !emptyInterfaceMergesInItsScope(node)

	if len(extended) == 0 {
		if !shouldSuggest {
			ctx.ReportNode(name, messageEmptyInterface)
			return
		}
		ctx.ReportNodeWithSuggestions(name, messageEmptyInterface,
			emptyInterfaceRewrite(ctx, node, "object", rule.Message{
				Id:          "replaceEmptyInterface",
				Description: "Replace the empty interface with `object`, which means any non-primitive.",
			}),
			emptyInterfaceRewrite(ctx, node, "unknown", rule.Message{
				Id:          "replaceEmptyInterface",
				Description: "Replace the empty interface with `unknown`, which means any value at all.",
			}),
		)
		return
	}

	if !shouldSuggest {
		ctx.ReportNode(name, messageEmptyInterfaceWithSuper)
		return
	}
	supertype := rule.TokenRange(ctx.SourceFile, extended[0])
	ctx.ReportNodeWithSuggestions(name, messageEmptyInterfaceWithSuper,
		emptyInterfaceRewrite(ctx, node, ctx.SourceFile.Text()[supertype.Pos():supertype.End()], rule.Message{
			Id:          "replaceEmptyInterfaceWithSuper",
			Description: "Replace the empty interface with a type alias of the type it extends.",
		}),
	)
}

// emptyInterfaceRewrite is the suggestion that rewrites an empty interface as `type Name<T> = target`.
//
// Upstream replaces its TSInterfaceDeclaration node, which an `export` wraps rather than starts, so an
// `export` stays in front of the alias. A `declare` is part of the node and goes with it: `export declare
// interface X {}` becomes `export type X = object`. Measured on the installed rule.
func emptyInterfaceRewrite(ctx rule.Context, node *ast.Node, target string, message rule.Message) rule.Suggestion {
	text := ctx.SourceFile.Text()
	declaration := node.AsInterfaceDeclaration()
	name := declaration.Name()
	typeParameters := ""
	if declaration.TypeParameters != nil {
		// The list's own range sits inside the angle brackets, and upstream's text includes them.
		opening := scanner.SkipTrivia(text, name.End())
		closing := scanner.SkipTrivia(text, declaration.TypeParameters.End())
		typeParameters = text[opening : closing+1]
	}
	replacement := core.NewTextRange(emptyInterfaceRewriteStart(ctx, node), node.End())
	return rule.Suggestion{
		Message: message,
		Fixes: []rule.Fix{rule.ReplaceRange(replacement,
			"type "+name.Text()+typeParameters+" = "+target)},
	}
}

// emptyInterfaceRewriteStart is where upstream's interface node begins: its first modifier that is not
// `export` or `default`, else the `interface` keyword.
func emptyInterfaceRewriteStart(ctx rule.Context, node *ast.Node) int {
	modifiers := node.Modifiers()
	if modifiers == nil || len(modifiers.Nodes) == 0 {
		return rule.TokenRange(ctx.SourceFile, node).Pos()
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind != ast.KindExportKeyword && modifier.Kind != ast.KindDefaultKeyword {
			return rule.TokenRange(ctx.SourceFile, modifier).Pos()
		}
	}
	return scanner.SkipTrivia(ctx.SourceFile.Text(), modifiers.Nodes[len(modifiers.Nodes)-1].End())
}

// emptyInterfaceMergesInItsScope answers upstream's mergedWithOtherDeclaration: whether a class or
// another interface of the same name is declared in the scope the interface sits in.
//
// That scope's declarations are the statement list holding the interface, a `case` clause's being the
// whole `switch`'s. An interface with type parameters is its own scope to upstream, where nothing else is
// declared, so it never merges; see the rule's doc comment.
func emptyInterfaceMergesInItsScope(node *ast.Node) bool {
	declaration := node.AsInterfaceDeclaration()
	if declaration.TypeParameters != nil && len(declaration.TypeParameters.Nodes) > 0 {
		return false
	}
	name := declaration.Name().Text()
	for _, statement := range scopeStatementsOf(node) {
		if statement == node {
			continue
		}
		if statement.Kind != ast.KindClassDeclaration && statement.Kind != ast.KindInterfaceDeclaration {
			continue
		}
		if other := statement.Name(); other != nil && other.Kind == ast.KindIdentifier && other.Text() == name {
			return true
		}
	}
	return false
}

// scopeStatementsOf returns the statements declared in the same scope as a statement.
func scopeStatementsOf(statement *ast.Node) []*ast.Node {
	parent := statement.Parent
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindSourceFile:
		return parent.AsSourceFile().Statements.Nodes
	case ast.KindBlock:
		return parent.AsBlock().Statements.Nodes
	case ast.KindModuleBlock:
		return parent.AsModuleBlock().Statements.Nodes
	case ast.KindCaseClause, ast.KindDefaultClause:
		var statements []*ast.Node
		if parent.Parent != nil && parent.Parent.Kind == ast.KindCaseBlock {
			for _, clause := range parent.Parent.AsCaseBlock().Clauses.Nodes {
				statements = append(statements, clause.AsCaseOrDefaultClause().Statements.Nodes...)
			}
		}
		return statements
	}
	return nil
}

// checkEmptyTypeLiteral reports a `{}` type literal, unless it sits in an intersection or names a type
// alias allowWithName exempts.
//
// Our parser keeps a parenthesized type, which upstream's folds away, so `T & ({})` is still in an
// intersection and `type BaseProps = ({})` still names the alias; the parentheses are looked through.
func checkEmptyTypeLiteral(ctx rule.Context, settings NoEmptyObjectTypeOptions, node *ast.Node) {
	typeLiteral := node.AsTypeLiteralNode()
	if typeLiteral.Members != nil && len(typeLiteral.Members.Nodes) != 0 {
		return
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedType {
		parent = parent.Parent
	}
	if parent != nil && parent.Kind == ast.KindIntersectionType {
		return
	}
	if settings.allowWithNameTester != nil && parent != nil && parent.Kind == ast.KindTypeAliasDeclaration &&
		settings.allowWithNameTester.Test(parent.Name().Text()) {
		return
	}

	ctx.ReportNodeWithSuggestions(node, messageEmptyObject,
		rule.Suggestion{
			Message: rule.Message{
				Id:          "replaceEmptyObjectType",
				Description: "Replace `{}` with `object`, which means any non-primitive.",
			},
			Fixes: []rule.Fix{ctx.ReplaceNode(node, "object")},
		},
		rule.Suggestion{
			Message: rule.Message{
				Id:          "replaceEmptyObjectType",
				Description: "Replace `{}` with `unknown`, which means any value at all.",
			},
			Fixes: []rule.Fix{ctx.ReplaceNode(node, "unknown")},
		},
	)
}

// heritageTypes returns the types an interface extends, across every heritage clause.
//
// The count is of types rather than clauses: an interface has one `extends` clause holding a list, so
// counting clauses would answer one for `extends Base, Derived` and miss the distinction the rule turns
// on.
func heritageTypes(clauses *ast.NodeList) []*ast.Node {
	if clauses == nil {
		return nil
	}
	var types []*ast.Node
	for _, clause := range clauses.Nodes {
		heritage := clause.AsHeritageClause()
		if heritage == nil || heritage.Types == nil {
			continue
		}
		types = append(types, heritage.Types.Nodes...)
	}
	return types
}
