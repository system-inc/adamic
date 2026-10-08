package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// nestedDeclarations binds a function body's declarations before any body is lowered.
// Sibling code shares a cell layout, never cells holding sibling closure values.
func (l *lowering) nestedDeclarations(nodes []*ast.Node) ([]ir.Statement, error) {
	if l.function == nil {
		return nil, nil
	}
	declarations := []*ast.Node{}
	for _, node := range nodes {
		if node.Kind != ast.KindFunctionDeclaration {
			continue
		}
		if node.Parent != nil && (node.Parent.Kind == ast.KindCaseClause || node.Parent.Kind == ast.KindDefaultClause) {
			// switchBindings owns the shared case environment and hoists these functions.
			continue
		}
		if node.Parent == nil || node.Parent.Kind != ast.KindBlock || !ast.IsFunctionLike(node.Parent.Parent) {
			return nil, l.notYet(node, "a block-scoped nested function declaration")
		}
		if node.Name() == nil {
			return nil, l.notYet(node, "a generic or unnamed nested function declaration")
		}
		if node.Body() == nil {
			// Overloads describe the implementation, not additional runtime functions.
			// Only skip a signature when this body contains its implementation.
			implemented := false
			for _, candidate := range nodes {
				if candidate.Kind == ast.KindFunctionDeclaration && candidate.Body() != nil && candidate.Name() != nil && l.symbol(candidate.Name()) == l.symbol(node.Name()) {
					implemented = true
					break
				}
			}
			if !implemented {
				return nil, l.notYet(node, "a nested function declaration without an implementation")
			}
			continue
		}
		for _, parameter := range node.Parameters() {
			if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
				return nil, l.notYet(parameter, "dynamic this in a nested function declaration")
			}
		}
		declarations = append(declarations, node)
	}
	if len(declarations) == 0 {
		return nil, nil
	}
	l.function.NestedFrame = true
	owner := l.functionIndex
	prologue := []ir.Statement{}
	// Captures of a declaration later in the source must already have a cell when
	// the hoisted functions are initialized. Noncaptured bindings emit no cell.
	for _, node := range nodes {
		if node.Kind != ast.KindVariableStatement {
			continue
		}
		for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			bound, err := l.preallocateNestedBindings(declaration.Name())
			if err != nil {
				return nil, err
			}
			prologue = append(prologue, bound...)
		}
	}
	functions := []int{}
	bodies := []*ast.Node{}
	for _, node := range declarations {
		local, err := l.declareLocal(node.Name())
		if err != nil {
			return nil, err
		}
		if len(node.TypeParameters()) > 0 {
			if l.generics == nil {
				l.generics = map[*ast.Symbol]*ast.Node{}
			}
			l.generics[l.symbol(node.Name())] = node
			// Negative marks a generic declaration with direct-call instantiations only.
			l.result.Locals[local].NestedFunction = -1
			continue
		}
		index := len(l.result.Functions)
		l.result.Locals[local].NestedFunction = index + 1
		l.result.Functions = append(l.result.Functions, ir.Function{Name: node.Name().Text(), Closure: true, NestedParent: owner + 1})
		l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node.Name())), function: index, node: node})
		if l.instance != nil {
			l.instance.templates = append(l.instance.templates, template{closure: index, proven: l.checker.GetTypeAtLocation(node.Name()), isClosure: true})
		}
		functions = append(functions, index)
		bodies = append(bodies, node)
		if err := l.signature(index, node, -1); err != nil {
			return nil, err
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
			return nil, l.notYet(invalid, "dynamic this in a nested function declaration")
		}
		l.closures = append(l.closures, index)
		outerThis := l.this
		l.this = -1
		err := l.lowerFunction(index, node, -1)
		l.this = outerThis
		l.closures = l.closures[:len(l.closures)-1]
		if err != nil {
			return nil, err
		}
	}
	environment := []int{}
	for _, index := range functions {
		for _, local := range l.result.Functions[index].Environment {
			if !slices.Contains(environment, local) {
				environment = append(environment, local)
			}
		}
	}
	for _, index := range functions {
		l.result.Functions[index].Environment = slices.Clone(environment)
	}
	return prologue, nil
}

func (l *lowering) nestedSibling(node *ast.Node) int {
	if !ast.IsIdentifier(node) || l.function == nil {
		return -1
	}
	local, ok := l.locals[l.symbol(node)]
	if !ok {
		return -1
	}
	index := l.result.Locals[local].NestedFunction - 1
	if index < 0 {
		return -1
	}
	parent := l.result.Functions[index].NestedParent
	if l.function.NestedParent != 0 {
		if parent == l.function.NestedParent {
			return index
		}
		return -1
	}
	if !l.function.Closure {
		return -1
	}
	// Forward only through anonymous closures in the same lexical group.
	// Crossing another named group still requires a separate binding strategy.
	boundary := -1
	for position := len(l.closures) - 1; position >= 0; position-- {
		function := l.closures[position]
		if function == parent-1 {
			boundary = position
			break
		}
		if group := l.result.Functions[function].NestedParent; group != 0 {
			if group != parent {
				return -1
			}
			boundary = position
			break
		}
	}
	for _, function := range l.closures[boundary+1:] {
		if forwarded := l.result.Functions[function].ForwardedNestedParent; forwarded != 0 && forwarded != parent {
			return -1
		}
	}
	for _, function := range l.closures[boundary+1:] {
		l.result.Functions[function].ForwardedNestedParent = parent
	}
	// Known slots propagate now; slots discovered by later sibling bodies are
	// added to every forwarding closure when the enclosing frame is completed.
	for _, captured := range l.result.Functions[index].Environment {
		l.touch(captured)
	}
	return index
}

// References use the canonical value cached weakly in the declaring frame.
// Only its cells travel between functions, never a strong sibling closure binding.
func (l *lowering) nestedReference(node *ast.Node) (ir.Expression, bool, error) {
	local, ok := l.locals[l.symbol(node)]
	if !ok || l.result.Locals[local].NestedFunction == 0 {
		return nil, false, nil
	}
	if l.result.Locals[local].NestedFunction < 0 {
		return nil, true, l.notYet(node, "a generic function as a value")
	}
	if l.result.Locals[local].Function != l.functionIndex {
		index := l.result.Locals[local].NestedFunction - 1
		l.nestedReferenceEnvironment(index)
		return ir.MakeClosure{Function: index}, true, nil
	}
	return ir.Read{Local: local, Of: ir.Closure}, true, nil
}

// A destructuring initializer keeps its source evaluation order; only the cells
// for the names it binds are allocated before hoisted closures can capture them.
func (l *lowering) preallocateNestedBindings(name *ast.Node) ([]ir.Statement, error) {
	if ast.IsIdentifier(name) {
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		l.result.Locals[local].Preallocated = true
		return []ir.Statement{ir.Declare{Local: local, Uninitialized: true}}, nil
	}
	if name.Kind != ast.KindArrayBindingPattern && name.Kind != ast.KindObjectBindingPattern {
		return nil, l.notYet(name, "a nested function's enclosing binding pattern")
	}
	statements := []ir.Statement{}
	for _, binding := range name.AsBindingPattern().Elements.Nodes {
		if skipped(binding) {
			continue
		}
		bound, err := l.preallocateNestedBindings(binding.Name())
		if err != nil {
			return nil, err
		}
		statements = append(statements, bound...)
	}
	return statements, nil
}

// finishNestedEnvironment gives a named-declaration frame one allocation site.
// A cell view retains its entire record, so cycle analysis must see all its slots.
func (l *lowering) finishNestedEnvironment(function *ir.Function, owner int) {
	if !function.NestedFrame {
		return
	}
	function.FrameIdentity = l.result.Functions[owner].FrameIdentity
	if function.FrameIdentity > 0 {
		function.Body = append([]ir.Statement{ir.Declare{Local: function.FrameIdentity - 1, Value: ir.NumberConstant{}}}, function.Body...)
	}
	direct := map[int]bool{}
	for _, statement := range function.Body {
		if declaration, ok := statement.(ir.Declare); ok {
			direct[declaration.Local] = true
		}
	}
	for local, binding := range l.result.Locals {
		if binding.Function != owner || !binding.Captured {
			continue
		}
		if !binding.Preallocated && !direct[local] && !slices.Contains(function.Parameters, local) {
			continue
		}
		l.result.Locals[local].EnvironmentCell = true
		l.result.Locals[local].Preallocated = true
		function.FrameEnvironment = append(function.FrameEnvironment, local)
	}
	if len(function.FrameEnvironment) > 0 {
		function.Body = append([]ir.Statement{ir.AllocateEnvironment{Cells: slices.Clone(function.FrameEnvironment)}}, function.Body...)
	}
	for index := range l.result.Functions {
		target := &l.result.Functions[index]
		retains := target.NestedParent == owner+1 || target.ForwardedNestedParent == owner+1 || slices.Contains(target.ReferenceParents, owner+1)
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
	for _, target := range l.result.Functions {
		if target.NestedParent != owner+1 {
			continue
		}
		for _, local := range target.Environment {
			if !slices.Contains(environment, local) {
				environment = append(environment, local)
			}
		}
	}
	for index := range l.result.Functions {
		target := &l.result.Functions[index]
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
