// Independent Go HIR control oracle. No bridge imports and no Go lint verdicts.
package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

type block struct {
	id                  int
	terminal            string
	predecessors, tests []int
}
type graph struct {
	entry     int
	blocks    []block
	controls  []int
	reference *hir.Function
}

func main() {
	directory := os.Args[1]
	graphs := []graph{{}}
	for n := 1; n <= 3; n++ {
		choices := (1 << n) + 1
		combinations := 1
		for i := 0; i < n; i++ {
			combinations *= choices
		}
		for code := 0; code < combinations; code++ {
			states := make([]int, n)
			value := code
			for i := range states {
				states[i] = value % choices
				value /= choices
			}
			for entry := 1; entry <= n; entry++ {
				for mask := 0; mask < (1 << n); mask++ {
					g := graph{entry: entry}
					for at, state := range states {
						b := block{id: at + 1}
						switch state {
						case 0:
							b.terminal = "Return"
						case 1:
							b.terminal = "Throw"
						default:
							switch (at + 1) % 3 {
							case 0:
								b.terminal = "Switch"
								b.tests = []int{90 + at, at + 1}
							case 1:
								b.terminal = "If"
								b.tests = []int{at + 1}
							case 2:
								b.terminal = "Branch"
								b.tests = []int{at + 1}
							}
						}
						g.blocks = append(g.blocks, b)
					}
					for from, state := range states {
						if state < 2 {
							continue
						}
						successors := state - 1
						for to := 0; to < n; to++ {
							if successors&(1<<to) != 0 {
								g.blocks[to].predecessors = append(g.blocks[to].predecessors, from+1)
							}
						}
					}
					for at := 0; at < n; at++ {
						if mask&(1<<at) != 0 {
							g.controls = append(g.controls, at+1)
						}
					}
					if code%2 == 1 {
						for a, b := 0, n-1; a < b; a, b = a+1, b-1 {
							g.blocks[a], g.blocks[b] = g.blocks[b], g.blocks[a]
						}
					}
					graphs = append(graphs, g)
				}
			}
		}
	}
	for _, controls := range [][]int{nil, {10}, {20}, {30}, {90}, {10, 20, 30}} {
		graphs = append(graphs, graph{entry: 10, controls: controls, blocks: []block{
			{id: 10, terminal: "Switch", tests: []int{90, 10}},
			{id: 20, terminal: "Goto", predecessors: []int{10}, tests: []int{20}},
			{id: 30, terminal: "Return", predecessors: []int{10, 20}},
		}})
	}

	sources := 0
	file, err := goparser.ParseFile(token.NewFileSet(), os.Args[2], nil, 0)
	if err != nil {
		panic(err)
	}
	goast.Inspect(file, func(n goast.Node) bool {
		literal, ok := n.(*goast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		code, err := strconv.Unquote(literal.Value)
		if err != nil || !strings.HasPrefix(code, "function f(") {
			return true
		}
		sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/control.tsx", Path: "/control.tsx"}, code, core.ScriptKindTSX)
		if len(sf.Diagnostics()) != 0 {
			panic("invalid source control")
		}
		var root *ast.Node
		sf.AsNode().ForEachChild(func(n *ast.Node) bool {
			if ast.IsFunctionLike(n) {
				root = n
				return true
			}
			return false
		})
		if root == nil {
			panic("missing source function")
		}
		function := hir.Lower(root, nil)
		g := graph{entry: int(function.Entry), reference: function}
		all := []int{}
		seen := map[int]bool{}
		for _, bb := range function.Blocks {
			b := block{id: int(bb.Id)}
			if bb.Terminal != nil {
				b.terminal = reflect.TypeOf(bb.Terminal).Elem().Name()
			}
			for _, id := range bb.Predecessors {
				b.predecessors = append(b.predecessors, int(id))
			}
			appendTest := func(p hir.Place) {
				id := int(p.Identifier)
				b.tests = append(b.tests, id)
				if !seen[id] {
					seen[id] = true
					all = append(all, id)
				}
			}
			switch v := bb.Terminal.(type) {
			case *hir.If:
				appendTest(v.Test)
			case *hir.Branch:
				appendTest(v.Test)
			case *hir.Switch:
				appendTest(v.Test)
				for _, c := range v.Cases {
					if c.Test != nil {
						appendTest(*c.Test)
					}
				}
			}
			g.blocks = append(g.blocks, b)
		}
		graphs = append(graphs, g)
		for _, id := range all {
			one := g
			one.controls = []int{id}
			graphs = append(graphs, one)
		}
		g.controls = all
		graphs = append(graphs, g)
		sources++
		return true
	})
	if sources < 20 {
		panic("source graph extraction is vacuous")
	}
	var frames, truth strings.Builder
	frame := func(v any) { s := fmt.Sprint(v); fmt.Fprintf(&frames, "%d\n%s", len(utf16.Encode([]rune(s))), s) }
	ids := func(values []int) {
		frame(len(values))
		for _, v := range values {
			frame(v)
		}
	}
	frame(len(graphs))
	elapsed := time.Duration(0)
	for index, g := range graphs {
		frame(g.entry)
		frame(len(g.blocks))
		function := g.reference
		if function == nil {
			function = &hir.Function{Entry: hir.BlockId(g.entry)}
		}
		query := []int{}
		for _, b := range g.blocks {
			frame(b.id)
			frame(b.terminal)
			ids(b.predecessors)
			ids(b.tests)
			if g.reference != nil {
				query = append(query, b.id)
				continue
			}
			bb := &hir.BasicBlock{Id: hir.BlockId(b.id)}
			for _, p := range b.predecessors {
				bb.Predecessors = append(bb.Predecessors, hir.BlockId(p))
			}
			test := func(id int) hir.Place { return hir.Place{Identifier: hir.IdentifierId(id)} }
			switch b.terminal {
			case "Return":
				bb.Terminal = &hir.Return{}
			case "Throw":
				bb.Terminal = &hir.Throw{}
			case "Goto":
				bb.Terminal = &hir.Goto{}
			case "If":
				bb.Terminal = &hir.If{Test: test(b.tests[0])}
			case "Branch":
				bb.Terminal = &hir.Branch{Test: test(b.tests[0])}
			case "Switch":
				v := test(b.tests[1])
				bb.Terminal = &hir.Switch{Test: test(b.tests[0]), Cases: []hir.SwitchCase{{}, {Test: &v}}}
			default:
				panic("unknown terminal")
			}
			function.Blocks = append(function.Blocks, bb)
			query = append(query, b.id)
		}
		query = append(query, 999)
		ids(g.controls)
		ids(query)
		values := map[int]bool{}
		for _, v := range g.controls {
			values[v] = true
		}
		before := time.Now()
		uncond := hir.UnconditionalBlocks(function)
		controlled := hir.ControlDominators(function, func(p hir.Place) bool { return values[int(p.Identifier)] })
		var results strings.Builder
		for _, id := range query {
			if controlled(hir.BlockId(id)) {
				results.WriteByte('1')
			} else {
				results.WriteByte('0')
			}
		}
		elapsed += time.Since(before)
		ordered := []int{}
		for id, yes := range uncond {
			if yes {
				ordered = append(ordered, int(id))
			}
		}
		sort.Ints(ordered)
		fmt.Fprintf(&truth, "%d\t", index)
		for _, id := range ordered {
			fmt.Fprintf(&truth, "%d,", id)
		}
		fmt.Fprintf(&truth, "\t%s\n", results.String())
	}
	for name, text := range map[string]string{"graphs.frames": frames.String(), "go.expected": truth.String()} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); err != nil {
			panic(err)
		}
	}
	fmt.Printf("graphs=%d sources=%d bytes=%d control_ns=%d\n", len(graphs), sources, truth.Len(), elapsed.Nanoseconds())
}
