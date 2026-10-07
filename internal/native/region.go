package native

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// Regions, the arenas of docs/memory.md (#272q6cv): a statement that makes values nothing reaches
// after it gets a region, and they're allocated in it and let go of together when it ends
// (runtime/region.c), instead of one malloc and one free each.
//
// Three facts decide it, each a fixed point over the program's functions:
//
//   - A function returns fresh when every value it returns is an object literal made in the return
//     itself (no spread), undefined, or what a call to a function that returns fresh returns. Such a
//     function has a second version taking a hidden region parameter: the literal its return makes
//     is allocated in it, and the calls whose results become that literal's fields, or its return,
//     are handed it on, to their own region versions. Every other call it makes is to a heap
//     version. That's Tofte and Talpin's region polymorphism:
//     which region a value lives in is the caller's to say, never a global's.
//   - A parameter flows nowhere when nothing the function does with it, or with a value read out of
//     it (a field, a field's field), lets it outlive the call: it's only read, tested, compared, and
//     handed to parameters that flow nowhere themselves. Stored, returned, assigned, captured, handed
//     to a closure or to anything else, it escapes.
//   - A statement gets a region when it hands what a call to a function that returns fresh returns
//     straight to a parameter that flows nowhere, as total += check(build(depth)) does: that value
//     and everything made into it are dead when the statement ends, and the region with them.
type regionPlan struct {
	// fresh are the functions that take a region.
	fresh map[int]bool

	// referenceWrites is conservative: an operation/callee may write a reference after allocation.
	referenceWrites map[int]bool

	// flows says, per function and parameter position, whether that parameter escapes (true) or
	// flows nowhere (false). A position not in it escapes.
	escapes map[int]map[int]bool

	// statements are the statements that get a region.
	statements   map[*ir.Statement]bool
	program      *ir.Program
	classObjects map[int]int
}

func planRegions(program *ir.Program) *regionPlan {
	plan := &regionPlan{fresh: map[int]bool{}, referenceWrites: map[int]bool{}, escapes: map[int]map[int]bool{}, statements: map[*ir.Statement]bool{}, program: program, classObjects: map[int]int{}}
	// Escape: start from every object parameter of a named function flowing nowhere, and mark any
	// that escapes, until none changes.
	for index, function := range program.Functions {
		plan.escapes[index] = map[int]bool{}
		for position, parameter := range function.Parameters {
			local := program.Locals[parameter]
			plan.escapes[index][position] = function.Closure || local.Captured || local.Type != ir.Object
		}
	}
	for changed := true; changed; {
		changed = false
		for index, function := range program.Functions {
			for position, parameter := range function.Parameters {
				if !plan.escapes[index][position] && plan.escaping(function.Body, parameter) {
					plan.escapes[index][position] = true
					changed = true
				}
			}
		}
	}
	// Fresh: start from every named function returning an object, and take away any that returns
	// something else, until none changes.
	for index, function := range program.Functions {
		if !function.Closure && function.Returns == ir.Object {
			plan.fresh[index] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for index := range plan.fresh {
			if !plan.returnsFresh(program, index) {
				delete(plan.fresh, index)
				changed = true
			}
		}
	}
	// A nonescaping consumer may still replace a regional object's reference field. It has no
	// hidden region argument, so force cleanup before calling any consumer that may do that.
	for changed := true; changed; {
		changed = false
		for index, function := range program.Functions {
			if !plan.referenceWrites[index] && plan.writesReferences(function.Body) {
				plan.referenceWrites[index] = true
				changed = true
			}
		}
	}
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for index := range list {
			switch statement := list[index].(type) {
			case ir.Assign, ir.Declare, ir.Evaluate, ir.WriteLine:
				if plan.feedsRegion(statement) {
					plan.statements[&list[index]] = true
				}
			}
			walkStatement(list[index], func(ir.Expression) {}, statements)
		}
	}
	for index := range program.Functions {
		statements(program.Functions[index].Body)
	}
	statements(program.Main)
	return plan
}

// returnsFresh reports whether every value a function returns is fresh, as the plan stands.
func (plan *regionPlan) returnsFresh(program *ir.Program, function int) bool {
	// Class allocators return their declared object after initialization. Its identity and
	// fields stay in that allocation; it may use a region only if no other path exposes it.
	for _, class := range program.Classes {
		if class.Constructor != function {
			continue
		}
		body := program.Functions[function].Body
		if len(body) < 2 {
			return false
		}
		declaration, ok := body[0].(ir.Declare)
		if !ok {
			return false
		}
		literal, ok := declaration.Value.(ir.ObjectLiteral)
		if !ok || literal.Class == 0 {
			return false
		}
		returned, ok := body[len(body)-1].(ir.Return)
		if !ok {
			return false
		}
		read, ok := returned.Value.(ir.Read)
		if !ok || read.Local != declaration.Local {
			return false
		}
		// A closure that captures the object (this, in the source) holds it through a cell that
		// escaping doesn't follow, since MakeClosure names only its function: keep it on the heap, as
		// a captured parameter is (planRegions).
		if program.Locals[declaration.Local].Captured || plan.escaping(body[1:len(body)-1], declaration.Local) {
			return false
		}
		plan.classObjects[function] = declaration.Local
		return true
	}
	fresh := true
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			if returned, ok := statement.(ir.Return); ok && !plan.freshValue(returned.Value) {
				fresh = false
			}
			walkStatement(statement, func(ir.Expression) {}, statements)
		}
	}
	statements(program.Functions[function].Body)
	return fresh
}

// freshValue reports whether a returned value is one a region can hold: an object literal made right
// there, undefined, or what a function that returns fresh returns.
func (plan *regionPlan) freshValue(value ir.Expression) bool {
	switch value := value.(type) {
	case ir.ObjectLiteral:
		return value.Spread == nil
	case ir.Undefined:
		return true
	case ir.Call:
		return plan.regionTarget(value) >= 0
	}
	return false
}

// escaping reports whether a parameter, or anything read out of it, can outlive the call, as the
// plan stands.
func (plan *regionPlan) escaping(body []ir.Statement, parameter int) bool {
	escaped := false
	// derived reports whether an expression's value may be the parameter's object or something read
	// out of it, and marks the escape when such a value is used for more than looking at.
	var derived func(expression ir.Expression) bool
	derived = func(expression ir.Expression) bool {
		switch expression := expression.(type) {
		case nil:
			return false
		case ir.Read:
			return expression.Local == parameter
		case ir.Property:
			return derived(expression.Object)
		case ir.Narrow:
			return derived(expression.Value)
		case ir.Defined:
			// A check that the value is there, and the value itself.
			return derived(expression.Value)
		case ir.Unwrap:
			return derived(expression.Value)
		case ir.CheckedCast:
			for _, allowed := range expression.Allowed {
				derived(allowed)
			}
			return derived(expression.Value)
		case ir.Call:
			for position, argument := range expression.Arguments {
				if derived(argument) && plan.callEscapes(expression, position) {
					escaped = true
				}
			}
			return false
		}
		// Anything else may hold on to an operand unless it's a consumer, done with its operands once
		// it's evaluated (borrow.go): a comparison, a test for undefined, a length.
		operands := false
		eachOperand(expression, func(operand ir.Expression) {
			if derived(operand) {
				operands = true
			}
		})
		if operands && !consumes(expression) {
			escaped = true
		}
		return false
	}
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			switch statement := statement.(type) {
			case ir.SetProperty:
				// Writing into the parameter's object is looking at it; writing it, or something read out
				// of it, anywhere is letting it go.
				derived(statement.Object)
				if derived(statement.Value) {
					escaped = true
				}
				continue
			case ir.SetIndex:
				derived(statement.Array)
				derived(statement.Index)
				if derived(statement.Value) {
					escaped = true
				}
				continue
			}
			// Every expression a statement holds directly is its own: a value derived from the
			// parameter there is declared, assigned, returned, printed or iterated, and escapes. A
			// condition is a boolean, which can't be.
			walkStatement(statement, func(ir.Expression) {}, statements)
			eachOperandOfStatement(statement, func(expression ir.Expression) {
				if derived(expression) {
					escaped = true
				}
			})
		}
	}
	statements(body)
	return escaped
}

// feedsRegion reports whether a statement hands what a fresh function returns straight to a
// parameter that flows nowhere: the statement's region holds it.
func (plan *regionPlan) feedsRegion(statement ir.Statement) bool {
	feeds := false
	walk(statement, func(expression ir.Expression) {
		if call, ok := expression.(ir.Call); ok {
			for position, argument := range call.Arguments {
				if inner, isCall := argument.(ir.Call); isCall && plan.regionTarget(inner) >= 0 && !plan.callEscapes(call, position) {
					feeds = true
				}
			}
		}
	})
	return feeds
}

// eachOperandOfStatement visits the expressions a statement holds directly: its own, not those of
// the statements it holds, and not what's inside them.
func eachOperandOfStatement(statement ir.Statement, visit func(ir.Expression)) {
	var each func(value reflect.Value)
	each = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if expression, ok := value.Interface().(ir.Expression); ok {
				visit(expression)
				return
			}
			each(value.Elem())
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				each(value.Field(index))
			}
		case reflect.Slice:
			if _, nested := value.Interface().([]ir.Statement); nested {
				return
			}
			for index := 0; index < value.Len(); index++ {
				each(value.Index(index))
			}
		}
	}
	each(reflect.ValueOf(statement))
}

// regionVariants is the versions a function is emitted in: every function on the heap, and a
// function that returns fresh again in a region. The region version is handed a region that is never
// NULL, so it knows which of its values are in it, and holds those without counting (regionValue);
// the heap version is the function as it was before regions.
func (e *emitter) regionVariants(function int) []bool {
	if e.regions.fresh[function] {
		return []bool{false, true}
	}
	return []bool{false}
}

// regionFunctionName is the C name of a fresh function's region version.
func (e *emitter) regionFunctionName(function int) string {
	return e.functionName(function) + "_in"
}

// regionFor is the region a call to a function that takes one is handed: the one its parent said
// (a statement's region, or the function's own), or none, and then it's the heap version's call.
func (e *emitter) regionFor(call ir.Call) string {
	if e.regions.regionTarget(call) < 0 || e.depth != e.regionCallDepth {
		return ""
	}
	return e.regionCallArgument
}

// regionValue puts an object made in a region in a temporary the statement doesn't own: immortal
// until its region ends, it takes no retain and no release, held in a field or returned.
func (e *emitter) regionValue(value string) string {
	name := e.temporary()
	e.line("adamic_object *%s = %s;", name, value)
	if e.regionValues == nil {
		e.regionValues = map[string]bool{}
	}
	e.regionValues[name] = true
	return name
}

// handRegion evaluates an expression with a region for it, when it's a call to a function that
// takes one; anything else is evaluated as it would be.
func (e *emitter) handRegion(expression ir.Expression, region string) string {
	call, isCall := expression.(ir.Call)
	if region == "" || !isCall || e.regions.regionTarget(call) < 0 {
		return e.value(expression)
	}
	outerDepth, outerArgument := e.regionCallDepth, e.regionCallArgument
	e.regionCallDepth, e.regionCallArgument = e.depth+1, region
	value := e.value(expression)
	e.regionCallDepth, e.regionCallArgument = outerDepth, outerArgument
	return value
}

// regionStatement emits a statement that gets a region: the region before it, and its end after.
func (e *emitter) regionStatement(at *ir.Statement) bool {
	if !e.regions.statements[at] {
		return false
	}
	region := e.temporary()
	e.line("adamic_region %s = ADAMIC_REGION;", region)
	if e.regions.writesReferences([]ir.Statement{*at}) {
		e.line("%s.holds_outside = true;", region)
	}
	outer := e.statementRegion
	e.statementRegion = region
	e.statement(*at)
	e.statementRegion = outer
	e.line("adamic_region_end(&%s);", region)
	return true
}

// Every possible implementation must finish with the argument before its region can end.
func (plan *regionPlan) callEscapes(call ir.Call, position int) bool {
	targets := plan.program.CallTargets(call)
	if len(targets) == 0 {
		return true
	}
	for _, target := range targets {
		escapes, known := plan.escapes[target][position]
		if !known || escapes {
			return true
		}
	}
	return false
}

// regionTarget selects a variant only when there is one proven target and the
// call uses the direct ABI. Virtual tables contain heap variants, even when all
// targets return fresh, so virtual calls stay on the counted heap.
func (plan *regionPlan) regionTarget(call ir.Call) int {
	targets := plan.program.CallTargets(call)
	if call.Virtual != 0 || len(targets) != 1 || !plan.fresh[targets[0]] {
		return -1
	}
	return targets[0]
}

// writesReferences is deliberately broader than a parameter-specific mutation summary. Reads and
// plain literals are safe; reference stores, opaque operations (callbacks included), and calls
// that transitively contain either require cleanup. Newly added operations fall back to cleanup.
func (plan *regionPlan) writesReferences(body []ir.Statement) bool {
	writes := false
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			if store, ok := statement.(ir.SetProperty); ok && store.Value.Type().IsReference() {
				writes = true
			}
			walkStatement(statement, func(expression ir.Expression) {
				switch expression := expression.(type) {
				case ir.Call:
					targets := plan.program.CallTargets(expression)
					if len(targets) == 0 {
						writes = true
					}
					for _, target := range targets {
						writes = writes || plan.referenceWrites[target]
					}
				case ir.ObjectLiteral, ir.MakeClosure:
					// These allocate new values; their operands are visited separately.
				default:
					writes = writes || !pureKind(expression)
				}
			}, statements)
		}
	}
	statements(body)
	return writes
}
