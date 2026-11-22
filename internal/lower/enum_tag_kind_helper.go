package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// #s8gy89r U2 applies #00rd54t only to a complete, pure helper body. Eight
// simultaneously active helpers are allowed; depth exhaustion or a back edge
// invalidates the whole outer predicate, never a partially collected set.
const kindHelperMaxDepth = 8

func (r *kindSetBody) kindHelperCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindIdentifier || call.QuestionDotToken != nil || node.Flags&ast.NodeFlagsOptionalChain != 0 || len(call.Arguments.Nodes) != 1 || (call.TypeArguments != nil && len(call.TypeArguments.Nodes) != 0) {
		return r.probeReject("kindHelperCall:callee-or-arity", node)
	}
	// The ruling names p.f, or the helper parameter itself, not a saved alias.
	if r.alias != nil {
		return r.probeReject("kindHelperCall:saved-alias", node)
	}
	if !r.read(call.Arguments.Nodes[0]) {
		return r.probeReject("kindHelperCall:argument-not-field", node)
	}
	symbol := r.l.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return r.probeReject("kindHelperCall:single-declaration", node)
	}
	function := symbol.Declarations[0]
	if function.Kind == ast.KindVariableDeclaration {
		v := function.AsVariableDeclaration()
		if function.Parent == nil || function.Parent.Flags&ast.NodeFlagsConst == 0 || v.Type != nil || v.Initializer == nil {
			return r.probeReject("kindHelperCall:const-arrow-binding", node)
		}
		function = ast.SkipParentheses(v.Initializer)
		if function.Kind != ast.KindArrowFunction {
			return r.probeReject("kindHelperCall:not-arrow", node)
		}
	} else if function.Kind != ast.KindFunctionDeclaration {
		return r.probeReject("kindHelperCall:not-function", node)
	}
	if function.Body() == nil || generatorDeclaration(function) || ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) || len(function.TypeParameters()) != 0 || r.kindHelperReassigned(symbol) {
		return r.probeReject("kindHelperCall:impure-generic-generator-or-reassigned", node)
	}
	parameters := function.Parameters()
	if len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) && parameters[0].Name().Text() == "this" {
		parameters = parameters[1:]
	}
	if len(parameters) != 1 || !ast.IsIdentifier(parameters[0].Name()) {
		return r.probeReject("kindHelperCall:one-plain-parameter", node)
	}
	p := parameters[0].AsParameterDeclaration()
	if p.Initializer != nil || p.QuestionToken != nil || p.DotDotDotToken != nil {
		return r.probeReject("kindHelperCall:default-optional-rest", node)
	}
	if r.depth >= kindHelperMaxDepth {
		return r.probeReject("kindHelperCall:depth-limit", node)
	}
	if r.active == nil {
		r.active = map[*ast.Node]bool{}
	}
	if r.active[function] {
		return r.probeReject("kindHelperCall:cycle", node)
	}
	r.active[function] = true
	defer delete(r.active, function)
	child := &kindSetBody{l: r.l, parameter: r.l.symbol(parameters[0].Name()), enum: r.enum, scalar: true, depth: r.depth + 1, active: r.active}
	// Only plain parameter reads exist in this grammar. A write, including one
	// hidden among statements before a return, makes the complete body fail.
	if !child.kindHelperBody(function.Body()) {
		r.probeWhy = child.probeWhy
		r.probeAt = child.probeAt
		return r.probeResult("kindHelperCall:child-body:"+r.probeWhy, node, false)
	}
	r.enum = child.enum
	for _, kind := range child.members {
		r.add(kind)
	}
	return true
}

// A function declaration can be rebound even though a const arrow cannot. Read
// value symbols at assignment targets, including shorthand/destructuring writes.
func (r *kindSetBody) kindHelperReassigned(symbol *ast.Symbol) bool {
	changed := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if ast.IsIdentifier(n) && ast.IsAssignmentTarget(n) && r.l.valueSymbol(n) == symbol {
			changed = true
		}
		if changed {
			return true
		}
		return n.ForEachChild(visit)
	}
	for _, file := range r.l.program.SourceFiles() {
		if visit(file.AsNode()) {
			break
		}
	}
	return changed
}

func (r *kindSetBody) kindHelperBody(body *ast.Node) bool {
	if body.Kind != ast.KindBlock {
		return r.comparisons(body)
	}
	statements := body.AsBlock().Statements.Nodes
	if len(statements) == 1 && kindSetFalse(statements[0]) {
		return true
	}
	if len(statements) > 0 && statements[0].Kind == ast.KindSwitchStatement {
		var tail []*checker.Type
		if len(statements) == 2 {
			value := statements[1]
			if !kindSetFalse(value) {
				if value.Kind != ast.KindReturnStatement || value.AsReturnStatement().Expression == nil {
					return r.probeReject("kindHelperBody:tail-not-return", body)
				}
				expression := ast.SkipParentheses(value.AsReturnStatement().Expression)
				if expression.Kind != ast.KindCallExpression {
					return r.probeReject("kindHelperBody:tail-not-call", body)
				}
				fork := *r
				fork.members = nil
				if !fork.kindHelperCall(expression) {
					r.probeWhy = fork.probeWhy
					r.probeAt = fork.probeAt
					return r.probeReject("kindHelperBody:tail-helper:"+r.probeWhy, body)
				}
				r.enum = fork.enum
				tail = fork.members
			}
		} else if len(statements) != 1 {
			return r.probeReject("kindHelperBody:statement-count", body)
		}
		return r.kindHelperSwitch(statements[0], len(statements) == 2, tail)
	}
	return r.body(body)
}

// A default helper is evaluated only after every explicit label failed. Its
// true set therefore excludes all labels, including explicit false cases.
func (r *kindSetBody) kindHelperSwitch(node *ast.Node, tail bool, tailSet []*checker.Type) bool {
	s := node.AsSwitchStatement()
	if !r.read(s.Expression) {
		return r.probeReject("kindHelperSwitch:not-field", node)
	}
	clauses := s.CaseBlock.AsCaseBlock().Clauses.Nodes
	var seen []*checker.Type
	fallback := tailSet
	outcome := false
	dynamic := tailSet
	falls := !tail
	hasDefault := false
	for i := len(clauses) - 1; i >= 0; i-- {
		c := clauses[i]
		var kind *checker.Type
		if c.Kind == ast.KindCaseClause {
			kind = r.member(c.AsCaseOrDefaultClause().Expression)
			if kind == nil {
				return r.probeReject("kindHelperSwitch:case-not-member", node)
			}
			for _, held := range seen {
				if checker.Checker_isTypeIdenticalTo(r.l.checker, held, kind) {
					return r.probeReject("kindHelperSwitch:duplicate-case", node)
				}
			}
			seen = append(seen, kind)
		} else if c.Kind == ast.KindDefaultClause {
			if hasDefault {
				return r.probeReject("kindHelperSwitch:duplicate-default", node)
			}
			hasDefault = true
		} else {
			return r.probeReject("kindHelperSwitch:unknown-clause", node)
		}
		statements := c.AsCaseOrDefaultClause().Statements.Nodes
		if len(statements) > 0 {
			if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement || statements[0].AsReturnStatement().Expression == nil {
				return r.probeReject("kindHelperSwitch:clause-not-return", node)
			}
			expression := ast.SkipParentheses(statements[0].AsReturnStatement().Expression)
			dynamic = nil
			if expression.Kind == ast.KindTrueKeyword || expression.Kind == ast.KindFalseKeyword {
				outcome = expression.Kind == ast.KindTrueKeyword
				if c.Kind == ast.KindDefaultClause {
					if outcome {
						return r.probeReject("kindHelperSwitch:true-default", node)
					}
					fallback = nil
				}
			} else if c.Kind == ast.KindDefaultClause && expression.Kind == ast.KindCallExpression {
				fork := *r
				fork.members = nil
				if !fork.kindHelperCall(expression) {
					r.probeWhy = fork.probeWhy
					r.probeAt = fork.probeAt
					return r.probeReject("kindHelperSwitch:default-helper:"+r.probeWhy, node)
				}
				r.enum = fork.enum
				dynamic = fork.members
				fallback = dynamic
				outcome = false
			} else {
				return r.probeReject("kindHelperSwitch:unrecognized-return", node)
			}
			falls = false
		}
		if c.Kind == ast.KindDefaultClause {
			fallback = dynamic // An empty default inherits its successor, not the tail.
		}
		if falls || c.Kind == ast.KindDefaultClause && outcome {
			return r.probeReject("kindHelperSwitch:fallthrough-or-true-default", node)
		}
		if kind != nil && (outcome || r.l.kindSetInside(kind, dynamic)) {
			r.add(kind)
		}
	}
	if !hasDefault && !tail {
		return r.probeReject("kindHelperSwitch:no-false-default-or-tail", node)
	}
	for _, kind := range fallback {
		excluded := false
		for _, label := range seen {
			if checker.Checker_isTypeIdenticalTo(r.l.checker, label, kind) {
				excluded = true
				break
			}
		}
		if !excluded {
			r.add(kind)
		}
	}
	return true
}
