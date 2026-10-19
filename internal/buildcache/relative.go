package buildcache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/system-inc/adamic/internal/buildcache/location"
)

// Absolute finds a product it reads back, and Resolved tells a product from another directory, as the cache lays them
// out, through these hooks (make_relative.go).
func init() {
	productDirectoryOf = productDirectory
	isProductDirectory = isProduct
}

func located(nameKey, directory string) {
	if previous, held := productDirectories.Swap(directory, nameKey); !held || previous != nameKey {
		productGeneration.Add(1)
	}
	productKeys.Store(nameKey, directory)
}

// forgotten drops directory from what this process holds.
func forgotten(nameKey, directory string) {
	if _, held := productDirectories.LoadAndDelete(directory); held {
		productGeneration.Add(1)
	}
	productKeys.CompareAndDelete(nameKey, directory)
}

// holding is directory, a product get returns, recorded under its name key so Relative names it and Absolute finds it.
func holding(nameKey, directory string) string {
	located(nameKey, directory)
	return directory
}

// relocating wraps a product's build. While it builds, its directory is <build cache>/<name key> to Relative; when it
// has built, relocatable reads what it made, under keying so the check's reads are never the build's own or a later
// build's (trace.go). The scratch directory is forgotten either way: filled renames it, and get holds the product
// where it ends up.
func relocating(inputs Inputs, nameKey string, build func(directory string) error) func(directory string) error {
	return func(directory string) error {
		located(nameKey, directory)
		err := build(directory)
		if err == nil {
			keying(func() { err = relocatable(inputs, directory) })
		}
		forgotten(nameKey, directory)
		return err
	}
}

// productDirectory is where the product named nameKey is on this machine: the directory this process holds it in;
// else the one its read sets name in the cache (settled, as Loom's runner unpacks it), the one this trace built and
// hasn't settled (pending), or the one built here for this machine alone (local), in that order; else <cache>/<name>,
// as a settled key names itself. Each is looked for in the output directory first, then in the build cache
// (location.ReadPlaces). What it finds is held, and journaled as found, so a build reading there under a trace reads
// that product (trace.go). Looking it up reads read sets, the tree and declared files, all under keying.
func productDirectory(nameKey string) string {
	if directory, ok := productKeys.Load(nameKey); ok {
		return directory.(string)
	}
	cache, err := cacheLocation()
	if err != nil {
		return filepath.Join(cache, nameKey)
	}
	output, err := outputLocation()
	if err != nil {
		return filepath.Join(cache, nameKey)
	}
	var directory string
	var upstreams []upstream
	if root, err := repositoryRoot(); err == nil {
		keying(func() { directory, upstreams = namedProduct(root, output, nameKey) })
	}
	if directory == "" {
		settled := filepath.Join(cache, nameKey)
		if output != cache {
			keying(func() {
				if found, ok := existingPlace(filepath.Join(output, nameKey)); ok {
					settled = found
				}
			})
		}
		return settled
	}
	journalFound(nameKey, "", directory, upstreams)
	return holding(nameKey, directory)
}

// namedProduct is productDirectory's place for nameKey, with the upstreams of the set that found it there, if a set did.
// cache is where this process writes products, and each place is looked for there and then in the build cache.
func namedProduct(root, cache, nameKey string) (string, []upstream) {
	if looked, err := productFor(root, cache, nameKey, 0, nil); err == nil && looked.directory != "" {
		return looked.directory, looked.upstreams
	}
	if traceDirectory() != "" {
		if pending, ok := directoryPlace(filepath.Join(pendingDirectory(cache), nameKey)); ok {
			return pending, nil
		}
	}
	if inputs, ok := localInputs(cache, nameKey); ok {
		if key, err := Key(root, inputs); err == nil {
			if built, ok := directoryPlace(filepath.Join(cache, "local", key)); ok {
				return built, nil
			}
		}
	}
	return "", nil
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// directoryPlace is the first place path is a directory at, under the output directory or the build cache
// (location.ReadPlaces), the output directory's first.
func directoryPlace(path string) (string, bool) {
	for _, place := range location.ReadPlaces(path) {
		if isDirectory(place) {
			return place, true
		}
	}
	return "", false
}

// localPointer is where a product built for this machine alone records the inputs that key it, by name key: its
// declared key needs its Files hashed on the tree at hand, which a reader holding only the name key can't name.
func localPointer(cache, nameKey string) string {
	return filepath.Join(cache, "local", nameKey+".json")
}

// pointLocal records inputs under nameKey, whole by rename, when they aren't there already. They are recorded as
// portable spells them, which is how they key: a flag naming this checkout (repository=/path/to/it) must key the same
// when another checkout of the tree reads the pointer, and Key reads a portable value back unchanged.
func pointLocal(cache, nameKey string, inputs Inputs) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	inputs.Name = portable(root, inputs.Name)
	inputs.Flags = slices.Clone(inputs.Flags)
	for index, flag := range inputs.Flags {
		inputs.Flags[index] = portable(root, flag)
	}
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	for index, tool := range inputs.Toolchain {
		inputs.Toolchain[index] = portable(root, tool)
	}
	encoded, err := json.Marshal(inputs)
	if err != nil {
		return err
	}
	path := localPointer(cache, nameKey)
	// Written where this process writes, unless the output directory or the build cache already holds the same pointer.
	for _, place := range location.ReadPlaces(path) {
		if existing, err := os.ReadFile(place); err == nil && bytes.Equal(existing, encoded) {
			return nil
		}
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".pointer-")
	if err != nil {
		return err
	}
	if _, err = temporary.Write(encoded); err == nil {
		err = temporary.Close()
	}
	if err == nil {
		err = os.Chmod(temporary.Name(), 0o644)
	}
	if err != nil {
		os.Remove(temporary.Name())
		return err
	}
	return os.Rename(temporary.Name(), path)
}

// localInputs are the inputs nameKey's pointer records, read in the output directory first, then in the build cache.
func localInputs(cache, nameKey string) (Inputs, bool) {
	for _, place := range location.ReadPlaces(localPointer(cache, nameKey)) {
		content, err := os.ReadFile(place)
		if err != nil {
			continue
		}
		var inputs Inputs
		return inputs, json.Unmarshal(content, &inputs) == nil
	}
	return Inputs{}, false
}

// relocatable refuses a product whose text names a path of the machine that built it: the repository, the build
// cache (so any product in it, and this one's scratch directory), a product built uncached, the home directory, the
// temporary directory, or a location an ADAMIC_ variable names (ADAMIC_TYPESCRIPT_SOURCE). A file holding a zero byte
// is a binary, whose paths are debug information no test reads (native prefix maps are their own work), and so is a
// macOS .dSYM bundle, unless it is data a test reads back (dataFormats); every other file is text a reader parses, and
// must say <repository> and <build cache>/<name key>.
//
// One path is not the builder's: a whole path token that occurs, as the same whole token, in a tracked file among the
// product's own declared inputs (its Files, read on this tree) is that input's text, which the product carries (#arw7837:
// json's pinned corpus is the repository's JSON, whose reports name /tmp/adamic-gate/... runs). It passes. A token naming
// the build cache or a product directory never does, since no input can vouch for where this machine builds.
func relocatable(inputs Inputs, directory string) error {
	places := machinePlaces()
	vouching := declaredText(inputs.Files)
	return readFiles(directory, true, func(file string, content []byte, _ fs.FileMode) error {
		text := string(content)
		// In data, the byte before a string is its length, any byte at all, so a path there starts wherever it is found.
		data := slices.Contains(dataFormats, filepath.Ext(file))
		for _, place := range places {
			for _, token := range pathTokens(text, place.path, data) {
				if vouching.holds(token, places) {
					continue
				}
				relative, _ := filepath.Rel(directory, file)
				return fmt.Errorf("product %s holds a path of the machine that built it: %s names %s (%s). A runner fetches this product and "+
					"reads it elsewhere, so a product names the tree as <repository> and a product as <build cache>/<name key> "+
					"(buildcache.Relative, read back with buildcache.Absolute), or keeps the path out of its bytes (#tqrqx60)",
					inputs.Name, filepath.ToSlash(relative), place.path, place.name)
			}
		}
		return nil
	})
}

// pathTokens are the whole path tokens of text in which path stands as a whole path, by replacePath's rule: each runs
// from where path starts to the last path byte or separator after it, so /tmp/adamic-gate/x/y.json is one token. In
// data (anyBefore), path may follow any byte.
func pathTokens(text, path string, anyBefore bool) []string {
	var tokens []string
	for offset := 0; ; {
		index := strings.Index(text[offset:], path)
		if index < 0 {
			return tokens
		}
		start := offset + index
		end := start + len(path)
		before := anyBefore || start == 0 || !pathByte(text[start-1]) && text[start-1] != '/'
		after := end == len(text) || text[end] == '/' || !pathByte(text[end])
		if before && after {
			for end < len(text) && (pathByte(text[end]) || text[end] == '/') {
				end++
			}
			tokens = append(tokens, text[start:end])
		}
		offset = start + len(path)
	}
}

// declaredVouching is the text of a product's declared, tracked inputs, read once and only when a token needs it.
type declaredVouching struct {
	files []string
	once  sync.Once
	texts [][]byte
	// tokens is every whole path token of texts: each maximal run of path bytes and separators.
	tokens map[string]bool
}

// vouchingScans counts the tokens holds looked for through every declared text: a test asks it.
var vouchingScans atomic.Int64

func declaredText(files []string) *declaredVouching {
	return &declaredVouching{files: files}
}

// holds says whether token occurs as a whole token (the same token, not a prefix of a longer path nor a part of one) in
// a tracked file among the declared inputs, and names no build cache or product directory.
func (vouching *declaredVouching) holds(token string, places []place) bool {
	for _, place := range places {
		if (place.name == "the build cache" || place.name == "a product directory") && containsPath(token, place.path) {
			return false
		}
	}
	vouching.once.Do(func() {
		root, err := repositoryRoot()
		if err != nil {
			return
		}
		for _, name := range vouching.files {
			start := filepath.Join(root, name)
			filepath.WalkDir(start, func(file string, entry fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if entry.IsDir() {
					if file != start && !Tracked(root, file, true) {
						return filepath.SkipDir
					}
					return nil
				}
				if entry.Type().IsRegular() && Tracked(root, file, false) {
					if content, err := os.ReadFile(file); err == nil {
						vouching.texts = append(vouching.texts, content)
					}
				}
				return nil
			})
		}
		vouching.tokens = map[string]bool{}
		for _, text := range vouching.texts {
			for start := 0; start < len(text); {
				if !pathByte(text[start]) && text[start] != '/' {
					start++
					continue
				}
				end := start
				for end < len(text) && (pathByte(text[end]) || text[end] == '/') {
					end++
				}
				vouching.tokens[string(text[start:end])] = true
				start = end
			}
		}
	})
	// A token of path bytes and separators alone stands whole in a text exactly when it is one of the text's maximal
	// runs of them: looked up, once the texts are read (json's pinned corpus, landable-14b's tree, Oct 11: every token
	// its product holds scanned against 819 declared files ran past go test's 10 minutes). Any other token is looked for.
	if !strings.ContainsFunc(token, func(r rune) bool { return r >= 0x80 || !pathByte(byte(r)) && r != '/' }) {
		return vouching.tokens[token]
	}
	vouchingScans.Add(1)
	for _, text := range vouching.texts {
		for _, found := range pathTokens(string(text), token, false) {
			if found == token {
				return true
			}
		}
	}
	return false
}

// isProduct says whether directory is a product: one this process holds, or a key's directory in the build cache or
// the output directory, in either's local directory or in this trace's pending one.
func isProduct(directory string) bool {
	if _, ok := productDirectories.Load(directory); ok {
		return true
	}
	roots, err := location.Roots()
	if err != nil || !productKey.MatchString(filepath.Base(directory)) {
		return false
	}
	parent := filepath.Dir(directory)
	var places []string
	for _, cache := range roots {
		places = append(places, cache, filepath.Join(cache, "local"))
		if traceDirectory() != "" {
			places = append(places, pendingDirectory(cache))
		}
	}
	return slices.ContainsFunc(places, func(place string) bool {
		return parent == place || resolved(parent) != "" && resolved(parent) == resolved(place)
	})
}

var productKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// machinePlaces are the paths relocatable looks for, longest first so a file under the repository is named as the
// repository's, not the home directory's.
func machinePlaces() []place {
	var found []place
	add := func(path, name string) {
		if !filepath.IsAbs(path) {
			return
		}
		for _, candidate := range []string{filepath.Clean(path), resolved(path)} {
			// / names every path, and a top-level directory such as /tmp is too short to tell a machine's own path from
			// one a product's text means (a lowered program's "/tmp/out").
			if candidate == "" || strings.Count(strings.Trim(candidate, string(filepath.Separator)), string(filepath.Separator)) < 1 {
				continue
			}
			if !slices.ContainsFunc(found, func(other place) bool { return other.path == candidate }) {
				found = append(found, place{candidate, name})
			}
		}
	}
	if root, err := repositoryRoot(); err == nil {
		add(root, "the repository")
	}
	cache, cacheErr := cacheLocation()
	if cacheErr == nil {
		add(cache, "the build cache")
	}
	// The output directory is the build cache's, whose products Loom copies into it: no product names it either.
	if output, err := outputLocation(); err == nil && (cacheErr != nil || output != cache) {
		add(output, "the build cache")
	}
	productDirectories.Range(func(directory, _ any) bool {
		add(directory.(string), "a product directory")
		return true
	})
	if home, err := os.UserHomeDir(); err == nil {
		add(home, "the home directory")
	}
	add(os.TempDir(), "the temporary directory")
	for name, directory := range locations() {
		add(directory, name)
	}
	sort.SliceStable(found, func(i, j int) bool { return len(found[i].path) > len(found[j].path) })
	return found
}

// containsPath says whether path stands in value as a whole path, by replacePath's rule.
func containsPath(value, path string) bool {
	return replacePath(value, path, "\x00") != value
}
