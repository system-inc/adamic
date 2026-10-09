package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// Every check here is shown failing on a tampered binary as well as passing on an intact one, because a
// verification that has never once refused anything has not been shown to look at the bytes. The tamper
// is one flipped byte, the smallest change there is, and in the launcher's fixture it is a flip that
// leaves the script runnable: if the launcher ran it anyway, its output would say so.

// placeholderChecksums stands in for a release's sums where a test builds the dispatcher for some other
// reason, since a dispatcher with none is refused.
var placeholderChecksums = []byte(strings.Repeat("0", 64) + "  cohere-linux-x64/bin/cohere\n")

// fixtureScript is the stand-in platform binary. The tamper flips the case of its first letter, which
// keeps it a working script, so a launcher that ran it after all would print "Fixture ran".
const fixtureScript = "#!/bin/sh\necho \"fixture ran: $*\"\n"

func TestChecksumsCheckWithShasumAndFailOnAFlippedByte(t *testing.T) {
	t.Parallel()

	// shasum is what the README tells a reader to run, so the format is proven by shasum rather than by a
	// parser of ours, which would agree with whatever this package wrote.
	shasum, err := exec.LookPath("shasum")
	if err != nil {
		t.Fatalf("shasum is not installed, and it is the checker these sums are written for: %v", err)
	}

	output := t.TempDir()
	stageFixturePlatform(t, output, Target{GoOperatingSystem: "darwin", GoArchitecture: "arm64"})
	stageFixturePlatform(t, output, Target{GoOperatingSystem: "linux", GoArchitecture: "amd64"})

	checksums, err := Checksums(output, []string{"cohere-linux-x64", "cohere-darwin-arm64"})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(output, ChecksumsFileName), string(checksums))

	lines := strings.Split(strings.TrimSuffix(string(checksums), "\n"), "\n")
	wanted := []string{
		"cohere-darwin-arm64/bin/cohere",
		"cohere-darwin-arm64/bin/cohere-swift",
		"cohere-linux-x64/bin/cohere",
		"cohere-linux-x64/bin/cohere-swift",
	}
	if len(lines) != len(wanted) {
		t.Fatalf("SHA256SUMS has %d lines, wanted one per file in each bin/:\n%s", len(lines), checksums)
	}
	for index, line := range lines {
		if !strings.HasSuffix(line, "  "+wanted[index]) {
			t.Errorf("line %d is %q, wanted the file %s, sorted by path", index, line, wanted[index])
		}
	}

	check := exec.Command(shasum, "-a", "256", "-c", ChecksumsFileName)
	check.Dir = output
	if result, err := check.CombinedOutput(); err != nil {
		t.Fatalf("shasum refused the sums of the files they were taken from: %v\n%s", err, result)
	}

	flipByte(t, filepath.Join(output, "cohere-linux-x64", "bin", "cohere"))
	check = exec.Command(shasum, "-a", "256", "-c", ChecksumsFileName)
	check.Dir = output
	result, err := check.CombinedOutput()
	if err == nil {
		t.Fatalf("shasum passed a binary with a flipped byte:\n%s", result)
	}
	if !strings.Contains(string(result), "cohere-linux-x64/bin/cohere: FAILED") {
		t.Fatalf("shasum failed, but not by naming the flipped binary:\n%s", result)
	}
}

func TestChecksumsRefuseAPackageWithNothingToCheck(t *testing.T) {
	t.Parallel()

	output := t.TempDir()
	if err := os.MkdirAll(filepath.Join(output, "cohere-linux-x64", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if checksums, err := Checksums(output, []string{"cohere-linux-x64"}); err == nil {
		t.Fatalf("an empty bin/ produced sums %q, so a release with no binary would look checked", checksums)
	}

	if checksums, err := Checksums(output, []string{"cohere-darwin-arm64"}); err == nil {
		t.Fatalf("a package that was never staged produced sums %q", checksums)
	}
}

func TestChecksumsRefuseALinkInBin(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("creating a symbolic link needs a privilege Windows does not grant by default")
	}
	output := t.TempDir()
	stageFixturePlatform(t, output, Target{GoOperatingSystem: "linux", GoArchitecture: "amd64"})
	if err := os.Symlink("/bin/sh", filepath.Join(output, "cohere-linux-x64", "bin", "shell")); err != nil {
		t.Fatal(err)
	}
	if checksums, err := Checksums(output, []string{"cohere-linux-x64"}); err == nil {
		t.Fatalf("a link in bin/ was hashed as though it were the package's own file:\n%s", checksums)
	}
}

func TestDispatcherPublishesItsChecksums(t *testing.T) {
	t.Parallel()

	// Staged and left out of `files` is the failure this catches: npm publish would drop SHA256SUMS, and
	// every install would then refuse to run, from a package that looked complete in dist/.
	manifest := decodeManifest(t, mustDispatcherManifest(t, "1.2.3"))
	files, _ := manifest["files"].([]any)
	if !slices.Contains(files, any(ChecksumsFileName)) {
		t.Fatalf("the dispatcher's files are %v, so npm publish would leave %s out", files, ChecksumsFileName)
	}

	module, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	staged, err := buildDispatcherPackage(Options{ModuleDirectory: module, OutputDirectory: t.TempDir(), Version: "1.0.0"}, placeholderChecksums)
	if err != nil {
		t.Fatal(err)
	}
	shipped, err := os.ReadFile(filepath.Join(staged.Directory, ChecksumsFileName))
	if err != nil {
		t.Fatalf("the dispatcher package does not hold %s: %v", ChecksumsFileName, err)
	}
	if !bytes.Equal(shipped, placeholderChecksums) {
		t.Fatalf("the dispatcher holds %q, not the sums it was given", shipped)
	}

	if _, err := buildDispatcherPackage(Options{ModuleDirectory: module, OutputDirectory: t.TempDir(), Version: "1.0.0"}, nil); err == nil {
		t.Fatalf("a dispatcher was staged with no checksums, and its launcher would run nothing anywhere")
	}
}

// The launcher, run by Node the way npm's shim runs it, against a release staged into a node_modules
// layout by the same functions the release uses.

func TestLauncherRunsAnIntactRelease(t *testing.T) {
	t.Parallel()

	installation := stageInstalledRelease(t)
	result := installation.run(t, "--version")
	result.requireRan(t, "--version")

	if _, err := os.Stat(installation.verdictPath(t)); err != nil {
		t.Fatalf("a passing check left no remembered verdict, so every run would hash the binaries again: %v", err)
	}
}

func TestLauncherRefusesAFlippedByte(t *testing.T) {
	t.Parallel()

	installation := stageInstalledRelease(t)
	flipByte(t, installation.binary)

	result := installation.run(t, "--version")
	result.requireRefused(t, installation.binary, "does not match the checksum")
}

func TestLauncherRefusesATamperAfterAVerdictIsRemembered(t *testing.T) {
	t.Parallel()

	// The remembered verdict is the shortcut a tamper would hide behind: the binary passed once, then
	// changed. The change here keeps the size and puts the modification time back, so only the change
	// time and the bytes differ, which is everything a careful edit can leave behind.
	installation := stageInstalledRelease(t)
	installation.run(t, "--version").requireRan(t, "--version")

	before, err := os.Stat(installation.binary)
	if err != nil {
		t.Fatal(err)
	}
	flipByte(t, installation.binary)
	if err := os.Chtimes(installation.binary, before.ModTime(), before.ModTime()); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(installation.binary)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) || after.Size() != before.Size() {
		t.Fatalf("the tamper changed the modification time or the size, so this would not test the change time")
	}

	result := installation.run(t, "--version")
	result.requireRefused(t, installation.binary, "does not match the checksum")
}

func TestLauncherHonorsARememberedVerdictOnlyForTheChecksumItWasReachedOn(t *testing.T) {
	t.Parallel()

	// Two halves, so neither passes by accident. A new release names a new hash for a file whose bytes
	// have not moved, and the old verdict must not vouch for it. And a verdict that does match is
	// honored without hashing: shown by pointing both the sums and the verdict at a hash the file does
	// not have, which only a skipped hash would let run. That second half is what keeps the check off
	// every run's clock.
	installation := stageInstalledRelease(t)
	installation.run(t, "--version").requireRan(t, "--version")

	checksums, err := os.ReadFile(installation.checksums)
	if err != nil {
		t.Fatal(err)
	}
	actual := sha256Hex([]byte(fixtureScript))
	other := sha256Hex([]byte("another release's binary"))
	if !bytes.Contains(checksums, []byte(actual)) {
		t.Fatalf("the staged sums do not hold the fixture's hash %s:\n%s", actual, checksums)
	}
	writeFile(t, installation.checksums, strings.ReplaceAll(string(checksums), actual, other))

	installation.run(t, "--version").requireRefused(t, installation.binary, "does not match the checksum")

	verdict, err := os.ReadFile(installation.verdictPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(verdict, []byte(actual)) {
		t.Fatalf("the remembered verdict does not name the hash it was reached on:\n%s", verdict)
	}
	writeFile(t, installation.verdictPath(t), strings.ReplaceAll(string(verdict), actual, other))

	installation.run(t, "--version").requireRan(t, "--version")
}

func TestLauncherRefusesWithoutItsChecksums(t *testing.T) {
	t.Parallel()

	installation := stageInstalledRelease(t)
	if err := os.Remove(installation.checksums); err != nil {
		t.Fatal(err)
	}
	installation.run(t, "--version").requireRefused(t, installation.checksums, "is missing")
}

func TestLauncherRefusesChecksumsThatDoNotListThisPlatform(t *testing.T) {
	t.Parallel()

	installation := stageInstalledRelease(t)
	writeFile(t, installation.checksums, string(placeholderChecksums))
	if strings.Contains(string(placeholderChecksums), installation.directoryName+"/") {
		t.Fatalf("the placeholder lists %s, so this would not test a platform the sums leave out", installation.directoryName)
	}
	installation.run(t, "--version").requireRefused(t, installation.checksums, "lists no binary for")
}

func TestLauncherRefusesWhenAShippedFileIsMissing(t *testing.T) {
	t.Parallel()

	// cohere-swift, which cohere runs on a Mac without the launcher in between, so it is checked here or
	// nowhere.
	installation := stageInstalledRelease(t)
	engine := filepath.Join(filepath.Dir(installation.binary), SwiftEngineFileName)
	if err := os.Remove(engine); err != nil {
		t.Fatal(err)
	}
	installation.run(t, "--version").requireRefused(t, engine, "is missing")
}

func TestLauncherRefusesATamperedEngine(t *testing.T) {
	t.Parallel()

	installation := stageInstalledRelease(t)
	engine := filepath.Join(filepath.Dir(installation.binary), SwiftEngineFileName)
	flipByte(t, engine)
	installation.run(t, "--version").requireRefused(t, engine, "does not match the checksum")
}

func TestLauncherRunsTheOverrideUnchecked(t *testing.T) {
	t.Parallel()

	// Someone who names a binary has said what to run, usually their own build, which no release sum
	// describes. The installed binary is tampered with too, so a launcher that checked it anyway refuses.
	installation := stageInstalledRelease(t)
	flipByte(t, installation.binary)

	override := filepath.Join(t.TempDir(), "cohere-local")
	writeFixtureExecutable(t, override)

	installation.run(t, "--version", BinaryOverrideVariable+"="+override).requireRan(t, "--version")
}

// installedRelease is a dispatcher and this machine's platform package, staged where npm installs them.
type installedRelease struct {
	root          string
	home          string
	launcher      string
	binary        string
	checksums     string
	directoryName string
}

// stageInstalledRelease stages a fixture release into node_modules/@system-inc, which the staging layout
// already is: one directory per package, named without the scope.
func stageInstalledRelease(t *testing.T) installedRelease {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fixture binary is a shell script; the Windows package is proven by check-platform-package.sh on Windows")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Fatalf("node is not installed, and the launcher is a Node script, so it cannot be tested: %v", err)
	}

	// Through every link, as Node resolves the launcher and the package, so the paths its messages name
	// are the paths this test holds (macOS's temporary directory is under a link).
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "node_modules", PackageScope)
	target := Target{GoOperatingSystem: runtime.GOOS, GoArchitecture: runtime.GOARCH}
	stageFixturePlatform(t, output, target)

	checksums, err := Checksums(output, []string{target.DirectoryName()})
	if err != nil {
		t.Fatal(err)
	}
	module, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dispatcher, err := buildDispatcherPackage(Options{ModuleDirectory: module, OutputDirectory: output, Version: "1.0.0"}, checksums)
	if err != nil {
		t.Fatal(err)
	}

	return installedRelease{
		root:          root,
		home:          t.TempDir(),
		launcher:      filepath.Join(dispatcher.Directory, launcherRelativePath),
		binary:        filepath.Join(output, target.DirectoryName(), "bin", BinaryFileName(runtime.GOOS)),
		checksums:     filepath.Join(dispatcher.Directory, ChecksumsFileName),
		directoryName: target.DirectoryName(),
	}
}

// verdictPath is where the launcher remembers a passing check for this installation: the user cache
// under this test's home, keyed by the package directory as Node resolves it, through every link.
func (installation installedRelease) verdictPath(t *testing.T) string {
	t.Helper()
	packageDirectory, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(installation.binary)))
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(installation.home, ".cache")
	if runtime.GOOS == "darwin" {
		cache = filepath.Join(installation.home, "Library", "Caches")
	}
	return filepath.Join(cache, "cohere", "verified-binaries", sha256Hex([]byte(packageDirectory)))
}

// launcherRun is one run's outcome.
type launcherRun struct {
	exitCode       int
	standardOutput string
	standardError  string
}

// run runs the launcher with this installation's home as the user's, so its remembered verdicts are the
// test's own, and with no override unless one is given.
func (installation installedRelease) run(t *testing.T, argument string, environment ...string) launcherRun {
	t.Helper()
	command := exec.Command("node", installation.launcher, argument)
	command.Dir = installation.root
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if name == "HOME" || name == "XDG_CACHE_HOME" || name == BinaryOverrideVariable {
			continue
		}
		command.Env = append(command.Env, variable)
	}
	command.Env = append(command.Env, "HOME="+installation.home)
	command.Env = append(command.Env, environment...)

	var standardOutput, standardError bytes.Buffer
	command.Stdout = &standardOutput
	command.Stderr = &standardError
	err := command.Run()
	exitCode := 0
	if exitError, isExit := err.(*exec.ExitError); isExit {
		exitCode = exitError.ExitCode()
	} else if err != nil {
		t.Fatalf("running the launcher: %v", err)
	}
	return launcherRun{exitCode: exitCode, standardOutput: standardOutput.String(), standardError: standardError.String()}
}

func (result launcherRun) requireRan(t *testing.T, argument string) {
	t.Helper()
	if result.exitCode != 0 || !strings.Contains(result.standardOutput, "fixture ran: "+argument) {
		t.Fatalf("the launcher did not run the intact binary: exit %d\nstdout: %s\nstderr: %s", result.exitCode, result.standardOutput, result.standardError)
	}
}

// requireRefused holds a refusal to three things: a non-zero exit, a message naming the file and the
// reason, and nothing run, which the tampered fixture would have announced in either case.
func (result launcherRun) requireRefused(t *testing.T, file string, reason string) {
	t.Helper()
	if result.exitCode == 0 {
		t.Fatalf("the launcher exited zero over %s\nstdout: %s\nstderr: %s", file, result.standardOutput, result.standardError)
	}
	if strings.Contains(strings.ToLower(result.standardOutput), "fixture ran") {
		t.Fatalf("the launcher refused, but only after running the binary:\n%s", result.standardOutput)
	}
	if !strings.Contains(result.standardError, file) || !strings.Contains(result.standardError, reason) {
		t.Fatalf("the refusal does not name %s and say %q:\n%s", file, reason, result.standardError)
	}
}

// stageFixturePlatform writes a platform package whose bin/ holds the fixture cohere and a cohere-swift.
func stageFixturePlatform(t *testing.T, output string, target Target) {
	t.Helper()
	directory := filepath.Join(output, target.DirectoryName())
	manifest, err := PlatformManifest(target, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(directory, "package.json"), string(manifest))
	for _, name := range []string{BinaryFileName(target.GoOperatingSystem), SwiftEngineFileName} {
		writeFixtureExecutable(t, filepath.Join(directory, "bin", name))
	}
}

// writeFixtureExecutable writes the fixture script, executable whatever the umask.
func writeFixtureExecutable(t *testing.T, path string) {
	t.Helper()
	writeExecutable(t, path, 0o755)
	if err := os.WriteFile(path, []byte(fixtureScript), 0o755); err != nil {
		t.Fatal(err)
	}
}

// flipByte flips one bit of a file in place: the case of its first "f", which in the fixture turns
// "fixture" into "Fixture" and leaves a script that still runs.
func flipByte(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	index := bytes.IndexByte(contents, 'f')
	if index < 0 {
		index = len(contents) / 2
	}
	contents[index] ^= 0x20
	if err := os.WriteFile(path, contents, 0o755); err != nil {
		t.Fatal(err)
	}
}

func sha256Hex(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}
