package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"sort"
)

// Quantified declarations are compared after alpha-renaming, without method
// bivariance. Only module producers are supported: their identity and environment
// remain unchanged when an instantiated body is selected at a call.
func (l *lowering) genericCallableShape(target *checker.Type) bool {
	signatures := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) == 0 || signatures[0].HasRestParameter() || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	for _, parameter := range signatures[0].TypeParameters() {
		constraint := l.checker.GetBaseConstraintOfType(parameter)
		if constraint == nil || !l.viewCallableBoxedRepresentation(constraint) {
			return false
		}
	}
	return true
}

func (l *lowering) genericCallableMatches(producer, wanted *checker.Signature) bool {
	if producer == nil || len(producer.TypeParameters()) != len(wanted.TypeParameters()) || len(producer.Parameters()) != len(wanted.Parameters()) || producer.HasRestParameter() || producer.MinArgumentCount() != wanted.MinArgumentCount() {
		return false
	}
	mapper := newTypeMapper(producer.TypeParameters(), wanted.TypeParameters())
	for i, parameter := range producer.TypeParameters() {
		actual, expected := l.checker.GetBaseConstraintOfType(parameter), l.checker.GetBaseConstraintOfType(wanted.TypeParameters()[i])
		if actual == nil || expected == nil || !identicalTypes(l.checker, instantiateType(l.checker, actual, mapper), expected) {
			return false
		}
	}
	for i, parameter := range producer.Parameters() {
		if !identicalTypes(l.checker, instantiateType(l.checker, l.censusCallableParameterType(parameter), mapper), l.censusCallableParameterType(wanted.Parameters()[i])) {
			return false
		}
	}
	return identicalTypes(l.checker, instantiateType(l.checker, l.checker.GetReturnTypeOfSignature(producer), mapper), l.checker.GetReturnTypeOfSignature(wanted))
}

func (l *lowering) genericCallableTemplate(declaration *ast.Node) int {
	key := "view-template:" + l.program.Where(declaration)
	if index, exists := l.genericInstances[key]; exists {
		return index
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: declaration.Name().Text() + "_template", Closure: true, GenericTemplate: true, Body: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("generic callable: template requires a call instantiation")}}}})
	if l.genericInstances == nil {
		l.genericInstances = map[string]int{}
	}
	l.genericInstances[key] = index
	return index
}

func (l *lowering) genericCallableValue(node, declaration *ast.Node) (ir.Expression, error) {
	if !l.genericCallableShape(l.checker.GetTypeAtLocation(declaration.Name())) {
		return nil, l.notYet(node, "a generic function as a value")
	}
	if declaration.Body() == nil {
		return nil, l.notYet(declaration, "a generic callable without a module body")
	}
	index := l.genericCallableTemplate(declaration)
	if held, exists := l.forwarders[index]; exists {
		return ir.Read{Local: held, Of: ir.Closure}, nil
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: declaration.Name().Text() + "_value", Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[index] = held
	return ir.Read{Local: held, Of: ir.Closure}, nil
}

func (l *lowering) genericCallableProducers(wanted *checker.Signature) []*ast.Node {
	producers := []*ast.Node{}
	for _, declaration := range l.generics {
		if declaration.Body() != nil && l.genericCallableMatches(l.checker.GetSignatureFromDeclaration(declaration), wanted) {
			producers = append(producers, declaration)
		}
	}
	sort.Slice(producers, func(i, j int) bool { return l.program.Where(producers[i]) < l.program.Where(producers[j]) })
	return producers
}

func (l *lowering) prepareGenericCallableRead(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	signature := l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)[0]
	id := l.result.ViewContractTypes[int(target.Id())]
	if id == 0 {
		var err error
		id, err = buildViewCallableContract(l, node, target, nil)
		if err != nil {
			return 0, err
		}
	}
	contract := l.result.ViewContracts[id-1]
	contract.Generic = true
	contract.Functions = nil
	for _, declaration := range l.genericCallableProducers(signature) {
		contract.Functions = append(contract.Functions, l.genericCallableTemplate(declaration))
	}
	sources, targets := signature.TypeParameters(), []*checker.Type{}
	for _, parameter := range sources {
		targets = append(targets, l.checker.GetBaseConstraintOfType(parameter))
	}
	mapper := newTypeMapper(sources, targets)
	child := func(proven *checker.Type) (ir.ViewContractID, error) {
		proven = instantiateType(l.checker, proven, mapper)
		of, known := l.representation(proven)
		if !known {
			return 0, l.notYet(node, "a generic callable constraint representation")
		}
		child := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: of, Name: l.checker.TypeToString(proven), RepresentationMask: l.viewCallableRepresentationMask(proven)})
		return child, nil
	}
	contract.Parameters = nil
	for _, parameter := range signature.Parameters() {
		parameterID, err := child(l.censusCallableParameterType(parameter))
		if err != nil {
			return 0, err
		}
		contract.Parameters = append(contract.Parameters, parameterID)
	}
	result, err := child(l.checker.GetReturnTypeOfSignature(signature))
	if err != nil {
		return 0, err
	}
	contract.Result = result
	contract.Unsupported = ""
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

func (l *lowering) genericCallableInstances(call *ast.Node, signatures []*checker.Signature) ([][2]int, error) {
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) == 0 {
		return nil, nil
	}
	if !l.genericCallableShape(l.checker.GetTypeAtLocation(call.AsCallExpression().Expression)) {
		return nil, l.notYet(call, "a generic callable whose constraint cannot be represented")
	}
	resolved := l.checker.GetResolvedSignature(call)
	if resolved == nil || resolved.Declaration() == nil || resolved.Declaration() != signatures[0].Declaration() {
		return nil, l.notYet(call, "a resolved generic callable outside its declared signature")
	}
	instances := [][2]int{}
	for _, declaration := range l.genericCallableProducers(signatures[0]) {
		template := l.genericCallableTemplate(declaration)
		instance, err := l.instantiateFunction(call, declaration)
		if err != nil {
			return nil, err
		}
		_, err = l.functionValue(declaration.Name(), instance)
		if err != nil {
			return nil, err
		}
		held := l.forwarders[instance]
		forwarder := -1
		for _, value := range l.forwarderValues {
			if declared, ok := value.(ir.Declare); ok && declared.Local == held {
				forwarder = declared.Value.(ir.MakeClosure).Function
			}
		}
		if forwarder < 0 {
			return nil, l.notYet(call, "a generic callable without its instantiated body")
		}
		instances = append(instances, [2]int{template, forwarder})
	}
	// A nonempty dispatch distinguishes a reached generic call from an ordinary
	// callable even when no producer has a compatible quantified declaration.
	if len(instances) == 0 {
		instances = append(instances, [2]int{-1, -1})
	}
	return instances, nil
}

// Recheck every returned type in the instantiation, independently of the generic
// annotation. Assertions and nested returns supply no proof here.
func (l *lowering) checkGenericCallableReturns(call, declaration *ast.Node, promised *checker.Type) error {
	var failure error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			value := node.AsReturnStatement().Expression
			if value == nil || !l.checker.IsTypeAssignableTo(l.concrete(l.checker.GetTypeAtLocation(value)), promised) {
				failure = l.notYet(call, "a generic callable result outside its instantiation")
				return false
			}
			var scan ast.Visitor
			scan = func(part *ast.Node) bool {
				if part.Kind == ast.KindAsExpression || part.Kind == ast.KindTypeAssertionExpression || part.Kind == ast.KindNonNullExpression {
					failure = l.notYet(call, "an asserted generic callable result without an instantiation proof")
				}
				return part.ForEachChild(scan)
			}
			value.ForEachChild(scan)
			if value.Kind == ast.KindAsExpression || value.Kind == ast.KindTypeAssertionExpression {
				failure = l.notYet(call, "an asserted generic callable result without an instantiation proof")
			}
		}
		return node.ForEachChild(visit)
	}
	declaration.Body().ForEachChild(visit)
	return failure
}
