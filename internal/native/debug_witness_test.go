package native

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func requireDebugWitnessBytes(test *testing.T, directory, first, second string) {
	test.Helper()
	left, err := os.ReadFile(filepath.Join(directory, first))
	if err != nil {
		test.Fatal(err)
	}
	right, err := os.ReadFile(filepath.Join(directory, second))
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		test.Fatalf("debug witnesses %s and %s differ between entry directories", first, second)
	}
	if !bytes.Contains(left, []byte("/adamic-debug-witness")) {
		test.Fatal("witness lost its canonical debug directory")
	}
	test.Logf("%s and %s: %d identical bytes, canonical witness directory", first, second, len(left))
}

// C compilation belongs to the products. Each consumer only compares the two
// executables built ahead from different disposable source directories.
func TestDebugLineWitnessReproducible(test *testing.T) {
	test.Parallel()
	directory := debugLineProduct(test, false)
	requireDebugWitnessBytes(test, directory, "program", "program-again")
	// A release build records no comp_dir, so it needs no map to match.
	release, err := os.ReadFile(filepath.Join(directory, "release"))
	if err != nil {
		test.Fatal(err)
	}
	again, err := os.ReadFile(filepath.Join(directory, "release-again"))
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(release, again) {
		test.Fatal("release witnesses differ between entry directories")
	}
}

func TestDebugCommandWitnessReproducible(test *testing.T) {
	test.Parallel()
	requireDebugWitnessBytes(test, debugCommandProduct(test, false), "witness", "witness-again")
}

func TestDebugSharingWitnessReproducible(test *testing.T) {
	test.Parallel()
	directory := debugSharingProduct(test, false)
	requireDebugWitnessBytes(test, directory, "one/entry/program", "two/entry/program")
	contents, err := os.ReadFile(filepath.Join(directory, "build.log"))
	if err != nil {
		test.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 5 || fields[4] != "0.000" {
			test.Fatalf("sharing product publishes variable timing: %q", line)
		}
	}
}
