package dispatch

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

// SwiftEngineOverrideVariable points cohere at a specific Swift engine binary, the counterpart of
// COHERE_BINARY for cohere itself.
//
// For whoever is developing the engine, and for running a known build while the sources on disk are
// mid-edit. An explicit binary wins over the rebuild for the reason resolveBinary gives: someone who
// names a binary has said what they want to run. It is announced on every run, because the findings
// then come from a binary no hash vouches for.
const SwiftEngineOverrideVariable = "COHERE_SWIFT_ENGINE"

// SwiftEngineDirectory is where the Swift engine's package lives inside the cohere module.
func (paths Paths) SwiftEngineDirectory() string {
	return filepath.Join(paths.ModuleDirectory, "swift")
}

// SwiftBuildDirectory is the scratch path the engine is built in.
//
// Its own, rather than the package's `.build`: that one belongs to whoever is developing the engine,
// and a build here taking its lock, or replacing its products mid-test, would be one process breaking
// another for no reason a reader of either could see.
func (paths Paths) SwiftBuildDirectory() string {
	return filepath.Join(paths.CacheDirectory, "swift-build")
}

// SwiftEngineBinaryPath is where the engine built from inputs with this hash lives, named the way
// BinaryPath names cohere's, platform included.
func (paths Paths) SwiftEngineBinaryPath(hash string) string {
	return filepath.Join(paths.BinaryDirectory(), fmt.Sprintf("cohere-swift-%s-%s-%s", runtime.GOOS, runtime.GOARCH, hash))
}

// ResolveSwiftEngine returns the path to a cohere-swift binary matching the engine's sources and the
// installed toolchain, building it when there is none.
//
// The same bargain Resolve makes for cohere, for the same reason: the rules are compiled in, so a
// stale binary runs old rules and looks exactly like a current one. The hash covers what decides the
// binary, which for Swift is the package's sources and manifests and the toolchain. The toolchain is
// in the hash because the engine links its libraries and reads its compiler's serialized diagnostics,
// so the same sources under a new Xcode are a different engine.
//
// What the hash does not cover is the one bit the manifest reads from git: whether the repository had
// uncommitted changes when the engine was built, which the engine reports as `sourceTreeModified`.
// Asking git on every run would cost more than the rest of the resolve together, and Go's own build
// stamp behaves the same way for the cohere binary. So a binary built from a modified tree goes on
// saying so after the change is committed, until its sources move. That errs toward the warning,
// which is the direction to err.
//
// That is the working-tree half. When commit is not empty the engine is built from that commit
// instead, which is what a gate gets: see resolveCommittedSwiftEngine. The contract is the one the
// caller speaks, used only to ask a freshly built committed engine for its provenance.
//
// The returned boolean reports whether a build ran.
func ResolveSwiftEngine(paths Paths, commit string, contract int) (string, bool, error) {
	toolchain, err := swiftToolchain()
	if err != nil {
		return "", false, err
	}

	if commit != "" {
		return resolveCommittedSwiftEngine(paths, commit, toolchain, contract)
	}

	hash, err := swiftEngineHash(paths.SwiftEngineDirectory(), toolchain)
	if err != nil {
		return "", false, err
	}

	binaryPath := paths.SwiftEngineBinaryPath(hash)
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, false, nil
	}

	if err := buildSwiftProduct(paths.SwiftEngineDirectory(), paths.SwiftBuildDirectory(), binaryPath); err != nil {
		return "", false, err
	}
	return binaryPath, true, nil
}

// swiftToolchain identifies the installed Swift toolchain, as `swift --version` states it.
//
// A missing toolchain is an error naming the command that failed, never a fallback. Without one the
// engine cannot be built and could not read the compiler's diagnostics if it were, so there is no
// engine to fall back to: only a run that checks nothing.
func swiftToolchain() (string, error) {
	output, err := exec.Command("swift", "--version").Output()
	if err != nil {
		return "", fmt.Errorf("no Swift toolchain answered `swift --version` (%w), so the Swift engine cannot be built and nothing was checked", err)
	}
	identity := strings.Join(strings.Fields(string(output)), " ")
	if identity == "" {
		return "", fmt.Errorf("`swift --version` printed nothing, so the toolchain the Swift engine would link against is unknown")
	}
	return identity, nil
}

// swiftEngineHash digests every input that decides what the engine binary does.
//
// `Package.swift` and `Package.resolved` pin the package graph, and every file under `Sources/` is
// compiled in. Tests are left out because they are not in the binary, and `.build` because it is
// output. Contents are hashed, never times, for the reason Inputs.Compute gives.
func swiftEngineHash(engineDirectory string, toolchain string) (string, error) {
	files := []string{
		filepath.Join(engineDirectory, "Package.swift"),
		filepath.Join(engineDirectory, "Package.resolved"),
	}

	sources := filepath.Join(engineDirectory, "Sources")
	walkErr := filepath.WalkDir(sources, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	if walkErr != nil {
		return "", fmt.Errorf("listing the Swift engine's sources under %s: %w", sources, walkErr)
	}
	if len(files) == 2 {
		// The same guard CollectInputs keeps: an empty source set hashes to a stable digest for a tree
		// holding nothing, and every later run would hit it.
		return "", fmt.Errorf("found no Swift engine sources under %s, which cannot be right", sources)
	}
	sort.Strings(files)

	digest := sha256.New()
	fmt.Fprintf(digest, "toolchain\x00%s\x00", toolchain)
	fmt.Fprintf(digest, "platform\x00%s/%s\x00", runtime.GOOS, runtime.GOARCH)
	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading Swift engine input %s: %w", path, err)
		}
		// The path relative to the package, rather than the base name Inputs.Compute uses: two files
		// named `Rule.swift` in different directories are different inputs, and moving one between
		// targets changes the binary.
		relative, err := filepath.Rel(engineDirectory, path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(digest, "file\x00%s\x00%d\x00", filepath.ToSlash(relative), len(contents))
		digest.Write(contents)
	}
	return hex.EncodeToString(digest.Sum(nil))[:hashLength], nil
}

// buildSwiftProduct builds the engine package at packageDirectory and copies the product to binaryPath.
//
// A variable so a test can stand in for SwiftPM: a cold build compiles swift-syntax and swift-format,
// measured at 225s, and the decisions worth testing are which tree is built and when, not SwiftPM.
var buildSwiftProduct = buildSwiftProductWithSwiftPM

// lastLines is a tool's kept output for an error: its last count lines, on lines of their own after the
// error's, or nothing when it printed nothing.
func lastLines(output string, count int) string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return ""
	}
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}
	return "\n" + strings.Join(lines, "\n")
}

// buildToolGrace is how long a build tool's group has to end after a terminate before it is killed.
const buildToolGrace = 5 * time.Second

// buildToolWaitDelay bounds how long Wait waits on a tool's output pipes once the tool has ended, so a
// compiler that left the group while holding a pipe cannot hold this process with it.
const buildToolWaitDelay = 2 * time.Second

// runBuildTool runs a build tool as the leader of a process group of its own, and ends the whole group if
// this process is interrupted, terminated or hung up while the tool runs.
//
// Without it a signal ended this process alone and left `swift build` and its compilers running, with
// nobody to read their result, and holding SwiftPM's lock on the scratch path so the next run's build waited
// behind them (#nqb3mjv). A kill this process cannot catch still leaves them: nothing can forward that.
func runBuildTool(command *exec.Cmd) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signals)
	return runBuildToolUntil(command, signals)
}

// runBuildToolUntil is runBuildTool with the signals given, so a test can send one without signaling
// itself.
func runBuildToolUntil(command *exec.Cmd, signals <-chan os.Signal) error {
	startInOwnGroup(command)
	command.WaitDelay = buildToolWaitDelay
	if err := command.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()

	select {
	case err := <-done:
		return err
	case received := <-signals:
		signalGroup(command, syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(buildToolGrace):
			signalGroup(command, syscall.SIGKILL)
			<-done
		}
		return fmt.Errorf("stopped by %v, and everything it started was ended with it", received)
	}
}

// buildSwiftProductWithSwiftPM builds the engine in release mode and copies the product to binaryPath.
//
// A cold build compiles swift-syntax and swift-format and takes minutes, so it announces itself on
// stderr before starting, whatever the verbosity: a command that is silent for three minutes reads as
// hung. SwiftPM's progress goes to stderr too, because this process's stdout is the report.
//
// Copied rather than linked, because the next build replaces the product in place, and a link would
// turn every hash-named binary into whichever engine was built last.
func buildSwiftProductWithSwiftPM(packageDirectory string, scratchDirectory string, binaryPath string) error {
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		return fmt.Errorf("creating the binary cache directory: %w", err)
	}

	Report.Step("cohere: building the Swift engine from %s (its sources or the toolchain changed; a cold build takes minutes)",
		packageDirectory)

	// `--manifest-cache none`, because the manifest reads the tree's git state into the engine's
	// `sourceTreeModified`, and SwiftPM's shared manifest cache answers it from an earlier evaluation of
	// the same commit, whatever the scratch path. An engine rebuilt after an edit to a tracked file would
	// then say it was built from a clean tree. Measured by @system_cohere_release, whose release build
	// passes the same flag.
	arguments := []string{
		"build", "-c", "release",
		"--package-path", packageDirectory,
		"--scratch-path", scratchDirectory,
		"--manifest-cache", "none",
		"--product", "cohere-swift",
	}
	// SwiftPM's own lines stream under --verbose. Otherwise they are kept, so a default run prints cohere's
	// report alone (#ytqqv8v), and a build that fails quotes their end in its error.
	build := exec.Command("swift", arguments...)
	var kept bytes.Buffer
	build.Stdout, build.Stderr = &kept, &kept
	if stream := Report.ToolOutput(); stream != nil {
		build.Stdout, build.Stderr = stream, stream
	}
	if err := runBuildTool(build); err != nil {
		return fmt.Errorf("building the Swift engine (swift %s): %w%s", strings.Join(arguments, " "), err, lastLines(kept.String(), 40))
	}

	binPath, err := exec.Command("swift", append(arguments, "--show-bin-path")...).Output()
	if err != nil {
		return fmt.Errorf("asking SwiftPM where it put the Swift engine: %w", err)
	}
	product := filepath.Join(strings.TrimSpace(string(binPath)), "cohere-swift")

	// The build said it worked; the file is what gets run. The same gap build() closes for cohere.
	if err := copyExecutable(product, binaryPath); err != nil {
		return fmt.Errorf("swift build reported success but the engine could not be taken from %s: %w", product, err)
	}
	return nil
}

// copyExecutable copies a file to a new path through a temporary name in the same directory, so a
// concurrent run never sees, and execs, a half-written binary.
func copyExecutable(source string, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	temporary, err := os.CreateTemp(filepath.Dir(destination), filepath.Base(destination)+".partial-*")
	if err != nil {
		return err
	}
	// Removed on every failure path below. After the rename the name no longer exists and the call
	// does nothing.
	defer os.Remove(temporary.Name())

	if _, err := io.Copy(temporary, input); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o755); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), destination)
}
