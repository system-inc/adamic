// Overlay test invokes the unchanged production validator on supplied HIR.
// This deliberately does not stand in for source-to-SSA lowering.
package react

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

type wave29Op struct {
	kind                         string
	target, source, binding, tag int
}
type wave29Block struct {
	instructions []int
	phiTarget    int
	operands     []int
}
type wave29Case struct {
	name                       string
	ops                        []wave29Op
	blocks                     []wave29Block
	missingCreator, missingTag bool
}

func wave29Written(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, h, l)
		}
	}
	return out.String()
}
func TestWave29StaticCore(t *testing.T) {
	output := os.Getenv("ADAMIC_WAVE29_STATIC_CORE")
	if output == "" {
		t.Fatal("missing output directory")
	}
	code := "/* 世界 🌍 */\r\nconst made = makeIt(); const Bound = made; const Widget = made;\r\n/* tag */ <Bound/>; const factory = () => null;"
	sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/input.tsx", Path: "/input.tsx"}, code, core.ScriptKindTSX)
	nodes := map[string][]*ast.Node{}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier || n.Kind == ast.KindArrowFunction {
			r := rule.TokenRange(sf, n)
			nodes[code[r.Pos():r.End()]] = append(nodes[code[r.Pos():r.End()]], n)
		}
		n.ForEachChild(func(c *ast.Node) bool { walk(c); return false })
	}
	walk(sf.AsNode())
	wanted := []*ast.Node{nodes["makeIt"][0], nodes["made"][0], nodes["Bound"][0], nodes["Bound"][1], nodes["() => null"][0], nodes["Widget"][0], nil, nil, nil, nil, nil, nil, nil, nil, nil, nil}
	op := func(k string, t, s, b, tag int) wave29Op { return wave29Op{k, t, s, b, tag} }
	jsx := op("JsxExpression", 15, -1, -1, 3)
	cases := []wave29Case{}
	add := func(name string, ops ...wave29Op) {
		ids := make([]int, len(ops))
		for i := range ids {
			ids[i] = i
		}
		cases = append(cases, wave29Case{name: name, ops: ops, blocks: []wave29Block{{instructions: ids, phiTarget: -1}}})
	}
	for _, kind := range []string{"CallExpression", "NewExpression", "MethodCall"} {
		add(kind, op(kind, 3, 0, -1, -1), jsx)
	}
	add("function", op("FunctionExpression", 4, -1, -1, -1), op("LoadLocal", 3, 4, -1, -1), jsx)
	add("alias", op("CallExpression", 1, 0, -1, -1), op("LoadLocal", 2, 1, -1, -1), op("LoadLocal", 3, 2, -1, -1), jsx)
	add("store-binding", op("CallExpression", 1, 0, -1, -1), op("StoreLocal", 7, 1, 3, -1), jsx)
	add("store-result", op("CallExpression", 1, 0, -1, -1), op("StoreLocal", 3, 1, 7, -1), jsx)
	add("clean", jsx)
	add("clean-load", op("LoadLocal", 3, 1, -1, -1), jsx)
	add("clean-store", op("StoreLocal", 3, 1, 7, -1), jsx)
	add("load-context-ignored", op("CallExpression", 1, 0, -1, -1), op("LoadContext", 3, 1, -1, -1), jsx)
	add("property-load-ignored", op("CallExpression", 1, 0, -1, -1), op("PropertyLoad", 3, 1, -1, -1), jsx)
	add("host", op("CallExpression", 3, 0, -1, -1), op("JsxExpression", 15, -1, -1, -1))
	add("duplicate", op("CallExpression", 3, 0, -1, -1), jsx, jsx)
	add("nil-instruction", op("CallExpression", 3, 0, -1, -1), op("", 0, 0, 0, 0), jsx)
	add("creator-fallback", op("CallExpression", 3, 6, -1, -1), jsx)
	add("tag-fallback", op("CallExpression", 3, 0, -1, -1), jsx)
	cases[len(cases)-1].missingTag = true
	add("both-fallback", op("CallExpression", 3, 0, -1, -1), jsx)
	cases[len(cases)-1].missingTag = true
	cases[len(cases)-1].missingCreator = true
	for _, dynamic := range []bool{false, true} {
		k := "Primitive"
		if dynamic {
			k = "CallExpression"
		}
		cases = append(cases, wave29Case{name: fmt.Sprintf("phi-%v", dynamic), ops: []wave29Op{op(k, 1, 0, -1, -1), jsx}, blocks: []wave29Block{{instructions: []int{0}, phiTarget: -1}, {instructions: []int{1}, phiTarget: 3, operands: []int{7, 1, 8}}}})
	}
	cases = append(cases, wave29Case{name: "forward-block", ops: []wave29Op{op("CallExpression", 3, 0, -1, -1), jsx}, blocks: []wave29Block{{instructions: []int{0}, phiTarget: -1}, {instructions: []int{1}, phiTarget: -1}}})
	cases = append(cases, wave29Case{name: "backedge-no-fixpoint", ops: []wave29Op{op("CallExpression", 3, 0, -1, -1), jsx}, blocks: []wave29Block{{instructions: []int{1}, phiTarget: -1}, {instructions: []int{0}, phiTarget: -1}}})
	// Long alias chains exercise allocation and propagation without changing the oracle.
	for length := 1; length <= 12; length++ {
		ops := []wave29Op{op("CallExpression", 1, 0, -1, -1)}
		prev := 1
		for i := 0; i < length; i++ {
			target := i + 2
			ops = append(ops, op("LoadLocal", target, prev, -1, -1))
			prev = target
		}
		ops = append(ops, op("JsxExpression", 15, -1, -1, prev))
		add(fmt.Sprintf("chain-%d", length), ops...)
	}
	var records, truth, names strings.Builder
	frame := func(v any) { s := fmt.Sprint(v); fmt.Fprintf(&records, "%d\n%s", len(utf16.Encode([]rune(s))), s) }
	ids := func(v []int) {
		frame(len(v))
		for _, n := range v {
			frame(n)
		}
	}
	u16 := func(bytePos int) int { return len(utf16.Encode([]rune(code[:bytePos]))) }
	place := func(id int) hir.Place { return hir.Place{Identifier: hir.IdentifierId(id)} }
	frame(len(cases))
	findings := 0
	for index, c := range cases {
		fmt.Fprintf(&names, "%d\t%s\n", index, c.name)
		fmt.Fprintf(&truth, "case %d\n", index)
		identifiers := make([]*hir.Identifier, len(wanted))
		frame(code)
		frame(len(wanted))
		for id, node := range wanted {
			if c.missingCreator && id == 0 || c.missingTag && id == 3 {
				node = nil
			}
			identifiers[id] = &hir.Identifier{Id: hir.IdentifierId(id), Node: node}
			if node == nil {
				frame(0)
			} else {
				frame(1)
				frame(u16(node.Pos()))
				frame(u16(node.End()))
			}
		}
		instructions := make([]*hir.Instruction, len(c.ops))
		frame(len(c.ops))
		rawPos := strings.Index(code, "/* tag */")
		rawEnd := rule.TokenRange(sf, wanted[3]).End()
		for id, o := range c.ops {
			frame(o.kind)
			if o.kind == "" {
				continue
			}
			frame(o.target)
			frame(o.source)
			frame(o.binding)
			frame(o.tag)
			frame(u16(rawPos))
			frame(u16(rawEnd))
			var value hir.InstructionValue
			switch o.kind {
			case "FunctionExpression":
				value = &hir.FunctionExpression{}
			case "CallExpression":
				value = &hir.CallExpression{Callee: place(o.source)}
			case "NewExpression":
				value = &hir.NewExpression{Callee: place(o.source)}
			case "MethodCall":
				value = &hir.MethodCall{Property: place(o.source)}
			case "LoadLocal":
				value = &hir.LoadLocal{Place: place(o.source)}
			case "LoadContext":
				value = &hir.LoadContext{Place: place(o.source)}
			case "StoreLocal":
				value = &hir.StoreLocal{LValue: place(o.binding), Value: place(o.source)}
			case "PropertyLoad":
				value = &hir.PropertyLoad{Object: place(o.source), Property: "Inner"}
			case "Primitive":
				value = &hir.Primitive{}
			case "JsxExpression":
				tag := hir.JsxTag{Name: "div"}
				if o.tag >= 0 {
					p := place(o.tag)
					p.Range = core.NewTextRange(rawPos, rawEnd)
					tag.Place = &p
				}
				value = &hir.JsxExpression{Tag: tag}
			default:
				t.Fatalf("unknown op %q", o.kind)
			}
			instructions[id] = &hir.Instruction{Id: hir.InstructionId(id), LValue: place(o.target), Value: value}
		}
		blocks := []*hir.BasicBlock{}
		frame(len(c.blocks))
		for _, b := range c.blocks {
			block := &hir.BasicBlock{}
			if b.phiTarget < 0 {
				frame(0)
			} else {
				frame(1)
				frame(b.phiTarget)
				ids(b.operands)
				phi := &hir.Phi{Place: place(b.phiTarget), Operands: map[hir.BlockId]hir.Place{}}
				for i, v := range b.operands {
					phi.Operands[hir.BlockId(i)] = place(v)
				}
				block.Phis = append(block.Phis, phi)
			}
			ids(b.instructions)
			for _, v := range b.instructions {
				block.Instructions = append(block.Instructions, hir.InstructionId(v))
			}
			blocks = append(blocks, block)
		}
		ctx := rule.Context{SourceFile: sf, Report: func(d rule.Diagnostic) {
			findings++
			fmt.Fprintf(&truth, "%d\t%d\treact-hooks/static-components\t%s\t%s\t%d\t%d\n", d.Range.Pos(), d.Range.End(), d.Message.Id, wave29Written(d.Message.Description), len(d.Fixes), len(d.Suggestions))
		}}
		reportDynamicComponents(ctx, &hir.Function{Blocks: blocks, Instructions: instructions, Identifiers: identifiers})
	}
	for name, text := range map[string]string{"graphs.frames": records.String(), "go.expected": truth.String(), "cases.tsv": names.String()} {
		if err := os.WriteFile(filepath.Join(output, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("cases=%d findings=%d bytes=%d", len(cases), findings, truth.Len())
}
