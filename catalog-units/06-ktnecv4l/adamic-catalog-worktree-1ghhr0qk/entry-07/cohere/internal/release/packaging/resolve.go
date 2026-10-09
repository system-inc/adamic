package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// BinaryOverrideVariable points the dispatcher at a specific binary.
//
// This is the development affordance: build cohere from source, export this, and every `cohere` in
// every repo on the machine runs that build instead of the installed one. It is deliberately an
// absolute-path override rather than a directory to search, so that what it selects is stated
// rather than resolved.
const BinaryOverrideVariable = "COHERE_BINARY"

// PlatformPackagePrefix is what every platform binary's package name begins with, the "@system-inc/
// cohere-" of "@system-inc/cohere-darwin-arm64".
//
// The platform packages share the dispatcher's scope and carry its name as a stem, the shape biome
// (`@biomejs/cli-darwin-arm64`) and rollup (`@rollup/rollup-darwin-arm64`) publish under, rather than
// esbuild's scope of their own (`@esbuild/darwin-arm64`). esbuild can drop the stem because its
// scope holds nothing else. `@system-inc` holds other tools, so `@system-inc/darwin-arm64` would not
// say whose binary it is, and a second tool shipping binaries the same way would collide with it.
const PlatformPackagePrefix = DispatcherPackageName + "-"

// ErrNoBinary reports that no cohere binary could be found for this platform.
//
// It is a distinct error so that a caller can tell "we looked and there is nothing for this
// platform" apart from "we could not look". Both are failures; only one is a packaging bug.
var ErrNoBinary = errors.New("no cohere binary for this platform")

// PlatformPackageName is the npm package holding the binary for a Go GOOS and GOARCH, like
// "@system-inc/cohere-darwin-arm64".
//
// The arguments are Go's names and the result is npm's, because npm is what resolves the package
// while Go is what built it. They disagree in two places: Go says `windows` where npm says `win32`,
// and Go says `amd64` where npm says `x64`. Publishing under the Go spelling produces a package
// that installs correctly and is never found, which on the machine is indistinguishable from a
// platform we never shipped — so the translation happens here, once, and every caller passes Go
// names in.
func PlatformPackageName(goOperatingSystem string, goArchitecture string) string {
	return fmt.Sprintf(
		"%s%s-%s",
		PlatformPackagePrefix,
		NpmOperatingSystem(goOperatingSystem),
		NpmArchitecture(goArchitecture),
	)
}

// CurrentPlatformPackageName is the platform package for the running machine.
func CurrentPlatformPackageName() string {
	return PlatformPackageName(runtime.GOOS, runtime.GOARCH)
}

// NpmOperatingSystem translates a Go GOOS to the name npm's `os` field uses.
//
// Only the disagreement is listed. `darwin` and `linux` are already npm's spelling, and mapping
// them explicitly would just be a second place to keep in sync.
func NpmOperatingSystem(goOperatingSystem string) string {
	if goOperatingSystem == "windows" {
		return "win32"
	}
	return goOperatingSystem
}

// NpmArchitecture translates a Go GOARCH to the name npm's `cpu` field uses.
func NpmArchitecture(goArchitecture string) string {
	switch goArchitecture {
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	default:
		return goArchitecture
	}
}

// BinaryFileName is the name of the executable inside a platform package.
//
// Windows needs the extension to execute at all, so the name carries the platform rather than being
// a bare "cohere" everywhere.
func BinaryFileName(operatingSystem string) string {
	if operatingSystem == "windows" {
		return "cohere.exe"
	}
	return "cohere"
}

// Resolve returns the cohere binary to run on this machine, or an error naming what it looked for.
//
// The search order is override, then the platform package, and there is no third step. This is the
// rule this package exists to hold, and it is worth restating at the place it is enforced: the gate
// cohere replaces printed a green checkmark over zero files for days because a resolver walked up
// looking for a binary, found none, fell through to the bare name `oxlint`, and the failed spawn
// produced an empty file list — which is indistinguishable from a clean tree. So there is no
// fallback to PATH, no fallback to a bare command name, and no fallback to a binary built for
// another platform. A missing binary exits non-zero naming the platform, every time.
//
// searchRoots are the directories to look for `node_modules` in, nearest first.
func Resolve(searchRoots []string) (string, error) {
	if override := strings.TrimSpace(os.Getenv(BinaryOverrideVariable)); override != "" {
		return resolveOverride(override)
	}

	packageName := CurrentPlatformPackageName()
	binaryFileName := BinaryFileName(runtime.GOOS)

	attempted := []string{}
	for _, root := range searchRoots {
		for _, candidate := range binaryCandidates(root, packageName, binaryFileName) {
			attempted = append(attempted, candidate)

			information, err := os.Stat(candidate)
			if err != nil {
				continue
			}
			if information.IsDir() {
				continue
			}
			if err := requireExecutable(candidate, information); err != nil {
				// A binary that exists but cannot run is reported rather than skipped. Continuing
				// past it would end in "no binary for this platform", which sends whoever reads it
				// looking for a packaging bug that is not there — the file shipped fine, its mode
				// did not survive.
				return "", err
			}
			return candidate, nil
		}
	}

	return "", fmt.Errorf(
		"%w: %s/%s wants %s, and it is not installed.\nLooked in:\n  %s\nInstall %s without skipping optional dependencies, which brings it in, or point %s at a local build.",
		ErrNoBinary, runtime.GOOS, runtime.GOARCH, packageName,
		strings.Join(attempted, "\n  "), DispatcherPackageName, BinaryOverrideVariable,
	)
}

// resolveOverride validates an explicitly requested binary.
//
// Every failure here is loud. Someone who sets this variable has stated which binary they want, so
// silently running a different one would be the worst possible outcome: they would read the results
// as coming from their build. A stale export in a shell profile is exactly how that happens, which
// is why a missing path says what was asked for and which variable asked for it.
func resolveOverride(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("%s is set to %q, which is not a usable path: %w", BinaryOverrideVariable, path, err)
	}

	information, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("%s points at %s, and there is nothing there: %w", BinaryOverrideVariable, absolutePath, err)
	}
	if information.IsDir() {
		return "", fmt.Errorf("%s points at %s, which is a directory rather than a binary", BinaryOverrideVariable, absolutePath)
	}
	if err := requireExecutable(absolutePath, information); err != nil {
		return "", err
	}

	return absolutePath, nil
}

// requireExecutable reports whether a file can actually be run.
//
// Without this the failure surfaces at exec as a permission error, which reads like a machine
// problem rather than a packaging one. The most common cause is a packaging step that lost the
// mode bit — a tarball repacked without preserving permissions — so the message names that.
func requireExecutable(path string, information os.FileInfo) error {
	if runtime.GOOS == "windows" {
		// Windows has no executable bit; the extension decides. Checking the mode there would
		// reject every correctly packaged binary.
		return nil
	}
	if information.Mode()&0o111 == 0 {
		// The mode is measured; the cause is not. Packaging that dropped the bit is the common one,
		// but a restrictive umask on this machine, a noexec mount, or an extraction that ignored
		// modes all land here identically. Asserting the first would send a reader to inspect a
		// tarball that is fine.
		return fmt.Errorf(
			"the cohere binary at %s has mode %s, so it cannot be executed.\nMost often the mode was lost in packaging or extraction; a restrictive umask or a noexec mount produce it too",
			path, information.Mode(),
		)
	}
	return nil
}

// binaryCandidates lists where the platform binary could live under one search root, nearest first.
func binaryCandidates(root string, packageName string, binaryFileName string) []string {
	nodeModules := filepath.Join(root, "node_modules")
	return []string{
		// The normal install: pnpm, npm, and yarn all place an optional platform dependency here.
		filepath.Join(nodeModules, filepath.FromSlash(packageName), "bin", binaryFileName),

		// pnpm's isolated store keeps real packages under `.pnpm/` and symlinks the top level. The
		// symlink is what the line above follows, so this is only reached when the top-level link
		// is missing while the package is genuinely installed — which happens with
		// `node-linker=isolated` and a dependency that only a nested package depends on.
		filepath.Join(nodeModules, ".pnpm", "node_modules", filepath.FromSlash(packageName), "bin", binaryFileName),
	}
}

// SearchRoots returns the directories to look for `node_modules` in, nearest first.
//
// Both the working directory and the dispatcher's own location are walked, and every parent of
// each. The working directory finds the install in the repo being checked; the executable's own
// path finds it when cohere is invoked from somewhere else entirely, which is what happens when a
// tool runs it by absolute path. Walking up handles monorepos, where the binary is hoisted to the
// workspace root rather than installed beside the package being checked.
func SearchRoots(workingDirectory string, executablePath string) []string {
	roots := []string{}
	seen := map[string]bool{}

	add := func(start string) {
		if start == "" {
			return
		}
		directory := start
		for {
			if !seen[directory] {
				seen[directory] = true
				roots = append(roots, directory)
			}
			parent := filepath.Dir(directory)
			if parent == directory {
				return
			}
			directory = parent
		}
	}

	add(workingDirectory)
	if executablePath != "" {
		add(filepath.Dir(executablePath))
	}

	return roots
}
