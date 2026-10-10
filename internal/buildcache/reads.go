package buildcache

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"
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

// readsVersion versions the read set file and every key derived from it.
const readsVersion = "buildcache reads v2"

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
}

// A readEntry is one thing the build read, by kind: content, listing, exists or link (paths in the tree), above (paths
// above the tree, relative to it), home (paths in the home directory, ~-relative), product (a name key), tool (a path
// under the tools directory, exec'd), toolset (a tools directory read but not run), go (the release), system (a
// machine file, by the dpkg package that installed it unchanged or by itself) or settings (the build settings of the
// binary that ran the build). Value is what it was when recorded, shortened, so a miss can
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

func loadReads(cache, nameKey string) (readsFile, error) {
	content, err := os.ReadFile(readsPath(cache, nameKey))
	if errors.Is(err, fs.ErrNotExist) {
		return readsFile{}, nil
	}
	if err != nil {
		return readsFile{}, err
	}
	var file readsFile
	if err = json.Unmarshal(content, &file); err != nil {
		return readsFile{}, fmt.Errorf("%s: %v", readsPath(cache, nameKey), err)
	}
	if file.Version != readsVersion {
		return readsFile{}, nil
	}
	return file, nil
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
// newest keptSets sets kept.
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

// mergeReads adds sets fetched from the store to nameKey's, after the cache's own, skipping any it already holds.
func mergeReads(cache, nameKey, name string, sets []readSet) error {
	return withReads(cache, nameKey, func(file *readsFile) error {
		file.Name = name
		for _, set := range sets {
			if !slices.ContainsFunc(file.Sets, func(kept readSet) bool { return kept.Key == set.Key && sameEntries(kept.Entries, set.Entries) }) {
				file.Sets = append(file.Sets, set)
			}
		}
		return nil
	})
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
func evaluate(root, cache, nameKey string, set readSet, depth int) (evaluation, error) {
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
		value, err := entryValue(root, cache, entry, depth)
		if err != nil {
			return evaluation{}, err
		}
		result.values[index] = value
		field(entry.Kind+" "+entry.Path, value)
	}
	result.key = hex.EncodeToString(hash.Sum(nil))
	return result, nil
}

func entryValue(root, cache string, entry readEntry, depth int) (string, error) {
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
	case "above":
		return existence(filepath.Join(root, filepath.FromSlash(entry.Path))), nil
	case "home":
		home, err := os.UserHomeDir()
		if err != nil {
			return "no home: " + err.Error(), nil
		}
		return existence(filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(entry.Path, "~/")))), nil
	case "product":
		found, err := productFor(root, cache, entry.Path, depth+1)
		if err != nil {
			return "", err
		}
		if found.key == "" {
			return "", fmt.Errorf("it reads product %s, which has no product for this tree (%s)", entry.Path[:12], found.why)
		}
		return found.key, nil
	case "tool":
		return Tool(filepath.Join(toolsRoot(), filepath.FromSlash(entry.Path)), "--version"), nil
	case "toolset":
		// A tools directory read but never run (a sysroot, a resource directory): named by the release it resolves to.
		if real := resolved(filepath.Join(toolsRoot(), entry.Path)); real != "" {
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
	case "settings":
		return binarySettings(), nil
	}
	return "", fmt.Errorf("a read set entry of unknown kind %q", entry.Kind)
}

// treeValue is a tree path as the tree carries it: a path git doesn't track (in a checkout; an unpacked source tracks
// everything) reads as absent, so a checkout's leftovers and an unpacked source of it value the same.
func treeValue(root, kind, name string) string {
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

// fileValue is a tracked file's content as git names it: its blob id and whether it is executable. A checkout's clean
// file is valued by the id its index holds, unread, so keying tens of thousands of files costs git's stat pass, not a
// hash of each; a dirty one, one changed since git was asked, or a file in a tree with no git is hashed as git hashes
// a blob, so a checkout and an unpacked source of it value the same.
func fileValue(root, path string, info fs.FileInfo) string {
	object, length, clean := cleanObject(root, path, info)
	if clean {
		return fmt.Sprintf("file %t %s", object.mode == "100755", object.id)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "unreadable: " + err.Error()
	}
	return fmt.Sprintf("file %t %s", info.Mode()&0o111 != 0, blobID(content, length))
}

// cleanObject is the index's object for a file whose working copy git found clean, unchanged since: modified a second
// before git was asked or later, it is hashed instead (git's own racy-clean margin covers what came before).
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
		if !ok || files.dirty[relative] || !strings.HasPrefix(object.mode, "100") || !info.ModTime().Before(files.taken.Add(-time.Second)) {
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

// toolsRoot is where Workshop's tools live: ADAMIC_TOOLS, or ~/adamic-tools. A tool entry names a path under it, so
// the same tools under another home value the same.
func toolsRoot() string {
	if directory := os.Getenv("ADAMIC_TOOLS"); directory != "" {
		return directory
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/opt/adamic-tools"
	}
	return filepath.Join(home, "adamic-tools")
}

// found is a product looked up by its read sets: the key and directory of the first set (newest first) whose key
// names a product in the cache, or why none did, and the keys each set gave (for the store to try).
type found struct {
	key, directory, why string
	keys                []string
}

func productFor(root, cache, nameKey string, depth int) (found, error) {
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
		evaluated, err := evaluate(root, cache, nameKey, set, depth)
		if err != nil {
			if index == 0 {
				result.why = err.Error()
			}
			continue
		}
		if index == 0 {
			newest = evaluated
		}
		result.keys = append(result.keys, evaluated.key)
		if _, err := os.Stat(filepath.Join(cache, evaluated.key)); err == nil {
			result.key, result.directory = evaluated.key, filepath.Join(cache, evaluated.key)
			return result, nil
		}
	}
	if result.why == "" {
		result.why = changed(file.Sets[0], newest)
	}
	return result, nil
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

// changed names the entries of a set whose values on this tree differ from the ones it recorded.
func changed(set readSet, now evaluation) string {
	differ := differing(set, now)
	switch {
	case len(differ) == 0:
		return fmt.Sprintf("its newest read set (%d entries) values as recorded, but no product is cached under %s", len(set.Entries), now.key[:min(12, len(now.key))])
	case len(differ) > 8:
		return fmt.Sprintf("%d of %d entries it read differ from its newest traced build: %s, and %d more", len(differ), len(set.Entries), strings.Join(differ[:8], ", "), len(differ)-8)
	}
	return fmt.Sprintf("%d of %d entries it read differ from its newest traced build: %s", len(differ), len(set.Entries), strings.Join(differ, ", "))
}

// shortValue is what a set records of a value: enough to tell two apart for a person, never the key's input.
func shortValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
