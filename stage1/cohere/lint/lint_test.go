package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."
const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"

var portFiles = []string{"finding.ts", "messages.ts", "settings.ts", "comments.ts", "unicode.ts", "lint.ts", "main.ts", "volume.ts", "volume_messages.ts"}

type execution struct {
	output   []byte
	duration time.Duration
}

// Output is a file, never a pipe: the large corpus must also work on Node's writev path.
func execute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	duration := time.Since(started)
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

func goOracle(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_lint_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	execute(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}

func buildPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "scanner")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func node(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}

func difference(got, want []byte) string {
	if bytes.Equal(got, want) {
		return ""
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		left, right := "<EOF>", "<EOF>"
		if i < len(a) {
			left = a[i]
		}
		if i < len(b) {
			right = b[i]
		}
		if left != right {
			caseLine := ""
			for j := i; j >= 0; j-- {
				if j < len(b) && strings.HasPrefix(b[j], "case ") {
					caseLine = b[j]
					break
				}
			}
			start := i - 5
			if start < 0 {
				start = 0
			}
			end := i + 5
			if end > len(b) {
				end = len(b)
			}
			return fmt.Sprintf("%s line %d: port %q, Go %q\nGo context:\n%s", caseLine, i+1, left, right, strings.Join(b[start:end], "\n"))
		}
	}
	return "different bytes"
}

func manifest(t *testing.T, rows []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.txt")
	if err := os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func generated(t *testing.T) []string {
	sources := []string{
		"debugger; if(x) debugger; while(x) debugger; label: debugger; function f(){ debugger; } function noop(){} switch(x){case 1: debugger;}",
		"if(x){} else { /* intent */ } try{}catch(e){}finally{}; switch('/*'){}; switch(x){/* intent */}; class C{static{}}",
		"a == b; 1 == 2; 1 == 1n; (typeof a) != 'number'; (null) === null; /a/ == null; `a` != 'b';",
		"var a; for(var x=0;x<3;x++){}; for(var y in z){}; for(var q of z){}; declare var b; declare global{ var c; namespace N{var d;} } namespace M{var e;}",
		"switch(x){case f( a /* c */ ):break;case f(a):break;case a?.b:break;case a.b:break;case (a):break;case a:break;case /[/*]/:break;case /[/*]/:break;case -1:break;case +1:break;case 1.0:break;case 1:break;}",
		"const é = '😀';\r\n é == null;\n debugger;\r\n if(é){} // outside\n",
		"outer: while(x){continue outer; break outer; await f(); async function nested(){await g();}} with(x){y();} new C(); void new D(); const holes=[,,x]; ([,a]=b); function* missing(){function* inner(){yield 1;} return 1;} function* good(){yield 1;} f(); var late;",
		"async function f(){for(let i=await init();await test();await update()){await body();} for await(const x of xs){await ignored();} for(const x of xs){await body();}} function g(){'use strict'; var top; f(); var late;} class C{static{f();var late;}}",
		"const placeholder='${x}${y}'; const empty='${}'; const regex=/=foo/g; const escaped=/\\=foo/; a|0; a|(0); 0|a; a>>>=b; ~a; for(a,b,c; a,b; a++,b++); for(let r=/;/; a,b; a++,b++); const arrow = a => (b,c); const deliberate = a => ((b,c));",
		"const nested=(a?true:false)?false:true; const bool=a == b ? false : true; const same=f()?true:true; const coerced=(a + b)?true:false; const assertion=(a as any)?false:true; const fallback=a?a:b??c; const nonnull=a!?true:false;",
		"// TODO: fix this xxx\nconst url='http://todo'; const fake=/[/*]TODO/; const template=`// TODO ${f() /* FIXME */} /* XXX */`; /* eslint no-warning-comments: error todo */\n// TODOé and fixme\n/** TODO decorated */ const arr=[1, // FIXME trailing comma\n]; import data from 'data' with {/* TODO attributes */}; type Empty=import('data', {with:{/* TODO import type */}});",
		"\ufeff\ufeff // TODO two marks\nconst x=1;",
		"// ſomething\n// K\n// ſa\n// Σ σ ς\n// TODO éééééééééééééééééééé é\n// TODO \\ quoted \"text\"\n// ABC TODO\n// K ſ TODO\n// -/ TODO\n",
		"async function f(){for await([, x] of values){await work();} for([,y] of values){} for([,z] in values){}}",
	}
	var rows []string
	for i, source := range sources {
		path := filepath.Join(t.TempDir(), fmt.Sprintf("generated-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		for _, options := range []string{"all\tAlways\tAlways\tfalse", "all\tSmart\t\ttrue", "all\tAlways\tNever\tfalse", "all\tAlways\tIgnore\ttrue", "all\tAlways\tAlways\tfalse\t{\"allowLoop\":true,\"allowSwitch\":true,\"int32Hint\":true,\"allow\":[\"~\"],\"allowInParentheses\":false,\"defaultAssignment\":false,\"terms\":[\"todo\",\"fixme\",\"xxx\"],\"location\":\"anywhere\",\"decoration\":[\"*\"]}", "all\tAlways\tAlways\tfalse\t{\"Require\":\"always\",\"terms\":[]}", "no-warning-comments\t\t\tfalse\t{\"terms\":[\"s\",\"k\",\"σ\",\"todo\"],\"location\":\"anywhere\"}", "no-warning-comments\t\t\tfalse\t{\"decoration\":[\"a\",\"-\",\"z\"]}", "no-warning-comments\t\t\tfalse\t{\"decoration\":[\"z\",\"-\",\"a\"]}", "no-warning-comments\t\t\tfalse\t{\"decoration\":[\"-\",\"/\"]}"} {
			rows = append(rows, path+"\t"+options)
		}
	}
	return rows
}

// Capture every Run, including tests that assert repair fields directly. The overlay changes no rule.
func upstream(t *testing.T) []string {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		t.Fatal(err)
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		t.Fatal("capture overlay anchor changed")
	}
	directory := t.TempDir()
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(directory, "capture")
	t.Setenv("COHERE_DOCS_CAPTURE", capture)
	execute(t, root, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/core", "-run", "Test(NoDebugger|NoEmpty|Eqeqeq|NoVar|NoDuplicateCase|NoContinue|NoWith|NoNew|NoSparseArrays|RequireYield|NoAwaitInLoop|VarsOnTop|NoTemplateCurlyInString|NoDivRegex|NoBitwise|NoLabels|NoSequences|UnicodeBom|NoUnneededTernary|NoWarningComments|NoPlusplus|NoNegatedCondition|NoReturnAssign)", "-count=1", "-timeout=10m")
	execute(t, root, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/base", "./internal/lint/rules/nexus", "./internal/lint/rules/adamic", "-run", "Test(ConsistencyNoConsole|ConsistencyRequireTypeSuffix|ConsistencyNoEnum|NoTypePredicate)", "-count=1", "-timeout=10m")
	execute(t, root, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/typescript", "-run", "Test(MethodSignatureStyle|NoWrapperObjectTypes|PreferLiteralEnumMember)", "-count=1", "-timeout=10m")
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Rule, Source, Outcome, FixedSource string
		Options                            json.RawMessage
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains("|no-debugger|no-empty|eqeqeq|no-var|no-duplicate-case|no-continue|no-with|no-new|no-sparse-arrays|require-yield|no-await-in-loop|vars-on-top|no-template-curly-in-string|no-div-regex|no-bitwise|no-labels|no-sequences|unicode-bom|no-unneeded-ternary|no-warning-comments|no-plusplus|base/consistency-no-console|nexus/consistency-require-type-suffix|adamic/no-type-predicate|@typescript-eslint/method-signature-style|@typescript-eslint/no-wrapper-object-types|@typescript-eslint/prefer-literal-enum-member|nexus/consistency-no-enum|no-negated-condition|no-return-assign|", "|"+row.Rule+"|") {
				continue
			}
			key := fmt.Sprintf("%s\t%+v\t%s", row.Rule, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		path := filepath.Join(directory, fmt.Sprintf("case-%03d.ts", i))
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			t.Fatal(err)
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				t.Fatal(err)
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		t.Fatalf("capture unexpectedly small: %d cases", len(rows))
	}
	t.Logf("cohere cases: %d unique source/rule/options combinations", len(rows))
	return rows
}
func compare(t *testing.T, oracle, binary, directory, path string) []byte {
	t.Helper()
	want := execute(t, "", oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
		if diff := difference(side.run.output, want.output); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	t.Logf("Go, Node, native identical: %d bytes", len(want.output))
	return want.output
}

// Not parallel: upstream capture uses t.Setenv and a process-wide fixture capture destination.
func TestRulesAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	rows := append(generated(t), volumeGenerated(t)...)
	for _, row := range upstream(t) {
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			t.Logf("EXPLICIT LIMIT: parser recovery is not ported for %s", row)
			checkRecoveryRefusal(t, oracle, binary, directory, row)
		} else {
			rows = append(rows, row)
		}
	}
	compare(t, oracle, binary, directory, manifest(t, rows))
}

// Not parallel: the large sanitized corpus runs before timing samples.
func TestCompilerAndStage1Agree(t *testing.T) {
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to pinned v6.0.3")
	}
	pin := execute(t, "", "git", "-C", source, "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin.output)) != compilerCommit {
		t.Fatal("compiler corpus pin differs")
	}
	var rows []string
	for _, root := range []string{filepath.Join(source, "src/compiler"), filepath.Join(repository, "stage1")} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				rows = append(rows, absolute)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("compiler and stage1: %d files", len(rows))
	compare(t, goOracle(t), buildPort(t, directory, true), directory, manifest(t, rows))
}
func mutant(t *testing.T, from, to string, targets ...string) string {
	target := "lint.ts"
	if len(targets) > 0 {
		target = targets[0]
	}
	directory := t.TempDir()
	for _, file := range portFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		if file == target {
			if strings.Count(source, from) != 1 {
				t.Fatalf("mutant anchor count for %q", from)
			}
			source = strings.Replace(source, from, to, 1)
		}
		typescript, err := filepath.Abs("../../typescript")
		if err != nil {
			t.Fatal(err)
		}
		source = strings.ReplaceAll(source, "../../typescript", typescript)
		if err := os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

// Not parallel: sanitized rebuilds run in sequence to bound memory and precede timing.
func TestMutants(t *testing.T) {
	path := manifest(t, generated(t))
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", path).output
	for _, change := range []struct{ name, from, to string }{
		{"suggestion applied as fix", "hasTypeOf || sameType ? 'fix' : 'suggestion'", "hasTypeOf || sameType ? 'fix' : 'fix'"},
		{"empty function body reported", "if(!functionBody && !(this.allowCatch", "if((functionBody || !functionBody) && !(this.allowCatch"},
		{"duplicate case suppressed", "if(seen.has(signature))", "if(!seen.has(signature))"},
		{"control statement omitted", "node.kind === 'ContinueStatement' && this.enabled('no-continue')", "node.kind === 'BreakStatement' && this.enabled('no-continue')"},
		{"destructuring hole reported", "!this.assignmentTarget(index) &&", ""},
		{"nested generator owns outer yield", "].includes(this.node(child).kind)", "].includes('no boundary')"},
		{"await crosses function boundary", "this.functionLike(parent) ||", " "},
		{"option ignored", "this.settings.read('int32hint', 'false') === 'true'", "this.settings.read('int32hint', 'false') === 'false'"},
		{"regex fix eats extra byte", "finding.editEnd = start + 2;", "finding.editEnd = start + 3;"},
		{"boolean inverse changed", "? '!='", "? '!=='"},
		{"comment self directive exemption removed", "if(selfDirective(value))", "if(selfDirective(value) && value === '')"},
		{"bom removes two marks", "finding.editEnd = 1;", "finding.editEnd = this.source.startsWith('\\ufeff\\ufeff') ? 2 : 1;"},
		{"label option widened", "this.settings.read('allowloop', 'false') === 'true'", "this.settings.read('allowloop', 'false') !== 'true'"},
		{"directive prefix ignored", "let prefix = !staticBlock;", "let prefix = false;"},
		{"comma chain reports inner", "!this.comma(parent)", "this.comma(parent)"},
		{"empty placeholder reported", "node.text.indexOf('}', body) > body", "node.text.indexOf('}', body) >= body"},
		{"overlap winner misreported", "${winner} overlaps another fix", "${this.selected} overlaps another fix"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutant(t, change.from, change.to)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s mutant survived on %s", change.name, side.name)
				}
				t.Logf("%s caught on %s: %s", change.name, side.name, difference(side.run.output, want))
			}
		})
	}
}

func TestDecorationOptionMutant(t *testing.T) {
	t.Parallel()
	path := manifest(t, generated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	directory := mutant(t, "foldedRange(character, first, last)", "foldedRange(character, first, first)", "comments.ts")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"native", execute(t, "", binary, "--manifest", path)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("decoration range mutant survived on %s", side.name)
		}
		t.Logf("decoration range collapsed to one character caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

func TestCountGuardMutant(t *testing.T) {
	t.Parallel()
	source := filepath.Join(t.TempDir(), "count.ts")
	if err := os.WriteFile(source, []byte("var a; debugger;"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source})
	oracle := goOracle(t)
	answer := execute(t, "", oracle, "--manifest", path).output
	count := execute(t, "", oracle, "--manifest", path, "--count").output
	directory := mutant(t, "if(countOnly) {\n        return linter.findings.length;", "if(countOnly) {\n        return linter.findings.length + 1;", "main.ts")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name            string
		ordinary, count execution
	}{
		{"Node", node(t, directory, path, false), node(t, directory, path, true)},
		{"native", execute(t, "", binary, "--manifest", path), execute(t, "", binary, "--manifest", path, "--count")},
	} {
		if !bytes.Equal(side.ordinary.output, answer) {
			t.Fatalf("%s mutant was caught outside the count check", side.name)
		}
		if bytes.Equal(side.count.output, count) {
			t.Fatalf("%s count mutant survived", side.name)
		}
		t.Logf("count check alone caught %s: mutant=%q Go=%q; ordinary output remains identical", side.name, side.count.output, count)
	}
}

// Not parallel: interleaved timing samples must not compete with tests in this package.
func TestThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		t.Skip("set ADAMIC_LINT_BENCH=1")
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("missing pinned compiler checkout")
	}
	var rows []string
	err := filepath.WalkDir(filepath.Join(source, "src/compiler"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			rows = append(rows, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	path := manifest(t, rows)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, false)
	machine := execute(t, "", "uname", "-a")
	cpu, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		t.Fatal(err)
	}
	model := ""
	for _, line := range strings.Split(string(cpu), "\n") {
		if strings.HasPrefix(line, "model name") {
			model = line
			break
		}
	}
	loadBefore, _ := os.ReadFile("/proc/loadavg")
	t.Logf("machine %s; %s; load before %s", strings.TrimSpace(string(machine.output)), model, strings.TrimSpace(string(loadBefore)))
	best := map[string]time.Duration{}
	var want []byte
	for round := 0; round < 5; round++ {
		for _, name := range []string{"Go", "native", "Node"} {
			var result execution
			switch name {
			case "Go":
				result = execute(t, "", oracle, "--manifest", path, "--count")
			case "native":
				result = execute(t, "", binary, "--manifest", path, "--count")
			case "Node":
				result = node(t, directory, path, true)
			}
			if want == nil {
				want = result.output
			}
			if !bytes.Equal(want, result.output) {
				t.Fatalf("count differs on %s: %q vs %q", name, result.output, want)
			}
			if best[name] == 0 || result.duration < best[name] {
				best[name] = result.duration
			}
			t.Logf("round %d %s %s findings=%s", round+1, name, result.duration, strings.TrimSpace(string(result.output)))
		}
	}
	var count int
	if _, err := fmt.Sscan(string(want), &count); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Go", "native", "Node"} {
		t.Logf("best of 5 %s: %.6fs, %.2f findings/s (%d files, %d findings)", name, best[name].Seconds(), float64(count)/best[name].Seconds(), len(rows), count)
	}
	loadAfter, _ := os.ReadFile("/proc/loadavg")
	t.Logf("load after %s", strings.TrimSpace(string(loadAfter)))
}
