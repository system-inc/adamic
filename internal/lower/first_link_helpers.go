package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A helper is a function lowering makes with no declaration of its own, which functionNodes never names:
// rangeFieldDispatch's range_field_read, an array builder's, a dispatcher. Its code is the code of the units that
// use it, so a first-link lowering stubs it with them and never alone (Compiler's ruling on #nxb3xh1). On tsc a
// trap of textRangeContainsPositionInclusive stubbed every function declared inside it and left its range_field_read
// helper holding the read checkViewWidenedReads had stopped at, so the pass met that read again, a Gap.
//
// A helper's owners are the unit whose lowering made it (firstLinkHelperUnits) and every unit whose code refers to
// it: a declared function's unit, a top-level statement's, or, through a helper that refers to it, that helper's
// owners, transitively. A helper is stubbed only once every owner is trapped, as the stub of the unit that made
// it. One with a live owner keeps its code, so a pass that meets it again records a Gap and abandons its judgment
// loudly, and every run-time trap's reason stays true whatever a builder caches across sites. A reference no unit
// holds, from a statement of Main outside every top-level statement's IR or from the program's own tables, is a
// live owner, which fails closed. A helper no unit made, a whole-program pass's, has no owner and is left: its stop
// names no site (stopWhere), a Gap.

// firstLinkNoteHelpers notes the unit whose lowering made each helper among the functions from on. The innermost
// unit wins, since a nested body lowers, and notes, before the body holding it.
func (l *lowering) firstLinkNoteHelpers(from int, unit *ast.Node) {
	for index := from; index < len(l.result.Functions); index++ {
		if l.functionNodes[index] != nil || l.firstLinkHelperUnits[index] != nil {
			continue
		}
		if l.firstLinkHelperUnits == nil {
			l.firstLinkHelperUnits = map[int]*ast.Node{}
		}
		l.firstLinkHelperUnits[index] = unit
	}
}

// firstLinkStubHelpers stubs every helper whose owners are all trapped, as the stub of the unit that made it, and
// leaves every other. A function noted while its declaration's body was still to lower, a generic instance's, has
// its declaration by now and is no helper.
func (l *lowering) firstLinkStubHelpers() {
	pending := false
	for index, unit := range l.firstLinkHelperUnits {
		if _, trapped := l.firstLinkTrappedAt(unit); trapped && l.functionNodes[index] == nil && !firstLinkStubBody(l.result.Functions[index].Body) {
			pending = true
		}
	}
	if !pending {
		return
	}
	uses := l.firstLinkHelperUses()
	const (
		weighing = iota + 1
		dead
		live
	)
	state := map[int]int{}
	var trappedWhole func(index int) bool
	trappedWhole = func(index int) bool {
		switch state[index] {
		case dead:
			return true
		case weighing, live:
			// A helper that refers back to itself, through any number of others, is weighed live, failing closed.
			return false
		}
		state[index] = weighing
		_, trapped := l.firstLinkTrappedAt(l.firstLinkHelperUnits[index])
		trapped = trapped && l.functionNodes[index] == nil
		for _, use := range uses[index] {
			if !trapped {
				break
			}
			switch {
			case use.function >= 0 && l.functionNodes[use.function] != nil:
				_, trapped = l.firstLinkTrappedAt(l.functionNodes[use.function])
			case use.function >= 0:
				trapped = trappedWhole(use.function)
			case use.statement != nil:
				_, trapped = l.firstLinkTrappedAt(use.statement)
			default:
				trapped = false
			}
		}
		state[index] = live
		if trapped {
			state[index] = dead
		}
		return trapped
	}
	for index, unit := range l.firstLinkHelperUnits {
		if trappedWhole(index) {
			stub, _ := l.firstLinkTrappedAt(unit)
			l.result.Functions[index].Body = []ir.Statement{stub}
		}
	}
}

// firstLinkHelperUse is where a helper is referred to: from a function's body, from a top-level statement's IR in
// Main, or, with neither, from somewhere no unit holds.
type firstLinkHelperUse struct {
	function  int
	statement *ast.Node
}

// firstLinkHelperUses is every reference to each noted helper: from each function's body, from each statement of
// Main by the top-level statement whose IR holds it, and from the program's tables, which no unit holds. Locals are
// passed over, since a local's Function is the function declaring it, never a reference.
func (l *lowering) firstLinkHelperUses() map[int][]firstLinkHelperUse {
	uses := map[int][]firstLinkHelperUse{}
	note := func(use firstLinkHelperUse) func(int) {
		return func(function int) {
			if _, helper := l.firstLinkHelperUnits[function]; helper {
				uses[function] = append(uses[function], use)
			}
		}
	}
	statements := map[int]*ast.Node{}
	for statement := range l.firstLinkMain {
		if start, length, placed := l.firstLinkMainAt(statement); placed {
			for index := start; index < start+length; index++ {
				statements[index] = statement
			}
		}
	}
	for index, statement := range l.result.Main {
		firstLinkFunctionReferences(reflect.ValueOf(statement), "", map[uintptr]bool{}, note(firstLinkHelperUse{function: -1, statement: statements[index]}))
	}
	for index := range l.result.Functions {
		function := reflect.ValueOf(l.result.Functions[index])
		for field := 0; field < function.NumField(); field++ {
			name := function.Type().Field(field).Name
			use := firstLinkHelperUse{function: index}
			if name != "Body" {
				use.function = -1
			}
			firstLinkFunctionReferences(function.Field(field), name, map[uintptr]bool{}, note(use))
		}
	}
	program := reflect.ValueOf(l.result).Elem()
	for field := 0; field < program.NumField(); field++ {
		name := program.Type().Field(field).Name
		if name == "Main" || name == "Functions" || name == "Locals" || !program.Type().Field(field).IsExported() {
			continue
		}
		firstLinkFunctionReferences(program.Field(field), name, map[uintptr]bool{}, note(firstLinkHelperUse{function: -1}))
	}
	return uses
}

// firstLinkFunctionFields are the IR's fields that name a function by its index, among them a call's, a closure's,
// a comparator's, a class's constructor and accessors, a method's, a nested function's parent and the method
// targets table. A field of another name naming one would be missed, so a new one belongs here.
var firstLinkFunctionFields = map[string]bool{"Function": true, "Functions": true, "Comparator": true, "Constructor": true, "Getter": true, "Setter": true, "SpreadGetters": true, "NestedFunction": true, "NestedParent": true, "ForwardedNestedParent": true, "MethodTargets": true}

// firstLinkFunctionReferences notes every function index value holds in a field firstLinkFunctionFields names, field
// being the name of the field holding value. seen keeps a pointer cyclic in the IR from being walked twice in one
// walk; each body and table is walked with its own, so a pointer two of them share is noted for each.
func firstLinkFunctionReferences(value reflect.Value, field string, seen map[uintptr]bool, note func(int)) {
	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			firstLinkFunctionReferences(value.Elem(), field, seen, note)
		}
	case reflect.Pointer:
		if value.IsNil() || seen[value.Pointer()] {
			return
		}
		seen[value.Pointer()] = true
		firstLinkFunctionReferences(value.Elem(), field, seen, note)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if value.Type().Field(index).IsExported() {
				firstLinkFunctionReferences(value.Field(index), value.Type().Field(index).Name, seen, note)
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			firstLinkFunctionReferences(value.Index(index), field, seen, note)
		}
	case reflect.Map:
		entries := value.MapRange()
		for entries.Next() {
			firstLinkFunctionReferences(entries.Key(), field, seen, note)
			firstLinkFunctionReferences(entries.Value(), field, seen, note)
		}
	case reflect.Int:
		if firstLinkFunctionFields[field] {
			note(int(value.Int()))
		}
	}
}
