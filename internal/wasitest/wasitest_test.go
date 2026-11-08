package wasitest

import (
	"fmt"
	"strings"
	"testing"
)

// A recordingTB is a testing.TB that records what Skipf, Fatalf and Logf are given and lets its caller go on, so a
// test can read what this package did with a test it was handed. Any other method is its nil TB's, and panics.
type recordingTB struct {
	testing.TB
	skipped, failed string
	logged          []string
}

func (recorder *recordingTB) Helper() {}

func (recorder *recordingTB) Skipf(format string, arguments ...any) {
	recorder.skipped = fmt.Sprintf(format, arguments...)
}

func (recorder *recordingTB) Fatalf(format string, arguments ...any) {
	recorder.failed = fmt.Sprintf(format, arguments...)
}

func (recorder *recordingTB) Logf(format string, arguments ...any) {
	recorder.logged = append(recorder.logged, fmt.Sprintf(format, arguments...))
}

// environment is a lookup reading values, as os.Getenv reads the process's, so these tests set no process-wide
// variable and run in parallel.
func environment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

// A test not opted in skips, with a sysroot or without, and is neither failed nor handed a sysroot. Mutant: require
// dropping its opt-in check, so a test not opted in runs whenever WASI_SYSROOT is set.
func TestRequireSkipsATestNotOptedIn(t *testing.T) {
	t.Parallel()
	for _, values := range []map[string]string{{}, {"WASI_SYSROOT": "/sdk/share/wasi-sysroot"}, {"ADAMIC_TEST_WASI": "0", "WASI_SYSROOT": "/sdk/share/wasi-sysroot"}} {
		recorder := &recordingTB{}
		if sysroot := require(recorder, environment(values), "ADAMIC_TEST_WASI"); sysroot != "" {
			t.Errorf("%v: handed sysroot %q to a test not opted in", values, sysroot)
		}
		if recorder.skipped != "set ADAMIC_TEST_WASI=1 and WASI_SYSROOT to run this wasm32-wasi test" || recorder.failed != "" {
			t.Errorf("%v: skipped %q, failed %q; want a skip naming the opt-in and no failure", values, recorder.skipped, recorder.failed)
		}
	}
}

// The ruling's pin: an opted-in test with no sysroot fails, naming what's missing, and doesn't skip. Mutant: require
// skipping when opted in without a sysroot, as runWASIUnit's "WASI toolchain missing" skip did.
func TestRequireFailsAnOptedInTestWithNoSysroot(t *testing.T) {
	t.Parallel()
	recorder := &recordingTB{}
	if sysroot := require(recorder, environment(map[string]string{"ADAMIC_TEST_WASI": "1"}), "ADAMIC_TEST_WASI"); sysroot != "" {
		t.Errorf("handed sysroot %q with WASI_SYSROOT unset", sysroot)
	}
	if recorder.skipped != "" {
		t.Errorf("skipped an opted-in test: %q", recorder.skipped)
	}
	if !strings.HasPrefix(recorder.failed, "ADAMIC_TEST_WASI=1 opts this test into wasm32-wasi, and WASI_SYSROOT is unset") {
		t.Errorf("failed with %q; want a failure naming the opt-in and the unset WASI_SYSROOT", recorder.failed)
	}
}

// An opted-in test with a sysroot is handed it, and neither skipped, failed nor logged. Mutant: require returning
// before it reads WASI_SYSROOT, handing an opted-in test an empty sysroot.
func TestRequireHandsAnOptedInTestItsSysroot(t *testing.T) {
	t.Parallel()
	recorder := &recordingTB{}
	sysroot := require(recorder, environment(map[string]string{"ADAMIC_ORACLE_WASI": "1", "WASI_SYSROOT": "/sdk/share/wasi-sysroot"}), "ADAMIC_ORACLE_WASI")
	if sysroot != "/sdk/share/wasi-sysroot" || recorder.skipped != "" || recorder.failed != "" || len(recorder.logged) != 0 {
		t.Errorf("sysroot %q, skipped %q, failed %q, logged %q; want the sysroot alone", sysroot, recorder.skipped, recorder.failed, recorder.logged)
	}
}

// A leg not opted in is dropped with a line in the log naming it and the opt-in, and its test goes on: it isn't
// skipped, since its other legs still run. Mutants: runsLeg skipping the whole test; runsLeg dropping the leg with
// nothing in the log, as the oracle's opt-in legs with no else did.
func TestLegNotOptedInIsLoggedAndTheTestGoesOn(t *testing.T) {
	t.Parallel()
	recorder := &recordingTB{}
	if runsLeg(recorder, environment(map[string]string{"WASI_SYSROOT": "/sdk/share/wasi-sysroot"}), "ADAMIC_ORACLE_WASI", "sweep.a on wasm32-wasi") {
		t.Error("ran a leg not opted in")
	}
	if want := "wasi leg not run: sweep.a on wasm32-wasi: set ADAMIC_ORACLE_WASI=1"; len(recorder.logged) != 1 || recorder.logged[0] != want {
		t.Errorf("logged %q, want the one line %q", recorder.logged, want)
	}
	if recorder.skipped != "" || recorder.failed != "" {
		t.Errorf("skipped %q, failed %q; want the test to go on", recorder.skipped, recorder.failed)
	}
}

// An opted-in leg runs when the sysroot is set and fails its test when it isn't. Mutant: runsLeg running an opted-in
// leg without asking for the sysroot, which then fails later as a wasm32 build error instead of naming WASI_SYSROOT.
func TestLegOptedInRunsOrFails(t *testing.T) {
	t.Parallel()
	recorder := &recordingTB{}
	if !runsLeg(recorder, environment(map[string]string{"ADAMIC_ORACLE_WASI": "1", "WASI_SYSROOT": "/sdk/share/wasi-sysroot"}), "ADAMIC_ORACLE_WASI", "sweep.a on wasm32-wasi") {
		t.Error("dropped an opted-in leg with its sysroot set")
	}
	if recorder.skipped != "" || recorder.failed != "" || len(recorder.logged) != 0 {
		t.Errorf("with its sysroot: skipped %q, failed %q, logged %q; want nothing", recorder.skipped, recorder.failed, recorder.logged)
	}
	recorder = &recordingTB{}
	if runsLeg(recorder, environment(map[string]string{"ADAMIC_ORACLE_WASI": "1"}), "ADAMIC_ORACLE_WASI", "sweep.a on wasm32-wasi") {
		t.Error("ran an opted-in leg with WASI_SYSROOT unset")
	}
	if !strings.HasPrefix(recorder.failed, "ADAMIC_ORACLE_WASI=1 opts this test into wasm32-wasi, and WASI_SYSROOT is unset") || recorder.skipped != "" {
		t.Errorf("without its sysroot: failed %q, skipped %q; want a failure naming WASI_SYSROOT", recorder.failed, recorder.skipped)
	}
}

// A leg that needs only the sysroot runs whenever it's set, opted in or not; with it unset, an opted-in test fails and
// one not opted in logs the drop and goes on. Mutants: sysrootFor dropping the leg with nothing in the log, as
// TestCommandRuntimeArchivesHaveOneBuilder's continue did; sysrootFor logging, not failing, when opted in, as
// TestArrayUnionRead did.
func TestSysrootLegRunsFailsOrIsLogged(t *testing.T) {
	t.Parallel()
	for _, values := range []map[string]string{{"WASI_SYSROOT": "/sdk/share/wasi-sysroot"}, {"ADAMIC_ORACLE_WASI": "1", "WASI_SYSROOT": "/sdk/share/wasi-sysroot"}} {
		recorder := &recordingTB{}
		sysroot, runs := sysrootFor(recorder, environment(values), "ADAMIC_ORACLE_WASI", "the wasm options' product key")
		if sysroot != "/sdk/share/wasi-sysroot" || !runs || recorder.skipped != "" || recorder.failed != "" || len(recorder.logged) != 0 {
			t.Errorf("%v: sysroot %q runs %t, skipped %q, failed %q, logged %q; want the sysroot alone", values, sysroot, runs, recorder.skipped, recorder.failed, recorder.logged)
		}
	}
	recorder := &recordingTB{}
	if _, runs := sysrootFor(recorder, environment(map[string]string{"ADAMIC_ORACLE_WASI": "1"}), "ADAMIC_ORACLE_WASI", "the wasm options' product key"); runs {
		t.Error("ran an opted-in leg with WASI_SYSROOT unset")
	}
	if !strings.HasPrefix(recorder.failed, "ADAMIC_ORACLE_WASI=1 opts this test into wasm32-wasi, and WASI_SYSROOT is unset") || len(recorder.logged) != 0 {
		t.Errorf("opted in: failed %q, logged %q; want a failure, not a log line", recorder.failed, recorder.logged)
	}
	recorder = &recordingTB{}
	if _, runs := sysrootFor(recorder, environment(map[string]string{}), "ADAMIC_ORACLE_WASI", "the wasm options' product key"); runs {
		t.Error("ran a leg with WASI_SYSROOT unset")
	}
	if want := "wasi leg not run: the wasm options' product key: WASI_SYSROOT is unset"; len(recorder.logged) != 1 || recorder.logged[0] != want || recorder.failed != "" || recorder.skipped != "" {
		t.Errorf("not opted in: logged %q, failed %q, skipped %q; want the one line %q", recorder.logged, recorder.failed, recorder.skipped, want)
	}
}
