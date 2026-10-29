package native

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func requireDebugWitnessBytes(t *testing.T, directory, first, second string) {
	t.Helper()
	left, err := os.ReadFile(filepath.Join(directory, first))
	if err != nil {
		t.Fatal(err)
	}
	right, err := os.ReadFile(filepath.Join(directory, second))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("debug witnesses %s and %s differ between entry directories", first, second)
	}
	if !bytes.Contains(left, []byte("/adamic-debug-witness")) {
		t.Fatal("witness lost its canonical debug directory")
	}
	t.Logf("%s and %s: %d identical bytes, canonical witness directory", first, second, len(left))
}

// C compilation belongs to the products. Each consumer only compares the two
// executables built ahead from different disposable source directories.
func TestDebugLineWitnessReproducible(t *testing.T) {
	t.Parallel()
	directory := debugLineProduct(t, false)
	requireDebugWitnessBytes(t, directory, "program", "program-again")
	// A release build records no comp_dir, so it needs no map to match.
	release, err := os.ReadFile(filepath.Join(directory, "release"))
	if err != nil {
		t.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(directory, "release-again"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(release, again) {
		t.Fatal("release witnesses differ between entry directories")
	}
}

func TestDebugCommandWitnessReproducible(t *testing.T) {
	t.Parallel()
	requireDebugWitnessBytes(t, debugCommandProduct(t, false), "witness", "witness-again")
}

func TestDebugSharingWitnessReproducible(t *testing.T) {
	t.Parallel()
	directory := debugSharingProduct(t, false)
	requireDebugWitnessBytes(t, directory, "one/entry/program", "two/entry/program")
	contents, err := os.ReadFile(filepath.Join(directory, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 5 || fields[4] != "0.000" {
			t.Fatalf("sharing product publishes variable timing: %q", line)
		}
	}
}
