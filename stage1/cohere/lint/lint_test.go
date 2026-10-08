package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."
const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"

func prepareRegistry(t *testing.T, directory string) []registry.Descriptor {
	t.Helper()
	descriptors, err := registry.Generate(directory)
	if err != nil {
		t.Fatal(err)
	}
	return descriptors
}

func portFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (path == "gaps" || strings.HasSuffix(path, "testdata")) {
			return filepath.SkipDir
		}
		if !entry.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a") || strings.HasSuffix(path, "rule.json") || strings.HasSuffix(path, "mutant.json") || strings.HasSuffix(path, "oracle.go")) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

type execution struct {
	output   []byte
	duration time.Duration
}

// moduleDownload is the one line the go command writes to stderr on a clean build: fetching a module the
// cache doesn't hold yet, as on a fresh box when cohere's pin adds one. Any other stderr still fails.
var moduleDownload = regexp.MustCompile(`(?m)^go: downloading \S+ \S+\n`)

// commandDiagnostics is the stderr that fails a command: all of it, less the go command's module
// downloads.
func commandDiagnostics(name string, stderr []byte) []byte {
	if name == "go" {
		return moduleDownload.ReplaceAll(stderr, nil)
	}
	return stderr
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
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

func goOracle(t *testing.T) string {
	return goOracleFrom(t, ".")
}
func goOracleFrom(t *testing.T, sourceRoot string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	descriptors := prepareRegistry(t, sourceRoot)
	directory := t.TempDir()
	replacements := map[string]string{}
	var virtualFiles []string
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			t.Fatal(err)
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	execute(t, root, "go", args...)
	return binary
}

func buildPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	prepareRegistry(t, directory)
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
	prepareRegistry(t, directory)
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

// recoveryRows marks "recovery" each row whose file typescript-go's parse reports a diagnostic, when no other
// mode claims it, so the oracle compares that row's findings rather than refusing it as invalid corpus or
// fixing a file Go would not. A legacy octal escape is the case: no-octal-escape exists to report one, and
// TypeScript's parser reports it too. Other rows are left as they were.
func recoveryRows(t *testing.T, oracle string, rows []string) []string {
	t.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := execute(t, "", oracle, "--manifest", manifest(t, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		t.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
	}
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row
		if flags[index] != "1" {
			continue
		}
		fields := strings.Split(row, "\t")
		for len(fields) < 7 {
			fields = append(fields, "")
		}
		if fields[6] == "" {
			fields[6] = "recovery"
		}
		result[index] = strings.Join(fields, "\t")
	}
	return result
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
	// One source exercising the ten rules that came from the frequency-ranked slice (VOLUME.md), by default
	// and under the options each takes.
	path := filepath.Join(t.TempDir(), "options.ts")
	if err := os.WriteFile(path, []byte("interface Bare { n: number }; interface WrongType { value: number }; type Alias = string; const Choice={Yes:'Yes'} as const; function guard(x: unknown): x is string { return true; } console.log('one'); console['warn']('two'); let count=0; count++; for(let i=0;i<3;i++){count++;} interface CallableType { method(value: string): number; readonly property: (value: string) => number; } enum Direction { Left=1, Right=Left|2, Other=compute() } let boxed: Number; class Thing implements Boolean {} if(!flag){doOne();}else{doTwo();} const chosen=!flag?left:right; const assigning=()=>left=right; const wrapped=()=>(left=right); function returning(){return left=right;}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return append(rows, path,
		path+"\tno-plusplus\t\t\t\t{\"AllowForLoopAfterthoughts\":true}",
		path+"\t@typescript-eslint/method-signature-style\t\t\t\t{\"Style\":\"method\"}",
		path+"\t@typescript-eslint/prefer-literal-enum-member\t\t\t\t{\"AllowBitwiseExpressions\":true}",
		path+"\tno-return-assign\t\t\t\t\"always\"",
	)
}

// Capture every Run, including tests that assert repair fields directly. The overlay changes no rule.
func upstream(t *testing.T) []string {
	return upstreamFrom(t, ".")
}
func upstreamFrom(t *testing.T, sourceRoot string) []string {
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
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range prepareRegistry(t, sourceRoot) {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	for name, tests := range packages {
		execute(t, root, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(tests, "|")+")", "-count=1", "-timeout=10m")
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
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
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
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
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
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
				mode = "recovery"
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
	return compareWithJavaScript(t, oracle, binary, directory, path, emittedJavaScript(t, directory))
}
func compareWithJavaScript(t *testing.T, oracle, binary, directory, path, module string) []byte {
	t.Helper()
	want := execute(t, "", oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", node(t, directory, path, false)}, {"emitted JavaScript", runJavaScript(t, module, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
		if diff := difference(side.run.output, want.output); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	t.Logf("Go, Node, emitted JavaScript, native identical: %d bytes", len(want.output))
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
	rows := generated(t)
	for _, row := range upstream(t) {
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			t.Logf("EXPLICIT LIMIT: parser recovery is not ported for %s", row)
			checkRecoveryRefusal(t, oracle, binary, directory, row)
		} else {
			rows = append(rows, row)
		}
	}
	compare(t, oracle, binary, directory, manifest(t, recoveryRows(t, oracle, rows)))
}

// Recovery is a parser dependency, not successful lint parity. Keep the exact
// upstream malformed cases and prove that both ports refuse instead of silently
// returning the oracle's recovered findings.
func checkRecoveryRefusal(t *testing.T, oracle, binary, directory, row string) {
	t.Helper()
	recovered := manifest(t, []string{strings.TrimSuffix(row, "unsupported-recovery") + "recovery"})
	answer := execute(t, "", oracle, "--manifest", recovered)
	t.Logf("Go recovered output: %s", answer.output)
	path := manifest(t, []string{row})
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		args []string
	}{
		{binary, []string{"--manifest", path}},
		{"node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		command := exec.CommandContext(ctx, side.name, side.args...)
		output, err := os.CreateTemp(t.TempDir(), "recovery-refusal-")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = output
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err = command.Run()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		output.Close()
		if err == nil || (!timedOut && !strings.Contains(stderr.String(), "adamic: panic:")) {
			t.Fatalf("expected parser refusal from %s, got %v: %s", side.name, err, stderr.String())
		}
		t.Logf("explicit unsupported recovery: %s: timeout=%t: %v: %s", side.name, timedOut, err, stderr.String())
	}
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
			if !entry.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a")) {
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
	return copyPort(t, t.TempDir(), from, to, targets...)
}
func copyPort(t *testing.T, directory, from, to string, targets ...string) string {
	if len(targets) == 0 && from != "" {
		targets = []string{"lint.ts"}
	}
	changed := 0
	prepareRegistry(t, ".")
	for _, file := range portFiles(t) {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		selected := len(targets) == 0 || filepath.Clean(file) == filepath.Clean(targets[0])
		if from != "" && selected && !strings.HasSuffix(file, "mutant.json") && strings.Contains(source, from) {
			if strings.Count(source, from) != 1 {
				t.Fatalf("mutant anchor repeated in %s", file)
			}
			source = strings.Replace(source, from, to, 1)
			changed++
		}
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".a") {
			source = rewritePortImports(t, file, source)
		}
		destination := filepath.Join(directory, file)
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Validation requires each rule's owned witnesses too.
	for _, d := range prepareRegistry(t, ".") {
		paths, err := registry.Witnesses(filepath.Join("rules", d.Slug))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(directory, path)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if from != "" && changed != 1 {
		t.Fatalf("mutant anchor count: %d for %q", changed, from)
	}
	return directory
}

// Not parallel: sanitized rebuilds run in sequence to bound memory and precede timing.
func TestLegacyMutants(t *testing.T) {
	path := manifest(t, generated(t))
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", path).output
	for _, change := range []struct{ name, from, to string }{
		{"overlap winner misreported", "${winner} overlaps another fix", "${this.selected} overlaps another fix"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutant(t, change.from, change.to)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"emitted JavaScript", emittedNode(t, directory, path, false)}, {"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)}} {
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
		if !entry.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a")) {
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

// TestNodeTableIsLinkOnly requires the same output with a copy of every row appended to the node table,
// attached to nothing. Stage 1 reads the table only by following links from the root, and the flat copy
// of typescript-go's tree (#k4fm1vf) depends on it: its tables hold rows no link reaches. A rule or
// harness pass that walks the table by row reports on the copies and fails here.
func TestNodeTableIsLinkOnly(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	rows := generated(t)
	for _, row := range upstream(t) {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	path := manifest(t, recoveryRows(t, oracle, rows))
	binary := buildPort(t, directory, false)
	plain := execute(t, "", binary, "--manifest", path)
	junk := execute(t, "", binary, "--manifest", path, "--junk-rows")
	if diff := difference(junk.output, plain.output); diff != "" {
		t.Fatalf("output changed with unattached rows in the node table: %s", diff)
	}
	t.Logf("%d rows: identical with and without unattached node rows, %d bytes", len(rows), len(plain.output))
}

// TestShardsAgree requires the driver's output to be byte-identical however many processes share the
// manifest (#tj6d455): one, two, and the machine's cores. Equal counts cannot see a reordered or repeated
// case, so the whole output is compared, and the count mode too.
func TestShardsAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	rows := generated(t)
	for _, row := range upstream(t) {
		if !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	// The compiler files are the corpus with large files, where shards differ most in what they hold, so
	// the test refuses to run without them rather than passing on a smaller corpus.
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout: TestShardsAgree needs its compiler files")
	}
	matches, err := filepath.Glob(filepath.Join(source, "src/compiler/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatalf("no compiler files under %s", source)
	}
	rows = append(rows, matches...)
	path := manifest(t, recoveryRows(t, oracle, rows))
	binary := buildPort(t, directory, false)
	want := execute(t, "", binary, "--manifest", path)
	wantCount := execute(t, "", binary, "--manifest", path, "--count")
	for _, count := range []int{1, 2, runtime.NumCPU()} {
		got, err := shards.Run(binary, path, count, false)
		if err != nil {
			t.Fatal(err)
		}
		if diff := difference(got, want.output); diff != "" {
			t.Fatalf("%d shards: %s", count, diff)
		}
		gotCount, err := shards.Run(binary, path, count, true)
		if err != nil {
			t.Fatal(err)
		}
		if string(gotCount) != string(wantCount.output) {
			t.Fatalf("%d shards count %q, want %q", count, gotCount, wantCount.output)
		}
	}
	t.Logf("%d rows: identical at 1, 2 and %d shards, %d bytes", len(rows), runtime.NumCPU(), len(want.output))
}

// mutantBuilds bounds how many of TestMutants' sanitized native builds run at once. The subtests run in
// parallel, since each works in its own copy of the port and the cost is one build per rule, which grows
// with every port batch. A sanitized clang build of the whole port is the memory-heavy step, so only a few
// run together, whatever the machine's core count.
var mutantBuilds = make(chan struct{}, min(runtime.NumCPU(), 4))

func TestMutants(t *testing.T) {
	oracle := goOracle(t)
	for _, descriptor := range prepareRegistry(t, ".") {
		var change struct{ Name, File, From, To string }
		data, err := os.ReadFile(filepath.Join("rules", descriptor.Slug, "mutant.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &change); err != nil {
			t.Fatal(err)
		}
		t.Run(change.Name, func(t *testing.T) {
			t.Parallel()
			// A mutant can change only its own rule's output, so it runs on the rows that select that rule:
			// its witnesses, and the generated rows selecting it or all rules. Rows naming another rule
			// cannot show it. An "all" row stays as it is, since its options are every rule's bag.
			var rows []string
			for _, row := range generated(t) {
				// A bare path selects all rules, as main.ts reads it.
				selected := "all"
				if fields := strings.Split(row, "\t"); len(fields) > 1 && fields[1] != "" {
					selected = fields[1]
				}
				if selected == "all" || selected == descriptor.Name {
					rows = append(rows, row)
				}
			}
			rows = append(rows, recoveryRows(t, oracle, ownedWitnessRows(t, ".", descriptor))...)
			path := manifest(t, rows)
			want := execute(t, "", oracle, "--manifest", path).output
			if change.File == "" {
				change.File = descriptor.Module
			}
			directory := mutant(t, change.From, change.To, filepath.Join("rules", descriptor.Slug, change.File))
			// The slot is released by defer: buildPort fails with t.Fatal, which ends this goroutine, and a
			// slot held by a failed build would leave every other subtest waiting until the package timed out.
			binary := func() string {
				mutantBuilds <- struct{}{}
				defer func() { <-mutantBuilds }()
				return buildPort(t, directory, true)
			}()
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"emitted JavaScript", emittedNode(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s mutant survived on %s", change.Name, side.name)
				}
				t.Logf("%s caught on %s: %s", change.Name, side.name, difference(side.run.output, want))
			}
		})
	}
}

var portImport = regexp.MustCompile(`(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])`)

// Preserve imports within the copied port, and resolve outside imports from their
// original module directory. Rules may be arbitrarily deeper than context.ts.
func rewritePortImports(t *testing.T, file, source string) string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return portImport.ReplaceAllStringFunc(source, func(declaration string) string {
		parts := portImport.FindStringSubmatch(declaration)
		if !strings.HasPrefix(parts[3], ".") {
			return declaration
		}
		absolute := filepath.Clean(filepath.Join(root, filepath.Dir(file), parts[3]))
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return declaration
		}
		return parts[1] + parts[2] + filepath.ToSlash(absolute) + parts[4]
	})
}

func emittedJavaScript(t *testing.T, directory string) string {
	t.Helper()
	prepareRegistry(t, directory)
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "lint.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func emittedNode(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	return runJavaScript(t, emittedJavaScript(t, directory), manifest, count)
}
func runJavaScript(t *testing.T, module, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, module, "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}
