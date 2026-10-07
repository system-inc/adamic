package lower

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// phantomParts separates a primitive from the object views that give it a checker-only name.
// Void is included because tsc's __String deliberately brands undefined through void.
func (l *lowering) phantomParts(proven *checker.Type) (*checker.Type, []*checker.Type) {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return nil, nil
	}
	var primitive *checker.Type
	var objects []*checker.Type
	for _, part := range proven.Types() {
		switch {
		case part.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0:
			if primitive != nil {
				return nil, nil
			}
			primitive = part
		case part.Flags()&checker.TypeFlagsObject != 0:
			objects = append(objects, part)
		default:
			return nil, nil
		}
	}
	if primitive == nil || len(objects) == 0 {
		return nil, nil
	}
	return primitive, objects
}

func (l *lowering) phantomBase(proven *checker.Type) *checker.Type {
	primitive, objects := l.phantomParts(proven)
	if primitive == nil {
		return nil
	}
	for _, object := range objects {
		if len(l.checker.GetSignaturesOfType(object, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(object, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(object)) != 0 {
			return nil
		}
		for _, field := range l.checker.GetPropertiesOfType(object) {
			if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) || primitiveMember(primitive.Flags(), field.Name) {
				return nil
			}
		}
	}
	return primitive
}

func phantomField(proven *checker.Type, optional bool) bool {
	flags := proven.Flags()
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if !phantomField(part, optional) {
				return false
			}
		}
		return true
	}
	return flags&checker.TypeFlagsVoid != 0 || optional && flags&checker.TypeFlagsUndefined != 0
}

// These inventories are the own names of boxed primitives and their complete prototype chains,
// observed on Node and independently checked in TestPhantomPrimitiveNames. The checker library is
// intentionally smaller than JavaScript's, so its declarations alone cannot prove absence.
var phantomObjectNames = strings.Fields("__defineGetter__ __defineSetter__ __lookupGetter__ __lookupSetter__ __proto__ constructor hasOwnProperty isPrototypeOf propertyIsEnumerable toLocaleString toString valueOf")
var phantomStringNames = strings.Fields("anchor at big blink bold charAt charCodeAt codePointAt concat endsWith fixed fontcolor fontsize includes indexOf isWellFormed italics lastIndexOf length link localeCompare match matchAll normalize padEnd padStart repeat replace replaceAll search slice small split startsWith strike sub substr substring sup toLocaleLowerCase toLocaleUpperCase toLowerCase toUpperCase toWellFormed trim trimEnd trimLeft trimRight trimStart")
var phantomNumberNames = strings.Fields("toExponential toFixed toPrecision")

func primitiveMember(flags checker.TypeFlags, name string) bool {
	if flags&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0 {
		return false
	}
	if slices.Contains(phantomObjectNames, name) {
		return true
	}
	if flags&checker.TypeFlagsStringLike != 0 {
		if slices.Contains(phantomStringNames, name) {
			return true
		}
		// A string may have any nonnegative integer index below its length. Noncanonical spellings
		// such as 00 and -0 aren't string exotic own properties.
		index, err := strconv.ParseUint(name, 10, 64)
		return err == nil && index < 1<<53 && strconv.FormatUint(index, 10) == name
	}
	return flags&checker.TypeFlagsNumberLike != 0 && slices.Contains(phantomNumberNames, name)
}

func (l *lowering) phantomUndefined(proven *checker.Type) bool {
	if base := l.phantomBase(proven); base != nil {
		return base.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0
	}
	return false
}

// Refuse even an unused brand declaration. Inspect the written constituents too: tsc may reduce a
// conflicting length or discriminant to never, which must not hide the member that needs fixing.
func (l *lowering) phantomRefusal(node *ast.Node) error {
	if node.Kind != ast.KindIntersectionType {
		return nil
	}
	var primitive *checker.Type
	var objects []*checker.Type
	for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
		proven := l.concrete(l.checker.GetTypeAtLocation(part))
		if base, brands := l.phantomParts(proven); base != nil {
			primitive = base
			objects = append(objects, brands...)
		} else if proven.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0 {
			primitive = proven
		} else if proven.Flags()&checker.TypeFlagsObject != 0 {
			objects = append(objects, proven)
		}
	}
	if primitive == nil {
		return nil
	}
	for _, object := range objects {
		for _, field := range l.checker.GetPropertiesOfType(object) {
			if primitiveMember(primitive.Flags(), field.Name) {
				return &Refused{Where: l.program.Where(node), What: "a primitive brand member " + field.Name + " that exists on the primitive", Fix: "use a member name the primitive does not have on its own properties or prototype chain"}
			}
			if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) {
				return &Refused{Where: l.program.Where(node), What: "a primitive brand member " + field.Name + " whose type is not void", Fix: "make " + field.Name + " void (or optional and typed undefined) so the brand is phantom"}
			}
		}
	}
	return nil
}

// phantomAssignable compares only primitives, preserving literal constraints. It never grants an
// object cast or a cast between different primitive kinds merely because both fit in one word.
func (l *lowering) phantomAssignable(from, to *checker.Type) bool {
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range from.Types() {
			if !l.phantomAssignable(part, to) {
				return false
			}
		}
		return true
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range to.Types() {
			if l.phantomAssignable(from, part) {
				return true
			}
		}
		return false
	}
	if base := l.phantomBase(from); base != nil {
		from = base
	}
	if base := l.phantomBase(to); base != nil {
		to = base
	}
	primitiveFlags := checker.TypeFlagsStringLike | checker.TypeFlagsNumberLike | checker.TypeFlagsBooleanLike | checker.TypeFlagsVoid | checker.TypeFlagsUndefined
	if from.Flags()&primitiveFlags == 0 || to.Flags()&primitiveFlags == 0 {
		return false
	}
	return l.checker.IsTypeAssignableTo(from, to)
}

func (l *lowering) hasPhantom(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if l.hasPhantom(part) {
				return true
			}
		}
	}
	return l.phantomBase(proven) != nil
}

func (l *lowering) phantomCast(source, target *checker.Type) bool {
	if !l.hasPhantom(source) && !l.hasPhantom(target) {
		return false
	}
	if l.phantomAssignable(source, target) {
		return true
	}
	// __String's void arm permits the same cast out as its string arm. Its null pointer stays
	// undefined; this erases the assertion rather than materializing or checking a brand object.
	if source.Flags()&checker.TypeFlagsUnion != 0 && target.Flags()&checker.TypeFlagsStringLike != 0 {
		present := false
		for _, part := range source.Types() {
			if l.phantomUndefined(part) {
				continue
			}
			if !l.phantomAssignable(part, target) {
				return false
			}
			present = true
		}
		return present
	}
	return false
}

func (l *lowering) phantomMemberType(proven *checker.Type, name string) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if !l.phantomMemberType(part, name) {
				return false
			}
		}
		return true
	}
	if l.phantomBase(proven) == nil {
		return false
	}
	_, objects := l.phantomParts(proven)
	for _, object := range objects {
		if l.checker.GetPropertyOfType(object, name) != nil {
			return true
		}
	}
	return false
}

func (l *lowering) phantomMember(node *ast.Node) (ir.Expression, bool, error) {
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
	if !l.phantomMemberType(l.concrete(l.checker.GetTypeAtLocation(receiver)), name) {
		return nil, false, nil
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	// A stale narrowing is checked here with a catchable TypeError, not Defined's panic.
	if defined, ok := value.(ir.Defined); ok {
		value = defined.Value
	}
	if unwrapped, ok := value.(ir.Unwrap); ok {
		value = unwrapped.Value
	}
	failure := ir.MakeError{Message: ir.StringConstant{Index: l.constant("Cannot read properties of undefined (reading '" + name + "')")}, Name: ir.StringConstant{Index: l.constant("TypeError")}}
	return ir.PhantomMember{Value: value, Name: name, Optional: optional, Failure: failure}, true, nil
}

// The result is always undefined if the read completes, but the receiver must still run, and may
// throw. A conditional evaluates it once before writing the known spelling.
func (l *lowering) phantomSpelling(value ir.Expression) ir.Expression {
	text := ir.StringConstant{Index: l.constant("undefined")}
	return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: text, WhenNot: text}
}
