package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A released binary that quietly stops being stripped grows about 18 MB, measured at 43.1 against
// 61.9 for the same commit. Nothing fails when that happens: the build succeeds, the packages stage,
// the binaries run. The only signal is a size nobody was watching, which is why it is asserted here
// rather than trusted to review.

func TestReleaseBuildStripsSymbols(t *testing.T) {
	t.Parallel()

	if len(StripFlags) == 0 {
		t.Fatalf("no strip flags, so released binaries would ship with symbols and DWARF")
	}

	stripped := strings.Join(StripFlags, " ")
	for _, required := range []string{"-s", "-w"} {
		if !strings.Contains(stripped, required) {
			t.Errorf("strip flags %q are missing %s", stripped, required)
		}
	}
}

func TestDescribeBuildNamesTheFlagsActuallyUsed(t *testing.T) {
	t.Parallel()

	// A size label that names the wrong build is worse than an unlabeled size: it is confidently
	// wrong rather than ambiguous. So the description is built from the same values the compiler
	// receives, and this asserts it cannot drift into a hand-written string.
	description := DescribeBuild()

	for _, flag := range BuildFlags {
		if !strings.Contains(description, flag) {
			t.Errorf("the build description %q omits %s, which the build actually passes", description, flag)
		}
	}
	for _, flag := range StripFlags {
		if !strings.Contains(description, flag) {
			t.Errorf("the build description %q omits %s, which the build actually passes", description, flag)
		}
	}
}

// A binary that exists and is the right size still says nothing about which machine it runs on.
// That distinction is the whole point of these: "verified to exist" answers a narrower question
// than its name implies, and a mis-staged binary passes every check that stops at existence.

func TestVerifyBinaryRefusesTheWrongPlatformsExecutable(t *testing.T) {
	t.Parallel()

	// The reachable failure: a staging bug writes one target's binary into another's package. It
	// publishes cleanly, installs cleanly, and fails at exec on a user's machine with a format
	// error that reads as a broken install rather than as our mistake.
	darwin := Target{GoOperatingSystem: "darwin", GoArchitecture: "arm64"}

	elf := writeFakeExecutable(t, []byte{0x7f, 'E', 'L', 'F'})
	if _, err := cohereBinary(elf, darwin); err == nil {
		t.Fatalf("a Linux binary passed verification as a darwin package's contents")
	}

	windows := writeFakeExecutable(t, []byte{'M', 'Z'})
	if _, err := cohereBinary(windows, darwin); err == nil {
		t.Fatalf("a Windows binary passed verification as a darwin package's contents")
	}
}

func TestVerifyBinaryRefusesSomethingThatIsNotAnExecutable(t *testing.T) {
	t.Parallel()

	// Large enough to clear the size floor and still not a program.
	notABinary := writeFakeExecutable(t, []byte("#!/bin/sh\necho nope\n"))
	if _, err := cohereBinary(notABinary, Target{GoOperatingSystem: "linux", GoArchitecture: "amd64"}); err == nil {
		t.Fatalf("plain text passed verification as a Linux binary")
	}
}

func TestVerifyBinaryAcceptsTheRightFormat(t *testing.T) {
	t.Parallel()

	// A guard that refuses everything is as useless as one that refuses nothing, and it is the
	// version that gets deleted rather than fixed.
	cases := map[string][]byte{
		"darwin":  {0xcf, 0xfa, 0xed, 0xfe},
		"linux":   {0x7f, 'E', 'L', 'F'},
		"windows": {'M', 'Z'},
	}
	for operatingSystem, magic := range cases {
		path := writeFakeExecutable(t, magic)
		target := Target{GoOperatingSystem: operatingSystem, GoArchitecture: "amd64"}
		if _, err := cohereBinary(path, target); err != nil {
			t.Errorf("a correct %s binary was refused: %v", operatingSystem, err)
		}
	}
}

func TestEveryTargetHasAKnownExecutableFormat(t *testing.T) {
	t.Parallel()

	// The guard skips an operating system it has no magic for, so that adding a target never blocks
	// a release on a missing table entry. That is the right runtime behavior and the wrong thing to
	// discover silently, so the omission fails here instead.
	for _, target := range Targets {
		if _, known := executableMagic[target.GoOperatingSystem]; !known {
			t.Errorf("%s ships with no executable-format check, so a mis-staged binary would pass", target)
		}
	}
}

// writeFakeExecutable writes a file that clears the size floor and begins with the given bytes.
func writeFakeExecutable(t *testing.T, magic []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cohere")

	const clearsTheSizeFloor = 2 << 20
	contents := make([]byte, clearsTheSizeFloor)
	copy(contents, magic)

	if err := os.WriteFile(path, contents, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExecutableMagicMatchesWhatTheCompilerActuallyEmits(t *testing.T) {
	t.Parallel()

	// Every other test of this table builds its fixtures out of the table, so they pass for any
	// value in it, including a wrong one. That is the shape where a corpus encodes a misreading:
	// the suite agrees with the belief that produced it and never touches the thing the belief is
	// about.
	//
	// So this compiles a real program for each operating system we ship and reads the bytes Go
	// actually wrote. About two seconds for three platforms, which is the entire cost of anchoring
	// the last constant in this package to something outside it.
	if testing.Short() {
		t.Skip("cross-compiles three binaries")
	}

	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "go.mod"), "module magiccheck\n\ngo 1.27\n")
	writeFile(t, filepath.Join(directory, "main.go"), "package main\n\nfunc main() {}\n")

	// One architecture per operating system: the magic is keyed by GOOS, and this asserts exactly
	// that scope rather than implying it covers architecture too.
	for _, target := range []Target{
		{GoOperatingSystem: "darwin", GoArchitecture: "arm64"},
		{GoOperatingSystem: "linux", GoArchitecture: "amd64"},
		{GoOperatingSystem: "windows", GoArchitecture: "amd64"},
	} {
		path := filepath.Join(directory, "out-"+target.GoOperatingSystem)

		command := exec.Command("go", "build", "-o", path, ".")
		command.Dir = directory
		command.Env = append(os.Environ(),
			"GOOS="+target.GoOperatingSystem,
			"GOARCH="+target.GoArchitecture,
			"CGO_ENABLED=0",
		)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("cross-compiling for %s: %v\n%s", target, err, output)
		}

		if _, err := cohereBinaryFormatOnly(path, target); err != nil {
			t.Errorf("the magic table rejects a binary Go actually produced for %s: %v", target, err)
		}
	}
}

// cohereBinaryFormatOnly runs the format check without the size floor.
//
// A hello-world binary is well under a megabyte, so `cohereBinary` would refuse it for a reason
// that has nothing to do with what this test is asking.
func cohereBinaryFormatOnly(path string, target Target) (int64, error) {
	return 0, requireExecutableFormat(path, target)
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
