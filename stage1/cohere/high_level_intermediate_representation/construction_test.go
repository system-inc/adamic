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
	provider, err := os.ReadFile(filepath.Join(root, "bridge/tsgo/checker/symbol_graph.go"))
	if err != nil {
		t.Fatal(err)
	}
	providerPath := filepath.Join(temp, "symbol_graph_test.go")
	if err := os.WriteFile(providerPath, []byte(strings.Replace(string(provider), "package checker", "package high_level_intermediate_representation", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	replacements[filepath.Join(upstream, "stage1_hir_symbol_graph_test.go")] = providerPath

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
	cachedNode := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "main.ts"), "--cached-coverage", manifest)
	cachedNative := command(t, root, nil, binary, "--cached-coverage", manifest)
	if !bytes.Equal(cachedNode, node) || !bytes.Equal(cachedNative, node) {
		t.Fatal("ForFunction census differs from constructed Go/Node/native census")
	}
	t.Log("ForFunction cache matches every admitted checker-backed row; checker-less cache requests decline")

	t.Logf("%d/%d context-distinct functions match Go on Node and natively; remaining rows are explicit declines", matched, total)
	t.Logf("%d Go tests skipped; names recorded in go-tests.log", strings.Count(string(output), "--- SKIP:"))
	if matched == 0 || total == 0 {
		t.Fatal("empty construction coverage would prove nothing")
	}
	checkConstructionMutants(t, root, lane, manifest, node, true)
	checkArenaIndexMutant(t, root, lane, manifest)
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
	{"cast changes type span", "lower.ts", "nodePos: this.byte(type.pos)", "nodePos: this.byte(type.pos) + 1"},
	{"property delete changes key", "lower.ts", "kind: 'PropertyDelete', object: this.expression(target.children[0] ?? -1), property: this.parser.node(target.children[target.children.length - 1] ?? -1).text", "kind: 'PropertyDelete', object: this.expression(target.children[0] ?? -1), property: 'other'"},
	{"computed delete loses key", "lower.ts", "kind: 'ComputedDelete', object: this.expression(target.children[0] ?? -1), property: this.expression(target.children[target.children.length - 1] ?? -1)", "kind: 'ComputedDelete', object: this.expression(target.children[0] ?? -1), property: this.fn.returns"},
	{"regex flags disappear", "lower.ts", "flags: node.text.slice(slash + 1)", "flags: ''"},
	{"template operand disappears", "lower.ts", "subexprs.push(this.expression(span.children[0] ?? -1))", "this.expression(span.children[0] ?? -1)"},
	{"tagged template changes tag", "lower.ts", "kind: 'TaggedTemplateExpression', tag, quasis, subexprs", "kind: 'TaggedTemplateExpression', tag: this.fn.returns, quasis, subexprs"},
	{"this changes global name", "lower.ts", "name: 'this', bindingKind: 0", "name: 'that', bindingKind: 0"},
	{"array hole disappears", "lower.ts", "spread: false, hole: true", "spread: false, hole: false"},
	{"array spread disappears", "lower.ts", "spread: element.kind === 'SpreadElement', hole: false", "spread: false, hole: false"},
	{"object spread disappears", "lower.ts", "value: this.expression(member.children[0] ?? -1), spread: true", "value: this.expression(member.children[0] ?? -1), spread: false"},
	{"object shorthand loses operand", "lower.ts", "const value = this.expression(member.kind === 'ShorthandPropertyAssignment' ? nameId : member.children[1] ?? -1);", "const value = member.kind === 'ShorthandPropertyAssignment' ? this.fn.returns : this.expression(member.children[1] ?? -1);"},
	{"computed object key loses operand", "lower.ts", "properties.push({ key: computedKey === undefined ? name.text : '', computedKey, value, spread: false });", "properties.push({ key: computedKey === undefined ? name.text : '', computedKey: computedKey === undefined ? undefined : value, value, spread: false });"},
	{"object key precedes initializer", "lower.ts", "const value = this.expression(member.kind === 'ShorthandPropertyAssignment' ? nameId : member.children[1] ?? -1);\n            const computedKey = name.kind === 'ComputedPropertyName' ? this.expression(name.children[0] ?? -1) : undefined;", "const computedKey = name.kind === 'ComputedPropertyName' ? this.expression(name.children[0] ?? -1) : undefined;\n            const value = this.expression(member.kind === 'ShorthandPropertyAssignment' ? nameId : member.children[1] ?? -1);"},
	{"star export origin disappears", "export_origin.ts", "for(const star of this.graph.symbol(module.symbol).stars)", "for(const star of this.graph.symbol(module.symbol).stars.slice(0, 0))"},

	{"jsx component tag loses value", "lower.ts", "tag = { name: '', place: this.loadIdentifier(tagId) };", "tag = { name: '', place: this.fn.returns };"},
	{"jsx host tag becomes component", "lower.ts", "if(name.text !== '' && name.text.charAt(0) >= 'a' && name.text.charAt(0) <= 'z')", "if(false)"},
	{"jsx bare prop becomes false", "lower.ts", "initializer === undefined ? this.emit({ kind: 'Primitive', literal: 'bool:true' }", "initializer === undefined ? this.emit({ kind: 'Primitive', literal: 'bool:false' }"},
	{"jsx spread loses marker", "lower.ts", "name: '', value: this.expression(attribute.children[0] ?? -1), spread: true", "name: '', value: this.expression(attribute.children[0] ?? -1), spread: false"},
	{"jsx trivia becomes text", "lower.ts", "if(child.semantic !== '1')", "if(true)"},
	{"jsx empty child becomes return place", "lower.ts", "if(inner !== undefined) { places.push(this.expression(inner)); }", "if(inner !== undefined) { places.push(this.expression(inner)); } else { places.push(this.fn.returns); }"},
	{"jsx fragment drops first child", "lower.ts", "{ kind: 'JsxFragment', children: this.jsxChildren(children) }", "{ kind: 'JsxFragment', children: this.jsxChildren(children.slice(1)) }"},

	{"do loop loses continue variant", "lower.ts", "this.jump(test.id, 1); this.current = test;", "this.jump(test.id, 0); this.current = test;"},
	{"for increment jumps past loop", "lower.ts", "this.expression(increment); this.jump(test.id, 0);", "this.expression(increment); this.jump(fallthrough.id, 0);"},
	{"iterator next loses collection", "lower.ts", "kind: 'IteratorNext', iterator, collection", "kind: 'IteratorNext', iterator, collection: iterator"},
	{"for in uses return slot as object", "lower.ts", "kind: 'NextPropertyOf', value: collection", "kind: 'NextPropertyOf', value: this.fn.returns"},
	{"switch cases stop falling through", "lower.ts", "this.jump(blocks[index + 1] ?? fallthrough.id, 0);", "this.jump(fallthrough.id, 0);"},
	{"label terminal targets exit", "lower.ts", "kind: 'Label', block: block.id, fallthrough: fallthrough.id", "kind: 'Label', block: fallthrough.id, fallthrough: fallthrough.id"},
	{"try completion loses variant", "lower.ts", "this.statement(node.children[0] ?? -1); this.jump(normal.id, 2);", "this.statement(node.children[0] ?? -1); this.jump(normal.id, 0);"},
	{"for condition disappears", "lower.ts", "const condition = node.slots[1] ?? -1;", "const condition = -1;"},

	{"capture read becomes local", "lower.ts", "kind: 'LoadContext', place:", "kind: 'LoadLocal', place:"},
	{"capture pairing loses outer value", "lower.ts", "const place = this.locals.get(symbol) ?? this.captureOf(symbol);", "const place = this.fn.returns;"},
	{"contextual outer writes become local", "lower.ts", "kind: this.contextual.has(symbol) ? 'StoreContext' : 'StoreLocal'", "kind: 'StoreLocal'"},
	{"context declaration registration disappears", "lower.ts", "this.fn.contextDeclarations.add(this.fn.identifier(place.identifier).declaration);", "this.fn.contextDeclarations.has(this.fn.identifier(place.identifier).declaration);"},
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
	{"if branch successors swap", "lower.ts", "kind: 'If', testPlace: test, consequent: consequent.id, alternate: alternate.id", "kind: 'If', testPlace: test, consequent: alternate.id, alternate: consequent.id"},
	{"ternary true arm loses value", "lower.ts", "place: firstValue }, firstNode, firstPlace", "place: logical === undefined ? this.fn.returns : firstValue }, firstNode, firstPlace"},
	{"logical short circuit chooses wrong arm", "lower.ts", "const swapped = logical === '||' || logical === '??';", "const swapped = logical === '&&';"},
	{"while back edge exits loop", "lower.ts", "this.jumps.pop(); this.jump(test.id, 1);", "this.jumps.pop(); this.jump(fallthrough.id, 0);"},
	{"continue uses break target", "lower.ts", "node.kind === 'BreakStatement' ? target.breakBlock : target.continueBlock", "target.breakBlock"},
	{"throw becomes return", "lower.ts", "this.close({ kind: 'Throw', value: this.expression(node.children[0] ?? -1) });", "this.close({ kind: 'Return', value: this.expression(node.children[0] ?? -1) });"},
	{"return becomes unreachable", "lower.ts", "this.close({ kind: 'Return', value: this.fn.returns });", "this.close({ kind: 'Unreachable' });"},
	{"post abrupt instructions join entry", "lower.ts", "if(this.current === undefined) { this.current = this.fn.newBlock('block'); }", "if(this.current === undefined) { this.current = this.fn.block(this.fn.entry); }"},
}

func checkConstructionMutants(t *testing.T, root, lane, input string, want []byte, census bool) {
	checkConstructionMutantsOn(t, root, lane, input, want, census, true)
}

// The full census always requires native. This separate opt-in Node certificate is
// useful while a documented native compiler gap is pending; it cannot green that gate.
func TestNodeConstructionReplay(t *testing.T) {
	manifest := os.Getenv("HIR_NODE_CENSUS")
	if manifest == "" {
		t.Skip("set HIR_NODE_CENSUS to the independently exported Go census")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "main.ts"), "--coverage", manifest)
	matched, total := compareConstructionCensus(t, node, manifest, false)
	cached := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "main.ts"), "--cached-coverage", manifest)
	if !bytes.Equal(node, cached) {
		t.Fatal("Node cached construction differs")
	}
	t.Logf("Node-only certificate: %d/%d including probes; native is not certified by this test", matched, total)
	checkConstructionMutantsOn(t, root, lane, manifest, node, true, false)
}
func checkConstructionMutantsOn(t *testing.T, root, lane, input string, want []byte, census, native bool) {
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
			if !native {
				return
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

// The index mutant must stop at a checked read, rather than silently produce a wrong graph.
func checkArenaIndexMutant(t *testing.T, root, lane, manifest string) {
	t.Helper()
	directory, err := os.MkdirTemp(filepath.Dir(lane), "hir-arena-mutant-")
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
		// Every copied module imports the same mutated home, preserving nominal identity.
		data = []byte(strings.ReplaceAll(string(data), "../arena/arena_index.a", "./arena_index.a"))
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(file)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	indices, err := os.ReadFile(filepath.Join(root, "stage1/cohere/arena/arena_index.a"))
	if err != nil {
		t.Fatal(err)
	}
	from := "new FunctionIndex(arena.length)"
	if strings.Count(string(indices), from) != 1 {
		t.Fatal("arena minting mutant anchor moved")
	}
	indices = []byte(strings.Replace(string(indices), from, "new FunctionIndex(arena.length + 1)", 1))
	if err := os.WriteFile(filepath.Join(directory, "arena_index.a"), indices, 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "arena-mutant")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(directory, "main.ts"), "-o", binary)
	for _, args := range [][]string{{"node", "--no-warnings", "oracle/node.mjs", filepath.Join(directory, "main.ts"), "--coverage", manifest}, {binary, "--coverage", manifest}} {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = root
		output, err := c.CombinedOutput()
		if err == nil {
			t.Fatal("off-by-one arena index did not stop")
		}
		if !strings.Contains(string(output), "FunctionIndex") || !strings.Contains(string(output), "out of range") {
			t.Fatalf("index mutant stopped for the wrong reason: %s", output)
		}
	}
	t.Log("off-by-one FunctionIndex mutant stopped at a checked arena read on Node and native")
}

func TestArenaIndexBrands(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	directory, err := os.MkdirTemp(filepath.Dir(lane), "hir-index-brand-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)
	// Import the production index types: they must be incompatible at the checker boundary.
	source := "import { FunctionIndex, BlockIndex } from '../arena/arena_index.a';\nconst functions: FunctionIndex[] = [];\nconst fn = FunctionIndex.push(functions);\nconst wrong: BlockIndex = fn;\nconsole.log(`${wrong.slot}`);\n"
	entry := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("go", "run", "./cmd/adamic", "build", entry, "-o", filepath.Join(t.TempDir(), "wrong-brand"))
	c.Dir = root
	output, err := c.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "nominal ancestry") {
		t.Fatalf("FunctionIndex accepted as BlockIndex or wrong refusal: %v\n%s", err, output)
	}
	_ = lane
}

func TestArenaOwnerIdentity(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	directory, err := os.MkdirTemp(filepath.Dir(lane), "hir-foreign-index-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)
	source := "import { HIRArena } from '../high_level_intermediate_representation/core.ts';\nconst first = new HIRArena(); first.create('First');\nconst second = new HIRArena(); const foreign = second.create('Second');\nconsole.log(first.read(foreign).name);\n"
	entry := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "foreign-index")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", entry, "-o", binary)
	for _, args := range [][]string{{"node", "--no-warnings", "oracle/node.mjs", entry}, {binary}} {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = root
		output, err := c.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "FunctionIndex belongs to another arena") {
			t.Fatalf("same-slot foreign index was accepted: %v\n%s", err, output)
		}
	}
	t.Log("same-slot index from another HIRArena is rejected on Node and native")
}
