package buildcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	"syscall"

	"github.com/system-inc/adamic/internal/buildcache/location"
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
	asKeying(func() {
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
		return productDirectoryOf(strings.TrimPrefix(name, "<build cache>/"))
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
	asKeying(func() {
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

// The machinery's part in Relative, Absolute and Resolved comes in through the hooks below, each set in an init of the
// file that owns it. Alone, each is a plain default, so nothing in this file keys, traces, finds or stores a product.

// asKeying runs work as keying: the machinery sets it (trace.go's init) so a trace never records where this machine
// keeps things; alone it just runs work.
var asKeying = func(work func()) { work() }

// productDirectoryOf is where the product named nameKey is on this machine: the machinery sets it (relative.go's init)
// to productDirectory, which looks it up by read sets and journals what it finds; alone it is <cache>/<name key>, as a
// settled key names itself.
var productDirectoryOf = func(nameKey string) string {
	cache, _ := cacheLocation()
	return filepath.Join(cache, nameKey)
}

// isProductDirectory says whether directory is a product: the machinery sets it (relative.go's init) to isProduct,
// which knows the cache's layout and a trace's pending directory; alone it is a directory this process holds.
var isProductDirectory = func(directory string) bool {
	_, ok := productDirectories.Load(directory)
	return ok
}

// foundProduct records that reading directory is reading the product named nameKey: the machinery sets it (trace.go's
// init) to journal a found entry; alone it records nothing.
var foundProduct = func(nameKey, directory string) {}

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
// tree for its own steps (a mutant port whose imports reach back into the repository, which the build then lowers) and
// ends by making them a product's: the last step of its build, before Product scans it. A reader that hands those files
// to a tool reads them through Resolved.
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
// text files are read back through Absolute, made once per machine and roots, or the product itself when no file holds a
// name. The copy is never a product; the product itself stays as it was built, and a key that names a file in the copy
// names it in the product, <build cache>/<name key>/... (places). A directory that isn't a product (a package's own,
// which a caller may pass alike) is read as it is, and nothing is written beside it.
func Resolved(directory string) (string, error) {
	if found, ok := resolvedDirectories.Load(directory); ok {
		return found.(string), nil
	}
	if !isProductDirectory(directory) {
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
		foundProduct(nameKey, place)
	}
	resolvedCopies.Store(place, nameKey)
	resolvedDirectories.Store(directory, place)
	return place, nil
}

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

type place struct{ path, name string }

func resolved(path string) string {
	resolutions.Add(1)
	if value, err := filepath.EvalSymlinks(path); err == nil {
		return value
	}
	return ""
}

// resolutions counts the paths this process has resolved: a test asks it.
var resolutions atomic.Int64

// replacePath replaces path with name where it stands as a whole path: not inside a longer name (/work/adamic in
// /work/adamic-bench stays) and not as the tail of another path (/root in /chroot stays).
func replacePath(value, path, name string) string {
	var built strings.Builder
	for {
		index := strings.Index(value, path)
		if index < 0 {
			built.WriteString(value)
			return built.String()
		}
		end := index + len(path)
		before := index == 0 || !pathByte(value[index-1]) && value[index-1] != '/'
		after := end == len(value) || value[end] == '/' || !pathByte(value[end])
		built.WriteString(value[:index])
		if before && after {
			built.WriteString(name)
		} else {
			built.WriteString(path)
		}
		value = value[end:]
	}
}

func pathByte(b byte) bool {
	return 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || '0' <= b && b <= '9' || strings.IndexByte("._-+@~", b) >= 0
}

// cacheLocation is where products live on this machine (location.BuildCache), the one rule internal/native's caches
// live under too.
func cacheLocation() (string, error) {
	return location.BuildCache()
}

var root struct {
	once      sync.Once
	directory string
	err       error
}

// repositoryRoot is the nearest directory above the working directory holding this module's go.mod.
func repositoryRoot() (string, error) {
	root.once.Do(func() {
		directory, err := os.Getwd()
		if err != nil {
			root.err = err
			return
		}
		for {
			content, err := os.ReadFile(filepath.Join(directory, "go.mod"))
			if err == nil && strings.HasPrefix(string(content), "module github.com/system-inc/adamic\n") {
				root.directory = directory
				return
			}
			parent := filepath.Dir(directory)
			if parent == directory {
				root.err = errors.New("buildcache: no adamic go.mod above the working directory")
				return
			}
			directory = parent
		}
	})
	return root.directory, root.err
}

// ProductRoot is where a walk over a product directory starts: what directory resolves to when it is a link, and every
// walk rooted at a product goes through it, a test's included (graphql/printer's printerWalk, mega's run 2). Loom's runner
// links each product a unit fetched into the unit's cache from the box's shared products, and filepath.WalkDir visits a
// link at its root as one entry, never what it points to. Mega's run 1 (Oct 12): every product read through another
// was valued as one link line naming that box's path, and missed on every runner; run 2: Resolved found no text file
// in estree's mutant port, so node read its <repository> imports raw. Links inside a product are its own, and stay.
func ProductRoot(directory string) (string, error) {
	if info, err := os.Lstat(directory); err != nil || info.Mode()&fs.ModeSymlink == 0 {
		return directory, nil
	}
	return filepath.EvalSymlinks(directory)
}
