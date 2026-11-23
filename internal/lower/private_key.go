package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// privateKeyHazard says why an own-key or presence test (hasOwnProperty, propertyIsEnumerable, a detached
// hasOwnProperty, in) can't lower yet when its key may spell a private slot's native name, or "" when it can.
// A private name is held as an ordinary slot under privateName's spelling, "#name@n", on native and the
// JavaScript backend alike, so a string key spelling it finds it, where Node finds no private name by any
// string key (#a0xw8mg: hasOwnProperty('#secret@1') answered true on both backends). A string literal can
// spell one only with a leading # and an @, so '#ordinary' and '#hidden' still lower (#a1b746w's corpus). A
// key the lowering can't read can spell anything, so it is held back on a receiver whose storage can hold a
// private slot. Object.hasOwn already refuses any # key (library_object.go).
func (l *lowering) privateKeyHazard(receiver, key *ast.Node) string {
	key = ast.SkipParentheses(key)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral {
		if strings.HasPrefix(key.Text(), "#") && strings.Contains(key.Text(), "@") {
			return "the key " + key.Text() + " spells a private slot's native name"
		}
		return ""
	}
	if l.mayHoldPrivateSlot(receiver) {
		return "a key that isn't a string literal may spell a private slot's native name"
	}
	return ""
}

// mayHoldPrivateSlot reports whether a value seen as the receiver's type can hold a private field's slot: the
// type itself, a class the program constructs that the type views, or a class whose constructor the type views
// (its static privates). When the program's shapes can't be read, it answers true.
func (l *lowering) mayHoldPrivateSlot(receiver *ast.Node) bool {
	view := l.concrete(l.checker.GetTypeAtLocation(receiver))
	if l.privateSlots(view, 0) {
		return true
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		var shape *checker.Type
		switch node.Kind {
		case ast.KindNewExpression, ast.KindClassExpression:
			shape = l.checker.GetTypeAtLocation(node)
		case ast.KindClassDeclaration:
			if node.Name() != nil {
				if symbol := l.checker.GetSymbolAtLocation(node.Name()); symbol != nil {
					shape = l.checker.GetTypeOfSymbol(symbol)
				}
			}
		}
		if shape != nil && l.privateSlots(shape, 0) && l.checker.IsTypeAssignableTo(shape, view) {
			found = true
			return true
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
		if found {
			return true
		}
	}
	return false
}

// privateSlots reports whether a type's storage holds a private field, its own or a base class's.
func (l *lowering) privateSlots(proven *checker.Type, depth int) bool {
	if depth > 32 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.privateSlots(member, depth+1) {
				return true
			}
		}
		return false
	}
	if l.hasPrivateStorage(proven) {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	if target.ObjectFlags()&checker.ObjectFlagsClass == 0 {
		return false
	}
	for _, base := range l.classBases(proven) {
		if l.privateSlots(base, depth+1) {
			return true
		}
	}
	return false
}
