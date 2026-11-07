package native_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// records_define_hidden.a is Object.defineProperty's hidden record key, as tsc's commandLineParser hides
// options.configFile, with the attributes Node gives it: held by graphCompilerWitness to source Node, the
// JavaScript backend, native under ASan and UBSan, and the leak check.
func recordDefineHiddenSource(test *testing.T) string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join("testdata", "records_define_hidden.a"))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

func TestRecordDefineHiddenAgainstNode(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, recordDefineHiddenSource(test), nil)
}

// An entries record, kept as a record but seen through fixed views, written through one under a name a define
// hides on another record. No program on main drives a defined record into a fixed object's write (a record seen
// as a fixed view is refused, through a union it's not yet, and a generic write is the record's own), so this is
// the record that reaches writeProperty's record branch, which takes the checked write there, and stores.
const recordHiddenBesideEntries = `const source: { visible: number; [key: string]: number | string } = { visible: 1 };
const alias = source;
alias['late'] = 2;
alias['visible'] = 3;
function show(view: { visible: number }): void {
    view.visible = 4;
}
show(source);
console.log(` + "`${Object.keys(source).join(',')} ${String(source.visible)}`" + `);
const hidden: Record<string, number> = {};
Object.defineProperty(hidden, 'visible', { enumerable: false, writable: false, value: 1 });
console.log(` + "`${String(hidden.visible)} ${String(Object.keys(hidden).length)}`" + `);
`

func TestRecordDefineHiddenBesideEntriesRecords(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, recordHiddenBesideEntries, nil)
}

var (
	recordWriteCheck = regexp.MustCompile(
		`if \(!adamic_record_set_checked(\(.*\))\) adamic_record_refused\(.*, false\);`,
	)
	recordDeleteCheck = regexp.MustCompile(`if \(!\w+\) adamic_record_refused\(.*, true\);`)
)

// Mutant: the write and delete checks dropped from the C, so a write or a delete of a hidden key is refused
// without a throw, while the define's own check, which is IR, stays. Node's output must catch it.
func TestRecordDefineHiddenWriteChecksMutant(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, recordDefineHiddenSource(test), func(code string) string {
		return recordDeleteCheck.ReplaceAllString(recordWriteCheck.ReplaceAllString(code, "adamic_record_set$1;"), "")
	})
}

// recordUncaughtWitness holds a program that ends on a throw nothing catches to source Node, which
// oracle/node.mjs reports as Adamic's panic: String(error) on one stderr line, and exit 70. The JavaScript
// backend and native under ASan and UBSan must exit with Node's code and write its stdout and stderr. A
// program that ends so holds what it held, which isn't a leak, so leak detection is off, as in the oracle's
// own comparison. Given a mutant of the C, native must disagree with Node instead.
func recordUncaughtWitness(test *testing.T, source string, mutant func(string) string) {
	test.Helper()
	directory := test.TempDir()
	path := filepath.Join(directory, "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	type ran struct {
		stdout, stderr []byte
		exitCode       int
	}
	run := func(environment []string, name string, arguments ...string) ran {
		timeoutContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		command := exec.CommandContext(timeoutContext, name, arguments...)
		command.Env = append(os.Environ(), environment...)
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		var exited *exec.ExitError
		if runError := command.Run(); runError != nil && !errors.As(runError, &exited) {
			test.Fatalf("running %s: %v", name, runError)
		}
		return ran{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
	}
	truth := run(nil, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	if truth.exitCode != 70 {
		test.Fatalf(
			"source Node did not end on an uncaught throw: exit %d\n%s\n%s",
			truth.exitCode,
			truth.stdout,
			truth.stderr,
		)
	}
	agrees := func(got ran) bool {
		return got.exitCode == truth.exitCode &&
			bytes.Equal(got.stdout, truth.stdout) &&
			bytes.Equal(got.stderr, truth.stderr)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerWithOptionsError := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	js := filepath.Join(directory, "main.mjs")
	if writeError := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	if got := run(nil, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js); !agrees(got) {
		test.Fatalf(
			"JavaScript: exit %d %q %q, Node exit %d %q %q",
			got.exitCode,
			got.stdout,
			got.stderr,
			truth.exitCode,
			truth.stdout,
			truth.stderr,
		)
	}
	code := native.C(program)
	if mutant != nil {
		changed := mutant(code)
		if changed == code {
			test.Fatal("mutant did not change C")
		}
		code = changed
	}
	binary := filepath.Join(directory, "native")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got := run([]string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if mutant != nil {
		if agrees(got) {
			test.Fatal("mutant survived: native agrees with Node")
		}
		test.Logf(
			"caught: native exit %d %q %q, Node exit %d %q %q",
			got.exitCode,
			got.stdout,
			got.stderr,
			truth.exitCode,
			truth.stdout,
			truth.stderr,
		)
		return
	}
	if !agrees(got) {
		test.Fatalf(
			"native: exit %d %q %q, Node exit %d %q %q",
			got.exitCode,
			got.stdout,
			got.stderr,
			truth.exitCode,
			truth.stdout,
			truth.stderr,
		)
	}
}

// A write to a hidden key that nothing catches: Node ends on its TypeError. The write is in a function, through
// a key known only as it runs, so the throw leaves the function, and main's call to it is what tests for it
// (RecordWriteMayThrow, through lowering's exceptions fixpoint) and ends the program as Node does.
const recordHiddenUncaught = `const settings: Record<string, number> = { size: 1 };
Object.defineProperty(settings, 'locked', { enumerable: false, writable: false, value: 2 });
function update(key: string, value: number): void {
    settings[key] = value;
}
update('size', 3);
console.log(` + "`${String(settings.size)} ${String(settings.locked)}`" + `);
update(['loc', 'ked'].join(''), 4);
console.log('not reached');
`

func TestRecordDefineHiddenUncaughtAgainstNode(test *testing.T) {
	test.Parallel()
	recordUncaughtWitness(test, recordHiddenUncaught, nil)
}

// Mutant: the uncaught write's check dropped from the C, so the write is refused without a throw and the
// program runs on. Node's exit code and stderr must catch it.
func TestRecordDefineHiddenUncaughtMutant(test *testing.T) {
	test.Parallel()
	recordUncaughtWitness(test, recordHiddenUncaught, func(code string) string {
		return recordWriteCheck.ReplaceAllString(code, "adamic_record_set$1;")
	})
}

// A define in main.a, and a write through a key known only as it runs in a function library.a exports and
// main.a calls, inside a try. Imports are lowered before the module that imports them, so the write is lowered
// before the define names its key: it's the exceptions fixpoint, once every module is lowered, that makes the
// library's function one that can throw, and main's call one that tests for it.
const recordHiddenLibrary = `export function store(record: Record<string, number>, key: string, value: number): void {
    record[key] = value;
}
`

const recordHiddenAcrossModules = `import { store } from './library.a';
const counts: Record<string, number> = { a: 1 };
Object.defineProperty(counts, 'hidden', { enumerable: false, writable: false, value: 5 });
store(counts, 'a', 2);
try {
    store(counts, ['hid', 'den'].join(''), 6);
    console.log('written');
}
catch(error) {
    if(error instanceof Error) console.log(` + "`${error.name}: ${error.message}`" + `);
}
console.log(` + "`${String(counts.hidden)} ${JSON.stringify(counts)}`" + `);
`

func TestRecordDefineHiddenAcrossModulesAgainstNode(test *testing.T) {
	test.Parallel()
	graphCompilerModulesWitness(
		test,
		recordHiddenAcrossModules,
		map[string]string{"library.a": recordHiddenLibrary},
		nil,
	)
}

// A record write or delete is checked only when its key may be hidden: one known only as it runs, or a literal
// a define names. A literal no define names emits the bare call, and without a define nothing is checked.
func TestRecordWriteChecksFollowHiddenKeys(test *testing.T) {
	test.Parallel()
	const writes = `const r: Record<string, number> = {};
const key = ['a', 'b'].join('');
r.plain = 1;
r[key] = 2;
delete r[key];
delete r.plain;
r.other ??= 3;
`
	const define = `Object.defineProperty(r, 'hidden', { enumerable: false, writable: false, value: 5 });
try {
    r.hidden = 6;
}
catch(error) {
    if(error instanceof Error) console.log(error.message);
}
`
	for _, probe := range []struct {
		name, source    string
		checks, defines int
	}{
		{"no define", writes + "console.log(String(JSON.stringify(r)));\n", 0, 0},
		// r[key] and delete r[key] are dynamic, and r.hidden names the hidden key; r.plain, delete r.plain and
		// r.other ??= name none.
		{"one define", writes + define + "console.log(String(JSON.stringify(r)));\n", 3, 1},
		// view.visible and alias['visible'] are writeProperty's, each checked in its record branch.
		{"entries record under a hidden name", recordHiddenBesideEntries, 2, 1},
	} {
		test.Run(probe.name, func(test *testing.T) {
			path := filepath.Join(test.TempDir(), "main.a")
			if writeError := os.WriteFile(path, []byte(probe.source), 0600); writeError != nil {
				test.Fatal(writeError)
			}
			checked, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			program, lowerWithOptionsError := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
			if lowerWithOptionsError != nil {
				test.Fatal(lowerWithOptionsError)
			}
			code := native.C(program)
			if got := strings.Count(code, "adamic_record_refused("); got != probe.checks {
				test.Errorf("%d checked writes and deletes, want %d", got, probe.checks)
			}
			if got := strings.Count(code, "adamic_record_define_hidden("); got != probe.defines {
				test.Errorf("%d defines, want %d", got, probe.defines)
			}
		})
	}
}
