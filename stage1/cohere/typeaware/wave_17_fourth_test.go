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
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Extract source inputs only, with production Go deciding every finding and fix.
func wave17FourthControls(h *harness) []string {
	unique := map[string]bool{}
	for _, name := range []string{"no_throw_literal", "no_useless_backreference", "prefer_arrow_callback"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			if lit, ok := n.(*goast.BasicLit); ok && lit.Kind == token.STRING {
				s, err := strconv.Unquote(lit.Value)
				if err == nil && (strings.Contains(s, "throw ") || strings.Contains(s, "function") || strings.Contains(s, "RegExp(") || (strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "/repository/"))) {
					unique[s] = true
				}
			}
			return true
		})
	}
	for _, s := range []string{
		"/* 世界 🌍 */\r\nthrow undefined;",
		"RegExp('\\\\18446744073709551617(a)','u');RegExp('\\\\18446744073709551616(a)','u');RegExp('\\\\1(🌍)');",
		"throw foo ? new Error() : 'x';throw foo ? 'x' : 'y';throw new Error(),1;",
		"const R=RegExp;new R('\\\\1(a)');let A;A=RegExp;A('(a\\\\1)');",
		"const {RegExp:R}=globalThis;R('\\\\1(a)');const G=globalThis;G['Reg'+'Exp']('(a\\\\1)');",
		"const p='\\\\1(a)';let f='u';RegExp(p,f);f='';RegExp(p,f);",
		"/* 世界 🌍 */\r\nfoo(function /* preserve */ named(x:string):number {return 1;});",
		"foo(function self(){const self=1;self;});foo(function self(){self();});",
		"foo(function(){(()=>arguments)();});foo(function(){function inner(){arguments;}});",
		"new Foo(function(){});foo((function(){return this;}).bind(this));foo(function(){return this;}.bind(this));",
		"foo(function(){}.bind(/* preserve */this));foo(async function\n(x){});",
		"foo(function(a,a){});foo(function(this:any,x:number){});",
	} {
		unique[s] = true
	}
	keys := []string{}
	for s := range unique {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	paths := []string{}
	for i, s := range keys {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), s+"\nexport {};\n"))
	}
	physical := h.write("declaration-control.a", "export {};declare const pattern='\\\\1(a)';RegExp(pattern);\n")
	alias := strings.TrimSuffix(physical, ".a") + ".d.ts"
	os.Remove(alias)
	if err := os.Symlink(filepath.Base(physical), alias); err != nil {
		h.t.Fatal(err)
	}
	paths = append(paths, alias)
	return paths
}

// Not parallel: native and sanitizer builds share the machine's compilation budget.
func TestWave17FourthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE17_FOURTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_fourth_suite.a")
	binary := h.build(stage0, "wave17-fourth", entry, archive, false)
	oracle := volumeOracle(h, "wave17-fourth-oracle", "oracle_wave_17_fourth.go")
	h.write("config-root.d.ts", "export {};\n")
	paths := wave17FourthControls(h)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","moduleDetection":"auto","allowJs":true,"lib":["ESNext"],"types":[],"noEmit":true},"files":["config-root.d.ts"]}`)
	manifest := h.write("all.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-sources", exec.Command(oracle, config, manifest, "--valid-sources"))
	paths = strings.Fields(string(valid.stdout))
	manifest = h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, id := range []string{"object", "undef", "preferArrowCallback", "forward", "nested", "disjunctive", "backward", "intoNegativeLookaround"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+id+"\t")) {
			t.Fatal("missing positive " + id)
		}
	}
	t.Logf("%d upstream and edge control files", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave17-fourth-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"throw", "no_throw_literal.a", "if(!this.possible(argument))", "if(true)"},
		{"regex", "backreference_structure.a", "if(group.start <= ref.start && ref.end <= group.end)", "if(false && group.start <= ref.start && ref.end <= group.end)"},
		{"arrow", "prefer_arrow_callback.a", "' =>'", "' => '"},
		{"tracker", "regexp_references.a", "p.children[0] === index && mode === 1", "p.children[0] === index && mode === 0"},
	} {
		original, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", m.file))
		if err != nil {
			t.Fatal(err)
		}
		s := strings.ReplaceAll(string(original), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		if strings.Count(s, m.from) != 1 {
			t.Fatal("mutant anchor " + m.name)
		}
		rule := h.write("mutant-"+m.name+".a", strings.Replace(s, m.from, m.to, 1))
		data, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		s = strings.ReplaceAll(string(data), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		s = strings.ReplaceAll(s, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
		if m.name == "regex" || m.name == "tracker" {
			data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/no_useless_backreference.a"))
			if err != nil {
				t.Fatal(err)
			}
			nativeRule := strings.ReplaceAll(string(data), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
			nativeRule = strings.Replace(nativeRule, filepath.Join(repository, "stage1/cohere/typeaware", m.file), rule, 1)
			rewritten := h.write("mutant-backreference-rule.a", nativeRule)
			s = strings.Replace(s, filepath.Join(repository, "stage1/cohere/typeaware/no_useless_backreference.a"), rewritten, 1)
		}

		s = strings.Replace(s, filepath.Join(repository, "stage1/cohere/typeaware", m.file), rule, 1)
		e := h.write("mutant-suite-"+m.name+".a", s)
		b := h.build(stage0, "mutant-"+m.name, e, sanitized, true)
		got := h.must("mutant-"+m.name+"-run", exec.Command(b, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or sanitizer caught " + m.name)
		}
		t.Logf("%s mutant exits 0, empty stderr; Go byte oracle catches byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, flags := range [][]string{{"--allow-named"}, {"--disallow-unbound"}, {"--allow-named", "--disallow-unbound"}} {
		commands := append([]string{config, manifest}, flags...)
		want := h.must("options-go", exec.Command(oracle, commands...))
		got := h.must("options-native", exec.Command(asan, commands...))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("options differ " + strings.Join(flags, " "))
		}
		t.Logf("options %v: %d identical bytes", flags, len(got.stdout))
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','global-binding-facts'));`)
	probe := h.write("released-input.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if exit, ok := got.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("binding question rejects released program: panic 70")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	registryArchive := h.archive("released-registry-checker", registryOverlay, false)
	registryBinary := h.build(stage0, "released-registry-mutant", released, registryArchive, true)
	gotRegistry := h.must("released-registry-mutant-run", exec.Command(registryBinary, config, probe))
	if len(gotRegistry.stderr) != 0 {
		t.Fatal("released registry mutant did not finish cleanly")
	}
	t.Log("released registry mutant exits 0, empty stderr; required panic 70 catches it")

	sourceProbe := h.write("released-source.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','symbol-identities'));`)
	sourceBinary := h.build(stage0, "released-source", sourceProbe, archive, false)
	sourceResult := h.run("released-source-run", exec.Command(sourceBinary, config, probe))
	if exit, ok := sourceResult.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(sourceResult.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatal("released source question escaped")
	}
	sourceMutant := h.build(stage0, "released-source-registry-mutant", sourceProbe, registryArchive, true)
	sourceGot := h.must("released-source-registry-run", exec.Command(sourceMutant, config, probe))
	if len(sourceGot.stderr) != 0 {
		t.Fatal("registry source mutant did not finish cleanly")
	}
	t.Log("symbol identity question rejects released handle; retained registry mutant finishes cleanly, required panic catches it")

	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE required")
		}
		if population.name == "compiler" {
			pin := h.must("corpus-pin", exec.Command("git", "-C", population.root, "rev-parse", "HEAD"))
			if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
				t.Fatal("wrong TypeScript pin")
			}
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if err != nil {
			t.Fatal(err)
		}
		paths = nil
		for _, p := range strings.Split(string(data), "\n") {
			if p != "" {
				paths = append(paths, filepath.Join(population.root, p))
			}
		}
		list := h.write(population.name+".manifest", strings.Join(paths, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, list)
		h.compare(population.name+"-asan", oracle, asan, population.config, list)
	}
}
