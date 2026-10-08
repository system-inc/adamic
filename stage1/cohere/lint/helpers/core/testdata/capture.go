package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf16"
)

var adamicCaptureMu sync.Mutex
var adamicFrames = map[*ast.Node]*adamicFrame{}

type adamicFrame struct {
	key string
	ids map[*ast.Node]int
}

func adamicU16(source string, pos int) int {
	if pos < 0 {
		return 0
	}
	if pos > len(source) {
		pos = len(source)
	}
	return len(utf16.Encode([]rune(source[:pos])))
}
func adamicFrameFor(node *ast.Node, sf *ast.SourceFile) *adamicFrame {
	root := node
	if sf != nil {
		root = sf.AsNode()
	} else if root != nil {
		for root.Parent != nil {
			root = root.Parent
		}
	}
	if frame := adamicFrames[root]; frame != nil {
		return frame
	}
	source := ""
	if root != nil && root.Kind == ast.KindSourceFile {
		sf = root.AsSourceFile()
		source = sf.Text()
	}
	ids := map[*ast.Node]int{}
	var nodes []*ast.Node
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
		nodes = append(nodes, n)
		n.ForEachChild(func(c *ast.Node) bool { add(c); return false })
		return id
	}
	add(root)
	id := func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		v, ok := ids[n]
		if !ok {
			panic("projection missed node")
		}
		return v
	}
	list := func(ns *ast.NodeList) []int {
		out := []int{}
		if ns != nil {
			for _, n := range ns.Nodes {
				out = append(out, id(n))
			}
		}
		return out
	}
	rows := []map[string]any{}
	for _, n := range nodes {
		start, end := n.Pos(), n.End()
		if sf != nil {
			r := rule.TokenRange(sf, n)
			start, end = r.Pos(), r.End()
		}
		children := []int{}
		n.ForEachChild(func(c *ast.Node) bool { children = append(children, id(c)); return false })
		row := map[string]any{"kind": strings.TrimPrefix(n.Kind.String(), "Kind"), "kindNumber": int(n.Kind), "start": adamicU16(source, start), "end": adamicU16(source, end), "parent": id(n.Parent), "children": children}
		switch n.Kind {
		case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral:
			row["text"] = n.Text()
		}
		switch n.Kind {
		case ast.KindParenthesizedExpression:
			row["expression"] = id(n.AsParenthesizedExpression().Expression)
		case ast.KindPropertyAccessExpression:
			row["expression"] = id(n.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			row["expression"] = id(n.AsElementAccessExpression().Expression)
		case ast.KindVoidExpression:
			row["expression"] = id(n.AsVoidExpression().Expression)
		case ast.KindCallExpression:
			c := n.AsCallExpression()
			row["expression"] = id(c.Expression)
			row["args"] = list(c.Arguments)
		case ast.KindBinaryExpression:
			b := n.AsBinaryExpression()
			row["left"] = id(b.Left)
			row["right"] = id(b.Right)
			row["operator"] = strings.TrimPrefix(b.OperatorToken.Kind.String(), "Kind")
		case ast.KindPropertyAssignment:
			p := n.AsPropertyAssignment()
			row["name"] = id(p.Name())
			row["initializer"] = id(p.Initializer)
		case ast.KindShorthandPropertyAssignment:
			row["name"] = id(n.Name())
		case ast.KindImportAttribute:
			row["name"] = id(n.Name())
		case ast.KindArrayLiteralExpression:
			row["elements"] = list(n.AsArrayLiteralExpression().Elements)
		case ast.KindObjectLiteralExpression:
			row["elements"] = list(n.AsObjectLiteralExpression().Properties)
		case ast.KindFunctionDeclaration:
			row["generator"] = n.AsFunctionDeclaration().AsteriskToken != nil
		case ast.KindFunctionExpression:
			row["generator"] = n.AsFunctionExpression().AsteriskToken != nil
		case ast.KindMethodDeclaration:
			row["generator"] = n.AsMethodDeclaration().AsteriskToken != nil
		case ast.KindBlock:
			row["statements"] = list(n.AsBlock().Statements)
		case ast.KindIfStatement:
			s := n.AsIfStatement()
			row["thenNode"] = id(s.ThenStatement)
			row["elseNode"] = id(s.ElseStatement)
		case ast.KindTryStatement:
			s := n.AsTryStatement()
			row["tryBlock"] = id(s.TryBlock.AsNode())
			if s.FinallyBlock != nil {
				row["finallyBlock"] = id(s.FinallyBlock.AsNode())
			}
			if s.CatchClause != nil {
				row["catchBlock"] = id(s.CatchClause.AsCatchClause().Block.AsNode())
			}
		case ast.KindSwitchStatement:
			row["clauses"] = list(n.AsSwitchStatement().CaseBlock.AsCaseBlock().Clauses)
		case ast.KindCaseClause, ast.KindDefaultClause:
			row["statements"] = list(n.AsCaseOrDefaultClause().Statements)
		case ast.KindLabeledStatement:
			row["statement"] = id(n.AsLabeledStatement().Statement)
		case ast.KindWithStatement:
			row["statement"] = id(n.AsWithStatement().Statement)
		}
		rows = append(rows, row)
	}
	data, err := json.Marshal(map[string]any{"source": source, "nodes": rows})
	if err != nil {
		panic(err)
	}
	key := fmt.Sprintf("%x", sha256.Sum256(data))
	if err = os.WriteFile(filepath.Join(os.Getenv("ADAMIC_CORE_CAPTURE"), key+".json"), data, 0644); err != nil {
		panic(err)
	}
	frame := &adamicFrame{key, ids}
	adamicFrames[root] = frame
	return frame
}
func adamicCapture(helper string, sf *ast.SourceFile, args []*ast.Node, extra any, want any) {
	if os.Getenv("ADAMIC_CORE_CAPTURE") == "" {
		return
	}
	adamicCaptureMu.Lock()
	defer adamicCaptureMu.Unlock()
	var first *ast.Node
	for _, n := range args {
		if n != nil {
			first = n
			break
		}
	}
	frame := adamicFrameFor(first, sf)
	indexes := []int{}
	for _, n := range args {
		if n == nil {
			indexes = append(indexes, -1)
		} else {
			indexes = append(indexes, frame.ids[n])
		}
	}
	if n, ok := want.(*ast.Node); ok {
		if n == nil {
			want = "-1"
		} else {
			want = fmt.Sprint(frame.ids[n])
		}
	}
	if sf != nil {
		if positions, ok := extra.([]int); ok {
			extra = []int{adamicU16(sf.Text(), positions[0]), adamicU16(sf.Text(), positions[1])}
		}
	}
	if b, ok := want.(bool); ok {
		want = fmt.Sprint(b)
	}
	if v, ok := want.(int); ok {
		want = fmt.Sprint(v)
	}
	data, err := json.Marshal(map[string]any{"helper": helper, "frame": frame.key, "args": indexes, "extra": extra, "want": want})
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(filepath.Join(os.Getenv("ADAMIC_CORE_CAPTURE"), "calls.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.Write(append(data, '\n')); err != nil {
		panic(err)
	}
}
