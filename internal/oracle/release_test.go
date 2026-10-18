package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/native"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestReleaseFMAAgreesWithNode(t *testing.T) {
	t.Parallel()
	releaseAgrees(t, "internal/oracle/testdata/release_fma.a", false)
}
func TestReleaseRecursionAgreesWithNode(t *testing.T) {
	t.Parallel()
	releaseAgrees(t, "internal/oracle/testdata/stack_forever.a", false)
}
func TestReleaseTailCallAgreesWithNode(t *testing.T) {
	t.Parallel()
	releaseAgrees(t, "internal/oracle/testdata/stack_tail_call.a", false)
}
func TestReleaseOracleCatchesOneByte(t *testing.T) {
	t.Parallel()
	releaseAgrees(t, "dedication/dedication.a", true)
}
func releaseAgrees(t *testing.T, fixture string, mutant bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if mutant {
		program.Strings[0] += "!"
	}
	binary := filepath.Join(t.TempDir(), "shipped")
	if err := native.Build(native.C(program), binary, native.Options{Release: true}); err != nil {
		t.Fatal(err)
	}
	difference := disagreement(want, execute(t, binary))
	if mutant {
		if difference != "stdout differs" {
			t.Fatalf("one-byte mutant escaped: %q", difference)
		}
		t.Log("one-byte mutant caught: stdout differs")
		return
	}
	if difference != "" {
		t.Fatal(difference)
	}
}

func TestReleaseRecursionGuardMutant(t *testing.T) {
	t.Parallel()
	releaseTailMutant(t, "stack_forever.a")
}
func TestReleaseTailCallGuardMutant(t *testing.T) {
	t.Parallel()
	releaseTailMutant(t, "stack_tail_call.a")
}
func releaseTailMutant(t *testing.T, fixture string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	options := native.Options{Release: true}
	flags := slices.DeleteFunc(native.LinkFlags(options), func(flag string) bool { return flag == "-fno-optimize-sibling-calls" })
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.BuildFlags(native.C(program), binary, options, flags); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	output, err := command.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Log("tail-call option omission caught: mutant loops beyond 2 s while Node panics")
		return
	}
	if command.ProcessState == nil {
		t.Fatalf("unexpected mutant result: %v %q", err, output)
	}
	if command.ProcessState.ExitCode() == want.exitCode {
		t.Fatalf("tail-call mutant survived: %q", output)
	}
	t.Logf("tail-call option omission caught: exit %d instead of %d", command.ProcessState.ExitCode(), want.exitCode)
}
