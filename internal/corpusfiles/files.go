// Package corpusfiles selects test inputs from Git, never from a directory walk. A source with no .git, as a Loom
// runner unpacks it, answers the same questions from the manifest build-tree writes into it (internal/tracked).
package corpusfiles

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/tracked"
)

const CohereCommit = "7945d102a6c18dd36adf9114a758ce646e8b2359"
const TypeScriptCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"
const TypeScriptGoCommit = "d92d9bfee114c80be2c375d72edae966176e3a4f"
const PrettierCommit = "cb4b33fba24a8428d00e54be85fc886288a374ea"

// Repository returns sorted absolute paths tracked at HEAD under each named root.
// Patterns are case-insensitive basename globs, such as *.md or *_test.go.
// Matching tracked files must be clean in the index and worktree separately.
func Repository(t testing.TB, checkout string, roots, patterns []string) []string {
	t.Helper()
	return checked(t, checkout, "", roots, patterns)
}

// Upstream requires the exact commit and an empty porcelain status, including
// untracked and ignored files, under every root. Sparse membership does not
// reduce ls-files: every selected tracked file must also exist on disk.
func Upstream(t testing.TB, checkout, pin string, roots, patterns []string) []string {
	t.Helper()
	if len(pin) != 40 || strings.Trim(pin, "0123456789abcdef") != "" {
		t.Fatalf("upstream corpus %s: invalid commit pin %q", checkout, pin)
	}
	return checked(t, checkout, pin, roots, patterns)
}

func checked(t testing.TB, checkout, pin string, roots, patterns []string) []string {
	t.Helper()
	files, actual, counts, err := selectFiles(checkout, pin, roots, patterns)
	if err != nil {
		t.Fatalf("corpus-files: %v", err)
	}
	kind := "repository"
	if pin != "" {
		kind = "upstream"
	}
	source := ""
	if !tracked.Git(checkout) {
		source = " (no .git: the tree's " + tracked.ManifestDirectory + " manifest)"
	}
	for i, root := range roots {
		t.Logf("corpus-files: %s %s root %s pin %s count %d%s", kind, checkout, root, actual, counts[i], source)
	}
	return files
}

func git(checkout string, args ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", checkout}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), checkout, err, output)
	}
	return output, nil
}

func names(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	return strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
}

func matches(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if ok, _ := path.Match(strings.ToLower(pattern), strings.ToLower(path.Base(name))); ok {
			return true
		}
	}
	return false
}

func selectFiles(checkout, pin string, roots, patterns []string) ([]string, string, []int, error) {
	if len(roots) == 0 || len(patterns) == 0 {
		return nil, "", nil, fmt.Errorf("%s: named roots and patterns are required", checkout)
	}
	for _, pattern := range patterns {
		if _, err := path.Match(pattern, ""); err != nil || strings.Contains(pattern, "/") {
			return nil, "", nil, fmt.Errorf("%s: invalid basename pattern %q", checkout, pattern)
		}
	}
	checkout, err := filepath.Abs(checkout)
	if err != nil {
		return nil, "", nil, err
	}
	for _, root := range roots {
		if root == "" || filepath.IsAbs(root) || filepath.Clean(root) != root || root == ".." || strings.HasPrefix(root, "../") {
			return nil, "", nil, fmt.Errorf("%s: invalid named root %q", checkout, root)
		}
	}
	if !tracked.Git(checkout) {
		return selectManifest(checkout, pin, roots, patterns)
	}
	head, err := git(checkout, "rev-parse", "HEAD")
	if err != nil {
		return nil, "", nil, err
	}
	actual := strings.TrimSpace(string(head))
	if pin != "" && actual != pin {
		return nil, actual, nil, fmt.Errorf("%s: HEAD %s differs from pin %s", checkout, actual, pin)
	}
	counts := make([]int, len(roots))
	selected := map[string]bool{}
	for i, root := range roots {
		if _, err := os.Stat(filepath.Join(checkout, root)); err != nil {
			return nil, actual, nil, fmt.Errorf("%s: missing named root %s: %w", checkout, root, err)
		}
		spec := ":(top,literal)" + filepath.ToSlash(root)
		if root == "." {
			spec = ":(top)**"
		}
		if pin != "" {
			status, err := git(checkout, "status", "--porcelain=v1", "--ignored", "--untracked-files=all", "-z", "--", spec)
			if err != nil {
				return nil, actual, nil, err
			}
			if len(status) != 0 {
				return nil, actual, nil, fmt.Errorf("%s root %s: dirty, untracked or ignored upstream paths: %q", checkout, root, names(status))
			}
		}
		// A staged edit restored only in the worktree must not hide the index.
		for _, side := range []struct {
			name string
			args []string
		}{
			{"index", []string{"diff", "--cached", "--name-only", "--no-renames", "-z", "HEAD", "--", spec}},
			{"worktree", []string{"diff", "--name-only", "--no-renames", "-z", "--", spec}},
		} {
			data, err := git(checkout, side.args...)
			if err != nil {
				return nil, actual, nil, err
			}
			var dirty []string
			for _, name := range names(data) {
				if matches(name, patterns) {
					dirty = append(dirty, name)
				}
			}
			if len(dirty) != 0 {
				return nil, actual, nil, fmt.Errorf("%s root %s: dirty %s paths against HEAD: %q", checkout, root, side.name, dirty)
			}
		}
		data, err := git(checkout, "ls-files", "--cached", "--full-name", "-z", "--", spec)
		if err != nil {
			return nil, actual, nil, err
		}
		var missing []string
		for _, name := range names(data) {
			if !matches(name, patterns) {
				continue
			}
			absolute := filepath.Join(checkout, filepath.FromSlash(name))
			info, err := os.Stat(absolute)
			if err != nil || info.IsDir() {
				missing = append(missing, name)
				continue
			}
			selected[absolute] = true
			counts[i]++
		}
		if len(missing) != 0 {
			return nil, actual, nil, fmt.Errorf("%s root %s: missing tracked corpus files (including sparse-checkout omissions): %q", checkout, root, missing)
		}
		if counts[i] == 0 {
			return nil, actual, nil, fmt.Errorf("%s root %s: no tracked files match %q", checkout, root, patterns)
		}
	}
	var files []string
	for name := range selected {
		files = append(files, name)
	}
	sort.Strings(files)
	return files, actual, counts, nil
}

// selectManifest is selectFiles for a source with no .git: the commit and the tracked entries come from the tree's
// manifest, and the checks git makes against HEAD are made against it, file by file. There is no index, so the
// manifest is HEAD and index at once. Upstream's status becomes: every file under a root is a recorded entry (anything
// else is untracked, ignored or not), and every entry under it is on disk as recorded. Repository's worktree check
// becomes: every matching entry is on disk as recorded, and no matching file that the .gitignore files leave in is
// missing from the record, which would make the selection smaller than git's.
func selectManifest(checkout, pin string, roots, patterns []string) ([]string, string, []int, error) {
	actual, err := tracked.Head(checkout)
	if err != nil {
		return nil, "", nil, err
	}
	if pin != "" && actual != pin {
		return nil, actual, nil, fmt.Errorf("%s: HEAD %s differs from pin %s", checkout, actual, pin)
	}
	entries, err := tracked.Files(checkout)
	if err != nil {
		return nil, actual, nil, err
	}
	counts := make([]int, len(roots))
	selected := map[string]bool{}
	for i, root := range roots {
		if _, err := os.Stat(filepath.Join(checkout, root)); err != nil {
			return nil, actual, nil, fmt.Errorf("%s: missing named root %s: %w", checkout, root, err)
		}
		under := tracked.Under(entries, filepath.ToSlash(root))
		var matching []tracked.Entry
		for _, entry := range under {
			if matches(entry.Path, patterns) {
				matching = append(matching, entry)
			}
		}
		if pin != "" {
			dirty, err := unrecorded(checkout, root, under)
			if err != nil {
				return nil, actual, nil, err
			}
			missing, changed, err := tracked.Changed(checkout, under)
			if err != nil {
				return nil, actual, nil, err
			}
			dirty = append(dirty, changed...)
			for _, name := range missing {
				// A missing corpus file is named below as one; any other is a deletion, which status reports.
				if !matches(name, patterns) {
					dirty = append(dirty, name)
				}
			}
			if len(dirty) != 0 {
				sort.Strings(dirty)
				return nil, actual, nil, fmt.Errorf("%s root %s: dirty, untracked or ignored upstream paths: %q", checkout, root, dirty)
			}
		} else {
			_, changed, err := tracked.Changed(checkout, matching)
			if err != nil {
				return nil, actual, nil, err
			}
			if len(changed) != 0 {
				return nil, actual, nil, fmt.Errorf("%s root %s: dirty worktree paths against HEAD: %q", checkout, root, changed)
			}
			// git would select a file its commit tracks; one the manifest leaves out means the manifest is stale.
			extra, err := tracked.Unrecorded(checkout, root)
			if err != nil {
				return nil, actual, nil, err
			}
			var unlisted []string
			for _, name := range extra {
				if matches(name, patterns) {
					unlisted = append(unlisted, name)
				}
			}
			if len(unlisted) != 0 {
				return nil, actual, nil, fmt.Errorf("%s root %s: corpus files the tree's manifest doesn't record and no .gitignore ignores (a stale manifest, or a leftover): %q", checkout, root, unlisted)
			}
		}
		var missing []string
		for _, entry := range matching {
			absolute := filepath.Join(checkout, filepath.FromSlash(entry.Path))
			info, err := os.Stat(absolute)
			if err != nil || info.IsDir() {
				missing = append(missing, entry.Path)
				continue
			}
			selected[absolute] = true
			counts[i]++
		}
		if len(missing) != 0 {
			return nil, actual, nil, fmt.Errorf("%s root %s: missing tracked corpus files (including sparse-checkout omissions): %q", checkout, root, missing)
		}
		if counts[i] == 0 {
			return nil, actual, nil, fmt.Errorf("%s root %s: no tracked files match %q", checkout, root, patterns)
		}
	}
	var files []string
	for name := range selected {
		files = append(files, name)
	}
	sort.Strings(files)
	return files, actual, counts, nil
}

// unrecorded is every file or link under root that no entry records, as git status --ignored --untracked-files=all
// lists untracked and ignored paths alike. A submodule's directory is its own repository's, as it is to status, and
// the manifest at the tree's root is the source's record, as .git is a checkout's.
func unrecorded(checkout, root string, under []tracked.Entry) ([]string, error) {
	recorded, submodules := map[string]bool{}, map[string]bool{}
	for _, entry := range under {
		if entry.Submodule() {
			submodules[entry.Path] = true
		} else {
			recorded[entry.Path] = true
		}
	}
	var found []string
	err := filepath.WalkDir(filepath.Join(checkout, root), func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(checkout, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if submodules[relative] || relative == tracked.ManifestDirectory {
				return filepath.SkipDir
			}
			return nil
		}
		if !recorded[relative] {
			found = append(found, relative)
		}
		return nil
	})
	return found, err
}
