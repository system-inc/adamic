// Package tracked answers what git answers about a repository, its commit and its tracked files, for a source that
// has no .git. Loom's runners run prebuilt tests in a source unpacked from chunks of the tree's tracked files, and
// a chunk carries no .git (#cyasrr4). build-tree writes git's own answers for the tree and each submodule into that
// source, under ManifestDirectory at the tree's root, and this package reads them there. A directory that holds
// .git is asked through git, exactly as before: the manifest is read only where git can't answer.
//
// The manifest is one directory per repository, at the repository's path under ManifestDirectory (the tree's own at
// its top, cohere's at cohere, cohere's TypeScript at cohere/TypeScript), holding three files, each a git command's
// output byte for byte, run in that repository:
//
//	HEAD    git rev-parse HEAD
//	commit  git cat-file commit HEAD
//	files   git ls-tree -r -z --full-tree HEAD
//
// The three are bound to each other as git binds them: commit hashes to HEAD, and its tree line is the tree that
// files rebuilds, so a listing from any other commit is refused. files is that commit's own listing, so its object
// ids are what each tracked file must hash to, and a submodule appears in its parent's files as a gitlink (mode
// 160000) naming the commit its own HEAD must equal. Write writes it.
package tracked

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// ManifestDirectory is where a source with no .git carries what git would answer about it, at the tree's root.
const ManifestDirectory = ".tracked"

const headFile = "HEAD"
const commitFile = "commit"
const filesFile = "files"

// An Entry is one tracked path: its git mode (100644, 100755, 120000, or 160000 for a submodule's gitlink), its object
// (the blob it holds, or the commit a gitlink names), and its path in the repository, slash-separated.
type Entry struct {
	Mode   string
	Object string
	Path   string
}

// Submodule reports whether the entry is a gitlink: a submodule's commit, not a file.
func (entry Entry) Submodule() bool {
	return entry.Mode == "160000"
}

// Git reports whether git answers for directory: it holds .git, a repository's directory or a submodule's file.
func Git(directory string) bool {
	_, err := os.Lstat(filepath.Join(directory, ".git"))
	return err == nil
}

// Head is the commit directory's repository is at: git rev-parse HEAD where git answers, else the manifest's HEAD,
// which must equal the gitlink its parent repository's manifest records for it.
func Head(directory string) (string, error) {
	if Git(directory) {
		output, err := git(directory, "rev-parse", "HEAD")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}
	found, err := manifestOf(directory)
	if err != nil {
		return "", err
	}
	return found.head, nil
}

// Files is directory's repository's tracked entries in git's order: git ls-files --stage where git answers, else the
// manifest's. A submodule is one gitlink entry; its own files are its repository's.
func Files(directory string) ([]Entry, error) {
	if Git(directory) {
		output, err := git(directory, "ls-files", "--stage", "-z")
		if err != nil {
			return nil, err
		}
		return parseStage(directory, output)
	}
	found, err := manifestOf(directory)
	if err != nil {
		return nil, err
	}
	return found.entries, nil
}

// Recursive is every tracked file of directory's repository and of each submodule inside it, repository-relative, in
// the order git ls-files --recurse-submodules prints them: a submodule's files where its gitlink sorts. Where git
// answers it runs that command; in a manifest a gitlink with no manifest of its own is refused, not skipped.
func Recursive(directory string) ([]string, error) {
	if Git(directory) {
		output, err := git(directory, "ls-files", "--recurse-submodules", "-z")
		if err != nil {
			return nil, err
		}
		return names(output), nil
	}
	found, err := manifestOf(directory)
	if err != nil {
		return nil, err
	}
	var paths []string
	var expand func(prefix string, entries []Entry, directory string) error
	expand = func(prefix string, entries []Entry, directory string) error {
		for _, entry := range entries {
			if !entry.Submodule() {
				paths = append(paths, prefix+entry.Path)
				continue
			}
			submodule := filepath.Join(directory, filepath.FromSlash(entry.Path))
			inner, err := manifestOf(submodule)
			if err != nil {
				return err
			}
			if err := expand(prefix+entry.Path+"/", inner.entries, submodule); err != nil {
				return err
			}
		}
		// A file on disk the record leaves out would be left out of the listing: a stale manifest, refused by name.
		extra, err := unrecorded(directory, ".", entries)
		if err != nil {
			return err
		}
		if len(extra) != 0 {
			return fmt.Errorf("%s: files the tree's manifest doesn't record and no .gitignore ignores (a stale manifest, or a leftover): %q", directory, extra)
		}
		return nil
	}
	if err := expand("", found.entries, found.directory); err != nil {
		return nil, err
	}
	return paths, nil
}

// Unrecorded is every file at or under root (repository-relative) in directory's repository that the manifest
// doesn't record and the repository's own ignore rules (its .gitignore files, as git reads them) don't ignore: what
// git status lists as untracked. In a source with no .git it is how a stale manifest shows, a file the commit tracks
// but the record left out. Where git answers it is nil: untracked files are git's to judge, as they always were.
func Unrecorded(directory, root string) ([]string, error) {
	if Git(directory) {
		return nil, nil
	}
	found, err := manifestOf(directory)
	if err != nil {
		return nil, err
	}
	return unrecorded(found.directory, root, found.entries)
}

// unrecorded walks root in the repository at directory the way git status does: into every directory the record
// holds a file under, past a submodule's (its own repository's), and, for anything else, asking git's ignore rules
// before it goes further, so an ignored node_modules is one question rather than thousands of files. A directory
// holding .git is another repository, listed as itself. The manifest at the tree's root is the source's record.
func unrecorded(directory, root string, entries []Entry) ([]string, error) {
	root = path.Clean(filepath.ToSlash(root))
	recorded, directories, submodules := map[string]bool{}, map[string]bool{".": true}, map[string]bool{}
	for _, entry := range entries {
		if entry.Submodule() {
			submodules[entry.Path] = true
		} else {
			recorded[entry.Path] = true
		}
		for parent := path.Dir(entry.Path); parent != "."; parent = path.Dir(parent) {
			directories[parent] = true
		}
	}
	_, statErr := os.Stat(filepath.Join(directory, ManifestDirectory))
	holdsManifest := statErr == nil
	var found, candidates []string
	var visit func(relative string, known bool) error
	// visit lists a directory's children: a known directory's unknown children become candidates for the ignore
	// rules, and a candidate directory that isn't ignored has all its children asked in turn.
	visit = func(relative string, known bool) error {
		children, err := os.ReadDir(filepath.Join(directory, filepath.FromSlash(relative)))
		if err != nil {
			return err
		}
		for _, child := range children {
			name := path.Join(relative, child.Name())
			switch {
			case child.IsDir() && (submodules[name] || name == ManifestDirectory && holdsManifest):
			case child.IsDir() && known && directories[name]:
				if err := visit(name, true); err != nil {
					return err
				}
			case !child.IsDir() && known && recorded[name]:
			default:
				candidates = append(candidates, name)
			}
		}
		return nil
	}
	info, err := os.Lstat(filepath.Join(directory, filepath.FromSlash(root)))
	if err != nil {
		return nil, err
	}
	switch {
	case !info.IsDir() && !recorded[root]:
		candidates = append(candidates, root)
	case info.IsDir() && directories[root] && !submodules[root]:
		if err := visit(root, true); err != nil {
			return nil, err
		}
	case info.IsDir() && !submodules[root]:
		candidates = append(candidates, root)
	}
	for len(candidates) != 0 {
		ignored, err := ignoredPaths(directory, candidates)
		if err != nil {
			return nil, err
		}
		asked := candidates
		candidates = nil
		for _, name := range asked {
			if ignored[name] {
				continue
			}
			info, err := os.Lstat(filepath.Join(directory, filepath.FromSlash(name)))
			if err != nil {
				return nil, err
			}
			if !info.IsDir() {
				found = append(found, name)
				continue
			}
			if _, err := os.Lstat(filepath.Join(directory, filepath.FromSlash(name), ".git")); err == nil {
				found = append(found, name+"/")
				continue
			}
			if err := visit(name, false); err != nil {
				return nil, err
			}
		}
	}
	sort.Strings(found)
	return found, nil
}

// ignoredPaths asks git which of asked (relative to directory) the repository's own ignore rules ignore: its
// .gitignore files and nothing of the machine's (no global or system configuration, no core.excludesFile, an empty
// info/exclude). The directory has no .git, so git runs against an empty scratch repository with directory as its
// work tree, and --no-index reads only the rules.
func ignoredPaths(directory string, asked []string) (map[string]bool, error) {
	scratch, err := os.MkdirTemp("", "tracked-ignore-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch)
	environment := append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	initialize := exec.Command("git", "init", "-q", "--bare", scratch)
	initialize.Env = environment
	if output, err := initialize.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git init for the ignore rules: %w: %s", err, output)
	}
	if err := os.WriteFile(filepath.Join(scratch, "info", "exclude"), nil, 0o644); err != nil {
		return nil, err
	}
	var input bytes.Buffer
	for _, name := range asked {
		input.WriteString(name)
		input.WriteByte(0)
	}
	command := exec.Command("git", "--git-dir="+scratch, "--work-tree="+directory, "-c", "core.bare=false", "-c", "core.excludesFile="+os.DevNull,
		"check-ignore", "--no-index", "-z", "--stdin")
	command.Dir = directory
	command.Env = environment
	command.Stdin = &input
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	// check-ignore exits 1 when it ignores nothing.
	if exit := (*exec.ExitError)(nil); errors.As(err, &exit) && exit.ExitCode() == 1 && stderr.Len() == 0 {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("git check-ignore in %s: %w: %s", directory, err, strings.TrimSpace(stderr.String()))
	}
	ignored := map[string]bool{}
	for _, name := range names(output) {
		ignored[name] = true
	}
	return ignored, nil
}

// Under is the entries at or under root, a repository-relative path; "." is every entry.
func Under(entries []Entry, root string) []Entry {
	root = path.Clean(filepath.ToSlash(root))
	var under []Entry
	for _, entry := range entries {
		if root == "." || entry.Path == root || strings.HasPrefix(entry.Path, root+"/") {
			under = append(under, entry)
		}
	}
	return under
}

// Changed reads each file entry (gitlinks are skipped) in directory and names those missing, and those that aren't
// what the record holds: another kind (a file for a link, a directory for a file), another executable bit, or content
// that doesn't hash to the entry's object. It is what git diff against HEAD answers, for a source with no .git.
func Changed(directory string, entries []Entry) (missing, changed []string, err error) {
	type verdict struct {
		missing, changed bool
		err              error
	}
	verdicts := make([]verdict, len(entries))
	work := make(chan int)
	var group sync.WaitGroup
	for range max(1, runtime.GOMAXPROCS(0)) {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range work {
				missing, changed, err := differs(directory, entries[index])
				verdicts[index] = verdict{missing, changed, err}
			}
		}()
	}
	for index, entry := range entries {
		if !entry.Submodule() {
			work <- index
		}
	}
	close(work)
	group.Wait()
	for index, verdict := range verdicts {
		switch {
		case verdict.err != nil:
			return nil, nil, verdict.err
		case verdict.missing:
			missing = append(missing, entries[index].Path)
		case verdict.changed:
			changed = append(changed, entries[index].Path)
		}
	}
	return missing, changed, nil
}

func differs(directory string, entry Entry) (missing, changed bool, err error) {
	name := filepath.Join(directory, filepath.FromSlash(entry.Path))
	info, err := os.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}
	var content io.Reader
	switch {
	case entry.Mode == "120000" && info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(name)
		if err != nil {
			return false, false, err
		}
		content = strings.NewReader(target)
	case (entry.Mode == "100644" || entry.Mode == "100755") && info.Mode().IsRegular():
		// git's rule: a file is executable when its owner may execute it.
		if (info.Mode().Perm()&0o100 != 0) != (entry.Mode == "100755") {
			return false, true, nil
		}
		file, err := os.Open(name)
		if err != nil {
			return false, false, err
		}
		defer file.Close()
		content = file
	default:
		return false, true, nil
	}
	object, err := blobObject(content, len(entry.Object))
	if err != nil {
		return false, false, fmt.Errorf("hashing %s: %w", name, err)
	}
	return false, object != entry.Object, nil
}

// blobObject is git's object id for content as a blob: sha1 for a 40-digit id, sha256 for a 64-digit one.
func blobObject(content io.Reader, digits int) (string, error) {
	data, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}
	return object("blob", data, digits)
}

// Write puts git's answers for checkout, and for each submodule under it that git has checked out, into tree's
// manifest: HEAD and files as the package names them, one repository at a time. It is what build-tree writes into the
// source it chunks, and tests use it to make a source with no .git.
func Write(checkout, tree string) error {
	var write func(directory, relative string) error
	write = func(directory, relative string) error {
		head, err := git(directory, "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		commit, err := git(directory, "cat-file", "commit", "HEAD")
		if err != nil {
			return err
		}
		listing, err := git(directory, "ls-tree", "-r", "-z", "--full-tree", "HEAD")
		if err != nil {
			return err
		}
		destination := filepath.Join(tree, ManifestDirectory, filepath.FromSlash(relative))
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destination, headFile), head, 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destination, commitFile), commit, 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destination, filesFile), listing, 0o644); err != nil {
			return err
		}
		entries, err := parseTree(directory, listing)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			submodule := filepath.Join(directory, filepath.FromSlash(entry.Path))
			if entry.Submodule() && Git(submodule) {
				if err := write(submodule, path.Join(relative, entry.Path)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return write(checkout, "")
}

// A manifest is one repository's record: the directory it describes, its commit and its entries.
type manifest struct {
	directory string
	head      string
	entries   []Entry
}

// manifests holds each directory's manifest once read, for the life of the process: a source's files are fixed while
// its tests run. A failed read is not kept, so a manifest written later is read then.
var manifests sync.Map

// manifestOf finds the tree root at or above directory that holds ManifestDirectory and reads directory's record
// there, refusing one whose commit disagrees with the gitlink its parent repository records for it. A directory with
// neither .git nor a record is refused by name: its commit and its tracked files are unknown.
func manifestOf(directory string) (*manifest, error) {
	directory, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	if found, ok := manifests.Load(directory); ok {
		return found.(*manifest), nil
	}
	found, err := readManifest(directory)
	if err != nil {
		return nil, err
	}
	manifests.Store(directory, found)
	return found, nil
}

func readManifest(directory string) (*manifest, error) {
	root := directory
	for {
		if info, err := os.Stat(filepath.Join(root, ManifestDirectory)); err == nil && info.IsDir() {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return nil, fmt.Errorf("%s has no .git, and no %s manifest at or above it records its commit and tracked files (build-tree writes one into the source it chunks)", directory, ManifestDirectory)
		}
		root = parent
	}
	relative, err := filepath.Rel(root, directory)
	if err != nil {
		return nil, err
	}
	relative = filepath.ToSlash(relative)
	if relative == "." {
		relative = ""
	}
	record := filepath.Join(root, ManifestDirectory, filepath.FromSlash(relative))
	head, err := os.ReadFile(filepath.Join(record, headFile))
	if err != nil {
		return nil, fmt.Errorf("%s has no .git, and the tree's manifest has no commit for it: %w", directory, err)
	}
	listing, err := os.ReadFile(filepath.Join(record, filesFile))
	if err != nil {
		return nil, fmt.Errorf("%s has no .git, and the tree's manifest has no tracked files for it: %w", directory, err)
	}
	found := &manifest{directory: directory, head: strings.TrimSpace(string(head))}
	if !objectID(found.head) {
		return nil, fmt.Errorf("%s: the manifest's commit %q isn't an object id", filepath.Join(record, headFile), found.head)
	}
	if found.entries, err = parseTree(filepath.Join(record, filesFile), listing); err != nil {
		return nil, err
	}
	commit, err := os.ReadFile(filepath.Join(record, commitFile))
	if err != nil {
		return nil, fmt.Errorf("%s has no .git, and the tree's manifest has no commit object for it, so nothing binds its files to its commit: %w", directory, err)
	}
	if err := bound(record, found, commit); err != nil {
		return nil, err
	}
	if relative == "" {
		return found, nil
	}
	// A submodule's record must be the commit its parent pins: the nearest repository above it that has a record.
	for parent := path.Dir(relative); ; parent = path.Dir(parent) {
		parentDirectory := filepath.Join(root, filepath.FromSlash(parent))
		if _, err := os.Stat(filepath.Join(root, ManifestDirectory, filepath.FromSlash(parent), headFile)); err == nil {
			above, err := manifestOf(parentDirectory)
			if err != nil {
				return nil, err
			}
			inside := strings.TrimPrefix(relative, parent+"/")
			if parent == "." {
				inside = relative
			}
			for _, entry := range above.entries {
				if entry.Path == inside && entry.Submodule() {
					if entry.Object != found.head {
						return nil, fmt.Errorf("%s: the manifest's commit %s isn't %s, the gitlink %s records", directory, found.head, entry.Object, parentDirectory)
					}
					return found, nil
				}
			}
			return nil, fmt.Errorf("%s: the manifest records it, but %s's records no gitlink at %s", directory, parentDirectory, inside)
		}
		if parent == "." {
			return nil, fmt.Errorf("%s: the manifest records it, but no repository above it", directory)
		}
	}
}

// parseTree reads git ls-tree -r -z's entries: "<mode> <type> <object>\t<path>", each ended by NUL.
func parseTree(source string, listing []byte) ([]Entry, error) {
	var entries []Entry
	for _, line := range names(listing) {
		header, name, ok := strings.Cut(line, "\t")
		fields := strings.Fields(header)
		if !ok || name == "" || len(fields) != 3 || !knownMode(fields[0]) || !objectID(fields[2]) || (fields[1] == "commit") != (fields[0] == "160000") || (fields[1] != "blob" && fields[1] != "commit") {
			return nil, fmt.Errorf("%s: %q isn't a git ls-tree -r entry", source, line)
		}
		entries = append(entries, Entry{Mode: fields[0], Object: fields[2], Path: name})
	}
	return entries, nil
}

// parseStage reads git ls-files --stage -z's entries: "<mode> <object> <stage>\t<path>", each ended by NUL.
func parseStage(source string, listing []byte) ([]Entry, error) {
	var entries []Entry
	for _, line := range names(listing) {
		header, name, ok := strings.Cut(line, "\t")
		fields := strings.Fields(header)
		if !ok || name == "" || len(fields) != 3 || !knownMode(fields[0]) || !objectID(fields[1]) {
			return nil, fmt.Errorf("%s: %q isn't a git ls-files --stage entry", source, line)
		}
		entries = append(entries, Entry{Mode: fields[0], Object: fields[1], Path: name})
	}
	return entries, nil
}

func knownMode(mode string) bool {
	return mode == "100644" || mode == "100755" || mode == "120000" || mode == "160000"
}

func objectID(text string) bool {
	return (len(text) == 40 || len(text) == 64) && strings.Trim(text, "0123456789abcdef") == ""
}

func names(data []byte) []string {
	data = bytes.TrimSuffix(data, []byte{0})
	if len(data) == 0 {
		return nil
	}
	return strings.Split(string(data), "\x00")
}

func git(directory string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", directory}, arguments...)...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s in %s: %w: %s", strings.Join(arguments, " "), directory, err, strings.TrimSpace(stderr.String()))
	}
	return output, nil
}

// bound refuses a record whose three files aren't one commit's: the commit object must hash to HEAD, and the tree it
// names must be the tree its files listing rebuilds.
func bound(record string, found *manifest, commit []byte) error {
	if id, err := object("commit", commit, len(found.head)); err != nil || id != found.head {
		return fmt.Errorf("%s: the commit object hashes to %s, not the manifest's commit %s (%v)", filepath.Join(record, commitFile), id, found.head, err)
	}
	line, _, _ := bytes.Cut(commit, []byte("\n"))
	named, ok := strings.CutPrefix(string(line), "tree ")
	if !ok || !objectID(named) {
		return fmt.Errorf("%s: the commit object names no tree: %q", filepath.Join(record, commitFile), line)
	}
	rebuilt, err := treeObject(found.entries, len(found.head))
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.Join(record, filesFile), err)
	}
	if rebuilt != named {
		return fmt.Errorf("%s: the listing rebuilds tree %s, and commit %s names tree %s: the files aren't that commit's", filepath.Join(record, filesFile), rebuilt, found.head, named)
	}
	return nil
}

// treeObject is the id of the root tree git would write for entries: each directory a tree object of its children,
// "<mode> <name>\0<raw id>", in git's order (a directory's name compares as if it ended in "/"), subtrees as 40000.
func treeObject(entries []Entry, digits int) (string, error) {
	type child struct {
		mode, name, id string
		tree           bool
	}
	children := map[string][]child{}
	directories := map[string]bool{".": true}
	for _, entry := range entries {
		children[path.Dir(entry.Path)] = append(children[path.Dir(entry.Path)], child{mode: entry.Mode, name: path.Base(entry.Path), id: entry.Object})
		for parent := path.Dir(entry.Path); parent != "." && !directories[parent]; parent = path.Dir(parent) {
			directories[parent] = true
			children[path.Dir(parent)] = append(children[path.Dir(parent)], child{mode: "40000", name: path.Base(parent), tree: true})
		}
	}
	var build func(directory string) (string, error)
	build = func(directory string) (string, error) {
		list := children[directory]
		sortKey := func(item child) string {
			if item.tree {
				return item.name + "/"
			}
			return item.name
		}
		sort.Slice(list, func(left, right int) bool { return sortKey(list[left]) < sortKey(list[right]) })
		var content bytes.Buffer
		for index, item := range list {
			if index > 0 && list[index-1].name == item.name {
				return "", fmt.Errorf("%q is listed twice", path.Join(directory, item.name))
			}
			if item.tree {
				id, err := build(path.Join(directory, item.name))
				if err != nil {
					return "", err
				}
				item.id = id
			}
			raw, err := hex.DecodeString(item.id)
			if err != nil || len(item.id) != digits {
				return "", fmt.Errorf("%q has object %q", path.Join(directory, item.name), item.id)
			}
			fmt.Fprintf(&content, "%s %s\x00", item.mode, item.name)
			content.Write(raw)
		}
		return object("tree", content.Bytes(), digits)
	}
	return build(".")
}

// object is git's id for data as an object of kind: sha1 for a 40-digit repository, sha256 for a 64-digit one.
func object(kind string, data []byte, digits int) (string, error) {
	var sum hash.Hash
	switch digits {
	case 40:
		sum = sha1.New()
	case 64:
		sum = sha256.New()
	default:
		return "", fmt.Errorf("an object id of %d digits", digits)
	}
	fmt.Fprintf(sum, "%s %d\x00", kind, len(data))
	sum.Write(data)
	return hex.EncodeToString(sum.Sum(nil)), nil
}
