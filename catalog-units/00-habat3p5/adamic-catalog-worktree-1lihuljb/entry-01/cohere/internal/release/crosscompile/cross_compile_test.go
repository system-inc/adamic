// Package crosscompile proves cohere compiles for every platform a release ships. It reads the whole
// program, so any edit reruns it, and it lives apart from internal/release/packaging so that the edit
// reruns it alone and not packaging's Swift and staging tests too (#nxgt2ca).
package crosscompile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	release "github.com/system-inc/cohere/internal/release/packaging"
)

// crossCompileVariable chooses how TestEveryReleaseTargetCompiles proves the targets. Unset, it
// type-checks cohere for each one from source, here and in release.yml. Set to "build", it builds each one
// as the release does, for whoever wants that proof from a test.
//
// Measured on private caches (#nkbfkvn), for all six targets: building costs 17.1 GB of Go cache cold,
// and 5.5 GB more after one real line in internal/lint/rule, because every package that imports it
// recompiles six times. Type-checking from source costs 12 MB cold and nothing after the same edit, in
// about five seconds, since nothing is compiled. `go vet` per target was no better than building, at
// 2.7 GB cold for one target: it compiles every dependency for its export data. Building on every
// local `go test ./...` was most of the 25 GB an hour that filled the disk on 2026-10-03.
//
// Type-checking catches what the Windows break was, a call to something one platform does not define,
// and any type error in a file only one platform compiles. What only a build catches, a link-time
// failure, is left to release.yml's staging, which builds all six right after its tests and fails on
// any that does not build. Asking the test to build them there as well cost a three-core runner 600
// seconds and a timeout (dry run 37238149770), for a proof staging gives anyway.
const crossCompileVariable = "COHERE_CROSS_COMPILE"

// TestEveryReleaseTargetCompiles builds cohere for every platform a release ships, so a target that
// stops compiling fails `go test` instead of the next release.
//
// `go test` otherwise builds only for the host. cohere stopped compiling for windows at `11a25134`
// and nothing in the ordinary gate noticed, because the break was in a Unix-only process-group call
// and every machine that ran the tests was a Mac. The first place it would have shown up was a
// release, the most expensive place there is to find it.
//
// It ranges over release.Targets itself rather than a list of its own, because a second list is how a
// dropped or renamed platform stops being checked without anything failing. It compiles through
// release.GoBuildCommand, the same command the release runs, for the same reason.
//
// It type-checks rather than builds, locally and in release.yml; crossCompileVariable says why, with the
// measurements. `-short` skips it for a quick local loop and says so.
func TestEveryReleaseTargetCompiles(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skipf("NOT MEASURED: -short skips checking cohere for the %d release targets, so a platform that no longer builds would pass here", len(release.Targets))
	}
	mode := os.Getenv(crossCompileVariable)
	if mode != "" && mode != "build" {
		t.Fatalf("%s is %q; it is unset to type-check each target, or \"build\" to build each one", crossCompileVariable, mode)
	}

	moduleDirectory := filepath.Join("..", "..", "..")

	for _, target := range release.Targets {
		// Six type-checks at once hold about 5.4 GB, 900 MB each. One after another they were most of
		// packaging's wall on every edit, about 6s against about 2s in parallel (#nxgt2ca).
		t.Run(target.String(), func(t *testing.T) {
			t.Parallel()
			if mode != "build" {
				if problems := typeCheckTarget(moduleDirectory, target); len(problems) > 0 {
					t.Fatalf("cohere does not type-check for %s, so a release would fail there:\n%s", target, strings.Join(problems, "\n"))
				}
				return
			}
			command := release.GoBuildCommand(moduleDirectory, target, strings.Join(release.StripFlags, " "), filepath.Join(t.TempDir(), target.BinaryFileName()))
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("cohere does not compile for %s, so a release would fail there: %v\n%s", target, err, output)
			}
		})
	}
}

// typeCheckTarget type-checks the package the release builds, and everything it imports, from source as
// target would compile it, and returns every error, or none. Nothing is compiled, so nothing is written
// to the build cache.
func typeCheckTarget(moduleDirectory string, target release.Target) []string {
	return typeCheck(moduleDirectory, target, false, "./command/cohere")
}

// typeCheck type-checks the packages patterns name, and everything they import, from source as target
// would compile them, their tests too when tests is set, and returns every error once, or none.
func typeCheck(moduleDirectory string, target release.Target, tests bool, patterns ...string) []string {
	loaded, err := packages.Load(&packages.Config{
		// Syntax and type information for every dependency is what makes go/packages type-check them all
		// from source rather than reading export data, which it would have to compile first.
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:   moduleDirectory,
		Env:   append(os.Environ(), "GOOS="+target.GoOperatingSystem, "GOARCH="+target.GoArchitecture, "CGO_ENABLED=0"),
		Tests: tests,
	}, patterns...)
	if err != nil {
		return []string{err.Error()}
	}
	var problems []string
	// With tests a package is loaded twice, alone and with its test files, and an error in the code both
	// share is reported by each, so each is kept once.
	seen := map[string]bool{}
	checked := 0
	packages.Visit(loaded, nil, func(pkg *packages.Package) {
		checked++
		for _, problem := range pkg.Errors {
			if !seen[problem.Error()] {
				seen[problem.Error()] = true
				problems = append(problems, problem.Error())
			}
		}
	})
	if checked == 0 {
		return []string{fmt.Sprintf("nothing was loaded for %s, so nothing was checked", target)}
	}
	return problems
}

// TestTypeCheckCatchesAPlatformOnlyBreak is the known-dirty control for the local path: the darwin-only
// call the build control below uses, and a type error only a windows file holds, both fail the
// windows type-check and neither fails darwin's.
func TestTypeCheckCatchesAPlatformOnlyBreak(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "go.mod"), "module crosscompile\n\ngo 1.21\n")
	writeFile(t, filepath.Join(directory, "command", "cohere", "main.go"), "package main\n\nfunc main() { onlyOnDarwin() }\n")
	writeFile(t, filepath.Join(directory, "command", "cohere", "only_darwin.go"), "package main\n\nfunc onlyOnDarwin() {}\n")
	writeFile(t, filepath.Join(directory, "command", "cohere", "broken_windows.go"), "package main\n\nvar broken int = \"text\"\n")

	if problems := typeCheckTarget(directory, release.Target{GoOperatingSystem: "darwin", GoArchitecture: "arm64"}); len(problems) > 0 {
		t.Fatalf("the control does not type-check even for darwin, so it cannot show anything: %v", problems)
	}
	problems := strings.Join(typeCheckTarget(directory, release.Target{GoOperatingSystem: "windows", GoArchitecture: "amd64"}), "\n")
	for _, want := range []string{"onlyOnDarwin", "broken_windows.go"} {
		if !strings.Contains(problems, want) {
			t.Errorf("the windows type-check did not report %s: %q", want, problems)
		}
	}
}

// TestEveryPackagesTestsTypeCheckForWindows type-checks every package of the module for Windows, its tests
// included, so a test that cannot build there fails here and not on a Windows machine someone finally runs
// it on.
//
// Release targets prove only what ships: command/cohere's tests used a Unix-only stand-in engine for weeks
// without anything noticing, since only three named Windows tests run in release.yml (#tejf9bc). Windows
// alone, because it is the one shipped platform that is not Unix: a file a Unix build constraint keeps
// out of linux builds would also be kept out of darwin's, which every test run here already compiles.
func TestEveryPackagesTestsTypeCheckForWindows(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("NOT MEASURED: -short skips type-checking the module's tests for Windows, so a test that no longer builds there would pass here")
	}
	var windows release.Target
	for _, target := range release.Targets {
		if target.GoOperatingSystem == "windows" {
			windows = target
			break
		}
	}
	if windows.GoOperatingSystem == "" {
		t.Fatal("no release target is Windows, so there is nothing for this test to check against")
	}
	if problems := typeCheck(filepath.Join("..", "..", ".."), windows, true, "./..."); len(problems) > 0 {
		t.Fatalf("the module's tests do not type-check for %s:\n%s", windows, strings.Join(problems, "\n"))
	}
}

// TestTypeCheckOfTestsCatchesAUnixOnlyTestHelper is the known-dirty control for the test above: a test
// that calls a helper only Unix defines type-checks for darwin and fails for windows, and only when tests
// are checked.
func TestTypeCheckOfTestsCatchesAUnixOnlyTestHelper(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "go.mod"), "module crosscompile\n\ngo 1.21\n")
	writeFile(t, filepath.Join(directory, "tool", "tool.go"), "package tool\n")
	writeFile(t, filepath.Join(directory, "tool", "helper_unix_test.go"), "//go:build unix\n\npackage tool\n\nfunc stand() string { return \"/bin/sh\" }\n")
	writeFile(t, filepath.Join(directory, "tool", "tool_test.go"), "package tool\n\nimport \"testing\"\n\nfunc TestTool(t *testing.T) { _ = stand() }\n")

	darwin := release.Target{GoOperatingSystem: "darwin", GoArchitecture: "arm64"}
	windows := release.Target{GoOperatingSystem: "windows", GoArchitecture: "amd64"}
	if problems := typeCheck(directory, darwin, true, "./..."); len(problems) > 0 {
		t.Fatalf("the control does not type-check even for darwin, so it cannot show anything: %v", problems)
	}
	if problems := typeCheck(directory, windows, false, "./..."); len(problems) > 0 {
		t.Fatalf("the control's code alone failed for windows, so it does not isolate the tests: %v", problems)
	}
	if problems := strings.Join(typeCheck(directory, windows, true, "./..."), "\n"); !strings.Contains(problems, "stand") {
		t.Errorf("the windows type-check of the tests did not report the Unix-only helper: %q", problems)
	}
}

// TestCrossCompileCatchesAPlatformOnlyBreak is the known-dirty control for the test above. A check
// that has never failed has not been shown to check anything, and the shape here is the shape of the
// windows break that motivated it: code that calls something only one platform defines.
func TestCrossCompileCatchesAPlatformOnlyBreak(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "go.mod"), "module crosscompile\n\ngo 1.21\n")
	writeFile(t, filepath.Join(directory, "command", "cohere", "main.go"), "package main\n\nfunc main() { onlyOnDarwin() }\n")
	writeFile(t, filepath.Join(directory, "command", "cohere", "only_darwin.go"), "package main\n\nfunc onlyOnDarwin() {}\n")

	darwin := release.Target{GoOperatingSystem: "darwin", GoArchitecture: "arm64"}
	if output, err := release.GoBuildCommand(directory, darwin, "", filepath.Join(t.TempDir(), "cohere")).CombinedOutput(); err != nil {
		t.Fatalf("the control does not compile even for darwin, so it cannot show anything: %v\n%s", err, output)
	}

	windows := release.Target{GoOperatingSystem: "windows", GoArchitecture: "amd64"}
	output, err := release.GoBuildCommand(directory, windows, "", filepath.Join(t.TempDir(), "cohere.exe")).CombinedOutput()
	if err == nil {
		t.Fatal("a call to a darwin-only function compiled for windows, so the cross-compile check cannot catch a platform-only break")
	}
	if !strings.Contains(string(output), "onlyOnDarwin") {
		t.Fatalf("the windows build failed, but not on the darwin-only call: %s", output)
	}
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
