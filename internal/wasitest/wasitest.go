// Package wasitest decides whether a test's wasm32-wasi leg runs. Only _test.go files import it, so no binary carries
// it. The rule (#p3rd3vn): a test that isn't opted in skips its wasi leg, as it always has; a test opted in with anything
// the leg needs missing fails, and never skips, so a run that asked for wasi and got none can't pass; and a leg dropped
// inside a test whose other legs still run says so in the test's log, so a drop is at least an event. A proof once ran
// the wasi product tests without the opt-ins: six of seven skipped, the run exited 0, and a wasm32 error reached the
// line.
//
// Each function returns after Skipf and Fatalf, which stop a *testing.T but needn't stop every testing.TB.
package wasitest

import (
	"os"
	"testing"
)

// sysrootVariable names the WASI SDK's sysroot, which every wasm32-wasi build reads (native's ValidateOptions refuses
// the target without it) and whose SDK's clang sits beside it.
const sysrootVariable = "WASI_SYSROOT"

// Require is the sysroot for a test that is all wasi leg. It skips the test when optIn isn't 1, and fails it, naming
// what's missing, when optIn is 1 and WASI_SYSROOT is unset.
func Require(tb testing.TB, optIn string) string {
	tb.Helper()
	return require(tb, os.Getenv, optIn)
}

// Leg is whether a wasi leg runs in a test whose other legs run either way: true when optIn is 1 and WASI_SYSROOT is
// set, a failure when optIn is 1 and it isn't, and false when optIn isn't 1, once the test's log says leg wasn't run and
// why.
func Leg(tb testing.TB, optIn, leg string) bool {
	tb.Helper()
	return runsLeg(tb, os.Getenv, optIn, leg)
}

// Sysroot is the sysroot for a wasi leg that needs only it, as a runtime archive's key does, in a test whose other legs
// run either way: the sysroot and true when WASI_SYSROOT is set, opted in or not; a failure when it's unset and optIn
// is 1; and false when it's unset and optIn isn't 1, once the test's log says leg wasn't run and why.
func Sysroot(tb testing.TB, optIn, leg string) (string, bool) {
	tb.Helper()
	return sysrootFor(tb, os.Getenv, optIn, leg)
}

// require is Require reading the environment through lookup, so a test can hand it one without setting the process's.
func require(tb testing.TB, lookup func(string) string, optIn string) string {
	tb.Helper()
	if lookup(optIn) != "1" {
		tb.Skipf("set %s=1 and %s to run this wasm32-wasi test", optIn, sysrootVariable)
		return ""
	}
	sysroot, _ := optedIn(tb, lookup, optIn)
	return sysroot
}

// runsLeg is Leg reading the environment through lookup.
func runsLeg(tb testing.TB, lookup func(string) string, optIn, leg string) bool {
	tb.Helper()
	if lookup(optIn) != "1" {
		notRun(tb, leg, "set "+optIn+"=1")
		return false
	}
	_, runs := optedIn(tb, lookup, optIn)
	return runs
}

// sysrootFor is Sysroot reading the environment through lookup.
func sysrootFor(tb testing.TB, lookup func(string) string, optIn, leg string) (string, bool) {
	tb.Helper()
	if sysroot := lookup(sysrootVariable); sysroot != "" {
		return sysroot, true
	}
	if lookup(optIn) == "1" {
		return optedIn(tb, lookup, optIn)
	}
	notRun(tb, leg, sysrootVariable+" is unset")
	return "", false
}

// optedIn is the sysroot of a test optIn opted in, and fails the test, naming what's missing, when WASI_SYSROOT is
// unset: a run that asked for wasi and can't build it is a failure, never a skip.
func optedIn(tb testing.TB, lookup func(string) string, optIn string) (string, bool) {
	tb.Helper()
	sysroot := lookup(sysrootVariable)
	if sysroot == "" {
		tb.Fatalf("%s=1 opts this test into wasm32-wasi, and %s is unset: set it to the WASI SDK's sysroot (bash cloud/setup.sh --wasi-sdk installs one), or unset %s to skip the wasi legs", optIn, sysrootVariable, optIn)
		return "", false
	}
	return sysroot, true
}

// notRun says in the test's log that leg wasn't run, and why, so a wasi leg a test drops while its others run is an
// event rather than nothing.
func notRun(tb testing.TB, leg, why string) {
	tb.Helper()
	tb.Logf("wasi leg not run: %s: %s", leg, why)
}
