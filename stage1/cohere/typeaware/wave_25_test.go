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

// Extract the unchanged production table sources, including the GraphQL preamble.
func wave25Controls(h *harness) []string {
	var sources []string
	for _, name := range []string{"graphql_nullable_parity", "matching_inject_type"} {
		file := filepath.Join(h.repository, "cohere/internal/lint/rules/base/correctness_require_"+name+"_test.go")
		tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		preamble := ""
		var literalText func(goast.Expr) string
		literalText = func(e goast.Expr) string {
			switch n := e.(type) {
			case *goast.BasicLit:
				if n.Kind == token.STRING {
					text, err := strconv.Unquote(n.Value)
					if err != nil {
						h.t.Fatal(err)
					}
					return text
				}
			case *goast.BinaryExpr:
				if n.Op == token.ADD {
					return literalText(n.X) + literalText(n.Y)
				}
			}
			return ""
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			if v, ok := n.(*goast.ValueSpec); ok && len(v.Names) > 0 && v.Names[0].Name == "graphQlNullableParityPreamble" {
				preamble = literalText(v.Values[0])
				return false
			}
			return true
		})
		seen := map[string]bool{}
		goast.Inspect(tree, func(n goast.Node) bool {
			v, ok := n.(*goast.BasicLit)
			if !ok || v.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(v.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			if strings.Contains(text, "class ") && strings.Contains(text, "@") && !seen[text] {
				seen[text] = true
				sources = append(sources, preamble+text)
			}
			return true
		})
	}
	if len(sources) < 50 {
		h.t.Fatalf("missing production controls: %d", len(sources))
	}
	return sources
}

// Not parallel: native builds, sanitizer runs and timings share this machine.
func TestWave25AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE25_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_25_suite.a")
	binary := h.build(stage0, "wave25", entry, archive, false)
	oracle := volumeOracle(h, "wave25-oracle", "oracle_wave_25.go")
	config := h.write("controls-config.json", fmt.Sprintf(`{"extends":%q,"compilerOptions":{"experimentalDecorators":true}}`, filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")))
	var paths []string
	for i, source := range wave25Controls(h) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	// Additional Unicode/CRLF and optional union-brand controls.
	paths = append(paths, h.write("edge.a", `type D<T> = ParameterDecorator & {readonly __resolvedType?:T};
declare function Inject<T>():D<T>;
declare const namespace:{bare:D<string>};
class C {constructor(@Inject<string|number>() value:string|number, @Inject<undefined>() absent:number, @namespace.bare bare:number) {}}
`))
	paths = append(paths, h.write("unicode.a", "/* 世界 🌍 */\r\ndeclare function GraphQlField(o?:unknown):PropertyDecorator;\r\nclass C { @GraphQlField({nullable:true}) é!:string; }\r\nexport {};\r\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"base/correctness-require-graphql-nullable-parity", "base/correctness-require-matching-inject-type"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave25-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"graphql-mask", "graphql_nullable_parity.a", "graph.has(actual, 1 | 2 | 4 | 8 | 16)", "graph.has(actual, 1 | 2 | 8 | 16)"},
		{"inject-strip", "matching_inject_type.a", "if(kept.length === 1)", "if(kept.length === 0)"},
	} {
		mutantDir := filepath.Join(directory, m.name)
		if err := os.MkdirAll(mutantDir, 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"wave_25_suite.a", "decorator_shape.a", "graphql_nullable_parity.a", "matching_inject_type.a"} {
			data, err := os.ReadFile(filepath.Join(filepath.Dir(entry), file))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if file == m.file {
				if strings.Count(source, m.from) != 1 {
					t.Fatal("nonunique mutant")
				}
				source = strings.Replace(source, m.from, m.to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			for _, dependency := range []string{"unary_minus", "rules", "checker_facts", "facts", "frames", "diagnostic"} {
				source = strings.ReplaceAll(source, "./"+dependency+".ts", filepath.Join(repository, "stage1/cohere/typeaware", dependency+".ts"))
			}
			if err := os.WriteFile(filepath.Join(mutantDir, file), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, m.name+"-native", filepath.Join(mutantDir, "wave_25_suite.a"), archive, false)
		got := h.must(m.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant survived: %s", m.name)
		}
		t.Logf("%s exits 0; only Go byte oracle catches byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, root, config string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json")},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")},
	} {
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
		corpusManifest := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, corpusManifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, corpusManifest)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, corpusManifest, "--count"))
		command := exec.Command(binary, population.config, corpusManifest, "--count")
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("timing count mismatch")
		}
		t.Logf("%s whole process native %s Go %s; %s", population.name, got.elapsed, want.elapsed, got.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);
console.log(tsgoInspect(program,path,0,1,'Identifier','raw-shape'));`)
	stale := h.build(stage0, "released", released, archive, false)
	probe := h.write("released-probe.a", "x;\n")
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released-handle check: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released program.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-registry-native", released, mutantArchive, false)
	h.must("released-registry-run", exec.Command(mutant, config, probe))
	t.Log("released registry mutant exits 0; required panic 70 catches it")
}
