package buildcache

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestRequireArtifacts(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	for _, name := range []string{"missing", "empty", "directory"} {
		if name == "empty" {
			write(test, directory, name, "")
		}
		if name == "directory" {
			if err := os.Mkdir(filepath.Join(directory, name), 0755); err != nil {
				test.Fatal(err)
			}
		}
		if err := requireArtifact(directory, Artifact{Name: name}); err == nil {
			test.Fatalf("accepted %s artifact", name)
		}
	}
	write(test, directory, "data", "built")
	RequireArtifacts(test, directory, Artifact{Name: "data"})
	for _, name := range []string{"", "../data", "/data"} {
		if err := requireArtifact(directory, Artifact{Name: name}); err == nil {
			test.Fatalf("accepted name %q", name)
		}
	}
	if err := requireArtifact("", Artifact{Name: "data"}); err == nil {
		test.Fatal("accepted empty directory")
	}
	if err := requireArtifact(directory, Artifact{Name: "data", Executable: true}); err == nil {
		test.Fatal("accepted unusable executable")
	}
	writeExecutable(
		test,
		filepath.Join(directory, "probe"),
		"#!/bin/sh\nread value\n[ \"$value\" = built ] || exit 3\nexit 7\n",
	)
	RequireArtifacts(test, directory, Artifact{Name: "probe", Executable: true, Stdin: "built\n", ExitCode: 7})
	if err := requireArtifact(directory, Artifact{Name: "probe", Executable: true, Stdin: "built\n"}); err == nil {
		test.Fatal("accepted wrong exit")
	}
}

// A probe that exits well while a child it started still holds the output pipe passes, and soon: its process group goes
// once it exits (childguard.Capture), so the child can't hold the read, and no WaitDelay after the exit makes the drain
// red by chance (#n3gk22h, adamiclint's waitdelay on artifacts.go). Mutant: CombinedOutput with the 5 s WaitDelay back,
// red (exec's WaitDelay error, the probe failed).
func TestAProbeWhoseChildHoldsItsOutputPasses(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	writeExecutable(test, filepath.Join(directory, "probe"), "#!/bin/sh\nsleep 30 &\necho ready\nexit 0\n")
	started := time.Now()
	if err := requireArtifact(directory, Artifact{Name: "probe", Executable: true}); err != nil {
		test.Fatalf("a probe whose child holds its output: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 20*time.Second {
		test.Fatalf("the probe took %v, held by its child", elapsed)
	}
}

func TestWithoutLeakDetection(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct{ environment, want []string }{
		{[]string{"HOME=/h"}, []string{"HOME=/h", "ASAN_OPTIONS=detect_leaks=0"}},
		{
			[]string{"ASAN_OPTIONS=halt_on_error=1", "HOME=/h"},
			[]string{"ASAN_OPTIONS=halt_on_error=1:detect_leaks=0", "HOME=/h"},
		},
		{[]string{"ASAN_OPTIONS=detect_leaks=1"}, []string{"ASAN_OPTIONS=detect_leaks=1:detect_leaks=0"}},
		{[]string{"ASAN_OPTIONS="}, []string{"ASAN_OPTIONS=detect_leaks=0"}},
	} {
		if got := withoutLeakDetection(testCase.environment); !slices.Equal(got, testCase.want) {
			test.Errorf("withoutLeakDetection(%q) = %q, want %q", testCase.environment, got, testCase.want)
		}
	}
}
