package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// smallMacBinary builds a tiny Go program for macOS on goArchitecture and returns its path. The guards
// below read real Mach-O load commands, so they are tested against a real Mach-O rather than a byte
// pattern written to look like one.
func smallMacBinary(t *testing.T, goArchitecture string) string {
	t.Helper()

	if runtime.GOOS != "darwin" {
		t.Skip("NOT MEASURED: lipo and vtool are macOS tools, and the Swift engine is only packaged on macOS")
	}

	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "go.mod"), "module small\n\ngo 1.21\n")
	writeFile(t, filepath.Join(directory, "main.go"), "package main\n\nfunc main() {}\n")

	binary := filepath.Join(directory, "small")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = directory
	command.Env = append(command.Environ(), "GOOS=darwin", "GOARCH="+goArchitecture, "CGO_ENABLED=0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("building the sample binary: %v\n%s", err, output)
	}
	return binary
}

// TestArchitectureIsReadFromTheBinary holds both directions: a binary for the package's architecture
// passes, and one for the other Mac is refused, naming what it was built for.
func TestArchitectureIsReadFromTheBinary(t *testing.T) {
	t.Parallel()

	arm := smallMacBinary(t, "arm64")
	if err := requireArchitecture(arm, "arm64"); err != nil {
		t.Fatalf("an arm64 binary was refused as arm64: %v", err)
	}

	err := requireArchitecture(arm, "x86_64")
	if err == nil {
		t.Fatal("an arm64 binary was accepted for an x86_64 package")
	}
	if !strings.Contains(err.Error(), "arm64") {
		t.Fatalf("the refusal does not say what the binary was built for: %v", err)
	}
}

// TestMinimumMacOSIsReadFromTheBinary uses one binary twice: as Go built it, which passes, and with its
// recorded minimum rewritten to macOS 27 by vtool, which is refused. The rewritten copy is the case the
// guard exists for, an engine built against a deployment target newer than the release promises.
func TestMinimumMacOSIsReadFromTheBinary(t *testing.T) {
	t.Parallel()

	binary := smallMacBinary(t, "arm64")
	if err := requireMinimumMacOS(binary, MinimumSwiftEngineMacOS); err != nil {
		t.Fatalf("a binary Go built for an old macOS was refused: %v", err)
	}

	tooNew := filepath.Join(t.TempDir(), "too-new")
	rewrite := exec.Command("vtool", "-set-build-version", "macos", "27.0", "27.0", "-replace", "-output", tooNew, binary)
	if output, err := rewrite.CombinedOutput(); err != nil {
		t.Fatalf("rewriting the minimum macOS: %v\n%s", err, output)
	}

	err := requireMinimumMacOS(tooNew, MinimumSwiftEngineMacOS)
	if err == nil {
		t.Fatal("a binary requiring macOS 27 was accepted for a release promising macOS 15")
	}
	if !strings.Contains(err.Error(), "requires macOS 27.0") {
		t.Fatalf("the refusal does not name the minimum it found: %v", err)
	}
}

// TestMacOSVersionsCompareByNumber holds the comparison the guard rests on, including the case a string
// comparison gets wrong: 9 sorts after 15 as text and before it as a version.
func TestMacOSVersionsCompareByNumber(t *testing.T) {
	t.Parallel()

	cases := []struct {
		version string
		than    string
		newer   bool
	}{
		{"27.0", "15.0", true},
		{"15.0", "15.0", false},
		{"15", "15.0", false},
		{"15.1", "15.0", true},
		{"14.9", "15.0", false},
		{"9.0", "15.0", false},
		{"15.0.1", "15.0", false},
	}
	for _, testCase := range cases {
		newer, err := macOSVersionIsNewer(testCase.version, testCase.than)
		if err != nil {
			t.Fatalf("comparing %s with %s: %v", testCase.version, testCase.than, err)
		}
		if newer != testCase.newer {
			t.Errorf("%s newer than %s: got %v, want %v", testCase.version, testCase.than, newer, testCase.newer)
		}
	}

	if _, err := macOSVersionIsNewer("fifteen", "15.0"); err == nil {
		t.Error("a version that is not a number was compared")
	}
}

// TestSwiftArchitectureNamesEveryMacTarget ties the Go-to-Swift architecture names to Targets, so a Mac
// target added later without a Swift name fails here rather than in a release.
func TestSwiftArchitectureNamesEveryMacTarget(t *testing.T) {
	t.Parallel()

	for _, target := range Targets {
		if !ShipsSwiftEngine(target.GoOperatingSystem) {
			continue
		}
		if _, err := swiftArchitecture(target.GoArchitecture); err != nil {
			t.Errorf("%s ships the Swift engine and has no Swift architecture: %v", target, err)
		}
	}
}

// contractPair writes a stand-in cohere whose `--version` prints cohereVersion, and a stand-in engine that
// runs engineScript with its arguments in "$@". Shell scripts rather than real builds, because what is
// under test is how the release reads the two answers, and a real engine takes fifteen minutes to build.
func contractPair(t *testing.T, cohereVersion string, engineScript string) (coherePath string, enginePath string) {
	t.Helper()

	directory := t.TempDir()
	coherePath = filepath.Join(directory, "cohere")
	enginePath = filepath.Join(directory, SwiftEngineFileName)
	writeExecutable(t, coherePath, 0o755)
	writeFile(t, coherePath, "#!/bin/sh\n[ \"$1\" = --version ] || exit 9\nprintf '%s\\n' '"+strings.ReplaceAll(cohereVersion, "\n", "' '")+"'\n")
	writeExecutable(t, enginePath, 0o755)
	writeFile(t, enginePath, "#!/bin/sh\n"+engineScript+"\n")
	return coherePath, enginePath
}

// stampedCommit is the commit the stand-in engines below say they were built from, and the one the
// release is told it stamped.
const stampedCommit = "0123456789abcdef0123456789abcdef01234567"

// speaksContract is an engine stamped with stampedCommit that answers `--contract <its> --version` with its
// provenance record and refuses any other contract the way the real one does: a sentence on stderr and
// exit 2.
func speaksContract(contract string) string {
	return `if [ "$1" != --contract ] || [ "$3" != --version ]; then exit 9; fi
if [ "$2" != ` + contract + ` ]; then echo "cohere-swift speaks contract ` + contract + `, and the front door asked for $2" >&2; exit 2; fi
echo '{"kind":"provenance","contract":` + contract + `,"engine":"cohere-swift","version":"0.1.0","commit":"` + stampedCommit + `"}'`
}

// TestSwiftContractAcceptsAMatchingPair is the positive half. Without it, a check that refused every pair
// would pass the controls below.
func TestSwiftContractAcceptsAMatchingPair(t *testing.T) {
	t.Parallel()

	coherePath, enginePath := contractPair(t, "cohere 1.0.0\n  swift contract: 3", speaksContract("3"))
	if err := requireSwiftContract(coherePath, enginePath, stampedCommit); err != nil {
		t.Fatalf("a pair speaking the same contract was refused: %v", err)
	}
}

// TestSwiftContractRefusesAMismatchedPair holds each way a pair can disagree, and that each refusal says
// which.
func TestSwiftContractRefusesAMismatchedPair(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name          string
		cohereVersion string
		engineScript  string
		says          string
	}{
		{"the engine refuses the flag", "cohere 1.0.0\n  swift contract: 4", speaksContract("3"), "refused contract 4"},
		{"the engine accepts the flag and its record names another contract", "cohere 1.0.0\n  swift contract: 3",
			`echo '{"kind":"provenance","contract":2}'`, "speaks contract 2"},
		{"the engine answers with something that is not provenance", "cohere 1.0.0\n  swift contract: 3",
			`echo '{"kind":"summary","contract":3}'`, "not a provenance record"},
		{"the engine's record names no contract", "cohere 1.0.0\n  swift contract: 3",
			`echo '{"kind":"provenance"}'`, "not a provenance record"},
		{"cohere names no contract", "cohere 1.0.0", speaksContract("3"), "names no Swift contract"},
		{"the engine was not stamped", "cohere 1.0.0\n  swift contract: 3",
			`echo '{"kind":"provenance","contract":3,"commit":"dev"}'`, `built from "dev"`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			coherePath, enginePath := contractPair(t, testCase.cohereVersion, testCase.engineScript)
			err := requireSwiftContract(coherePath, enginePath, stampedCommit)
			if err == nil {
				t.Fatal("a mismatched pair was staged")
			}
			if !strings.Contains(err.Error(), testCase.says) {
				t.Fatalf("refused, but without saying %q: %v", testCase.says, err)
			}
		})
	}
}

// TestSwiftEngineBuildReadsTheTreeAsItIsNow builds a stand-in engine through buildSwiftEngine twice in one
// repository: clean, then with a tracked file edited. It is laid out as cohere-swift is, a CohereSwift
// target with the stamp gitignored under Command/, its manifest defines a flag from
// `Context.gitInformation.hasUncommittedChanges` as cohere-swift's does, and the product prints that flag
// and the commit it was stamped with.
//
// SwiftPM's shared manifest cache would answer the second build from the first evaluation and stamp a
// modified tree as clean; the release turns that cache off. The first build also proves the stamp: it
// reaches the binary, it does not count as a modification, and it is gone afterwards.
func TestSwiftEngineBuildReadsTheTreeAsItIsNow(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "darwin" {
		t.Skip("NOT MEASURED: the Swift engine is only built on macOS")
	}
	if _, err := exec.LookPath("swift"); err != nil {
		t.Skip("NOT MEASURED: no swift toolchain on PATH")
	}

	module := t.TempDir()
	writeFile(t, filepath.Join(module, "swift", "Package.swift"), `// swift-tools-version:6.2
import PackageDescription
let modified = Context.gitInformation?.hasUncommittedChanges ?? true
let package = Package(name: "Stand", products: [.executable(name: "cohere-swift", targets: ["CohereSwift"])],
    targets: [.executableTarget(name: "CohereSwift", swiftSettings: modified ? [.define("TREE_MODIFIED")] : [])])
`)
	source := filepath.Join(module, "swift", "Sources", "CohereSwift", "main.swift")
	program := "#if TREE_MODIFIED\nlet tree = \"modified\"\n#else\nlet tree = \"clean\"\n#endif\n" +
		"#if COHERE_RELEASE_STAMP\nprint(tree, EngineReleaseStamp.commit)\n#else\nprint(tree, \"dev\")\n#endif\n"
	writeFile(t, source, program)
	writeFile(t, filepath.Join(module, "swift", "Sources", "CohereSwift", "Command", "Placeholder.swift"), "enum Placeholder {}\n")
	writeFile(t, filepath.Join(module, "swift", ".gitignore"), "/Sources/CohereSwift/Command/"+SwiftEngineStampFileName+"\n")
	writeFile(t, filepath.Join(module, ".gitignore"), ".scratch/\n")
	gitIn(t, module, "init", "--quiet")
	gitIn(t, module, "add", ".")
	gitIn(t, module, "commit", "--quiet", "-m", "stand-in engine")
	commit, err := readReleaseCommit(module)
	if err != nil {
		t.Fatal(err)
	}

	target := Target{GoOperatingSystem: "darwin", GoArchitecture: runtime.GOARCH}
	scratch := filepath.Join(module, ".scratch")
	says := func() string {
		t.Helper()
		destination := filepath.Join(t.TempDir(), SwiftEngineFileName)
		if err := buildSwiftEngine(module, scratch, target, destination, commit); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(destination).Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(output))
	}

	if got := says(); got != "clean "+commit {
		t.Fatalf("an engine built from a clean tree says %q, and should say clean and its commit %s", got, commit)
	}
	if _, err := os.Stat(filepath.Join(module, "swift", "Sources", "CohereSwift", "Command", SwiftEngineStampFileName)); err == nil {
		t.Fatal("the stamp was left in the checkout after the build")
	}
	writeFile(t, source, program+"// edited\n")
	if got := says(); got != "modified "+commit {
		t.Fatalf("an engine built after a tracked file was edited says %q, so its provenance would call the tree clean", got)
	}
}
