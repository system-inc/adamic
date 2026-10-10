package buildcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
)

// A product's bytes mean the same on every machine (#tqrqx60). Workshop builds each product once and runners fetch it
// by a key that names no machine, so a manifest, an overlay or a mutant's import holding Workshop's /home/ahra/...
// would point nowhere on the runner, and the test reading it would fail as the author's red. A product names the tree
// as <repository>, the build cache as <build cache>, another product (or itself) as <build cache>/<key>, and a gate
// input by the variable that locates it, such as <ADAMIC_TYPESCRIPT_SOURCE>; Relative
// writes those names and Absolute reads them back against the roots of the machine reading it. A product built here
// that still names this machine fails its build, naming the file (relocatable).

// Relative rewrites this machine's paths in value to the roots a reader supplies: the repository becomes <repository>,
// a product directory <build cache>/<key> (an uncached build's temporary directory too, and a product's own directory
// while it builds), the build cache <build cache>, and a directory an ADAMIC_ variable locates <ADAMIC_NAME>. A path
// outside them all stays as it is, and relocatable refuses a product that holds one of this machine's.
func Relative(value string) string {
	var found []place
	add := func(path, name string) {
		for _, candidate := range []string{filepath.Clean(path), resolved(path)} {
			if candidate != "" && candidate != string(filepath.Separator) {
				found = append(found, place{candidate, name})
			}
		}
	}
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
	sort.SliceStable(found, func(i, j int) bool { return len(found[i].path) > len(found[j].path) })
	for _, place := range found {
		value = replacePath(value, place.path, place.name)
	}
	return value
}

// Absolute is Relative read back on this machine: <build cache>/<key> is that product's directory (where this process
// built it uncached, or under the build cache), <build cache> the build cache, <repository> the repository, and
// <ADAMIC_NAME> where this machine's ADAMIC_NAME locates it (left as it is when unset, so the read fails naming it).
func Absolute(value string) string {
	cache, _ := cacheLocation()
	value = productName.ReplaceAllStringFunc(value, func(name string) string {
		key := strings.TrimPrefix(name, "<build cache>/")
		if directory, ok := productKeys.Load(key); ok {
			return directory.(string)
		}
		return filepath.Join(cache, key)
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
// install), by variable, absolute. The build cache's own variable is <build cache>.
func locations() map[string]string {
	found := map[string]string{}
	for _, variable := range os.Environ() {
		name, value, _ := strings.Cut(variable, "=")
		if !strings.HasPrefix(name, "ADAMIC_") || name == "ADAMIC_BUILD_CACHE_DIR" || value == "" {
			continue
		}
		if info, err := os.Stat(value); err == nil && info.IsDir() {
			if absolute, err := filepath.Abs(value); err == nil {
				found[name] = absolute
			}
		}
	}
	return found
}

// productDirectories maps each product directory this process builds outside the cache's own <key> (a build's scratch
// directory, an uncached build's temporary one) to its key, and productKeys each key to the latest such directory.
var productDirectories, productKeys sync.Map

func located(key, directory string) {
	productDirectories.Store(directory, key)
	productKeys.Store(key, directory)
}

// relocatable refuses a product whose text names a path of the machine that built it: the repository, the build
// cache (so any product in it, and this one's scratch directory), a product built uncached, the home directory, the
// temporary directory, or a location an ADAMIC_ variable names (ADAMIC_TYPESCRIPT_SOURCE). A file holding a zero byte
// is a binary, whose paths are debug information no test reads (native prefix maps are their own work), and so is a
// macOS .dSYM bundle; every other file is text a reader parses, and must say <repository> and <build cache>/<key>.
func relocatable(name, directory string) error {
	places := machinePlaces()
	return textFiles(directory, func(file string, content []byte, _ fs.FileMode) error {
		text := string(content)
		for _, place := range places {
			if containsPath(text, place.path) {
				relative, _ := filepath.Rel(directory, file)
				return fmt.Errorf("product %s holds a path of the machine that built it: %s names %s (%s). A runner fetches this product and "+
					"reads it elsewhere, so a product names the tree as <repository> and a product as <build cache>/<key> "+
					"(buildcache.Relative, read back with buildcache.Absolute), or keeps the path out of its bytes (#tqrqx60)",
					name, filepath.ToSlash(relative), place.path, place.name)
			}
		}
		return nil
	})
}

// textFiles calls each with every text file under directory: a regular file holding no zero byte in its first 8000
// (a binary's paths are debug information no test reads, and native prefix maps are their own work), outside a macOS
// .dSYM bundle.
func textFiles(directory string, each func(file string, content []byte, mode fs.FileMode) error) error {
	return filepath.WalkDir(directory, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
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
		if bytes.IndexByte(content[:min(len(content), 8000)], 0) >= 0 {
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
// or found, by the key of the product it copies, which is how a key names the copy (portable).
var resolvedDirectories, resolvedCopies sync.Map

// Resolved is a product's directory as this machine reads it, for a reader that hands its files to a tool that can't
// read Relative's names itself (node running a mutant port, the loader lowering it): a copy beside the product whose
// text files are read back through Absolute, made once per machine and roots, or the product itself when no file holds
// a name. The copy is never a product and never published; the product itself stays as it was built, and a key that
// names a file in the copy names it in the product, <build cache>/<key>/... (places). A directory that
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
	key := filepath.Base(directory)
	if built, ok := productDirectories.Load(directory); ok {
		key = built.(string)
	}
	resolvedCopies.Store(place, key)
	resolvedDirectories.Store(directory, place)
	return place, nil
}

// isProduct says whether directory is a product: a key's directory in the build cache, or one this process built
// uncached.
func isProduct(directory string) bool {
	if _, ok := productDirectories.Load(directory); ok {
		return true
	}
	cache, err := cacheLocation()
	if err != nil || !productKey.MatchString(filepath.Base(directory)) {
		return false
	}
	parent := filepath.Dir(directory)
	return parent == cache || resolved(parent) != "" && resolved(parent) == resolved(cache)
}

var productKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// resolve copies directory to place whole, each text file through Absolute, by way of a scratch directory renamed into
// place, so a copy on disk is always complete.
func resolve(directory, place string) error {
	scratch, err := os.MkdirTemp(filepath.Dir(place), ".resolving-")
	if err != nil {
		return err
	}
	err = filepath.WalkDir(directory, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(directory, file)
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
