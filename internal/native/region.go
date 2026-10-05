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
//     function takes a hidden region parameter: the literal its return makes is allocated in it, and
//     the calls whose results become that literal's fields, or its return, are handed it on. Every
//     other call it makes is handed NULL, the heap. That's Tofte and Talpin's region polymorphism:
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

	// flows says, per function and parameter position, whether that parameter escapes (true) or
	// flows nowhere (false). A position not in it escapes.
	escapes map[int]map[int]bool

	// statements are the statements that get a region.
	statements map[*ir.Statement]bool
}

func planRegions(program *ir.Program) *regionPlan {
	plan := &regionPlan{fresh: map[int]bool{}, escapes: map[int]map[int]bool{}, statements: map[*ir.Statement]bool{}}
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
		return plan.fresh[value.Function]
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
		case ir.Unwrap:
			return derived(expression.Value)
		case ir.CheckedCast:
			for _, allowed := range expression.Allowed {
				derived(allowed)
			}
			return derived(expression.Value)
		case ir.Call:
			for position, argument := range expression.Arguments {
				if derived(argument) && plan.escapes[expression.Function][position] {
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
				if inner, isCall := argument.(ir.Call); isCall && plan.fresh[inner.Function] && !plan.escapes[call.Function][position] {
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

// regionFor is the region a call to a function that takes one is handed: the one its parent said
// (a statement's region, or the function's own), or the heap.
func (e *emitter) regionFor(call ir.Call) string {
	if !e.regions.fresh[call.Function] {
		return ""
	}
	if e.depth == e.regionCallDepth {
		return e.regionCallArgument
	}
	return "NULL"
}

// handRegion evaluates an expression with a region for it, when it's a call to a function that
// takes one; anything else is evaluated as it would be.
func (e *emitter) handRegion(expression ir.Expression, region string) string {
	call, isCall := expression.(ir.Call)
	if region == "" || !isCall || !e.regions.fresh[call.Function] {
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
	outer := e.statementRegion
	e.statementRegion = region
	e.statement(*at)
	e.statementRegion = outer
	e.line("adamic_region_end(&%s);", region)
	return true
}
