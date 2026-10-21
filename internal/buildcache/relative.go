package buildcache

import (
	"bytes"
	"encoding/json"
	"errors"
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
// where it ends up. named is set to the name key of each product its bytes name (relocatable), so get can record where
// each was for a product it builds for this machine alone (recordParts).
func relocating(inputs Inputs, nameKey string, build func(directory string) error, named *[]string) func(directory string) error {
	return func(directory string) error {
		located(nameKey, directory)
		err := build(directory)
		if err == nil {
			keying(func() { *named, err = relocatable(inputs, directory) })
		}
		forgotten(nameKey, directory)
		return err
	}
}

// productDirectory is where the product named nameKey is on this machine: the directory this process holds it in;
// else the one its read sets name in the cache (settled, as Loom's runner unpacks it), the one this trace built and
// hasn't settled (pending), or the one built here for this machine alone (local), in that order; else <cache>/<name>,
// as a settled key names itself, when it is there. Each is looked for in the output directory first, then in the build
// cache (location.ReadPlaces). What it finds is held, and journaled as found, so a build reading there under a trace
// reads that product (trace.go). Looking it up reads read sets, the tree and declared files, all under keying.
//
// It never answers with a place that isn't there (#z8mg5ks: on Home, markdown-list-layout-shared-v4 named a part whose
// declared files had moved, <cache>/<name key> was read back, and 16 tests ran a missing binary, charged to the change).
// A product found nowhere panics with productPlace's ErrNotBuilt error, and the panic is deliberate: Absolute, the make
// file's hook that calls it, answers with a string, and answering with an error would re-key every product to fix a
// path Loom's runners never take, since they read settled products by sum. The panic reaches the test that read the
// name back and fails it with adamic's not-built line as the whole of its message, which go test writes as that test's
// output, so Loom's judge reads it as that miss, the tree's, never the change's (#z8mg5ks; a panic in TestMain, which
// fails no test, is #rf855c8's).
func productDirectory(nameKey string) string {
	directory, err := productPlace(nameKey)
	if err != nil {
		panic(err)
	}
	return directory
}

// productPlace is productDirectory's place for nameKey, or an ErrNotBuilt error naming the product when it has none.
func productPlace(nameKey string) (string, error) {
	return productPlaceChecking(nameKey, map[string]bool{})
}

// productPlaceChecking is productPlace while checking holds the local products whose parts are being checked further up
// this lookup (partsHoldChecking).
func productPlaceChecking(nameKey string, checking map[string]bool) (string, error) {
	if directory, ok := productKeys.Load(nameKey); ok {
		return directory.(string), nil
	}
	cache, err := cacheLocation()
	if err != nil {
		return "", notBuiltPart(cache, nameKey, err.Error())
	}
	output, err := outputLocation()
	if err != nil {
		return "", notBuiltPart(cache, nameKey, err.Error())
	}
	var directory string
	var upstreams []upstream
	if root, err := repositoryRoot(); err == nil {
		keying(func() { directory, upstreams = namedProductChecking(root, output, nameKey, checking) })
	}
	if directory == "" {
		var settled string
		var ok bool
		keying(func() { settled, ok = existingPlace(filepath.Join(output, nameKey)) })
		if !ok {
			return "", notBuiltPart(output, nameKey, fmt.Sprintf("no read set, pending build or product built for this machine "+
				"alone finds it, and %s holds nothing", strings.Join(location.ReadPlaces(filepath.Join(output, nameKey)), " nor ")))
		}
		return settled, nil
	}
	journalFound(nameKey, "", directory, upstreams)
	return holding(nameKey, directory), nil
}

// notBuiltPart is the ErrNotBuilt error for a product read back by its name key alone, in get's words for a missing
// product, which Loom's judge matches (`product (.+?) \(name key ([0-9a-f]{12})\) is not built: `). The product is
// named as its local pointer or its read sets name it, when either does, else by the name it was read back by.
func notBuiltPart(cache, nameKey, why string) error {
	name := "<build cache>/" + nameKey
	keying(func() {
		if inputs, ok := localInputs(cache, nameKey); ok && inputs.Name != "" {
			name = inputs.Name
		} else if file, err := loadReads(cache, nameKey); err == nil && file.Name != "" {
			name = file.Name
		}
	})
	return fmt.Errorf("product %s (name key %s) is %w: %s", name, nameKey[:min(12, len(nameKey))], ErrNotBuilt, why)
}

// namedProduct is productDirectory's place for nameKey, with the upstreams of the set that found it there, if a set did.
// cache is where this process writes products, and each place is looked for there and then in the build cache.
func namedProduct(root, cache, nameKey string) (string, []upstream) {
	return namedProductChecking(root, cache, nameKey, map[string]bool{})
}

// namedProductChecking is namedProduct while checking holds the local products whose parts are being checked further up
// this lookup (partsHoldChecking).
func namedProductChecking(root, cache, nameKey string, checking map[string]bool) (string, []upstream) {
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
			// Only one whose parts are still where they were when it was built (partsHold, #z8mg5ks).
			for _, place := range location.ReadPlaces(filepath.Join(cache, "local", key)) {
				if isDirectory(place) && partsHoldChecking(place, checking) {
					return place, nil
				}
			}
		}
	}
	return "", nil
}

// partsPath is where a product built for this machine alone records its parts, beside it: the products its bytes name
// by <build cache>/<name key>, each with the place, relative to the cache, it was at when the product was built
// (local/<key>, <key> for a settled set's, pending/<trace>/<name key>), as JSON, name key to place. Loom ships only a
// settled product's <key>/ and <key>.inputs, and a runner never builds a local product, so the record never leaves
// the machine.
func partsPath(directory string) string {
	return directory + ".parts"
}

// recordParts writes target's parts record, whole by rename, for a product just built and not yet placed at target:
// each of named but nameKey (the product's own) with where productPlace finds it now (#z8mg5ks). A product that
// names none records an empty one. A part found nowhere, or outside the cache (an uncached build's directory), leaves
// no record, so the product misses next time and is built again rather than hit with a part no record can check.
func recordParts(target, nameKey string, named []string) error {
	parts := map[string]string{}
	for _, part := range named {
		if part == nameKey {
			continue
		}
		directory, err := productPlace(part)
		if err == nil {
			if place, ok := cachePlace(directory); ok {
				parts[part] = place
				continue
			}
		}
		note("parts %s: %s is in no cache place (%v), so it is built again next time", nameKey[:12], part[:12], err)
		if err = os.Remove(partsPath(target)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	encoded, err := json.Marshal(parts)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".parts-")
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
	return os.Rename(temporary.Name(), partsPath(target))
}

// partsHold says whether the product built for this machine alone at directory is still the one its parts make: its
// parts record is there, and every part in it is found now, as productDirectory finds it, at exactly the place it was
// recorded at, a directory that is there. A part found elsewhere moved (its declared files changed, so it keys
// elsewhere), and the product holding its old name is stale; a product with no record was built before records were
// kept. Either misses, and is built again (#z8mg5ks).
func partsHold(directory string) bool {
	return partsHoldChecking(directory, map[string]bool{})
}

// partsHoldChecking is partsHold while checking holds the local products whose parts are being checked further up this
// lookup. A product met again while its own check is under way names itself through its parts, a cycle no record can
// settle: it doesn't hold, so the products in the cycle miss and are built again, rather than the check recursing
// until the stack overflows.
func partsHoldChecking(directory string, checking map[string]bool) bool {
	if checking[directory] {
		return false
	}
	checking[directory] = true
	defer delete(checking, directory)
	content, err := os.ReadFile(partsPath(directory))
	if err != nil {
		return false
	}
	var parts map[string]string
	if err = json.Unmarshal(content, &parts); err != nil {
		return false
	}
	for part, recorded := range parts {
		found, err := productPlaceChecking(part, checking)
		if err != nil || !isDirectory(found) {
			return false
		}
		if place, ok := cachePlace(found); !ok || place != recorded {
			return false
		}
	}
	return true
}

// cachePlace is directory relative to the root that holds it (location.Roots), slash-separated, so a place recorded
// under the output directory is the same place once Loom copies it into the build cache. A directory under no root has
// none.
func cachePlace(directory string) (string, bool) {
	roots, err := location.Roots()
	if err != nil {
		return "", false
	}
	for _, root := range roots {
		relative, err := filepath.Rel(root, directory)
		if err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return filepath.ToSlash(relative), true
		}
	}
	return "", false
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
//
// It returns the name key of each product the files it reads name by <build cache>/<name key>, once each, in the order
// first met: the product's parts, whose places a product built for this machine alone records (recordParts).
func relocatable(inputs Inputs, directory string) ([]string, error) {
	places := machinePlaces()
	vouching := declaredText(inputs.Files)
	var named []string
	err := readFiles(directory, true, func(file string, content []byte, _ fs.FileMode) error {
		text := string(content)
		for _, name := range productName.FindAllString(text, -1) {
			if nameKey := strings.TrimPrefix(name, "<build cache>/"); !slices.Contains(named, nameKey) {
				named = append(named, nameKey)
			}
		}
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
	return named, err
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
