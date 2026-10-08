package consistentreturn

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"os"
	"strconv"
	"strings"
	"sync"
)

type captureNode struct {
	Kind, Text                                string
	Expression, Name, Parent, Arrow, Pos, End int
	Generator, StaticModifier, AsyncModifier  bool
}
type captureRow struct {
	Kind, Text, Want                               string
	Nodes                                          []captureNode
	Index                                          int
	Capitalise, TreatUndefined, HasHook, HookValue bool
}

var captureMutex sync.Mutex

func captureBytes(s string) string {
	parts := []string{}
	for _, b := range []byte(s) {
		parts = append(parts, strconv.Itoa(int(b)))
	}
	return strings.Join(parts, ",")
}
func captureWrite(r captureRow) {
	path := os.Getenv("ADAMIC_RETURN_CAPTURE")
	if path == "" {
		return
	}
	b, e := json.Marshal(r)
	if e != nil {
		panic(e)
	}
	captureMutex.Lock()
	defer captureMutex.Unlock()
	f, e := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		panic(e)
	}
}
func captureGraph(root *ast.Node, token func(*ast.Node) core.TextRange) ([]captureNode, int) {
	nodes := []captureNode{}
	ids := map[*ast.Node]int{}
	var add func(*ast.Node) int
	add = func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		if id, ok := ids[n]; ok {
			return id
		}
		id := len(nodes)
		ids[n] = id
		nodes = append(nodes, captureNode{})
		r := captureNode{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Expression: -1, Name: -1, Parent: -1, Arrow: -1, Pos: n.Pos(), End: n.End()}
		switch n.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
			r.Text = n.Text()
		}
		r.Name = add(n.Name())
		// Only immediate parent kind/key matters; ancestor chains are unnecessary.
		if n.Parent != nil && (n.Kind == ast.KindFunctionExpression) {
			r.Parent = add(n.Parent)
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			r.Expression = add(n.AsReturnStatement().Expression)
		case ast.KindComputedPropertyName:
			r.Expression = add(n.AsComputedPropertyName().Expression)
		case ast.KindParenthesizedExpression:
			r.Expression = add(n.AsParenthesizedExpression().Expression)
		case ast.KindArrowFunction:
			r.Arrow = add(n.AsArrowFunction().EqualsGreaterThanToken)
		}
		r.Generator = adamicRawIsGenerator(n)
		r.StaticModifier = ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic)
		r.AsyncModifier = ast.HasSyntacticModifier(n, ast.ModifierFlagsAsync)
		if token != nil {
			tr := token(n)
			r.Pos = tr.Pos()
			r.End = tr.End()
		}
		nodes[id] = r
		return id
	}
	index := add(root)
	return nodes, index
}
func captureAst(kind string, n *ast.Node, want string, flag bool, token func(*ast.Node) core.TextRange) {
	nodes, index := captureGraph(n, token)
	captureWrite(captureRow{Kind: kind, Nodes: nodes, Index: index, Want: want, Capitalise: flag})
}
func IsScope(n *ast.Node) bool {
	v := adamicRawIsScope(n)
	captureAst("IsScope", n, strconv.FormatBool(v), false, nil)
	return v
}
func IsGenerator(n *ast.Node) bool {
	v := adamicRawIsGenerator(n)
	captureAst("IsGenerator", n, strconv.FormatBool(v), false, nil)
	return v
}
func isExemptFromEndJudgment(n *ast.Node) bool {
	v := adamicRawisExemptFromEndJudgment(n)
	captureAst("isExemptFromEndJudgment", n, strconv.FormatBool(v), false, nil)
	return v
}
func Name(n *ast.Node, c bool) string {
	v := adamicRawName(n, c)
	captureAst("Name", n, captureBytes(v), c, nil)
	return v
}
func staticName(n *ast.Node) string {
	v := adamicRawstaticName(n)
	captureAst("staticName", n, captureBytes(v), false, nil)
	return v
}
func capitaliseFirst(s string, c bool) string {
	v := adamicRawcapitaliseFirst(s, c)
	captureWrite(captureRow{Kind: "capitaliseFirst", Text: s, Capitalise: c, Want: captureBytes(v)})
	return v
}
func Verb(s string) string {
	v := adamicRawVerb(s)
	captureWrite(captureRow{Kind: "Verb", Text: s, Want: captureBytes(v)})
	return v
}
func ReportRange(n *ast.Node, token func(*ast.Node) core.TextRange) core.TextRange {
	v := adamicRawReportRange(n, token)
	captureAst("ReportRange", n, strconv.Itoa(v.Pos())+":"+strconv.Itoa(v.End()), false, token)
	return v
}
func HasValue(n *ast.Node, s Settings, h Hooks) bool {
	hookValue := false
	calls := 0
	original := h.TreatsArgumentAsUnspecified
	if original != nil {
		h.TreatsArgumentAsUnspecified = func(a *ast.Node) bool { calls++; hookValue = original(a); return hookValue }
	}
	v := adamicRawHasValue(n, s, h)
	nodes, index := captureGraph(n, nil)
	captureWrite(captureRow{Kind: "HasValue", Nodes: nodes, Index: index, TreatUndefined: s.TreatUndefinedAsUnspecified, HasHook: original != nil, HookValue: hookValue, Want: strconv.FormatBool(v) + ":" + strconv.Itoa(calls)})
	return v
}
