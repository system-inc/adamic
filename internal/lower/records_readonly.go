package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A readonly string-index type (Readonly<OptionsBase>, Readonly<Record<string, V>>) has two uses, and each
// program gives each such type one of them, since a function whose parameter has it is compiled once and
// reads it in one layout (Runtime's must on #6c4swvt, piece 4):
//   - a view of an enumeration's object, held in that object's own fixed storage (enumerationReadonlyType,
//     tsc's getEnumMembers, #52zja36), which is what every such type was before;
//   - a view of a record, held as that record (tsc's serializeOptionBaseObject and
//     handleOptionConfigDirTemplateSubstitution over Readonly<OptionsBase>, #6c4swvt piece 4).
//
// The first value the program puts into the type, at a view site in module order (recordViewSite: an
// initializer, a cast, an argument, a return, a property or shorthand), decides which. A record there makes
// the type a record (readonlyRecordType). Anything else, or no site at all, keeps it what it was. A later site
// putting the other kind in stops at the storage view, whose words then name the type and the site that
// decided it (readonlyRecordReason). A source that is itself a readonly string-index type decides nothing,
// since its own kind is decided the same way, and the storage view compares the two when it meets them.

// readonlyKind is what a program's first view site decided a readonly string-index type holds.
type readonlyKind int

const (
	readonlyUndecided readonlyKind = iota
	readonlyRecord
	readonlyFixed
)

// readonlyDecision is a readonly string-index type's kind and the site that decided it.
type readonlyDecision struct {
	kind  readonlyKind
	site  *ast.Node
	value *checker.Type
}

// readonlyIndexType is whether t is an object whose only index signature is a readonly string one, with no
// call or construct signature, and isn't an array, a tuple or a class instance.
func (l *lowering) readonlyIndexType(t *checker.Type) bool {
	if t == nil || t.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(t) || checker.IsTupleType(t) || isClassInstance(t) {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(t)
	return len(infos) == 1 && infos[0].KeyType().Flags() == checker.TypeFlagsString && infos[0].IsReadonly() && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) == 0
}

// readonlyRecordType is whether t, a readonly string-index type, is held as a record: the program's first
// view site into it put a record there.
func (l *lowering) readonlyRecordType(t *checker.Type) bool {
	if t == nil || !l.readonlyIndexType(t) {
		return false
	}
	decisions := l.readonlyDecisions()
	if decisions == nil {
		return false
	}
	decision, decided := decisions[t]
	return decided && decision.kind == readonlyRecord
}

// readonlyDecisions walks the program's view sites once, in module order, and decides each readonly
// string-index type a value is put into. It's nil while the walk runs, so a type asked about during it keeps
// what it was before, and a source's own representation never waits on the type it goes into.
func (l *lowering) readonlyDecisions() map[*checker.Type]readonlyDecision {
	if l.readonlyKinds != nil || l.readonlyDeciding || l.program == nil || len(l.program.Files()) == 0 {
		return l.readonlyKinds
	}
	l.readonlyDeciding = true
	defer func() { l.readonlyDeciding = false }()
	decisions := map[*checker.Type]readonlyDecision{}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		l.readonlyKinds = decisions
		return decisions
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if source, target := l.recordViewSite(node); source != nil && target != nil {
			for _, member := range l.definedMembers(l.concrete(l.checker.GetNonNullableType(target))) {
				member = l.concrete(member)
				if _, decided := decisions[member]; decided || !l.readonlyIndexType(member) {
					continue
				}
				value := l.checker.GetTypeAtLocation(source)
				if kind := l.readonlySourceKind(value); kind != readonlyUndecided {
					decisions[member] = readonlyDecision{kind: kind, site: source, value: value}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	l.readonlyKinds = decisions
	return decisions
}

// readonlySourceKind is what a value of type t puts into a readonly string-index type: a record, a fixed
// object, or undecided for anything else (a readonly string-index type of its own, a union of both kinds, a
// value no storage is known for).
func (l *lowering) readonlySourceKind(t *checker.Type) readonlyKind {
	kind := readonlyUndecided
	for _, member := range l.definedMembers(l.concrete(l.checker.GetNonNullableType(t))) {
		member = l.concrete(member)
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			continue
		}
		var this readonlyKind
		switch {
		case l.readonlyIndexType(member):
			return readonlyUndecided
		case l.recordElement(member) != nil:
			this = readonlyRecord
		default:
			if of, known := l.representation(member); known && of == ir.Object {
				this = readonlyFixed
			} else {
				return readonlyUndecided
			}
		}
		if kind != readonlyUndecided && kind != this {
			return readonlyUndecided
		}
		kind = this
	}
	return kind
}

// readonlyRecordReason names, for the storage view's stop, why a readonly string-index type in target holds
// what it holds: the site that decided it, and what that site put there. It's "" when no member of target
// is a decided readonly string-index type.
func (l *lowering) readonlyRecordReason(target *checker.Type) string {
	decisions := l.readonlyDecisions()
	if target == nil || decisions == nil {
		return ""
	}
	for _, member := range l.definedMembers(l.concrete(l.checker.GetNonNullableType(target))) {
		decision, decided := decisions[l.concrete(member)]
		if !decided {
			continue
		}
		held := "a fixed object"
		if decision.kind == readonlyRecord {
			held = "a record"
		}
		return "; " + l.checker.TypeToString(member) + " is held as " + held + ", since " + l.program.Where(decision.site) + " puts a " + l.checker.TypeToString(decision.value) + " there first"
	}
	return ""
}
