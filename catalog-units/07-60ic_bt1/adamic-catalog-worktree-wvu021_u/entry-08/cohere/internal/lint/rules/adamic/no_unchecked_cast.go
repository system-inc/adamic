package adamic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// noUncheckedCastText is the rule's message, whose wording lives in `policy/messages/no-unchecked-cast.json`.
var noUncheckedCastText = policy.MessageOf("adamic/no-unchecked-cast", "uncheckedCast")

// CheckedDowncastNote is the note a cast Adamic compiles to a runtime check records: counted under
// --coverage, never a finding, so readiness does not fail on it.
const CheckedDowncastNote = "checked downcast"

/*
 * NoUncheckedCast reports a cast the runtime could not check (#drbrp8c).
 *
 *     invalid: load() as User                  unknown to an interface: no tag to confirm it
 *     invalid: value as unknown as User        the outer cast is the unchecked one
 *     invalid: JSON.parse(text) as User        any, cast to anything, is a guess
 *     valid:   shape as Circle                 a union member with a discriminant: a checked downcast
 *     valid:   dog as Animal                   an upcast; invariant-mutable judges what it widens
 *     valid:   [1, 2] as const
 *
 * # What is allowed, and what Adamic does with it
 *
 * - `as const`, and a cast to the same type.
 * - An upcast: the source is assignable to the target. Nothing is claimed that tsc did not prove, and
 *   flow.Walker offers it to invariant-mutable, nominal-class and no-optional-widening as a site.
 * - A checked downcast: the source is a union, the target keeps some of its members, and the runtime can
 *   tell every kept member from every dropped one, by a property whose unit literal differs between them,
 *   by typeof, or by instanceof. Adamic compiles it to a check that panics (adamic docs/0.1.md, decision
 *   5), so it is no hole. It is counted as a note, CheckedDowncastNote, and not reported.
 *
 * Everything else is reported, `as unknown as T` and `as NonNullable<T>` included. A source typed `any`
 * is assignable to every target, so it is caught before the upcast test: the cast is the only claim the
 * value's type ever gets.
 *
 * The checked forms are this rule's judgment for now. #bw3xg7c decides Adamic's escape hatches against
 * the readiness survey's counts, and may widen or narrow them; checkedDowncast is the one place to change.
 *
 * # No fix
 *
 * The repair is where the value comes from, which the cast hides.
 */
var NoUncheckedCast = rule.Rule{
	Name:             "adamic/no-unchecked-cast",
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return nil
		}
		typeChecker := ctx.TypeChecker
		judge := func(node *ast.Node) {
			typeNode := node.Type()
			if typeNode == nil || ast.IsConstTypeReference(typeNode) {
				return
			}
			source := typeChecker.GetTypeAtLocation(node.Expression())
			target := checker.Checker_getTypeFromTypeNode(typeChecker, typeNode)
			if source == nil || target == nil || source == target {
				return
			}
			// An `any` anywhere in the source (`any`, `any[]`, `Promise<any>`) is assignable to everything, so
			// the upcast test would pass it: the cast is the only claim the value's type gets, and nothing
			// checks it. The consumers showed `result.rows as RowType[]` on an `any[]` passing every rule.
			if !type_checking.CanBeUnsafeAssignment(typeChecker, source) {
				if checker.Checker_isTypeAssignableTo(typeChecker, source, target) {
					return
				}
				if checkedDowncast(typeChecker, source, target) {
					ctx.Note(CheckedDowncastNote)
					return
				}
			} else if target.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || type_checking.IsTypeUnknownArrayType(target, typeChecker) {
				return
			}
			ctx.ReportNode(node, rule.Message{
				Id: "uncheckedCast",
				Description: noUncheckedCastText.Render(map[string]string{
					"source": typeChecker.TypeToString(source),
					"target": typeChecker.TypeToString(target),
				}),
			})
		}
		return rule.Listeners{
			ast.KindAsExpression:            judge,
			ast.KindTypeAssertionExpression: judge,
		}
	},
}

// checkedDowncast is a cast from a union to some of its members that the runtime can check.
func checkedDowncast(typeChecker *checker.Checker, source *checker.Type, target *checker.Type) bool {
	if source.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	targetMembers := []*checker.Type{target}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		targetMembers = target.Types()
	}
	var kept, dropped []*checker.Type
	for _, member := range source.Types() {
		if containsIdentical(typeChecker, targetMembers, member) {
			kept = append(kept, member)
		} else {
			dropped = append(dropped, member)
		}
	}
	// Every target member must be a source member, or the cast claims something the union never held.
	if len(kept) != len(targetMembers) || len(kept) == 0 {
		return false
	}
	for _, keep := range kept {
		for _, drop := range dropped {
			if !distinguishable(typeChecker, keep, drop) {
				return false
			}
		}
	}
	return true
}

func containsIdentical(typeChecker *checker.Checker, types []*checker.Type, wanted *checker.Type) bool {
	for _, candidate := range types {
		if candidate == wanted || checker.Checker_isTypeIdenticalTo(typeChecker, candidate, wanted) {
			return true
		}
	}
	return false
}

// distinguishable is two union members a runtime check can tell apart: by typeof, by a property whose
// unit literal differs, or by instanceof a class one is and the other is not.
func distinguishable(typeChecker *checker.Checker, left *checker.Type, right *checker.Type) bool {
	if typeofClass(left) != typeofClass(right) {
		return true
	}
	if typeofClass(left) != "object" {
		// Two members of one primitive class: literals of it, told apart by ===.
		return isUnitLiteral(left) && isUnitLiteral(right)
	}
	if isClassInstance(left) && !isClassInstance(right) || !isClassInstance(left) && isClassInstance(right) {
		return true
	}
	if isClassInstance(left) && isClassInstance(right) && classSymbol(left) != classSymbol(right) &&
		!derivesFrom(typeChecker, declaredClassType(typeChecker, left), classSymbol(right), map[*ast.Symbol]bool{}) &&
		!derivesFrom(typeChecker, declaredClassType(typeChecker, right), classSymbol(left), map[*ast.Symbol]bool{}) {
		return true
	}
	for _, property := range checker.Checker_getPropertiesOfType(typeChecker, left) {
		other := checker.Checker_getPropertyOfType(typeChecker, right, property.Name)
		if other == nil {
			continue
		}
		leftType := checker.Checker_getTypeOfSymbol(typeChecker, property)
		rightType := checker.Checker_getTypeOfSymbol(typeChecker, other)
		if isUnitLiteral(leftType) && isUnitLiteral(rightType) &&
			!checker.Checker_isTypeIdenticalTo(typeChecker, leftType, rightType) {
			return true
		}
	}
	return false
}

// typeofClass is what `typeof` would say of a value of the type, with null told apart from object.
func typeofClass(t *checker.Type) string {
	flags := t.Flags()
	switch {
	case flags&checker.TypeFlagsStringLike != 0:
		return "string"
	case flags&checker.TypeFlagsNumberLike != 0:
		return "number"
	case flags&checker.TypeFlagsBooleanLike != 0:
		return "boolean"
	case flags&checker.TypeFlagsBigIntLike != 0:
		return "bigint"
	case flags&checker.TypeFlagsESSymbolLike != 0:
		return "symbol"
	case flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0:
		return "undefined"
	case flags&checker.TypeFlagsNull != 0:
		return "null"
	}
	return "object"
}

func isUnitLiteral(t *checker.Type) bool {
	return t.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral|
		checker.TypeFlagsEnumLiteral|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
}
