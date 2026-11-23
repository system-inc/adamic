package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// dictionaryView is whether a cast castProof would stop as castIndexSignature is the checked dictionary view
// instead (#2v3xtgm, re-ruled Nov 23): tsc's 22 `object as MapLike<unknown>`, a value typed object read by its own
// keys. Natively a record's values are stored unboxed by the element its static type names (records.go's
// recordSlot), and only a read out of a slot held as a Union is checked (recordReadDeclared), so the view is admitted
// only where every read through it is a Union and the record holds Union-readable values:
//   - the target's non-nullish part is one dictionary type, held as ir.Record, whose one index signature is a string's
//     of exactly unknown, with no named property, call or construct signature beside it: reads through it are Unions,
//     and every use of one narrows it by its run-time tag. Any other value type keeps castIndexSignature's NotYet, since
//     its reads are unboxed by a storage the value may not have;
//   - the source's non-nullish part is object, held as ir.Union. A source with fields of its own keeps #mft3zc6's
//     storage rule's stop, unknown keeps its refusal, and a dictionary already held as a record is an upcast here;
//   - undefined beside the object passes through only where the target keeps it, and null keeps the NotYet, since a
//     record's reference has no null of its own;
//   - .a keeps main's refusal (castStop);
//   - no write anywhere in the program reaches a dictionary the target is assignable to (dictionaryViewWritten): the
//     record may be another alias's Record<string, string>, and a number written through the view would be read there
//     as a string.
//
// The test at the cast (dictionaryViewValue) is that the value is a record whose values are references, which a
// Union reads as they are (union.c's keyed_record_value and its note at 433).
func (l *lowering) dictionaryView(node *ast.Node, source, target *checker.Type) bool {
	if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return false
	}
	if l.includesNull(source) || l.includesNull(target) || l.includesUndefined(source) && !l.includesUndefined(target) {
		return false
	}
	presentSource, present := l.checker.GetNonNullableType(source), l.checker.GetNonNullableType(target)
	if presentSource.Flags()&checker.TypeFlagsNonPrimitive == 0 {
		return false
	}
	for _, member := range castMembers(source) {
		if member.Flags() != checker.TypeFlagsUndefined && member != presentSource {
			return false
		}
	}
	for _, member := range castMembers(target) {
		if member.Flags() != checker.TypeFlagsUndefined && member != present {
			return false
		}
	}
	if present.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(present) || len(l.checker.GetPropertiesOfType(present)) != 0 {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(present)
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString || infos[0].ValueType().Flags() != checker.TypeFlagsUnknown {
		return false
	}
	if len(l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(present, checker.SignatureKindConstruct)) != 0 || l.recordElement(present) == nil {
		return false
	}
	if from, known := l.representation(source); !known || from != ir.Union {
		return false
	}
	if to, known := l.representation(target); !known || to != ir.Record {
		return false
	}
	// The mutant (M7): dictionaryViewWritten answers false, and TestDictionaryViewCasts/written_through_a_dictionary
	// and /deleted_through_a_dictionary go red, admitted where a write through MapLike<unknown> could put a number in
	// a Record<string, string>.
	return !l.dictionaryViewWritten(present)
}

// dictionaryViewWritten is whether some write in the program may reach a dictionary present is assignable to. It's
// keyed by type, since nothing tracks a cast's result through the aliases it's handed to: every property or element
// write, a delete, a compound assignment, an increment and a destructuring's or a loop's target (propertyWritesOf's
// targets), and every call of a library member that writes any name (Object.assign, Reflect.set and the rest), judged
// by its first argument. It fails closed: Object or Reflect handed on whole, a library member read and not called,
// a receiver typed any, and a dictionary whose type mentions a type parameter each answer yes, since an instantiation
// may make it one present is assignable to.
func (l *lowering) dictionaryViewWritten(present *checker.Type) bool {
	writes := l.propertyWritesOf()
	if writes.libraryWhole {
		return true
	}
	for _, target := range writes.targets {
		var receiver *ast.Node
		if target.Kind == ast.KindPropertyAccessExpression {
			receiver = target.AsPropertyAccessExpression().Expression
		} else {
			receiver = target.AsElementAccessExpression().Expression
		}
		if l.dictionaryViewReceiver(receiver, present) {
			return true
		}
	}
	for _, member := range writes.libraryWrites {
		call := member.Parent
		if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != member || len(call.AsCallExpression().Arguments.Nodes) == 0 {
			return true
		}
		if l.dictionaryViewReceiver(call.AsCallExpression().Arguments.Nodes[0], present) {
			return true
		}
	}
	return false
}

// dictionaryViewReceiver is whether a write's receiver may be a dictionary present is assignable to: some member,
// nullish aside and a type parameter seen as its constraint, that holds an index signature and that present is
// assignable to, or that holds a string's index signature and mentions a type parameter, or that is any.
func (l *lowering) dictionaryViewReceiver(receiver *ast.Node, present *checker.Type) bool {
	held := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
	for _, member := range castMembers(held) {
		if member.Flags()&checker.TypeFlagsTypeParameter != 0 {
			constraint := l.checker.GetBaseConstraintOfType(member)
			if constraint == nil || constraint == member {
				continue // An unconstrained type parameter has no index signature to write through.
			}
			member = constraint
		}
		if member.Flags()&checker.TypeFlagsAny != 0 {
			return true
		}
		infos := l.checker.GetIndexInfosOfType(member)
		if len(infos) == 0 {
			continue
		}
		if l.checker.IsTypeAssignableTo(present, member) {
			return true
		}
		// Only a string's index signature, never an array's number index: a T[] element write, which every generic
		// list helper makes, reaches no dictionary.
		for _, info := range infos {
			if info.KeyType().Flags()&checker.TypeFlagsString != 0 && l.mentionsTypeParameter(member, map[*checker.Type]bool{}) {
				return true
			}
		}
	}
	return false
}

// dictionaryViewValue lowers the cast dictionaryView proved, in a helper that takes the value once, as
// checkedClassCast's does: undefined passes through where the target keeps it, and anything else is a record whose
// values are references, as one checked narrow tests it on both backends alike, or a run-time NotYet in the words
// that say which test it failed (ir's DictionaryViewRecordStop and DictionaryViewReferencesStop). A fixed object, a
// class instance, an array, a function and a host value fail the first; a record of numbers or booleans the second.
func (l *lowering) dictionaryViewValue(node *ast.Node) (ir.Expression, error) {
	as := node.AsAsExpression()
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	value, err := l.expression(as.Expression)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Union {
		return nil, l.notYet(node, "a checked dictionary view of "+l.checker.TypeToString(l.concrete(l.checker.GetTypeAtLocation(node)))+" over a value of "+l.checker.TypeToString(source)+" held as other than a union")
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	if l.includesUndefined(source) {
		b.body = append(b.body, ir.If{Condition: ir.IsUndefined{Value: held}, Then: []ir.Statement{ir.Return{Value: ir.Undefined{Of: ir.Record}}}})
	}
	// The mutants (cast_stop.mutants.md): drop the two tests, an unchecked narrow, and oracle's
	// TestDictionaryViewStopsAtTheCast goes red on its array, literal and class rows; drop ReferencesMessage alone,
	// and on its number and boolean rows.
	return b.finish("dictionary_view", ir.Narrow{Value: held, To: ir.Record, Checked: true, Message: ir.DictionaryViewRecordStop, ReferencesMessage: ir.DictionaryViewReferencesStop}), nil
}
