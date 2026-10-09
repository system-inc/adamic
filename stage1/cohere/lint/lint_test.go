package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
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
	command := exec.Command(name, args...)
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
	err = testguard.Run(command, testguard.Budget, testguard.Ceiling)
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
	if isPackage(sourceRoot) {
		value := shared("oracle", func(value *sharedValue) {
			directory, err := os.MkdirTemp(sharedDirectory, "oracle-")
			if err != nil {
				value.err = err
				return
			}
			value.path, value.err = goOracleIn(sourceRoot, directory)
		})
		if value.err != nil {
			t.Fatal(value.err)
		}
		return value.path
	}
	binary, err := goOracleIn(sourceRoot, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return binary
}

func buildPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	if isPackage(directory) {
		value := shared(fmt.Sprintf("port sanitize=%t", sanitize), func(value *sharedValue) {
			value.path, value.err = sharedPath("scanner")
			if value.err == nil {
				value.err = buildPortTo(directory, sanitize, value.path)
			}
		})
		if value.err != nil {
			t.Fatal(value.err)
		}
		return value.path
	}
	binary := filepath.Join(t.TempDir(), "scanner")
	if err := buildPortTo(directory, sanitize, binary); err != nil {
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

// upstream is the run's capture of every asserted upstream case for this package's rules (captureUpstream).
func upstream(t *testing.T) []string {
	return upstreamFrom(t, ".")
}
func upstreamFrom(t *testing.T, sourceRoot string) []string {
	t.Helper()
	if isPackage(sourceRoot) {
		value := shared("upstream", func(value *sharedValue) {
			directory, err := os.MkdirTemp(sharedDirectory, "upstream-")
			if err != nil {
				value.err = err
				return
			}
			value.rows, value.err = captureUpstream(sourceRoot, directory)
		})
		if value.err != nil {
			t.Fatal(value.err)
		}
		t.Logf("cohere cases: %d unique source/rule/options combinations", len(value.rows))
		return append([]string(nil), value.rows...)
	}
	rows, err := captureUpstream(sourceRoot, t.TempDir())
	if err != nil {
		t.Fatal(err)
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
	started := time.Now()
	rulesAgreeLowered(t)
	t.Logf("TestRulesAgree (setup lowered): %.3fs", time.Since(started).Seconds())
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: setup over 60s budget")
	}
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
		command := exec.Command(side.name, side.args...)
		output, err := os.CreateTemp(t.TempDir(), "recovery-refusal-")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = output
		var stderr bytes.Buffer
		command.Stderr = &stderr
		err = testguard.Run(command, 2*time.Second, testguard.Ceiling)
		cpuGuard := strings.Contains(fmt.Sprint(err), "child CPU hang guard exceeded")
		output.Close()
		if err == nil || (!cpuGuard && !strings.Contains(stderr.String(), "adamic: panic:")) {
			t.Fatalf("expected parser refusal from %s, got %v: %s", side.name, err, stderr.String())
		}
		t.Logf("explicit unsupported recovery: %s: CPU-guard=%t: %v: %s", side.name, cpuGuard, err, stderr.String())
	}
}

// Not parallel: the large sanitized corpus runs before timing samples.
// Largest-first scheduling is deterministic: equal sizes retain manifest order, and equal
// loads choose the lowest shard index. Original case numbers survive the local manifests.
func compilerShardAssignments(t *testing.T, rows []string, count int) [][]int {
	t.Helper()
	sizes := make([]int64, len(rows))
	order := make([]int, len(rows))
	for index, row := range rows {
		info, err := os.Stat(row)
		if err != nil {
			t.Fatal(err)
		}
		sizes[index], order[index] = info.Size(), index
	}
	sort.SliceStable(order, func(i, j int) bool { return sizes[order[i]] > sizes[order[j]] })
	assignments := make([][]int, count)
	loads := make([]int64, count)
	for _, index := range order {
		shard := 0
		for candidate := 1; candidate < count; candidate++ {
			if loads[candidate] < loads[shard] {
				shard = candidate
			}
		}
		assignments[shard] = append(assignments[shard], index)
		loads[shard] += sizes[index]
	}
	for shard := range assignments {
		sort.Ints(assignments[shard])
		t.Logf("compiler shard %d/%d: %d files, %d bytes", shard, count, len(assignments[shard]), loads[shard])
		if len(assignments[shard]) == 1 {
			t.Logf("compiler shard %d/%d only file: %s", shard, count, rows[assignments[shard][0]])
		}
	}
	return assignments
}

// Both comparison failures and Merge's missing/extra-case failures retain the input path.
func compilerCasePath(rows []string, message string) string {
	var index int
	if _, err := fmt.Sscanf(message, "case %d", &index); err == nil && index >= 0 && index < len(rows) {
		return rows[index] + ": " + strings.TrimPrefix(message, fmt.Sprintf("case %d ", index))
	}
	return message
}

// shards.Run accepts an executable, so this adapter supplies Node's fixed arguments too. Each
// backend still writes to a file, fails on stderr, and has execute's CPU hang guard. Durations belong
// to individual shards, not the merged run, so a slow shard remains visible on a loaded gate seat.
func compilerShardLauncher(t *testing.T, command string, arguments []string, rows []string, assignments [][]int) (string, string) {
	t.Helper()
	directory := t.TempDir()
	configuration, err := json.Marshal(struct {
		Command     string
		Arguments   []string
		Directory   string
		Executable  string
		Assignments [][]int
	}{command, arguments, directory, guardExecutable(t), assignments})
	if err != nil {
		t.Fatal(err)
	}
	for shard, indices := range assignments {
		local := make([]string, len(indices))
		for index, original := range indices {
			local[index] = rows[original]
		}
		if err := os.WriteFile(filepath.Join(directory, fmt.Sprintf("%d.manifest", shard)), []byte(strings.Join(local, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	launcher := filepath.Join(directory, "shard.cjs")
	source := `#!/usr/bin/env node
const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const configuration = ` + string(configuration) + `;
const args = process.argv.slice(2);
const shard = args[args.indexOf('--shard') + 1].replace('/', '-');
const shardIndex = Number(shard.split('-')[0]);
const indices = configuration.Assignments[shardIndex];
const backendArgs = ['--manifest', path.join(configuration.Directory, shardIndex + '.manifest')];
const outputPath = path.join(configuration.Directory, shard + '.stdout');
const errorPath = path.join(configuration.Directory, shard + '.stderr');
const output = fs.openSync(outputPath, 'w');
const errors = fs.openSync(errorPath, 'w');
const started = process.hrtime.bigint();
const result = spawnSync(configuration.Executable, ['-test.run=^TestCompilerGuardBackend$', '--', configuration.Command, ...(configuration.Arguments || []), ...backendArgs], {
  env: {...process.env, ADAMIC_LINT_GUARD_BACKEND: '1'}, stdio: ['ignore', output, errors]
});
fs.writeFileSync(path.join(configuration.Directory, shard + '.time'),
  String(Number(process.hrtime.bigint() - started) / 1e6));
fs.closeSync(output);
fs.closeSync(errors);
const stderr = fs.readFileSync(errorPath);
if (result.error || result.status !== 0 || stderr.length !== 0) {
  console.error('shard inputs: ' + fs.readFileSync(path.join(configuration.Directory, shardIndex + '.manifest'), 'utf8'));
  console.error(result.error || ('exit ' + result.status + ', signal ' + result.signal));
  process.stderr.write(stderr);
  process.exitCode = 1;
} else {
  // Latin-1 preserves every output byte while only ASCII case headers are rewritten.
  const local = fs.readFileSync(outputPath).toString('latin1');
  const merged = local.replace(/^case ([0-9]+)$/gm, (_, number) => {
    const original = indices[Number(number)];
    if (original === undefined) throw new Error('unexpected local case ' + number);
    return 'case ' + original;
  });
  process.stdout.write(Buffer.from(merged, 'latin1'));
}
`
	if err := os.WriteFile(launcher, []byte(source), 0755); err != nil {
		t.Fatal(err)
	}
	return launcher, directory
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

// Not parallel: this semantic overlap check precedes timing samples.
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
			}{{"Node", node(t, directory, path, false)}, {"emitted JavaScript", emittedNode(t, directory, path, false)}} {
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
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
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
	for _, side := range []struct {
		name            string
		ordinary, count execution
	}{
		{"Node", node(t, directory, path, false), node(t, directory, path, true)},
		{"emitted JavaScript", emittedNode(t, directory, path, false), emittedNode(t, directory, path, true)},
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
// Not parallel: uses shared upstream capture and oracle products.
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

// One edited JSX rule is the native canary. Semantic mutations are held to Go on
// Node and emitted JavaScript; this copy also proves sanitized native matches the
// mutated Node result, including JSX parsing, text spans and finding serialization.
const nativeCanaryRule = "react/jsx-no-comment-textnodes"

// Not parallel: prepares the shared live registry and oracle before mutant comparisons.
func TestMutants(t *testing.T) {
	oracle := goOracle(t)
	descriptors := prepareRegistry(t, ".")
	hasCanary := false
	for _, descriptor := range descriptors {
		hasCanary = hasCanary || descriptor.Name == nativeCanaryRule
	}
	if !hasCanary {
		t.Fatalf("native canary rule %s is missing", nativeCanaryRule)
	}
	for _, descriptor := range descriptors {
		var change struct{ Name, File, From, To string }
		data, err := os.ReadFile(filepath.Join("rules", descriptor.Slug, "mutant.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &change); err != nil {
			t.Fatal(err)
		}
		// This canary is covered by the top-level JSX textnode shard tests.
		if change.Name == "react-jsx-no-comment-textnodes mutant from batch 8" {
			continue
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
			mutatedNode := node(t, directory, path, false)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", mutatedNode}, {"emitted JavaScript", emittedNode(t, directory, path, false)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s mutant survived on %s", change.Name, side.name)
				}
				t.Logf("%s caught on %s: %s", change.Name, side.name, difference(side.run.output, want))
			}
			if descriptor.Name == nativeCanaryRule {
				binary := buildPort(t, directory, true)
				got := execute(t, "", binary, "--manifest", path)
				if diff := difference(got.output, mutatedNode.output); diff != "" {
					t.Fatalf("native canary differs from mutated Node: %s", diff)
				}
				t.Logf("sanitized native canary %s equals mutated Node: %d bytes", descriptor.Name, len(got.output))
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
	if isPackage(directory) {
		value := shared("emitted JavaScript", func(value *sharedValue) {
			value.path, value.err = sharedPath("lint.mjs")
			if value.err == nil {
				value.err = emitJavaScriptTo(directory, value.path)
			}
		})
		if value.err != nil {
			t.Fatal(value.err)
		}
		return value.path
	}
	path := filepath.Join(t.TempDir(), "lint.mjs")
	if err := emitJavaScriptTo(directory, path); err != nil {
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
