package lower

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A predicate parameter forwarded where the program takes a narrowing predicate (#jqf1drp, on the road
// rulings.md:140 ruled for #vdt9f1q): tsc's getFirstJSDocTag takes `predicate: (tag: JSDocTag) => tag is T`
// and hands it to core.ts's find, whose narrowing overload's result believes every true answer while its
// implementation holds the callback as a plain boolean one, so nothing there checks the claim. The
// parameter proves nothing, since whatever its caller passed is behind it, so in .ts it's passed as a
// checked closure, built at the pass site: it calls the parameter's value and, when it answers true,
// checks the argument against the parameter's own claim before it hands true on, so a lie stops loudly
// there. Its body is checkedPredicateValueBody's, the forwarder's own. The false side is the walk's
// (predicate_false_flow.go), which the parameter's checked claim already turns on. In .a the argument stays
// refused (predicateArgumentProof). A proof that every function the closure can call checks or proves its
// claim (ClosureTargets) would erase the check; it's a later step, and nothing here assumes it's absent.

// predicateParameterArgument is the parameter declaration argument names, when argument is a predicate
// parameter read as a value: a plain parameter, neither optional, defaulted nor a rest, never assigned,
// written as a function type whose claim names one of its parameters and a type. nil otherwise. An optional
// one could hold undefined, which a closure made around it would hide.
func (l *lowering) predicateParameterArgument(argument *ast.Node) *ast.Node {
	if argument.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := l.symbol(argument)
	if symbol == nil || len(symbol.Declarations) != 1 || l.assignedSymbol(symbol) {
		return nil
	}
	parameter := symbol.Declarations[0]
	if parameter.Kind != ast.KindParameter || parameter.AsParameterDeclaration().QuestionToken != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil || parameter.Initializer() != nil {
		return nil
	}
	written := parameter.Type()
	if written == nil || written.Kind != ast.KindFunctionType || written.Type() == nil || written.Type().Kind != ast.KindTypePredicate {
		return nil
	}
	claim := written.Type().AsTypePredicateNode()
	if claim.AssertsModifier != nil || claim.Type == nil || claim.ParameterName == nil || claim.ParameterName.Kind != ast.KindIdentifier {
		return nil
	}
	return parameter
}

// checkedPredicateParameterSite is whether parameter, a predicate parameter (predicateParameterArgument)
// passed for contract, is passed as its checked closure: its own claim is a checked claim in .ts
// (checkedFunctionTypeClaim), and contract is the program's. A library's contract is called inside the
// library, and a forward to lib's filter or find keeps its refusal (Lower's must, predicates_function_type.go),
// which this step doesn't widen.
func (l *lowering) checkedPredicateParameterSite(contract, parameter *ast.Node) bool {
	file := ast.GetSourceFileOfNode(contract)
	return !load.IsLibrary(file) && !load.IsPrelude(file) && l.checkedFunctionTypeClaim(parameter.Type().Type())
}

// checkedPredicateParameterPassed is whether node, a name read as a value, is a predicate parameter passed
// as a call's argument where predicateArguments admits it as its checked closure: the call's parameter
// there claims a predicate that isn't checked at the calls through it (predicateContract,
// checkedFunctionTypeClaim), and checkedPredicateParameterSite holds. It reads predicateArguments' own
// contract, so lowering builds the closure exactly where the proof admitted the argument for one.
func (l *lowering) checkedPredicateParameterPassed(node *ast.Node) bool {
	parameter := l.predicateParameterArgument(node)
	if parameter == nil {
		return false
	}
	argument := node
	for argument.Parent != nil && argument.Parent.Kind == ast.KindParenthesizedExpression {
		argument = argument.Parent
	}
	call := argument.Parent
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	index := slices.Index(call.AsCallExpression().Arguments.Nodes, argument)
	if index < 0 {
		return false
	}
	signature := l.checker.GetResolvedSignature(call)
	if signature == nil || index >= len(signature.Parameters()) {
		return false
	}
	contract := predicateContract(signature.Parameters()[index])
	return contract != nil && !l.checkedFunctionTypeClaim(contract) && l.checkedPredicateParameterSite(contract, parameter)
}

// checkedPredicateParameterValue is value, node's read, as the argument it's passed as: the checked
// closure when node is a predicate parameter passed where checkedPredicateParameterPassed says, and value
// itself otherwise. The closure is made as bind makes a bound function (function_bind.go): a factory call
// at the pass site takes the parameter's value into a captured cell of its own, which the closure reads,
// so the cell takes part in the cycle proof, and the closure is recorded with the parameter's type, so
// every call through a function type it can stand for counts it among its targets. It takes the
// parameters the parameter's function type takes, calls the value with them, and checks the claimed one
// against the claim on a true answer, in the words "predicate parameter <name> passed at <site>". It's no
// call site, so PredicateChecks doesn't count it, as it doesn't count a forwarder.
func (l *lowering) checkedPredicateParameterValue(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if value.Type() != ir.Closure || !l.checkedPredicateParameterPassed(node) {
		return value, nil
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, l.notYet(node, "a predicate parameter passed as a checked closure without one call signature")
	}
	claim := predicateOfSignature(l.checker, signatures[0])
	if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || claim.Type() == nil {
		return nil, l.notYet(node, "a checked predicate without a reifiable target")
	}
	parameters := signatures[0].Parameters()
	index := int(claim.ParameterIndex())
	if index < 0 || index >= len(parameters) {
		return nil, l.notYet(node, "a checked predicate on a missing argument")
	}
	maker := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "checked_predicate_parameter_factory", Returns: ir.Closure})
	wrapper := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "checked_predicate_parameter", Closure: true, Returns: ir.Boolean})
	captured := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "checked_predicate", Type: ir.Closure, Function: maker, Captured: true})
	l.noteLocal(captured, proven, node)
	function := l.result.Functions[wrapper]
	function.Source.Role = "checked_predicate_parameter"
	function.Environment = []int{captured}
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: wrapper})
		return index
	}
	arguments := []ir.Expression{}
	for _, parameter := range parameters {
		if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil {
			return nil, l.notYet(node, "a predicate parameter passed as a checked closure whose function type takes a rest")
		}
		of, known := l.bindParameterType(parameter)
		if !known {
			return nil, l.notYet(node, "a predicate parameter passed as a checked closure taking a parameter without a value representation")
		}
		taken := local("argument", of)
		function.Parameters = append(function.Parameters, taken)
		arguments = append(arguments, ir.Read{Local: taken, Of: of})
	}
	call := ir.CallClosure{Closure: ir.Read{Local: captured, Of: ir.Closure}, Arguments: arguments, FunctionType: int(proven.Id()), Returns: ir.Boolean}
	source := l.checker.GetTypeOfSymbol(parameters[index])
	where := l.program.Where(node)
	sourceName, targetName := l.checker.TypeToString(source), l.checker.TypeToString(claim.Type())
	body, err := l.checkedPredicateValueBody(node, node, call, arguments[index], source, claim.Type(), false, local, func(branch string) string {
		return fmt.Sprintf("predicate parameter %s passed at %s, %s branch: source %s, target %s", node.Text(), where, branch, sourceName, targetName)
	})
	if err != nil {
		return nil, err
	}
	function.Body = body
	l.result.Functions[wrapper] = function
	l.closureRecords = append(l.closureRecords, closureRecord{proven: proven, function: wrapper, node: node})
	if l.instance != nil {
		l.instance.templates = append(l.instance.templates, template{closure: wrapper, proven: l.checker.GetTypeAtLocation(node), isClosure: true})
	}
	made, err := l.programAllocation(node, ir.MakeClosure{Function: wrapper})
	if err != nil {
		return nil, err
	}
	l.result.Functions[maker].Parameters = []int{captured}
	l.result.Functions[maker].Body = []ir.Statement{ir.Return{Value: made}}
	if l.predicateParametersBuilt == nil {
		l.predicateParametersBuilt = map[*ast.Node]bool{}
	}
	l.predicateParametersBuilt[node] = true
	return ir.Call{Function: maker, Arguments: []ir.Expression{value}, Returns: ir.Closure}, nil
}

// checkCheckedPredicateParameters stops a predicate parameter's checked closure wherever a function it can
// call observes its argument count, collects a rest or declares more parameters than the closure passes:
// the closure passes on the arguments its function type names, which can be fewer than its caller passed
// (find passes an index the claim's type may not name), as checkBoundArguments stops a bound function. A
// value assignable to the type may declare more, `(shape: Shape, index?: number) => shape is Circle`,
// and Node hands it find's index where the closure would leave it undefined. Forwarding the whole
// argument list is the later fix. And every parameter predicateArgumentProof admitted for its closure
// that lowering lowered must have been passed as one, whatever road lowering took to the argument: one
// lowered another way would hand the program's narrowing predicate the parameter raw, so it stops there,
// the first in the source first. Only the lowered regions count (walkLoweredRegions): the contracts pass
// admits every call in every module, but a body lowering never lowers, such as an uncalled generic
// forwarder's, passes nothing at all, so stopping it would stop a program with no hazard in it. It runs
// once FunctionTypeTargets is known.
func (l *lowering) checkCheckedPredicateParameters(modules []*ast.SourceFile) error {
	unbuilt := []*ast.Node{}
	l.walkLoweredRegions(modules, func(node *ast.Node) {
		if l.predicateParametersAdmitted[node] && !l.predicateParametersBuilt[node] {
			unbuilt = append(unbuilt, node)
		}
	})
	if len(unbuilt) > 0 {
		first := slices.MinFunc(unbuilt, func(a, b *ast.Node) int {
			return cmp.Or(strings.Compare(string(ast.GetSourceFileOfNode(a).FileName()), string(ast.GetSourceFileOfNode(b).FileName())), cmp.Compare(a.Pos(), b.Pos()))
		})
		return l.notYet(first, "a predicate parameter admitted for its checked closure that lowering passed another way")
	}
	for _, record := range l.closureRecords {
		function := l.result.Functions[record.function]
		if function.Source.Role != "checked_predicate_parameter" {
			continue
		}
		var found error
		walk(function.Body, func(node any) bool {
			if call, ok := node.(ir.CallClosure); ok {
				for _, target := range l.result.FunctionTypeTargets[call.FunctionType] {
					candidate := l.result.Functions[target]
					// Parameters holds the declared parameters in order (a destructured or defaulted one
					// arrives as one incoming local), behind a method's leading receiver, which Receiver
					// marks and no argument fills, so it's left out of the count. The environment is
					// Environment, never Parameters. A lone this parameter is counted, which can only stop
					// more.
					declared := len(candidate.Parameters)
					if candidate.Receiver {
						declared--
					}
					if candidate.ReadsArguments || candidate.RestElement != 0 || declared > len(call.Arguments) {
						found = l.notYet(record.node, "a predicate parameter passed as a checked closure, which may call a function that observes its argument count, collects rest arguments or declares more parameters than its type names")
					}
				}
			}
			return found == nil
		})
		if found != nil {
			return found
		}
	}
	return nil
}
