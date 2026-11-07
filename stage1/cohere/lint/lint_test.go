package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/executable"
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

func prepareRegistry(test *testing.T, directory string) []registry.Descriptor {
	test.Helper()
	descriptors, generateError := registry.Generate(directory)
	if generateError != nil {
		test.Fatal(generateError)
	}
	return descriptors
}

// portFiles are the port's sources, relative to this directory. A rule's sources are the files directly in
// rules/<slug>/ (registry.Discover and RuleModule read nothing deeper), so the walk never enters a rule's evidence/,
// testdata/ or any other directory below it: listing one put every rule's evidence into the read sets of every test and
// product this walk keys, so a doc-only change under one rule's evidence/ moved 861 of 990 keys (#s57yq6y).
func portFiles(test *testing.T) []string {
	test.Helper()
	var files []string
	walkError := filepath.WalkDir(".", func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() && (path == "gaps" || strings.HasSuffix(path, "testdata") || lintRuleSubdirectory(path)) {
			return filepath.SkipDir
		}
		if !entry.IsDir() && (strings.HasSuffix(path, ".ts") ||
			strings.HasSuffix(path, ".a") ||
			strings.HasSuffix(path, "rule.json") ||
			strings.HasSuffix(path, "mutant.json") ||
			strings.HasSuffix(path, "oracle.go")) {
			files = append(files, path)
		}
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return files
}

// lintRuleSubdirectory is whether path is a directory below a rule's own, rules/<slug>/<directory> (evidence/,
// testdata/, validation/, evidence-completion/), given relative to this directory, to the repository, or absolute. A
// rule's sources sit directly in its directory, so a walk that keys or copies them skips these: what churns there is
// evidence and reports, and a listing of it in any lint test's read set moved every later test's key (#s57yq6y).
func lintRuleSubdirectory(path string) bool {
	slashed := filepath.ToSlash(filepath.Clean(path))
	if index := strings.LastIndex(slashed, "stage1/cohere/lint/rules/"); index >= 0 {
		slashed = slashed[index+len("stage1/cohere/lint/"):]
	}
	parts := strings.Split(slashed, "/")
	return len(parts) >= 3 && parts[0] == "rules"
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
func execute(test *testing.T, directory, name string, args ...string) execution {
	test.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node, go, git, clang) is a tool: the one internal/toolchain resolves, with a tool's environment,
		// not PATH's, run in directory with PWD to match; git gets GitEnv. A program a test built (the Go oracle, a
		// port) runs as named, with this process's environment.
		if name == "git" {
			command = toolchain.GitCommandIn(directory, args...)
		} else {
			command = toolchain.CommandIn(directory, name, args...)
		}
	}
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	runError := testguard.Run(command, testguard.Budget, testguard.Ceiling)
	duration := time.Since(started)
	if runError != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return execution{data, duration}
}

func buildPort(test *testing.T, directory string, sanitize bool) string {
	test.Helper()
	if isPackage(directory) {
		value := shared(fmt.Sprintf("port sanitize=%t", sanitize), func(value *sharedValue) {
			value.path, value.err = sharedPath("scanner")
			if value.err == nil {
				value.err = buildPortTo(directory, sanitize, value.path)
			}
		})
		if value.err != nil {
			test.Fatal(value.err)
		}
		return value.path
	}
	binary := filepath.Join(test.TempDir(), "scanner")
	if buildPortToError := buildPortTo(directory, sanitize, binary); buildPortToError != nil {
		test.Fatal(buildPortToError)
	}
	return binary
}

func node(test *testing.T, directory, manifest string, count bool) execution {
	test.Helper()
	prepareRegistry(test, directory)
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	args := []string{
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(directory, "main.ts"),
		"--manifest",
		manifest,
	}
	if count {
		args = append(args, "--count")
	}
	return execute(test, "", "node", args...)
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
			return fmt.Sprintf(
				"%s line %d: port %q, Go %q\nGo context:\n%s",
				caseLine,
				i+1,
				left,
				right,
				strings.Join(b[start:end], "\n"),
			)
		}
	}
	return "different bytes"
}

func manifest(test *testing.T, rows []string) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "manifest.txt")
	if writeError := os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

// recoveryRows marks "recovery" each row whose file typescript-go's parse reports a diagnostic, when no other
// mode claims it, so the oracle compares that row's findings rather than refusing it as invalid corpus or
// fixing a file Go would not. A legacy octal escape is the case: no-octal-escape exists to report one, and
// TypeScript's parser reports it too. Other rows are left as they were.
func recoveryRows(test *testing.T, oracle string, rows []string) []string {
	test.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := execute(test, "", oracle, "--manifest", manifest(test, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		test.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
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
func generated(test *testing.T) []string {
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
	for index, source := range sources {
		path := filepath.Join(test.TempDir(), fmt.Sprintf("generated-%d.ts", index))
		if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		for _, options := range []string{
			"all\tAlways\tAlways\tfalse",
			"all\tSmart\t\ttrue",
			"all\tAlways\tNever\tfalse",
			"all\tAlways\tIgnore\ttrue",
			"all\tAlways\tAlways\tfalse\t{\"allowLoop\":true,\"allowSwitch\":true,\"int32Hint\":true,\"allow\":[\"~\"],\"allowInParentheses\":false,\"defaultAssignment\":false,\"terms\":[\"todo\",\"fixme\",\"xxx\"],\"location\":\"anywhere\",\"decoration\":[\"*\"]}",
			"all\tAlways\tAlways\tfalse\t{\"Require\":\"always\",\"terms\":[]}",
			"no-warning-comments\t\t\tfalse\t{\"terms\":[\"s\",\"k\",\"σ\",\"todo\"],\"location\":\"anywhere\"}",
			"no-warning-comments\t\t\tfalse\t{\"decoration\":[\"a\",\"-\",\"z\"]}",
			"no-warning-comments\t\t\tfalse\t{\"decoration\":[\"z\",\"-\",\"a\"]}",
			"no-warning-comments\t\t\tfalse\t{\"decoration\":[\"-\",\"/\"]}",
		} {
			rows = append(rows, path+"\t"+options)
		}
	}
	// One source exercising the ten rules that came from the frequency-ranked slice (VOLUME.md), by default
	// and under the options each takes.
	path := filepath.Join(test.TempDir(), "options.ts")
	if writeError := os.WriteFile(path, []byte(
		"interface Bare { n: number }; interface WrongType { value: number }; type Alias = string; const Choice={Yes:'Yes'} as const; function guard(x: unknown): x is string { return true; } console.log('one'); console['warn']('two'); let count=0; count++; for(let i=0;i<3;i++){count++;} interface CallableType { method(value: string): number; readonly property: (value: string) => number; } enum Direction { Left=1, Right=Left|2, Other=compute() } let boxed: Number; class Thing implements Boolean {} if(!flag){doOne();}else{doTwo();} const chosen=!flag?left:right; const assigning=()=>left=right; const wrapped=()=>(left=right); function returning(){return left=right;}\n",
	), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return append(rows, path,
		path+"\tno-plusplus\t\t\t\t{\"AllowForLoopAfterthoughts\":true}",
		path+"\t@typescript-eslint/method-signature-style\t\t\t\t{\"Style\":\"method\"}",
		path+"\t@typescript-eslint/prefer-literal-enum-member\t\t\t\t{\"AllowBitwiseExpressions\":true}",
		path+"\tno-return-assign\t\t\t\t\"always\"",
	)
}

// upstream is the run's capture of every asserted upstream case for this package's rules (captureUpstream).
func upstream(test *testing.T) []string {
	return upstreamFrom(test, ".")
}
func upstreamFrom(test *testing.T, sourceRoot string) []string {
	test.Helper()
	if isPackage(sourceRoot) {
		value := shared("upstream", func(value *sharedValue) {
			directory, temporaryDirectoryError := os.MkdirTemp(sharedDirectory, "upstream-")
			if temporaryDirectoryError != nil {
				value.err = temporaryDirectoryError
				return
			}
			value.rows, value.err = captureUpstream(sourceRoot, directory)
		})
		if value.err != nil {
			test.Fatal(value.err)
		}
		test.Logf("cohere cases: %d unique source/rule/options combinations", len(value.rows))
		return append([]string(nil), value.rows...)
	}
	rows, captureUpstreamError := captureUpstream(sourceRoot, test.TempDir())
	if captureUpstreamError != nil {
		test.Fatal(captureUpstreamError)
	}
	test.Logf("cohere cases: %d unique source/rule/options combinations", len(rows))
	return rows
}
func compare(test *testing.T, oracle, binary, directory, path string) []byte {
	test.Helper()
	return compareWithJavaScript(test, oracle, binary, directory, path, emittedJavaScript(test, directory))
}
func compareWithJavaScript(test *testing.T, oracle, binary, directory, path, module string) []byte {
	test.Helper()
	want := execute(test, "", oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"emitted JavaScript", runJavaScript(test, module, path, false)},
		{"native", execute(test, "", binary, "--manifest", path)},
	} {
		if diff := difference(side.run.output, want.output); diff != "" {
			test.Fatalf("%s: %s", side.name, diff)
		}
	}
	test.Logf("Go, Node, emitted JavaScript, native identical: %d bytes", len(want.output))
	return want.output
}

// Recovery is a parser dependency, not successful lint parity. Keep the exact
// upstream malformed cases and prove that both ports refuse instead of silently
// returning the oracle's recovered findings.
func checkRecoveryRefusal(test *testing.T, oracle, binary, directory, row string) {
	test.Helper()
	recovered := manifest(test, []string{strings.TrimSuffix(row, "unsupported-recovery") + "recovery"})
	answer := execute(test, "", oracle, "--manifest", recovered)
	test.Logf("Go recovered output: %s", answer.output)
	path := manifest(test, []string{row})
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, side := range []struct {
		name string
		args []string
	}{
		{binary, []string{"--manifest", path}},
		{"node", []string{
			"--disable-warning=ExperimentalWarning",
			runner,
			filepath.Join(directory, "main.ts"),
			"--manifest",
			path,
		}},
	} {
		command := exec.Command(side.name, side.args...)
		if side.name == "node" {
			// Node is a tool, the one internal/toolchain resolves, as execute runs it.
			command = toolchain.Command(side.name, side.args...)
		}
		output, temporaryFileError := os.CreateTemp(test.TempDir(), "recovery-refusal-")
		if temporaryFileError != nil {
			test.Fatal(temporaryFileError)
		}
		command.Stdout = output
		var stderr bytes.Buffer
		command.Stderr = &stderr
		runError := testguard.Run(command, 2*time.Second, testguard.Ceiling)
		cpuGuard := strings.Contains(fmt.Sprint(runError), "child CPU hang guard exceeded")
		output.Close()
		if runError == nil || (!cpuGuard && !strings.Contains(stderr.String(), "adamic: panic:")) {
			test.Fatalf("expected parser refusal from %s, got %v: %s", side.name, runError, stderr.String())
		}
		test.Logf(
			"explicit unsupported recovery: %s: CPU-guard=%t: %v: %s",
			side.name,
			cpuGuard,
			runError,
			stderr.String(),
		)
	}
}

// Not parallel: the large sanitized corpus runs before timing samples.
// Largest-first scheduling is deterministic: equal sizes retain manifest order, and equal
// loads choose the lowest shard index. Original case numbers survive the local manifests.
func compilerShardAssignments(test *testing.T, rows []string, count int) [][]int {
	test.Helper()
	sizes := make([]int64, len(rows))
	order := make([]int, len(rows))
	for index, row := range rows {
		info, statError := os.Stat(row)
		if statError != nil {
			test.Fatal(statError)
		}
		sizes[index], order[index] = info.Size(), index
	}
	sort.SliceStable(order, func(left, right int) bool { return sizes[order[left]] > sizes[order[right]] })
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
		test.Logf("compiler shard %d/%d: %d files, %d bytes", shard, count, len(assignments[shard]), loads[shard])
		if len(assignments[shard]) == 1 {
			test.Logf("compiler shard %d/%d only file: %s", shard, count, rows[assignments[shard][0]])
		}
	}
	return assignments
}

// Both comparison failures and Merge's missing/extra-case failures retain the input path.
func compilerCasePath(rows []string, message string) string {
	var index int
	if _, scanError := fmt.Sscanf(message, "case %d", &index); scanError == nil && index >= 0 && index < len(rows) {
		return rows[index] + ": " + strings.TrimPrefix(message, fmt.Sprintf("case %d ", index))
	}
	return message
}

// shards.Run accepts an executable, so this adapter supplies Node's fixed arguments too. Each
// backend still writes to a file, fails on stderr, and has execute's CPU hang guard. Durations belong
// to individual shards, not the merged run, so a slow shard remains visible on a loaded gate seat.
func compilerShardLauncher(
	test *testing.T,
	command string,
	arguments []string,
	rows []string,
	assignments [][]int,
) (string, string) {
	test.Helper()
	directory := test.TempDir()
	configuration, encodeError := json.Marshal(struct {
		Command     string
		Arguments   []string
		Directory   string
		Executable  string
		Assignments [][]int
	}{command, arguments, directory, guardExecutable(test), assignments})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	for shard, indices := range assignments {
		local := make([]string, len(indices))
		for index, original := range indices {
			local[index] = rows[original]
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, fmt.Sprintf("%d.manifest", shard)),
			[]byte(strings.Join(local, "\n")+"\n"),
			0644,
		); writeError != nil {
			test.Fatal(writeError)
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
	if writeError := executable.Write(launcher, []byte(source)); writeError != nil {
		test.Fatal(writeError)
	}
	return launcher, directory
}

func mutant(test *testing.T, from, to string, targets ...string) string {
	return copyPort(test, test.TempDir(), from, to, targets...)
}
func copyPort(test *testing.T, directory, from, to string, targets ...string) string {
	return copyPortImporting(test, directory, rewritePortImports, from, to, targets...)
}

// copyPortImporting is copyPort with imports, which spells each copied file's imports: rewritePortImports points the
// ones leaving the package at this checkout, and keepPortImports leaves them as written (portTree).
func copyPortImporting(
	test *testing.T,
	directory string,
	imports func(test *testing.T, file, source string) string,
	from, to string,
	targets ...string,
) string {
	if len(targets) == 0 && from != "" {
		targets = []string{"lint.ts"}
	}
	changed := 0
	prepareRegistry(test, ".")
	for _, file := range portFiles(test) {
		data, readError := os.ReadFile(file)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(data)
		selected := len(targets) == 0 || filepath.Clean(file) == filepath.Clean(targets[0])
		if from != "" && selected && !strings.HasSuffix(file, "mutant.json") && strings.Contains(source, from) {
			if strings.Count(source, from) != 1 {
				test.Fatalf("mutant anchor repeated in %s", file)
			}
			source = strings.Replace(source, from, to, 1)
			changed++
		}
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".a") {
			source = imports(test, file, source)
		}
		destination := filepath.Join(directory, file)
		if createDirectoryError := os.MkdirAll(filepath.Dir(destination), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(destination, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	// Validation requires each rule's owned witnesses too.
	for _, descriptor := range prepareRegistry(test, ".") {
		paths, witnessesError := registry.Witnesses(filepath.Join("rules", descriptor.Slug))
		if witnessesError != nil {
			test.Fatal(witnessesError)
		}
		for _, path := range paths {
			data, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			target := filepath.Join(directory, path)
			if createDirectoryError := os.MkdirAll(filepath.Dir(target), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if writeError := os.WriteFile(target, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	if from != "" && changed != 1 {
		test.Fatalf("mutant anchor count: %d for %q", changed, from)
	}
	return directory
}

// Not parallel: shares the CPU with interleaved timing samples.
func TestLegacyMutants(test *testing.T) {
	path := manifest(test, generated(test))
	oracle := lintGoOracleProduct(test)
	want := execute(test, "", oracle, "--manifest", path).output
	for _, change := range []struct{ name, from, to string }{
		{"overlap winner misreported", "${winner} overlaps another fix", "${this.selected} overlaps another fix"},
	} {
		test.Run(change.name, func(test *testing.T) {
			directory := mutant(test, change.from, change.to)
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", node(test, directory, path, false)},
				{"emitted JavaScript", emittedNode(test, directory, path, false)},
			} {
				if bytes.Equal(side.run.output, want) {
					test.Fatalf("%s mutant survived on %s", change.name, side.name)
				}
				test.Logf("%s caught on %s: %s", change.name, side.name, difference(side.run.output, want))
			}
		})
	}
}

func TestCountGuardMutant(test *testing.T) {
	test.Parallel()
	source := filepath.Join(test.TempDir(), "count.ts")
	if writeError := os.WriteFile(source, []byte("var a; debugger;"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	path := manifest(test, []string{source})
	oracle := lintGoOracleProduct(test)
	answer := execute(test, "", oracle, "--manifest", path).output
	count := execute(test, "", oracle, "--manifest", path, "--count").output
	directory := mutant(
		test,
		"if(countOnly) {\n        return linter.findings.length;",
		"if(countOnly) {\n        return linter.findings.length + 1;",
		"main.ts",
	)
	for _, side := range []struct {
		name            string
		ordinary, count execution
	}{
		{"Node", node(test, directory, path, false), node(test, directory, path, true)},
		{"emitted JavaScript", emittedNode(test, directory, path, false), emittedNode(test, directory, path, true)},
	} {
		if !bytes.Equal(side.ordinary.output, answer) {
			test.Fatalf("%s mutant was caught outside the count check", side.name)
		}
		if bytes.Equal(side.count.output, count) {
			test.Fatalf("%s count mutant survived", side.name)
		}
		test.Logf(
			"count check alone caught %s: mutant=%q Go=%q; ordinary output remains identical",
			side.name,
			side.count.output,
			count,
		)
	}
}

// Not parallel: interleaved timing samples must not compete with tests in this package.
func TestThroughput(test *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		test.Skip("set ADAMIC_LINT_BENCH=1")
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Fatal("missing pinned compiler checkout")
	}
	var rows []string
	walkError := filepath.WalkDir(filepath.Join(
		source,
		"src/compiler",
	), func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if !entry.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a")) {
			rows = append(rows, path)
		}
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	path := manifest(test, rows)
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	oracle := lintGoOracleProduct(test)
	binary := buildPort(test, directory, false)
	machine := execute(test, "", "uname", "-a")
	cpu, readError := os.ReadFile("/proc/cpuinfo")
	if readError != nil {
		test.Fatal(readError)
	}
	model := ""
	for _, line := range strings.Split(string(cpu), "\n") {
		if strings.HasPrefix(line, "model name") {
			model = line
			break
		}
	}
	loadBefore, _ := os.ReadFile("/proc/loadavg")
	test.Logf(
		"machine %s; %s; load before %s",
		strings.TrimSpace(string(machine.output)),
		model,
		strings.TrimSpace(string(loadBefore)),
	)
	best := map[string]time.Duration{}
	var want []byte
	for round := 0; round < 5; round++ {
		for _, name := range []string{"Go", "native", "Node"} {
			var result execution
			switch name {
			case "Go":
				result = execute(test, "", oracle, "--manifest", path, "--count")
			case "native":
				result = execute(test, "", binary, "--manifest", path, "--count")
			case "Node":
				result = node(test, directory, path, true)
			}
			if want == nil {
				want = result.output
			}
			if !bytes.Equal(want, result.output) {
				test.Fatalf("count differs on %s: %q vs %q", name, result.output, want)
			}
			if best[name] == 0 || result.duration < best[name] {
				best[name] = result.duration
			}
			test.Logf(
				"round %d %s %s findings=%s",
				round+1,
				name,
				result.duration,
				strings.TrimSpace(string(result.output)),
			)
		}
	}
	var count int
	if _, scanError := fmt.Sscan(string(want), &count); scanError != nil {
		test.Fatal(scanError)
	}
	for _, name := range []string{"Go", "native", "Node"} {
		test.Logf(
			"best of 5 %s: %.6fs, %.2f findings/s (%d files, %d findings)",
			name,
			best[name].Seconds(),
			float64(count)/best[name].Seconds(),
			len(rows),
			count,
		)
	}
	loadAfter, _ := os.ReadFile("/proc/loadavg")
	test.Logf("load after %s", strings.TrimSpace(string(loadAfter)))
}

// One edited JSX rule is the native canary. Semantic mutations are held to Go on
// Node and emitted JavaScript; this copy also proves sanitized native matches the
// mutated Node result, including JSX parsing, text spans and finding serialization.
const nativeCanaryRule = "react/jsx-no-comment-textnodes"

var portImport = regexp.MustCompile(`(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])`)

// Preserve imports within the copied port, and resolve outside imports from their
// original module directory. Rules may be arbitrarily deeper than context.ts.
func rewritePortImports(test *testing.T, file, source string) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return portImport.ReplaceAllStringFunc(source, func(declaration string) string {
		parts := portImport.FindStringSubmatch(declaration)
		if !strings.HasPrefix(parts[3], ".") {
			return declaration
		}
		absolute := filepath.Clean(filepath.Join(root, filepath.Dir(file), parts[3]))
		relative, relativePathError := filepath.Rel(root, absolute)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return declaration
		}
		return parts[1] + parts[2] + filepath.ToSlash(absolute) + parts[4]
	})
}

func emittedJavaScript(test *testing.T, directory string) string {
	test.Helper()
	if isPackage(directory) {
		value := shared("emitted JavaScript", func(value *sharedValue) {
			value.path, value.err = sharedPath("lint.mjs")
			if value.err == nil {
				value.err = emitJavaScriptTo(directory, value.path)
			}
		})
		if value.err != nil {
			test.Fatal(value.err)
		}
		return value.path
	}
	path := filepath.Join(test.TempDir(), "lint.mjs")
	if emitJavaScriptToError := emitJavaScriptTo(directory, path); emitJavaScriptToError != nil {
		test.Fatal(emitJavaScriptToError)
	}
	return path
}
func emittedNode(test *testing.T, directory, manifest string, count bool) execution {
	test.Helper()
	return runJavaScript(test, emittedJavaScript(test, directory), manifest, count)
}
func runJavaScript(test *testing.T, module, manifest string, count bool) execution {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, module, "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return execute(test, "", "node", args...)
}
