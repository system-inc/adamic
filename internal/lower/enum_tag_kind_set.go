package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// #s8gy89r U1 applies #00rd54t: a predicate's call is only the kind comparisons
// its complete body makes. This parser has no partial answer: an unrecognized
// statement, operand or disjunct invalidates the whole body. The flow walk still
// judges hazards, the checker's narrowing and accessor names at the call.
type kindSetBody struct {
	l                *lowering
	parameter, alias *ast.Symbol
	field            string
	enum             *ast.Symbol
	members          []*checker.Type
}

func (l *lowering) kindPredicateBody(body, parameter *ast.Node) (string, []*checker.Type, bool) {
	r := &kindSetBody{l: l, parameter: l.checker.GetSymbolAtLocation(parameter)}
	if body == nil || r.parameter == nil {
		return "", nil, false
	}
	if !r.body(body) || r.field == "" || len(r.members) == 0 {
		return "", nil, false
	}
	return r.field, r.members, true
}

func (r *kindSetBody) parameterRead(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	return n.Kind == ast.KindIdentifier && r.l.checker.GetSymbolAtLocation(n) == r.parameter
}
func (r *kindSetBody) read(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if r.alias != nil {
		return n.Kind == ast.KindIdentifier && r.l.checker.GetSymbolAtLocation(n) == r.alias
	}
	if n.Kind != ast.KindPropertyAccessExpression || n.Flags&ast.NodeFlagsOptionalChain != 0 {
		return false
	}
	a := n.AsPropertyAccessExpression()
	if a.QuestionDotToken != nil || !r.parameterRead(a.Expression) {
		return false
	}
	name := a.Name().Text()
	if r.field != "" && r.field != name {
		return false
	}
	r.field = name
	return true
}
func (r *kindSetBody) member(n *ast.Node) *checker.Type {
	n = ast.SkipParentheses(n)
	symbol := r.l.symbol(n)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsEnumMember == 0 {
		return nil
	}
	t := r.l.checker.GetTypeAtLocation(n)
	identity := r.l.enumIdentity(t)
	if identity == nil || r.enum != nil && r.enum != identity {
		return nil
	}
	r.enum = identity
	return t
}
func (r *kindSetBody) add(t *checker.Type) {
	for _, held := range r.members {
		if checker.Checker_isTypeIdenticalTo(r.l.checker, held, t) {
			return
		}
	}
	r.members = append(r.members, t)
}

// Only object-or-nullish parameters make p and !!p nullish guards. A primitive
// truthiness test could exclude 0, false or an empty string as well.
func (r *kindSetBody) guard(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if n.Kind == ast.KindPrefixUnaryExpression && n.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		n = ast.SkipParentheses(n.AsPrefixUnaryExpression().Operand)
		if n.Kind != ast.KindPrefixUnaryExpression || n.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
			return false
		}
		n = ast.SkipParentheses(n.AsPrefixUnaryExpression().Operand)
	}
	if !r.parameterRead(n) {
		return false
	}
	t := r.l.checker.GetTypeOfSymbol(r.parameter)
	parts := []*checker.Type{t}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		parts = t.Types()
	}
	for _, part := range parts {
		if part.Flags()&checker.TypeFlagsPrimitive != 0 && part.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			return false
		}
	}
	return true
}
func (r *kindSetBody) comparisons(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if n.Kind != ast.KindBinaryExpression {
		return false
	}
	b := n.AsBinaryExpression()
	if b.OperatorToken.Kind == ast.KindBarBarToken {
		return r.comparisons(b.Left) && r.comparisons(b.Right)
	}
	if b.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
		return false
	}
	// Probe the member before the field, so reversing the operands cannot leave
	// behind a field name from an unsuccessful interpretation.
	if t := r.member(b.Right); t != nil && r.read(b.Left) {
		r.add(t)
		return true
	}
	if t := r.member(b.Left); t != nil && r.read(b.Right) {
		r.add(t)
		return true
	}
	return false
}
func (r *kindSetBody) expression(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if n.Kind == ast.KindBinaryExpression && n.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
		b := n.AsBinaryExpression()
		return r.guard(b.Left) && r.comparisons(b.Right)
	}
	return r.comparisons(n)
}
func kindSetFalse(n *ast.Node) bool {
	return n.Kind == ast.KindReturnStatement && n.AsReturnStatement().Expression != nil && ast.SkipParentheses(n.AsReturnStatement().Expression).Kind == ast.KindFalseKeyword
}
func (r *kindSetBody) body(n *ast.Node) bool {
	if n.Kind != ast.KindBlock {
		return r.expression(n)
	}
	statements := n.AsBlock().Statements.Nodes
	if len(statements) == 0 {
		return false
	}
	// The alias is local to this grammar: exactly one const and no other write or
	// statement can appear before its use. A let or an assignment is never read.
	if statements[0].Kind == ast.KindVariableStatement {
		list := statements[0].AsVariableStatement().DeclarationList
		declarations := list.AsVariableDeclarationList().Declarations.Nodes
		if list.Flags&ast.NodeFlagsConst == 0 || len(declarations) != 1 {
			return false
		}
		d := declarations[0]
		if d.Name().Kind != ast.KindIdentifier || d.Initializer() == nil || !r.read(d.Initializer()) {
			return false
		}
		r.alias = r.l.checker.GetSymbolAtLocation(d.Name())
		if r.alias == nil {
			return false
		}
		statements = statements[1:]
		if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
			return false
		}
	}
	if len(statements) == 1 && statements[0].Kind == ast.KindReturnStatement {
		value := statements[0].AsReturnStatement().Expression
		return value != nil && r.expression(value)
	}
	if len(statements) == 2 && statements[0].Kind == ast.KindIfStatement && kindSetFalse(statements[1]) {
		test := statements[0].AsIfStatement()
		return test.ElseStatement == nil && test.ThenStatement.Kind == ast.KindBlock && r.guard(test.Expression) && r.body(test.ThenStatement)
	}
	if len(statements) == 1 || len(statements) == 2 && kindSetFalse(statements[1]) {
		if statements[0].Kind == ast.KindSwitchStatement {
			return r.switchBody(statements[0], len(statements) == 2)
		}
	}
	return false
}
func (r *kindSetBody) switchBody(n *ast.Node, tailFalse bool) bool {
	s := n.AsSwitchStatement()
	if !r.read(s.Expression) {
		return false
	}
	clauses := s.CaseBlock.AsCaseBlock().Clauses.Nodes
	outcome := false
	falls := !tailFalse
	hasDefault := false
	// Empty cases inherit the next clause's literal return. No break, call or
	// other expression is ignored, even in a case that returns false.
	for i := len(clauses) - 1; i >= 0; i-- {
		c := clauses[i]
		var statements []*ast.Node
		var value *checker.Type
		if c.Kind == ast.KindCaseClause {
			value = r.member(c.AsCaseOrDefaultClause().Expression)
			if value == nil {
				return false
			}
			statements = c.AsCaseOrDefaultClause().Statements.Nodes
		} else if c.Kind == ast.KindDefaultClause {
			if hasDefault {
				return false
			}
			hasDefault = true
			statements = c.AsCaseOrDefaultClause().Statements.Nodes
		} else {
			return false
		}
		if len(statements) != 0 {
			if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
				return false
			}
			expression := statements[0].AsReturnStatement().Expression
			if expression == nil {
				return false
			}
			expression = ast.SkipParentheses(expression)
			if expression.Kind != ast.KindTrueKeyword && expression.Kind != ast.KindFalseKeyword {
				return false
			}
			outcome = expression.Kind == ast.KindTrueKeyword
			falls = false
		}
		if falls || c.Kind == ast.KindDefaultClause && outcome {
			return false
		}
		if value != nil && outcome {
			r.add(value)
		}
	}
	return hasDefault || tailFalse
}

func (l *lowering) kindSetMeets(tag *checker.Type, set []*checker.Type) bool {
	for _, kind := range set {
		if checker.Checker_isTypeAssignableTo(l.checker, kind, tag) {
			return true
		}
	}
	return false
}
func (l *lowering) kindSetInside(tag *checker.Type, set []*checker.Type) bool {
	// Preserve the singleton rule exactly, including never (an empty K(S)).
	parts := []*checker.Type{tag}
	if tag.Flags()&checker.TypeFlagsUnion != 0 {
		parts = tag.Types()
	}
	for _, part := range parts {
		found := false
		for _, kind := range set {
			if checker.Checker_isTypeIdenticalTo(l.checker, part, kind) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
