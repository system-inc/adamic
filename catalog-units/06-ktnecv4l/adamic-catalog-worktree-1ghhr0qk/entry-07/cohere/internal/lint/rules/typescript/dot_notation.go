package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/dotnotation"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// DotNotationOptions is upstream's option surface: the core rule's two, and three that ask the checker.
//
// `AllowKeywords` is a pointer because its default is TRUE, as in the core rule: a plain bool would
// decode an absent option to false and start reporting every `a.class` in the tree. The other three
// default to false, which is their zero value.
type DotNotationOptions struct {
	// AllowKeywords permits `a.class` and friends. Absent means true, which is upstream's default.
	AllowKeywords *bool `json:"allowKeywords"`

	// AllowPattern is a regular expression source. A computed key matching it is left alone.
	AllowPattern string `json:"allowPattern"`

	// AllowIndexSignaturePropertyAccess leaves alone a bracket access that resolves to no property and
	// reaches a string index signature. A tsconfig setting noPropertyAccessFromIndexSignature turns it
	// on whatever this says, because there the dot form is a type error.
	AllowIndexSignaturePropertyAccess bool `json:"allowIndexSignaturePropertyAccess"`

	// AllowPrivateClassPropertyAccess leaves alone a bracket access to a member declared `private`.
	AllowPrivateClassPropertyAccess bool `json:"allowPrivateClassPropertyAccess"`

	// AllowProtectedClassPropertyAccess leaves alone a bracket access to a member declared `protected`.
	AllowProtectedClassPropertyAccess bool `json:"allowProtectedClassPropertyAccess"`
}

// DotNotation is typescript-eslint's extension of the core `dot-notation`.
//
//	invalid: a['b'];
//	valid:   class X { protected p = 1 } new X()['p'];                  with allowProtectedClassPropertyAccess
//	valid:   class X { private p = 1 } new X()['p'];                    with allowPrivateClassPropertyAccess
//	valid:   declare const t: Record<string, number>; t['row'];         with allowIndexSignaturePropertyAccess
//
// Bracket access is TypeScript's sanctioned way past `private` and `protected`: `x['p']` type-checks
// where `x.p` is TS2445 or TS2341, which is how a test reaches a protected hook. The core rule cannot
// see that, so it proposes a fix that does not compile. api's OrmTrackingEntity.test.ts reaches
// clearChangedFields and afterSave this way (#qv7evaa).
//
// Upstream's wrapper is `baseRule.create(context)` with one question asked first, and only when one
// of the three options is on: does the accessed property's FIRST declaration carry `private` or
// `protected` as its FIRST modifier, or, failing any property, does the object reach a string index
// signature. Everything that reports comes from the core judgment in `ecmascript/dotnotation`, shared
// with `dot-notation`, so this file carries the question and nothing else.
//
// # How the property is found, which decides the corner cases
//
// Upstream asks `getSymbolAtLocation` of the key, and failing that searches the object's non-nullable
// type's properties for a string literal key's text. This port asks the same two questions of the
// same checker, so their answers are upstream's in the shapes that are not obvious, each replayed
// through the installed 8.71.0:
//
//	x[('p')]            the parenthesized key has no symbol, and the search finds it: exempt
//	x[`p`]              a template key has a symbol: exempt
//	x[null], x[true]    neither question finds `private null = 1`: reported
//	(A | B)['v']        with A's `v` private and B's public, neither finds a property: reported
//	@tracked private v  a decorator is not a modifier, so the first modifier is `private`: exempt
//	static v, readonly  the first modifier is neither: reported
//
// The index-signature question needs a key type that is string-like, so `[key: number]` and
// `[key: symbol]` do not exempt, and `any` has no index signature at all.
var DotNotation = rule.Rule{
	Name: "@typescript-eslint/dot-notation",

	// The accessed property's declaration and the object's index signatures are the checker's answers
	NeedsTypeChecker: true,
	// Whether noPropertyAccessFromIndexSignature is on
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		resolved, isDotNotationOptions := rule.OptionsAs[DotNotationOptions](options)
		if !isDotNotationOptions {
			resolved = DefaultDotNotationOptions()
		}
		settings := dotnotation.Settings{AllowKeywords: true, AllowPattern: dotnotation.CompileAllowPattern(resolved.AllowPattern)}
		if resolved.AllowKeywords != nil {
			settings.AllowKeywords = *resolved.AllowKeywords
		}

		allowIndexSignature := resolved.AllowIndexSignaturePropertyAccess ||
			(ctx.Program != nil && ctx.Program.Options().NoPropertyAccessFromIndexSignature.IsTrue())
		if ctx.TypeChecker != nil &&
			(resolved.AllowPrivateClassPropertyAccess || resolved.AllowProtectedClassPropertyAccess || allowIndexSignature) {
			settings.DeclinesComputed = func(access *ast.Node, key *ast.Node) bool {
				object := access.AsElementAccessExpression().Expression
				property := ctx.TypeChecker.GetSymbolAtLocation(key)
				if property == nil && key.Kind == ast.KindStringLiteral {
					objectType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(object))
					for _, candidate := range ctx.TypeChecker.GetPropertiesOfType(objectType) {
						if candidate.Name == key.Text() {
							property = candidate
							break
						}
					}
				}

				switch dotNotationFirstModifier(property) {
				case ast.KindPrivateKeyword:
					if resolved.AllowPrivateClassPropertyAccess {
						return true
					}
				case ast.KindProtectedKeyword:
					if resolved.AllowProtectedClassPropertyAccess {
						return true
					}
				}

				if property == nil && allowIndexSignature {
					objectType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(object))
					for _, indexInfo := range ctx.TypeChecker.GetIndexInfosOfType(objectType) {
						if type_checking.IsTypeFlagSet(indexInfo.KeyType(), checker.TypeFlagsStringLike) {
							return true
						}
					}
				}
				return false
			}
		}
		return dotnotation.Listeners(ctx, settings)
	},
}

// dotNotationFirstModifier is the kind of the first modifier on the property's first declaration,
// upstream's `getModifiers(symbol?.getDeclarations()?.[0])?.[0].kind`, and KindUnknown when there is
// none. Decorators sit in the same list here and are not modifiers there, so they are passed over.
func dotNotationFirstModifier(property *ast.Symbol) ast.Kind {
	if property == nil || len(property.Declarations) == 0 {
		return ast.KindUnknown
	}
	modifiers := property.Declarations[0].Modifiers()
	if modifiers == nil {
		return ast.KindUnknown
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind != ast.KindDecorator {
			return modifier.Kind
		}
	}
	return ast.KindUnknown
}

// DefaultDotNotationOptions is the unconfigured answer: keywords allowed, everything else off.
func DefaultDotNotationOptions() DotNotationOptions {
	allowKeywords := true
	return DotNotationOptions{AllowKeywords: &allowKeywords}
}

// DecodeDotNotationOptions reads this rule's configuration from the config layer.
//
// Hand-rolled for the core rule's reason: `allowKeywords` defaults to true, so decoding an absent
// option into a plain bool would invert the rule, and a bare `"error"` arrives as empty input. Unknown
// keys are refused, as upstream's schema does with additionalProperties false.
func DecodeDotNotationOptions(raw []byte) (any, error) {
	options := DefaultDotNotationOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var decoded DotNotationOptions
	if err := rule.UnmarshalOptions(raw, &decoded); err != nil {
		return options, err
	}
	if decoded.AllowKeywords == nil {
		decoded.AllowKeywords = options.AllowKeywords
	}
	return decoded, nil
}
