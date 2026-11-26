package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A SpreadWith literal's source not stored as a plain object or a class instance stops in a keyless form of its own
// (Adamic on #k7g6f9h, Nov 26, under #7fg2fp3's form, lifted by #n0744ek), said whole, and it reads under #k7g6f9h
// alone. Mutant: its line dropped from runtimeNotYetKeyless, and the ruled message is refused.
func TestRuntimeNotYetSpreadWithSourceForm(t *testing.T) {
	t.Parallel()
	source, stop, list := runtimeNotYetFixture(t)
	path := filepath.Join(t.TempDir(), "stops.json")
	read := func(stderr, ruling string) error {
		keyless := stop
		keyless.Stderr, keyless.RulingTask, keyless.LiftingTask = stderr, ruling, "#n0744ek"
		writeRuntimeNotYetList(t, path, runtimeNotYetStopList{Comment: list.Comment, Stops: map[string]runtimeNotYetStop{contentHash(source): keyless}})
		_, err := readRuntimeNotYetStops(path)
		return err
	}
	ruled := "adamic: panic: " + ir.SpreadWithSourceStop + "\n"
	if err := read(ruled, "#k7g6f9h"); err != nil {
		t.Errorf("the ruled message refused: %v", err)
	}
	for name, stderr := range map[string]string{
		"the first words":     "adamic: panic: a spread of a value stored as an array or a host value into an object literal isn't built yet (#k7g6f9h)\n",
		"a key read after it": "adamic: panic: " + ir.SpreadWithSourceStop + " (reading '0')\n",
		"no newline":          "adamic: panic: " + ir.SpreadWithSourceStop,
		"no panic prefix":     ir.SpreadWithSourceStop + "\n",
	} {
		if err := read(stderr, "#k7g6f9h"); err == nil {
			t.Errorf("%s accepted: %q", name, stderr)
		}
	}
	if err := read(ruled, "#0sps03r"); err == nil {
		t.Error("the message under another ruling accepted")
	}
}

// The tuple pin, the run-time stop's one witness (a tuple held in a union with a plain object type), binds its
// exact bytes to the form, its ruling and its lifting task. A function stops at compile time (checkLibrarySpreads),
// so backstop_closure.ts holds no row, and this test says so.
func TestRuntimeNotYetSpreadWithSourceRows(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	list, err := readRuntimeNotYetStops(filepath.Join(root, "cloud/admission-corpus/runtime-notyet-stops.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"backstop_tuple"} {
		path := "internal/oracle/testdata/spread_with/" + name + ".ts"
		source, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		got, ok := list.Stops[contentHash(source)]
		want := runtimeNotYetStop{Path: path, RulingTask: "#k7g6f9h", LiftingTask: "#n0744ek", Stderr: "adamic: panic: " + ir.SpreadWithSourceStop + "\n"}
		if !ok || got != want {
			t.Errorf("%s: got %+v, listed %v; want %+v", name, got, ok, want)
		}
	}
	closure, err := os.ReadFile(filepath.Join(root, "internal/oracle/testdata/spread_with/backstop_closure.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := list.Stops[contentHash(closure)]; ok {
		t.Errorf("backstop_closure is listed as a run-time stop (%+v), but checkLibrarySpreads refuses it at compile time", got)
	}
}
