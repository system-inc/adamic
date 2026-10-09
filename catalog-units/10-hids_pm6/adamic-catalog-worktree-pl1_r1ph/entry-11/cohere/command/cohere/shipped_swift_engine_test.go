//go:build unix

// The stand-in engines here are /bin/sh scripts, so the file builds on Unix, where the engine ships.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/release/dispatch"
	"github.com/system-inc/cohere/internal/release/packaging"
)

// installCohere puts cohere at `<install>/bin/cohere`, the layout of a platform package, stamped with
// version when it is not empty, and returns the binary. The install directory has no cohere module above
// it, so the only engine a run can find is one placed beside the binary.
//
// Each install is a directory of its own, holding a hard link to a build shared by every install of that
// stamp: the package's binary for no stamp, and stampedCohere's for a version. Linking cohere once per
// install was three links of about 50 MB in each package run, for two distinct binaries (#nxgt2ca). Nothing
// writes to an installed binary; a test places files beside it.
func installCohere(t *testing.T, version string) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "bin", "cohere")
	if err := os.MkdirAll(filepath.Dir(binary), 0o755); err != nil {
		t.Fatal(err)
	}
	source := buildCohere(t)
	if version != "" {
		source = stampedCohere(t, version)
	}
	if err := os.Link(source, binary); err != nil {
		// Across filesystems a link fails, and a copy is the same install.
		contents, readError := os.ReadFile(source)
		if readError != nil {
			t.Fatal(readError)
		}
		if writeError := os.WriteFile(binary, contents, 0o755); writeError != nil {
			t.Fatal(writeError)
		}
	}
	return binary
}

// stampedCohere returns this command's binary stamped with version, built on the first call for it.
func stampedCohere(t *testing.T, version string) string {
	t.Helper()
	stampedBinaries.Lock()
	stamped := stampedBinaries.byVersion[version]
	if stamped == nil {
		stamped = &stampedBinary{}
		stampedBinaries.byVersion[version] = stamped
	}
	stampedBinaries.Unlock()
	stamped.once.Do(func() {
		stamped.directory, stamped.err = os.MkdirTemp("", "cohere-test-stamped-binary-")
		if stamped.err != nil {
			return
		}
		stamped.path = filepath.Join(stamped.directory, "cohere")
		build := exec.Command("go", "build", "-o", stamped.path,
			"-ldflags=-X github.com/system-inc/cohere/internal/release/packaging.version="+version, ".")
		stamped.output, stamped.err = build.CombinedOutput()
	})
	if stamped.err != nil {
		t.Fatalf("cannot build cohere stamped %s: %v\n%s", version, stamped.err, stamped.output)
	}
	return stamped.path
}

// placeSiblingEngine puts a stand-in cohere-swift beside binary that records its arguments in a file
// and streams a clean run, and returns the file it records into.
func placeSiblingEngine(t *testing.T, binary string) string {
	t.Helper()
	clean, err := filepath.Abs(filepath.Join("..", "..", "swift", "Contract", "Clean.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	invoked := filepath.Join(t.TempDir(), "invoked")
	script := "#!/bin/sh\necho \"$@\" > '" + invoked + "'\ncat '" + clean + "'\n"
	if err := os.WriteFile(filepath.Join(filepath.Dir(binary), release.SwiftEngineFileName), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return invoked
}

// swiftPackage writes a Swift package, a Package.swift and one source, under parent, or under a new
// temporary directory when parent is empty.
func swiftPackage(t *testing.T, parent string) string {
	t.Helper()
	if parent == "" {
		parent = t.TempDir()
	}
	root := filepath.Join(parent, "Toy")
	writeTree(t, root, map[string]string{
		"Package.swift":         "// swift-tools-version:6.0\nimport PackageDescription\nlet package = Package(name: \"Toy\")\n",
		"Sources/Toy/Toy.swift": "let toy = 1\n",
	})
	return root
}

// runInstalled runs an installed cohere from directory with no engine named by the override, so the
// run resolves its engine the way an installed one does. The override is cleared in the child's
// environment alone, not this process's, so the tests that call it run in parallel (#nxgt2ca).
func runInstalled(t *testing.T, binary string, directory string, arguments ...string) (string, int) {
	t.Helper()
	return runCohereWithEnvironment(t, binary, directory, []string{dispatch.SwiftEngineOverrideVariable + "="}, arguments...)
}

// A released cohere runs the engine its package ships beside it, with no checkout anywhere, reached
// through a symlink the way a package manager's `bin` link reaches it.
func TestAReleasedCohereRunsTheEngineShippedBesideIt(t *testing.T) {
	t.Parallel()

	binary := installCohere(t, "1.0.0")
	invoked := placeSiblingEngine(t, binary)
	link := filepath.Join(t.TempDir(), "cohere")
	if err := os.Symlink(binary, link); err != nil {
		t.Fatal(err)
	}
	root := swiftPackage(t, "")

	// --verbose, so an engine build would say so: by default a build is silent off a terminal (#ytqqv8v),
	// and the check below would pass over one.
	output, exitCode := runInstalled(t, link, root, "--no-fix", "--verbose")
	if exitCode != 0 {
		t.Fatalf("a released cohere with its engine beside it exited %d:\n%s", exitCode, output)
	}
	recorded, err := os.ReadFile(invoked)
	if err != nil {
		t.Fatalf("the engine beside cohere was never run:\n%s", output)
	}
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	if want := "--contract 3 --root "; !strings.HasPrefix(string(recorded), want) ||
		!(strings.Contains(string(recorded), root) || strings.Contains(string(recorded), resolvedRoot)) {
		t.Fatalf("the shipped engine was run with %q, expected the contract and the package root %s", recorded, root)
	}
	if strings.Contains(output, "building the Swift engine") {
		t.Fatalf("a released cohere built an engine:\n%s", output)
	}

	// The release reads the contract from the cohere it stages, outside any project, and refuses an
	// engine that will not speak it, so the line is held here in the shape it parses.
	version, exitCode := runInstalled(t, binary, t.TempDir(), "--version")
	if exitCode != 0 || !strings.HasPrefix(version, "cohere 1.0.0\n") ||
		!strings.Contains(version, "\n  swift contract: "+strconv.Itoa(swiftContractVersion)+"\n") {
		t.Fatalf("a released cohere's --version does not state its Swift contract (exit %d):\n%s", exitCode, version)
	}
}

// A released cohere with no engine beside it is a broken install, and says so, even standing inside a
// cohere checkout it could build one from: an engine built there would run unpaired rules under the
// release's version.
func TestAReleasedCohereWithNoEngineBesideItNeverBuildsOne(t *testing.T) {
	t.Parallel()

	binary := installCohere(t, "1.0.0")
	checkout := t.TempDir()
	writeTree(t, checkout, map[string]string{
		"go.mod":                       "module github.com/system-inc/cohere\n",
		"swift/Package.swift":          "// swift-tools-version:6.0\n",
		"swift/Sources/Engine/A.swift": "let a = 1\n",
	})
	root := swiftPackage(t, checkout)

	// --verbose, so a build from the checkout would say so (see above).
	output, exitCode := runInstalled(t, binary, root, "--no-fix", "--verbose")
	if exitCode != 1 {
		t.Fatalf("a released cohere with no engine exited %d, expected 1:\n%s", exitCode, output)
	}
	// Named as the run names it, with the executable's symlinks resolved, which on macOS turns /tmp into
	// /private/tmp.
	installed, err := filepath.EvalSymlinks(filepath.Dir(binary))
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(installed, release.SwiftEngineFileName)
	if !strings.Contains(output, "this cohere 1.0.0 install has no Swift engine: expected "+expected) ||
		!strings.Contains(output, "reinstall cohere") {
		t.Fatalf("the missing engine was not named with the way out:\n%s", output)
	}
	if strings.Contains(output, "building the Swift engine") || strings.Contains(output, "swift --version") {
		t.Fatalf("a released cohere reached for the checkout:\n%s", output)
	}
}

// A development cohere never takes an engine from beside itself, so a stale cohere-swift next to a
// `go build -o` cannot run in place of the checkout's. Outside a checkout it has none, and says so.
func TestADevelopmentCohereIgnoresAnEngineBesideIt(t *testing.T) {
	t.Parallel()

	binary := installCohere(t, "")
	invoked := placeSiblingEngine(t, binary)
	root := swiftPackage(t, "")

	output, exitCode := runInstalled(t, binary, root, "--no-fix")
	if exitCode != 1 || !strings.Contains(output, "the Swift engine is built from a cohere checkout") {
		t.Fatalf("a development cohere outside a checkout exited %d:\n%s", exitCode, output)
	}
	if _, err := os.Stat(invoked); err == nil {
		t.Fatalf("a development cohere ran the engine beside it:\n%s", output)
	}
}

// Where the engine does not ship, a Swift package is refused by name before anything is looked for or
// built, whichever kind of cohere is running. The engine named by the override still runs: it is the
// engine developer's, and naming one is saying what to run.
// Not parallel: it sets the Swift engine override variable with t.Setenv, the working directory with t.Chdir,
// and the package-level swiftEnginePlatform
func TestSwiftIsRefusedByNameWhereTheEngineDoesNotShip(t *testing.T) {
	t.Setenv(dispatch.SwiftEngineOverrideVariable, "")
	// Outside any cohere module, so a resolver that skipped the refusal fails at once on the checkout it
	// cannot find, rather than starting a Swift build from the module this test runs in.
	t.Chdir(t.TempDir())
	original := swiftEnginePlatform
	t.Cleanup(func() { swiftEnginePlatform = original })
	location := projectLocation{Root: "/work/Toy"}
	for _, platform := range []release.Target{
		{GoOperatingSystem: "linux", GoArchitecture: "amd64"},
		{GoOperatingSystem: "linux", GoArchitecture: "arm64"},
		{GoOperatingSystem: "windows", GoArchitecture: "amd64"},
		{GoOperatingSystem: "windows", GoArchitecture: "arm64"},
	} {
		swiftEnginePlatform = platform
		for _, provenance := range []release.Provenance{{Version: "1.0.0"}, {Version: "dev"}} {
			_, _, err := resolveSwiftEngineBinary(location, provenance)
			want := "Swift is not available on this platform (" + platform.String() +
				"): cohere-swift ships only for macOS, so nothing in /work/Toy was checked"
			if err == nil || err.Error() != want {
				t.Errorf("%s, cohere %s: got %v\nwant %s", platform, provenance.Version, err, want)
			}
		}
	}

	// The control: on macOS the same released cohere gets past the platform, to its sibling.
	swiftEnginePlatform = release.Target{GoOperatingSystem: "darwin", GoArchitecture: "arm64"}
	if _, _, err := resolveSwiftEngineBinary(location, release.Provenance{Version: "1.0.0"}); err == nil ||
		!strings.Contains(err.Error(), "install has no Swift engine") {
		t.Errorf("on macOS a released cohere was not sent to its sibling: %v", err)
	}

	engine := fakeSwiftEngine(t, "")
	t.Setenv(dispatch.SwiftEngineOverrideVariable, engine)
	swiftEnginePlatform = release.Target{GoOperatingSystem: "linux", GoArchitecture: "amd64"}
	if binaryPath, _, err := resolveSwiftEngineBinary(location, release.Provenance{Version: "1.0.0"}); err != nil || binaryPath != engine {
		t.Errorf("the override was refused on linux: %q, %v", binaryPath, err)
	}
}
