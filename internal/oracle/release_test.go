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
	path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	if mutant {
		program.Strings[0] += "!"
	}
	binary := filepath.Join(test.TempDir(), "shipped")
	if buildError := native.Build(native.C(program), binary, native.Options{Release: true}); buildError != nil {
		test.Fatal(buildError)
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
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	options := native.Options{Release: true}
	flags := slices.DeleteFunc(
		native.LinkFlags(options),
		func(flag string) bool { return flag == "-fno-optimize-sibling-calls" },
	)
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildFlagsError := native.BuildFlags(native.C(program), binary, options, flags); buildFlagsError != nil {
		test.Fatal(buildFlagsError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(timeoutContext, binary)
	output, combinedOutputError := command.CombinedOutput()
	if timeoutContext.Err() == context.DeadlineExceeded {
		test.Log("tail-call option omission caught: mutant loops beyond 2 s while Node panics")
		return
	}
	if command.ProcessState == nil {
		test.Fatalf("unexpected mutant result: %v %q", combinedOutputError, output)
	}
	if command.ProcessState.ExitCode() == want.exitCode {
		test.Fatalf("tail-call mutant survived: %q", output)
	}
	test.Logf("tail-call option omission caught: exit %d instead of %d", command.ProcessState.ExitCode(), want.exitCode)
}
