//go:build lintoracle

package decorators

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode/utf16"
	"unicode/utf8"
)

type adamicPair struct {
	Name  string
	Found bool
}
type adamicNode struct {
	Kind, Text string
	Expression int
	Modifiers  []int
	Pos, End   int
	Children   []int
}
type adamicArena struct {
	ID, Root int
	Source   string
	Nodes    []adamicNode
	Parents  []int
	ids      map[*ast.Node]int
}

var adamicAstMutex sync.Mutex
var adamicArenas = map[*ast.Node]*adamicArena{}

func recordAdamicAst(name string, input *ast.Node, names map[string]struct{}, value any) {
	adamicAstMutex.Lock()
	defer adamicAstMutex.Unlock()
	f, e := os.OpenFile(os.Getenv("ADAMIC_AST_CAPTURE"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	root := input
	for root != nil && root.Parent != nil {
		root = root.Parent
	}
	arena := adamicArenas[root]
	if arena == nil {
		arena = &adamicArena{ID: len(adamicArenas), Root: -1, Nodes: []adamicNode{}, Parents: []int{}, ids: map[*ast.Node]int{}}
		if root != nil && root.Kind == ast.KindSourceFile {
			arena.Source = root.AsSourceFile().Text()
		}
		if !utf8.ValidString(arena.Source) {
			panic("invalid UTF-8 source: " + name)
		}
		offsets := map[int]int{}
		unit := 0
		for position, point := range arena.Source {
			offsets[position] = unit
			if point > 65535 {
				unit += 2
			} else {
				unit++
			}
		}
		offsets[len(arena.Source)] = unit
		var visit func(*ast.Node) int
		visit = func(node *ast.Node) int {
			if node == nil {
				return -1
			}
			children := []int{}
			node.ForEachChild(func(child *ast.Node) bool { children = append(children, visit(child)); return false })
			data := adamicNode{Kind: strings.TrimPrefix(node.Kind.String(), "Kind"), Pos: offsets[node.Pos()], End: offsets[node.End()], Children: children, Expression: -1, Modifiers: []int{}}
			switch node.Kind {
			case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral:
				data.Text = node.Text()
			}

			switch node.Kind {
			case ast.KindDecorator:
				data.Expression = arena.ids[node.AsDecorator().Expression]
			case ast.KindCallExpression:
				data.Expression = arena.ids[node.AsCallExpression().Expression]
			}
			if modifiers := node.Modifiers(); modifiers != nil {
				for _, modifier := range modifiers.Nodes {
					if modifier != nil {
						data.Modifiers = append(data.Modifiers, arena.ids[modifier])
					}
				}
			}

			id := len(arena.Nodes)
			arena.Nodes = append(arena.Nodes, data)
			arena.Parents = append(arena.Parents, -1)
			arena.ids[node] = id
			for _, child := range children {
				arena.Parents[child] = id
			}
			return id
		}
		arena.Root = visit(root)
		adamicArenas[root] = arena
		if e = encoder.Encode(arena); e != nil {
			panic(e)
		}
	}
	index := func(node *ast.Node) int {
		if node == nil {
			return -1
		}
		id, ok := arena.ids[node]
		if !ok {
			panic("uncaptured node")
		}
		return id
	}
	out := ""
	switch v := value.(type) {
	case bool:
		out = fmt.Sprint(v)
	case string:
		out = v
	case *ast.Node:
		out = fmt.Sprint(index(v))
	case []*ast.Node:
		if v == nil {
			out = "nil"
			break
		}
		parts := []string{}
		for _, node := range v {
			parts = append(parts, fmt.Sprint(index(node)))
		}
		out = strings.Join(parts, ",")
	case adamicPair:
		out = fmt.Sprintf("%t|%s", v.Found, v.Name)
	default:
		panic(fmt.Sprintf("result %T", value))
	}
	if !utf8.ValidString(out) {
		panic("invalid UTF-8 helper result: " + name)
	}
	units := ""
	for _, u := range utf16.Encode([]rune(out)) {
		units += fmt.Sprintf("%d,", u)
	}
	keys := []string{}
	for key := range names {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if e = encoder.Encode(struct {
		Names        []string
		Name         string
		Arena, Input int
		Want         string
	}{keys, name, arena.ID, index(input), name + "|" + units}); e != nil {
		panic(e)
	}
}
