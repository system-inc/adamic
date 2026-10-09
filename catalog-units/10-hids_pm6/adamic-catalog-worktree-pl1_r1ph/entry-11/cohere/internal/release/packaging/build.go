package release

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// BuildFlags are the flags every released binary is built with.
//
// They match `dispatch.ReleaseBuildFlags` deliberately: a locally rebuilt binary and a shipped one
// should differ only in their stamps, so that a bug reproduced against one reproduces against the
// other. `-trimpath` also removes the build machine's absolute paths from the binary, which a
// published artifact should not carry.
var BuildFlags = []string{"-trimpath"}

// StripFlags remove the symbol table and DWARF from a released binary.
//
// Named rather than inlined so that DescribeBuild reports the flags a release actually used. A
// hand-written description drifts from the build it claims to describe, and a size label that names
// the wrong build is worse than an unlabeled size: it is confidently wrong rather than ambiguous.
var StripFlags = []string{"-s", "-w"}

// DescribeBuild names the build a released binary comes from, for any size reported about it.
//
// A size without its build is not a measurement. The same commit measures 43.1 MB stripped and
// 61.9 MB from a plain `go build`, an 18 MB spread, so two people quoting sizes from different
// builds can both be right and still disagree — which is exactly what happened here, and cost two
// messages to reconcile. The same trap has a sharper form for anyone attributing those bytes:
// `go tool nm` reports every symbol as size zero on a stripped binary, so an attribution run
// against a release build sums to zero rather than failing.
func DescribeBuild() string {
	return strings.Join(BuildFlags, " ") + " -ldflags=\"" + strings.Join(StripFlags, " ") + " ...\""
}

// Options configure a release build.
type Options struct {
	// ModuleDirectory is the root of the cohere module.
	ModuleDirectory string

	// OutputDirectory is where the staged packages are written, each in a directory named for its
	// package without the scope: the dispatcher in "cohere" and each platform in one like
	// "cohere-darwin-arm64". A scoped name would nest the dispatcher in "@system-inc/cohere" and turn
	// the publish loop's one directory per package into a directory holding a package.
	OutputDirectory string

	// Version is the published version, like "0.3.1".
	Version string

	// Targets are the platforms to build. Empty means every target.
	Targets []Target

	// Signing configures macOS codesigning. A zero value stages unsigned binaries, which is a
	// supported outcome rather than a failure: signing needs credentials a contributor may not
	// have, and an npm install does not set the quarantine attribute that Gatekeeper checks. What
	// is not supported is an unsigned release that looks signed, so the state is always reported.
	Signing Signing

	// SwiftScratchDirectory is where SwiftPM builds the Swift engine. Empty means the user cache
	// directory's cohere/release-swift-build, so a second release on one machine builds incrementally.
	// Never inside OutputDirectory: the scratch is over a gigabyte, and everything in the output is
	// packed into the artifact every proof downloads.
	SwiftScratchDirectory string
}

// Result reports what a release build produced.
type Result struct {
	// Packages are the staged package directories, in the order they were built.
	Packages []StagedPackage
}

// StagedPackage is one built package on disk.
type StagedPackage struct {
	// Name is the npm package name.
	Name string

	// Directory is where it was staged.
	Directory string

	// BinaryPath is the executable inside it, empty for the dispatcher-only package.
	BinaryPath string

	// SizeInBytes is the binary's size, for the release summary. A binary that comes out
	// drastically smaller than its siblings is usually a build that failed into an empty file.
	SizeInBytes int64
}

// Build stages every platform package and the dispatcher package.
//
// It fails on the first target that does not build, rather than collecting errors and reporting at
// the end. A release missing one platform is not a release with a warning: it is a published
// version that resolves to nothing on that platform, and the dispatcher's loud failure would be
// reported as a bug against a version that looked fine everywhere else.
func Build(options Options) (Result, error) {
	if options.Version == "" {
		// A release with no version stamps "dev" into the binary, and a binary that says "dev"
		// cannot be traced back to what shipped. That is precisely the thing --version exists for.
		return Result{}, fmt.Errorf("a release needs a version")
	}

	if err := requireReleaseVersion(options.Version); err != nil {
		return Result{}, err
	}

	scratch, err := swiftScratchDirectory(options)
	if err != nil {
		return Result{}, err
	}
	options.SwiftScratchDirectory = scratch

	// Read before the minimum is checked, though neither depends on the other, because a fixture can
	// pin a compiler and cannot contain MinimumReleaseCommit. In the other order no test could reach
	// this refusal through Build, and a test that drives readCompilerPin directly passes with the call
	// deleted from here.
	pin, err := readCompilerPin(options.ModuleDirectory)
	if err != nil {
		return Result{}, err
	}

	if err := requireAncestor(options.ModuleDirectory, MinimumReleaseCommit, minimumReleaseReason); err != nil {
		return Result{}, err
	}

	targets := options.Targets
	if len(targets) == 0 {
		targets = Targets
	}

	goToolchain, err := readGoToolchain(options.ModuleDirectory)
	if err != nil {
		return Result{}, err
	}

	result := Result{}

	packageDirectoryNames := []string{}
	for _, target := range targets {
		staged, err := buildPlatformPackage(options, target, pin, goToolchain)
		if err != nil {
			return Result{}, fmt.Errorf("building %s: %w", target, err)
		}
		result.Packages = append(result.Packages, staged)
		packageDirectoryNames = append(packageDirectoryNames, target.DirectoryName())
	}

	// After every platform package is signed, so the sums describe the bytes that ship. One rendering is
	// written twice, into the dispatcher and beside the packages for the GitHub release, so the copy a
	// reader downloads and the copy the launcher checks cannot disagree.
	checksums, err := Checksums(options.OutputDirectory, packageDirectoryNames)
	if err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(options.OutputDirectory, ChecksumsFileName), checksums, 0o644); err != nil {
		return Result{}, fmt.Errorf("writing %s: %w", ChecksumsFileName, err)
	}

	dispatcher, err := buildDispatcherPackage(options, checksums)
	if err != nil {
		return Result{}, fmt.Errorf("building the dispatcher package: %w", err)
	}
	result.Packages = append(result.Packages, dispatcher)

	return result, nil
}

// releaseVersionPattern is a semantic version: three numbers with no leading zeros, and an
// optional pre-release such as `-rc.1`. Build metadata (`+...`) is refused, because npm ignores it
// when comparing versions, so two releases differing only there would collide.
var releaseVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)

// requireReleaseVersion refuses a version that is not semver, or whose major version is below 1.
//
// cohere's versions start at 1.0.0, by Kirk's ruling on #ed2dp27. The version is typed by hand into
// the release workflow, so the floor is enforced here rather than left to an example someone has to
// remember: a `0.x` publish would tell every consumer the API is unstable, and npm cannot take a
// published version back.
func requireReleaseVersion(version string) error {
	match := releaseVersionPattern.FindStringSubmatch(version)
	if match == nil {
		return fmt.Errorf("%q is not a semantic version like 1.0.0 or 1.0.0-rc.1, so it cannot be published", version)
	}
	if match[1] == "0" {
		return fmt.Errorf("%q is below 1.0.0, and cohere's versions start at 1.0.0", version)
	}
	return nil
}

// MinimumReleaseCommit is the oldest cohere commit a release may be built from.
//
// It is `1bd41fb1`, where formatting options started coming only from CohereSettings.json's format
// block. An engine from before it does not know that block, and in a repository whose package.json
// no longer has a prettier key it formats silently with Prettier's defaults rather than failing:
// one printed 2-space indentation over api's GraphQL. A repository that pins a published version
// gets that version's engine and nothing newer can intervene, so the only place to stop it is here,
// before it is published.
//
// A launcher check was considered and rejected. Platform packages are pinned to the launcher's
// exact version, so a normal install never pairs a new launcher with an old engine, and an old
// published version carries an old launcher that no check written later can reach.
//
// Written out in full because an abbreviated hash is only unambiguous for the history it was
// abbreviated against.
const MinimumReleaseCommit = "1bd41fb1bdb465f4453c472bcb1e17da5d2d6cf1"

const minimumReleaseReason = "older engines ignore CohereSettings.json's format block and format with Prettier's defaults"

// requireAncestor refuses a release built from a commit that does not contain minimum.
//
// git answers with three exit codes, and only one of them is a pass. 0 means minimum is in the
// history. 1 means it is not. Anything else means git could not decide, and the usual cause is a
// shallow clone where minimum was never fetched; that is refused too, because a check that cannot
// tell is not a check that passed.
func requireAncestor(moduleDirectory string, minimum string, reason string) error {
	command := exec.Command("git", "-C", moduleDirectory, "merge-base", "--is-ancestor", minimum, "HEAD")
	output, err := command.CombinedOutput()
	if err == nil {
		return nil
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) && exitError.ExitCode() == 1 {
		return fmt.Errorf(
			"this release would be built from a commit that does not contain %s, and %s. Release from a checkout that includes it",
			ShortCommit(minimum), reason,
		)
	}
	return fmt.Errorf(
		"could not tell whether this release contains %s, so it is refused rather than assumed: %w\n%s\nA shallow clone is the usual cause; fetch full history and release again",
		ShortCommit(minimum), err, strings.TrimSpace(string(output)),
	)
}

// buildPlatformPackage cross-compiles one target and writes its package around the binary.
func buildPlatformPackage(options Options, target Target, pin compilerPin, goToolchain string) (StagedPackage, error) {
	directory := filepath.Join(options.OutputDirectory, target.DirectoryName())
	binaryPath := filepath.Join(directory, "bin", target.BinaryFileName())

	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		return StagedPackage{}, fmt.Errorf("creating the package directory: %w", err)
	}

	if err := compile(options, target, binaryPath, pin, goToolchain); err != nil {
		return StagedPackage{}, err
	}

	size, err := cohereBinary(binaryPath, target)
	if err != nil {
		return StagedPackage{}, err
	}

	// The Swift engine, beside cohere in the same bin/, on the platforms that ship it. A released cohere
	// looks for it there and nowhere else, so a package missing it is a broken package, not a smaller
	// one: the build fails rather than staging without it.
	executables := []string{binaryPath}
	enginePath := filepath.Join(directory, "bin", SwiftEngineFileName)
	engineCommit := ""
	if ShipsSwiftEngine(target.GoOperatingSystem) {
		engineCommit, err = readReleaseCommit(options.ModuleDirectory)
		if err != nil {
			return StagedPackage{}, err
		}
		if err := buildSwiftEngine(options.ModuleDirectory, options.SwiftScratchDirectory, target, enginePath, engineCommit); err != nil {
			return StagedPackage{}, err
		}
		executables = append(executables, enginePath)
	}

	// Signed before the manifest is written, so a signing failure leaves an obviously incomplete
	// package rather than one that looks finished. Only macOS is signed: Linux and Windows have no
	// equivalent gate on an npm-installed binary.
	if target.IsMacOS() && options.Signing.IsConfigured() {
		for _, executable := range executables {
			if err := Sign(executable, options.Signing); err != nil {
				return StagedPackage{}, err
			}
		}
	}

	// After signing, so the pair checked is the pair that ships.
	if ShipsSwiftEngine(target.GoOperatingSystem) {
		if err := requireSwiftContract(binaryPath, enginePath, engineCommit); err != nil {
			return StagedPackage{}, err
		}
	}

	manifest, err := PlatformManifest(target, options.Version)
	if err != nil {
		return StagedPackage{}, err
	}
	if err := os.WriteFile(filepath.Join(directory, "package.json"), manifest, 0o644); err != nil {
		return StagedPackage{}, fmt.Errorf("writing the package manifest: %w", err)
	}

	// Beside bin/, not in it, so the checksums stay a list of executables.
	if err := stageLicenses(options.ModuleDirectory, directory); err != nil {
		return StagedPackage{}, err
	}

	return StagedPackage{
		Name:        target.PackageName(),
		Directory:   directory,
		BinaryPath:  binaryPath,
		SizeInBytes: size,
	}, nil
}

// buildDispatcherPackage writes the package consumers actually install.
//
// Nothing is compiled here, which is the point. The dispatcher is one package every machine
// installs regardless of platform, so it cannot be a Go binary without becoming a seventh platform
// package and defeating the single-install premise. It is a Node script instead — Node is present
// by construction in an npm install — and it resolves the platform package through the package
// manager rather than by guessing at directory layouts.
//
// It takes no compiler pin and no toolchain for that reason: there is nothing here to stamp them
// into, and threading them in would imply this package carries provenance that it does not. What it
// does carry is the platform binaries' checksums, which the launcher checks before it runs one. The
// dispatcher pins every platform package exactly, so the sums for this version describe the only
// binaries it will ever be installed beside.
func buildDispatcherPackage(options Options, checksums []byte) (StagedPackage, error) {
	if len(checksums) == 0 {
		// The launcher refuses to run any binary its checksums do not list, so this package would
		// install and then run nothing anywhere.
		return StagedPackage{}, fmt.Errorf("the dispatcher package needs the platform binaries' checksums")
	}

	directory := filepath.Join(options.OutputDirectory, UnscopedDispatcherPackageName)
	if err := os.MkdirAll(filepath.Join(directory, "bin"), 0o755); err != nil {
		return StagedPackage{}, fmt.Errorf("creating the dispatcher package directory: %w", err)
	}

	manifest, err := DispatcherManifest(options.Version)
	if err != nil {
		return StagedPackage{}, err
	}
	if err := os.WriteFile(filepath.Join(directory, "package.json"), manifest, 0o644); err != nil {
		return StagedPackage{}, fmt.Errorf("writing the dispatcher manifest: %w", err)
	}

	launcher := DispatcherLauncher()
	launcherPath := filepath.Join(directory, "bin", FullCommandName)
	if err := os.WriteFile(launcherPath, []byte(launcher), 0o755); err != nil {
		return StagedPackage{}, fmt.Errorf("writing the dispatcher launcher: %w", err)
	}
	// WriteFile's mode is masked by umask, so the executable bit is set explicitly. A launcher that
	// ships without it fails at exec with a permission error on every machine that installs it.
	if err := os.Chmod(launcherPath, 0o755); err != nil {
		return StagedPackage{}, fmt.Errorf("making the dispatcher launcher executable: %w", err)
	}

	if err := stageSettingsSchemas(options.ModuleDirectory, directory); err != nil {
		return StagedPackage{}, err
	}

	if err := stageLicenses(options.ModuleDirectory, directory); err != nil {
		return StagedPackage{}, err
	}

	if err := os.WriteFile(filepath.Join(directory, ChecksumsFileName), checksums, 0o644); err != nil {
		return StagedPackage{}, fmt.Errorf("writing the dispatcher's %s: %w", ChecksumsFileName, err)
	}

	return StagedPackage{Name: DispatcherPackageName, Directory: directory}, nil
}

// swiftScratchDirectory resolves where SwiftPM builds the engine, and refuses a directory inside the
// output, which the release packs and uploads whole: 1.3 GB of scratch beside 200 MB of packages, measured
// on a full staging when the scratch was `<output>/.swift-build`.
func swiftScratchDirectory(options Options) (string, error) {
	scratch := options.SwiftScratchDirectory
	if scratch == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("finding a cache directory for the Swift engine's build: %w", err)
		}
		scratch = filepath.Join(cache, "cohere", "release-swift-build")
	}
	scratch, err := filepath.Abs(scratch)
	if err != nil {
		return "", err
	}
	output, err := filepath.Abs(options.OutputDirectory)
	if err != nil {
		return "", err
	}
	if relative, err := filepath.Rel(output, scratch); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("the Swift engine's build directory %s is inside the output %s, which is packed and uploaded whole", scratch, output)
	}
	return scratch, nil
}

// SchemaDirectoryName is where the settings schemas sit, in the module and in the dispatcher package
// alike, so a project names them as ./node_modules/@system-inc/cohere/schema/<file>.
const SchemaDirectoryName = "schema"

// SettingsSchemaFileNames are the JSON schemas the dispatcher package ships, one per settings tier.
// They are generated from the loader (internal/settingsschema) and committed, and a test there fails when
// they drift from it, so the release copies the committed files rather than generating its own.
var SettingsSchemaFileNames = []string{"CohereSettings.schema.json", "NexusCohereSettings.schema.json"}

// stageSettingsSchemas copies the settings schemas into the dispatcher package.
//
// The dispatcher and not a platform package, because the schema is the same on every platform and
// every install has exactly one dispatcher. A schema that is missing, empty or not JSON stops the
// release: a "$schema" path that resolves to nothing validates nothing, and an editor shows no error for
// that, so it would read as a settings file with no mistakes in it.
func stageSettingsSchemas(moduleDirectory string, packageDirectory string) error {
	if err := os.MkdirAll(filepath.Join(packageDirectory, SchemaDirectoryName), 0o755); err != nil {
		return fmt.Errorf("creating the dispatcher's schema directory: %w", err)
	}
	for _, name := range SettingsSchemaFileNames {
		source := filepath.Join(moduleDirectory, SchemaDirectoryName, name)
		contents, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("reading the settings schema the dispatcher ships: %w", err)
		}
		var schema map[string]any
		if err := json.Unmarshal(contents, &schema); err != nil || len(schema) == 0 {
			return fmt.Errorf("%s is not a JSON schema (%v), so the dispatcher would ship a $schema target that validates nothing", source, err)
		}
		if err := os.WriteFile(filepath.Join(packageDirectory, SchemaDirectoryName, name), contents, 0o644); err != nil {
			return fmt.Errorf("staging the settings schema %s: %w", name, err)
		}
	}
	return nil
}

// compile cross-compiles one target, stamping the provenance in.
func compile(options Options, target Target, binaryPath string, pin compilerPin, goToolchain string) error {
	const packagePath = "github.com/system-inc/cohere/internal/release/packaging"

	// Strip the symbol table and DWARF. Measured on a comparable binary: marginally faster to link
	// and 29% smaller, with nothing traded away that a released binary needs.
	stamps := append([]string{}, StripFlags...)
	stamps = append(stamps,
		"-X", packagePath+".version="+options.Version,
		"-X", packagePath+".compilerCommit="+pin.Commit,
		"-X", packagePath+".compilerUpstream="+pin.Upstream,
		"-X", packagePath+".goToolchain="+goToolchain,
	)

	command := GoBuildCommand(options.ModuleDirectory, target, strings.Join(stamps, " "), binaryPath)
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}
	return nil
}

// GoBuildCommand is the `go build` that produces cohere for one target.
//
// It is the single definition of how a target is compiled, shared by the release and by
// TestEveryReleaseTargetCompiles (internal/release/crosscompile), so a flag or an environment variable added here for the release
// is one the test compiles with too. A test with its own copy would keep passing on the copy after
// the release had changed.
func GoBuildCommand(moduleDirectory string, target Target, linkerFlags string, binaryPath string) *exec.Cmd {
	arguments := append([]string{"build"}, BuildFlags...)
	arguments = append(arguments, "-ldflags="+linkerFlags, "-o", binaryPath, "./command/cohere")

	command := exec.Command("go", arguments...)
	command.Dir = moduleDirectory
	command.Env = append(os.Environ(),
		"GOOS="+target.GoOperatingSystem,
		"GOARCH="+target.GoArchitecture,
		// Nothing here needs cgo, and disabling it is what makes the Linux binaries statically
		// linked. A dynamically linked binary would depend on the glibc of the build machine and
		// fail on any container older than it, which is a failure that only shows up in someone
		// else's CI.
		"CGO_ENABLED=0",
	)
	return command
}

// cohereBinary confirms the build produced an executable for the platform it claims, returning its
// size.
//
// Three questions, because the first two answer something narrower than the name suggests. Go
// reports success by exit code, but the artifact is the file. A file that exists says nothing about
// whether it holds a program. And a program says nothing about which machine it runs on.
//
// The third check is the one with a reachable failure behind it: a staging bug that wrote one
// target's binary into another's package would pass existence and size perfectly, publish cleanly,
// install cleanly, and fail at exec with a format error on a user's machine, reported as a broken
// install rather than as the packaging mistake it is. Magic bytes separate all three families we
// ship, so the check costs four bytes of read.
func cohereBinary(path string, target Target) (int64, error) {
	information, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("go build reported success but produced no binary at %s: %w", path, err)
	}

	// A real cohere binary statically links the whole TypeScript compiler and is over ten megabytes.
	// Anything near zero is a build that failed into an empty file, which packages perfectly and
	// installs perfectly and does nothing.
	const implausiblySmall = 1 << 20
	if information.Size() < implausiblySmall {
		return 0, fmt.Errorf("the binary at %s is %d bytes, which is far too small to be a real build", path, information.Size())
	}

	if err := requireExecutableFormat(path, target); err != nil {
		return 0, err
	}

	return information.Size(), nil
}

// executableMagic is the leading byte sequence of each executable format we ship.
//
// Keyed by GOOS because that is what decides the format: both darwin targets are Mach-O and both
// windows targets are PE, so the architecture is carried inside the file rather than in its first
// bytes. That means this catches a binary built for the wrong operating system and not one built
// for the wrong architecture of the right system, which is a real limit and is stated in the test
// rather than implied away here.
var executableMagic = map[string][]byte{
	// Mach-O 64-bit, little-endian. Go emits this for both darwin targets.
	"darwin": {0xcf, 0xfa, 0xed, 0xfe},
	// ELF.
	"linux": {0x7f, 'E', 'L', 'F'},
	// PE, which still begins with the DOS stub's "MZ".
	"windows": {'M', 'Z'},
}

// requireExecutableFormat reports whether a file is an executable of the format its target expects.
//
// An unknown operating system passes rather than failing, because this is a guard against a
// mis-staged binary and not a gate on which platforms may exist. A new target added to `Targets`
// without a magic entry should not block a release; `TestEveryTargetHasAKnownExecutableFormat`
// catches the omission at test time, which is where it belongs.
func requireExecutableFormat(path string, target Target) error {
	magic, known := executableMagic[target.GoOperatingSystem]
	if !known {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("reading the binary at %s: %w", path, err)
	}
	defer file.Close()

	header := make([]byte, len(magic))
	if _, err := io.ReadFull(file, header); err != nil {
		return fmt.Errorf("reading the header of %s: %w", path, err)
	}

	if !bytes.Equal(header, magic) {
		// The observation and the likely cause are stated as separate claims, because they are: the
		// bytes are measured and the cause is a guess. A message that asserts one explanation sends
		// a reader hunting that bug, and several things produce these bytes — another target's
		// build staged here, a truncated write that kept the path, a file that was never a program.
		// Naming the most likely one while marking it a guess costs nothing and misdirects nobody.
		return fmt.Errorf(
			"the binary at %s begins with %x, and a %s executable begins with %x.\nMost often this is another target's build staged into this package, but a truncated write or a non-executable file at that path produce it too",
			path, header, target.GoOperatingSystem, magic,
		)
	}
	return nil
}

// compilerPin is which commit of which repository the vendored compiler is pinned to.
type compilerPin struct {
	// Commit is the gitlink HEAD records for the submodule, which the submodule's checkout was
	// required to match.
	Commit string

	// Upstream is the repository that commit lives in, as "owner/name".
	Upstream string
}

// readCompilerPin reads the pinned commit of the vendored compiler and the repository it came from,
// both from the commit being released rather than from the checkout around it.
//
// Both halves are read rather than written down here, because the upstream has already moved: the
// compiler was vendored from `microsoft/typescript-go` until that repository was archived, then from
// `microsoft/TypeScript`, then from the fork `kirkouimet/TypeScript`, and is now System, Inc.'s fork
// `system-inc/TypeScript`. A hardcoded label survives a migration like that while quietly becoming
// false, and a commit reported against the wrong repository is worse than no commit at all: it resolves
// to nothing and gives a reader no hint why.
//
// The commit is the gitlink, and the submodule's checkout has to agree with it. The build compiles
// whatever the checkout holds, so a checkout moved without a commit would ship a compiler that no
// cohere commit names, and the release could not be rebuilt from history. `--version` would name it
// accurately and still point at nothing. Commits are compared and trees are not: the compiler patches
// change files in the checkout without moving its HEAD, so a tree comparison would refuse every
// patched release, which is every correct one.
func readCompilerPin(moduleDirectory string) (compilerPin, error) {
	recorded, err := PinnedCompilerCommit(moduleDirectory, "HEAD")
	if err != nil {
		return compilerPin{}, err
	}

	checkedOut, err := checkedOutCompilerCommit(moduleDirectory)
	if err != nil {
		return compilerPin{}, err
	}

	if checkedOut != recorded {
		return compilerPin{}, fmt.Errorf(
			"the compiler at TypeScript is checked out at %s, and the commit being released pins %s.\nA release built this way ships a compiler no cohere commit names. Commit the new pin, or run `git submodule update TypeScript` to return to the recorded one",
			checkedOut, recorded,
		)
	}

	return compilerPin{Commit: recorded, Upstream: readCompilerUpstream(moduleDirectory)}, nil
}

// PinnedCompilerCommit reads the vendored compiler's commit from a commit's tree. Exported because the
// launcher builds from a commit too, and two readers of one gitlink would drift.
//
// The gitlink recorded in the commit is the pin. The submodule's own HEAD is what the working tree
// has checked out, and it differs exactly when someone has moved it without committing.
func PinnedCompilerCommit(moduleDirectory string, commit string) (string, error) {
	command := exec.Command("git", "-C", moduleDirectory, "ls-tree", commit, "TypeScript")
	var standardError bytes.Buffer
	command.Stderr = &standardError
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("reading the compiler pin from %s: %w: %s", ShortCommit(commit), err, strings.TrimSpace(standardError.String()))
	}
	// "160000 commit <sha>\tTypeScript" for a submodule. Anything else means the commit does not pin
	// a compiler at all, and building would pick one up from somewhere unnamed.
	fields := strings.Fields(string(output))
	if len(fields) < 3 || fields[0] != "160000" || fields[1] != "commit" {
		return "", fmt.Errorf("commit %s does not pin the vendored compiler as a submodule at TypeScript (ls-tree printed %q)",
			ShortCommit(commit), strings.TrimSpace(string(output)))
	}
	return fields[2], nil
}

// checkedOutCompilerCommit reads the commit the compiler's checkout is at.
//
// The checkout has to be a repository of its own first. A submodule that was never initialized is an
// empty directory, and `git -C` from an empty directory walks up and answers with cohere's own HEAD,
// which would be reported as a moved compiler naming a commit that is not a compiler at all.
func checkedOutCompilerCommit(moduleDirectory string) (string, error) {
	submoduleDirectory := filepath.Join(moduleDirectory, "TypeScript")

	notCheckedOut := fmt.Errorf("the compiler at %s is not checked out. Run `git submodule update --init TypeScript`", submoduleDirectory)
	output, err := exec.Command("git", "-C", submoduleDirectory, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", notCheckedOut
	}
	topLevel, err := filepath.EvalSymlinks(strings.TrimSpace(string(output)))
	if err != nil {
		return "", notCheckedOut
	}
	// Absolute before resolving, because git answers with an absolute path and EvalSymlinks keeps a
	// relative one relative. Without it a module named as "." reads as never checked out.
	expected, err := filepath.Abs(submoduleDirectory)
	if err == nil {
		expected, err = filepath.EvalSymlinks(expected)
	}
	if err != nil || topLevel != expected {
		return "", notCheckedOut
	}

	output, err = exec.Command("git", "-C", submoduleDirectory, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("reading the checked-out compiler commit: %w", err)
	}
	commit := strings.TrimSpace(string(output))
	if commit == "" {
		return "", fmt.Errorf("the checked-out compiler commit came back empty")
	}
	return commit, nil
}

// readCompilerUpstream names the compiler's repository from the `.gitmodules` HEAD commits.
//
// The committed file and not the submodule's `origin` remote, which is local configuration no commit
// records. The pin moved to the fork in a commit, and a checkout cloned before that still has an
// origin pointing at Microsoft until someone runs `git submodule sync` by hand, so reading the remote
// stamps a repository beside a commit that repository does not contain.
//
// A file that does not say degrades to "unknown" rather than failing the release. The commit is the
// fact a bug report needs most, and it was already read from the commit and checked above.
func readCompilerUpstream(moduleDirectory string) string {
	output, err := exec.Command("git", "-C", moduleDirectory, "show", "HEAD:.gitmodules").Output()
	if err != nil {
		return "unknown"
	}
	return CompilerUpstreamFromGitmodules(string(output))
}

// CompilerUpstreamFromGitmodules names the compiler's repository, as "owner/name", from the contents of
// a `.gitmodules`, or "unknown" when it does not say. Exported because the launcher reads the same file
// from a snapshot, and two parsers of one file would drift.
//
// Sections are `[submodule "<name>"]`, each with `path` and `url`. The one whose path is the compiler's
// directory is the compiler, whatever it is named.
func CompilerUpstreamFromGitmodules(contents string) string {
	path, url := "", ""
	found := ""
	flush := func() {
		if path == "TypeScript" && url != "" {
			found = url
		}
		path, url = "", ""
	}
	for line := range strings.SplitSeq(contents, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			flush()
			continue
		}
		key, value, isSetting := strings.Cut(line, "=")
		if !isSetting {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"`)
		switch strings.TrimSpace(key) {
		case "path":
			path = strings.TrimSuffix(value, "/")
		case "url":
			url = value
		}
	}
	flush()
	return NormalizeUpstream(found)
}

// NormalizeUpstream reduces a git remote url to "owner/name". Exported because the launcher names the
// compiler's repository from the same kind of url, read from a committed `.gitmodules`, and two parsers
// of one url would drift.
//
// The two url shapes git accepts — the ssh `git@github.com:owner/name.git` and the https
// `https://github.com/owner/name.git` — have to produce the same label, because otherwise the same
// pin reads differently depending on how the build machine happened to clone, and a reader
// comparing two reports would see a difference that is not one.
//
// It is separate from the command that reads the remote so it can be tested without a git
// repository: the parsing is where the bugs are, and it should not need a fixture clone to exercise.
func NormalizeUpstream(remoteUrl string) string {
	url := strings.TrimSpace(remoteUrl)
	if url == "" {
		return "unknown"
	}

	url = strings.TrimSuffix(url, ".git")
	if _, path, found := strings.Cut(url, ":"); found && !strings.HasPrefix(url, "http") {
		url = path
	}

	segments := strings.Split(strings.Trim(url, "/"), "/")
	if len(segments) < 2 {
		return "unknown"
	}

	// Both halves have to be non-empty, not just present. A count alone passes for
	// `https://github.com/`, whose segments are ["https:", "", "github.com", ""], and the last two
	// join to "/github.com" — a label that is not a repository, printed by `--version` as though it
	// were one. A guard on position says the pieces exist; this says they mean something.
	owner, name := segments[len(segments)-2], segments[len(segments)-1]
	if owner == "" || name == "" {
		return "unknown"
	}
	return owner + "/" + name
}

// readGoToolchain reads the version of the toolchain performing the build.
func readGoToolchain(moduleDirectory string) (string, error) {
	command := exec.Command("go", "env", "GOVERSION")
	command.Dir = moduleDirectory

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("reading the Go toolchain version: %w", err)
	}

	toolchain := strings.TrimSpace(string(output))
	if toolchain == "" {
		return "", fmt.Errorf("the Go toolchain version came back empty")
	}
	return toolchain, nil
}
