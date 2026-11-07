package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// phantomParts separates a primitive from the object views that give it a checker-only name.
// Void is included because tsc's __String deliberately brands undefined through void.
func (lowering *lowering) phantomRefusal(node *ast.Node) error {
	if node.Kind != ast.KindIntersectionType {
		return nil
	}
	var primitive *checker.Type
	var objects []*checker.Type
	for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
		proven := lowering.concrete(lowering.checker.GetTypeAtLocation(part))
		if base, brands := lowering.phantomParts(proven); base != nil {
			primitive = base
			objects = append(objects, brands...)
		} else if proven.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0 {
			primitive = proven
		} else if proven.Flags()&checker.TypeFlagsObject != 0 {
			objects = append(objects, proven)
		}
	}
	if primitive == nil || lowering.phantomOverloadOnly(node) {
		return nil
	}
	for _, object := range objects {
		for _, field := range lowering.checker.GetPropertiesOfType(object) {
			if checker.IsKnownSymbol(field) {
				return &Refused{
					Where: lowering.program.Where(node),
					What:  "a primitive brand member keyed by a symbol",
					Fix:   "key the brand member by a name the primitive does not have on its own properties or prototype chain",
				}
			}
			if primitiveMember(primitive.Flags(), field.Name) {
				return &Refused{
					Where: lowering.program.Where(node),
					What:  "a primitive brand member " + field.Name + " that exists on the primitive",
					Fix:   "use a member name the primitive does not have on its own properties or prototype chain",
				}
			}
			if !phantomField(lowering.checker.GetTypeOfSymbol(field)) {
				return &Refused{
					Where: lowering.program.Where(node),
					What:  "a primitive brand member " + field.Name + " whose type is not void",
					Fix:   "make " + field.Name + " undefined (or void) so the brand is phantom",
				}
			}
		}
	}
	return nil
}

func init() { refusalRules.register(400, "phantomRefusal", (*lowering).phantomRefusal) }

// phantomAssignable compares only primitives, preserving literal constraints. It never grants an
// object cast or a cast between different primitive kinds merely because both fit in one word.
func (lowering *lowering) phantomAssignable(from, to *checker.Type) bool {
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range from.Types() {
			if !lowering.phantomAssignable(part, to) {
				return false
			}
		}
		return true
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range to.Types() {
			if lowering.phantomAssignable(from, part) {
				return true
			}
		}
		return false
	}
	if base := lowering.phantomBase(from); base != nil {
		from = base
	}
	if base := lowering.phantomBase(to); base != nil {
		to = base
	}
	primitiveFlags := checker.TypeFlagsStringLike | checker.TypeFlagsNumberLike | checker.TypeFlagsBooleanLike | checker.TypeFlagsVoid | checker.TypeFlagsUndefined
	if from.Flags()&primitiveFlags == 0 || to.Flags()&primitiveFlags == 0 {
		return false
	}
	return lowering.checker.IsTypeAssignableTo(from, to)
}

func (lowering *lowering) hasPhantom(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if lowering.hasPhantom(part) {
				return true
			}
		}
	}
	return lowering.phantomBase(proven) != nil
}

func (lowering *lowering) phantomCast(source, target *checker.Type) bool {
	if !lowering.hasPhantom(source) && !lowering.hasPhantom(target) {
		return false
	}
	if lowering.phantomAssignable(source, target) {
		return true
	}
	// __String's void arm permits the same cast out as its string arm. Its null pointer stays
	// undefined; this erases the assertion rather than materializing or checking a brand object.
	if source.Flags()&checker.TypeFlagsUnion != 0 && target.Flags()&checker.TypeFlagsStringLike != 0 {
		present := false
		for _, part := range source.Types() {
			if lowering.phantomUndefined(part) {
				continue
			}
			if !lowering.phantomAssignable(part, target) {
				return false
			}
			present = true
		}
		return present
	}
	return false
}

func (lowering *lowering) phantomMemberType(proven *checker.Type, name string) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if !lowering.phantomMemberType(part, name) {
				return false
			}
		}
		return true
	}
	if lowering.phantomBase(proven) == nil {
		return false
	}
	_, objects := lowering.phantomParts(proven)
	for _, object := range objects {
		if lowering.checker.GetPropertyOfType(object, name) != nil {
			return true
		}
	}
	return false
}

func (lowering *lowering) phantomMember(node *ast.Node) (ir.Expression, bool, error) {
	var receiver *ast.Node
	var name string
	optional := false
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		receiver, name, optional = access.Expression, access.Name().Text(), access.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind != ast.KindStringLiteral {
			return nil, false, nil
		}
		receiver, name, optional = access.Expression, key.Text(), access.QuestionDotToken != nil
	default:
		return nil, false, nil
	}
	if !lowering.phantomMemberType(lowering.concrete(lowering.checker.GetTypeAtLocation(receiver)), name) {
		return nil, false, nil
	}
	value, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	// A stale narrowing is checked here with a catchable TypeError, not Defined's panic.
	if defined, ok := value.(ir.Defined); ok {
		value = defined.Value
	}
	if unwrapped, ok := value.(ir.Unwrap); ok {
		value = unwrapped.Value
	}
	failure := ir.MakeError{
		Message: ir.StringConstant{
			Index: lowering.constant("Cannot read properties of undefined (reading '" + name + "')"),
		},
		Name: ir.StringConstant{Index: lowering.constant("TypeError")},
	}
	return ir.PhantomMember{Value: value, Name: name, Optional: optional, Failure: failure}, true, nil
}

// The result is always undefined if the read completes, but the receiver must still run, and may
// throw. A conditional evaluates it once before writing the known spelling.
func (lowering *lowering) phantomSpelling(value ir.Expression) ir.Expression {
	text := ir.StringConstant{Index: lowering.constant("undefined")}
	return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: text, WhenNot: text}
}

// Main already admits brands mentioned only in erased overload signatures. They
// never reach a runtime slot, so the new primitive representation is not needed.
// Any reference outside a bodyless function declaration keeps the refusal.
func (lowering *lowering) phantomOverloadOnly(node *ast.Node) bool {
	alias := node.Parent
	if alias == nil || alias.Kind != ast.KindTypeAliasDeclaration {
		return false
	}
	symbol := lowering.symbol(alias.Name())
	if symbol == nil {
		return false
	}
	seen, safe := false, true
	var visit ast.Visitor
	visit = func(use *ast.Node) bool {
		if !safe {
			return false
		}
		if ast.IsIdentifier(use) && use != alias.Name() && lowering.symbol(use) == symbol {
			seen = true
			var declaration *ast.Node
			for parent := use.Parent; parent != nil; parent = parent.Parent {
				if parent.Kind == ast.KindFunctionDeclaration {
					declaration = parent
					break
				}
			}
			if declaration == nil || declaration.Body() != nil {
				safe = false
				return false
			}
		}
		return use.ForEachChild(visit)
	}
	for _, file := range lowering.program.CompilerProgram().GetSourceFiles() {
		file.Node.ForEachChild(visit)
	}
	return seen && safe
}
