package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// nestedDeclarations binds a function body's declarations before any body is lowered.
// Sibling code shares a cell layout, never cells holding sibling closure values.
func (lowering *lowering) nestedDeclarations(nodes []*ast.Node) ([]ir.Statement, error) {
	if lowering.function == nil {
		return nil, nil
	}
	declarations := []*ast.Node{}
	for _, node := range nodes {
		if node.Kind != ast.KindFunctionDeclaration {
			continue
		}
		if generatorDeclaration(node) && !loweredGenerator(node) {
			// Before a generic one is put by to instantiate where it's called: its declaration is
			// reached here, whether or not anything calls it (#qk8rztp, open question 7).
			return nil, lowering.generatorNotYet(node, generatorForm(node))
		}
		if generatorDeclaration(node) && lowering.function.Suspends() {
			// A suspending parent's locals live in its own frame's cells, which no closure's
			// environment can be yet (async_normalize.go's canonical suspension stop, and
			// checkGeneratorBody's), so a nested generator there would be a closure over them.
			return nil, lowering.generatorNotYet(node, "declared inside an async function or a generator")
		}
		if node.Parent != nil && (node.Parent.Kind == ast.KindCaseClause || node.Parent.Kind == ast.KindDefaultClause) {
			// switchBindings owns the shared case environment and hoists these functions.
			continue
		}
		if node.Parent == nil || node.Parent.Kind != ast.KindBlock || !ast.IsFunctionLike(node.Parent.Parent) {
			return nil, lowering.notYet(node, "a block-scoped nested function declaration")
		}
		if node.Name() == nil {
			return nil, lowering.notYet(node, "a generic or unnamed nested function declaration")
		}
		if node.Body() == nil {
			// Overloads describe the implementation, not additional runtime functions.
			// Only skip a signature when this body contains its implementation.
			implemented := false
			for _, candidate := range nodes {
				if candidate.Kind == ast.KindFunctionDeclaration &&
					candidate.Body() != nil &&
					candidate.Name() != nil &&
					lowering.symbol(candidate.Name()) == lowering.symbol(node.Name()) {
					implemented = true
					break
				}
			}
			if !implemented {
				return nil, lowering.notYet(node, "a nested function declaration without an implementation")
			}
			continue
		}
		for _, parameter := range node.Parameters() {
			if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
				return nil, lowering.notYet(parameter, "dynamic this in a nested function declaration")
			}
		}
		declarations = append(declarations, node)
	}
	if len(declarations) == 0 {
		return nil, nil
	}
	lowering.function.NestedFrame = true
	owner := lowering.functionIndex
	prologue := []ir.Statement{}
	// Captures of a declaration later in the source must already have a cell when
	// the hoisted functions are initialized. Noncaptured bindings emit no cell.
	for _, node := range nodes {
		if node.Kind != ast.KindVariableStatement {
			continue
		}
		for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			bound, preallocateNestedBindingsError := lowering.preallocateNestedBindings(declaration.Name())
			if preallocateNestedBindingsError != nil {
				return nil, preallocateNestedBindingsError
			}
			prologue = append(prologue, bound...)
		}
	}
	functions := []int{}
	bodies := []*ast.Node{}
	for _, node := range declarations {
		local, declareLocalError := lowering.declareLocal(node.Name())
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		if len(node.TypeParameters()) > 0 {
			if lowering.generics == nil {
				lowering.generics = map[*ast.Symbol]*ast.Node{}
			}
			lowering.generics[lowering.symbol(node.Name())] = node
			// Negative marks a generic declaration with direct-call instantiations only.
			lowering.result.Locals[local].NestedFunction = -1
			continue
		}
		index := len(lowering.result.Functions)
		lowering.result.Locals[local].NestedFunction = index + 1
		lowering.result.Functions = append(
			lowering.result.Functions,
			ir.Function{Name: node.Name().Text(), Closure: true, NestedParent: owner + 1},
		)
		lowering.closureRecords = append(
			lowering.closureRecords,
			closureRecord{
				proven:   lowering.concrete(lowering.checker.GetTypeAtLocation(node.Name())),
				function: index,
				node:     node,
			},
		)
		if lowering.instance != nil {
			lowering.instance.templates = append(
				lowering.instance.templates,
				template{closure: index, proven: lowering.checker.GetTypeAtLocation(node.Name()), isClosure: true},
			)
		}
		functions = append(functions, index)
		bodies = append(bodies, node)
		if signatureError := lowering.signature(index, node, -1); signatureError != nil {
			return nil, signatureError
		}
		prologue = append(prologue, ir.Declare{Local: local, Value: ir.MakeClosure{Function: index}})
	}
	for position, node := range bodies {
		index := functions[position]
		// Ordinary declarations own their this. Dynamic receivers are not implemented.
		var invalid *ast.Node
		var visit ast.Visitor
		visit = func(inner *ast.Node) bool {
			if ast.IsFunctionLike(inner) && inner.Kind != ast.KindArrowFunction {
				return false
			}
			if inner.Kind == ast.KindThisKeyword && !ast.IsPartOfTypeNode(inner) {
				invalid = inner
				return true
			}
			return inner.ForEachChild(visit)
		}
		node.Body().ForEachChild(visit)
		if invalid != nil {
			return nil, lowering.notYet(invalid, "dynamic this in a nested function declaration")
		}
		lowering.closures = append(lowering.closures, index)
		outerThis := lowering.this
		lowering.this = -1
		lowerFunctionError := lowering.lowerFunction(index, node, -1)
		lowering.this = outerThis
		lowering.closures = lowering.closures[:len(lowering.closures)-1]
		if lowerFunctionError != nil {
			return nil, lowerFunctionError
		}
	}
	environment := []int{}
	for _, index := range functions {
		for _, local := range lowering.result.Functions[index].Environment {
			if !slices.Contains(environment, local) {
				environment = append(environment, local)
			}
		}
	}
	for _, index := range functions {
		lowering.result.Functions[index].Environment = slices.Clone(environment)
	}
	return prologue, nil
}

func (lowering *lowering) nestedSibling(node *ast.Node) int {
	if !ast.IsIdentifier(node) || lowering.function == nil {
		return -1
	}
	local, ok := lowering.locals[lowering.symbol(node)]
	if !ok {
		return -1
	}
	index := lowering.result.Locals[local].NestedFunction - 1
	if index < 0 {
		return -1
	}
	parent := lowering.result.Functions[index].NestedParent
	if lowering.function.NestedParent != 0 {
		if parent == lowering.function.NestedParent {
			return index
		}
		return -1
	}
	if !lowering.function.Closure {
		return -1
	}
	// Forward only through anonymous closures in the same lexical group.
	// Crossing another named group still requires a separate binding strategy.
	boundary := -1
	for position := len(lowering.closures) - 1; position >= 0; position-- {
		function := lowering.closures[position]
		if function == parent-1 {
			boundary = position
			break
		}
		if group := lowering.result.Functions[function].NestedParent; group != 0 {
			if group != parent {
				return -1
			}
			boundary = position
			break
		}
	}
	for _, function := range lowering.closures[boundary+1:] {
		if forwarded := lowering.result.Functions[function].ForwardedNestedParent; forwarded != 0 &&
			forwarded != parent {
			return -1
		}
	}
	for _, function := range lowering.closures[boundary+1:] {
		lowering.result.Functions[function].ForwardedNestedParent = parent
	}
	// Known slots propagate now; slots discovered by later sibling bodies are
	// added to every forwarding closure when the enclosing frame is completed.
	for _, captured := range lowering.result.Functions[index].Environment {
		lowering.touch(captured)
	}
	return index
}

// References use the canonical value cached weakly in the declaring frame.
// Only its cells travel between functions, never a strong sibling closure binding.
func (lowering *lowering) nestedReference(node *ast.Node) (ir.Expression, bool, error) {
	local, ok := lowering.locals[lowering.symbol(node)]
	if !ok || lowering.result.Locals[local].NestedFunction == 0 {
		return nil, false, nil
	}
	if lowering.result.Locals[local].NestedFunction < 0 {
		return nil, true, lowering.notYet(node, "a generic function as a value")
	}
	if lowering.result.Functions[lowering.result.Locals[local].NestedFunction-1].Generator {
		// A call through the value would make a frame no consumer keeps to itself, as a module
		// generator's functionValue stops (#qk8rztp S3a). Its own calls take their carrier from
		// nestedGeneratorCall, never through here.
		return nil, true, lowering.generatorNotYet(node, "used as a value")
	}
	return lowering.nestedValue(local), true, nil
}

// nestedValue is the function value of the nested declaration local binds: the declaring function's
// own binding where it is read there, and elsewhere one made with every cell of its group's layout
// carried in (nestedReferenceEnvironment).
func (lowering *lowering) nestedValue(local int) ir.Expression {
	if lowering.result.Locals[local].Function != lowering.functionIndex {
		index := lowering.result.Locals[local].NestedFunction - 1
		lowering.nestedReferenceEnvironment(index)
		return ir.MakeClosure{Function: index}
	}
	return ir.Read{Local: local, Of: ir.Closure}
}

// A destructuring initializer keeps its source evaluation order; only the cells
// for the names it binds are allocated before hoisted closures can capture them.
func (lowering *lowering) preallocateNestedBindings(name *ast.Node) ([]ir.Statement, error) {
	if ast.IsIdentifier(name) {
		local, declareLocalError := lowering.declareLocal(name)
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		lowering.result.Locals[local].Preallocated = true
		return []ir.Statement{ir.Declare{Local: local, Uninitialized: true}}, nil
	}
	if name.Kind != ast.KindArrayBindingPattern && name.Kind != ast.KindObjectBindingPattern {
		return nil, lowering.notYet(name, "a nested function's enclosing binding pattern")
	}
	statements := []ir.Statement{}
	for _, binding := range name.AsBindingPattern().Elements.Nodes {
		if skipped(binding) {
			continue
		}
		bound, preallocateNestedBindingsError := lowering.preallocateNestedBindings(binding.Name())
		if preallocateNestedBindingsError != nil {
			return nil, preallocateNestedBindingsError
		}
		statements = append(statements, bound...)
	}
	return statements, nil
}

// finishNestedEnvironment gives a named-declaration frame one allocation site.
// A cell view retains its entire record, so cycle analysis must see all its slots.
func (lowering *lowering) finishNestedEnvironment(function *ir.Function, owner int) {
	if !function.NestedFrame {
		return
	}
	function.FrameIdentity = lowering.result.Functions[owner].FrameIdentity
	if function.FrameIdentity > 0 {
		function.Body = append(
			[]ir.Statement{ir.Declare{Local: function.FrameIdentity - 1, Value: ir.NumberConstant{}}},
			function.Body...,
		)
	}
	direct := map[int]bool{}
	for _, statement := range function.Body {
		if declaration, ok := statement.(ir.Declare); ok {
			direct[declaration.Local] = true
		}
	}
	for local, binding := range lowering.result.Locals {
		if binding.Function != owner || !binding.Captured {
			continue
		}
		if !binding.Preallocated && !direct[local] && !slices.Contains(function.Parameters, local) {
			continue
		}
		lowering.result.Locals[local].EnvironmentCell = true
		lowering.result.Locals[local].Preallocated = true
		function.FrameEnvironment = append(function.FrameEnvironment, local)
	}
	if len(function.FrameEnvironment) > 0 {
		function.Body = append(
			[]ir.Statement{ir.AllocateEnvironment{Cells: slices.Clone(function.FrameEnvironment)}},
			function.Body...,
		)
	}
	for index := range lowering.result.Functions {
		target := &lowering.result.Functions[index]
		retains := target.NestedParent == owner+1 ||
			target.ForwardedNestedParent == owner+1 ||
			slices.Contains(target.ReferenceParents, owner+1)
		for _, local := range target.Environment {
			retains = retains || slices.Contains(function.FrameEnvironment, local)
		}
		if !retains {
			continue
		}
		for _, local := range function.FrameEnvironment {
			if !slices.Contains(target.Environment, local) {
				target.Environment = append(target.Environment, local)
			}
		}
	}
	// Preserve equal layouts after adding slots retained by other frame closures.
	environment := []int{}
	for _, target := range lowering.result.Functions {
		if target.NestedParent != owner+1 {
			continue
		}
		for _, local := range target.Environment {
			if !slices.Contains(environment, local) {
				environment = append(environment, local)
			}
		}
	}
	for index := range lowering.result.Functions {
		target := &lowering.result.Functions[index]
		if target.NestedParent == owner+1 {
			target.Environment = slices.Clone(environment)
		} else if target.ForwardedNestedParent == owner+1 || slices.Contains(target.ReferenceParents, owner+1) {
			for _, local := range environment {
				if !slices.Contains(target.Environment, local) {
					target.Environment = append(target.Environment, local)
				}
			}
		}
	}
}
