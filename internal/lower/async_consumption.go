package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
)

// A rejecting promise needs a consumer before control can leave its scope. A
// direct synchronous return transfers that obligation to the caller. Aliases
// remain the same obligation; a fulfilled library promise needs no rejection
// observer. This deliberately does not infer an observer in another closure.
func (l *lowering) checkPromiseConsumption() error {
	initializers := map[int]ir.Expression{}
	mutable := map[int]bool{}
	collect := func(statements []ir.Statement) {
		walk(statements, func(node any) bool {
			switch value := node.(type) {
			case ir.Declare:
				if value.Value != nil && value.Value.Type() == ir.Promise {
					initializers[value.Local] = value.Value
				}
			case ir.Assign:
				if value.Value.Type() == ir.Promise {
					mutable[value.Local] = true
				}
			}
			return true
		})
	}
	collect(l.result.Main)
	for _, function := range l.result.Functions {
		collect(function.Body)
	}
	var origin func(int, map[int]bool) int
	origin = func(local int, seen map[int]bool) int {
		if seen[local] || mutable[local] {
			return local
		}
		seen[local] = true
		if read, ok := initializers[local].(ir.Read); ok {
			return origin(read.Local, seen)
		}
		return local
	}
	// A graph with no escaping Throw has handled every rejection and synchronous
	// exception. Unknown targets and ordinary Promise forwarders remain unproved.
	safeFunctions := map[int]bool{}
	for index, function := range l.result.Functions {
		if !function.Async {
			continue
		}
		safe := true
		for _, block := range flow.Build(l.result, index).Blocks {
			if _, throws := block.Terminal.(*flow.Throw); throws {
				safe = false
			}
		}
		safeFunctions[index] = safe
	}
	var fulfilled func(ir.Expression, map[int]bool) bool
	fulfilled = func(value ir.Expression, seen map[int]bool) bool {
		switch value := value.(type) {
		case ir.PromiseValue:
			return !value.Reject
		case ir.Read:
			if seen[value.Local] || mutable[value.Local] {
				return false
			}
			seen[value.Local] = true
			return fulfilled(initializers[value.Local], seen)
		case ir.Call:
			targets := l.result.CallTargets(value)
			if len(targets) == 0 {
				return false
			}
			for _, target := range targets {
				if !safeFunctions[target] {
					return false
				}
			}
			return true
		}
		return false
	}
	var consumedAfter func([]ir.Statement, int) bool
	consumedAfter = func(statements []ir.Statement, local int) bool {
		for _, statement := range statements {
			var expression ir.Expression
			switch value := statement.(type) {
			case ir.Declare:
				expression = value.Value
			case ir.Evaluate:
				expression = value.Value
			case ir.Return:
				read, ok := value.Value.(ir.Read)
				return ok && read.Of == ir.Promise && origin(read.Local, map[int]bool{}) == origin(local, map[int]bool{})
			case ir.Block:
				if consumedAfter(value.Body, local) {
					return true
				}
			case ir.If:
				if consumedAfter(value.Then, local) && consumedAfter(value.Else, local) {
					return true
				}
			case ir.Try:
				if consumedAfter(value.Body, local) {
					return true
				}
			}
			if awaited, ok := expression.(ir.Await); ok {
				if read, ok := awaited.Value.(ir.Read); ok && origin(read.Local, map[int]bool{}) == origin(local, map[int]bool{}) {
					return true
				}
			}
			// A prior rejection or abrupt completion could bypass the later consumer.
			barrier := false
			walk(statement, func(node any) bool {
				switch value := node.(type) {
				case ir.Return, ir.Throw, ir.Panic, ir.Break, ir.Continue:
					barrier = true
				case ir.Await:
					barrier = barrier || !fulfilled(value.Value, map[int]bool{})
				}
				return !barrier
			})
			at := statement
			if barrier || flow.CanThrow(l.result, &flow.Instruction{At: &at, Expression: expression}) {
				return false
			}
		}
		return false
	}
	var checkScope func([]ir.Statement, *ast.Node) error
	checkScope = func(statements []ir.Statement, where *ast.Node) error {
		for index, statement := range statements {
			if declared, ok := statement.(ir.Declare); ok && declared.Value != nil && declared.Value.Type() == ir.Promise {
				node := where
				if original := l.localNodes[declared.Local]; original != nil {
					node = original
				}
				if !fulfilled(declared.Value, map[int]bool{}) && (mutable[declared.Local] || l.result.Locals[declared.Local].Captured || !consumedAfter(statements[index+1:], declared.Local)) {
					return l.notYet(node, "a promise not awaited in the same function before an escaping completion")
				}
			}
			var failure error
			var inspect func(reflect.Value, bool)
			inspect = func(value reflect.Value, consumed bool) {
				if !value.IsValid() || failure != nil {
					return
				}
				if value.CanInterface() {
					if body, ok := value.Interface().([]ir.Statement); ok {
						failure = checkScope(body, where)
						return
					}
					if expression, ok := value.Interface().(ir.Expression); ok {
						if expression.Type() == ir.Promise && !consumed && !fulfilled(expression, map[int]bool{}) {
							if _, read := expression.(ir.Read); !read {
								failure = l.notYet(where, "a promise not awaited in the same function before an escaping completion")
								return
							}
						}
						if awaited, ok := expression.(ir.Await); ok {
							inspect(reflect.ValueOf(awaited.Value), true)
							return
						}
					}
					switch node := value.Interface().(type) {
					case ir.Declare:
						inspect(reflect.ValueOf(node.Value), true)
						return
					case ir.Return:
						inspect(reflect.ValueOf(node.Value), true)
						return
					}
				}
				switch value.Kind() {
				case reflect.Interface, reflect.Pointer:
					if !value.IsNil() {
						inspect(value.Elem(), consumed)
					}
				case reflect.Struct:
					for field := 0; field < value.NumField(); field++ {
						inspect(value.Field(field), false)
					}
				case reflect.Slice, reflect.Array:
					for item := 0; item < value.Len(); item++ {
						inspect(value.Index(item), false)
					}
				}
			}
			inspect(reflect.ValueOf(statement), false)
			if failure != nil {
				return failure
			}
		}
		return nil
	}
	if err := checkScope(l.result.Main, l.entry.AsNode()); err != nil {
		return err
	}
	for _, function := range l.result.Functions {
		where := l.entry.AsNode()
		if function.AsyncSite > 0 {
			where = l.asyncSites[function.AsyncSite-1]
		}
		if err := checkScope(function.Body, where); err != nil {
			return err
		}
	}
	return nil
}
