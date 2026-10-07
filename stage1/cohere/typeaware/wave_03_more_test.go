package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func wave03MoreBodies(t *testing.T, repository string) []string {
	t.Helper()
	var bodies []string
	for _, name := range []string{"no_throw_literal", "prefer_arrow_callback", "no_useless_backreference", "no_useless_backreference_corpus"} {
		tree, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok || len(literal.Elts) == 0 {
				return true
			}
			for _, e := range literal.Elts {
				kv, ok := e.(*goast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*goast.Ident)
				if !ok || (key.Name != "source" && key.Name != "sourceText" && key.Name != "code") {
					continue
				}
				value, ok := kv.Value.(*goast.BasicLit)
				if ok && value.Kind == token.STRING {
					text, err := strconv.Unquote(value.Value)
					if err != nil {
						t.Fatal(err)
					}
					bodies = append(bodies, text)
					return false
				}
			}
			at := 0
			if name == "no_useless_backreference_corpus" {
				at = 1
			}
			if len(literal.Elts) <= at {
				return true
			}
			value, ok := literal.Elts[at].(*goast.BasicLit)
			if !ok || value.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(value.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(text, "throw") || strings.Contains(text, "function") || strings.Contains(text, "RegExp") || strings.HasPrefix(text, "/") {
				bodies = append(bodies, text)
				return false
			}
			return true
		})
	}
	return bodies
}
func wave03MoreMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/wave_03_more/*.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		// Rewrite the parser imports before rewriting the one-parent imports.
		source = strings.ReplaceAll(source, "../../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "'../", "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "main.a"), archive, false)
}

func TestWave03MoreAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE03_MORE_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("more-checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_03_more/main.a")
	binary := h.build(stage0, "more", entry, archive, false)
	oracle := volumeOracle(h, "more-oracle", "oracle_wave_03_more.go")
	types := h.write("more.d.ts", "export {};\n")
	config := h.write("more-config.json", fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022","DOM"],"moduleDetection":"auto"},"files":[%q]}`, types))
	bodies := append(wave03MoreBodies(t, repository), []string{"throw (new Error());", "throw undefined;", "function f(undefined:Error){throw undefined;}", `RegExp('\\1(a)');`, "foo(function(){ return this.x; });", "foo(function(this:any){ return 1; });", `const p='\\1(a)';const R=globalThis.RegExp;new R(p);`, `const R=flag?RegExp:RegExp;R('\\1(a)');`, `const {RegExp:R}=globalThis;R('\\1(a)');`, `let p='\\1(a)';p='x';RegExp(p);`, "foo(function(){return '世界🌍';});"}...)
	var paths []string
	for i, source := range bodies {
		paths = append(paths, h.write(fmt.Sprintf("more-control-%03d.ts", i), source+"\nexport {};\n"))
	}
	candidates := h.write("more-candidates.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, candidates, "--valid-sources"))
	manifest := h.write("more-controls.manifest", string(valid.stdout))
	t.Logf("controls: %d candidates %d independently parseable", len(paths), len(strings.Fields(string(valid.stdout))))
	truth := h.compare("more-controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-throw-literal", "no-useless-backreference", "prefer-arrow-callback"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	for _, option := range []string{"--allow-named", "--no-unbound-this"} {
		want := h.must(option+"-go", exec.Command(oracle, config, manifest, option))
		got := h.must(option+"-native", exec.Command(binary, config, manifest, option))
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("%s differs at byte %d", option, firstDifference(want.stdout, got.stdout))
		}
	}
	if os.Getenv("ADAMIC_WAVE03_MORE_QUICK") != "" {
		return
	}
	sanitized := h.archive("more-checker-asan", "", true)
	asan := h.build(stage0, "more-asan", entry, sanitized, true)
	h.compare("more-controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{{"throw", "no_throw_literal.a", "if(!this.couldBeError(argument))", "if(this.couldBeError(argument))"}, {"backreference", "regex_structure.a", "if(!backwards && reference.end <= group.start)", "if(backwards && reference.end <= group.start)"}, {"callback", "prefer_arrow_callback.a", "insertion, insertion, ' =>'", "insertion, insertion, ' => '"}} {
		mutant := wave03MoreMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s builds, exits 0, empty stderr, byte comparison catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, population := range []struct{ name, config, manifest string }{{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE03_REPOSITORY_MANIFEST")}, {"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE03_COMPILER_MANIFEST")}} {
		if population.manifest == "" {
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		command := exec.Command(binary, population.config, population.manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("timed %s differs", population.name)
		}
		t.Logf("%s native %s Go %s; native %s Go %s", population.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	probe := h.write("more-released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier','source-access-context'));`)
	source := h.write("more-probe.ts", "x;")
	stale := h.build(stage0, "more-released", probe, archive, false)
	got := h.run("more-released-run", exec.Command(stale, config, source))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released access query: panic 70")
	overlay := h.overlay("more-released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains a released program.")
	mutantArchive := h.archive("more-released-registry", overlay, false)
	mutant := h.build(stage0, "more-released-mutant", probe, mutantArchive, false)
	h.must("more-released-mutant-run", exec.Command(mutant, config, source))
	t.Log("released registry mutant exits 0 and is caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
