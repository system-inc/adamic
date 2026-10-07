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

func wave25FifthControls(h *harness) []string {
	var sources []string
	values := map[string]string{}
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
	for _, name := range []string{"unsupported_syntax_test.go", "use_memo_test.go", "boolean_prop_naming_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/react", name), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			if v, ok := n.(*goast.ValueSpec); ok {
				for at, name := range v.Names {
					if at < len(v.Values) {
						values[name.Name] = unquote(v.Values[at])
					}
				}
			}
			return true
		})
		goast.Inspect(tree, func(n goast.Node) bool {
			if call, ok := n.(*goast.CallExpr); ok {
				if f, ok := call.Fun.(*goast.SelectorExpr); ok && strings.HasPrefix(f.Sel.Name, "RunTyped") && len(call.Args) >= 4 {
					value := unquote(call.Args[3])
					if name, ok := call.Args[3].(*goast.Ident); ok {
						value = values[name.Name]
					}
					if value != "" {
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
					at := 1
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
	sources = append(sources, `function Component(p) { eval('x'); with(p){} class Inner {} return <div/>; }`, `function Component(p) { useMemo(p.fn, [f()]); return <div/>; }`, `interface Props { enabled: boolean; isFine: boolean } function Component(p: Props) { return <div/>; }`)

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
		logical := strings.TrimSuffix(physical, ".a") + ".tsx"
		if _, err := os.Lstat(logical); os.IsNotExist(err) {
			if err := os.Symlink(filepath.Base(physical), logical); err != nil {
				h.t.Fatal(err)
			}
		}
		paths = append(paths, logical)
	}
	if len(paths) < 100 {
		h.t.Fatalf("missing controls: %d", len(paths))
	}
	h.t.Logf("%d distinct production and edge inputs, boolean options object with defaults", len(paths))
	return paths
}

func wave25FifthCompare(h *harness, name, oracle, binary, config, manifest string) result {
	want := h.must(name+"-go", exec.Command(oracle, config, manifest, "--boolean-defaults"))
	got := h.must(name+"-native", exec.Command(binary, config, manifest, "--boolean-defaults"))
	if len(got.stderr) != 0 {
		h.t.Fatalf("native stderr: %s", got.stderr)
	}
	if !bytes.Equal(got.stdout, want.stdout) {
		at := firstDifference(got.stdout, want.stdout)
		h.t.Fatalf("%s mismatch byte %d: native %q Go %q", name, at, got.stdout[max(0, at-80):min(len(got.stdout), at+250)], want.stdout[max(0, at-80):min(len(want.stdout), at+250)])
	}
	h.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}

// Not parallel: native builds, sanitizers and timings share the machine.
func TestWave25FifthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE25_FIFTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_25_fifth_suite.a")
	binary := h.build(stage0, "wave25-fifth", entry, archive, false)
	oracle := volumeOracle(h, "wave25-fifth-oracle", "oracle_wave_25_fifth.go")
	config := h.write("controls-config.json", fmt.Sprintf(`{"extends":%q,"compilerOptions":{"target":"ES2024","lib":["ES2024","DOM"],"moduleDetection":"auto","jsx":"react-jsx"},"include":["controls/**/*.tsx"]}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")))
	paths := wave25FifthControls(h)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("source-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	t.Logf("excluded %d parse-invalid inputs", len(paths)-strings.Count(string(valid.stdout), "\n"))
	manifest = h.write("controls-valid.manifest", string(valid.stdout))
	h.compare("controls-no-options", oracle, binary, config, manifest)
	truth := wave25FifthCompare(h, "controls", oracle, binary, config, manifest)
	for _, name := range []string{"react-hooks/unsupported-syntax", "react-hooks/use-memo", "react/boolean-prop-naming"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave25-fifth-asan", entry, sanitized, true)
	wave25FifthCompare(h, "controls-asan", oracle, asan, config, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"eval-library", "unsupported_syntax.a", "declaration.library", "false"},
		{"memo-inline", "use_memo.a", "if(!inline) {", "if(inline) {"},
		{"boolean-capital", "boolean_prop_naming.a", "/^(is|has)[A-Z]([A-Za-z0-9]?)+/u", "/^(is|has)[a-z]([A-Za-z0-9]?)+/u"},
	} {
		mutantDirectory := filepath.Join(directory, mutation.name)
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"wave_25_fifth_suite.a", "unsupported_syntax.a", "use_memo.a", "boolean_prop_naming.a", "react_compilation_gate.a", "react_wave_25_messages.a", "declared_syntax_type.a", "syntax_projection.a", "process_symbol_details.a"} {
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
		mutant := h.build(stage0, mutation.name+"-native", filepath.Join(mutantDirectory, "wave_25_fifth_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest, "--boolean-defaults"))
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
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic'; const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,3,'SourceFile','declared-syntax-type'));`)
	stale := h.build(stage0, "released", released, archive, false)
	probe := h.write("released-probe.a", "x;\n")
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released syntax question: panic 70, invalid or released checker handle")
}

// Not parallel: native builds and sanitizer runs share this machine.
func TestWave25FifthBooleanRefusals(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE25_FIFTH_REFUSALS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_25_fifth_suite.a")
	data, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(data), "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
	for _, name := range []string{"unary_minus.ts", "rules.ts", "syntax_projection.a", "unsupported_syntax.a", "use_memo.a", "boolean_prop_naming.a"} {
		source = strings.ReplaceAll(source, "./"+name, filepath.Join(filepath.Dir(entry), name))
	}
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2024","jsx":"react-jsx"},"files":["input.tsx","props.d.ts"]}`)
	physical := h.write("input.a", "export {};\n")
	if err := os.Symlink(filepath.Base(physical), filepath.Join(directory, "input.tsx")); err != nil {
		t.Fatal(err)
	}
	declarations := h.write("props-declarations.a", "interface GlobalProps { enabled: boolean }\n")
	if err := os.Symlink(filepath.Base(declarations), filepath.Join(directory, "props.d.ts")); err != nil {
		t.Fatal(err)
	}
	manifest := h.write("controls.manifest", filepath.Join(directory, "input.tsx")+"\n")
	for _, row := range []struct{ name, pattern, input, reason string }{
		{"custom-pattern", "^enabled$", "export {};\n", "boolean-prop-naming requires a native regular-expression engine for this configured pattern"},
		{"foreign-annotation", "", "function C(p: GlobalProps) { return <div/>; } export {};\n", "boolean-prop-naming imported props syntax requires a cross-file annotation projection"},
		{"typed-wrapper", "", "interface Props {enabled:boolean} const C: React.FC<Props> = React.memo((p) => <div/>); export {};\n", "boolean-prop-naming typed wrapper component detection is not yet ported"},
	} {
		own := source
		if row.pattern != "" {
			own = strings.Replace(own, "args.includes('--boolean-defaults')", "true, '"+row.pattern+"'", 1)
		}
		path := h.write(row.name+".a", own)
		binary := h.build(stage0, row.name, path, archive, false)
		if err := os.WriteFile(physical, []byte(row.input), 0644); err != nil {
			t.Fatal(err)
		}
		got := h.run(row.name+"-run", exec.Command(binary, config, manifest, "--boolean-defaults"))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || len(got.stdout) != 0 || string(got.stderr) != "adamic: panic: "+row.reason+"\n" {
			t.Fatalf("%s did not refuse before findings: %v %q %q", row.name, got.err, got.stdout, got.stderr)
		}
		t.Logf("%s: exit 70, no partial findings", row.name)
	}
}
