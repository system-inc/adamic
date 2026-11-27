package lower

import (
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// narrowedArmPropertyReadGap is narrowedArmReadGap's rule for a property read, m.n, this.n, a.m.n or m['n'],
// as narrowedNullishPropertyReadGap is the nullish rule's (#8nj5pn4). The checker narrows a property by a
// discriminant check and keeps the narrowing across a call that writes the property again:
//
//	function reset(m: M): void { m.n = b; }
//	if (m.n.kind === 'a') { reset(m); return String(m.n.a + 1); }
//
// Node prints NaN, since m.n holds b by the read, and native read b's storage through a's layout and printed 1.
// A property has no declaration that keeps its writes in one function, so it's proved by name, as the nullish
// twin proves it: a getter of the name stops the read first, since a getter answers each read itself. Then
// the read lowers as it stands when the place it was last seen narrowed (an assignment, or a check that mentions
// it, armObservationOf) is followed by nothing that can run program code before the read, or by nothing that can
// write its name or install an accessor (propertyWriteMayRunBetween). A read with no such place is judged
// from the start of its own function.
// Any other such read is a narrowing kept across a call that may rewrite the property, a claim the walk can't
// prove whatever the field's representation (Adamic's ruling on #p1p7k3g): even where every type a stale value
// could be holds the field at one representation, the stale value's field is read through the arm it isn't. So
// every one is the floor's (narrowedArmView): refused in .a and, in .ts, read through a checked view of the
// declared type, so a stale arm stops loudly at the read and an honest one runs.
func (l *lowering) narrowedArmPropertyReadGap(node *ast.Node) error {
	if !l.isExpression(node) || (ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	// The read's name, and its property where it has one. An element's key is read as tsc narrows it: a literal,
	// or any key whose type is a string or number literal, a const index among them (Language's e02).
	var name string
	var property *ast.Symbol
	if node.Kind == ast.KindPropertyAccessExpression {
		name = node.Name().Text()
		property = l.checker.GetSymbolAtLocation(node.Name())
	} else {
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		switch key.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral:
			name = key.Text()
			property = l.checker.GetSymbolAtLocation(key)
		default:
			keyType := l.checker.GetTypeAtLocation(key)
			switch {
			case keyType.Flags()&checker.TypeFlagsNumberLiteral != 0:
				// Its value, never its spelling: an enum member's literal type prints as Slot.First (Language's e05).
				value := reflect.ValueOf(keyType.AsLiteralType().Value())
				if value.Kind() != reflect.Float64 {
					return nil
				}
				name = strconv.FormatFloat(value.Float(), 'f', -1, 64)
			case keyType.Flags()&checker.TypeFlagsStringLiteral != 0:
				text, isText := keyType.AsLiteralType().Value().(string)
				if !isText {
					return nil
				}
				name = text
			default:
				return nil // A key of no literal type narrows nothing.
			}
			property = l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression)), name)
		}
	}
	narrowed := l.checker.GetTypeAtLocation(node)
	var declared *checker.Type
	if property != nil {
		declared = l.checker.GetTypeOfSymbol(property)
	} else if _, numberError := strconv.ParseFloat(name, 64); node.Kind == ast.KindElementAccessExpression && numberError == nil {
		// An array's element by a literal index, arr[0], has no property of its own: its declared type is the
		// array's element type, and a write by any index or an array method may rewrite it (Language's r08).
		receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression))
		if !l.isArrayType(receiver) {
			return nil
		}
		declared = l.elementTypeOfArray(receiver)
	}
	if declared == nil {
		return nil // No declared property or element, so no declared type the checker narrowed from.
	}
	if narrowed.Id() == declared.Id() || narrowed.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	if _, nullish := l.nullishUnit(narrowed); nullish {
		return nil // narrowedNullishPropertyReadGap's.
	}
	present := l.checker.GetNonNullableType(declared)
	if present.Id() == narrowed.Id() || l.checker.GetNonNullableType(narrowed).Id() == present.Id() {
		// A presence strip, checked or observed at run time, as the binding rule says, or the declared type with
		// only undefined added, an element read under noUncheckedIndexedAccess, which narrows nothing (Language's e01).
		return nil
	}
	if storage, known := l.representation(declared); known && storage == ir.Union {
		return nil // The held union's member is checked at run time before it's narrowed.
	}
	if narrowedStorage, narrowedKnown := l.representation(narrowed); narrowedKnown && narrowedStorage != ir.Object {
		return nil // Only an object arm reads through an object's layout: a scalar, a string, an array, a closure or a record arm doesn't. An arm whose storage isn't known is judged.
	}
	if property != nil && property.Flags&ast.SymbolFlagsEnumMember != 0 {
		return nil // An enum's member is a constant no write reaches.
	}
	what := "a read of property " + name + " narrowed to " + l.checker.TypeToString(narrowed) + " where a call may have rewritten it"
	// A getter of the name answers each read itself, so it stops the read whatever the flow says, not yet in .a and
	// .ts alike. A program that could install one some other way (writes.accessors: a descriptor, a prototype) stops
	// it only where program code can run between the narrowing and the read, since nothing else can install it there.
	writes := l.propertyWritesOf()
	if l.accessorName(name) {
		return l.notYet(node, what)
	}
	_, notNumeric := strconv.ParseFloat(name, 64)
	numeric := notNumeric == nil
	observation := l.armObservationOf(node)
	if observation == nil {
		// No one place the narrowing was seen (a join of two checks, as a ternary chain's), so what may run is
		// judged from the start of the read's own code, where every narrowing of it begins. A read with no such
		// start stops if anything can write its name.
		boundary := callBoundary(node)
		if boundary == nil || boundary.Pos() < 0 {
			if numeric || writes.everyName || len(writes.named[name]) > 0 {
				return l.narrowedArmView(node, declared, narrowed, what)
			}
			return nil
		}
		observation = &nullishObservation{from: boundary.Pos()}
	}
	exempt := l.appendsBetween(observation, node, numeric)
	if writes.accessors && l.codeBetween(observation.from, node, exempt) {
		return l.narrowedArmView(node, declared, narrowed, what)
	}
	if l.propertyWriteMayRunBetween(node, name, numeric, observation.from, observation.target, exempt) {
		return l.narrowedArmView(node, declared, narrowed, what)
	}
	return nil
}

// narrowedArmView is the floor's answer to a read narrowedArmPropertyReadGap can't prove, what naming it. In .a it's
// refused, since .a reads no narrowing it can't prove (Adamic's ruling on #p1p7k3g), and the fix is a narrowing the
// walk can prove. In .ts it's the read through a checked view of declared for each member it was narrowed to
// (narrowedCheckedViews), as the swing reads an open-enum refinement it can't prove (Adamic's ruling on #8nj5pn4).
// The view checks the field at every read through the member's type, so a stale arm stops loudly there, exit 70,
// and an honest one reads its field. nil when the views register; where one can't be built, its own NotYet, which
// fails closed.
func (l *lowering) narrowedArmView(node *ast.Node, declared, narrowed *checker.Type, what string) error {
	file := l.program.FileName(ast.GetSourceFileOfNode(node))
	if strings.HasSuffix(file, ".a") {
		return &Refused{Where: l.program.Where(node), What: what + ": the checker keeps the narrowing across the call but can't prove it (adamic/narrowed-arm)", Fix: "test the property again after the call, or read it into a const before the call and narrow that"}
	}
	if narrowedArmViewSites != nil {
		narrowedArmViewSites.record(filepath.Dir(file), l.program.Where(node))
	}
	return l.narrowedCheckedViews(node, declared, narrowed)
}

// narrowedArmViewSites, when a test sets it, records each .ts read the floor takes through checked views in place of
// its stop (narrowedArmView), at the read's site, keyed by its file's directory, as enumTagSwingSites records the
// swing's. A view taken or skipped where the program prints the same either way shows only in the record, so a
// test pins that every read the floor judges leaves a site and every read it passes over leaves none, and a mutant
// that moves a read across that line goes red there. Production leaves it nil.
var narrowedArmViewSites *enumTagSwingRecord

// armObservationOf walks the read's flow straight back to the place it was last seen narrowed: a plain
// assignment of it, after which it holds what was assigned, or a condition that mentions it, the check that
// narrowed it. A call, which codeBetween judges, an assignment of something else and a join whose every way
// in reaches the same place are walked through. A loop, a label of any other kind, a write of what holds the
// read or the start of a function ends the walk with none. A condition's target is nil, since it writes
// nothing to exempt.
func (l *lowering) armObservationOf(read *ast.Node) *nullishObservation {
	data := read.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return nil
	}
	return l.armObservationAlong(read, data.FlowNode, 0, map[*ast.FlowNode]*nullishObservation{})
}

func (l *lowering) armObservationAlong(read *ast.Node, flow *ast.FlowNode, joins int, seen map[*ast.FlowNode]*nullishObservation) *nullishObservation {
	for flow != nil {
		switch {
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(l.checker, read, flow.Node) {
				parent := flow.Node.Parent
				if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Left == flow.Node {
					return &nullishObservation{target: flow.Node, from: parent.End(), assigned: true}
				}
				return nil
			}
			if enumTagContainsReference(l.checker, read, flow.Node) || l.mentionsReference(read, flow.Node) {
				return nil
			}
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if l.mentionsReference(read, flow.Node) {
				return &nullishObservation{from: flow.Node.End()}
			}
		case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
		case flow.Flags&ast.FlowFlagsBranchLabel != 0:
			if found, walked := seen[flow]; walked {
				return found
			}
			seen[flow] = nil
			if joins >= nullishObservationDepth {
				return nil
			}
			var common *nullishObservation
			for list := flow.Antecedents; list != nil; list = list.Next {
				if list.Flow.Flags&ast.FlowFlagsUnreachable != 0 {
					continue
				}
				found := l.armObservationAlong(read, list.Flow, joins+1, seen)
				if found == nil || common != nil && found.from != common.from {
					return nil
				}
				common = found
			}
			seen[flow] = common
			return common
		default:
			return nil
		}
		flow = flow.Antecedent
	}
	return nil
}
