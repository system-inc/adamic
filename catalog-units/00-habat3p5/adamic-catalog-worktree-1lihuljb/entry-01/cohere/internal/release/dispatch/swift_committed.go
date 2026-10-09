package dispatch

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/system-inc/cohere/internal/release/packaging"
)

// The Swift engine a gate runs is built from the same commit as the cohere binary running it.
//
// Until this existed it was built from the shared working tree's `swift/`, so any member's uncommitted
// Swift rule was live in every Swift project's gate, the defect #yd6yprj removed for the cohere binary.
//
// The tree it builds is a detached git worktree at the commit, not a `git archive` extraction, for two
// reasons that were measured rather than assumed (2026-10-02):
//
//   - The engine's manifest reads `Context.gitInformation?.hasUncommittedChanges ?? true`. An extraction
//     has no git context, so every committed engine would report itself modified. A worktree at the
//     commit reports false, and that reading also rules out SwiftPM having found the enclosing checkout,
//     which was dirty at the time.
//   - Moving one worktree between commits rewrites only the files that differ, so SwiftPM's build stays
//     warm: a cold build took 225s, a move three Swift files away rebuilt in 20s, an unchanged tree in 2s.
//
// The worktree belongs to this cache and nobody edits it. If it is ever found dirty the build is refused
// rather than forced, because forcing would discard something without knowing what it was.

// SwiftSourceWorktree is the worktree the committed engine is built from.
func (paths Paths) SwiftSourceWorktree() string {
	return filepath.Join(paths.CacheDirectory, "swift-source")
}

// SwiftCommittedBuildDirectory is the SwiftPM scratch path for committed builds. Separate from the
// working-tree engine's, because one scratch path fed two package directories would rebuild cold every
// time the other was used.
func (paths Paths) SwiftCommittedBuildDirectory() string {
	return filepath.Join(paths.CacheDirectory, "swift-build-committed")
}

// swiftSourceLock serializes moving the worktree and building in it.
func (paths Paths) swiftSourceLock() string {
	return filepath.Join(paths.CacheDirectory, "swift-source.lock")
}

// swiftEngineInputs are what decide the engine, as `commit:path` objects: the same three the
// working-tree hash reads, so a commit that changes only tests reuses the binary.
var swiftEngineInputs = []string{"swift/Package.swift", "swift/Package.resolved", "swift/Sources"}

// resolveCommittedSwiftEngine returns the engine for commit, building it when the cache has none.
func resolveCommittedSwiftEngine(paths Paths, commit string, toolchain string, contract int) (string, bool, error) {
	hash, err := committedSwiftEngineHash(paths, commit, toolchain)
	if err != nil {
		return "", false, err
	}
	binaryPath := paths.SwiftEngineBinaryPath(hash)
	if _, statErr := os.Stat(binaryPath); statErr == nil {
		return binaryPath, false, nil
	}

	if err := os.MkdirAll(paths.CacheDirectory, 0o755); err != nil {
		return "", false, fmt.Errorf("creating %s: %w", paths.CacheDirectory, err)
	}
	unlock, err := lockExclusive(paths.swiftSourceLock())
	if err != nil {
		return "", false, err
	}
	defer unlock()

	// Asked again under the lock: a run waiting here behind another that built this same engine finds
	// it, and does not move the worktree or build again.
	if _, statErr := os.Stat(binaryPath); statErr == nil {
		return binaryPath, false, nil
	}

	worktree := paths.SwiftSourceWorktree()
	if err := checkOutWorktree(paths.ModuleDirectory, worktree, commit); err != nil {
		return "", false, err
	}

	// Built to a temporary name and renamed only once proven, so a concurrent run never execs a binary
	// that failed its proof.
	temporary := binaryPath + fmt.Sprintf(".partial-%d", os.Getpid())
	defer os.Remove(temporary)

	if err := buildSwiftProduct(filepath.Join(worktree, "swift"), paths.SwiftCommittedBuildDirectory(), temporary); err != nil {
		return "", false, err
	}
	if err := proveCommittedSwiftEngine(temporary, commit, contract); err != nil {
		return "", false, err
	}
	if err := os.Rename(temporary, binaryPath); err != nil {
		return "", false, fmt.Errorf("moving the built Swift engine into the cache: %w", err)
	}
	return binaryPath, true, nil
}

// committedSwiftEngineHash names the engine for commit without reading a single file from disk.
//
// Git object ids are content addresses, so `commit:swift/Sources` names exactly the sources at that
// commit. A cache hit therefore never touches the worktree, and after the first run at a commit it
// starts no git either: see swiftEngineInputObjects.
func committedSwiftEngineHash(paths Paths, commit string, toolchain string) (string, error) {
	objects, err := swiftEngineInputObjects(paths, commit)
	if err != nil {
		return "", err
	}

	digest := sha256.New()
	fmt.Fprintf(digest, "kind\x00%s\x00", snapshotTag)
	fmt.Fprintf(digest, "toolchain\x00%s\x00", toolchain)
	fmt.Fprintf(digest, "platform\x00%s/%s\x00", runtime.GOOS, runtime.GOARCH)
	for index, input := range swiftEngineInputs {
		fmt.Fprintf(digest, "input\x00%s\x00%s\x00", input, objects[index])
	}
	return hex.EncodeToString(digest.Sum(nil))[:hashLength], nil
}

// swiftEngineInputObjects are the object ids of the engine's inputs at commit, one per
// swiftEngineInputs entry.
//
// A commit's tree never changes, so the answer is asked of git once and kept, named for the commit and
// for the list of inputs asked about: a cohere whose list differs reads its own file. A kept answer that
// is not the right number of object names is asked again rather than trusted.
func swiftEngineInputObjects(paths Paths, commit string) ([]string, error) {
	list := sha256.Sum256([]byte(strings.Join(swiftEngineInputs, "\x00")))
	kept := filepath.Join(paths.CacheDirectory, "swift-inputs", commit+"-"+hex.EncodeToString(list[:])[:hashLength])
	if contents, err := os.ReadFile(kept); err == nil {
		objects := strings.Fields(string(contents))
		valid := len(objects) == len(swiftEngineInputs)
		for _, object := range objects {
			valid = valid && isObjectName(object)
		}
		if valid {
			return objects, nil
		}
	}

	specifications := make([]string, 0, len(swiftEngineInputs))
	for _, input := range swiftEngineInputs {
		specifications = append(specifications, commit+":"+input)
	}
	// No --verify: it accepts exactly one argument. Each specification here is `<hex commit>:<path>`.
	output, err := gitOutput(paths.ModuleDirectory, append([]string{"rev-parse"}, specifications...)...)
	if err != nil {
		// rev-parse with several arguments stops at the first it cannot resolve, so the message names
		// what was asked rather than guessing which one was missing.
		return nil, fmt.Errorf("reading the Swift engine's inputs at %s (%s): %w", release.ShortCommit(commit), strings.Join(swiftEngineInputs, ", "), err)
	}
	objects := strings.Fields(output)
	if len(objects) != len(swiftEngineInputs) {
		return nil, fmt.Errorf("asked git for %d Swift engine inputs at %s and got %d answers", len(swiftEngineInputs), release.ShortCommit(commit), len(objects))
	}

	// Kept through a temporary and a rename, so a concurrent run reads a whole answer or none. Failing to
	// keep it costs the next run one git call, so it is not an error.
	if err := os.MkdirAll(filepath.Dir(kept), 0o755); err == nil {
		if temporary, err := os.CreateTemp(filepath.Dir(kept), ".inputs-*"); err == nil {
			_, writeErr := temporary.WriteString(strings.Join(objects, "\n") + "\n")
			closeErr := temporary.Close()
			if writeErr != nil || closeErr != nil || os.Rename(temporary.Name(), kept) != nil {
				os.Remove(temporary.Name())
			}
		}
	}
	return objects, nil
}

// checkOutWorktree puts the worktree at commit, creating it the first time.
func checkOutWorktree(moduleDirectory string, worktree string, commit string) error {
	if _, err := os.Stat(worktree); errors.Is(err, os.ErrNotExist) {
		if _, err := gitOutput(moduleDirectory, "worktree", "add", "--detach", worktree, commit); err != nil {
			return fmt.Errorf("creating the Swift engine's worktree at %s: %w", worktree, err)
		}
		return requireWorktreeAt(worktree, commit)
	}

	// The worktree exists, so it must be ours and clean before it moves. A plain checkout refuses to
	// overwrite local changes on its own; this says why before git does, and names the path.
	if err := requireCleanWorktree(worktree); err != nil {
		return err
	}
	if _, err := gitOutput(worktree, "checkout", "--quiet", "--detach", commit); err != nil {
		return fmt.Errorf("moving the Swift engine's worktree to %s: %w", release.ShortCommit(commit), err)
	}
	return requireWorktreeAt(worktree, commit)
}

// requireCleanWorktree refuses a worktree with anything uncommitted in it, untracked files included.
func requireCleanWorktree(worktree string) error {
	status, err := gitOutput(worktree, "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("%s is not a usable git worktree, so the Swift engine was not built (remove it by name and it will be recreated): %w", worktree, err)
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("the Swift engine's worktree %s has changes nobody should have made there, so it was not moved or built:\n%s",
			worktree, strings.TrimRight(status, "\n"))
	}
	return nil
}

// requireWorktreeAt confirms the worktree is clean and at commit, which is what the build is about to
// claim.
func requireWorktreeAt(worktree string, commit string) error {
	head, err := gitOutput(worktree, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("reading the Swift engine's worktree: %w", err)
	}
	if strings.TrimSpace(head) != commit {
		return fmt.Errorf("the Swift engine's worktree is at %s after checking out %s", release.ShortCommit(strings.TrimSpace(head)), release.ShortCommit(commit))
	}
	return requireCleanWorktree(worktree)
}

// proveCommittedSwiftEngine asks the built engine what it is and refuses it unless it reports a clean
// source tree.
//
// The engine's commit field reads "dev" until the release build stamps it (EngineVersion.swift), so it
// cannot be checked here yet. What can be checked is the bit its manifest computed from the tree it was
// built in, which is exactly the property this build exists to have.
func proveCommittedSwiftEngine(binaryPath string, commit string, contract int) error {
	command := exec.Command(binaryPath, "--contract", fmt.Sprint(contract), "--version")
	var standardError bytes.Buffer
	command.Stderr = &standardError
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("the Swift engine built from %s failed its own --version (%w), so it was not cached: %s",
			release.ShortCommit(commit), err, strings.TrimSpace(standardError.String()))
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		var record struct {
			Kind               string `json:"kind"`
			SourceTreeModified *bool  `json:"sourceTreeModified"`
		}
		if json.Unmarshal(scanner.Bytes(), &record) != nil || record.Kind != "provenance" {
			continue
		}
		if record.SourceTreeModified == nil {
			return fmt.Errorf("the Swift engine built from %s reported provenance without sourceTreeModified, so it was not cached", release.ShortCommit(commit))
		}
		if *record.SourceTreeModified {
			return fmt.Errorf("the Swift engine built from %s reports a modified source tree, which a build from a clean worktree at that commit cannot have, so it was not cached",
				release.ShortCommit(commit))
		}
		return nil
	}
	return fmt.Errorf("the Swift engine built from %s printed no provenance record for --version, so it was not cached:\n%s", release.ShortCommit(commit), output)
}
