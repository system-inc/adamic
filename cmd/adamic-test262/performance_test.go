package main

import (
	"context"
	"os"

	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Opt in so ordinary correctness gates do not spend time measuring process startup.
func TestCompilerStartupMeasurement(t *testing.T) {
	if os.Getenv("ADAMIC_TEST262_MEASURE") != "1" {
		// census: measurement Opt-in timing or profile artifact comparison; does not replace correctness verification.
		t.Skip("measurement only")
	}
	directory := t.TempDir()
	compiler := filepath.Join(directory, "adamic")
	build, release := boundedrun.Command(boundedrun.Build, "go", "build", "-o", compiler, "./cmd/adamic")
	defer release()
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	source := filepath.Join(directory, "program.a")
	if err := os.WriteFile(source, []byte(program("assertSameValue(Math.abs(-4), 4);")), 0600); err != nil {
		t.Fatal(err)
	}
	var subprocess, process, startup time.Duration
	for range 30 {
		start := time.Now()
		actual := runCommandWithLimit(2*time.Minute, nil, 16<<20, compiler, "c", source)
		subprocess += time.Since(start)
		if actual.Exit != 0 {
			t.Fatal(actual.Stderr)
		}
		start = time.Now()
		loaded, err := load.Load([]string{source})
		if err != nil {
			t.Fatal(err)
		}
		lowered, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
		}
		code := native.C(lowered)
		process += time.Since(start)
		if code != actual.Stdout {
			t.Fatal("in process C differs")
		}
		start = time.Now()
		command, release := boundedrun.Command(boundedrun.Probe, compiler)
		_ = command.Run()
		release()
		startup += time.Since(start)
	}
	t.Logf("30 interleaved compiles: subprocess=%s in-process=%s startup-only=%s", subprocess, process, startup)
}
