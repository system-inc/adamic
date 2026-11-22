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
	scalar           bool
	depth            int
	active           map[*ast.Node]bool
	probeWhy         string
	probeAt          *ast.Node
}

func (l *lowering) kindPredicateBody(body, parameter *ast.Node) (string, []*checker.Type, bool) {
	return l.kindPredicateBodyProbe(body, parameter, nil)
}

func (l *lowering) kindPredicateBodyProbe(body, parameter *ast.Node, walk *enumTagWalk) (string, []*checker.Type, bool) {
	r := &kindSetBody{l: l, parameter: l.checker.GetSymbolAtLocation(parameter)}
	if body == nil || r.parameter == nil {
		if walk != nil {
			walk.probeMiss("kindPredicateBody:missing-body-or-parameter", body)
		}
		return "", nil, false
	}
	if !r.body(body) || r.field == "" || len(r.members) == 0 {
		if walk != nil {
			why := r.probeWhy
			if why == "" {
				why = "kindPredicateBody:no-field-or-members"
			}
			walk.probeMiss(why, r.probeAt)
		}
		return "", nil, false
	}
	return r.field, r.members, true
}

func (r *kindSetBody) parameterRead(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	return r.probeResult("parameterRead:not-parameter", n, n.Kind == ast.KindIdentifier && r.l.checker.GetSymbolAtLocation(n) == r.parameter)
}
func (r *kindSetBody) read(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if r.alias != nil {
		return r.probeResult("read:not-alias", n, n.Kind == ast.KindIdentifier && r.l.checker.GetSymbolAtLocation(n) == r.alias)
	}
	if r.scalar {
		return r.parameterRead(n)
	}
	if n.Kind != ast.KindPropertyAccessExpression || n.Flags&ast.NodeFlagsOptionalChain != 0 {
		return r.probeReject("read:not-direct-field", n)
	}
	a := n.AsPropertyAccessExpression()
	if a.QuestionDotToken != nil || !r.parameterRead(a.Expression) {
		return r.probeReject("read:optional-or-other-parameter", n)
	}
	name := a.Name().Text()
	if r.field != "" && r.field != name {
		return r.probeReject("read:second-field", n)
	}
	r.field = name
	return true
}
func (r *kindSetBody) member(n *ast.Node) *checker.Type {
	n = ast.SkipParentheses(n)
	symbol := r.l.symbol(n)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsEnumMember == 0 {
		return r.probeMember("member:not-enum-member", n)
	}
	t := r.l.checker.GetTypeAtLocation(n)
	identity := r.l.enumIdentity(t)
	if identity == nil || r.enum != nil && r.enum != identity {
		return r.probeMember("member:different-or-missing-enum", n)
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
			return r.probeReject("guard:single-negation", n)
		}
		n = ast.SkipParentheses(n.AsPrefixUnaryExpression().Operand)
	}
	if !r.parameterRead(n) {
		return r.probeReject("guard:not-parameter", n)
	}
	t := r.l.checker.GetTypeOfSymbol(r.parameter)
	parts := []*checker.Type{t}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		parts = t.Types()
	}
	for _, part := range parts {
		if part.Flags()&checker.TypeFlagsPrimitive != 0 && part.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			return r.probeReject("guard:primitive-parameter", n)
		}
	}
	return true
}
func (r *kindSetBody) comparisons(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if n.Kind == ast.KindCallExpression {
		return r.kindHelperCall(n)
	}
	if n.Kind != ast.KindBinaryExpression {
		return r.probeReject("comparisons:not-comparison", n)
	}
	b := n.AsBinaryExpression()
	if b.OperatorToken.Kind == ast.KindBarBarToken {
		return r.comparisons(b.Left) && r.comparisons(b.Right)
	}
	if b.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
		return r.probeReject("comparisons:not-strict-equality", n)
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
	return r.probeReject("comparisons:unrecognized-operands", n)
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
		return r.probeReject("body:empty", n)
	}
	// The alias is local to this grammar: exactly one const and no other write or
	// statement can appear before its use. A let or an assignment is never read.
	if statements[0].Kind == ast.KindVariableStatement {
		list := statements[0].AsVariableStatement().DeclarationList
		declarations := list.AsVariableDeclarationList().Declarations.Nodes
		if list.Flags&ast.NodeFlagsConst == 0 || len(declarations) != 1 {
			return r.probeReject("body:alias-not-single-const", n)
		}
		d := declarations[0]
		if d.Name().Kind != ast.KindIdentifier || d.Initializer() == nil || !r.read(d.Initializer()) {
			return r.probeReject("body:alias-not-field", n)
		}
		r.alias = r.l.checker.GetSymbolAtLocation(d.Name())
		if r.alias == nil {
			return r.probeReject("body:alias-no-symbol", n)
		}
		statements = statements[1:]
		if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
			return r.probeReject("body:alias-not-return", n)
		}
	}
	if len(statements) == 1 && statements[0].Kind == ast.KindReturnStatement {
		value := statements[0].AsReturnStatement().Expression
		return r.probeResult("body:missing-or-unrecognized-return", n, value != nil && r.expression(value))
	}
	if len(statements) == 2 && statements[0].Kind == ast.KindIfStatement && kindSetFalse(statements[1]) {
		test := statements[0].AsIfStatement()
		return r.probeResult("body:unrecognized-if-guard-or-body", n, test.ElseStatement == nil && test.ThenStatement.Kind == ast.KindBlock && r.guard(test.Expression) && r.body(test.ThenStatement))
	}
	if len(statements) == 1 || len(statements) == 2 && kindSetFalse(statements[1]) {
		if statements[0].Kind == ast.KindSwitchStatement {
			return r.switchBody(statements[0], len(statements) == 2)
		}
	}
	return r.probeReject("body:unrecognized-statements", n)
}
func (r *kindSetBody) switchBody(n *ast.Node, tailFalse bool) bool {
	s := n.AsSwitchStatement()
	if !r.read(s.Expression) {
		return r.probeReject("switchBody:not-field", n)
	}
	clauses := s.CaseBlock.AsCaseBlock().Clauses.Nodes
	outcome := false
	falls := !tailFalse
	hasDefault := false
	var seen []*checker.Type
	// Empty cases inherit the next clause's literal return. No break, call or
	// other expression is ignored, even in a case that returns false.
	for i := len(clauses) - 1; i >= 0; i-- {
		c := clauses[i]
		var statements []*ast.Node
		var value *checker.Type
		if c.Kind == ast.KindCaseClause {
			value = r.member(c.AsCaseOrDefaultClause().Expression)
			if value == nil {
				return r.probeReject("switchBody:case-not-member", n)
			}
			// Duplicate labels are legal TypeScript, but the first wins at runtime.
			// Reading backwards must never let a later true case replace it.
			for _, earlier := range seen {
				if checker.Checker_isTypeIdenticalTo(r.l.checker, earlier, value) {
					return r.probeReject("switchBody:duplicate-case", n)
				}
			}
			seen = append(seen, value)
			statements = c.AsCaseOrDefaultClause().Statements.Nodes
		} else if c.Kind == ast.KindDefaultClause {
			if hasDefault {
				return r.probeReject("switchBody:duplicate-default", n)
			}
			hasDefault = true
			statements = c.AsCaseOrDefaultClause().Statements.Nodes
		} else {
			return r.probeReject("switchBody:unknown-clause", n)
		}
		if len(statements) != 0 {
			if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
				return r.probeReject("switchBody:non-return-clause", n)
			}
			expression := statements[0].AsReturnStatement().Expression
			if expression == nil {
				return r.probeReject("switchBody:empty-return", n)
			}
			expression = ast.SkipParentheses(expression)
			if expression.Kind != ast.KindTrueKeyword && expression.Kind != ast.KindFalseKeyword {
				return r.probeReject("switchBody:nonliteral-return", n)
			}
			outcome = expression.Kind == ast.KindTrueKeyword
			falls = false
		}
		if falls || c.Kind == ast.KindDefaultClause && outcome {
			return r.probeReject("switchBody:fallthrough-or-true-default", n)
		}
		if value != nil && outcome {
			r.add(value)
		}
	}
	return r.probeResult("switchBody:no-false-default-or-tail", n, hasDefault || tailFalse)
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
