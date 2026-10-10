// Package tracked answers what git answers about a repository, its commit and its tracked files, for a source that
// has no .git. Loom's runners run prebuilt tests in a source unpacked from chunks of the tree's tracked files, and
// a chunk carries no .git (#cyasrr4). build-tree writes git's own answers for the tree and each submodule into that
// source, under ManifestDirectory at the tree's root, and this package reads them there. A directory that holds
// .git is asked through git, exactly as before: the manifest is read only where git can't answer.
//
// The manifest is one directory per repository, at the repository's path under ManifestDirectory (the tree's own at
// its top, cohere's at cohere, cohere's TypeScript at cohere/TypeScript), holding two files, each a git command's
// output byte for byte, run in that repository:
//
//	HEAD   git rev-parse HEAD
//	files  git ls-tree -r -z --full-tree HEAD
//
// files is the commit's own listing, so its object ids are what each tracked file must hash to, and a submodule
// appears in its parent's files as a gitlink (mode 160000) naming the commit its own HEAD must equal. Write writes it.
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
	"strings"
	"sync"
)

// ManifestDirectory is where a source with no .git carries what git would answer about it, at the tree's root.
const ManifestDirectory = ".tracked"

const headFile = "HEAD"
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
		return nil
	}
	if err := expand("", found.entries, found.directory); err != nil {
		return nil, err
	}
	return paths, nil
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
		if (info.Mode().Perm()&0o111 != 0) != (entry.Mode == "100755") {
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
	var sum hash.Hash
	switch digits {
	case 40:
		sum = sha1.New()
	case 64:
		sum = sha256.New()
	default:
		return "", fmt.Errorf("an object id of %d digits", digits)
	}
	fmt.Fprintf(sum, "blob %d\x00", len(data))
	sum.Write(data)
	return hex.EncodeToString(sum.Sum(nil)), nil
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
