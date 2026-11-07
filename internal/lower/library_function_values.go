package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A library function that reads no receiver is a function value as it stands (#4q1b2er head 2):
// Date.now is `() => number`, and calling it through any binding gives the clock, so reading it as a
// value loses nothing, unlike a method whose this is its object. tsc's performanceCore.ts picks it as
// timestamp's fallback (`nativePerformanceTime ? () => nativePerformanceTime.now() : Date.now`).
//
// Only an allowlisted member whose direct call already lowers is admitted, and only where it's read
// into a slot that sees it exactly as itself: the same parameters and the same result, `() => number`.
// The function value returns an unboxed number, which a wider slot (`() => unknown`, `() => number |
// string`) would read as boxed. It lowers to one function value whose body is the direct call, made
// before anything runs and held by a global of its own, as a module function's (functionValue) and
// Error.captureStackTrace's (nodeCaptureStackTraceValue) are, so Date.now read twice is === as Node
// says. An === operand is not a slot, so `f === Date.now` stays refused, and that refusal is what keeps
// the backends agreeing: the JavaScript backend's value is a wrapper, not Date.now itself, so whoever
// lifts it must make the comparison hold on both. Every other read stays refused as a method read as a
// value. Intl.Collator's compare is not one
// of these: it reads its collator (#zjndrx5, #9tw2b9z).

// libraryFunctionValue is an allowlisted receiver-free library member: its global and name, its
// ECMAScript length, and its direct call.
type libraryFunctionValue struct {
	global, name string
	length       int
	call         func() ir.Expression
}

// libraryFunctionValues are the members a function value can stand for.
var libraryFunctionValues = []libraryFunctionValue{
	{global: "Date", name: "now", length: 0, call: dateNowCall},
}

// libraryFunctionMember is the allowlisted member node reads, unless node reads none or calls it.
func (lowering *lowering) libraryFunctionMember(node *ast.Node) (libraryFunctionValue, bool) {
	if node.Kind != ast.KindPropertyAccessExpression || called(node) || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return libraryFunctionValue{}, false
	}
	for _, member := range libraryFunctionValues {
		if node.Name().Text() == member.name &&
			lowering.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, member.global) {
			return member, true
		}
	}
	return libraryFunctionValue{}, false
}

// libraryFunctionValueRead reports whether node reads an allowlisted member into a slot that sees it
// exactly as itself, the one read of it methodValueRefusal admits (libraryMethodReadAllowed).
func (lowering *lowering) libraryFunctionValueRead(node *ast.Node) bool {
	_, known := lowering.libraryFunctionMember(node)
	return known && lowering.exactFunctionSlot(node)
}

// exactFunctionSlot reports whether the slot node goes into is a plain function type with node's own
// one signature: no this, no rest, no type parameters, the same parameter types in order and the same
// result type, by the checker's identity, and nothing else besides.
func (lowering *lowering) exactFunctionSlot(node *ast.Node) bool {
	slot := lowering.slotType(node)
	if slot == nil {
		return false
	}
	slot = lowering.concrete(slot)
	if slot.Flags()&checker.TypeFlagsObject == 0 ||
		len(lowering.checker.GetPropertiesOfType(slot)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(slot, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	viewed := lowering.checker.GetSignaturesOfType(slot, checker.SignatureKindCall)
	own := lowering.checker.GetSignaturesOfType(lowering.checker.GetTypeAtLocation(node), checker.SignatureKindCall)
	if len(viewed) != 1 || len(own) != 1 {
		return false
	}
	for _, signature := range []*checker.Signature{viewed[0], own[0]} {
		if signature.ThisParameter() != nil ||
			signature.Flags()&checker.SignatureFlagsHasRestParameter != 0 ||
			len(signature.TypeParameters()) != 0 {
			return false
		}
	}
	if len(viewed[0].Parameters()) != len(own[0].Parameters()) {
		return false
	}
	for index, parameter := range viewed[0].Parameters() {
		if lowering.concrete(
			lowering.checker.GetTypeOfSymbol(parameter),
		) != lowering.checker.GetTypeOfSymbol(own[0].Parameters()[index]) {
			return false
		}
	}
	return lowering.concrete(
		lowering.checker.GetReturnTypeOfSignature(viewed[0]),
	) == lowering.checker.GetReturnTypeOfSignature(own[0])
}

// libraryFunctionValueOf lowers an admitted read to the member's one function value, held by a global.
// Any other node is not handled, and lowers as it always has.
func (lowering *lowering) libraryFunctionValueOf(node *ast.Node) (ir.Expression, bool, error) {
	member, known := lowering.libraryFunctionMember(node)
	if !known || !lowering.exactFunctionSlot(node) {
		return nil, false, nil
	}
	role := "library_value_" + member.global + "_" + member.name
	for index, function := range lowering.result.Functions {
		if function.Source.Role == role {
			return ir.Read{Local: lowering.forwarders[index], Of: ir.Closure}, true, nil
		}
	}
	call := member.call()
	index := len(lowering.result.Functions)
	function := ir.Function{
		Name:         role,
		SourceLength: member.length,
		Closure:      true,
		Returns:      call.Type(),
		Body:         []ir.Statement{ir.Return{Value: call}},
	}
	function.Source.Role = role
	lowering.result.Functions = append(lowering.result.Functions, function)
	lowering.closureRecords = append(
		lowering.closureRecords,
		closureRecord{proven: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), function: index, node: node},
	)
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: role, Type: ir.Closure, Global: true, Function: -1},
	)
	lowering.forwarderValues = append(
		lowering.forwarderValues,
		ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}},
	)
	if lowering.forwarders == nil {
		lowering.forwarders = map[int]int{}
	}
	lowering.forwarders[index] = held
	return ir.Read{Local: held, Of: ir.Closure}, true, nil
}
