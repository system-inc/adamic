package lower

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// firstLinkTwoFunctions calls a function that prints, then one whose body stops today: Math.random is
// NotYet (refusedRandom), at main.a:5:9.
const firstLinkTwoFunctions = "function first(): void {\n\tconsole.log('first ran');\n}\nfunction second(): number {\n\treturn Math.random();\n}\nfirst();\nconsole.log(`${second()}`);\n"

// firstLinkNested is a function whose own work prints before it calls a nested function whose body stops
// today, at main.a:4:10.
const firstLinkNested = "function outer(): number {\n\tconsole.log('outer ran');\n\tfunction inner(): number {\n\t\treturn Math.random();\n\t}\n\treturn inner();\n}\nconsole.log(`${outer()}`);\n"

// firstLinkRefused calls a function that prints, then one holding ==, which the refusal pass refuses, at
// main.a:5:11.
const firstLinkRefused = "function first(): void {\n\tconsole.log('first ran');\n}\nfunction second(n: number): boolean {\n\treturn n == 1;\n}\nfirst();\nconsole.log(`${second(1)}`);\n"

// firstLinkSignature prints, then calls a function that calls one whose signature stops today: a rest
// parameter of a union of arrays is NotYet at main.a:1:17 (TestRestParametersThatArentArraysAreNotYet).
const firstLinkSignature = "function either(...args: string[] | number[]): number {\n\treturn 0;\n}\nfunction run(): void {\n\tconsole.log(`${either('a')}`);\n}\nconsole.log('before the call');\nrun();\n"

// lowerFirstLink lowers a .a program, as a first-link lowering when traps is set.
func lowerFirstLink(t *testing.T, source string, traps bool) (*ir.Program, error) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return LowerWithOptions(context.Background(), program, Options{FirstLinkTraps: traps})
}

// firstLinkRun is one backend's run of a lowered program: what it printed and its exit code.
type firstLinkRun struct {
	backend        string
	code           int
	stdout, stderr string
}

// firstLinkRuns builds a lowered program as sanitized native and on the JavaScript backend, and runs each.
func firstLinkRuns(t *testing.T, program *ir.Program) []firstLinkRun {
	t.Helper()
	directory := t.TempDir()
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	commands := []struct {
		backend string
		command *exec.Cmd
	}{
		{"native", exec.Command(binary)},
		{"javascript", exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)},
	}
	runs := []firstLinkRun{}
	for _, one := range commands {
		one.command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var out, errs bytes.Buffer
		one.command.Stdout, one.command.Stderr = &out, &errs
		code := 0
		if err := one.command.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			code = exit.ExitCode()
		}
		runs = append(runs, firstLinkRun{backend: one.backend, code: code, stdout: out.String(), stderr: errs.String()})
	}
	return runs
}

// firstLinkLocalSites is program's Traps with every site in a Site or a Reason read from its file's base name
// on: load names a file by the path it was loaded from, a test's temporary directory here, which differs on
// every run.
func firstLinkLocalSites(program *ir.Program) []ir.Trap {
	rows := []ir.Trap{}
	for _, row := range program.Traps {
		row.Site = firstLinkLocalSite.ReplaceAllString(row.Site, "$1")
		row.Reason = firstLinkLocalSite.ReplaceAllString(row.Reason, "$1")
		rows = append(rows, row)
	}
	return rows
}

// firstLinkLocalSite is a path ending in a file:line:col site, with the site's file from its base name on.
var firstLinkLocalSite = regexp.MustCompile(`[^\s]*/([^/\s]+:[0-9]+:[0-9]+)`)

// firstLinkFromFile is a stop's text from its site's file's base name on.
func firstLinkFromFile(stop error, file string) string {
	if stop == nil {
		return ""
	}
	text := stop.Error()
	if at := strings.Index(text, file+":"); at >= 0 {
		return text[at:]
	}
	return text
}

// firstLinkTrapped holds a program's runs to its trap: each prints stdout, then exits 75 with stderr
// starting with the first-link prefix and then stop, and naming site.
func firstLinkTrapped(t *testing.T, program *ir.Program, stdout, stop, site string) {
	t.Helper()
	for _, run := range firstLinkRuns(t, program) {
		if run.code != 75 || run.stdout != stdout || !strings.HasPrefix(run.stderr, "adamic: first-link trap: "+stop) || !strings.Contains(run.stderr, site) {
			t.Errorf("%s: exit %d stdout %q stderr %q; want 75, %q, and a first-link %s at %s", run.backend, run.code, run.stdout, run.stderr, stdout, stop, site)
		}
	}
}

// A first-link lowering lowers a body that stops as its trap, and the program runs up to it: the first
// function prints, then the second traps with exit 75, on native and on the JavaScript backend, and the
// program's Traps holds the one row. Mutant: adamic_first_link_trap and firstLinkTrap exit 0, and this
// goes red.
func TestFirstLinkTrapRunsUpToTheStop(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, firstLinkTwoFunctions, true)
	if err != nil {
		t.Fatal(err)
	}
	want := ir.Trap{Kind: "NotYet", Reason: "Math.random", Site: "main.a:5:9", Unit: "main.a second", Phase: "Body"}
	if rows := firstLinkLocalSites(program); len(rows) != 1 || rows[0] != want {
		t.Fatalf("Traps: got %+v, want %+v", rows, want)
	}
	firstLinkTrapped(t, program, "first ran\n", "NotYet: ", "main.a:5:9")
}

// Without the option, the same program stops at lowering with the same NotYet, word for word as Lower stops it.
func TestFirstLinkTrapOffStopsAtCompileTime(t *testing.T) {
	t.Parallel()
	_, err := lowerFirstLink(t, firstLinkTwoFunctions, false)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), "main.a:5:9: stage 0 can't lower Math.random yet") {
		t.Fatalf("got %v, want the NotYet for Math.random at main.a:5:9", err)
	}
	_, stop := lowerSource(t, firstLinkTwoFunctions)
	if stop == nil || firstLinkFromFile(stop, "main.a") != firstLinkFromFile(err, "main.a") {
		t.Fatalf("got %v, want Lower's stop %v", err, stop)
	}
}

// A nested function whose body stops lowers as its own trap alone: the function it's nested in still runs
// its own work before it calls the nested one, which traps.
func TestFirstLinkTrapNestedStubsOnlyItself(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, firstLinkNested, true)
	if err != nil {
		t.Fatal(err)
	}
	want := ir.Trap{Kind: "NotYet", Reason: "Math.random", Site: "main.a:4:10", Unit: "main.a outer.inner", Phase: "Body"}
	if rows := firstLinkLocalSites(program); len(rows) != 1 || rows[0] != want {
		t.Fatalf("Traps: got %+v, want %+v", rows, want)
	}
	firstLinkTrapped(t, program, "outer ran\n", "NotYet: ", "main.a:4:10")
}

// A unit the refusal pass refuses is trapped before anything lowers, and lowers as its stub: the first
// function runs, then the second traps with the refusal's words. Without the option, the program is
// refused at compile time, word for word as Lower refuses it.
func TestFirstLinkTrapRefusedUnitStubs(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, firstLinkRefused, true)
	if err != nil {
		t.Fatal(err)
	}
	want := ir.Trap{Kind: "Refused", Reason: "==", Fix: "use ===, which doesn't coerce", Site: "main.a:5:11", Unit: "main.a second", Phase: "Refusal"}
	if rows := firstLinkLocalSites(program); len(rows) != 1 || rows[0] != want {
		t.Fatalf("Traps: got %+v, want %+v", rows, want)
	}
	firstLinkTrapped(t, program, "first ran\n", "Refused: ==", "main.a:5:11")
	_, off := lowerFirstLink(t, firstLinkRefused, false)
	var refused *Refused
	if !errors.As(off, &refused) || !strings.HasSuffix(off.Error(), "main.a:5:11: Adamic 0.1 refuses ==; use ===, which doesn't coerce") {
		t.Fatalf("without the option: got %v, want the refusal of == at main.a:5:11", off)
	}
	if _, stop := lowerSource(t, firstLinkRefused); stop == nil || firstLinkFromFile(stop, "main.a") != firstLinkFromFile(off, "main.a") {
		t.Fatalf("without the option: got %v, want Lower's stop %v", off, stop)
	}
}

// A module function whose signature stops is trapped at its signature, and a reference to it stops where
// it's met, naming that trap, so the function holding the call traps: what runs before that call runs.
func TestFirstLinkTrapSignatureCascade(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, firstLinkSignature, true)
	if err != nil {
		t.Fatal(err)
	}
	signature := ir.Trap{Kind: "NotYet", Reason: "a rest parameter that isn't an array", Site: "main.a:1:17", Unit: "main.a either", Phase: "Signature"}
	reference := ir.Trap{Kind: "NotYet", Reason: "a reference to either, whose signature trapped: NotYet: a rest parameter that isn't an array at main.a:1:17", Site: "main.a:5:17", Unit: "main.a run", Phase: "Body"}
	if rows := firstLinkLocalSites(program); len(rows) != 2 || rows[0] != signature || rows[1] != reference {
		t.Fatalf("Traps: got %+v, want %+v then %+v", rows, signature, reference)
	}
	firstLinkTrapped(t, program, "before the call\n", "NotYet: a reference to either, whose signature trapped", "main.a:1:17")
}

// A whole-program pass's stop traps its unit in the IR already made: setter.ts stops at checkBoundArguments,
// at 9:15, in the top-level statement that binds (TestBoundNameAccessorsStayStopped's setter). In a first-link
// lowering that pass records the row with its own name as the phase, the statement's IR in Main becomes the
// trap, and the program runs up to it, printing 1, then exits 75.
func TestFirstLinkTrapPassStopRecordsAndTraps(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join("testdata", "bind_name_accessors", "setter.ts"))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := LowerWithOptions(context.Background(), loaded, Options{FirstLinkTraps: true})
	if err != nil {
		t.Fatal(err)
	}
	want := ir.Trap{Kind: "NotYet", Reason: "bind of a function that declares 2 parameters where the type it's bound as names 1", Site: "setter.ts:9:15", Unit: "setter.ts module", Phase: "checkBoundArguments"}
	if rows := firstLinkLocalSites(program); len(rows) == 0 || rows[0] != want {
		t.Fatalf("Traps: got %+v, want %+v first", rows, want)
	}
	firstLinkTrapped(t, program, "1\n", "NotYet: bind of a function that declares 2 parameters", ":9:15")
}

// A judging pass that ranges over a table lowering recorded passes over the entries a trapped unit left there,
// so it judges on past its first stop: two functions each read an optional field a record may reach, which
// checkRecordGuardedReads stops at the read (TestBrandedSupertypeRecordGuardIsRouted's plain row), and in a
// first-link lowering both reads are rows with that pass's phase, each in its own function, and no Gap.
// Mutant: firstLinkSkips answers false, so the pass meets show's read again past its trap, records a Gap and
// skips show2's.
func TestFirstLinkTrapPassRunsPastItsStops(t *testing.T) {
	t.Parallel()
	const records = `const counts: Record<string, number> = {};
counts["a"] = 1;
interface Loose { readonly kind?: number; readonly pos?: number; }
interface Sized extends Declaration { _sizedBrand: undefined; readonly size?: number; readonly label: string | number; }
function isSizedLie(n: Loose): n is Sized { return n !== undefined; }
`
	const body = ` { if (isSizedLie(n)) { console.log(n.size === undefined ? "none" : String(n.size)); console.log(String(n.label)); } }
`
	source := brandedPrelude + records + "function show(n: Loose): void" + body + "function show2(n: Loose): void" + body + "const w = { kind: 80, pos: 2, size: 3, label: \"x\" }; const v: Node = w; show(v);\nshow2(v);\nconsole.log(String(counts[\"a\"]));\n"
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	program, err := LowerWithOptions(context.Background(), loaded, Options{FirstLinkTraps: true})
	if err != nil {
		t.Fatal(err)
	}
	units := []string{}
	for _, row := range firstLinkLocalSites(program) {
		if row.Phase != "checkRecordGuardedReads" {
			continue
		}
		if row.Kind == "Gap" {
			t.Errorf("a Gap row for checkRecordGuardedReads: %+v", row)
			continue
		}
		if row.Kind != "NotYet" || row.Reason != "a union or optional field read in a program with record storage" || row.Echo {
			t.Errorf("a checkRecordGuardedReads row that isn't the read's NotYet: %+v", row)
		}
		units = append(units, row.Unit)
	}
	if len(units) != 2 || units[0] != "main.ts show" || units[1] != "main.ts show2" {
		t.Errorf("checkRecordGuardedReads rows in units %q, want one in main.ts show, then one in main.ts show2; all rows: %+v", units, firstLinkLocalSites(program))
	}
}

// An error that's neither a NotYet nor a Refused is a row of kind "Error" in a first-link lowering, never the
// end of it, and the rows after it are Suspect; a context's cancellation is still returned. No source road
// reaches such an error honestly (each sits behind a checker guarantee or is a compiler bug), so this drives
// a real first-link lowering's catch and pass with one: a body step whose lowering returns it traps its
// function with the error's words at the function, and a pass that returns it records a row with no unit and
// goes on.
func TestFirstLinkTrapPlainErrorIsARow(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", firstLinkTwoFunctions)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	built, err := lowerWithElementKey(&viewElementKey{}, loaded, check, entry, nil, false, false, ModuleFormatEcmaScript, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	first := entry.Statements.Nodes[0]
	stub, trapped, err := built.firstLinkCatch("Body", first, func() error { return errors.New("lower: compiler bug: a plain error") })
	if err != nil || !trapped || !stub.Trap {
		t.Fatalf("catch of a plain error: stub %+v trapped %v err %v; want a trap and no error", stub, trapped, err)
	}
	row := built.traps[len(built.traps)-1]
	row.Site = firstLinkLocalSite.ReplaceAllString(row.Site, "$1")
	if want := (ir.Trap{Kind: "Error", Reason: "lower: compiler bug: a plain error", Site: "main.a:1:1", Unit: "main.a first", Phase: "Body"}); row != want {
		t.Fatalf("the plain error's row: got %+v, want %+v", row, want)
	}
	if err := built.firstLinkPass("checkSomething", false, errors.New("lower: compiler bug: a pass's plain error"), nil); err != nil {
		t.Fatalf("a pass's plain error: got %v, want it recorded and nil", err)
	}
	row = built.traps[len(built.traps)-1]
	if want := (ir.Trap{Kind: "Error", Reason: "lower: compiler bug: a pass's plain error", Phase: "checkSomething", Suspect: true}); row != want {
		t.Fatalf("the pass's plain error's row: got %+v, want %+v, Suspect after the first error", row, want)
	}
	rows := len(built.traps)
	_, trapped, err = built.firstLinkCatch("Body", first, func() error { return fmt.Errorf("lowering stopped: %w", context.Canceled) })
	if trapped || !errors.Is(err, context.Canceled) || len(built.traps) != rows {
		t.Fatalf("a cancellation: trapped %v err %v rows %d, want it returned with no row (%d rows before)", trapped, err, len(built.traps), rows)
	}
}
