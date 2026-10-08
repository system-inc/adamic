package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The private proof tracks allocation origins, not reference counts or dead
// bindings. Values monotonically keep every origin from branches and loop trips.
// A private parent can point at shared data; that child never becomes private.
type workerValue struct {
	objects map[*workerObject]bool
	shared  bool
}
type workerObject struct {
	class      *ast.Node
	fields     map[string]workerValue
	elements   workerValue
	collection string
}
type workerFrame struct {
	owner    *ast.Node
	locals   map[*ast.Symbol]workerValue
	receiver workerValue
	result   workerValue
}
type workerProof struct {
	frames               map[*ast.Node]*workerFrame
	constructing         map[*ast.Node]bool
	task                 *parallelProof
	allocations          map[*ast.Node]*workerObject
	active               map[*ast.Node]bool
	changed, constructed bool
	chain                []string
}

func (p *parallelProof) taskFunction(node *ast.Node) error {
	original := p.function(node, nil)
	if original == nil {
		return nil
	}
	private := &workerProof{task: p, allocations: map[*ast.Node]*workerObject{}, active: map[*ast.Node]bool{}, frames: map[*ast.Node]*workerFrame{}, constructing: map[*ast.Node]bool{}}
	// Revisit the whole task until every alias and graph edge is stable. This
	// includes writes in branches and after a use, so a later shared origin cannot
	// hide behind a previously private origin. Failure is always conservative.
	for pass := 0; pass < 32; pass++ {
		private.changed = false
		args := make([]workerValue, len(node.Parameters()))
		for i, param := range node.Parameters() {
			args[i] = private.external(param.Name())
		}
		result, err := private.function(node, args, workerValue{})
		if err != nil {
			if !private.constructed {
				return original
			}
			return err
		}
		if !private.constructed {
			return original
		}
		if len(result.objects) > 0 || result.shared {
			signatures := p.l.checker.GetSignaturesOfType(p.l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)
			if len(signatures) != 1 || p.l.shareable(p.l.checker.GetReturnTypeOfSignature(signatures[0]), "work.result", node) != "" {
				return private.fail(node, "returns a private graph without a Shareable result proof")
			}
		}
		if !private.changed {
			return nil
		}
	}
	return private.fail(node, "private graph fixed point did not converge")
}
func (w *workerProof) fail(node *ast.Node, what string) error {
	return w.task.effect(node, w.chain, "cannot prove worker-private ownership: "+what)
}
func (w *workerProof) external(node *ast.Node) workerValue {
	var reference func(*checker.Type) bool
	reference = func(t *checker.Type) bool {
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range t.Types() {
				if reference(member) {
					return true
				}
			}
			return false
		}
		return t.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0
	}
	return workerValue{shared: reference(w.task.l.checker.GetTypeAtLocation(node))}
}
func (w *workerProof) merge(to *workerValue, from workerValue) {
	if from.shared && !to.shared {
		to.shared = true
		w.changed = true
	}
	for object := range from.objects {
		if to.objects == nil {
			to.objects = map[*workerObject]bool{}
		}
		if !to.objects[object] {
			to.objects[object] = true
			w.changed = true
		}
	}
}
func (w *workerProof) union(to *workerValue, from workerValue) {
	changed := w.changed
	w.merge(to, from)
	w.changed = changed
}
func (w *workerProof) allocation(node *ast.Node, class *ast.Node, collection string) workerValue {
	object := w.allocations[node]
	if object == nil {
		object = &workerObject{class: class, fields: map[string]workerValue{}, collection: collection}
		w.allocations[node] = object
		w.changed = true
	}
	return workerValue{objects: map[*workerObject]bool{object: true}}
}
func (w *workerProof) function(node *ast.Node, args []workerValue, receiver workerValue) (workerValue, error) {
	if w.active[node] {
		return workerValue{}, w.fail(node, "recursive private call requires an ownership summary")
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		return workerValue{}, w.fail(node, "async code is not a synchronous private task")
	}
	w.active[node] = true
	defer delete(w.active, node)
	old := w.chain
	name := "work"
	if node.Name() != nil {
		name = node.Name().Text()
	}
	w.chain = append(append([]string{}, old...), name)
	defer func() { w.chain = old }()
	frame := w.frames[node]
	if frame == nil {
		frame = &workerFrame{owner: node, locals: map[*ast.Symbol]workerValue{}}
		w.frames[node] = frame
	}
	w.merge(&frame.receiver, receiver)
	for i, param := range node.Parameters() {
		if !ast.IsIdentifier(param.Name()) || param.AsParameterDeclaration().DotDotDotToken != nil {
			return workerValue{}, w.fail(param, "non-plain parameter")
		}
		value := workerValue{}
		if i < len(args) {
			value = args[i]
		} else if initial := param.AsParameterDeclaration().Initializer; initial != nil {
			var err error
			value, err = w.expression(initial, frame)
			if err != nil {
				return workerValue{}, err
			}
		}
		old := frame.locals[w.task.symbol(param.Name())]
		w.merge(&old, value)
		frame.locals[w.task.symbol(param.Name())] = old
	}
	if ast.IsExpressionNode(node.Body()) {
		value, err := w.expression(node.Body(), frame)
		return value, err
	}
	if err := w.statement(node.Body(), frame); err != nil {
		return workerValue{}, err
	}
	return frame.result, nil
}
func (w *workerProof) statement(node *ast.Node, frame *workerFrame) error {
	if node == nil || ast.IsPartOfTypeNode(node) {
		return nil
	}
	if node.Kind == ast.KindVariableDeclaration {
		if !ast.IsIdentifier(node.Name()) {
			return w.fail(node, "destructuring local ownership")
		}
		value := workerValue{}
		if init := node.AsVariableDeclaration().Initializer; init != nil {
			var err error
			value, err = w.expression(init, frame)
			if err != nil {
				return err
			}
		} else {
			value = w.external(node.Name())
		}
		symbol := w.task.symbol(node.Name())
		old := frame.locals[symbol]
		w.merge(&old, value)
		frame.locals[symbol] = old
		return nil
	}
	if node.Kind == ast.KindReturnStatement {
		if value := node.AsReturnStatement().Expression; value != nil {
			result, err := w.expression(value, frame)
			if err != nil {
				return err
			}
			w.merge(&frame.result, result)
		}
		return nil
	}
	if node.Kind == ast.KindForOfStatement {
		loop := node.AsForInOrOfStatement()
		array, err := w.expression(loop.Expression, frame)
		if err != nil {
			return err
		}
		element := workerValue{shared: array.shared}
		for object := range array.objects {
			w.union(&element, object.elements)
		}
		if loop.Initializer.Kind != ast.KindVariableDeclarationList || len(loop.Initializer.AsVariableDeclarationList().Declarations.Nodes) != 1 {
			return w.fail(node, "non-declaration for-of binding")
		}
		local := loop.Initializer.AsVariableDeclarationList().Declarations.Nodes[0]
		if !ast.IsIdentifier(local.Name()) {
			return w.fail(local, "destructuring for-of binding")
		}
		if !w.external(local.Name()).shared {
			element = workerValue{}
		}
		old := frame.locals[w.task.symbol(local.Name())]
		w.merge(&old, element)
		frame.locals[w.task.symbol(local.Name())] = old
		return w.statement(loop.Statement, frame)
	}
	if ast.IsFunctionLike(node) {
		return w.fail(node, "a nested closure could retain a private graph")
	}
	if ast.IsExpressionNode(node) {
		_, err := w.expression(node, frame)
		return err
	}
	var found error
	node.ForEachChild(func(child *ast.Node) bool {
		if found == nil {
			found = w.statement(child, frame)
		}
		return found != nil
	})
	return found
}
func (w *workerProof) expression(node *ast.Node, frame *workerFrame) (workerValue, error) {
	node = ast.SkipParentheses(node)
	if node == nil || ast.IsPartOfTypeNode(node) {
		return workerValue{}, nil
	}
	if target := parallelWrite(node); target != nil {
		value := workerValue{}
		if node.Kind == ast.KindBinaryExpression {
			var err error
			value, err = w.expression(node.AsBinaryExpression().Right, frame)
			if err != nil {
				return workerValue{}, err
			}
		}
		if err := w.write(target, value, frame); err != nil {
			return workerValue{}, err
		}
		return value, nil
	}
	if node.Kind == ast.KindThisKeyword {
		if len(frame.receiver.objects) == 0 || frame.receiver.shared {
			return workerValue{}, w.fail(node, "this is not a private receiver")
		}
		return frame.receiver, nil
	}
	if ast.IsIdentifier(node) {
		if value, ok := frame.locals[w.task.symbol(node)]; ok {
			return value, nil
		}
		if node.Text() == "undefined" {
			return workerValue{}, nil
		}
		if err := w.task.capture(node, frame.owner, w.chain); err != nil {
			return workerValue{}, err
		}
		return w.external(node), nil
	}
	if node.Kind == ast.KindNewExpression {
		return w.construct(node, frame)
	}
	if node.Kind == ast.KindCallExpression {
		return w.call(node, frame)
	}
	if ast.IsFunctionLike(node) {
		return workerValue{}, w.fail(node, "a function value could retain a private graph")
	}
	if node.Kind == ast.KindAwaitExpression || node.Kind == ast.KindDeleteExpression {
		return workerValue{}, w.fail(node, "unsupported effect in private task")
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		for _, name := range []string{"Number", "Math", "String"} {
			if w.task.l.isLibraryGlobal(access.Expression, name) && w.task.l.libraryMember(node) && !w.external(node).shared {
				return workerValue{}, nil
			}
		}
		if symbol := w.task.symbol(node); symbol != nil {
			for _, declaration := range symbol.Declarations {
				if declaration.Kind == ast.KindGetAccessor || declaration.Kind == ast.KindSetAccessor {
					return workerValue{}, w.fail(node, "accessor effects are not proven")
				}
			}
		}
		receiver, err := w.expression(access.Expression, frame)
		if err != nil {
			return workerValue{}, err
		}
		if !w.external(node).shared {
			return workerValue{}, nil
		}
		result := workerValue{shared: receiver.shared}
		for object := range receiver.objects {
			value, ok := object.fields[access.Name().Text()]
			if !ok {
				return workerValue{}, w.fail(node, "reference field has no proven origin")
			}
			w.union(&result, value)
		}
		return result, nil
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		receiver, err := w.expression(access.Expression, frame)
		if err != nil {
			return workerValue{}, err
		}
		if _, err = w.expression(access.ArgumentExpression, frame); err != nil {
			return workerValue{}, err
		}
		if !w.external(node).shared {
			return workerValue{}, nil
		}
		result := workerValue{shared: receiver.shared}
		for object := range receiver.objects {
			w.union(&result, object.elements)
		}
		return result, nil
	}
	if node.Kind == ast.KindArrayLiteralExpression {
		value := w.allocation(node, nil, "array")
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			child, err := w.expression(element, frame)
			if err != nil {
				return workerValue{}, err
			}
			for object := range value.objects {
				w.merge(&object.elements, child)
			}
		}
		return value, nil
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		value := w.allocation(node, nil, "")
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if !ast.IsIdentifier(property.Name()) || (property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment) {
				return workerValue{}, w.fail(property, "non-plain object field")
			}
			initializer := property.Name()
			if property.Kind == ast.KindPropertyAssignment {
				initializer = property.AsPropertyAssignment().Initializer
			}
			child, err := w.expression(initializer, frame)
			if err != nil {
				return workerValue{}, err
			}
			for object := range value.objects {
				old := object.fields[property.Name().Text()]
				w.merge(&old, child)
				object.fields[property.Name().Text()] = old
			}
		}
		return value, nil
	}
	// Reference-valued wrappers keep every origin; scalar operators only read.
	result := workerValue{}
	var found error
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if found != nil || ast.IsPartOfTypeNode(child) {
			return false
		}
		if ast.IsExpressionNode(child) {
			value, err := w.expression(child, frame)
			found = err
			w.union(&result, value)
			return false
		}
		return child.ForEachChild(visit)
	}
	node.ForEachChild(visit)
	if found != nil {
		return workerValue{}, found
	}
	if node.Kind == ast.KindAsExpression || node.Kind == ast.KindNonNullExpression {
		return result, nil
	}
	if !w.external(node).shared {
		if (result.shared || len(result.objects) > 0) && node.Kind == ast.KindTemplateExpression {
			return workerValue{}, w.fail(node, "template coercion has unproven object dispatch")
		}
		if (result.shared || len(result.objects) > 0) && node.Kind == ast.KindBinaryExpression {
			operator := node.AsBinaryExpression().OperatorToken.Kind
			if operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
				return workerValue{}, w.fail(node, "operator coercion has unproven object dispatch")
			}
		}
		return workerValue{}, nil
	}
	switch node.Kind {
	case ast.KindConditionalExpression, ast.KindBinaryExpression, ast.KindAsExpression, ast.KindNonNullExpression:
		return result, nil
	}
	return workerValue{}, w.fail(node, "reference expression has no ownership rule")
}
func (w *workerProof) write(target *ast.Node, value workerValue, frame *workerFrame) error {
	target = ast.SkipParentheses(target)
	if ast.IsIdentifier(target) {
		symbol := w.task.symbol(target)
		if _, ok := frame.locals[symbol]; !ok {
			return w.fail(target, "publishes into a capture or global")
		}
		old := frame.locals[symbol]
		w.merge(&old, value)
		frame.locals[symbol] = old
		return nil
	}
	var receiver *ast.Node
	name := ""
	if target.Kind == ast.KindPropertyAccessExpression {
		receiver = target.AsPropertyAccessExpression().Expression
		name = target.Name().Text()
	} else if target.Kind == ast.KindElementAccessExpression {
		receiver = target.AsElementAccessExpression().Expression
		if _, err := w.expression(target.AsElementAccessExpression().ArgumentExpression, frame); err != nil {
			return err
		}
	} else {
		return w.fail(target, "write target is not private")
	}
	held, err := w.expression(receiver, frame)
	if err != nil {
		return err
	}
	if held.shared || len(held.objects) == 0 {
		return w.fail(target, "writes or publishes into a shared receiver")
	}
	if symbol := w.task.symbol(target); symbol != nil {
		for _, d := range symbol.Declarations {
			if d.Kind == ast.KindMethodDeclaration || d.Kind == ast.KindSetAccessor || d.Kind == ast.KindGetAccessor {
				return w.fail(target, "replaces a method or invokes an accessor")
			}
		}
	}
	for object := range held.objects {
		if name == "" {
			w.merge(&object.elements, value)
		} else {
			old := object.fields[name]
			w.merge(&old, value)
			object.fields[name] = old
		}
	}
	return nil
}
