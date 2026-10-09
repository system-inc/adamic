package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoMisusedNewInterfaceConstruct = rule.Message{
	Id: "interfaceConstruct",
	Description: "This construct signature returns the interface that declares it, which describes " +
		"a constructor whose instances are more constructors. An interface describing a static " +
		"class object should have its `new` signature return the instance type being constructed, " +
		"not the constructor interface itself.",
}

var messageNoMisusedNewInterfaceConstructor = rule.Message{
	Id: "interfaceConstructor",
	Description: "An interface cannot be constructed, so a member named `constructor` declares an " +
		"ordinary method that happens to share a name with the class keyword and constructs " +
		"nothing. Use a `new` construct signature if a constructor is meant, and rename the method " +
		"otherwise.",
}

var messageNoMisusedNewClassNew = rule.Message{
	Id: "classNew",
	Description: "This method is named `new` and declares that it returns the class it belongs to, " +
		"which reads as a constructor and is not one. Calling it does not construct anything, so " +
		"the name misleads every reader; a constructor is spelled `constructor`.",
}

// NoMisusedNew flags declarations that look like constructors and are not.
//
//	valid:   class C { constructor() {} }
//	valid:   interface I { new (): C; }
//	valid:   type T = { new (): T };
//	valid:   declare class C { new(): D; }
//	invalid: interface I { new (): I; }
//	invalid: interface I { constructor(): void; }
//	invalid: declare class C { new(): C; }
//
// Ported from oxc's `typescript/no-misused-new`, which is itself a port of
// `@typescript-eslint/no-misused-new`. No option surface: the oxc rule file carries no
// `from_configuration` and no schema, and none of its nineteen corpus cases supplies a second
// tuple element.
//
// # Three judgments, not one
//
// The name suggests a single question and the rule asks three, each anchored on a different node
// and each with its own message. They are worth naming separately because two of them reach type
// literals and one does not, which is the single most counterintuitive thing here.
//
//	a construct signature whose return type names the interface declaring it
//	    anchored on the INTERFACE, so a type literal is never visited
//	a method signature named `constructor`
//	    anchored on the SIGNATURE, so a type literal IS visited
//	a bodyless class method named `new` whose return type names the class
//	    anchored on the CLASS, and only a named one
//
// The corpus pins the asymmetry directly and it reads as a contradiction until the anchors are
// laid out: `type T = { new (): T };` is a PASS while `type T = { constructor(): void;};` is a
// FAIL. Both are members of a type literal. The first is unreachable because the construct-
// signature arm only ever walks an interface's own body, and the second is reachable because the
// `constructor` arm listens for the signature wherever it lives. Reproducing this faithfully means
// resisting the urge to make the two arms symmetric.
//
// # The return type is compared by NAME TEXT, never by resolution
//
// oxc reads the written identifier out of the type annotation and compares its text to the
// declaration's own name. There is no symbol lookup anywhere in the rule, so this needs no type
// checker, and more importantly a port that reached for one would diverge on inputs the corpus
// never writes.
//
// Measured against the oxc release binary, both directions:
//
//	declare class C { new(): D; }  +  type D = C;   SILENT   (resolves to C, text is D)
//	interface Outer {}
//	namespace M { interface Outer { new (): Outer; } }      REPORTS (text matches, symbol does not)
//
// The first is the one that matters: `D` resolves to the enclosing class and a resolving port
// reports it. Both are pinned as fixtures.
//
// Type arguments are read past for the same reason. `declare class C { new(): C<T>; }` reports,
// because only the bare identifier `C` is compared and the arguments are never examined. A
// qualified name is a different node shape rather than an identifier, so `interface I { new (): I.K; }`
// declines, as does a parenthesized `(C)`.
//
// # A getter is a MethodDefinition upstream and its own kind here
//
// oxc's class arm iterates `ClassElement::MethodDefinition`, and in its abstract syntax tree a
// getter is a MethodDefinition carrying `kind: Get`. Our parser gives a getter `KindGetAccessor`,
// distinct from `KindMethodDeclaration`. So a port that listened only for method declarations
// would go silent on a getter named `new`, and no corpus case could see it: the corpus's only
// getter is `get new()` with no return type, which passes under either reading.
//
// Measured: `declare abstract class C { get new(): C; }` REPORTS. The getter is handled.
//
// A setter is also a MethodDefinition upstream and is handled here too, but it can never match,
// because a setter has no return type annotation to compare. That is upstream's behavior rather
// than an omission, and the fixture records it.
//
// # The `constructor` key must be a static identifier
//
// oxc destructures `PropertyKey::StaticIdentifier`, which is the identifier form alone. A string
// literal key and a computed key both carry the text `constructor` and both decline upstream,
// measured on the release binary.
//
// This is the one place `ast.TryGetTextOfPropertyName` is the wrong tool despite being the shelf's
// recommended answer for reading a property key. Probed on our own parse, it answers
// `("constructor", true)` for the identifier key, the string literal key, AND the computed key, so
// building on it over-reports on two inputs the corpus does not write. The kind is checked
// directly instead.
//
// # Where the findings point
//
// Neither arm reports the node its listener receives, so the spans are load-bearing and are pinned
// by their own test.
//
//	construct signature   the `new` keyword: three bytes from the signature's first token
//	method signature      the whole `constructor` key
//	class method          the whole `new` key
//
// oxc writes the first as `Span::sized(sig.span.start, 3)`, taking on faith that a construct
// signature begins with `new`. It does: the parser only produces one when it sees that keyword.
var NoMisusedNew = rule.Rule{
	Name: "@typescript-eslint/no-misused-new",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			// The construct-signature arm walks the interface's own members rather than listening
			// for the signature, which is what confines it to interfaces and leaves type literals
			// alone. Listening for KindConstructSignature directly would be the natural shape and
			// would report `type T = { new (): T };`, which upstream passes.
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				declaration := node.AsInterfaceDeclaration()
				if declaration.Name() == nil {
					return
				}
				interfaceName := declaration.Name().Text()

				for _, member := range declaration.Members.Nodes {
					if member.Kind != ast.KindConstructSignature {
						continue
					}
					returnTypeName, ok := returnTypeIdentifierText(
						member.AsConstructSignatureDeclaration().Type)
					if !ok || returnTypeName != interfaceName {
						continue
					}
					ctx.ReportRange(newKeywordRange(ctx, member),
						messageNoMisusedNewInterfaceConstruct)
				}
			},

			// The `constructor` arm listens for the signature itself, so it reaches an interface
			// body and a type literal alike. That is upstream's shape and it is the reason the two
			// arms disagree about type literals.
			ast.KindMethodSignature: func(node *ast.Node) {
				name := node.AsMethodSignatureDeclaration().Name()
				if name == nil {
					return
				}
				// A static identifier only. A string-literal key and a computed key both hold the
				// text `constructor` and both decline upstream, so the text cannot be read before
				// the kind is checked.
				if name.Kind != ast.KindIdentifier || name.Text() != "constructor" {
					return
				}
				ctx.ReportNode(name, messageNoMisusedNewInterfaceConstructor)
			},

			// The class arm anchors on the declaration and walks its members, and it requires a
			// name: upstream returns early on a class with no id, which is what makes
			// `const foo = class { new(): X;};` and `export default class { constructor(); }`
			// unreachable rather than merely non-matching.
			ast.KindClassDeclaration: func(node *ast.Node) {
				reportClassMethodsNamedNew(ctx, node.AsClassDeclaration().Name(),
					node.AsClassDeclaration().Members)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				reportClassMethodsNamedNew(ctx, node.AsClassExpression().Name(),
					node.AsClassExpression().Members)
			},
		}
	},
}

// reportClassMethodsNamedNew flags each bodyless member named `new` whose return type names the
// class, and is shared by the declaration and expression forms.
//
// A class expression can carry a name (`const foo = class C { new(): C; }`), and upstream's arm is
// anchored on `AstKind::Class`, which covers both forms. Splitting them here is a consequence of
// our parser giving them separate kinds, not a difference in what is decided.
func reportClassMethodsNamedNew(ctx rule.Context, className *ast.Node, members *ast.NodeList) {
	// Upstream returns early when the class has no id, so an anonymous class is exempt entirely
	// rather than merely failing the name comparison. The distinction is invisible in the corpus
	// but real: an anonymous class has no name for any return type to match, so both readings
	// agree, and reproducing the early return keeps the reasoning honest.
	if className == nil || members == nil {
		return
	}
	name := className.Text()

	for _, member := range members.Nodes {
		memberName, returnType, hasBody, ok := classMemberParts(member)
		if !ok || memberName == nil {
			continue
		}
		// The key must be a plain identifier reading exactly `new`. Upstream's
		// `is_specific_id("new")` matches the identifier form alone, same as the `constructor`
		// arm above.
		if memberName.Kind != ast.KindIdentifier || memberName.Text() != "new" {
			continue
		}
		// A member with a body is a real method rather than a declaration, and upstream requires
		// `method.value.body.is_none()`. This is what keeps `class C { new() {} }` clean.
		if hasBody {
			continue
		}
		returnTypeName, hasReturnTypeName := returnTypeIdentifierText(returnType)
		if !hasReturnTypeName || returnTypeName != name {
			continue
		}
		ctx.ReportNode(memberName, messageNoMisusedNewClassNew)
	}
}

// classMemberParts pulls the key, the return type annotation, and whether a body is present out of
// a class member, for the member kinds upstream reaches through `ClassElement::MethodDefinition`.
//
// Upstream's abstract syntax tree folds a method, a getter, and a setter into one MethodDefinition
// carrying a `kind` discriminant, and its match names the whole node rather than the discriminant,
// so all three are visited there. Only two are named here, and the missing one is a real decision
// rather than an oversight.
//
// # Why a setter is deliberately absent, and how that was established
//
// A setter cannot legally carry a return type annotation, so the obvious reading is that a setter
// arm would be inert and including it would cost nothing. That reading is wrong in our tree.
// Probed on our own parse, `declare class C { set new(v: C): C; }` produces a SetAccessor whose
// `Type` is a live KindTypeReference: the annotation is illegal, so it is a parse error, and
// typescript-go recovers by attaching the node anyway. Every other condition of the class arm is
// satisfied by that input, so a setter arm reports it.
//
// Upstream does not. oxc's parser refuses to attach a return type to a setter at all, so its
// `method.value.return_type` is None and `get_return_type_identifier` yields nothing. The rule
// never gets the chance to compare.
//
// This was measured rather than reasoned, because the release binary cannot answer it: the input
// raises TS1095 from the type-error channel, which suppresses rule output on that file, so a probe
// there returns a silence that means nothing. The case was added to oxc's own pass vector and
// `cargo test -p oxc_linter --lib no_misused_new` run: the only diagnostic emitted was TS1095 from
// the parser and `no-misused-new` reported nothing. oxc's tree was restored and verified clean
// afterwards.
//
// So the setter arm is omitted to reproduce upstream's silence. It was present in the first draft
// of this rule and it did report that input, which is a divergence no corpus case could have shown:
// upstream writes no setter with a return type, because upstream's parser cannot represent one.
func classMemberParts(member *ast.Node) (name *ast.Node, returnType *ast.Node, hasBody bool, ok bool) {
	switch member.Kind {
	case ast.KindMethodDeclaration:
		declaration := member.AsMethodDeclaration()
		return declaration.Name(), declaration.Type, declaration.Body != nil, true
	case ast.KindGetAccessor:
		declaration := member.AsGetAccessorDeclaration()
		return declaration.Name(), declaration.Type, declaration.Body != nil, true
	}
	return nil, nil, false, false
}

// returnTypeIdentifierText reads the name text out of a return type annotation, when and only when
// the annotation is a type reference whose name is a bare identifier.
//
// This is upstream's `get_return_type_identifier`, and the narrowness is the whole point. Anything
// that is not `TSTypeReference` holding a `TSTypeName::IdentifierReference` yields nothing, so an
// array type, a literal type, an object type, a parenthesized type, and a qualified name all
// decline. Type arguments are not examined, so `C<T>` yields `C` and compares equal to `C`.
func returnTypeIdentifierText(returnType *ast.Node) (string, bool) {
	if returnType == nil || returnType.Kind != ast.KindTypeReference {
		return "", false
	}
	typeName := returnType.AsTypeReferenceNode().TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier {
		return "", false
	}
	return typeName.Text(), true
}

// newKeywordRange is the span of the `new` keyword that opens a construct signature.
//
// Upstream writes `Span::sized(sig.span.start, 3)`, which relies on a construct signature always
// beginning with that keyword. The parser only produces one when it sees `new`, so the assumption
// holds, but the node's own start includes leading trivia here where upstream's does not, so the
// three bytes have to be measured from the first token rather than from the node position.
func newKeywordRange(ctx rule.Context, signature *ast.Node) core.TextRange {
	tokenRange := rule.TokenRange(ctx.SourceFile, signature)
	return tokenRange.WithEnd(tokenRange.Pos() + len("new"))
}
