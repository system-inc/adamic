package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A generic value can use the ordinary closure ABI when its context fixes every
// binder. Only one value specialization per declaration is admitted: otherwise
// two forwarders would make the same source function compare unequal.
func (l *lowering) genericFunctionValue(node, declaration *ast.Node) (ir.Expression, error) {
	context := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if context == nil {
		return nil, l.notYet(node, "a generic function as a value without a concrete contextual signature")
	}
	signatures := l.checker.GetSignaturesOfType(l.concrete(context), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 {
		return nil, l.notYet(node, "a generic function as a value without a concrete contextual signature")
	}
	target := l.checker.GetSignatureFromDeclaration(declaration)
	given := signatures[0]
	if target == nil || len(target.Parameters()) > len(given.Parameters()) {
		return nil, l.notYet(node, "a generic function value whose contextual parameter count differs")
	}
	// The closure ABI evaluates every supplied argument, then a forwarder reads
	// only the declaration's parameters. Extra contextual parameters are ignored.
	concreteTypes := map[*checker.Type]*checker.Type{}
	for i, parameter := range target.Parameters() {
		l.inferTypes(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(given.Parameters()[i]), concreteTypes)
	}
	l.inferTypes(l.checker.GetReturnTypeOfSignature(target), l.checker.GetReturnTypeOfSignature(given), concreteTypes)
	substitution := map[*checker.Type]ir.Type{}
	sources, targets := []*checker.Type{}, []*checker.Type{}
	key := "value:" + l.program.Where(declaration)
	for _, parameter := range declaration.TypeParameters() {
		binder := l.checker.GetTypeAtLocation(parameter.Name())
		concrete := concreteTypes[binder]
		if concrete == nil || concrete.Flags()&checker.TypeFlagsTypeParameter != 0 {
			return nil, l.notYet(node, "a generic function value with an unresolved type parameter")
		}
		held, known := l.representation(concrete)
		if !known {
			return nil, l.notYet(node, "a generic function value with an unrepresented type argument")
		}
		substitution[binder] = held
		sources, targets = append(sources, binder), append(targets, concrete)
		key += "," + l.genericTypeKey(concrete)
	}
	prefix := "value:" + l.program.Where(declaration) + ","
	for existingKey := range l.genericInstances {
		if len(existingKey) >= len(prefix) && existingKey[:len(prefix)] == prefix && existingKey != key {
			return nil, l.notYet(node, "multiple specializations of a generic function value need one shared source identity")
		}
	}
	if l.genericDepth >= maximumGenericDepth {
		return nil, l.notYet(node, "a generic function value instantiated without end")
	}
	outerSubstitution, outerLocals, outerClosures, outerMapper := l.substitution, l.locals, l.closures, l.typeMapper
	l.substitution, l.closures, l.typeMapper = substitution, nil, newTypeMapper(sources, targets)
	defer func() {
		l.substitution, l.locals, l.closures, l.typeMapper = outerSubstitution, outerLocals, outerClosures, outerMapper
	}()
	// Inference is only a candidate. Prove its parameter and result types match the
	// context before constructing an ABI, including all nested keeping conventions.
	for i, parameter := range target.Parameters() {
		from, to := l.concrete(l.checker.GetTypeOfSymbol(parameter)), l.checker.GetTypeOfSymbol(given.Parameters()[i])
		if !identicalTypes(l.checker, from, to) || !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {
			return nil, l.notYet(node, "a generic function value whose instantiated parameters differ from its context")
		}
	}
	if !identicalTypes(l.checker, l.concrete(l.checker.GetReturnTypeOfSignature(target)), l.checker.GetReturnTypeOfSignature(given)) {
		return nil, l.notYet(node, "a generic function value whose instantiated result differs from its context")
	}
	if existing, known := l.genericInstances[key]; known {
		return l.functionValue(node, existing)
	}
	if err := l.refuseInstantiatedMutation(declaration); err != nil {
		return nil, err
	}
	l.locals = map[*ast.Symbol]int{}
	for symbol, local := range outerLocals {
		if l.result.Locals[local].Global {
			l.locals[symbol] = local
		}
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: declaration.Name().Text() + "_value_instance_" + strconv.Itoa(index)})
	if l.genericInstances == nil {
		l.genericInstances = map[string]int{}
	}
	l.genericInstances[key] = index
	l.genericDepth++
	defer func() { l.genericDepth-- }()
	if err := l.lowerFunction(index, declaration, -1); err != nil {
		return nil, err
	}
	return l.functionValue(node, index)
}
