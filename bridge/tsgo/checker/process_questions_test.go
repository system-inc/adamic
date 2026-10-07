package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestProcessQuestionsRawSyntaxAndFlow(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	text := "/* 世界 🌍 */\r\nconst chosen = true;\nexport function go(): void { if(chosen) console.log('yes'); else console.log('no'); }\nconst delayed = new Promise((resolve) => { setTimeout(resolve, 5); });\n"
	for path, contents := range map[string]string{file: text, config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"]},"files":["input.a"]}`} {
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := program.Compiler.GetSourceFile(file)
	nodes, ids := syntaxNodes(source)
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), nodeKind(node), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	fields := ask(source.AsNode(), "syntax-projection")
	if fields[0] != "1" || fields[1] != "syntax-projection" || fields[2] != file || fields[3] != "0" || fields[4] != "1" || fields[5] != strconv.Itoa(len(nodes)) {
		t.Fatalf("projection header: %q", fields[:6])
	}
	position := 6
	for _, node := range nodes {
		if position+16 >= len(fields) {
			t.Fatal("truncated syntax")
		}
		record := fields[position : position+16]
		position += 16
		if record[0] != nodeKind(node) || record[1] != strconv.Itoa(node.Pos()) || record[2] != strconv.Itoa(node.End()) || record[4] != strconv.FormatUint(ids[node.Parent], 10) {
			t.Fatal("syntax identity mismatch", record)
		}
		start, err := strconv.Atoi(record[3])
		if err != nil || start < node.Pos() || start > node.End() {
			t.Fatal("bad token start", record[3])
		}
		if node.Kind == ast.KindVariableDeclarationList && record[6] != "2" {
			t.Fatal("lost const flags")
		}
		for count := 0; count < 2; count++ {
			length, err := strconv.Atoi(fields[position])
			if err != nil {
				t.Fatal(err)
			}
			position++
			for _, field := range fields[position : position+length] {
				id, err := strconv.Atoi(field)
				if err != nil || id < 1 || id > len(nodes) {
					t.Fatal("bad syntax edge")
				}
			}
			position += length
		}
	}
	if position != len(fields) {
		t.Fatal("trailing syntax fields")
	}
	var function, call, identifier *ast.Node
	for _, node := range nodes {
		if node.Kind == ast.KindFunctionDeclaration {
			function = node
		}
		if node.Kind == ast.KindCallExpression && strings.Contains(text[node.Pos():node.End()], "setTimeout(") {
			call = node
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "chosen" {
			identifier = node
		}
	}
	if function == nil || call == nil || identifier == nil {
		t.Fatal("missing fixture nodes")
	}
	flow := ask(function, "syntax-control-flow")
	count, _ := strconv.Atoi(flow[2])
	if count < 3 {
		t.Fatal("branch was flattened")
	}
	at := 3
	reachable := 0
	events := 0
	for block := 0; block < count; block++ {
		if flow[at] == "1" {
			reachable++
		}
		at++
		length, _ := strconv.Atoi(flow[at])
		at++
		for _, edge := range flow[at : at+length] {
			id, err := strconv.Atoi(edge)
			if err != nil || id < 0 || id >= count {
				t.Fatal("bad control-flow edge")
			}
		}
		at += length
		length, _ = strconv.Atoi(flow[at])
		at++
		for event := 0; event < length; event++ {
			hook := flow[at]
			id, err := strconv.Atoi(flow[at+1])
			if (hook != "0" && hook != "1") || err != nil || id < 1 || id > len(nodes) {
				t.Fatal("bad syntax event")
			}
			events++
			at += 2
		}
	}
	if reachable < 3 || events < 2 || at != len(flow) {
		t.Fatalf("lost reachable branches or events: %d %d", reachable, events)
	}
	symbol := ask(identifier, "process-symbol-details\nreference")
	if symbol[2] == "0" || symbol[4] != "chosen" || symbol[5] != "1" || symbol[10] != "VariableDeclaration" {
		t.Fatal("lost declaration identity", symbol)
	}
	signature := ask(call, "resolved-call-declaration")
	if signature[2] != "1" {
		t.Fatal("lost resolved timer signature")
	}
	imports := ask(source.AsNode(), "program-imports")
	if imports[0] != "1" || imports[1] != "program-imports" || !strings.Contains(strings.Join(imports, "\n"), file) {
		t.Fatal("program omitted its root")
	}
	for _, row := range []struct {
		node     *ast.Node
		question string
	}{{source.AsNode(), "syntax-projection\nextra"}, {identifier, "syntax-projection"}, {identifier, "syntax-control-flow"}, {call, "resolved-call-declaration\nextra"}, {identifier, "resolved-call-declaration"}, {identifier, "program-imports"}, {source.AsNode(), "program-imports\nextra"}, {identifier, "process-symbol-details"}, {identifier, "process-symbol-details\nunknown"}, {identifier, "process-symbol-details\nown\nextra"}} {
		if _, err := program.Inspect(file, uint64(row.node.Pos()), uint64(row.node.End()), nodeKind(row.node), row.question); err == nil {
			t.Fatal("accepted malformed question", row.question)
		}
	}
}
