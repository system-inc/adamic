package lower

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// In .a, a declared property's type can't lie through an index signature (#npqr3sx, Adamic's ruling
// on #bq3fvdt). A record that declares a property narrower than its index signature's values, as
// interface Options { outFile?: string; [name: string]: string | number | undefined } does, keeps
// outFile a string only while nothing writes a number under its name, and tsc lets three roads do it:
//   - A wider view: Options seen as { [name: string]: string | number | undefined } can be written 3
//     under outFile. widened holds the declared property invariant against the view's index signature.
//   - A dictionary seen as Options: its entry under outFile, whatever it holds, is read as a string.
//     widened holds the dictionary's values against each declared property they'd be read through.
//   - A write through a key that isn't one literal, options[name] = 3 or { [name]: 3 }, which tsc
//     checks against the index signature alone. Only the value written can land under a declared
//     property, so the value is judged by recordNamedWrite's own check: refused when it fails one of
//     the rules 0.1 forbids (declaredKeyRefusable), and not yet, as in .ts, when it fails only the
//     identity gate, since the registry may admit it later (l.slotType).
// .ts keeps the checked read instead (recordReadType), which stops loudly at the declared read.

// declaredKeysFix is what each of the three refusals offers.
const declaredKeysFix = "give the record no declared keys, or write the declared property by its name"

// recordDeclaredKeyRefusal refuses, in .a, a write that can store a wider value under a declared
// property through the index signature, and a view widened finds reaching one. It runs before
// recordStorageView (2000), whose not yet would otherwise name the same view a storage gap, and
// refuses only these views: every other view keeps the stop it had.
func (l *lowering) recordDeclaredKeyRefusal(node *ast.Node) error {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	if err := l.declaredKeyWrite(node); err != nil {
		return err
	}
	l.declaredKeys = true
	defer func() { l.declaredKeys = false }()
	if source, target := l.recordViewSite(node); source != nil {
		// recordStorageView's own site and its own relation: the whole value seen as the slot, a
		// fresh array's elements and all, so the view is named where that rule would name it.
		own := l.checker.GetTypeAtLocation(source)
		if found := l.widened(own, target, map[[2]*checker.Type]bool{}); found != nil && found.declaredKey != "" {
			return l.wideningRefusal(node, own, target, found)
		}
	}
	// Every other view site refuseWidening judges: a shorthand property, a spread, a branch's implied
	// target. Only its declared-key refusals are this rule's; the rest it makes again at 3100.
	var refused *Refused
	if err := l.stringWideningRefusal(node); errors.As(err, &refused) && refused.Fix == declaredKeysFix {
		return err
	}
	return nil
}

func init() {
	refusalRules.register(1975, "recordDeclaredKeyRefusal", (*lowering).recordDeclaredKeyRefusal)
}

// declaredKeyWrite refuses a write through a key that isn't one literal, into a record with a
// writable index signature, of a value a declared property it can land under can't hold, and stops
// one only the identity gate stops, not yet, as recordNamedWrite does in .ts (declaredKeyRefusable). The writes are an
// element access written to (by =, a compound assignment, ++, a destructuring target or a for-in or
// for-of head), and an object literal's computed name, whose entry is written as the literal is made,
// into its own type and the type it's written into. Where the value written isn't one expression's
// (a destructuring target, a loop's head), it's whatever the index signature holds.
func (l *lowering) declaredKeyWrite(node *ast.Node) error {
	var key *ast.Node
	var receivers []*checker.Type
	var value *checker.Type
	switch node.Kind {
	case ast.KindElementAccessExpression:
		if !ast.IsAssignmentTarget(node) {
			return nil
		}
		access := node.AsElementAccessExpression()
		key = access.ArgumentExpression
		receivers = []*checker.Type{l.checker.GetTypeAtLocation(access.Expression)}
		value = l.writtenValue(node)
	case ast.KindComputedPropertyName:
		member := node.Parent
		if member == nil || member.Parent == nil || member.Parent.Kind != ast.KindObjectLiteralExpression || ast.IsAssignmentTarget(member.Parent) {
			return nil
		}
		// { outFile: 'a', [name]: 3 } is itself a record declaring outFile, whatever it's written into.
		key = node.AsComputedPropertyName().Expression
		receivers = []*checker.Type{l.checker.GetTypeAtLocation(member.Parent)}
		if contextual := l.checker.GetContextualType(member.Parent, checker.ContextFlagsNone); contextual != nil {
			receivers = append(receivers, contextual)
		}
		if member.Kind == ast.KindPropertyAssignment {
			value = l.checker.GetTypeAtLocation(member.AsPropertyAssignment().Initializer)
		}
	default:
		return nil
	}
	names, single := l.writtenNames(key)
	if single {
		return nil
	}
	for _, receiver := range receivers {
		for _, member := range l.definedMembers(l.checker.GetNonNullableType(receiver)) {
			for _, info := range l.checker.GetIndexInfosOfType(member) {
				if info.KeyType().Flags() != checker.TypeFlagsString || info.IsReadonly() {
					continue
				}
				written := value
				if written == nil {
					written = info.ValueType()
				}
				if property, identity := l.namedWriteMisfit(member, written, names); property != nil && !identity && l.declaredKeyRefusable(property, written) {
					return &Refused{Where: l.program.Where(node), What: "a write into a record with declared properties through a key that is not one literal, of a value a declared property it can land under cannot take (" + l.checker.TypeToString(written) + " written into " + l.checker.TypeToString(member) + " by " + sourceExpression(key) + ": " + property.Name + " is declared " + l.checker.TypeToString(l.checker.GetTypeOfSymbol(property)) + ")", Fix: declaredKeysFix}
				}
				if err := l.namedWriteStop(node, member, written, names); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// writtenNames is the names a key can be: single when it's one literal, or a value of one literal's
// type, whose write tsc checks against the property it names, or which lands in the index signature
// under a name no declared property takes. Otherwise it's each name a union of literals can be, or
// the name strings joined by + fold to (constantStringKey), which tsc types as string and doesn't
// check, so ["" + "constructor"] is judged as constructor is; or nil, when it can be any name.
func (l *lowering) writtenNames(key *ast.Node) (map[string]bool, bool) {
	key = ast.SkipParentheses(key)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral {
		return nil, true
	}
	proven := l.checker.GetTypeAtLocation(key)
	if proven.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsUniqueESSymbol) != 0 {
		return nil, true
	}
	if folded, known := constantStringKey(key); known {
		return map[string]bool{folded: true}, false
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return nil, false
	}
	names := map[string]bool{}
	for _, member := range proven.Types() {
		switch {
		case member.Flags()&checker.TypeFlagsStringLiteral != 0:
			names[member.AsLiteralType().Value().(string)] = true
		case member.Flags()&checker.TypeFlagsNumberLiteral != 0:
			names[strconv.FormatFloat(reflect.ValueOf(member.AsLiteralType().Value()).Float(), 'f', -1, 64)] = true
		default:
			return nil, false
		}
	}
	return names, false
}

// declaredKeyRefusable reports whether a value recordNamedWrite's list doesn't fit to a declared
// property is one 0.1 forbids there: not assignable to it, its optional undefined included, or
// failing the list's enum, invariance or storage rule, or the property readonly. A value that fails
// only the optional undefined (undefined written under an optional property, which may hold it), or
// only the identity gate (a SubPoint under names: Point, or an Other tsc relates to it structurally),
// is not yet, as in .ts: the registry may admit it once it judges the write (l.slotType, #b6m00y7).
func (l *lowering) declaredKeyRefusable(property *ast.Symbol, written *checker.Type) bool {
	to := l.checker.GetNonMissingTypeOfSymbol(property)
	return l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(written, l.checker.GetTypeOfSymbol(property)) || !l.enumAssignable(written, to) || l.widened(written, to, map[[2]*checker.Type]bool{}) != nil || !l.sameRecordStorage(written, to, map[[2]*checker.Type]bool{})
}

// writtenValue is the type of what an assignment or an increment writes into its target, or nil when
// the target isn't one's: the right side of =, and what a compound assignment or ++ makes.
func (l *lowering) writtenValue(target *ast.Node) *checker.Type {
	for target.Parent != nil && target.Parent.Kind == ast.KindParenthesizedExpression {
		target = target.Parent
	}
	parent := target.Parent
	switch {
	case parent == nil:
		return nil
	case parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == target && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind):
		if parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			return l.checker.GetTypeAtLocation(parent.AsBinaryExpression().Right)
		}
		return l.checker.GetTypeAtLocation(parent)
	case parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression:
		return l.checker.GetTypeAtLocation(parent)
	}
	return nil
}

// declaredNarrower is the first property declares has, but those skip names, that can't hold a value
// of type value: written under its name through an index signature, value would break it.
func (l *lowering) declaredNarrower(declares, value *checker.Type, skip func(name string) bool) *ast.Symbol {
	for _, property := range l.checker.GetPropertiesOfType(declares) {
		if skip != nil && skip(property.Name) {
			continue
		}
		declared := l.checker.GetTypeOfSymbol(property)
		if !l.checker.IsTypeAssignableTo(value, declared) || !l.enumAssignable(value, declared) {
			return property
		}
	}
	return nil
}

// declaredKeyWidened is the declared property an index signature reaches when a value of type from is
// seen as type to, in either direction, or nil:
//   - from declares it, and to's writable index signature can write under its name what it can't
//     hold. A copy (fresh) only reads from's properties, and isn't judged this way.
//   - to declares it, from's index signature (readonly too, since this direction only reads) holds
//     what it can't hold, and from doesn't declare it, or skip names it as written after a spread.
func (l *lowering) declaredKeyWidened(from, to *checker.Type, skip map[string]bool, fresh bool) *widening {
	if value := l.recordElement(to); value != nil && !fresh && !l.libraryIndexSignature(to) {
		viewDeclares := func(name string) bool { return l.checker.GetPropertyOfType(to, name) != nil }
		if property := l.declaredNarrower(from, value, viewDeclares); property != nil {
			return &widening{source: l.checker.GetTypeOfSymbol(property), target: value, declaredKey: property.Name}
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(from) {
		if info.KeyType().Flags() != checker.TypeFlagsString {
			continue
		}
		sourceDeclares := func(name string) bool { return skip[name] || l.checker.GetPropertyOfType(from, name) != nil }
		if property := l.declaredNarrower(to, info.ValueType(), sourceDeclares); property != nil {
			return &widening{source: info.ValueType(), target: l.checker.GetTypeOfSymbol(property), declaredKey: property.Name, dictionary: true}
		}
	}
	return nil
}

// libraryIndexSignature reports whether a type's index signature is the library's own, a parameter
// of Object.values or Object.entries, which reads the record it's handed.
func (l *lowering) libraryIndexSignature(proven *checker.Type) bool {
	for _, info := range l.checker.GetIndexInfosOfType(l.concrete(proven)) {
		if declaration := info.Declaration(); declaration != nil && load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			return true
		}
	}
	return false
}

// declaredKeyRefusal is a declared-key widening's refusal: one reason for each direction.
func (l *lowering) declaredKeyRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {
	if found.dictionary {
		return &Refused{Where: l.program.Where(node), What: "a dictionary read through a declared property narrower than its values (" + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ": " + found.declaredKey + " is declared " + l.checker.TypeToString(found.target) + ", and the index signature holds " + l.checker.TypeToString(found.source) + ")", Fix: declaredKeysFix}
	}
	return &Refused{Where: l.program.Where(node), What: "a value with declared properties seen through a wider index signature, which can write under a declared property what it does not hold (" + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ": " + found.declaredKey + " is declared " + l.checker.TypeToString(found.source) + ", and the index signature holds " + l.checker.TypeToString(found.target) + ")", Fix: declaredKeysFix}
}
