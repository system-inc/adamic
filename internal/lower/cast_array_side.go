package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// arrayArm is a cast from a union of arrays and other members to one side of it, which Array.isArray at the cast
// decides (#9z2ksv7, under Adamic's casts principle on #b5w3ycg: in .ts a cast's stop is NotYet when a run-time
// check could decide it). tsc's first stops are its shapes: tsbuildPublic.ts's AnyBuildOrder cast to
// CircularBuildOrder, and watchUtilities.ts's readonly string[] | Program | BuilderProgram | undefined cast to
// readonly string[]. In .a the cast is refused, as every checked downcast is (keptCheckDowncastRefusal).
type arrayArm struct {
	// array is a cast to the source's arrays; otherwise it's to the source's other members.
	array bool
	// dropsUndefined is a source holding undefined whose target doesn't: an undefined value fails the cast.
	dropsUndefined bool
	// keepsUndefined is a target holding undefined: an undefined value goes through.
	keepsUndefined bool
	// to is how the target is held, the picked side's Array or Object.
	to ir.Type
}

// arrayArmCast is the arm a cast from source to target picks, or nil, where castProof goes on to the stop it
// reached before. Every condition fails closed:
//   - the stop the cast would reach is castArrayArm, unionCastCause over the members unionMemberProof reads;
//   - Array.isArray's domain admits the source without its undefined, so every array member is a real array,
//     every other member rules an array out (arrayExcludedArm), and no member is a tuple;
//   - the target, undefined aside, is exactly one side, castAssignable both ways to the union of the source's
//     arrays or to the union of its other members, so nothing narrower on the picked side passes on a test that
//     only tells the sides apart (an element downcast, one member of several);
//   - undefined, on either side, is a plain undefined, never null, void or a phantom one;
//   - the source is held as a Union, and the target as the picked side is, Array or Object.
func (l *lowering) arrayArmCast(source, target *checker.Type) *arrayArm {
	if source.Flags()&checker.TypeFlagsUnion == 0 || l.includesNull(source) || l.includesNull(target) {
		return nil
	}
	for _, member := range castMembers(target) {
		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 && member.Flags() != checker.TypeFlagsUndefined || l.phantomUndefined(member) {
			return nil
		}
	}
	undefined := false
	arrays, others := []*checker.Type{}, []*checker.Type{}
	for _, member := range source.Types() {
		switch {
		case member.Flags() == checker.TypeFlagsUndefined:
			undefined = true
		case member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 || l.phantomUndefined(member):
			return nil
		case l.castArrayLike(member):
			arrays = append(arrays, member)
		default:
			others = append(others, member)
		}
	}
	if len(arrays) == 0 || len(others) == 0 {
		return nil
	}
	members := castMembers(source)
	if present := l.absentArm(source, target); present != nil {
		members = castMembers(present)
	}
	if l.unionCastCause(members, target) != castArrayArm || !l.arrayPredicateDomain(l.checker.GetNonNullableType(source)) {
		return nil
	}
	picked := l.checker.GetNonNullableType(target)
	arm := &arrayArm{keepsUndefined: l.includesUndefined(target)}
	arm.dropsUndefined = undefined && !arm.keepsUndefined
	switch arraySide, otherSide := l.checker.GetUnionType(arrays), l.checker.GetUnionType(others); {
	case l.castAssignable(picked, arraySide) && l.castAssignable(arraySide, picked):
		arm.array = true
	case l.castAssignable(picked, otherSide) && l.castAssignable(otherSide, picked):
	default:
		return nil
	}
	side := ir.Object
	if arm.array {
		side = ir.Array
	}
	held, heldKnown := l.representation(source)
	to, toKnown := l.representation(target)
	if !heldKnown || held != ir.Union || !toKnown || to != side {
		return nil
	}
	arm.to = to
	return arm
}

// arrayArmValue lowers a cast arrayArmCast proved: one helper, shared by both backends, that tests the held
// Union with Array.isArray, panics with the cast's own message where the value isn't the side the cast claims,
// and narrows it to that side. An undefined is no array, so the array side's test fails it already; the other
// side's tests it apart. It's the checked narrowed-union read's idiom (localRead). The arms the brand decides are
// recorded with the cast (admitArrayPredicate) where the test is built, as every such test's are: a checked view
// elsewhere, with no test at its own cast, could put a tuple behind an object arm unread, which natively is no
// array where Node says it is one, and checkArrayArmViews stops the cast in its own words once lowering is done
// (Lower's must on #9z2ksv7).
func (l *lowering) arrayArmValue(node *ast.Node, arm *arrayArm, source, target *checker.Type, message string) (ir.Expression, error) {
	if !l.admitArrayPredicate(node, l.checker.GetNonNullableType(source)) {
		// arrayArmCast asked the same domain, so this is a source it read otherwise: the stop it had before.
		return nil, l.castStop(node, castArrayArm, source, target, "")
	}
	value, err := l.expression(node.AsAsExpression().Expression)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Union {
		// castProof read the source's type as a Union; a value lowered otherwise gets the stop it had before.
		return nil, l.castStop(node, castArrayArm, source, target, "")
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.ArrayIsArray{Value: held})
	if !arm.array {
		matches = ir.Unary{Operator: ir.Not, Operand: matches}
	}
	if !arm.array && arm.dropsUndefined {
		matches = ir.Binary{Operator: ir.And, Left: matches, Right: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: held}}}
	}
	if arm.keepsUndefined {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	return b.finish("cast_array_arm", ir.Narrow{Value: held, To: arm.to}), nil
}

// arrayArmNarrowing is the statement .a writes for a cast arrayArmCast proves, which .a refuses: the Array.isArray
// test that panics where the value isn't the side the cast claims, written with the operand's own name where a
// narrowing can name it (castNarrowingOperand).
func (l *lowering) arrayArmNarrowing(node *ast.Node, arm *arrayArm) string {
	held, reference := castNarrowingOperand(node)
	test := held + " === undefined || Array.isArray(" + held + ")"
	switch {
	case arm.array && arm.keepsUndefined:
		test = held + " !== undefined && !Array.isArray(" + held + ")"
	case arm.array:
		test = "!Array.isArray(" + held + ")"
	case !arm.dropsUndefined:
		test = "Array.isArray(" + held + ")"
	}
	narrowing := "if (" + test + ") panic('why')"
	if !reference {
		return "hold it as const value, then " + narrowing
	}
	return narrowing
}

// arrayArmOperandType is the type the container fit walk (containerWhereOtherGoes) judges the value at node by,
// where node is the slot's value: node's own type, except at the operand of a cast castProof takes as an array
// arm to the object side, where it's that type without its array members (#9z2ksv7). The operand's slot is the cast's target,
// held as an object, so its arrays read to the walk as an array going where a plain object type is held
// (shapeFitArrayStop); but the cast's Array.isArray test panics on an array before the value reaches the slot,
// and Narrow hands the slot only the other members, which are still judged. Any other operand, a cast to the array
// side, and a cast arrayArmCast doesn't prove keep their own type, so their stops are as they were.
func (l *lowering) arrayArmOperandType(node *ast.Node) *checker.Type {
	own := l.checker.GetTypeAtLocation(node)
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindAsExpression || ast.SkipParentheses(parent.AsAsExpression().Expression) != node || own.Flags()&checker.TypeFlagsUnion == 0 {
		return own
	}
	source, target := l.concrete(l.checker.GetTypeAtLocation(parent.AsAsExpression().Expression)), l.concrete(l.checker.GetTypeAtLocation(parent))
	if arm := l.arrayArmCast(source, target); arm == nil || arm.array {
		return own
	}
	// The exemption reads the proof the cast takes, never a second judgment of its own: arrayArmCast above is a
	// cheap filter, and castProof, which castValue asks too, confirms the cast takes the arm. An earlier proof branch
	// added later then takes the cast, and its operand is judged whole (Lower's should on #9z2ksv7).
	if proof, err := l.castProof(parent); err != nil || proof.arrayArm == nil || proof.arrayArm.array {
		return own
	}
	kept := []*checker.Type{}
	for _, member := range own.Types() {
		if !l.castArrayLike(member) {
			kept = append(kept, member)
		}
	}
	return l.checker.GetUnionType(kept)
}
