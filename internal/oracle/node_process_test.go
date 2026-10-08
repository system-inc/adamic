package oracle

import (
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
		{"cwd", "node_process_host.a", "strlen(buffer));\n            free(buffer);\n            current_directory = result;\n            return adamic_retain(result);", "strlen(buffer));\n            free(buffer);\n            adamic_release(result);\n            return adamic_node_platform();"},
		{"platform", "node_process_host.a", `ADAMIC_STRING("linux")`, `ADAMIC_STRING("wrong")`},
		{"EOL", "node_process_host.a", `ADAMIC_STRING("\n")`, `ADAMIC_STRING("\r\n")`},
		{"nextTick feature", "node_process_host.a", "bool adamic_node_next_tick_feature(void) { return true; }", "bool adamic_node_next_tick_feature(void) { return false; }"},
		{"pid range", "node_process_host.a", "return (double)getpid();", "return 0;"},
		{"argv", "node_process_host.a", "index < saved_count; index++", "index < saved_count - 1; index++"},
		{"execArgv", "node_process_host.a", "execution_arguments = adamic_array_new(0, true);", "execution_arguments = adamic_array_new(1, true); static adamic_string flag = ADAMIC_STRING(\"--prof\"); adamic_array_push(execution_arguments, (adamic_value){.reference = &flag});"},
		{"columns", "node_process_terminal.a", "result.number = size.ws_col;", "result.number = 80;"},
		{"handle feature", "node_process_terminal.a", " || S_ISREG(info.st_mode)", " || !S_ISREG(info.st_mode)"},
		{"write", "node_process_host.a", "return adamic_write_raw(adamic_stdout, text);", "adamic_write_line(adamic_stdout, text); return true;"},
		{"memory range", "node_process_host.a", "result->slots[0].number = used;", "result->slots[0].number = used - used - 1;"},
		{"monotonic now", "node_process_performance.a", "milliseconds(CLOCK_MONOTONIC) - monotonic_origin", "monotonic_origin - milliseconds(CLOCK_MONOTONIC)"},
		{"time origin range", "node_process_performance.a", "return epoch_origin;", "return epoch_origin - epoch_origin - 1;"},
		{"performance hooks", "node_process_performance_core.a", "performance_object->slots[0].number = epoch_origin;", "performance_object->slots[0].number = 0;"},
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
	if runtime.GOOS != "linux" {
		t.Skip("Linux F_SETPIPE_SZ and inherited nonblocking pipe semantics")
	}
	t.Parallel()
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
	_, mutant := nodeProcessMutant(t, "node_process_blocking.a", "(void)fcntl(STDOUT_FILENO, F_SETFL, flags);", "(void)flags;")
	bad := nodeProcessTerminal(t, "backpressure", mutant)
	if len(bad.stdout) >= len(truth.stdout) || len(bad.stderr) != 0 || disagreement(truth, bad) == "" {
		t.Fatal("blocking mutation survived, or failed outside Node comparison")
	}
	t.Logf("setBlocking mutant runs without sanitizer errors; Node has %d bytes and exit %d, mutant has %d bytes and exit %d", len(truth.stdout), truth.exitCode, len(bad.stdout), bad.exitCode)
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
