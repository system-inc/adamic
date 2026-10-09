package lint

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	nativechecker "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func TestCheckerNoProgramCoverage(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	source := ownedWitnesses(t, directory, "no-unnecessary-boolean-literal-compare")[0]
	path := manifest(t, []string{source + "\t@typescript-eslint/no-unnecessary-boolean-literal-compare"})
	want := []byte("skipped @typescript-eslint/no-unnecessary-boolean-literal-compare no program\n")
	changed := mutant(t, "this.skipped.push(`skipped ${name} no program`);", "this.skipped.push(`skipped ${name} ignored`);", "context.ts")
	before := node(t, changed, path, false).output
	if bytes.Contains(before, want) {
		t.Fatal("no-program coverage omission survived")
	}
	t.Logf("before: coverage check caught mutant: %s", before)
	after := node(t, directory, path, false).output
	if !bytes.Contains(after, want) {
		t.Fatalf("typed selection was silently empty: %s", after)
	}
	t.Logf("after: coverage check passes: %s", after)
	oracle := execute(t, "", goOracle(t), "--manifest", path).output
	if diff := difference(after, oracle); diff != "" {
		t.Fatal(diff)
	}
}

func TestCheckerBridgeRefusalPending(t *testing.T) {
	t.Parallel()
	declarations, err := os.ReadFile(filepath.Join(repository, "internal/load/prelude.d.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(declarations, []byte("TSGoError")) {
		t.Skip("awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer")
	}
	directory := mutant(t, "const strict = frames.yes(); frames.end();", "const strict = frames.yes(); frames.end(); checker.ask(root, 'pending-refusal-control');", "rules/no-unnecessary-boolean-literal-compare/rule.a")
	source := ownedWitnesses(t, directory, "no-unnecessary-boolean-literal-compare")[0]
	project := t.TempDir()
	config := filepath.Join(project, "tsconfig.json")
	options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, source)
	if err := os.WriteFile(config, []byte(options), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{"program " + config, source + "\t@typescript-eslint/no-unnecessary-boolean-literal-compare"})
	prefix := filepath.Join(project, "facts")
	binary := buildPort(t, directory, true)
	native := execute(t, "", binary, "--manifest", path, "--record", prefix).output
	if !bytes.Contains(native, []byte("refused @typescript-eslint/no-unnecessary-boolean-literal-compare ")) {
		t.Fatalf("bridge error lost: %s", native)
	}
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	replay := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--replay", prefix).output
	if diff := difference(replay, native); diff != "" {
		t.Fatal(diff)
	}
	emitted := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emittedJavaScript(t, directory), "--manifest", path, "--replay", prefix).output
	if diff := difference(emitted, native); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("bridge refusal matches native, Node and emitted JavaScript: %s", native)
	checkerPath := filepath.Join(directory, "checker.a")
	original, err := os.ReadFile(checkerPath)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "this.refusals.push(new Refusal(this.rule, start, end, reason));"
	if strings.Count(string(original), anchor) != 1 {
		t.Fatal("refusal mutant anchor not unique")
	}
	changed := strings.Replace(string(original), anchor, "", 1)
	if err := os.WriteFile(checkerPath, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	mutantPrefix := filepath.Join(project, "mutant-facts")
	mutantNative := execute(t, "", buildPort(t, directory, true), "--manifest", path, "--record", mutantPrefix).output
	mutantNode := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--replay", mutantPrefix).output
	mutantEmitted := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emittedJavaScript(t, directory), "--manifest", path, "--replay", mutantPrefix).output
	for runtime, got := range map[string][]byte{"native": mutantNative, "Node": mutantNode, "emitted JavaScript": mutantEmitted} {
		if bytes.Equal(got, native) || bytes.Contains(got, []byte("refused ")) {
			t.Fatalf("%s refusal handling removal survived: %s", runtime, got)
		}
		t.Logf("%s refusal handling mutant builds and runs, caught by wire comparison", runtime)
	}
	if err := os.WriteFile(checkerPath, original, 0644); err != nil {
		t.Fatal(err)
	}
	for runtime, got := range map[string][]byte{
		"native":             execute(t, "", binary, "--manifest", path, "--record", prefix).output,
		"Node":               execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--replay", prefix).output,
		"emitted JavaScript": execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emittedJavaScript(t, directory), "--manifest", path, "--replay", prefix).output,
	} {
		if diff := difference(got, native); diff != "" {
			t.Fatalf("restored %s: %s", runtime, diff)
		}
	}
	t.Log("restoring refusal handling passes all three runtime wire comparisons")
}

// Replay is checked against live bridge facts. When the result API is present,
// native Adamic also produces and verifies this exact transcript.
func TestCheckerReplayEntryControl(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	source := filepath.Join(project, "source.ts")
	text := "declare const b: boolean; if (b === true) {}\n"
	if os.Getenv("ADAMIC_LINT_CONSTRAINT_CONTROL") == "1" {
		text = "function f<T extends boolean>(b: T) { return b === true; }\n"
	}
	config := filepath.Join(project, "tsconfig.json")
	options := `{"compilerOptions":{"strict":true},"files":["source.ts"]}`
	if err := os.WriteFile(source, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(options), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := nativechecker.Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(source)))
	var operand *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindBinaryExpression {
			operand = node.AsBinaryExpression().Left
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	if operand == nil {
		t.Fatal("no pilot operand")
	}
	frame := func(text string) string { return fmt.Sprintf("%d\n%s", len(utf16.Encode([]rune(text))), text) }
	entry := func(node *ast.Node, question string) string {
		kind := strings.TrimPrefix(node.Kind.String(), "Kind")
		value, err := program.Inspect(source, uint64(node.Pos()), uint64(node.End()), kind, question)
		if err != nil {
			t.Fatal(err)
		}
		key := frame(source) + frame(fmt.Sprint(node.Pos())) + frame(fmt.Sprint(node.End())) + frame(kind) + frame(question)
		return frame(key) + frame("Value") + frame(value)
	}
	first := entry(file.AsNode(), "options")
	second := entry(operand, "type-shape")
	prefix := filepath.Join(project, "facts")
	digest := func(text string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(text))) }
	header := "checker transcript 1\nprogram " + config + "\nsha256 " + digest(options) + "\n"
	rowHeader := source + "\n" + digest(text) + "\n"
	if err := os.WriteFile(prefix+".header", []byte(header), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{"program " + config, source + "\t@typescript-eslint/no-unnecessary-boolean-literal-compare"})
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	run := func(port, records string) []byte {
		if err := os.WriteFile(prefix+".0", []byte(rowHeader+records), 0644); err != nil {
			t.Fatal(err)
		}
		return execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(port, "main.ts"), "--manifest", path, "--replay", prefix).output
	}
	want := execute(t, "", goOracle(t), "--manifest", path).output
	declarations, err := os.ReadFile(filepath.Join(repository, "internal/load/prelude.d.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(declarations, []byte("TSGoError")) {
		live := execute(t, "", buildPort(t, directory, true), "--manifest", path, "--record", prefix).output
		if diff := difference(live, want); diff != "" {
			t.Fatal(diff)
		}
		actualHeader, err := os.ReadFile(prefix + ".header")
		if err != nil {
			t.Fatal(err)
		}
		if string(actualHeader) != header {
			t.Fatalf("native program header differs from Go SHA-256: %s", actualHeader)
		}
		transcript, err := os.ReadFile(prefix + ".0")
		if err != nil {
			t.Fatal(err)
		}
		if string(transcript) != rowHeader+first+second {
			t.Fatalf("native transcript differs from live bridge: %s", transcript)
		}
		t.Log("native recording verified against live bridge and Go findings")
	}
	before := run(directory, first)
	if bytes.Equal(before, want) || !bytes.Contains(before, []byte("missing transcript entry")) {
		t.Fatalf("removed entry survived: %s", before)
	}
	t.Logf("before: removed entry fails Node comparison: %s", before)
	after := run(directory, first+second)
	if diff := difference(after, want); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("after: complete live-bridge transcript matches Go: %s", after)
	extra := run(directory, first+second+second)
	if bytes.Equal(extra, want) || !bytes.Contains(extra, []byte("unasked transcript entry")) {
		t.Fatalf("extra entry survived: %s", extra)
	}
	t.Log("reverse guard rejects an unasked entry")
	if err := os.WriteFile(prefix+".header", []byte(header+"changed program hash\n"), 0644); err != nil {
		t.Fatal(err)
	}
	wrongProgram := run(directory, first+second)
	if bytes.Equal(wrongProgram, want) || !bytes.Contains(wrongProgram, []byte("transcript program differs")) {
		t.Fatalf("changed program header survived: %s", wrongProgram)
	}
	if err := os.WriteFile(prefix+".header", []byte(header), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("program header mismatch refuses replay; restoring it passes")
	if diff := difference(run(directory, first+second), want); diff != "" {
		t.Fatal(diff)
	}
	typeMutant := mutant(t, "(!plain && !nullable)", "(plain && !nullable)", "rules/no-unnecessary-boolean-literal-compare/rule.a")
	if bytes.Equal(run(typeMutant, first+second), want) {
		t.Fatal("pilot type-verdict mutant survived Node comparison")
	}
	t.Log("pilot type-verdict mutant caught only by findings comparison on Node")

	changed := mutant(t, "const strict = frames.yes(); frames.end();", "const strict = frames.yes(); frames.end(); checker.askFile(new FileQuestion('ReadsOtherFiles', 'scope-locals'));", "rules/no-unnecessary-boolean-literal-compare/rule.a")
	undeclared := run(changed, first+second)
	if !bytes.Contains(undeclared, []byte("undeclared program read ReadsOtherFiles")) {
		t.Fatalf("undeclared read survived: %s", undeclared)
	}
	t.Logf("pilot's harness-owned refusal reaches the wire: %s", undeclared)
}

func TestCheckerHashes(t *testing.T) {
	t.Parallel()
	module, err := filepath.Abs("checker_hash.a")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "hashes.a")
	text := "import { hash } from '" + filepath.ToSlash(module) + "';\nconsole.log(hash('')); console.log(hash('abc')); console.log(hash('😀'));\n"
	if err := os.WriteFile(source, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	want := []byte(fmt.Sprintf("%x\n%x\n%x\n", sha256.Sum256(nil), sha256.Sum256([]byte("abc")), sha256.Sum256([]byte("😀"))))
	got := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, source).output
	if !bytes.Equal(got, want) {
		t.Fatalf("configuration hash differs: %s", got)
	}
	native := execute(t, "", buildNative(t, source, true)).output
	if !bytes.Equal(native, want) {
		t.Fatalf("native configuration hash differs: %s", native)
	}
}
