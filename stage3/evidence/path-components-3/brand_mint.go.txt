package lower

import (
	"slices"
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
// that extends exactly one base and declares nothing of its own but members phantomBrandMember calls
// phantom, and the source must relate to that base, instantiated as the target is, as an upcast
// would, by castAssignableAt, which judges a fresh literal as freshOrWidened does. So `readonly T[]`
// as a mutable SortedArray<T> is refused (the base is T[], which the source doesn't relate to), and
// so is an Animal[] as SortedArray<Dog>. A fresh `[]` relates to any element type, and a mutable T[] relates
// to a mutable base, because the brand carries no layout. Array brands are phantom only while every
// observation of a member's presence is refused (rulings.md:31), which phantom_array_presence.go does,
// so a minted value is held as its base (unbranded.go).
type brandMint struct {
	base    *checker.Type
	members []string
}

// brandMintOf is the mint a cast of operand makes, or nil when it isn't one: a target phantomExtension
// calls a phantom extension of its base, and a source that relates to that base.
func (l *lowering) brandMintOf(operand *ast.Node, source, target *checker.Type) *brandMint {
	// The innermost value is what's judged, so one typed any or unknown has nothing to judge.
	if source.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return nil
	}
	base, members := l.phantomExtension(target)
	if base == nil {
		return nil
	}
	if !l.checker.IsTypeAssignableTo(source, base) || !l.castAssignableAt(operand, source, base) {
		return nil
	}
	return &brandMint{base: base, members: members}
}

// phantomExtension is the one judgement of a phantom extension (#gb0fqaa): an interface that extends
// exactly one base and adds nothing to it but members phantomBrandMember calls phantom. It's target's
// base, instantiated as target is, and the members it adds, or nil when target isn't one. A mint
// (brandMintOf) adds a source relation to it, and unbranded reads a value of target as its base.
func (l *lowering) phantomExtension(target *checker.Type) (*checker.Type, []string) {
	if target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) {
		return nil, nil
	}
	declared := target
	var arguments []*checker.Type
	if target.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.Target() != nil {
		declared = target.Target()
		arguments = l.checker.GetTypeArguments(target)
	}
	if declared.ObjectFlags()&checker.ObjectFlagsInterface == 0 || declared.AsInterfaceType() == nil {
		return nil, nil
	}
	bases := checker.Checker_getBaseTypes(l.checker, declared)
	if len(bases) != 1 {
		// Several bases, or none, would need a relation per base; a brand tsc writes extends one.
		return nil, nil
	}
	base := bases[0]
	if parameters := declared.AsInterfaceType().TypeParameters(); len(parameters) > 0 {
		if len(arguments) < len(parameters) {
			return nil, nil
		}
		base = instantiateType(l.checker, base, newTypeMapper(parameters, arguments[:len(parameters)]))
	}
	// Nothing of its own but phantom members: the same signatures and index entries as its base,
	// and every property the base lacks is one phantomBrandMember admits.
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if len(l.checker.GetSignaturesOfType(target, kind)) != len(l.checker.GetSignaturesOfType(base, kind)) {
			return nil, nil
		}
	}
	if len(l.checker.GetIndexInfosOfType(target)) != len(l.checker.GetIndexInfosOfType(base)) {
		return nil, nil
	}
	// Judged from what the interface itself declares, never from its members' types: a member that
	// redeclares one of the base's (`v: number` over `v: number | string`) would otherwise pass as the
	// base's own and mint a narrowing no check holds (Language's must on #b5w3ycg, Nov 5). So every
	// member the interface declares is a property signature phantomBrandMember admits that the base
	// doesn't have, and nothing else: no method, signature, index entry or redeclaration.
	symbol := declared.Symbol()
	if symbol == nil {
		return nil, nil
	}
	var members []string
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindInterfaceDeclaration {
			return nil, nil
		}
		for _, member := range declaration.AsInterfaceDeclaration().Members.Nodes {
			if member.Kind != ast.KindPropertySignature || member.Name() == nil {
				return nil, nil
			}
			property := l.checker.GetSymbolAtLocation(member.Name())
			if property == nil || !l.phantomBrandMember(property) || l.checker.GetPropertyOfType(base, property.Name) != nil {
				return nil, nil
			}
			members = append(members, property.Name)
		}
	}
	if len(members) == 0 {
		return nil, nil
	}
	// Every property the target has is the base's or one of those, so nothing reaches it another way.
	for _, member := range l.checker.GetPropertiesOfType(target) {
		if l.checker.GetPropertyOfType(base, member.Name) == nil && !slices.Contains(members, member.Name) {
			return nil, nil
		}
	}
	return base, members
}

// mintOperand is what a cast mints from: its operand, or, when the operand is itself a cast to unknown
// (`x as unknown as SortedArray<T>`, the spelling tsc accepts for a required undefined brand member),
// that cast's operand. The hop to unknown is part of the mint and never lowered (brandMintValue lowers
// the operand), so no unknown value is ever made; the hop on its own is an upcast castProof admits as
// on main. Any other hop (`as any`) is judged as its own cast.
func (l *lowering) mintOperand(as *ast.AsExpression) (*ast.Node, *checker.Type) {
	operand := ast.SkipParentheses(as.Expression)
	if operand.Kind == ast.KindAsExpression && l.checker.GetTypeAtLocation(operand).Flags()&checker.TypeFlagsUnknown != 0 {
		inner := operand.AsAsExpression().Expression
		return inner, l.concrete(l.checker.GetTypeAtLocation(inner))
	}
	return as.Expression, l.concrete(l.checker.GetTypeAtLocation(as.Expression))
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
