// Package typecheck answers questions about types that only the type checker can settle.
//
// It is the shelf's type-aware layer, and until it landed there was no such layer: no package under
// `internal/utilities/` imported `shim/checker` at all. Every type-aware rule needs the same handful of
// primitives -- what a union is made of, whether a type is thenable, whether a symbol came from the
// default library -- and before this they were reachable only through a path named `upstream`.
//
// # Provenance
//
// These helpers are vendored from tsgolint (https://github.com/typescript-eslint/tsgolint), MIT
// licensed, `Copyright (c) 2025 typescript-eslint and other contributors`. The full text sits in
// `LICENSE` beside this file, and each file names the tsgolint source it came from.
//
// tsgolint is a dead project and we will never re-sync, so the `upstream/` path these lived under
// was actively misleading: it told a reader to expect a sync that cannot happen and to treat edits
// as merge hazards. This code is ours to maintain now, and the next change to one of these helpers
// is an ordinary edit. The vendored commit was `4178710` (2025-07-13), recorded here because it is
// the provenance a reader wants rather than a baseline anyone will diff against.
//
// # What was changed from tsgolint
//
// Import paths, and nothing else about the logic. Upstream vendored against `microsoft/typescript-go`
// and we are on `microsoft/TypeScript` at `tsc/`, which renamed part of the surface: `ast.IsParameter`
// became `IsParameterDeclaration` and `AsTypeParameter` became `AsTypeParameterDeclaration`. Those
// renames are the only edits inside a function body. A reader who expects byte-identical files needs
// this note.
//
// # Why it compiles against our shim
//
// Every `github.com/microsoft/TypeScript/tsc/shim/...` path is `replace`-directed in `go.mod` to
// `./shim/...`. There is exactly one shim tree in this repo and it is ours, so `ast.Node` and
// `checker.Checker` are not merely the same shape across two trees, they are the same package. A
// mismatch fails loudly at compile rather than quietly at runtime.

// Source: tsgolint `internal/utilities/ts_api_utils.go`.
//
// The type predicates: what a type's flags say, what a union or intersection is made of, and the
// signature and thenable questions that read the checker to answer.
package type_checking

import (
	"iter"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

func UnionTypeParts(t *checker.Type) []*checker.Type {
	if IsUnionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}

// UnionTypePartsSeq yields what UnionTypeParts returns, for a caller that only loops over it, as
// strings.SplitSeq does for strings.Split.
//
// UnionTypeParts builds a one-element slice for every type that is not a union, and that slice
// reaches the heap: the function is over the inlining budget (cost 88 of 80, the call to Types alone
// is most of it), so the compiler cannot see the caller drop it. That was 1.05M objects on a cold ahra
// run, three quarters of them from strict-void-return (#rwsffzm). This one is small enough to inline,
// so a range loop over it allocates nothing, which TestUnionTypePartsSeqAllocatesNothing holds.
func UnionTypePartsSeq(t *checker.Type) iter.Seq[*checker.Type] {
	return func(yield func(*checker.Type) bool) {
		if !IsUnionType(t) {
			yield(t)
			return
		}
		for _, part := range t.Types() {
			if !yield(part) {
				return
			}
		}
	}
}
func IntersectionTypeParts(t *checker.Type) []*checker.Type {
	if IsIntersectionType(t) {
		return t.Types()
	}
	return []*checker.Type{t}
}

func IsTypeFlagSet(t *checker.Type, flags checker.TypeFlags) bool {
	return t != nil && checker.Type_flags(t)&flags != 0
}

func IsIntrinsicType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntrinsic)
}

func IsIntrinsicErrorType(t *checker.Type) bool {
	return IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "error"
}

func IsIntrinsicVoidType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsVoid)
}

func IsUnionType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUnion)
}
func IsIntersectionType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsIntersection)
}
func IsTypeAnyType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsAny)
}
func IsTypeUnknownType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsUnknown)
}
func IsObjectType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsObject)
}
func IsTypeParameter(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsTypeParameter)
}
func IsBooleanLiteralType(t *checker.Type) bool {
	return IsTypeFlagSet(t, checker.TypeFlagsBoolean)
}
func IsTrueLiteralType(t *checker.Type) bool {
	return IsBooleanLiteralType(t) && IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "true"
}
func IsFalseLiteralType(t *checker.Type) bool {
	return IsBooleanLiteralType(t) && IsIntrinsicType(t) && t.AsIntrinsicType().IntrinsicName() == "false"
}

func GetCallSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindCall)
}
func GetConstructSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindConstruct)
}

// ex. getCallSignaturesOfType
func CollectAllCallSignatures(typeChecker *checker.Checker, t *checker.Type) []*checker.Signature {
	if IsUnionType(t) {
		signatures := []*checker.Signature{}
		for _, subtype := range t.Types() {
			signatures = append(signatures, GetCallSignatures(typeChecker, subtype)...)
		}
		return signatures
	}
	if IsIntersectionType(t) {
		var signatures []*checker.Signature
		for _, subtype := range t.Types() {
			sig := GetCallSignatures(typeChecker, subtype)
			if len(sig) != 0 {
				if signatures != nil {
					return []*checker.Signature{}
				}
				signatures = sig
			}
		}
		if signatures == nil {
			return []*checker.Signature{}
		}
		return signatures
	}
	return checker.Checker_getSignaturesOfType(typeChecker, t, checker.SignatureKindCall)
}

func IsSymbolFlagSet(symbol *ast.Symbol, flag ast.SymbolFlags) bool {
	return symbol != nil && symbol.Flags&flag != 0
}

func IsCallback(
	typeChecker *checker.Checker,
	param *ast.Symbol,
	node *ast.Node,
) bool {
	t := checker.Checker_getApparentType(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(param, node))

	if param.ValueDeclaration != nil && ast.IsParameterDeclaration(param.ValueDeclaration) && param.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
		t = checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
		if t == nil {
			return false
		}
	}

	for subType := range UnionTypePartsSeq(t) {
		if len(GetCallSignatures(typeChecker, subType)) != 0 {
			return true
		}
	}

	return false
}

// TODO(note): why there is no IntersectionTypeParts
func IsThenableType(
	typeChecker *checker.Checker,
	node *ast.Node,
	t *checker.Type,
) bool {
	if t == nil {
		t = typeChecker.GetTypeAtLocation(node)
	}
	for typePart := range UnionTypePartsSeq(checker.Checker_getApparentType(typeChecker, t)) {
		then := checker.Checker_getPropertyOfType(typeChecker, typePart, "then")
		if then == nil {
			continue
		}

		thenType := typeChecker.GetTypeOfSymbolAtLocation(then, node)

		for subTypePart := range UnionTypePartsSeq(thenType) {
			for _, signature := range checker.Checker_getSignaturesOfType(typeChecker, subTypePart, checker.SignatureKindCall) {
				if len(checker.Signature_parameters(signature)) != 0 && IsCallback(typeChecker, checker.Signature_parameters(signature)[0], node) {
					return true
				}
			}
		}
	}
	return false
}

func GetWellKnownSymbolPropertyOfType(t *checker.Type, name string, typeChecker *checker.Checker) *ast.Symbol {
	return checker.Checker_getPropertyOfType(typeChecker, t, checker.Checker_getPropertyNameForKnownSymbolName(typeChecker, name))
}

// IsStrictCompilerOptionEnabled reports whether a strict-family option is on, as the compiler resolves
// it: an explicit value wins, and an unset one follows `strict`, which is itself on unless set false.
//
// That last clause is TypeScript 6's default, and it is what ts-api-utils 2.5 reads too
// (`options[option] ?? options.strict ?? isTsVersionAtLeast(6)`), so upstream's rules see it. This
// helper used to read an unset `strict` as off, TypeScript 5's default. Structure's tsconfig never
// writes `strict`, so in ahra and www every caller believed `strictNullChecks` and `noImplicitThis`
// were off. no-useless-default-assignment then declined every file there while it fired in api,
// whose Base tsconfig writes `strict: true` (#6ar414z). The compiler's own GetStrictOptionValue is
// the resolution the checker uses, so asking it keeps a rule's premise and the types it reads in step.
//
// Upstream also ties `strictPropertyInitialization` to `strictNullChecks`. No caller asks about that
// option, so the clause is not reproduced.
func IsStrictCompilerOptionEnabled(
	options *core.CompilerOptions,
	option core.Tristate,
) bool {
	return options.GetStrictOptionValue(option)
}
