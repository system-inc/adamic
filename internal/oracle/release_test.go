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

func TestReleaseFMAAgreesWithNode(test *testing.T) {
	test.Parallel()
	releaseAgrees(test, "internal/oracle/testdata/release_fma.a", false)
}
func TestReleaseRecursionAgreesWithNode(test *testing.T) {
	test.Parallel()
	releaseAgrees(test, "internal/oracle/testdata/stack_forever.a", false)
}
func TestReleaseTailCallAgreesWithNode(test *testing.T) {
	test.Parallel()
	releaseAgrees(test, "internal/oracle/testdata/stack_tail_call.a", false)
}
func TestReleaseOracleCatchesOneByte(test *testing.T) {
	test.Parallel()
	releaseAgrees(test, "dedication/dedication.a", true)
}
func releaseAgrees(test *testing.T, fixture string, mutant bool) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	if mutant {
		program.Strings[0] += "!"
	}
	binary := filepath.Join(test.TempDir(), "shipped")
	if err := native.Build(native.C(program), binary, native.Options{Release: true}); err != nil {
		test.Fatal(err)
	}
	difference := disagreement(want, execute(test, binary))
	if mutant {
		if difference != "stdout differs" {
			test.Fatalf("one-byte mutant escaped: %q", difference)
		}
		test.Log("one-byte mutant caught: stdout differs")
		return
	}
	if difference != "" {
		test.Fatal(difference)
	}
}

func TestReleaseRecursionGuardMutant(test *testing.T) {
	test.Parallel()
	releaseTailMutant(test, "stack_forever.a")
}
func TestReleaseTailCallGuardMutant(test *testing.T) {
	test.Parallel()
	releaseTailMutant(test, "stack_tail_call.a")
}
func releaseTailMutant(test *testing.T, fixture string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	options := native.Options{Release: true}
	flags := slices.DeleteFunc(
		native.LinkFlags(options),
		func(flag string) bool { return flag == "-fno-optimize-sibling-calls" },
	)
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.BuildFlags(native.C(program), binary, options, flags); err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	output, err := command.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		test.Log("tail-call option omission caught: mutant loops beyond 2 s while Node panics")
		return
	}
	if command.ProcessState == nil {
		test.Fatalf("unexpected mutant result: %v %q", err, output)
	}
	if command.ProcessState.ExitCode() == want.exitCode {
		test.Fatalf("tail-call mutant survived: %q", output)
	}
	test.Logf("tail-call option omission caught: exit %d instead of %d", command.ProcessState.ExitCode(), want.exitCode)
}
