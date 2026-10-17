package buildcache

import (
	"cmp"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/toolchain"
)

// A product's key is its read set (#vt46geg): what its last traced build read, evaluated on the tree at hand. The
// declared Files key every file under a directory a product might read (178 products named all of cohere/, so a
// cohere bump rebuilt 584 of 647); a measured read set names only what the build opened, listed, stat'ed or looked for
// and missed, so a change misses only the products that read it.
//
// A deterministic build whose every read answers the same makes the same bytes, so a read set is sound for any tree
// on which it evaluates the same, data-dependent reads included: a build that reads another file when one changes
// read the one that changed. A read set is written only by Workshop's traced builds (trace.go, settle.go), which
// refuse a build that read something no key can name. Its name key holds what isn't a file (the name, flags and tool
// reports, made portable), and the read set what is:
//
//   - tree entries, by path relative to the repository: a file's content and exec bit, a directory's listing of what
//     the tree carries, or only whether a path exists and as what (a stat, or a lookup that missed);
//   - the recipe, the test binary's linked packages (its Go files, directories, go.mod and go.sum), since strace can't
//     see what a build does in process;
//   - other products, by their name key, valued at the key that product has on the same tree;
//   - the Go release, and each tool from Workshop's tools directory, by its report;
//   - paths above the tree and in the home directory that were only looked for, by whether and as what they exist.
//
// The sets live beside the products as <name key>.reads, the last few kept, newest first, since a product's reads can
// change with the tree. A product is found when one set's key names a product in the cache; a miss names the entries
// whose values differ from the ones its newest set recorded.
const keptSets = 4

// readsVersion versions the read set file and every key derived from it. v4: a product read is valued by its bytes
// (productDigest), where v3 valued it by its key.
const readsVersion = "buildcache reads v4"

type readsFile struct {
	Version string    `json:"version"`
	Name    string    `json:"name"`
	Sets    []readSet `json:"sets"`
}

type readSet struct {
	Recorded string      `json:"recorded"`
	Key      string      `json:"key"`
	Trace    string      `json:"trace"`
	Entries  []readEntry `json:"entries"`
	Drift    *drift      `json:"drift,omitempty"`
	// Lean is a set keyed by the lean recipe (leanRecipeOf, #earlycut), so a lookup under Loom's switch refuses it
	// (leanKeysOff): while a lean hole is open, no lean key is matched, built or found.
	Lean bool `json:"lean,omitempty"`
}

// A readEntry is one thing the build read, by kind: content, listing, exists or link (paths in the tree), above (paths
// above the tree, relative to it), above cache (paths beside the build cache's parents, relative to it), home (paths in the home directory, ~-relative), absolute (any other path only looked
// for, by whether and as what it exists), product (a name key), tool (a path
// under the tools directory, exec'd), toolset (a tools directory read but not run), go (the release), system (a
// machine file, by the dpkg package that installed it unchanged or by itself), filesystem (case: whether the file system
// is case-sensitive, which osvfs asks as it starts) or settings (the build settings of the binary that ran the build). Value is what it was when recorded, shortened, so a miss can
// name what changed; the key hashes the full value.
type readEntry struct {
	Kind  string `json:"kind"`
	Path  string `json:"path"`
	Value string `json:"value,omitempty"`
}

// drift is the declared Files held against the measured reads: the tripwire Kirk kept until traces earn trust.
// Undeclared reads are what today's declared key would miss (a stale product); declared and unread is what it keys
// for nothing.
type drift struct {
	Undeclared      int      `json:"undeclared"`
	UndeclaredPaths []string `json:"undeclaredPaths,omitempty"`
	Unread          int      `json:"unread"`
	UnreadPaths     []string `json:"unreadPaths,omitempty"`
}

// NameKey is the part of a product's address that isn't a file: its name, flags and tool reports as portable
// spells them. Its read sets are filed under it.
func NameKey(root string, inputs Inputs) string {
	hash := sha256.New()
	field := func(kind, value string) {
		fmt.Fprintf(hash, "%s %d\n%s\n", kind, len(value), value)
	}
	field("buildcache", readsVersion)
	field("name", portable(root, inputs.Name))
	for _, flag := range inputs.Flags {
		field("flag", portable(root, flag))
	}
	for _, tool := range inputs.Toolchain {
		field("tool", portable(root, tool))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func readsPath(cache, nameKey string) string {
	return filepath.Join(cache, nameKey+".reads")
}

// loadReads is nameKey's read sets where this process writes them (cache), and, when that is an output directory of
// its own, those the build cache holds after them (location.ReadPlaces): the union, the output directory's sets first,
// then each of the cache's that reads other entries. withReads writes the whole union back to the output directory, so
// what Loom copies into the cache holds every set either held.
func loadReads(cache, nameKey string) (readsFile, error) {
	var union readsFile
	for index, path := range location.ReadPlaces(readsPath(cache, nameKey)) {
		file, err := readsAt(path)
		if err != nil {
			return readsFile{}, err
		}
		if index == 0 {
			union = file
		} else {
			union = unionReads(union, file)
		}
	}
	return union, nil
}

// readsAt is the read sets file at path: none when it isn't there or is another version's.
func readsAt(path string) (readsFile, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return readsFile{}, nil
	}
	if err != nil {
		return readsFile{}, err
	}
	var file readsFile
	if err = json.Unmarshal(content, &file); err != nil {
		return readsFile{}, fmt.Errorf("%s: %v", path, err)
	}
	if file.Version != readsVersion {
		return readsFile{}, nil
	}
	return file, nil
}

// unionReads is newer's sets, then each of older's that reads entries none of newer's does: a set of the same entries
// in both is the same set, valued later in newer, as recordReads keeps it.
func unionReads(newer, older readsFile) readsFile {
	switch {
	case older.Version == "":
		return newer
	case newer.Version == "":
		return older
	}
	union := readsFile{Version: newer.Version, Name: cmp.Or(newer.Name, older.Name), Sets: slices.Clone(newer.Sets)}
	for _, set := range older.Sets {
		if !slices.ContainsFunc(newer.Sets, func(kept readSet) bool { return sameEntries(kept.Entries, set.Entries) }) {
			union.Sets = append(union.Sets, set)
		}
	}
	return union
}

// recordReads puts set first among nameKey's sets, replacing one that reads the same entries.
func recordReads(cache, nameKey, name string, set readSet) error {
	return withReads(cache, nameKey, func(file *readsFile) error {
		file.Name = name
		file.Sets = append([]readSet{set}, slices.DeleteFunc(file.Sets, func(kept readSet) bool { return sameEntries(kept.Entries, set.Entries) })...)
		return nil
	})
}

// withReads changes nameKey's read sets under a lock every writer on the machine takes, so two settlements of one
// product never lose each other's reads: change reads the file, changes it, and it is written whole by rename, the
// newest keptSets sets kept. cache is where this process writes: the lock, the temporary file and the file itself are
// all there, and the file read is loadReads' union with what the build cache holds.
func withReads(cache, nameKey string, change func(file *readsFile) error) error {
	lock, err := os.OpenFile(readsPath(cache, nameKey)+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	file, err := loadReads(cache, nameKey)
	if err != nil {
		return err
	}
	if err = change(&file); err != nil {
		return err
	}
	file.Version = readsVersion
	file.Sets = file.Sets[:min(len(file.Sets), keptSets)]
	encoded, err := json.MarshalIndent(file, "", "\t")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(cache, ".reads-")
	if err != nil {
		return err
	}
	if _, err = temporary.Write(append(encoded, '\n')); err == nil {
		err = temporary.Close()
	}
	if err != nil {
		os.Remove(temporary.Name())
		return err
	}
	return os.Rename(temporary.Name(), readsPath(cache, nameKey))
}

// sortedEntries is entries without repeats, in the order every machine reads them: by path, then kind.
func sortedEntries(lists ...[]readEntry) []readEntry {
	seen := map[readEntry]bool{}
	var all []readEntry
	for _, entries := range lists {
		for _, entry := range entries {
			bare := readEntry{Kind: entry.Kind, Path: entry.Path}
			if !seen[bare] {
				seen[bare] = true
				all = append(all, bare)
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Path != all[j].Path {
			return all[i].Path < all[j].Path
		}
		return all[i].Kind < all[j].Kind
	})
	return all
}

func sameEntries(one, other []readEntry) bool {
	return slices.EqualFunc(one, other, func(a, b readEntry) bool { return a.Kind == b.Kind && a.Path == b.Path })
}

// evaluation is a read set valued on the tree at hand: its key, and each entry's value in order.
type evaluation struct {
	key    string
	values []string
}

// evaluate values every entry of set on the tree at root and hashes them under nameKey. A product entry is valued
// at that product's key on the same tree, which needs one of its sets to name a product in the cache; one that
// can't is an error naming it, so the product reading it can't be keyed either.
func evaluate(root, cache, nameKey string, set readSet, depth int, known productMemo) (evaluation, error) {
	if known == nil {
		known = productMemo{}
	}
	if depth > 16 {
		return evaluation{}, fmt.Errorf("products read each other more than 16 deep at %s", nameKey[:12])
	}
	hash := sha256.New()
	field := func(kind, value string) {
		fmt.Fprintf(hash, "%s %d\n%s\n", kind, len(value), value)
	}
	field("buildcache", readsVersion)
	field("name key", nameKey)
	result := evaluation{values: make([]string, len(set.Entries))}
	for index, entry := range set.Entries {
		value, err := entryValue(root, cache, entry, depth, known)
		if err != nil {
			return evaluation{}, err
		}
		result.values[index] = value
		field(entry.Kind+" "+entry.Path, value)
	}
	result.key = hex.EncodeToString(hash.Sum(nil))
	return result, nil
}

func entryValue(root, cache string, entry readEntry, depth int, known productMemo) (string, error) {
	switch entry.Kind {
	case "content", "listing", "exists":
		if !local(entry.Path) {
			return "", fmt.Errorf("a tree entry %q isn't inside the tree", entry.Path)
		}
		return treeValue(root, entry.Kind, entry.Path), nil
	case "link":
		// A symbolic link in the tree, read on the way to what it names: valued by where it points.
		if !local(entry.Path) {
			return "", fmt.Errorf("a tree entry %q isn't inside the tree", entry.Path)
		}
		path := filepath.Join(root, filepath.FromSlash(entry.Path))
		target, err := os.Readlink(path)
		if err != nil || !Tracked(root, path, false) {
			return treeValue(root, "exists", entry.Path), nil
		}
		return "link " + target, nil
	case nodeInstallKind:
		if !local(entry.Path) {
			return "", fmt.Errorf("a node install %q isn't inside the tree", entry.Path)
		}
		return treeValue(root, nodeInstallKind, entry.Path), nil
	case "above":
		return existence(filepath.Join(root, filepath.FromSlash(entry.Path))), nil
	case "above cache":
		// Relative to the build cache, where the product is read once Loom copies what this process wrote into it, so
		// a product built in an output directory of its own keys as one built in the cache.
		places := location.ReadPlaces(cache)
		return existence(filepath.Join(places[len(places)-1], filepath.FromSlash(entry.Path))), nil
	case "absolute":
		if !filepath.IsAbs(entry.Path) {
			return "", fmt.Errorf("an absolute entry %q isn't an absolute path", entry.Path)
		}
		return existence(entry.Path), nil
	case "home":
		home, err := os.UserHomeDir()
		if err != nil {
			return "no home: " + err.Error(), nil
		}
		return existence(filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(entry.Path, "~/")))), nil
	case "product":
		found, err := productFor(root, cache, entry.Path, depth+1, known)
		if err != nil {
			return "", err
		}
		if found.key == "" {
			return "", fmt.Errorf("it reads product %s, which has no product for this tree (%s)", entry.Path[:12], found.why)
		}
		// By the bytes it read, not by the key they're under: another trace settling the same product records the union
		// of both read sets, which moves its key and not its bytes, and a reader keyed by the key went stale with it
		// (home-7's dry run, Oct 11: 6 products, typeaware volume and the two that read it among them).
		return productDigest(found.directory)
	case productKeyKind:
		// A test read set's product, by the key it's found under: its set holds every read of the window the test's set
		// leaves out, so the key is what moves with them (#mmpt615).
		found, err := productFor(root, cache, entry.Path, depth+1, known)
		if err != nil {
			return "", err
		}
		if found.key == "" {
			return "", fmt.Errorf("it used product %s, which has no product for this tree (%s)", entry.Path[:12], found.why)
		}
		return found.key, nil
	case "tool":
		tools, err := toolsRoot()
		if err != nil {
			return "", err
		}
		return Tool(filepath.Join(tools, filepath.FromSlash(entry.Path)), "--version"), nil
	case "toolset":
		// A tools directory read but never run (a sysroot, a resource directory): named by the release it resolves to.
		tools, err := toolsRoot()
		if err != nil {
			return "", err
		}
		if real := resolved(filepath.Join(tools, entry.Path)); real != "" {
			return filepath.Base(real), nil
		}
		return "absent", nil
	case "go":
		return Tool("go", "version"), nil
	case "system":
		if !filepath.IsAbs(entry.Path) {
			return "", fmt.Errorf("a system entry %q isn't an absolute path", entry.Path)
		}
		return systemValue(entry.Path), nil
	case "system directory":
		if !filepath.IsAbs(entry.Path) {
			return "", fmt.Errorf("a system directory entry %q isn't an absolute path", entry.Path)
		}
		return existence(entry.Path), nil
	case "settings":
		return binarySettings(), nil
	case "filesystem":
		if entry.Path != "case" {
			return "", fmt.Errorf("a file system entry %q names nothing a machine answers", entry.Path)
		}
		return caseSensitivity(), nil
	}
	return "", fmt.Errorf("a read set entry of unknown kind %q", entry.Kind)
}

// settledTree holds each tree entry's value while Settle runs, by root, kind and path. Settle values every product it
// settles and every product each one reads, each in a pass of its own, so a family of products that read one another
// and the same directories valued each listing again in every pass (home-5's dry run, Oct 11: css's PostCSS answers,
// up to 40 sibling products and 3,047 listings each, settled for over 20 minutes). The tree is the same from Settle's
// start to its end, or every build is refused (TreeState), so a value read once holds for all of it. Nil outside
// Settle and Find: a builder, or a test that changes the tree between lookups, values afresh.
var settledTree atomic.Pointer[sync.Map]

// readTree holds each tree entry's value for the life of a process whose tree can't change under it (heldForLife), as
// settledTree does for Settle's: a test that looks up several products valued every entry of each one's read set again
// (#szswjtc: TypeScript's readproofs, 8 to 27 s a hit on a loaded Workshop). It is its own map, so Settle and Find,
// which set settledTree and clear it as they end, never drop what such a process holds.
var readTree sync.Map

// readMode is ADAMIC_BUILD_CACHE=read, a runner's process: it reads products and never builds them, over a source
// that never changes under it.
func readMode() bool {
	return os.Getenv("ADAMIC_BUILD_CACHE") == "read"
}

// heldForLife says whether what this process values of the tree at root, and each product it finds there, holds for
// its life: in read mode, whose source never changes under it; under a trace, whose tree is the same from the run's
// start to its end or every build of the run is refused (TreeState before and after, and any write into the tree,
// Settle); and over an unpacked source (unpackedSource), which Loom assembles once from the tree's chunks and nothing
// changes after. A traced process that looked up each product afresh valued the whole chain each one reads every time:
// css's PostCSS answers read the products their process read before them, about 30 siblings each, so one hit valued
// 163,866 entries under strace and took 62 to 182 s, and a shard of 49 never finished inside the 14-minute bound (lap 3,
// #47ykkmr). Loom's key check runs a package's product tests over an unpacked source untraced and not in read mode, and
// css's ran into its 10-minute timeout the same way (lap 4, #y7c0t4q).
func heldForLife(root string) bool {
	return readMode() || traceDirectory() != "" || unpackedSource(root)
}

// unpackedSources holds, by root, whether a root is an unpacked source.
var unpackedSources sync.Map

// unpackedSource says whether root is a source Loom unpacked: no .git, and git's answers in its .tracked manifest
// (manifestListing), as a runner's source and the key check's stand-in for one are. Looking is keying, as Tracked's
// look at .git is: no read set may hold .git or the manifest.
func unpackedSource(root string) bool {
	if found, ok := unpackedSources.Load(root); ok {
		return found.(bool)
	}
	unpacked := false
	keying(func() {
		_, gitErr := statGit(filepath.Join(root, ".git"))
		info, manifestErr := os.Stat(filepath.Join(root, trackedManifestDirectory, "files"))
		unpacked = gitErr != nil && manifestErr == nil && info.Mode().IsRegular()
	})
	unpackedSources.Store(root, unpacked)
	return unpacked
}

// treeValuesRead counts the tree values this process read rather than held: a test asks it.
var treeValuesRead atomic.Int64

// productDigests holds each product directory's digest once a process has it: a key's bytes never change.
var productDigests sync.Map

// productDigest is a product's bytes as one value: each file under directory by its path, kind, executable bit and
// content (a link by where it points), kept beside it as <directory>.digest so each machine hashes a product once. A
// directory reached through a link is valued by what it links to (ProductRoot). The digest is read beside the product
// in the output directory, then in the build cache, and kept in the output directory (location.ReadPlaces).
func productDigest(directory string) (string, error) {
	if value, ok := productDigests.Load(directory); ok {
		return value.(string), nil
	}
	places := location.ReadPlaces(directory + ".digest")
	for _, kept := range places {
		if content, err := os.ReadFile(kept); err == nil && len(content) == 64 {
			productDigests.Store(directory, string(content))
			return string(content), nil
		}
	}
	root, err := ProductRoot(directory)
	if err != nil {
		return "", fmt.Errorf("digest of product %s: %w", filepath.Base(directory), err)
	}
	hash := sha256.New()
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			fmt.Fprintf(hash, "directory %q\n", relative)
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			fmt.Fprintf(hash, "link %q %q\n", relative, target)
		case info.Mode().IsRegular():
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(content)
			fmt.Fprintf(hash, "file %q %t %x\n", relative, info.Mode()&0o111 != 0, sum)
		default:
			fmt.Fprintf(hash, "other %q %s\n", relative, info.Mode().Type())
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("digest of product %s: %w", filepath.Base(directory), err)
	}
	value := hex.EncodeToString(hash.Sum(nil))
	productDigests.Store(directory, value)
	// Kept where this process writes, through a temporary file whose name starts with a dot, as every scratch name in
	// the cache does, so nothing copying the output directory takes it for a digest. Keeping it is a convenience: a
	// failure to is no error.
	kept := places[0]
	if kept != directory+".digest" {
		if err := os.MkdirAll(filepath.Dir(kept), 0o755); err != nil {
			return value, nil
		}
	}
	if temporary, err := os.CreateTemp(filepath.Dir(kept), ".digest-*"); err == nil {
		_, writeErr := temporary.WriteString(value)
		if closeErr := temporary.Close(); writeErr != nil || closeErr != nil || os.Rename(temporary.Name(), kept) != nil {
			os.Remove(temporary.Name())
		}
	}
	return value, nil
}

// treeValue is a tree path as the tree carries it: a path git doesn't track (in a checkout; an unpacked source tracks
// everything) reads as absent, so a checkout's leftovers and an unpacked source of it value the same.
func treeValue(root, kind, name string) string {
	held := settledTree.Load()
	if heldForLife(root) {
		held = &readTree
	}
	if held == nil {
		return readTreeValue(root, kind, name)
	}
	key := root + "\x00" + kind + "\x00" + name
	if value, ok := held.Load(key); ok {
		return value.(string)
	}
	value := readTreeValue(root, kind, name)
	held.Store(key, value)
	return value
}

func readTreeValue(root, kind, name string) string {
	treeValuesRead.Add(1)
	if kind == nodeInstallKind {
		return nodeInstallValue(filepath.Join(root, filepath.FromSlash(name), "node_modules"))
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	info, err := os.Stat(path)
	if err != nil || name != "." && !Tracked(root, path, info.IsDir()) {
		return "absent"
	}
	switch {
	case info.IsDir() && kind != "exists":
		return "directory " + listing(root, path)
	case info.IsDir():
		return "directory"
	case kind == "content" && info.Mode().IsRegular():
		return fileValue(root, path, info)
	case info.Mode().IsRegular():
		return fmt.Sprintf("file %t", info.Mode()&0o111 != 0)
	}
	return "other " + info.Mode().Type().String()
}

// nodeInstallValue is an installed node_modules at top as its manifest names it: each file's path, sha256 and executable
// bit, and each link's target, sorted, hashed whole (#qx8fkaf). It's what's there, not what the lockfile meant, so
// any byte npm made, any platform or install script that made another, and a file no lockfile names each move it. A
// tree holds Loom's copy of the install and a runner the chunk it unpacked, the same entries, so the two value alike
// at any checkout path, and no file the install holds stands in for its bytes.
func nodeInstallValue(top string) string {
	if real, err := filepath.EvalSymlinks(top); err == nil {
		top = real
	}
	var lines []string
	err := filepath.WalkDir(top, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || path == top {
			return err
		}
		relative := filepath.ToSlash(strings.TrimPrefix(path, top+string(filepath.Separator)))
		switch {
		case entry.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			lines = append(lines, "link "+relative+" "+target)
		case entry.Type().IsRegular():
			info, err := entry.Info()
			if err != nil {
				return err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			lines = append(lines, fmt.Sprintf("file %s %x %t", relative, sha256.Sum256(content), info.Mode()&0o111 != 0))
		}
		return nil
	})
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "absent"
	case err != nil:
		return "unreadable: " + err.Error()
	}
	sort.Strings(lines)
	return fmt.Sprintf("install %x", sha256.Sum256([]byte(strings.Join(lines, "\n"))))
}

// fileValue is a file's content entry in a read set: gitValue, or why the file couldn't be read.
func fileValue(root, path string, info fs.FileInfo) string {
	value, err := gitValue(root, path, info)
	if err != nil {
		return "unreadable: " + err.Error()
	}
	return "file " + value
}

// gitValue is a regular file's content as git names it: whether it is executable, and its blob id. A checkout's clean
// file is valued by the id its index holds, unread, so keying tens of thousands of files costs git's stat pass, not a
// hash of each, and an unpacked source's file by the id its manifest's listing holds; a dirty one, one changed since
// git was asked or the listing was written, or a file in a tree with neither git nor a manifest is hashed as git hashes
// a blob, so a checkout and an unpacked source of it value the same. A read set's content entry (fileValue) and a
// declared file (hashPath) are both valued here.
func gitValue(root, path string, info fs.FileInfo) (string, error) {
	object, length, clean := cleanObject(root, path, info)
	if clean {
		return fmt.Sprintf("%t %s", object.mode == "100755", object.id), nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%t %s", info.Mode()&0o111 != 0, blobID(content, length)), nil
}

// cleanObject is the index's object for a file whose working copy git found clean, unchanged since: modified, or its
// inode changed, a second before git was asked or later, it is hashed instead (git's own racy-clean margin covers what
// came before). In an unpacked source it is the manifest's object, under the same margin before the listing was
// written. The inode's change time (ctime) is asked as well as the modification time because a test can edit a tracked
// file and set its modification time back (os.Chtimes), which a modification time alone reads as unchanged, and
// nothing but the clock sets a ctime: the edit and the Chtimes each set it to now (Operations' scout, #szswjtc). A
// file whose ctime can't be read is hashed.
func cleanObject(root, name string, info fs.FileInfo) (gitObject, int, bool) {
	for at := filepath.Dir(name); inside(root, at) || at == root; at = filepath.Dir(at) {
		files := repositoryAt(at)
		if files == nil {
			if at == root {
				break
			}
			continue
		}
		length := 40
		for _, any := range files.objects {
			length = len(any.id)
			break
		}
		relative, err := filepath.Rel(at, name)
		if err != nil || files.all {
			return gitObject{}, length, false
		}
		relative = filepath.ToSlash(relative)
		object, ok := files.objects[relative]
		margin := files.taken.Add(-time.Second)
		changed, known := changeTime(info)
		if !ok || files.dirty[relative] || !strings.HasPrefix(object.mode, "100") || !info.ModTime().Before(margin) || !known || !changed.Before(margin) {
			return gitObject{}, length, false
		}
		return object, length, true
	}
	return gitObject{}, 40, false
}

// blobID is git's id of a blob: sha1 (or sha256, for a repository whose ids are 64 long) of "blob <size>\x00" and
// the content.
func blobID(content []byte, length int) string {
	header := fmt.Sprintf("blob %d\x00", len(content))
	if length == 64 {
		sum := sha256.Sum256(append([]byte(header), content...))
		return hex.EncodeToString(sum[:])
	}
	sum := sha1.Sum(append([]byte(header), content...))
	return hex.EncodeToString(sum[:])
}

// listing is a directory's tracked entries, by name and kind, sorted; .git never counts.
func listing(root, directory string) string {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "unreadable: " + err.Error()
	}
	var names []string
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		info, err := os.Stat(path)
		isDirectory := err == nil && info.IsDir()
		if entry.Name() == ".git" || !Tracked(root, path, isDirectory) {
			continue
		}
		suffix := ""
		if isDirectory {
			suffix = "/"
		}
		names = append(names, entry.Name()+suffix)
	}
	sort.Strings(names)
	sum := sha256.Sum256([]byte(strings.Join(names, "\n")))
	return hex.EncodeToString(sum[:])
}

func local(name string) bool {
	return !filepath.IsAbs(name) && name != ".." && !strings.HasPrefix(name, "../")
}

// existence is whether a path outside the tree exists and as what, never its content.
func existence(path string) string {
	info, err := os.Stat(path)
	switch {
	case err != nil:
		return "absent"
	case info.IsDir():
		return "directory"
	case info.Mode().IsRegular():
		return fmt.Sprintf("file %t", info.Mode()&0o111 != 0)
	}
	return "other " + info.Mode().Type().String()
}

// caseSensitivity is the file system's answer to osvfs's probe, asked as osvfs asks it: whether this process's own
// executable exists with every letter's case swapped.
func caseSensitivity() string {
	executable, err := os.Executable()
	if err != nil {
		return "unknown: " + err.Error()
	}
	swapped := swapCase(executable)
	switch _, err := os.Stat(swapped); {
	case swapped == executable:
		return "unknown: its executable's path has no letters"
	case err == nil:
		return "insensitive"
	case errors.Is(err, fs.ErrNotExist):
		return "sensitive"
	default:
		return "unknown: " + err.Error()
	}
}

// toolsRoot is where Workshop's tools live: ADAMIC_TOOLS, or ~/adamic-tools. A tool entry names a path under it, so
// the same tools under another home value the same. In a sandbox (ADAMIC_SANDBOX=1) it is ADAMIC_TOOLS or an error:
// a tool entry never resolves under a home the sandbox didn't declare.
func toolsRoot() (string, error) {
	if directory := os.Getenv("ADAMIC_TOOLS"); directory != "" {
		return directory, nil
	}
	if toolchain.Sandboxed() {
		return "", errors.New("ADAMIC_SANDBOX=1 finds tools only where ADAMIC_TOOLS says, and ADAMIC_TOOLS is unset")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/opt/adamic-tools", nil
	}
	return filepath.Join(home, "adamic-tools"), nil
}

// found is a product looked up by its read sets: the key and directory of the first set (newest first) whose key
// names a product in the cache, or why none did, and the keys each set gave (for the store to try). With the product,
// its upstreams: every product the set that found it reads, through their sets in turn; with each key, the same for
// the set that gave it (keyUpstreams). With each, the set that gave it as it values on this tree, when it stored another
// key (refound, keyRefound): the set the caller re-records once it has the product (rerecord).
type found struct {
	key, directory, why string
	keys                []string
	upstreams           []upstream
	keyUpstreams        [][]upstream
	refound             *readSet
	keyRefound          []*readSet
}

// An upstream is a product a found product's set reads, by name key and the directory this process found it in. A run
// that finds a product journals its upstreams as found too, so it names every product its lookup valued: a runner
// values the same set and needs each one's read set and product (#mmpt615, 2e95d29e's void: tsprinter's mutant
// corpora were found whole, their Go oracle was never named, and no runner could value them).
type upstream struct {
	nameKey, directory string
}

// setUpstreams is the upstreams of set, which evaluate valued already on this pass (known), each once, in its entries'
// order, each followed by its own.
func setUpstreams(root, cache string, set readSet, depth int, known productMemo) []upstream {
	var all []upstream
	seen := map[string]bool{}
	for _, entry := range set.Entries {
		if entry.Kind != "product" && entry.Kind != productKeyKind {
			continue
		}
		read, err := productFor(root, cache, entry.Path, depth+1, known)
		if err != nil || read.directory == "" {
			continue
		}
		for _, one := range append([]upstream{{nameKey: entry.Path, directory: read.directory}}, read.upstreams...) {
			if !seen[one.nameKey] {
				seen[one.nameKey] = true
				all = append(all, one)
			}
		}
	}
	return all
}

// journalFound journals the product nameKey (name, when known) as found in directory, and each of its upstreams.
func journalFound(nameKey, name, directory string, upstreams []upstream) {
	journal(journalEntry{Event: "found", NameKey: nameKey, Name: name, Directory: directory})
	for _, one := range upstreams {
		journal(journalEntry{Event: "found", NameKey: one.nameKey, Directory: one.directory})
	}
}

// A productMemo holds each product one pass has looked up, by name key (Home's dry run, Oct 11: nine traces settled for
// over two hours, each product read valuing every product it reads again, for every reader, ten deep). A pass is one
// top-level productFor or evaluate, over one tree and one cache, so no answer outlives what it was given; a read-mode
// process keeps what it found for its life besides (readProducts).
type productMemo map[string]*productLookup

// readProducts holds each product a process whose tree can't change under it (heldForLife) has found, by root, cache and
// name key, so a test that looks up several products, or one product many times, values each one's read set once
// (#szswjtc, #47ykkmr). Only a found product is held: its tree never changes under the process and its directory is
// never taken away, so the set that found it still finds it. A missing one is looked up again, since a product that
// arrives later is found. What a traced build reads is never changed by it: a lookup is keying, which no read set takes
// (trace.go).
var readProducts sync.Map

// productsLookedUp counts the products this process looked up by their read sets rather than held: a test asks it.
var productsLookedUp atomic.Int64

// A productLookup is one product's answer in a pass; done is false while it is still being looked up, so meeting it
// again then is a cycle.
type productLookup struct {
	done   bool
	result found
	err    error
}

// productFor is the product nameKey names on the tree at root, once per pass (known: nil starts one). A product read
// while it is still being looked up reads itself through others, a cycle no key can value, and fails at once rather
// than at depth 16 down every branch.
func productFor(root, cache, nameKey string, depth int, known productMemo) (found, error) {
	held := heldForLife(root)
	// By every directory the lookup reads, so a process that reads a build cache through an output directory of its own
	// and one that reads it alone never share an answer.
	heldKey := root + "\x00" + strings.Join(location.ReadPlaces(cache), "\x00") + "\x00" + nameKey
	if held {
		if result, ok := readProducts.Load(heldKey); ok {
			return result.(found), nil
		}
	}
	if known == nil {
		known = productMemo{}
	}
	if lookup, seen := known[nameKey]; seen {
		if !lookup.done {
			return found{}, fmt.Errorf("products read each other in a cycle at %s", nameKey[:min(12, len(nameKey))])
		}
		return lookup.result, lookup.err
	}
	lookup := &productLookup{}
	known[nameKey] = lookup
	lookup.result, lookup.err = lookupProduct(root, cache, nameKey, depth, known)
	lookup.done = true
	if held && lookup.err == nil && lookup.result.directory != "" {
		readProducts.Store(heldKey, lookup.result)
	}
	return lookup.result, lookup.err
}

func lookupProduct(root, cache, nameKey string, depth int, known productMemo) (found, error) {
	productsLookedUp.Add(1)
	file, err := loadReads(cache, nameKey)
	if err != nil {
		return found{}, err
	}
	if len(file.Sets) == 0 {
		return found{why: "no traced build has recorded what it reads"}, nil
	}
	var result found
	var newest evaluation
	for index, set := range file.Sets {
		if set.Lean && leanKeysOff() {
			if index == 0 {
				result.why = "lean keys are off (ADAMIC_BUILD_LEAN=off), so its newest read set, keyed by the lean recipe, isn't matched"
			}
			continue
		}
		evaluated, err := evaluate(root, cache, nameKey, set, depth, known)
		if err != nil {
			if index == 0 {
				result.why = err.Error()
			}
			continue
		}
		if index == 0 {
			newest = evaluated
		}
		upstreams := setUpstreams(root, cache, set, depth, known)
		refound := revalued(set, evaluated)
		result.keys, result.keyUpstreams = append(result.keys, evaluated.key), append(result.keyUpstreams, upstreams)
		result.keyRefound = append(result.keyRefound, refound)
		// In the output directory first, then in the build cache: the product is wherever it was found.
		if directory, ok := existingPlace(filepath.Join(cache, evaluated.key)); ok {
			result.key, result.directory, result.upstreams, result.refound = evaluated.key, directory, upstreams, refound
			return result, nil
		}
	}
	if result.why == "" {
		result.why = changed(file.Sets[0], newest)
	}
	return result, nil
}

// revalued is set as it values on this tree (evaluated), under the key it gave, when that isn't the key it stored: the
// set to record again if it finds or fetches a product there. Nil when it gave its own key, and in read mode, which
// never writes the cache's read sets.
func revalued(set readSet, evaluated evaluation) *readSet {
	if set.Key == evaluated.key || readMode() {
		return nil
	}
	again := readSet{Recorded: now(), Key: evaluated.key, Trace: set.Trace, Drift: set.Drift, Entries: make([]readEntry, len(set.Entries))}
	for index, entry := range set.Entries {
		again.Entries[index] = readEntry{Kind: entry.Kind, Path: entry.Path, Value: shortValue(evaluated.values[index])}
	}
	return &again
}

// rerecorded holds each name key and key this process has recorded again (rerecord), so a product looked up many times
// writes its file once.
var rerecorded sync.Map

// rerecord records again, first among nameKey's sets, the set that found or fetched its product under a key it didn't
// store (#xaz1487, system_adamic's yes on #2dkn3fb): each set stores what it valued to when it was recorded, so a product
// found by one that values otherwise now sits under a key no set in the file names, and once newer sets push that set
// past the file's keptSets, nothing finds the product any more. A failure to write is noted and never fails the
// lookup.
func rerecord(cache, nameKey, name string, set *readSet) {
	if set == nil {
		return
	}
	if _, done := rerecorded.LoadOrStore(cache+"\x00"+nameKey+"\x00"+set.Key, true); done {
		return
	}
	if err := recordReads(cache, nameKey, name, *set); err != nil {
		note("reads %s %s: recording the set that found %s again: %v", name, nameKey[:min(12, len(nameKey))], set.Key[:12], err)
	}
}

// differing is the entries of a set whose values on this tree differ from the ones it recorded.
func differing(set readSet, now evaluation) []string {
	var differ []string
	for index, entry := range set.Entries {
		if index >= len(now.values) || entry.Value == "" || shortValue(now.values[index]) != entry.Value {
			differ = append(differ, entry.Kind+" "+entry.Path)
		}
	}
	return differ
}

// changed names the entries of a set whose values on this tree differ from the ones it recorded: how many of each
// kind, then the first 8 by kind and path. The counts are every entry's, so a reader of the miss sees the kind of the
// ones past the eighth too: Loom's judge voids a miss as the box's only when every differing entry is a system file,
// and an unnamed tree file must not hide behind eight named system ones (Judge, #m5z9d6q). The shape is
// "N of M entries it read differ from its newest traced build (by kind: content 2, system 18): <entries>[, and K more]",
// kinds in byte order, each named as readEntry's Kind is.
func changed(set readSet, now evaluation) string {
	differ := differing(set, now)
	if len(differ) == 0 {
		return fmt.Sprintf("its newest read set (%d entries) values as recorded, but no product is cached under %s", len(set.Entries), now.key[:min(12, len(now.key))])
	}
	counts := map[string]int{}
	for _, entry := range differ {
		kind, _, _ := strings.Cut(entry, " ")
		counts[kind]++
	}
	var kinds []string
	for _, kind := range slices.Sorted(maps.Keys(counts)) {
		kinds = append(kinds, fmt.Sprintf("%s %d", kind, counts[kind]))
	}
	named, more := differ, ""
	if len(differ) > 8 {
		named, more = differ[:8], fmt.Sprintf(", and %d more", len(differ)-8)
	}
	return fmt.Sprintf("%d of %d entries it read differ from its newest traced build (by kind: %s): %s%s", len(differ), len(set.Entries), strings.Join(kinds, ", "), strings.Join(named, ", "), more)
}

// shortValue is what a set records of a value: enough to tell two apart for a person, never the key's input.
func shortValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
