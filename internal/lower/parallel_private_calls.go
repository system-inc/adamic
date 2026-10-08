package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func (w *workerProof) construct(node *ast.Node, frame *workerFrame) (workerValue, error) {
	expression := node.AsNewExpression()
	args := []workerValue{}
	for _, arg := range nodesOf(expression.Arguments) {
		value, err := w.expression(arg, frame)
		if err != nil {
			return workerValue{}, err
		}
		args = append(args, value)
	}
	callee := ast.SkipParentheses(expression.Expression)
	for _, name := range []string{"Map", "Set"} {
		if w.task.l.isLibraryGlobal(callee, name) {
			if len(args) > 0 {
				return workerValue{}, w.fail(node, "collection initializer ownership is not proven")
			}
			return w.allocation(node, nil, name), nil
		}
	}
	if w.task.l.isLibraryGlobal(callee, "Error") {
		return w.allocation(node, nil, "Error"), nil
	}
	symbol := w.task.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindClassDeclaration {
		return workerValue{}, w.fail(node, "constructor target is not a known source class")
	}
	class := symbol.Declarations[0]
	if w.constructing[class] {
		return workerValue{}, w.fail(node, "recursive construction is not proven")
	}
	w.constructing[class] = true
	defer delete(w.constructing, class)
	if class.AsClassDeclaration().HeritageClauses != nil && len(class.AsClassDeclaration().HeritageClauses.Nodes) > 0 {
		return workerValue{}, w.fail(node, "constructor hierarchy is not proven")
	}
	value := w.allocation(node, class, "")
	w.constructed = true
	var constructor *ast.Node
	for _, member := range class.AsClassDeclaration().Members.Nodes {
		if member.Kind == ast.KindConstructor {
			if constructor != nil {
				return workerValue{}, w.fail(node, "multiple constructor targets")
			}
			constructor = member
		}
	}
	owner := class
	if constructor != nil {
		owner = constructor
	}
	initial := &workerFrame{owner: owner, locals: map[*ast.Symbol]workerValue{}, receiver: value}
	for _, member := range class.AsClassDeclaration().Members.Nodes {
		if member.Kind != ast.KindPropertyDeclaration || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if !ast.IsIdentifier(member.Name()) {
			return workerValue{}, w.fail(member, "computed/private field initializer")
		}
		child := workerValue{}
		if initializer := member.AsPropertyDeclaration().Initializer; initializer != nil {
			var err error
			child, err = w.expression(initializer, initial)
			if err != nil {
				return workerValue{}, err
			}
		} else {
			continue
		}
		for object := range value.objects {
			old := object.fields[member.Name().Text()]
			w.merge(&old, child)
			object.fields[member.Name().Text()] = old
		}
	}
	if constructor != nil {
		returned, err := w.function(constructor, args, value)
		if err != nil {
			return workerValue{}, err
		}
		if returned.shared || len(returned.objects) > 0 {
			return workerValue{}, w.fail(constructor, "constructor returns an alternate receiver")
		}
	}
	return value, nil
}

func (w *workerProof) methodStable(method *ast.Node) bool {
	modules, err := w.task.l.moduleOrder(w.task.l.program.Files()[0])
	if err != nil {
		return false
	}
	stable := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if target := parallelWrite(node); target != nil {
			target = ast.SkipParentheses(target)
			if target.Kind == ast.KindPropertyAccessExpression && target.Name().Text() == method.Name().Text() {
				stable = false
			}
			if target.Kind == ast.KindElementAccessExpression && w.task.l.checker.GetTypeAtLocation(target.AsElementAccessExpression().ArgumentExpression).Flags()&checker.TypeFlagsNumberLike == 0 {
				stable = false
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return stable
}

func (w *workerProof) call(node *ast.Node, frame *workerFrame) (workerValue, error) {
	expression := node.AsCallExpression()
	callee := ast.SkipParentheses(expression.Expression)
	args := []workerValue{}
	for _, argument := range nodesOf(expression.Arguments) {
		value, err := w.expression(argument, frame)
		if err != nil {
			return workerValue{}, err
		}
		args = append(args, value)
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		name := access.Name().Text()
		for _, library := range []string{"Math", "Number", "String"} {
			if w.task.l.isLibraryGlobal(access.Expression, library) && w.task.l.libraryMember(callee) {
				if library == "Math" && name == "random" {
					return workerValue{}, w.fail(node, "effectful Math.random")
				}
				if w.external(node).shared {
					return w.allocation(node, nil, "array"), nil
				}
				return workerValue{}, nil
			}
		}
		receiver, err := w.expression(access.Expression, frame)
		if err != nil {
			return workerValue{}, err
		}
		symbol := w.task.symbol(callee)
		library := symbol != nil && len(symbol.Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
		if library {
			return w.collectionCall(node, name, receiver, args, frame)
		}
		result := workerValue{}
		if receiver.shared {
			// Shared methods retain the independent deeply readonly source proof.
			if err := w.task.call(node, frame.owner, w.chain); err != nil {
				return workerValue{}, err
			}
			w.union(&result, w.external(node))
		}
		if !receiver.shared && len(receiver.objects) == 0 {
			return workerValue{}, w.fail(node, "method receiver has no private allocation")
		}
		for object := range receiver.objects {
			if object.class == nil {
				return workerValue{}, w.fail(node, "method dispatch target is not known")
			}
			var method *ast.Node
			for _, member := range object.class.AsClassDeclaration().Members.Nodes {
				if member.Kind == ast.KindMethodDeclaration && member.Name().Text() == name && !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
					if method != nil {
						return workerValue{}, w.fail(node, "multiple method targets")
					}
					method = member
				}
			}
			if method == nil || method.Body() == nil || !w.methodStable(method) {
				return workerValue{}, w.fail(node, "method dispatch targets are not all known and stable")
			}
			returned, err := w.function(method, args, workerValue{objects: map[*workerObject]bool{object: true}})
			if err != nil {
				return workerValue{}, err
			}
			w.union(&result, returned)
		}
		return result, nil
	}
	for _, name := range []string{"String", "Number", "Boolean"} {
		if w.task.l.isLibraryGlobal(callee, name) {
			for _, arg := range args {
				if arg.shared || len(arg.objects) > 0 {
					return workerValue{}, w.fail(node, "coercion may dispatch an object method")
				}
			}
			return workerValue{}, nil
		}
	}
	for _, name := range []string{"panic", "utf8At", "utf8Length"} {
		if w.task.l.isPreludeFunction(callee, name) {
			return workerValue{}, nil
		}
	}
	function := w.task.functionValue(callee, map[*ast.Symbol]bool{})
	if function == nil {
		return workerValue{}, w.fail(node, "call targets are not all known")
	}
	declaration := w.task.declaration(callee)
	if declaration != nil && !w.task.immutable(declaration) {
		return workerValue{}, w.fail(node, "callee binding can change")
	}
	w.task.global(callee)
	return w.function(function, args, workerValue{})
}

func (w *workerProof) collectionCall(node *ast.Node, name string, receiver workerValue, args []workerValue, frame *workerFrame) (workerValue, error) {
	mutates := false
	switch name {
	case "set", "add", "delete", "clear", "push", "pop", "shift", "unshift", "splice", "fill", "sort", "reverse", "copyWithin":
		mutates = true
	}
	if mutates && (receiver.shared || len(receiver.objects) == 0) {
		return workerValue{}, w.fail(node, "mutates a shared collection")
	}
	result := workerValue{}
	if mutates {
		for object := range receiver.objects {
			for _, arg := range args {
				w.merge(&object.elements, arg)
			}
		}
		if !w.external(node).shared {
			return result, nil
		}
		switch name {
		case "set", "add", "reverse", "sort", "fill", "copyWithin":
			return receiver, nil
		case "pop", "shift":
			for object := range receiver.objects {
				w.union(&result, object.elements)
			}
			return result, nil
		case "splice":
			value := w.allocation(node, nil, "array")
			for object := range value.objects {
				for original := range receiver.objects {
					w.merge(&object.elements, original.elements)
				}
			}
			return value, nil
		}
		return workerValue{}, w.fail(node, "mutating collection result is not proven")
	}
	// The list mirrors the shared task proof. Unknown library effects stay refused.
	switch name {
	case "get", "at":
		result.shared = receiver.shared
		for object := range receiver.objects {
			w.union(&result, object.elements)
		}
		if !w.external(node).shared {
			return workerValue{}, nil
		}
		return result, nil
	case "slice", "concat", "split", "keys", "values", "entries":
		if !w.external(node).shared {
			return workerValue{}, nil
		}
		value := w.allocation(node, nil, "array")
		for object := range value.objects {
			if receiver.shared && w.external(node).shared {
				w.merge(&object.elements, workerValue{shared: true})
			}
			for original := range receiver.objects {
				w.merge(&object.elements, original.elements)
			}
			for _, arg := range args {
				w.merge(&object.elements, arg)
			}
		}
		return value, nil
	case "has", "join", "includes", "indexOf", "lastIndexOf", "charCodeAt", "codePointAt", "charAt", "startsWith", "endsWith", "substring", "substr", "toLowerCase", "toUpperCase", "trim", "trimStart", "trimEnd", "replace", "replaceAll", "repeat", "toString", "toFixed", "toExponential", "toPrecision":
		return workerValue{}, nil
	}
	return workerValue{}, w.fail(node, "library call effect is not proven: "+name)
}
