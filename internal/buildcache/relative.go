package buildcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	"syscall"
)

// A product's bytes mean the same on every machine (#tqrqx60). Workshop builds each product once and runners fetch it
// by a key that names no machine, so a manifest, an overlay or a mutant's import holding Workshop's /home/ahra/...
// would point nowhere on the runner, and the test reading it would fail as the author's red. A product names the tree
// as <repository>, the build cache as <build cache>, another product (or itself) as <build cache>/<name key>, and a
// gate input by the variable that locates it, such as <ADAMIC_TYPESCRIPT_SOURCE>; Relative writes those names and
// Absolute reads them back against the roots of the machine reading it. A product built here that still names this
// machine fails its build, naming the file (relocatable).
//
// A name key, not a key: a product's key is the read set its traced build measured, known only once Settle has keyed
// it, and it lives under a different directory while it builds (pending), once settled (<key>) and when built for this
// machine alone (local/<declared key>). Its name key is the same in all three, on every machine, so Absolute reads a
// name key back to wherever the reader's cache holds that product (productDirectory). Loom's runner unpacks each
// product with its <key>.inputs and <name key>.reads (#0gcg7vf's shape B), so that lookup needs no network.

// Relative rewrites this machine's paths in value to the roots a reader supplies: the repository becomes <repository>,
// a product directory <build cache>/<name key> (every product this process got, an uncached build's temporary
// directory too, and a product's own directory while it builds), the build cache <build cache>, and a directory an
// ADAMIC_ variable locates <ADAMIC_NAME>. A path
// outside them all stays as it is, and relocatable refuses a product that holds one of this machine's.
func Relative(value string) string {
	for _, place := range relativePlaces() {
		value = replacePath(value, place.path, place.name)
	}
	return value
}

// madePlaces is what Relative's places were made from, the product directories this process held (productGeneration)
// and its environment, and the places.
type madePlaces struct {
	generation  int64
	environment string
	places      []place
}

var relativeMade atomic.Pointer[madePlaces]

// relativePlaces are Relative's places, longest first, made again only when a product directory was located or
// forgotten or the environment changed, and made as keying. Making them stats every ADAMIC_ variable's value and
// resolves every place, a stat per path element, each one a stop under a trace, and a loader calls Relative for every
// path it displays: lint's profile lowering called it about 8,400 times, 2.3 million stats, 9.7 s untraced and 637 s
// traced (Home, Oct 12), so lint's four ProfileCompilation product tests ran past go test's 14 minutes under the trace
// and fell back untraced. Where this machine keeps things is never what a product read.
func relativePlaces() []place {
	generation, environment := productGeneration.Load(), strings.Join(os.Environ(), "\x00")
	if made := relativeMade.Load(); made != nil && made.generation == generation && made.environment == environment {
		return made.places
	}
	var found []place
	add := func(path, name string) {
		for _, candidate := range []string{filepath.Clean(path), resolved(path)} {
			if candidate != "" && candidate != string(filepath.Separator) {
				found = append(found, place{candidate, name})
			}
		}
	}
	keying(func() {
		productDirectories.Range(func(directory, key any) bool {
			add(directory.(string), "<build cache>/"+key.(string))
			return true
		})
		if cache, err := cacheLocation(); err == nil {
			add(cache, "<build cache>")
		}
		if root, err := repositoryRoot(); err == nil {
			add(root, "<repository>")
		}
		for name, directory := range locations() {
			add(directory, "<"+name+">")
		}
	})
	sort.SliceStable(found, func(i, j int) bool { return len(found[i].path) > len(found[j].path) })
	relativeMade.Store(&madePlaces{generation, environment, found})
	return found
}

// Absolute is Relative read back on this machine: <build cache>/<name key> is that product's directory here
// (productDirectory), <build cache> the build cache, <repository> the repository, and <ADAMIC_NAME> where this
// machine's ADAMIC_NAME locates it (left as it is when unset, so the read fails naming it).
func Absolute(value string) string {
	cache, _ := cacheLocation()
	value = productName.ReplaceAllStringFunc(value, func(name string) string {
		return productDirectory(strings.TrimPrefix(name, "<build cache>/"))
	})
	value = strings.ReplaceAll(value, "<build cache>", cache)
	if root, err := repositoryRoot(); err == nil {
		value = strings.ReplaceAll(value, "<repository>", root)
	}
	here := locations()
	return locationName.ReplaceAllStringFunc(value, func(name string) string {
		if directory, ok := here[strings.Trim(name, "<>")]; ok {
			return directory
		}
		return name
	})
}

var productName = regexp.MustCompile(`<build cache>/[0-9a-f]{64}`)
var locationName = regexp.MustCompile(`<ADAMIC_[A-Z0-9_]+>`)

// locations are the directories this machine's ADAMIC_ variables name (ADAMIC_TYPESCRIPT_SOURCE, a library's npm
// install), by variable, absolute. The build cache's own variable is <build cache>. They're looked for once per
// environment, as keying (relativePlaces): Absolute runs once per name it reads back, too.
func locations() map[string]string {
	environment := os.Environ()
	joined := strings.Join(environment, "\x00")
	if made := locationsMade.Load(); made != nil && made.environment == joined {
		return made.found
	}
	found := map[string]string{}
	keying(func() {
		for _, variable := range environment {
			name, value, _ := strings.Cut(variable, "=")
			if !strings.HasPrefix(name, "ADAMIC_") || name == "ADAMIC_BUILD_CACHE_DIR" || value == "" {
				continue
			}
			if info, err := statLocation(value); err == nil && info.IsDir() {
				if absolute, err := filepath.Abs(value); err == nil {
					found[name] = absolute
				}
			}
		}
	})
	locationsMade.Store(&madeLocations{joined, found})
	return found
}

// madeLocations is the environment locations were looked for in, and what they found. Nothing writes found once made.
type madeLocations struct {
	environment string
	found       map[string]string
}

var locationsMade atomic.Pointer[madeLocations]

// statLocation is os.Stat, for locations: a test asks whether it ran as keying.
var statLocation = os.Stat

// productDirectories maps each product directory this process holds (every one get returned, a build's scratch
// directory while it builds, one productDirectory found) to its name key, and productKeys each name key to the latest
// such directory. Only located and forgotten change them, so productGeneration moves with every change.
var productDirectories, productKeys sync.Map

// productGeneration counts the changes to productDirectories, so relativePlaces knows its places are still whole.
var productGeneration atomic.Int64

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
// as a settled key names itself. What it finds is held, and journaled as found, so a build reading there under a trace
// reads that product (trace.go). Looking it up reads read sets, the tree and declared files, all under keying.
func productDirectory(nameKey string) string {
	if directory, ok := productKeys.Load(nameKey); ok {
		return directory.(string)
	}
	cache, err := cacheLocation()
	if err != nil {
		return filepath.Join(cache, nameKey)
	}
	var directory string
	if root, err := repositoryRoot(); err == nil {
		keying(func() { directory = namedProduct(root, cache, nameKey) })
	}
	if directory == "" {
		return filepath.Join(cache, nameKey)
	}
	journal(journalEntry{Event: "found", NameKey: nameKey, Directory: directory})
	return holding(nameKey, directory)
}

func namedProduct(root, cache, nameKey string) string {
	if looked, err := productFor(root, cache, nameKey, 0, nil); err == nil && looked.directory != "" {
		return looked.directory
	}
	if traceDirectory() != "" {
		if pending := filepath.Join(pendingDirectory(cache), nameKey); isDirectory(pending) {
			return pending
		}
	}
	if inputs, ok := localInputs(cache, nameKey); ok {
		if key, err := Key(root, inputs); err == nil && isDirectory(filepath.Join(cache, "local", key)) {
			return filepath.Join(cache, "local", key)
		}
	}
	return ""
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
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
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, encoded) {
		return nil
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

func localInputs(cache, nameKey string) (Inputs, bool) {
	content, err := os.ReadFile(localPointer(cache, nameKey))
	if err != nil {
		return Inputs{}, false
	}
	var inputs Inputs
	return inputs, json.Unmarshal(content, &inputs) == nil
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

// textFiles calls each with every text file under directory: a regular file holding no zero byte in its first 8000
// (a binary's paths are debug information no test reads, and native prefix maps are their own work), outside a macOS
// .dSYM bundle.
func textFiles(directory string, each func(file string, content []byte, mode fs.FileMode) error) error {
	return readFiles(directory, false, each)
}

// dataFormats are binary formats a product writes as data a test reads back, never a binary's debug information, so
// relocatable reads their paths as it reads text's (landable-13b: estree's program.gob named its random build
// directory, and the zero-byte skip let it through to the key check). Nothing rewrites them: a name Relative writes
// changes a string's length, which gob records beside it.
var dataFormats = []string{".gob"}

// readFiles is textFiles, and with data, every file in one of dataFormats too. A directory reached through a link is
// read where it points (ProductRoot), and each file is named under directory as given.
func readFiles(directory string, data bool, each func(file string, content []byte, mode fs.FileMode) error) error {
	root, err := ProductRoot(directory)
	if err != nil {
		return err
	}
	return filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if relative, err := filepath.Rel(root, file); err == nil {
			file = filepath.Join(directory, relative)
		}
		if entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".dSYM") {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if bytes.IndexByte(content[:min(len(content), 8000)], 0) >= 0 && !(data && slices.Contains(dataFormats, filepath.Ext(file))) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return each(file, content, info.Mode())
	})
}

// RelativeFiles rewrites every text file under directory through Relative, for a build that writes files naming this
// tree for its own steps (a mutant port whose imports reach back into the repository, which the build then lowers)
// and ends by making them a product's: the last step of its build, before Product scans and publishes it. A reader
// that hands those files to a tool reads them through Resolved.
func RelativeFiles(directory string) error {
	return textFiles(directory, func(file string, content []byte, mode fs.FileMode) error {
		if relative := Relative(string(content)); relative != string(content) {
			return os.WriteFile(file, []byte(relative), mode.Perm())
		}
		return nil
	})
}

// resolvedDirectories holds each directory Resolved has answered in this process, and resolvedCopies each copy it made
// or found, by the name key of the product it copies, which is how a key names the copy (portable).
var resolvedDirectories, resolvedCopies sync.Map

// Resolved is a product's directory as this machine reads it, for a reader that hands its files to a tool that can't
// read Relative's names itself (node running a mutant port, the loader lowering it): a copy beside the product whose
// text files are read back through Absolute, made once per machine and roots, or the product itself when no file holds
// a name. The copy is never a product and never published; the product itself stays as it was built, and a key that
// names a file in the copy names it in the product, <build cache>/<name key>/... (places). A directory that
// isn't a product (a package's own, which a caller may pass alike) is read as it is, and nothing is written beside it.
func Resolved(directory string) (string, error) {
	if found, ok := resolvedDirectories.Load(directory); ok {
		return found.(string), nil
	}
	if !isProduct(directory) {
		resolvedDirectories.Store(directory, directory)
		return directory, nil
	}
	named := false
	err := textFiles(directory, func(_ string, content []byte, _ fs.FileMode) error {
		named = named || Absolute(string(content)) != string(content)
		return nil
	})
	if err != nil || !named {
		if err == nil {
			resolvedDirectories.Store(directory, directory)
		}
		return directory, err
	}
	// The copy depends on where this machine keeps what Absolute reads names against, so two checkouts sharing one
	// cache each get their own.
	roots := []string{Absolute("<repository>"), Absolute("<build cache>")}
	for name, location := range locations() {
		roots = append(roots, name+"="+location)
	}
	sort.Strings(roots)
	sum := sha256.Sum256([]byte(strings.Join(roots, "\n")))
	place := filepath.Join(filepath.Dir(directory), "."+filepath.Base(directory)+".resolved-"+hex.EncodeToString(sum[:6]))
	if _, err := os.Stat(place); err != nil {
		lock, err := os.OpenFile(place+".lock", os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return "", err
		}
		defer lock.Close()
		if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
			return "", err
		}
		defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		if _, err = os.Stat(place); err != nil {
			if err = resolve(directory, place); err != nil {
				return "", err
			}
		}
	}
	// Under a trace, reading the copy is reading the product (trace.go's found entries). A key names the copy as it names
	// the product, <build cache>/<name key> (places); a directory this process doesn't hold is named by its own base.
	nameKey := filepath.Base(directory)
	if held, ok := productDirectories.Load(directory); ok {
		nameKey = held.(string)
		journal(journalEntry{Event: "found", NameKey: nameKey, Directory: place})
	}
	resolvedCopies.Store(place, nameKey)
	resolvedDirectories.Store(directory, place)
	return place, nil
}

// isProduct says whether directory is a product: one this process holds, or a key's directory in the build cache,
// in its local directory or in this trace's pending one.
func isProduct(directory string) bool {
	if _, ok := productDirectories.Load(directory); ok {
		return true
	}
	cache, err := cacheLocation()
	if err != nil || !productKey.MatchString(filepath.Base(directory)) {
		return false
	}
	parent := filepath.Dir(directory)
	places := []string{cache, filepath.Join(cache, "local")}
	if traceDirectory() != "" {
		places = append(places, pendingDirectory(cache))
	}
	return slices.ContainsFunc(places, func(place string) bool {
		return parent == place || resolved(parent) != "" && resolved(parent) == resolved(place)
	})
}

var productKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// resolve copies directory to place whole, each text file through Absolute, by way of a scratch directory renamed into
// place, so a copy on disk is always complete.
func resolve(directory, place string) error {
	root, err := ProductRoot(directory)
	if err != nil {
		return err
	}
	scratch, err := os.MkdirTemp(filepath.Dir(place), ".resolving-")
	if err != nil {
		return err
	}
	err = filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		target := filepath.Join(scratch, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("%s is not a regular file", file)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if bytes.IndexByte(content[:min(len(content), 8000)], 0) < 0 && !strings.Contains(relative, ".dSYM"+string(filepath.Separator)) {
			content = []byte(Absolute(string(content)))
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
	if err == nil {
		err = os.Rename(scratch, place)
	}
	if err != nil {
		os.RemoveAll(scratch)
	}
	return err
}

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
	if cache, err := cacheLocation(); err == nil {
		add(cache, "the build cache")
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
