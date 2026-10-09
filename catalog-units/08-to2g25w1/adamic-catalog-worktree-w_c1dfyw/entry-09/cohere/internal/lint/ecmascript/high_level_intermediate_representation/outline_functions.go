// Outlining function expressions that capture nothing.
//
// React's `Optimization/OutlineFunctions.ts`, run at `Pipeline.ts:359` -- before scope assignment,
// which is what makes it a scope-structure pass rather than a codegen one.
//
// # What it does and why the timing matters
//
// A `FunctionExpression` closing over nothing is the same function on every render, so it does not
// need to be built each time and does not need memoizing. Upstream lifts it to a module-level
// declaration and replaces the expression with a `LoadGlobal` of that name.
//
// Because this runs before `InferReactiveScopeVariables`, the lifted lambda never gets a reactive
// scope at all. Measured on `repro-object-fromEntries-entries.js`, where upstream compiles both
// callbacks to `_temp` and `_temp2` at module scope: this tree kept two `FunctionExpression` scopes
// upstream does not have, and every attempt to explain that inside the escape analysis found a
// faithful port. The scopes were never supposed to exist.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"strconv"
	"strings"
)

// OutlineFunctions replaces capture-free function expressions with a load of an outlined function.
//
// Returns how many were outlined, for measurement.
//
// # The gate is upstream's, and each clause earns its place
//
// A function that captures something cannot be lifted: the captured values live in the enclosing
// frame. A named function expression is excluded because upstream has not handled the naming
// collision, and its own comment says so. An fbt operand is excluded because the fbt transform reads
// the expression back and a `LoadGlobal` is not what it expects; this tree has no fbt handling, so
// that clause has nothing to select and is not ported -- recorded here rather than silently absent.
//
// Recursive, because an inner function may itself be capture-free.
func OutlineFunctions(function *Function) int {
	if function == nil {
		return 0
	}
	outlined := 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			expression, isFunction := instruction.Value.(*FunctionExpression)
			if !isFunction {
				continue
			}
			nested := function.Functions[expression.Function]
			// Recurse first, matching upstream: an inner capture-free function is outlined whether
			// or not its parent can be.
			outlined += OutlineFunctions(nested)
			if nested == nil || len(nested.Context) != 0 || len(expression.Captures) != 0 {
				continue
			}
			if nested.Name != "" {
				// Upstream: "TODO: handle outlining named functions".
				continue
			}
			// The lifted function's name is not observable here -- nothing in this tree emits code,
			// and every consumer of a `LoadGlobal` reads `BindingKind` rather than resolving the
			// name -- so a stable synthetic name is enough. Upstream generates a unique one through
			// its environment for the same reason its codegen needs one.
			instruction.Value = &LoadGlobal{
				Name:        outlinedFunctionName(expression.Function),
				BindingKind: GlobalBindingKindGlobal,
			}
			// The `Functions` entry is now reachable only through this map. Upstream keeps the
			// lowered function in its environment for the same reason: outlining moves a function,
			// it does not discard one, and a later pass that reads the callee's own effects still
			// needs to find it.
			if function.Outlined == nil {
				function.Outlined = map[static_single_assignment.IdentifierId]FunctionId{}
			}
			function.Outlined[instruction.LValue.Identifier] = expression.Function
			outlined++
		}
	}
	return outlined
}

// outlinedFunctionName is the synthetic module-level name a lifted function takes.
//
// Upstream spells these `_temp`, `_temp2` and so on. The numbering here is by `FunctionId` rather
// than by a counter, which is stable across runs and unique within one function.
// outlinedFunctionByName inverts `outlinedFunctionName`, resolving a `LoadGlobal` back to the
// function it names.
//
// The `Outlined` map is keyed by the identifier the `LoadGlobal` was written into, and that key does
// not survive a second `Construct`: inlining an immediately invoked function expression re-runs SSA,
// every identifier is renumbered, and the map still holds the pre-inline keys. Measured on
// `error.validate-object-values-mutation`, where the callback resolver found nothing after the
// inline and `values.map` stopped widening its receiver's range, dropping the golden.
//
// The name is the durable half. It is synthesized from the function id and nothing rewrites it, so
// reading it back is the same answer the map was built to give and it is renumber-proof.
func outlinedFunctionByName(function *Function, name string) (FunctionId, bool) {
	if function == nil || name == "" || !strings.HasPrefix(name, "_temp") {
		return 0, false
	}
	suffix := strings.TrimPrefix(name, "_temp")
	if suffix == "" {
		return 0, len(function.Functions) > 0
	}
	parsed, err := strconv.Atoi(suffix)
	if err != nil || parsed < 1 {
		return 0, false
	}
	id := FunctionId(parsed - 1)
	if int(id) >= len(function.Functions) {
		return 0, false
	}
	return id, true
}

func outlinedFunctionName(id FunctionId) string {
	if id == 0 {
		return "_temp"
	}
	return "_temp" + strconv.Itoa(int(id)+1)
}
