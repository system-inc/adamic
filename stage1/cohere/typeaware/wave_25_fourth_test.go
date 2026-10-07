package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func wave25FourthControls(h *harness) []string {
	var sources []string
	unquote := func(e goast.Expr) string {
		if n, ok := e.(*goast.BasicLit); ok && n.Kind == token.STRING {
			v, err := strconv.Unquote(n.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			return v
		}
		return ""
	}
	for _, name := range []string{"no_throw_literal_test.go", "no_useless_backreference_test.go", "no_useless_backreference_corpus_test.go", "prefer_arrow_callback_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/core", name), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			if call, ok := n.(*goast.CallExpr); ok {
				if f, ok := call.Fun.(*goast.SelectorExpr); ok && strings.HasPrefix(f.Sel.Name, "RunTyped") && len(call.Args) >= 4 {
					if value := unquote(call.Args[3]); value != "" {
						sources = append(sources, value)
					}
				}
			}
			if list, ok := n.(*goast.CompositeLit); ok && list.Type == nil && len(list.Elts) > 0 {
				keyed := false
				for _, item := range list.Elts {
					if pair, ok := item.(*goast.KeyValueExpr); ok {
						keyed = true
						if key, ok := pair.Key.(*goast.Ident); ok && (key.Name == "source" || key.Name == "sourceText" || key.Name == "code") {
							if v := unquote(pair.Value); v != "" {
								sources = append(sources, v)
							}
						}
					}
				}
				if !keyed {
					at := 0
					if name == "no_useless_backreference_corpus_test.go" || unquote(list.Elts[0]) == "an unclosed group" || unquote(list.Elts[0]) == "an unopened group" || unquote(list.Elts[0]) == "a trailing backslash" || unquote(list.Elts[0]) == "an unterminated class" {
						at = 1
					}
					if at < len(list.Elts) {
						if v := unquote(list.Elts[at]); v != "" {
							sources = append(sources, v)
						}
					}
				}
			}
			return true
		})
	}
	sources = append(sources, `const {RegExp: R} = globalThis; R('\\1(a)');`, `let R; R = RegExp; R('\\1(a)');`, `RegExp = other; RegExp('\\1(a)');`, `const p = '\\1'; RegExp(p + '(a)');`, `throw flag ? new Error() : new Error();`, "/* 世界 🌍 */\r\nfoo(function café(a) { return a; });\r\n")
	sources = append(sources,
		"const suffix = '(a)'; RegExp(`\\\\1${suffix}`);",
		`const G = globalThis; const R = G['RegExp']; R('\\1(a)');`,
		`let R; ({RegExp: R} = globalThis); R('\\1(a)');`,
		`function f(R = RegExp) { R('\\1(a)'); }`,
		`let R; const o = {R = RegExp}; R('\\1(a)');`,
		`const R = RegExp; const o = {R}; export {R as Exported}; R('\\1(a)');`,
		`let p = '\\1(a)'; [p] = ['(a)']; RegExp(p);`,
		`let p = '\\1(a)'; ({x:p} = obj); RegExp(p);`,
		`const R = RegExp as typeof RegExp; R('\\1(a)');`,
		`const R = flag ? RegExp : RegExp; R('\\1(a)');`,
		`const R = RegExp; R('\\1(a)'); globalThis.RegExp('\\1(a)');`,
		`const p = p + '(a)'; RegExp(p);`,
		`const R = RegExp; const recursive = recursive; R('\\1(a)');`,
		`foo(function /* preserve */ café(a) { return a; });`,
		"foo(async function\n(a) { return a; });",
	)

	if err := os.MkdirAll(filepath.Join(h.directory, "controls"), 0755); err != nil {
		h.t.Fatal(err)
	}
	seen := map[string]bool{}
	var paths []string
	for _, source := range sources {
		if seen[source] {
			continue
		}
		seen[source] = true
		physical := h.write(fmt.Sprintf("controls/control-%03d.a", len(paths)), source+"\nexport {};\n")
		logical := strings.TrimSuffix(physical, ".a") + ".ts"
		if _, err := os.Lstat(logical); os.IsNotExist(err) {
			if err := os.Symlink(filepath.Base(physical), logical); err != nil {
				h.t.Fatal(err)
			}
		}
		paths = append(paths, logical)
	}
	if len(paths) < 150 {
		h.t.Fatalf("missing controls: %d", len(paths))
	}
	h.t.Logf("%d distinct production and edge inputs, default options", len(paths))
	return paths
}

// Not parallel: native builds, sanitizers and timings share the machine.
func TestWave25FourthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE25_FOURTH_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_25_fourth_suite.a")
	binary := h.build(stage0, "wave25-fourth", entry, archive, false)
	oracle := volumeOracle(h, "wave25-fourth-oracle", "oracle_wave_25_fourth.go")
	config := h.write("controls-config.json", fmt.Sprintf(`{"extends":%q,"compilerOptions":{"target":"ES2024","lib":["ES2024","DOM"],"moduleDetection":"auto"},"include":["controls/**/*.ts"]}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")))
	paths := wave25FourthControls(h)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("source-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	t.Logf("excluded %d parse-invalid inputs", len(paths)-strings.Count(string(valid.stdout), "\n"))
	manifest = h.write("controls-valid.manifest", string(valid.stdout))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-throw-literal", "no-useless-backreference", "prefer-arrow-callback"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave25-fourth-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"throw-conditional", "no_throw_literal.a", "node.kind === 'ConditionalExpression'", "false"},
		{"regex-forward", "no_useless_backreference.a", "return 2;", "return 3;"},
		{"arrow-fix", "prefer_arrow_callback.a", "' =>'", "' ->'"},
	} {
		mutantDirectory := filepath.Join(directory, mutation.name)
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"wave_25_fourth_suite.a", "no_throw_literal.a", "no_useless_backreference.a", "prefer_arrow_callback.a", "regexp_references.a", "source_bytes.a", "syntax_reference_facts.a", "syntax_projection.a", "process_symbol_details.a"} {
			data, err := os.ReadFile(filepath.Join(filepath.Dir(entry), file))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if file == mutation.file {
				if strings.Count(source, mutation.from) != 1 {
					t.Fatal("nonunique mutant", mutation.name)
				}
				source = strings.Replace(source, mutation.from, mutation.to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			for _, dependency := range []string{"unary_minus", "rules", "facts", "frames", "diagnostic", "type_fact", "repair"} {
				source = strings.ReplaceAll(source, "./"+dependency+".ts", filepath.Join(repository, "stage1/cohere/typeaware", dependency+".ts"))
			}
			if err := os.WriteFile(filepath.Join(mutantDirectory, file), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, mutation.name+"-native", filepath.Join(mutantDirectory, "wave_25_fourth_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", mutation.name)
		}
		t.Logf("%s exits 0; only Go byte comparison catches byte %d", mutation.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, root, config string }{{"repository", repository, filepath.Join(repository, "tsconfig.json")}, {"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")}} {
		if population.root == "" {
			t.Log("compiler corpus not supplied")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			roots = append(roots, filepath.Join(population.root, path))
		}
		corpus := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, corpus)
		h.compare(population.name+"-asan", oracle, asan, population.config, corpus)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, corpus, "--count"))
		got := h.must(population.name+"-timed-native", exec.Command(binary, population.config, corpus, "--count"))
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("timing count mismatch")
		}
		t.Logf("%s native %s Go %s", population.name, got.elapsed, want.elapsed)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic'; const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,3,'SourceFile','syntax-reference-facts'));`)
	stale := h.build(stage0, "released", released, archive, false)
	probe := h.write("released-probe.a", "x;\n")
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released syntax question: panic 70, invalid or released checker handle")
}
