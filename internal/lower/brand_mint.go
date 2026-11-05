package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A brand mint is a cast whose target differs from its source only by phantom brand members: tsc's
// `[] as SortedArray<T>`, an empty array claimed sorted. A phantom member has no run-time content,
// so the cast is an erased no-op in .ts (rulings.md:29), and each one gets its own line in the
// checked report, since the claim it makes (sorted, a path, an escaped name) is one no check can
// hold. It's refused in .a for now (Compiler and Adamic, Nov 5, #yth65gz): .a mints come back to
// Adamic as their own question.
//
// "Only by phantom members" is judged by Adamic's relation, never tsc's. The target is an interface
// that extends exactly one base and declares nothing of its own but members phantomMember calls
// phantom, and the source must relate to that base, instantiated as the target is, as an upcast
// would: castAssignable's rules, with a fresh literal judged as freshOrWidened judges one. So `readonly T[]` as a mutable
// SortedArray<T> is refused (the base is T[], which the source doesn't relate to), and so is an
// Animal[] as SortedArray<Dog>. A fresh `[]` relates to any element type, and a mutable T[] relates
// to a mutable base, because the brand carries no layout. Array brands are phantom only while every
// observation of a member's presence is refused (rulings.md:31); Language holds those roads.
type brandMint struct {
	base    *checker.Type
	members []string
}

// brandMintOf is the mint a cast of operand makes, or nil when it isn't one.
func (l *lowering) brandMintOf(operand *ast.Node, source, target *checker.Type) *brandMint {
	// The innermost value is what's judged, so one typed any or unknown has nothing to judge.
	if source.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) {
		return nil
	}
	declared := target
	var arguments []*checker.Type
	if target.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.Target() != nil {
		declared = target.Target()
		arguments = l.checker.GetTypeArguments(target)
	}
	if declared.ObjectFlags()&checker.ObjectFlagsInterface == 0 || declared.AsInterfaceType() == nil {
		return nil
	}
	bases := checker.Checker_getBaseTypes(l.checker, declared)
	if len(bases) != 1 {
		// Several bases, or none, would need a relation per base; a brand tsc writes extends one.
		return nil
	}
	base := bases[0]
	if parameters := declared.AsInterfaceType().TypeParameters(); len(parameters) > 0 {
		if len(arguments) < len(parameters) {
			return nil
		}
		base = instantiateType(l.checker, base, newTypeMapper(parameters, arguments[:len(parameters)]))
	}
	// Nothing of its own but phantom members: the same signatures and index entries as its base,
	// and every property the base lacks is one phantomMember admits.
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if len(l.checker.GetSignaturesOfType(target, kind)) != len(l.checker.GetSignaturesOfType(base, kind)) {
			return nil
		}
	}
	if len(l.checker.GetIndexInfosOfType(target)) != len(l.checker.GetIndexInfosOfType(base)) {
		return nil
	}
	mint := &brandMint{base: base}
	for _, member := range l.checker.GetPropertiesOfType(target) {
		if l.checker.GetPropertyOfType(base, member.Name) != nil {
			continue
		}
		if !l.phantomMember(member) {
			return nil
		}
		mint.members = append(mint.members, member.Name)
	}
	if len(mint.members) == 0 {
		return nil
	}
	// castAssignable's rules, with the widening judged as at any slot: a fresh literal ([]) is held by
	// nothing else, so its own elements are the slot's (freshOrWidened).
	if !l.checker.IsTypeAssignableTo(source, base) || !l.classAssignable(source, base) || !l.enumAssignable(source, base) || l.freshOrWidened(operand, source, base) != nil || !l.castOptionalFields(source, base, map[[2]*checker.Type]bool{}) || l.optionalRelationFailure(source, base, nil, map[[2]*checker.Type]bool{}) != "" {
		return nil
	}
	return mint
}

// mintOperand is what a cast mints from: its operand, or, when the operand is itself a cast to unknown
// (`x as unknown as SortedArray<T>`, the spelling tsc accepts for a required undefined brand member),
// that cast's operand. The hop to unknown is part of the mint and erased with it, so no unknown value
// is ever made; any other hop (`as any`) is judged as its own cast.
func (l *lowering) mintOperand(as *ast.AsExpression) (*ast.Node, *checker.Type) {
	operand := ast.SkipParentheses(as.Expression)
	if operand.Kind == ast.KindAsExpression && l.checker.GetTypeAtLocation(operand).Flags()&checker.TypeFlagsUnknown != 0 {
		inner := operand.AsAsExpression().Expression
		return inner, l.concrete(l.checker.GetTypeAtLocation(inner))
	}
	return as.Expression, l.concrete(l.checker.GetTypeAtLocation(as.Expression))
}

// mintHop is whether node is the `as unknown` hop of a brand mint, which castProof erases with the mint.
func (l *lowering) mintHop(node *ast.Node) bool {
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUnknown == 0 {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindAsExpression {
		return false
	}
	operand, source := l.mintOperand(parent.AsAsExpression())
	return operand != node && l.brandMintOf(operand, source, l.concrete(l.checker.GetTypeAtLocation(parent))) != nil
}

// brandMintValue lowers a mint as the value it mints from, under any `as unknown` hop, which makes no
// value of its own. A literal made right there ([], [a, b]) is made as the brand's base, the slot it's
// written into, so its elements are stored as the base's are. Any other value must already be stored
// as the branded type reads it, or the mint is NotYet: a number[] read as (number | string)[] would read
// its elements by the wrong representation.
func (l *lowering) brandMintValue(node *ast.Node, as *ast.AsExpression, mint *brandMint) (ir.Expression, error) {
	operand, from := l.mintOperand(as)
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	literal := ast.SkipParentheses(operand)
	if literal.Kind == ast.KindArrayLiteralExpression || literal.Kind == ast.KindObjectLiteralExpression {
		l.keepSlot(literal, mint.base)
	} else if !l.sameKeeping(from, target, map[[2]*checker.Type]bool{}) {
		return nil, l.notYet(node, "a brand mint that changes the runtime representation of "+l.checker.TypeToString(from))
	}
	l.recordBrandMint(node, target, mint)
	return l.expression(operand)
}

// brandMintRefusal is a mint written in .a, which .a doesn't admit yet.
func (l *lowering) brandMintRefusal(node *ast.Node, target *checker.Type, mint *brandMint) error {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "a brand mint in .a: a cast to " + l.checker.TypeToString(target) + " that adds only its phantom member " + strings.Join(mint.members, ", "), Fix: "mint the brand in a .ts file, where the cast is erased and listed in the checked report (rulings.md:29)"}
}

// recordBrandMint lists a lowered mint once per site and target, for the checked report.
func (l *lowering) recordBrandMint(node *ast.Node, target *checker.Type, mint *brandMint) {
	record := ir.BrandMint{Where: l.program.Where(node), Target: l.checker.TypeToString(target), Members: mint.members}
	for _, existing := range l.result.BrandMints {
		if existing.Where == record.Where && existing.Target == record.Target {
			return
		}
	}
	l.result.BrandMints = append(l.result.BrandMints, record)
}
