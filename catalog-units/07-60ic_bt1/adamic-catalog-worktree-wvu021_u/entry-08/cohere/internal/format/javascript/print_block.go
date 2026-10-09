package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// print/block.js and print/statement-sequence.js.

// printBlock is upstream's printBlock: Program, BlockStatement, StaticBlock and TSModuleBlock.
func printBlock(path *Path, options *Options, print PrintFunc) Doc {
	bodyDoc := printBlockBody(path, options, print)
	current := node(path)
	parent := parentOf(path)

	if current.Is("Program") && !parent.Is("ModuleExpression") {
		if bodyDoc != nil {
			return concatIn(path, bodyDoc, hardline)
		}
		return emptyDoc
	}

	var parts []any
	if current.Is("StaticBlock") {
		parts = append(parts, "static ")
	}
	parts = append(parts, "{")
	if bodyDoc != nil {
		parts = append(parts, indentIn(path, concatIn(path, hardline, bodyDoc)), hardline)
	} else {
		parentParent := grandparentOf(path)
		if !(parent.Is("ArrowFunctionExpression", "FunctionExpression", "FunctionDeclaration", "ComponentDeclaration",
			"HookDeclaration", "ObjectMethod", "ClassMethod", "ClassPrivateMethod", "ForStatement", "WhileStatement",
			"DoWhileStatement", "DoExpression", "ModuleExpression") ||
			parent.Is("CatchClause") && parentParent.Child("finalizer") == nil ||
			parent.Is("TSModuleDeclaration", "DeclareModule", "MatchStatementCase") ||
			current.Is("StaticBlock")) {
			parts = append(parts, hardline)
		}
	}
	parts = append(parts, "}")
	return concatIn(path, parts...)
}

// printBlockBody is upstream's printBlockBody. It returns nil where upstream returns "", the falsy
// value printBlock tests for. typescript-estree has no directives array (directives are expression
// statements in body), so that branch is the Babel one and is not reached.
func printBlockBody(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	hasDirectives := len(current.List("directives")) > 0
	hasBody := false
	for _, statement := range current.List("body") {
		if !statement.Is("EmptyStatement") {
			hasBody = true
			break
		}
	}
	hasDanglingComments := hasComment(current, commentDangling, nil)

	if !hasDirectives && !hasBody && !hasDanglingComments {
		return nil
	}

	var parts []any
	if hasDirectives {
		parts = append(parts, printStatementSequence(path, options, print, "directives"))
		if hasBody || hasDanglingComments {
			parts = append(parts, hardline)
			directives := current.List("directives")
			if isNextLineEmptyAfter(directives[len(directives)-1], options) {
				parts = append(parts, hardline)
			}
		}
	}
	if hasBody {
		parts = append(parts, printStatementSequence(path, options, print, "body"))
	}
	if hasDanglingComments {
		parts = append(parts, printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}))
	}
	return concatIn(path, parts...)
}

// printStatementSequence is upstream's printStatementSequence.
func printStatementSequence(path *Path, options *Options, print PrintFunc, property string) Doc {
	current := node(path)
	var parts []any

	var lastStatement Node
	statements := current.List(property)
	for index := len(statements) - 1; index >= 0; index-- {
		if !statements[index].Is("EmptyStatement") {
			lastStatement = statements[index]
			break
		}
	}

	each(path, func(path *Path, _ int) {
		statement := node(path)
		if statement.Is("EmptyStatement") {
			return
		}
		parts = append(parts, print(nil, nil))
		if statement != lastStatement {
			parts = append(parts, hardline)
			if isNextLineEmptyAfter(statement, options) {
				parts = append(parts, hardline)
			}
		}
	}, property)

	return concatIn(path, parts...)
}
