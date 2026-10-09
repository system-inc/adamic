package dispatch

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/release/packaging"
)

// The launcher builds the cohere every project's gate runs from the checkout's committed tree, never
// from its working tree.
//
// The checkout is shared. Several members edit it at once, and until this existed the binary was
// named for a hash of whatever `go list` found on disk, so any member's half-written rule went live in
// every project's gate on the next run. On 2026-10-02 one member's uncommitted reading of nested
// literals turned 22 Tailwind findings on in ahra before anyone had classified the work. The work was
// right; the gate was still running code nobody had committed, which is the honesty doctrine failing
// inside the toolchain that enforces it.
//
// So a gate run builds from a snapshot of HEAD: this repository extracted with `git archive`, and the
// vendored compiler extracted at the commit HEAD pins it to. Nothing in either comes from a working
// tree, the compiler submodule's included. Testing uncommitted work is still possible, and it is
// explicit: `--dev` builds the working tree and says so on every run.

// snapshotTag is mixed into the binary hash so a committed-tree build can never share a name with a
// binary built any other way. Both kinds land in the same directory.
const snapshotTag = "committed-tree"

// compilerSubdirectory is the part of the vendored compiler the build reads. `go.work` uses only
// `./TypeScript/tsc`, so the rest of the submodule (another few hundred files of tooling) is left out of
// the snapshot.
const compilerSubdirectory = "tsc"

// committedBuild is what one committed-tree build is made from.
type committedBuild struct {
	// Commit is the commit of this repository being built.
	Commit string

	// CompilerCommit is the vendored compiler commit that Commit pins, read from Commit's own tree.
	// The submodule's checkout is not consulted, because a member bumping the pin locally is exactly
	// an uncommitted change.
	CompilerCommit string

	// GoVersion identifies the toolchain, as Inputs.GoVersion does.
	GoVersion string
}

// hash names the binary this build produces.
//
// The commit decides everything committed: the rules, the shims and the pinned compiler commit are
// all in its tree. What it does not decide is the toolchain and the flags, so those
// are framed in beside it.
func (build committedBuild) hash() string {
	digest := sha256.New()
	fmt.Fprintf(digest, "kind\x00%s\x00", snapshotTag)
	fmt.Fprintf(digest, "commit\x00%s\x00", build.Commit)
	fmt.Fprintf(digest, "goVersion\x00%s\x00", build.GoVersion)
	for _, flag := range ReleaseBuildFlags {
		fmt.Fprintf(digest, "buildFlag\x00%s\x00", flag)
	}
	return hex.EncodeToString(digest.Sum(nil))[:hashLength]
}

// SnapshotDirectory holds the per-build extractions of this repository. Each is removed once its
// binary is built, because the binary is self-contained and the next commit needs its own.
func (paths Paths) SnapshotDirectory() string {
	return filepath.Join(paths.CacheDirectory, "snapshot")
}

// CompilerDirectory holds compiler extractions, one per pinned commit.
//
// These are kept, because the compiler is 66,000 files and takes about seven seconds to extract,
// while the pin changes rarely. Two commits of this repository that pin the same compiler build
// against the same directory.
func (paths Paths) CompilerDirectory() string {
	return filepath.Join(paths.CacheDirectory, "compiler")
}

// ResolveCommitted returns a cohere binary built from the checkout's HEAD, building it when the
// cache has none.
//
// The returned string is the commit the binary was built from, so the caller can say which one ran,
// and the boolean reports whether a build happened.
func ResolveCommitted(paths Paths, packagePath string) (binaryPath string, commit string, built bool, err error) {
	build, err := readCommittedBuild(paths.ModuleDirectory)
	if err != nil {
		return "", "", false, err
	}

	binaryPath = paths.BinaryPath(build.hash())
	if _, statErr := os.Stat(binaryPath); statErr == nil {
		return binaryPath, build.Commit, false, nil
	}

	// Asked only for a build. The commit already decides the binary's name, the pin included, so a run
	// that finds its binary has no use for it, and reading a tree is the one question that needs git.
	build.CompilerCommit, err = release.PinnedCompilerCommit(paths.ModuleDirectory, build.Commit)
	if err != nil {
		return "", "", false, err
	}
	if err := buildCommitted(paths, packagePath, build, binaryPath); err != nil {
		return "", "", false, err
	}
	return binaryPath, build.Commit, true, nil
}

// readCommittedBuild reads HEAD and the toolchain, which is everything a binary's name depends on. The
// compiler commit HEAD pins is read only when a build needs it; see ResolveCommitted.
func readCommittedBuild(moduleDirectory string) (committedBuild, error) {
	commit, err := readHead(moduleDirectory)
	if err != nil {
		return committedBuild{}, fmt.Errorf("reading the commit to build cohere from: %w", err)
	}

	goVersion, err := goEnvironment(moduleDirectory)
	if err != nil {
		return committedBuild{}, err
	}

	return committedBuild{Commit: commit, GoVersion: goVersion}, nil
}

// buildCommitted extracts the snapshot, builds it, and proves the result before it can be found.
func buildCommitted(paths Paths, packagePath string, build committedBuild, binaryPath string) error {
	for _, directory := range []string{paths.BinaryDirectory(), paths.SnapshotDirectory(), paths.CompilerDirectory()} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", directory, err)
		}
	}

	Report.Step("cohere: building from commit %s, its committed tree only", release.ShortCommit(build.Commit))

	// A directory of its own, so two runs building at once never extract over each other.
	created, err := os.MkdirTemp(paths.SnapshotDirectory(), release.ShortCommit(build.Commit)+"-*")
	if err != nil {
		return fmt.Errorf("creating a snapshot directory: %w", err)
	}
	defer os.RemoveAll(created)

	// Every path below is the resolved one. Go resolves the build's working directory through any
	// symlink and compares it with the workspace file's location, so a snapshot named through a link
	// (macOS's `/var` is one, to `/private/var`) is reported as outside its own workspace and fails to
	// build. Found by the test fixture, which lives under `/var`.
	snapshot, err := filepath.EvalSymlinks(created)
	if err != nil {
		return fmt.Errorf("resolving the snapshot directory: %w", err)
	}

	if err := extractGitArchive(paths.ModuleDirectory, build.Commit, "", snapshot); err != nil {
		return fmt.Errorf("extracting commit %s: %w", release.ShortCommit(build.Commit), err)
	}

	compiler, err := ensureCompiler(paths, build.CompilerCommit)
	if err != nil {
		return err
	}

	// `git archive` writes the submodule as an empty directory. The compiler goes in its place
	// as a link, which Go follows for a workspace module, and which keeps the snapshot to this
	// repository's own few thousand files.
	submodulePath := filepath.Join(snapshot, "TypeScript")
	if err := os.Remove(submodulePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("clearing the submodule placeholder in the snapshot: %w", err)
	}
	if err := os.Symlink(compiler, submodulePath); err != nil {
		return fmt.Errorf("linking the compiler into the snapshot: %w", err)
	}

	// Built to a temporary name and renamed only once proven, so a concurrent run never finds, and
	// execs, a binary that is half-written or that failed its proof.
	// The suffix stays last, so Windows runs the proof against the temporary name as it runs the binary.
	temporary := strings.TrimSuffix(binaryPath, executableSuffix) + fmt.Sprintf(".partial-%d", os.Getpid()) + executableSuffix
	defer os.Remove(temporary)

	if err := goBuildSnapshot(paths, snapshot, packagePath, build, temporary); err != nil {
		return err
	}
	if err := proveCommittedBinary(temporary, snapshot, build); err != nil {
		return err
	}
	if err := os.Rename(temporary, binaryPath); err != nil {
		return fmt.Errorf("moving the built binary into the cache: %w", err)
	}

	pruneAfterBuild(paths, binaryPath, compiler)
	return nil
}

// pruneAfterBuild trims the binary cache once a new binary is in it, which is when the cache grows.
//
// It never fails the build: the binary is built and proven, and a prune that could not finish leaves
// the cache as large as it was, which is the state it started in. It says so instead. Every outcome
// goes to the prune log, a prune that removed nothing included, and what was removed is a note for
// --verbose, so an ordinary build stays quiet.
func pruneAfterBuild(paths Paths, keep string, currentCompiler string) {
	now := time.Now()
	plan, err := PlanPrune(paths, []string{keep}, currentCompiler, now)
	if err != nil {
		Report.Fail("cohere: the cache was not pruned: %v", err)
		return
	}
	removed, applyErr := ApplyPrune(paths, plan)
	if err := RecordPrune(paths, plan, removed, now); err != nil {
		Report.Fail("cohere: the prune log was not written: %v", err)
	}
	if len(removed) > 0 {
		bytes := int64(0)
		for _, file := range removed {
			bytes += file.Bytes
		}
		Report.Note("cohere: pruned %d cached binaries and extractions (%.1f GB) unused for over an hour; names in %s",
			len(removed), float64(bytes)/1e9, paths.PruneLogPath())
	}
	if applyErr != nil {
		Report.Fail("cohere: the prune stopped part way: %v", applyErr)
	}
}

// ensureCompiler returns an extraction of the compiler at compilerCommit, creating it once.
//
// It is keyed by the full commit, so a moved pin can never reuse an extraction of the previous one.
func ensureCompiler(paths Paths, compilerCommit string) (string, error) {
	key := compilerCommit
	compiler := filepath.Join(paths.CompilerDirectory(), key)

	if _, err := os.Stat(compiler); err == nil {
		// Marked as used, because the prune removes an extraction untouched for an hour and an
		// extraction's own time is when it was made. A build against a pin extracted yesterday would
		// otherwise compile against a directory the next prune is free to take.
		if err := os.Chtimes(compiler, time.Now(), time.Now()); err != nil {
			return "", fmt.Errorf("marking the compiler at %s in use: %w", release.ShortCommit(compilerCommit), err)
		}
		return compiler, nil
	}

	Report.Step("cohere: extracting the compiler at %s, once per pin (about seven seconds)", release.ShortCommit(compilerCommit))

	partial, err := os.MkdirTemp(paths.CompilerDirectory(), key+".partial-*")
	if err != nil {
		return "", fmt.Errorf("creating a compiler directory: %w", err)
	}
	// After a successful rename the name no longer exists and this does nothing.
	defer os.RemoveAll(partial)

	submodule := filepath.Join(paths.ModuleDirectory, "TypeScript")
	if err := extractGitArchive(submodule, compilerCommit, compilerSubdirectory, partial); err != nil {
		return "", fmt.Errorf("extracting the compiler at %s (if the commit is missing, the submodule has not fetched it: "+
			"run `git submodule sync && git submodule update --init` in %s): %w", release.ShortCommit(compilerCommit), paths.ModuleDirectory, err)
	}

	if err := os.Rename(partial, compiler); err != nil {
		// Another run finished the same extraction first. Its directory is identical by construction,
		// so it is used and this one is discarded.
		if _, statErr := os.Stat(compiler); statErr == nil {
			return compiler, nil
		}
		return "", fmt.Errorf("moving the compiler into the cache: %w", err)
	}
	return compiler, nil
}

// goBuildSnapshot compiles the snapshot.
//
// VCS stamping is off, and the commits are stamped explicitly instead. The snapshot has no `.git` of
// its own and sits inside this checkout, so with stamping on Go walks up to the checkout and records
// the working tree's revision and its dirty state, which is the state this build exists to exclude.
//
// GOWORK is set to the snapshot's own workspace file. An inherited GOWORK pointing at the checkout
// would build the checkout's modules from inside the snapshot without any error.
func goBuildSnapshot(paths Paths, snapshot string, packagePath string, build committedBuild, outputPath string) error {
	const packagingPath = "github.com/system-inc/cohere/internal/release/packaging"

	linkerFlags := []string{}
	arguments := []string{"build", "-buildvcs=false"}
	for _, flag := range ReleaseBuildFlags {
		// The release flags carry their own `-ldflags`, and Go keeps only the last one given, so the
		// stamps are joined onto it rather than passed as a second flag that would erase the first.
		if value, isLinkerFlags := strings.CutPrefix(flag, "-ldflags="); isLinkerFlags {
			linkerFlags = append(linkerFlags, value)
			continue
		}
		arguments = append(arguments, flag)
	}
	workspace := "GOWORK=" + filepath.Join(snapshot, "go.work")
	formatter, err := FormatterIdentity(snapshot, build.GoVersion, []string{workspace})
	if err != nil {
		return err
	}
	linkerFlags = append(linkerFlags,
		"-X", packagingPath+".selfCommit="+build.Commit,
		"-X", packagingPath+".compilerCommit="+build.CompilerCommit,
		"-X", packagingPath+".formatterIdentity="+formatter,
	)
	// The repository the pin lives in, so `--version` names it beside the commit. Since ffaddb2 the pin is on
	// a fork, and a bare commit sends a reader to microsoft/TypeScript, where it does not exist.
	if upstream := committedCompilerUpstream(snapshot); upstream != "" {
		linkerFlags = append(linkerFlags, "-X", packagingPath+".compilerUpstream="+upstream)
	}
	arguments = append(arguments, "-ldflags="+strings.Join(linkerFlags, " "), "-o", outputPath, packagePath)

	command := exec.Command("go", arguments...)
	command.Dir = snapshot
	// The shared Go cache, as build() uses: see ReleaseBuildFlags.
	command.Env = append(os.Environ(), workspace, "CGO_ENABLED=0")
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("building cohere from commit %s: %w", release.ShortCommit(build.Commit), err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("go build reported success but produced no binary at %s: %w", outputPath, err)
	}
	return nil
}

// proveCommittedBinary asks the built binary what it is, and refuses it unless the answer matches the
// build.
//
// Go's exit code says the compile succeeded. It does not say the stamps took. The binary's own
// `--version` does, so the proof is read off the artifact rather than off the steps that made it.
func proveCommittedBinary(binaryPath string, workingDirectory string, build committedBuild) error {
	command := exec.Command(binaryPath, "--version")
	command.Dir = workingDirectory
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("the binary built from %s failed its own --version (%w), so it was not cached:\n%s",
			release.ShortCommit(build.Commit), err, output)
	}
	return checkCommittedVersion(string(output), build.Commit)
}

// checkCommittedVersion is the judgement proveCommittedBinary makes, separate so it can be tested
// against the text a real binary prints without building one.
func checkCommittedVersion(version string, commit string) error {
	commitNamed := false
	for line := range strings.SplitSeq(version, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "commit:" && fields[1] == release.ShortCommit(commit) {
			commitNamed = true
		}
	}

	if !commitNamed {
		return fmt.Errorf("the binary built from %s does not name that commit in --version, so it cannot be traced back to it "+
			"and was not cached:\n%s", release.ShortCommit(commit), version)
	}
	if strings.Contains(version, "uncommitted changes") {
		return fmt.Errorf("the binary built from %s reports uncommitted changes, which a committed-tree build cannot have, "+
			"so it was not cached:\n%s", release.ShortCommit(commit), version)
	}
	return nil
}

// extractGitArchive writes a commit's tree, or one subdirectory of it, into an empty directory.
//
// `git archive` reads objects rather than files, so nothing in the working tree can reach the
// extraction, which is the whole point. The tar is read here rather than handed to `tar`, so every
// entry's path is checked against the destination before anything is written.
func extractGitArchive(repository string, commit string, subdirectory string, destination string) error {
	arguments := []string{"-C", repository, "archive", "--format=tar", commit}
	if subdirectory != "" {
		arguments = append(arguments, subdirectory)
	}
	command := exec.Command("git", arguments...)
	var standardError bytes.Buffer
	command.Stderr = &standardError
	stream, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}

	written, extractErr := extractTar(stream, destination)
	// Drained so git is never blocked writing into a pipe nobody reads after an extraction error.
	_, _ = io.Copy(io.Discard, stream)
	waitErr := command.Wait()

	if extractErr != nil {
		return extractErr
	}
	if waitErr != nil {
		return fmt.Errorf("git archive: %w: %s", waitErr, strings.TrimSpace(standardError.String()))
	}
	if written == 0 {
		// An empty extraction builds into nothing that runs, or worse, into something that runs and
		// checks nothing. A commit always has files, so zero is a failure that has to say so.
		return fmt.Errorf("git archive of %s wrote no files, which cannot be right", release.ShortCommit(commit))
	}
	return nil
}

// extractTar writes a tar stream under destination, returning how many files it wrote.
func extractTar(stream io.Reader, destination string) (int, error) {
	reader := tar.NewReader(stream)
	written := 0
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return written, nil
		}
		if err != nil {
			return written, fmt.Errorf("reading the archive: %w", err)
		}

		target := filepath.Join(destination, filepath.FromSlash(header.Name))
		if target != destination && !strings.HasPrefix(target, destination+string(filepath.Separator)) {
			return written, fmt.Errorf("archive entry %q would be written outside %s", header.Name, destination)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return written, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return written, err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode)&0o777)
			if err != nil {
				return written, err
			}
			if _, err := io.Copy(file, reader); err != nil {
				file.Close()
				return written, err
			}
			if err := file.Close(); err != nil {
				return written, err
			}
			written++
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return written, err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return written, err
			}
		case tar.TypeXGlobalHeader:
			// git archive opens with a pax header carrying the commit id. It describes the archive and
			// is not a file.
		default:
			return written, fmt.Errorf("archive entry %q has type %q, which the extraction does not handle", header.Name, header.Typeflag)
		}
	}
}

// gitOutput runs git in a directory and returns its standard output, carrying stderr into the error.
func gitOutput(directory string, arguments ...string) (string, error) {
	command := exec.Command("git", append([]string{"-C", directory}, arguments...)...)
	var standardError bytes.Buffer
	command.Stderr = &standardError
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(standardError.String()))
	}
	return string(output), nil
}

// platformBinaryPrefix is how every hash-named cohere binary for this platform begins.
func platformBinaryPrefix() string {
	return fmt.Sprintf("cohere-%s-%s-", runtime.GOOS, runtime.GOARCH)
}

// committedCompilerUpstream names the vendored compiler's repository, as "owner/name", from the
// snapshot's committed `.gitmodules`, or answers empty when it does not say.
//
// The committed file and not the submodule's `origin`. A remote re-pointed in one checkout is an
// uncommitted change, read from no commit, by the same reasoning that reads the pin from the commit's
// gitlink rather than from the submodule's checkout. A snapshot has no `.git` to ask anyway.
func committedCompilerUpstream(snapshot string) string {
	contents, err := os.ReadFile(filepath.Join(snapshot, ".gitmodules"))
	if err != nil {
		return ""
	}
	if upstream := release.CompilerUpstreamFromGitmodules(string(contents)); upstream != "unknown" {
		return upstream
	}
	return ""
}
