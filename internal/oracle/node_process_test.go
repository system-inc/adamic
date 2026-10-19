package oracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"host", "performance", "performance_core"} {
		var arguments []string
		if name == "host" {
			arguments = []string{"plain", "", "with space", "héllo 🌍", "--prof", "bad\xff"}
		}
		inputFixtures = append(inputFixtures, struct {
			path       string
			arguments  []string
			unreadable bool
			writes     bool
		}{"internal/oracle/testdata/node_process_" + name + ".a", arguments, false, false})
	}
}

func nodeProcessTerminal(t *testing.T, destination string, command ...string) run {
	t.Helper()
	result := execute(t, "python3", append([]string{filepath.Join(repository, "internal/oracle/testdata/node_process_terminal.py"), destination}, command...)...)
	if result.exitCode != 0 {
		t.Fatalf("terminal driver: %s", result.stderr)
	}
	var observed struct {
		Stdout, Stderr []byte
		ExitCode       int
	}
	if err := json.Unmarshal(result.stdout, &observed); err != nil {
		t.Fatal(err)
	}
	return run{stdout: observed.Stdout, stderr: observed.Stderr, exitCode: observed.ExitCode}
}

func TestNodeProcessTerminal(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_process_terminal.a")
	runner := filepath.Join(repository, "oracle/node.mjs")
	for _, destination := range []string{"file", "pipe", "tty"} {
		truth := nodeProcessTerminal(t, destination, "node", "--disable-warning=ExperimentalWarning", runner, path)
		for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
			got := nodeProcessTerminal(t, destination, command...)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s: %s\nNode %q\ngot %q %q", destination, difference, truth.stdout, got.stdout, got.stderr)
			}
		}
		t.Logf("%s: Node %q", destination, truth.stdout)
	}
}

func TestNodeProcessMissingMark(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_process_missing_mark.a")
	truth := onNode(t, path)
	if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), `The "drop" performance mark has not been set`) {
		t.Fatalf("unexpected Node result %d %s", truth.exitCode, truth.stderr)
	}
	for _, got := range []run{execute(t, binary), onNode(t, script)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s", difference, got.stderr)
		}
	}
}

// Renaming the unit's exported symbols makes a second runtime in this translation unit.
// Each mutant compiles and runs under sanitizers; only its Node comparison may reject it.
func nodeProcessMutant(t *testing.T, fixture, before, after string) (string, string) {
	t.Helper()
	path := filepath.Join(repository, "internal/oracle/testdata", fixture)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if strings.Count(source, before) != 1 {
		t.Fatalf("mutant anchor %q occurs %d times", before, strings.Count(source, before))
	}
	source = strings.Replace(source, before, after, 1)
	source = strings.ReplaceAll(source, "adamic_node_", "mutant_node_")
	code := strings.ReplaceAll(native.C(program), "adamic_node_", "mutant_node_")
	code = strings.Replace(code, "adamic_start(argc, argv);", "adamic_start(argc, argv); mutant_node_process_start(argc, argv);", 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source+"\n"+code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return path, binary
}

func TestNodeProcessMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, fixture, before, after string }{
		{"cwd", "node_process_host.a", "current_directory = result;\n    return adamic_retain(result);", "adamic_release(result);\n    return adamic_node_platform();"},
		{"platform", "node_process_host.a", `ADAMIC_STRING("linux")`, `ADAMIC_STRING("wrong")`},
		{"EOL", "node_process_host.a", "return adamic_host_eol();", "static adamic_string newline = ADAMIC_STRING(\"\\r\\n\"); return &newline;"},
		{"nextTick feature", "node_process_host.a", "bool adamic_node_next_tick_feature(void) { return true; }", "bool adamic_node_next_tick_feature(void) { return false; }"},
		{"pid range", "node_process_host.a", "return (double)getpid();", "return 0;"},
		{"argv", "node_process_host.a", "arguments = adamic_host_argv();", "arguments = adamic_host_exec_argv();"},
		{"execArgv", "node_process_host.a", "execution_arguments = adamic_host_exec_argv();", "execution_arguments = adamic_host_exec_argv(); static adamic_string flag = ADAMIC_STRING(\"--prof\"); adamic_array_push(execution_arguments, (adamic_value){.reference = &flag});"},
		{"columns", "node_process_terminal.a", "result.number = size.ws_col;", "result.number = 80;"},
		{"handle feature", "node_process_terminal.a", " || S_ISREG(info.st_mode)", " || !S_ISREG(info.st_mode)"},
		{"write", "node_process_host.a", "return adamic_write_raw(adamic_stdout, text);", "adamic_write_line(adamic_stdout, text); return true;"},
		{"memory range", "node_process_host.a", "result->slots[0].number = observations.heapUsed;", "result->slots[0].number = observations.heapUsed - observations.heapUsed - 1;"},
		{"monotonic now", "node_process_performance.a", "return adamic_host_performance_now();", "return -adamic_host_performance_now();"},
		{"time origin range", "node_process_performance.a", "return adamic_host_time_origin();", "return -adamic_host_time_origin();"},
		{"performance hooks", "node_process_performance_core.a", "performance_object->slots[0].number = adamic_host_time_origin();", "performance_object->slots[0].number = 0;"},
		{"mark", "node_process_performance.a", "return entry(name, false, now, 0);", "return entry(name, false, now, 1);"},
		{"measure", "node_process_performance.a", "finish - begin", "begin - finish"},
		{"clear all marks", "node_process_cleared_marks.a", "name == NULL || (record->length", "name != NULL && (record->length"},
		{"clear named marks", "node_process_missing_mark.a", "void adamic_node_clear_marks(const adamic_string *name) { clear_marks(name); }", "void adamic_node_clear_marks(const adamic_string *name) { (void)name; }"},
		{"clear measures", "node_process_performance.a", "void adamic_node_clear_measures(const adamic_string *name) { (void)name; }", "void adamic_node_clear_measures(const adamic_string *name) { clear_marks(name); }"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			if mutant.name == "platform" && runtime.GOOS == "darwin" {
				mutant.before = `ADAMIC_STRING("darwin")`
			}
			path, binary := nodeProcessMutant(t, mutant.fixture, mutant.before, mutant.after)
			if mutant.fixture == "node_process_terminal.a" {
				truth := nodeProcessTerminal(t, "tty", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
				bad := nodeProcessTerminal(t, "tty", binary)
				if bad.exitCode != truth.exitCode || string(bad.stderr) != string(truth.stderr) || disagreement(truth, bad) != "stdout differs" {
					t.Fatalf("mutant not caught only by stdout: %d %q %q", bad.exitCode, bad.stdout, bad.stderr)
				}
			} else {
				absolute, err := filepath.Abs(path)
				if err != nil {
					t.Fatal(err)
				}
				how := inputRun{directory: filepath.Dir(absolute), arguments: []string{"plain", "", "with space", "héllo 🌍", "--prof", "bad\xff"}}
				truth := onNodeWith(t, how, absolute)
				bad := executeInput(t, how, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, how.arguments...)
				difference := disagreement(truth, bad)
				expected := "stdout differs"
				if mutant.fixture == "node_process_missing_mark.a" {
					expected = "stderr differs"
				}
				if mutant.name == "clear measures" || mutant.name == "clear all marks" {
					expected = "exit codes differ"
				}
				if difference != expected || strings.Contains(string(bad.stderr), "Sanitizer") {
					t.Fatalf("mutant not caught only by Node bytes: %s, exit %d vs %d\nNode %q %q\nbad %q %q", difference, bad.exitCode, truth.exitCode, truth.stdout, truth.stderr, bad.stdout, bad.stderr)
				}
			}
			t.Log("mutant compiled and ran without sanitizer errors; Node comparison caught it")
		})
	}
}

func TestNodeProcessUnicodeDirectory(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_host.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	directory := filepath.Join(shared, "héllo 🌍")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: directory, arguments: []string{"real directory", "bad\xff"}}
	truth := onNodeWith(t, how, path)
	backend := inputBackend(t, how, program, shared)
	got, binary := inputNatively(t, how, program, shared)
	for _, result := range []run{backend, got} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatal(difference)
		}
	}
	if report := inputLeaks(t, func() inputRun { return how }, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestNodeProcessBlocking(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux F_SETPIPE_SZ and inherited nonblocking pipe semantics")
	}
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_process_blocking.a")
	runner := filepath.Join(repository, "oracle/node.mjs")
	truth := nodeProcessTerminal(t, "backpressure", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if truth.exitCode != 0 || len(truth.stdout) != 200000 || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node blocking result %d %d %s", truth.exitCode, len(truth.stdout), truth.stderr)
	}
	for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
		if difference := disagreement(truth, nodeProcessTerminal(t, "backpressure", command...)); difference != "" {
			t.Fatal(difference)
		}
	}
	// Main retries EAGAIN, so both blocking and nonblocking descriptors can drain
	// every byte. Observe the descriptor itself so retrying cannot mask a no-op.
	flagsSource := `const fs = await import('node:fs');
function flags() {
 const line = fs.readFileSync('/proc/self/fdinfo/1','utf8').split('\n').find(line => line.startsWith('flags:'));
 if (line === undefined) throw Error('missing descriptor flags');
 return Number.parseInt(line.slice(6).trim(), 8);
}
const before = (flags() & 2048) !== 0;
process.stdout._handle.setBlocking(true);
console.log(before + ' ' + ((flags() & 2048) !== 0));`
	flagsTruth := nodeProcessTerminal(t, "backpressure", "node", "--input-type=module", "-e", flagsSource)
	if flagsTruth.exitCode != 0 || len(flagsTruth.stderr) != 0 || string(flagsTruth.stdout) != "true false\n" {
		t.Fatalf("Node must clear an inherited nonblocking flag: %+v", flagsTruth)
	}
	code := `#include "adamic.h"
#include <fcntl.h>
int main(void) {
 adamic_start(0, NULL);
 if ((fcntl(1, F_GETFL) & O_NONBLOCK) == 0) return 2;
 adamic_object *handle = adamic_node_stdout_handle();
 if (handle == NULL) return 3;
 adamic_closure *method = handle->slots[0].reference;
 adamic_value arguments[] = {{.boolean = true}};
 (void)method->code(method, arguments);
 int flags = fcntl(1, F_GETFL);
 adamic_release(handle);
 if (flags < 0) return 4;
 static adamic_string blocked = ADAMIC_STRING("true false"), unchanged = ADAMIC_STRING("true true");
 adamic_write_line(adamic_stdout, (flags & O_NONBLOCK) != 0 ? &unchanged : &blocked);
 return 0;
}`
	for _, mutant := range []bool{false, true} {
		implementation := code + "\n"
		if mutant {
			data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
			if err != nil {
				t.Fatal(err)
			}
			anchor := "(void)fcntl(STDOUT_FILENO, F_SETFL, flags);"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("blocking mutation anchor changed")
			}
			changed := strings.Replace(string(data), anchor, "(void)flags;", 1)
			implementation = strings.ReplaceAll(changed+"\n"+code+"\n", "adamic_node_", "blocking_probe_node_")
		}
		probe := filepath.Join(t.TempDir(), "flags")
		if err := native.Build(implementation, probe, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := nodeProcessTerminal(t, "backpressure", probe)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("descriptor probe must run cleanly: %+v", got)
		}
		difference := disagreement(flagsTruth, got)
		if !mutant && difference != "" {
			t.Fatalf("blocking descriptor control: %s", difference)
		}
		if mutant && difference != "stdout differs" {
			t.Fatalf("Node alone must catch unchanged descriptor flags: %q", difference)
		}
	}
	t.Log("setBlocking no-op ran cleanly under ASan/UBSan; only Node's descriptor flags caught it; the 200 KB controls still drain completely")

}

func TestNodeProcessCachedDirectory(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_process_cached_cwd.a")
	_, mutant := nodeProcessMutant(t, "node_process_cached_cwd.a", "if (current_directory != NULL) { return adamic_retain(current_directory); }", "if (false) { return adamic_retain(current_directory); }")
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_cached_cwd.py"))
	if err != nil {
		t.Fatal(err)
	}
	observe := func(command ...string) run {
		result := execute(t, "python3", append([]string{driver}, command...)...)
		if result.exitCode != 0 {
			t.Fatalf("cwd driver: %s", result.stderr)
		}
		var observed struct {
			Stdout, Stderr []byte
			ExitCode       int
		}
		if err := json.Unmarshal(result.stdout, &observed); err != nil {
			t.Fatal(err)
		}
		return run{stdout: observed.Stdout, stderr: observed.Stderr, exitCode: observed.ExitCode}
	}
	truth := observe("node", "--disable-warning=ExperimentalWarning", runner, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\ncached true\n" {
		t.Fatalf("unexpected Node cached cwd: %d %q %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
		if difference := disagreement(truth, observe(command...)); difference != "" {
			t.Fatal(difference)
		}
	}
	bad := observe(mutant)
	if bad.exitCode != 70 || !strings.Contains(string(bad.stderr), "ENOENT") || strings.Contains(string(bad.stderr), "Sanitizer") {
		t.Fatal("uncached cwd mutant did not fail only Node comparison")
	}
	t.Log("uncached cwd mutant compiled and ran without sanitizer errors; Node keeps its cached path, mutant panics with ENOENT")
}

func TestNodeProcessUnsupportedOperationsStayLoud(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"performance entries": "import { performance } from 'node:perf_hooks'; performance.getEntries();",
		"measure options":     "import { performance } from 'node:perf_hooks'; performance.measure('x', {start: 0});",
		"write backpressure":  "import 'node:process'; const accepted = process.stdout.write('x'); console.log(String(accepted));",
		"nextTick callback":   "import 'node:process'; process.nextTick(() => { console.log('tick'); });",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "node_process_unsupported.a")
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var missing *lower.NotYet
			if !errors.As(err, &missing) {
				t.Fatalf("must fail before C generation with NotYet, got %v", err)
			}
			t.Log(missing)
		})
	}
}

// A raw write must not put a regular file into pipe-buffer mode. fstat reads
// the real descriptor without invoking an Adamic operation that would flush it.
func TestNodeProcessRawWriteFileModeMutant(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := `import fs from 'node:fs'; process.stdout.write('raw|'); console.log('line'); console.log(String(fs.fstatSync(1).size));`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	observe := func(name string, arguments ...string) run {
		file, err := os.Create(filepath.Join(t.TempDir(), "stdout"))
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		command := bounded(t, name, arguments...)
		var stderr bytes.Buffer
		command.Stdout, command.Stderr = file, &stderr
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		if err := command.Run(); err != nil {
			t.Fatalf("file observation must finish cleanly: %v\n%s", err, stderr.Bytes())
		}
		if stderr.Len() != 0 {
			t.Fatalf("file observation must have clean stderr: %s", stderr.Bytes())
		}
		data, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatal(err)
		}
		return run{stdout: data, stderr: stderr.Bytes()}
	}
	truth := observe("node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	if string(truth.stdout) != "raw|line\n9\n" || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node file observation: %+v", truth)
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/adamic.c"))
	if err != nil {
		t.Fatal(err)
	}
	runtimeSource := string(data)
	prefix := ""
	for _, name := range []string{"adamic_start", "adamic_write_line", "adamic_write_raw", "adamic_output_flush", "adamic_panic", "adamic_unreachable", "adamic_process_exit_now"} {
		prefix += "#define " + name + " raw_file_probe_" + name + "\n"
	}
	harness := `
int main(void) {
 adamic_start(0, NULL);
 static adamic_string raw = ADAMIC_STRING("raw|"), line = ADAMIC_STRING("line");
 (void)adamic_write_raw(adamic_stdout, &raw);
 adamic_write_line(adamic_stdout, &line);
 struct stat information;
 if (fstat(1, &information) != 0) return 1;
 adamic_string *size = adamic_string_from_number((double)information.st_size);
 adamic_write_line(adamic_stdout, size);
 adamic_release(size);
 return 0;
}
`
	for _, mutant := range []bool{false, true} {
		implementation := runtimeSource
		if mutant {
			start := strings.Index(implementation, "void adamic_write_line(")
			before := "output_mode = isatty(adamic_stdout) || regular_file ? 2 : 1;"
			if start < 0 || strings.Count(implementation[start:], before) != 1 {
				t.Fatal("raw file-mode mutation anchor changed")
			}
			implementation = implementation[:start] + strings.Replace(implementation[start:], before, "output_mode = isatty(adamic_stdout) ? 2 : 1; (void)regular_file;", 1)
		}
		binary := filepath.Join(t.TempDir(), "probe")
		if err := native.Build(prefix+implementation+harness, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := observe(binary)
		difference := disagreement(truth, got)
		if !mutant && difference != "" {
			t.Fatalf("raw file control: %s", difference)
		}
		if mutant && difference != "stdout differs" {
			t.Fatalf("Node alone must catch raw pipe mode: %q", difference)
		}
	}
	t.Log("raw file-mode mutant exited 0 with clean sanitizers and leaks; only Node's descriptor-size observation caught it")
}
