package high_level_intermediate_representation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Rebind test calls to observer functions. Only test identifiers are replaced, never Go HIR code.
func observeTestCalls(data []byte) []byte {
	var scan scanner.Scanner
	files := token.NewFileSet()
	file := files.AddFile("test.go", files.Base(), len(data))
	scan.Init(file, data, nil, 0)
	type edit struct {
		start, end int
		text       string
	}
	var edits []edit
	for {
		pos, tok, text := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.IDENT && (text == "Lower" || text == "ForFunction" || text == "ForFunctionWithoutManualMemoization") {
			offset := file.Offset(pos)
			edits = append(edits, edit{offset, offset + len(text), "stage1Observed" + text})
		}
	}
	for index := len(edits) - 1; index >= 0; index-- {
		e := edits[index]
		data = append(append(append([]byte{}, data[:e.start]...), []byte(e.text)...), data[e.end:]...)
	}
	return data
}
func TestWholeConstructionCensus(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	temp := t.TempDir()
	destination := os.Getenv("HIR_CENSUS_EXPORT")
	if destination == "" {
		destination = filepath.Join(temp, "census")
	}
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	replacements := map[string]string{
		filepath.Join(upstream, "stage1_hir_oracle_test.go"): filepath.Join(lane, "testdata/oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_census_test.go"): filepath.Join(lane, "testdata/census_test.go"),
	}
	files, err := filepath.Glob(filepath.Join(upstream, "*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		observed := observeTestCalls(data)
		copy := filepath.Join(temp, filepath.Base(file))
		if err := os.WriteFile(copy, observed, 0600); err != nil {
			t.Fatal(err)
		}
		replacements[file] = copy
	}
	encoded, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlay, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("go", "test", "-v", "-count=1", "-timeout=15m", "-tags=lintoracle", "-overlay", overlay, "./internal/lint/ecmascript/high_level_intermediate_representation")
	c.Dir = filepath.Join(root, "cohere")
	c.Env = append(os.Environ(), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "HIR_CENSUS=" + destination, "HIR_CORPUS=" + filepath.Join(lane, "testdata/corpus.txt"), "HIR_OUTPUT=" + filepath.Join(temp, "small.dump")}...)
	output, runErr := c.CombinedOutput()
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "go-tests.log"), output, 0600); err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("Go construction suite failed: %v; full log: %s", runErr, filepath.Join(destination, "go-tests.log"))
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "HIR construction census:") {
			t.Log(line)
		}
	}
	manifest := filepath.Join(destination, "manifest.tsv")
	node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "main.ts"), "--coverage", manifest)
	matched, total := compareConstructionCensus(t, node, manifest, false)
	binary := filepath.Join(temp, "hir")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(lane, "main.ts"), "-o", binary)
	native := command(t, root, nil, binary, "--coverage", manifest)
	if !bytes.Equal(native, node) {
		t.Fatal("native and Node census outputs differ: " + firstDifference(native, node))
	}
	t.Logf("%d/%d context-distinct functions match Go on Node and natively; remaining rows are explicit declines", matched, total)
	t.Logf("%d Go tests skipped; names recorded in go-tests.log", strings.Count(string(output), "--- SKIP:"))
	if matched == 0 || total == 0 {
		t.Fatal("empty construction coverage would prove nothing")
	}
	checkConstructionMutants(t, root, lane, manifest, node, true)
}
func compareConstructionCensus(t *testing.T, output []byte, manifest string, mutant bool) (int, int) {
	t.Helper()
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{}
	for _, chunk := range strings.Split(string(output), "case\t")[1:] {
		key, payload, ok := strings.Cut(chunk, "\n")
		if !ok {
			t.Fatal("missing case payload")
		}
		if _, exists := cases[key]; exists {
			t.Fatal("duplicate output key")
		}
		cases[key] = payload
	}
	total, matched := 0, 0
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		n, err := strconv.Atoi(fields[6])
		if err != nil {
			t.Fatal(err)
		}
		total += n
		got, ok := cases[fields[0]]
		if !ok {
			t.Fatal("missing census case: " + fields[0])
		}
		delete(cases, fields[0])
		if fields[7] != "true" {
			if got != "decline\n" {
				t.Fatal("out-of-slice case was silently admitted")
			}
			continue
		}
		want, err := os.ReadFile(fields[5])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte(got), want) {
			if mutant {
				continue
			}
			t.Fatalf("case %s (%s %s:%s checker=%s): %s", fields[0], fields[1], fields[2], fields[3], fields[4], firstDifference([]byte(got), want))
		}
		matched += n
	}
	if len(cases) != 0 {
		t.Fatal("unexpected extra census cases")
	}
	return matched, total
}

type constructionMutant struct{ name, file, from, to string }

var constructionMutants = []constructionMutant{
	{"capture read becomes local", "lower.ts", "kind: 'LoadContext', place:", "kind: 'LoadLocal', place:"},
	{"capture pairing loses outer value", "lower.ts", "const place = this.locals.get(symbol) ?? this.captureOf(symbol);", "const place = this.fn.returns;"},
	{"contextual outer writes become local", "lower.ts", "kind: this.contextual.has(symbol) ? 'StoreContext' : 'StoreLocal'", "kind: 'StoreLocal'"},
	{"context declaration registration disappears", "lower.ts", "this.fn.contextDeclarations.add((this.fn.identifiers[place.identifier] ?? panic('missing context identifier')).declaration);", "this.fn.contextDeclarations.has((this.fn.identifiers[place.identifier] ?? panic('missing context identifier')).declaration);"},
	{"hoisted function becomes ordinary let", "lower.ts", "value, declarationKind: 6 }, id, undefined);", "value, declarationKind: 1 }, id, undefined);"},

	{"optional call becomes unconditional", "lower.ts", "const optional = node.children.some((child) => this.parser.node(child).kind === 'QuestionDotToken');", "const optional = false;"},
	{"call arguments lose spread", "lower.ts", "spread: child.kind === 'SpreadElement'", "spread: false"},
	{"method call loses receiver", "lower.ts", "kind: 'MethodCall', receiver, property, args", "kind: 'MethodCall', receiver: property, property, args"},
	{"constructor uses return slot as callee", "lower.ts", "kind: 'NewExpression', callee, args: this.arguments(id)", "kind: 'NewExpression', callee: this.fn.returns, args: this.arguments(id)"},
	{"React callee origin disappears", "lower.ts", "const origin = this.exportOrigin(calleeId, new Set<number>());", "const origin: ModuleExportOriginInterface = { module: '', exported: '' };"},

	{"concise arrow returns unused slot", "lower.ts", "if(conciseId >= 0) { builder.close({ kind: 'Return', value: builder.expression(conciseId) }); }", "if(conciseId >= 0) { builder.expression(conciseId); builder.close({ kind: 'Return', value: fn.returns }); }"},
	{"anonymous assignment name disappears", "lower.ts", "if(name === '') {", "if(name !== '') {"},

	{"return store to nil", "lower.ts", "this.emit(value, id, this.fn.returns);", "this.emit({ kind: 'Primitive', literal: 'nil' }, id, this.fn.returns);"},
	{"binary right operand becomes left", "lower.ts", "left, operator: operators.get(operator) ?? panic('unsupported binary'), right", "left, operator: operators.get(operator) ?? panic('unsupported binary'), right: left"},
	{"unary operator becomes plus", "lower.ts", "{ kind: 'UnaryExpression', operator, value }", "{ kind: 'UnaryExpression', operator: '+', value }"},
	{"comma returns left operand", "lower.ts", "if(operator === 'CommaToken') { return right; }", "if(operator === 'CommaToken') { return left; }"},
	{"literal boolean flips", "lower.ts", "return 'bool:true';", "return 'bool:false';"},
	{"parameter becomes capture", "lower.ts", "fn.params.push(builder.bind(id));", "fn.context.push(builder.bind(id));"},
	{"symbol local read loses binding", "lower.ts", "const local = this.locals.get(this.identity(id));", "const local = this.locals.get(-1);"},
	{"import classification loses module", "lower.ts", "source = declaration.module;", "source = '';"},
	{"declaration kind becomes const", "lower.ts", "const declarationKind = list.semantic === '2' ? 0 : 1;", "const declarationKind = 0;"},
	{"assignment result loses RHS", "lower.ts", "return this.assign(node.children[0] ?? -1, this.expression(node.children[2] ?? -1));", "const rhs = this.expression(node.children[2] ?? -1); this.assign(node.children[0] ?? -1, rhs); return this.emit({ kind: 'Primitive', literal: 'nil' }, id, undefined);"},
	{"compound assignment uses plus", "lower.ts", "left, operator: compound, right", "left, operator: '-', right"},
	{"update operation flips", "lower.ts", "node.operator === 'PlusPlusToken' ? '++' : '--'", "node.operator === 'PlusPlusToken' ? '--' : '++'"},
	{"property load loses name", "lower.ts", "return this.emit({ kind: 'PropertyLoad', object, property: this.parser.node(node.children[1] ?? -1).text }", "return this.emit({ kind: 'PropertyLoad', object, property: '' }"},
	{"computed load loses key", "lower.ts", "{ kind: 'ComputedLoad', object, property: this.expression(node.children[1] ?? -1) }", "{ kind: 'ComputedLoad', object, property: object }"},
	{"if branch successors swap", "lower.ts", "kind: 'If', test, consequent: consequent.id, alternate: alternate.id", "kind: 'If', test, consequent: alternate.id, alternate: consequent.id"},
	{"ternary true arm becomes false", "lower.ts", "logical === undefined ? 2 : 0", "logical === undefined ? 4 : 0"},
	{"logical short circuit chooses wrong arm", "lower.ts", "const swapped = logical === '||' || logical === '??';", "const swapped = logical === '&&';"},
	{"while back edge exits loop", "lower.ts", "this.jumps.pop(); this.jump(test.id, 1);", "this.jumps.pop(); this.jump(fallthrough.id, 0);"},
	{"continue uses break target", "lower.ts", "node.kind === 'BreakStatement' ? target.breakBlock : target.continueBlock", "target.breakBlock"},
	{"throw becomes return", "lower.ts", "this.close({ kind: 'Throw', value: this.expression(node.children[0] ?? -1) });", "this.close({ kind: 'Return', value: this.expression(node.children[0] ?? -1) });"},
	{"return becomes unreachable", "lower.ts", "this.close({ kind: 'Return', value: this.fn.returns });", "this.close({ kind: 'Unreachable' });"},
	{"post abrupt instructions disappear", "lower.ts", "for(const id of ids) { this.statement(id); }", "for(const id of ids) { this.statement(id); if(this.current === undefined) { break; } }"},
}

func checkConstructionMutants(t *testing.T, root, lane, input string, want []byte, census bool) {
	t.Helper()
	// The old corpus has no binary/unary/comma/parameter witnesses; the full census plus probes does.
	for index, mutant := range constructionMutants {
		if !census && mutant.name != "return store to nil" {
			continue
		}
		t.Run("catches "+mutant.name, func(t *testing.T) {
			directory, err := os.MkdirTemp(filepath.Dir(lane), "hir-mutant-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(directory)
			files, err := filepath.Glob(filepath.Join(lane, "*.ts"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if filepath.Base(file) == mutant.file {
					if strings.Count(string(data), mutant.from) != 1 {
						t.Fatalf("mutant anchor moved: %s", mutant.name)
					}
					data = []byte(strings.Replace(string(data), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, filepath.Base(file)), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"node", "--no-warnings", "oracle/node.mjs", filepath.Join(directory, "main.ts")}
			if census {
				args = append(args, "--coverage")
			}
			args = append(args, input)
			badNode := command(t, root, nil, args...)
			if bytes.Equal(badNode, want) {
				t.Fatal("semantic mutant survived on Node")
			}
			binary := filepath.Join(t.TempDir(), fmt.Sprintf("mutant-%d", index))
			command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(directory, "main.ts"), "-o", binary)
			args = []string{binary}
			if census {
				args = append(args, "--coverage")
			}
			args = append(args, input)
			badNative := command(t, root, nil, args...)
			if !bytes.Equal(badNative, badNode) {
				t.Fatal("mutant differs between Node and native")
			}
			if bytes.Equal(badNative, want) {
				t.Fatal("semantic mutant survived natively")
			}
		})
	}
}
