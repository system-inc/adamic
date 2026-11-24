package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// enumComparedMember is whether the program's own comparison proves node, a read of a const local whose
// declared type is the whole open numeric enum identity, one of the members slot takes, where it's
// written (Adamic's ruling of Nov 19 on #snypn8c, c3). tsc's parser writes token() into a slot of a few
// members after a case chain or an === test (parser.ts:4754, 8417, 8701). Only a const local is taken,
// so nothing writes it between the comparison and the write. The checker's narrowed type at the read is
// never trusted: over an open enum it can't tell `!==` exclusion, which leaves any other number in, from
// a comparison, which names the member. So the proof is read off the program's own syntax: the
// comparisons that guard the read (comparedMembers), each member they name fitting the slot.
func (l *lowering) enumComparedMember(node *ast.Node, identity *ast.Symbol, slot *checker.Type) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return false
	}
	// A shorthand's name is the property to the checker; the proof belongs to the local it reads.
	symbol := l.valueSymbol(node)
	if symbol == nil || symbol.ValueDeclaration == nil {
		return false
	}
	if symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration || symbol.ValueDeclaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	members := l.comparedMembers(node, symbol, identity)
	if members == nil {
		return false
	}
	for _, member := range members {
		if !l.checker.IsTypeAssignableTo(member, slot) {
			return false
		}
	}
	return true
}

// comparedMembers is the members of identity the innermost comparison guarding node proves symbol holds
// there, or nil for none. The walk goes up from the read and takes the first proof it meets: an if's
// then-branch, a conditional's true arm or an && right operand whose test proves a set, or a case clause
// of a switch over symbol itself. A branch reached where its test failed (an else, a false arm, an &&
// left operand, an || operand) proves nothing, and the walk stops there. So does a function boundary: a
// function made in a branch runs when it's called, not where it's written, and this proof reads only the
// straight path from the test to the write. A class body is such a boundary too, since its field
// initializers run in its constructor. Proofs from several levels are never combined.
func (l *lowering) comparedMembers(node *ast.Node, symbol, identity *ast.Symbol) []*checker.Type {
	for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(parent) || ast.IsClassLike(parent) {
			return nil
		}
		switch parent.Kind {
		case ast.KindIfStatement:
			statement := parent.AsIfStatement()
			if child == statement.ElseStatement {
				return nil
			}
			if child == statement.ThenStatement {
				if members := l.provedMembers(statement.Expression, symbol, identity); members != nil {
					return members
				}
			}
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if child == conditional.WhenFalse {
				return nil
			}
			if child == conditional.WhenTrue {
				if members := l.provedMembers(conditional.Condition, symbol, identity); members != nil {
					return members
				}
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindBarBarToken:
				return nil
			case ast.KindAmpersandAmpersandToken:
				if child == binary.Left {
					return nil
				}
				if members := l.provedMembers(binary.Left, symbol, identity); members != nil {
					return members
				}
			}
		case ast.KindCaseClause, ast.KindDefaultClause:
			switchStatement := parent.Parent.Parent
			read := ast.SkipParentheses(switchStatement.AsSwitchStatement().Expression)
			if ast.IsIdentifier(read) && l.symbol(read) == symbol {
				return l.caseMembers(child, parent, identity)
			}
		}
	}
	return nil
}

// caseMembers is the members a switch over the value proves where child sits in clause: that clause's
// member and those of the clauses with no statements of their own right before it, which reach it only by
// falling through with nothing between. A default proves nothing over an open enum, which holds any
// number, and a clause before with statements may fall through with another value still in hand, so
// either one ends the proof with nil, as does a read in the case's own expression or a case that isn't
// one of identity's members.
func (l *lowering) caseMembers(child, clause *ast.Node, identity *ast.Symbol) []*checker.Type {
	if child == clause.AsCaseOrDefaultClause().Expression {
		return nil
	}
	clauses := clause.Parent.AsCaseBlock().Clauses.Nodes
	position := slices.Index(clauses, clause)
	var members []*checker.Type
	for index := position; index >= 0; index-- {
		if clauses[index].Kind == ast.KindDefaultClause {
			return nil
		}
		if index < position && len(clauses[index].AsCaseOrDefaultClause().Statements.Nodes) != 0 {
			return nil
		}
		member := l.identityMember(clauses[index].AsCaseOrDefaultClause().Expression, identity)
		if member == nil {
			return nil
		}
		members = append(members, member)
	}
	return members
}

// provedMembers is the set of identity's members a condition proves symbol holds where it's true, or nil
// where it proves nothing. `x === M` and `x == M`, either way round, prove M. `a || b` proves the union of
// both, and nothing when either side proves nothing. `a && b` proves what either side proves, both
// intersected where both do. Anything else proves nothing: `!==` and `!=` leave every other number in, and
// a call, a predicate and `!` are held for a later ruling.
func (l *lowering) provedMembers(condition *ast.Node, symbol, identity *ast.Symbol) []*checker.Type {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := condition.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken:
		if member := l.comparedMember(binary.Left, binary.Right, symbol, identity); member != nil {
			return []*checker.Type{member}
		}
		if member := l.comparedMember(binary.Right, binary.Left, symbol, identity); member != nil {
			return []*checker.Type{member}
		}
	case ast.KindBarBarToken:
		left := l.provedMembers(binary.Left, symbol, identity)
		right := l.provedMembers(binary.Right, symbol, identity)
		if left != nil && right != nil {
			return slices.Concat(left, right)
		}
	case ast.KindAmpersandAmpersandToken:
		left := l.provedMembers(binary.Left, symbol, identity)
		right := l.provedMembers(binary.Right, symbol, identity)
		if left == nil {
			return right
		}
		if right == nil {
			return left
		}
		// Both sides hold at once, so only a member both name can be there. None is a branch that
		// never runs, which writes nothing, so the empty set proves it.
		both := []*checker.Type{}
		for _, member := range left {
			if slices.Contains(right, member) {
				both = append(both, member)
			}
		}
		return both
	}
	return nil
}

// comparedMember is the member a comparison's member side names, when its other side reads symbol and the
// member is one of identity's.
func (l *lowering) comparedMember(read, member *ast.Node, symbol, identity *ast.Symbol) *checker.Type {
	read = ast.SkipParentheses(read)
	if !ast.IsIdentifier(read) || l.symbol(read) != symbol {
		return nil
	}
	return l.identityMember(member, identity)
}

// identityMember is the type of the member node reads, when it's one of identity's, as enumMemberOrigin
// asks it.
func (l *lowering) identityMember(node *ast.Node, identity *ast.Symbol) *checker.Type {
	if node == nil {
		return nil
	}
	declaration := l.enumMember(node)
	if declaration == nil || l.symbol(declaration.Parent.Name()) != identity {
		return nil
	}
	return l.checker.GetTypeAtLocation(ast.SkipParentheses(node))
}
