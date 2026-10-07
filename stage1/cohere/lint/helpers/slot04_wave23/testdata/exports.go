package control_flow_graph

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

type AdamicElement struct {
	Valid, Assignment, Shorthand, Spread, Computed bool
	Key, Target, Fallback                          int
}
type AdamicNode struct {
	Present, Identifier, Throwable, Parenthesized, Binary, Conditional, Prefix, Postfix, Update, Access, YieldNode, ClassExpression, Root, Hole, Wrapper, Binding, Object, Array, Spread, Equals, SourceFile, StaticBlock, Property, BodyBlock bool
	Operand, Left, Right, Type, Body, Initializer                                                                                                                                                                                              int
	Children, Statements, Parameters                                                                                                                                                                                                           []int
	Elements                                                                                                                                                                                                                                   []AdamicElement
}

func AdamicView(n *ast.Node, ids map[*ast.Node]int) AdamicNode {
	id := func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		x, ok := ids[n]
		if !ok {
			panic("missing child identity")
		}
		return x
	}
	v := AdamicNode{Present: n != nil, Operand: -1, Left: -1, Right: -1, Type: -1, Body: -1, Initializer: -1, Children: []int{}, Statements: []int{}, Parameters: []int{}, Elements: []AdamicElement{}}
	if n == nil {
		return v
	}
	v.Root = IsRoot(n)
	n.ForEachChild(func(child *ast.Node) bool { v.Children = append(v.Children, id(child)); return false })
	switch n.Kind {
	case ast.KindIdentifier:
		v.Identifier = true
		v.Throwable = isThrowableIdentifier(n)
	case ast.KindParenthesizedExpression:
		v.Parenthesized = true
		v.Wrapper = true
		v.Operand = id(n.AsParenthesizedExpression().Expression)
	case ast.KindBinaryExpression:
		v.Binary = true
		b := n.AsBinaryExpression()
		v.Left = id(b.Left)
		v.Right = id(b.Right)
		v.Equals = b.OperatorToken.Kind == ast.KindEqualsToken
	case ast.KindConditionalExpression:
		v.Conditional = true
	case ast.KindPrefixUnaryExpression:
		v.Prefix = true
		u := n.AsPrefixUnaryExpression()
		v.Operand = id(u.Operand)
		v.Update = u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken
	case ast.KindPostfixUnaryExpression:
		v.Postfix = true
		v.Operand = id(n.AsPostfixUnaryExpression().Operand)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		v.Access = true
	case ast.KindYieldExpression:
		v.YieldNode = true
		v.Operand = id(n.AsYieldExpression().Expression)
	case ast.KindClassExpression:
		v.ClassExpression = true
	case ast.KindOmittedExpression:
		v.Hole = true
	case ast.KindNonNullExpression:
		v.Wrapper = true
		v.Operand = id(n.AsNonNullExpression().Expression)
	case ast.KindAsExpression:
		v.Wrapper = true
		v.Operand = id(n.AsAsExpression().Expression)
	case ast.KindSatisfiesExpression:
		v.Wrapper = true
		v.Operand = id(n.AsSatisfiesExpression().Expression)
	case ast.KindTypeAssertionExpression:
		v.Wrapper = true
		v.Operand = id(n.AsTypeAssertion().Expression)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		v.Binding = true
		p := n.AsBindingPattern()
		if p.Elements != nil {
			for _, element := range p.Elements.Nodes {
				e := AdamicElement{Key: -1, Target: -1, Fallback: -1}
				b := element.AsBindingElement()
				if b != nil {
					e.Valid = true
					e.Target = id(b.Name())
					e.Fallback = id(b.Initializer)
					if b.PropertyName != nil && b.PropertyName.Kind == ast.KindComputedPropertyName {
						e.Computed = true
						e.Key = id(b.PropertyName.AsComputedPropertyName().Expression)
					}
				}
				v.Elements = append(v.Elements, e)
			}
		}
	case ast.KindObjectLiteralExpression:
		v.Object = true
		for _, property := range n.AsObjectLiteralExpression().Properties.Nodes {
			e := AdamicElement{Key: -1, Target: -1, Fallback: -1}
			switch property.Kind {
			case ast.KindPropertyAssignment:
				e.Assignment = true
				a := property.AsPropertyAssignment()
				e.Target = id(a.Initializer)
				if name := a.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					e.Computed = true
					e.Key = id(name.AsComputedPropertyName().Expression)
				}
			case ast.KindShorthandPropertyAssignment:
				e.Shorthand = true
				s := property.AsShorthandPropertyAssignment()
				e.Target = id(s.Name())
				e.Fallback = id(s.ObjectAssignmentInitializer)
			case ast.KindSpreadAssignment:
				e.Spread = true
				e.Target = id(property.AsSpreadAssignment().Expression)
			}
			v.Elements = append(v.Elements, e)
		}
	case ast.KindArrayLiteralExpression:
		v.Array = true
		v.Children = []int{}
		for _, element := range n.AsArrayLiteralExpression().Elements.Nodes {
			v.Children = append(v.Children, id(element))
		}
	case ast.KindSpreadElement:
		v.Spread = true
		v.Operand = id(n.AsSpreadElement().Expression)
	}
	if v.Root && n.Kind != ast.KindSourceFile && n.Kind != ast.KindClassStaticBlockDeclaration && n.Kind != ast.KindPropertyDeclaration {
		for _, parameter := range n.Parameters() {
			v.Parameters = append(v.Parameters, id(parameter))
		}
		v.Type = id(n.Type())
		v.Body = id(n.Body())
		if n.Body() != nil {
			v.BodyBlock = n.Body().Kind == ast.KindBlock
		}
	}
	switch n.Kind {
	case ast.KindSourceFile:
		v.SourceFile = true
		for _, statement := range n.AsSourceFile().Statements.Nodes {
			v.Statements = append(v.Statements, id(statement))
		}
	case ast.KindClassStaticBlockDeclaration:
		v.StaticBlock = true
		v.Body = id(n.AsClassStaticBlockDeclaration().Body)
	case ast.KindPropertyDeclaration:
		v.Property = true
		v.Initializer = id(n.AsPropertyDeclaration().Initializer)
	}
	return v
}

var adamicMode string
var adamicAbrupt bool
var adamicIDs map[*ast.Node]int
var adamicActions []string

func adamicID(n *ast.Node) int {
	if n == nil {
		return -1
	}
	return adamicIDs[n]
}
func adamicAction(name string, n *ast.Node) {
	adamicActions = append(adamicActions, fmt.Sprintf("%s:%d", name, adamicID(n)))
}
func adamicDependency[E any](b *Builder[E], name string, n *ast.Node) bool {
	adamicAction(name, n)
	if adamicMode == "build" && adamicAbrupt {
		b.cur.Reachable = false
	}
	return true
}
func adamicList[E any](b *Builder[E], list *ast.NodeList) bool {
	ids := []string{}
	if list != nil {
		for _, n := range list.Nodes {
			ids = append(ids, fmt.Sprint(adamicID(n)))
		}
	}
	adamicActions = append(adamicActions, "statements:"+strings.Join(ids, ","))
	if adamicAbrupt {
		b.cur.Reachable = false
	}
	return true
}
func AdamicObserve(nodes []*ast.Node, ids map[*ast.Node]int) {
	adamicIDs = ids
	for hook := 0; hook < 2; hook++ {
		for _, abrupt := range []bool{false, true} {
			adamicAbrupt = abrupt
			for index, n := range nodes {
				if !IsRoot(n) {
					continue
				}
				adamicMode = "build"
				adamicActions = []string{}
				g := Build[int](n, Hooks[int]{})
				fmt.Printf("build|%d|%d|%t|%s|%t|%d|%d|%d|%t|%t\n", index, hook, abrupt, strings.Join(adamicActions, ";"), g.EndReachable, len(g.Blocks), len(g.FinalBlocks), len(g.ThrownBlocks), g.Blocks[0].hasIncoming, g.Blocks[0].Reachable)
			}
		}
	}
}
