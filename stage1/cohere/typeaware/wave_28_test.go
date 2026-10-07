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

// Extract source rows from the pinned production tests rather than inventing
// equivalents. Default options are judged by the independent production oracle.
func wave28Controls(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, item := range []struct {
		path   string
		column int
	}{
		{"core/block_scoped_var_test.go", 0}, {"core/getter_return_test.go", 1},
		{"base/correctness_require_verify_optional_parity_test.go", -1},
	} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", item.path), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			literal, ok := n.(*goast.CompositeLit)
			if !ok {
				return true
			}
			// Getter clean cases also use a plain string slice rather than tuples.
			if item.path == "core/getter_return_test.go" {
				if array, ok := literal.Type.(*goast.ArrayType); ok {
					if element, ok := array.Elt.(*goast.Ident); ok && element.Name == "string" {
						for _, row := range literal.Elts {
							if value, ok := row.(*goast.BasicLit); ok && value.Kind == token.STRING {
								source, err := strconv.Unquote(value.Value)
								if err != nil {
									t.Fatal(err)
								}
								if strings.ContainsAny(source, "{};") {
									sources = append(sources, source)
								}
							}
						}
						return true
					}
				}
			}
			var value *goast.BasicLit
			if item.column >= 0 && len(literal.Elts) > item.column {
				value, _ = literal.Elts[item.column].(*goast.BasicLit)
			} else if item.column < 0 {
				for _, element := range literal.Elts {
					pair, ok := element.(*goast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := pair.Key.(*goast.Ident)
					if ok && key.Name == "source" {
						value, _ = pair.Value.(*goast.BasicLit)
					}
				}
			}
			if value == nil || value.Kind != token.STRING {
				return true
			}
			source, err := strconv.Unquote(value.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.ContainsAny(source, "{};") {
				return true
			}
			if item.column < 0 {
				source = "declare function VerifyIsOptional(): PropertyDecorator & ParameterDecorator;\ndeclare function VerifyIsString(): PropertyDecorator & ParameterDecorator;\ndeclare function VerifyBy(check: unknown): PropertyDecorator & ParameterDecorator;\ndeclare function NotAVerifyDecorator(): PropertyDecorator;\ndeclare const namespaced: { Verify(): PropertyDecorator };\n" + source
			}
			sources = append(sources, source)
			return true
		})
	}
	// Additional byte-span, binding identity and type-graph discriminations.
	sources = append(sources,
		"/* 世界 🌍 */\r\nfunction f(){\r\n{var é=1;}\r\né;}\r\n",
		"function f(){ {var {x:y=1, z:{w}}=obj;} y;w;} function g(){ {var [a,,...b]=rows;} a;b;}",
		"{var x;} const holder={x}; holder.x; type X=typeof x;",
		"import {type T} from 'missing'; function f(){ {var type=1;} type; }",
		"/* 世界 🌍 */\r\nObject.defineProperty({},'x',{get: ( /* parameter */ p) => {if(p)return 1;}});",
		"Object.defineProperty({},'x',{get: p => {}}); Object.defineProperty({},'x',{get: <T extends () => number>() => {}});",
		"Object.defineProperty({},'x',{['get'](){}}); Object.defineProperties({}, {x:{[`get`]:function(){}}});",
		"Object['defineProperty']({},'x',{get(){}}); (Reflect?.['defineProperty'])({},'x',({get(){}}));",
		"const Object={defineProperty(...x:unknown[]){}}; Object.defineProperty({},'x',{get(){}});",
		"declare function VerifyIsOptional():PropertyDecorator;declare function VerifyIsString():PropertyDecorator;class Entity<T extends string|undefined>{@VerifyIsString() value:T; @VerifyIsOptional() strict:T; @VerifyIsString() nothing:void; @VerifyIsString() never:never; @VerifyIsString() both:(string|undefined)&{};}",
	)
	return sources
}

func wave28Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	folder := filepath.Join(h.repository, "stage1/cohere/typeaware")
	data, err := os.ReadFile(filepath.Join(folder, file))
	if err != nil {
		h.t.Fatal(err)
	}
	if strings.Count(string(data), from) != 1 {
		h.t.Fatalf("nonunique %s mutant", name)
	}
	source := strings.Replace(string(data), from, to, 1)
	source = wave28Imports(source, folder)
	changed := h.write(name+".a", source)
	entry, err := os.ReadFile(filepath.Join(folder, "wave_28_suite.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	runner := wave28Imports(string(entry), folder)
	runner = strings.Replace(runner, filepath.Join(folder, file), changed, 1)
	return h.build(stage0, name, h.write(name+"-suite.a", runner), archive, false)
}
func wave28Imports(source, folder string) string {
	// Rewrite every relative source import, including the existing .ts modules.
	for _, prefix := range []string{"./", "../../typescript/"} {
		for {
			at := strings.Index(source, "from '"+prefix)
			if at < 0 {
				break
			}
			begin := at + len("from '")
			end := begin + strings.Index(source[begin:], "'")
			absolute := filepath.Clean(filepath.Join(folder, source[begin:end]))
			source = source[:begin] + absolute + source[end:]
		}
	}
	return source
}

// Not parallel: native compilation, sanitizers and cost observations share the machine.
func TestWave28AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE28_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_28_suite.a")
	binary := h.build(stage0, "wave28", entry, archive, false)
	oracle := volumeOracle(h, "wave28-oracle", "oracle_wave_28.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave28Controls(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	accepted := strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")
	if len(accepted) == 0 || len(accepted) > len(paths) {
		t.Fatal("no valid controls")
	}
	t.Logf("production test sources and extra controls: %d accepted, %d parse-invalid excluded", len(accepted), len(paths)-len(accepted))
	manifest = h.write("valid-controls.manifest", string(valid.stdout))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"base/correctness-require-verify-optional-parity", "block-scoped-var", "getter-return"} {
		count := bytes.Count(truth.stdout, []byte("\t"+name+"\t"))
		if count == 0 {
			t.Fatalf("no positive control for %s", name)
		}
		t.Logf("%s: %d control findings", name, count)
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave28-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"parity", "verify_optional_parity.a", "facts.has(facts.root(), 1 | 2 | 4 | 8 | 16)", "(facts.root().flags & (1 | 2 | 4 | 8 | 16)) !== 0"},
		{"scope", "block_scoped_var.a", "if(use.pos >= bounds.pos && use.end <= bounds.end)", "if(use.pos >= 0 && use.end <= this.rules.scanner.text.length)"},
		{"getter", "getter_return.a", "this.exits(node.children[1] ?? -1) && this.exits(node.children[2] ?? -1)", "this.exits(node.children[1] ?? -1) || this.exits(node.children[2] ?? -1)"},
	} {
		mutant := wave28Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, population := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE28_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE28_COMPILER_MANIFEST")},
	} {
		if population.manifest == "" {
			t.Logf("%s corpus not supplied", population.name)
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		for _, subject := range []struct{ name, path string }{{"native", binary}, {"Go", oracle}} {
			command := exec.Command(subject.path, population.config, population.manifest)
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			observed := h.must(population.name+"-timed-"+subject.name, command)
			t.Logf("%s %s process_s=%.6f %s %s", population.name, subject.name, observed.elapsed.Seconds(), summary(observed.stdout), strings.TrimSpace(string(observed.stderr)))
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','symbol-identities'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant: exit 0, caught by required panic 70")
}
