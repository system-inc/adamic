package native

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// A program's main.c is keyed by what its compile reads, read just before clang compiles it (#2r9xhqs, the C bytes
// object cache): clang's build as a split unit's key reads it (unitClang), the environment clang reads, every argument of
// the compile with each per-build place written as a placeholder, every file the preprocessing read by its bytes, and
// main.c preprocessed, which holds what resolution decided of every header it reached. An object the cache holds under
// the key is copied to main.o and nothing compiles. A miss compiles, and after the compile the preprocessing runs again
// and every file it read is looked at again (objectKeying.recheck), so a header changed while main.c compiled is caught
// and its object is never stored.
//
// Preprocessed bytes keep the column of a line's first token but no other spacing and no comment, and a Debug or
// sanitized object's line table records each token's column: two sources a space apart inside a line preprocess to the
// same bytes and compile, under -g, to other objects (Golem, clang 20.1.8). So every file the preprocessing read is keyed
// by its own bytes too, which hold every column.

// objectTrustVariable is how far a build's objects are trusted, read once a process: main, a build of main's own, reads
// and writes the main directory alone; candidate reads main's, then its own, and writes its own alone, so nothing a
// candidate compiled is ever read by main. Unset or anything else, main.c is neither keyed nor cached. BuildLean, which
// reads no environment, never is.
const objectTrustVariable = "ADAMIC_NATIVE_OBJECT_TRUST"

// objectTrust is objectTrustVariable as this process reads it, once.
var objectTrust = sync.OnceValue(func() string { return os.Getenv(objectTrustVariable) })

// errUncachedObject is a reason main.c gets no key: its compile reads something the key can't account for.
var errUncachedObject = errors.New("native: main.c compiles uncached")

// objectPlaceholders are the per-build places a compile's prefix maps take out of its object, each by the name it is
// mapped to (linkOne's maps and clangIn's toolMaps), and the placeholder the key writes it as: the build directory, random
// per build, the runtime archive's directory, under each machine's cache, and where the compiler is installed. The argv
// names each of them in its map, and in a Debug build names the build directory in main.c's path and the -I.
var objectPlaceholders = map[string]string{"/adamic-build": "<build>", "/adamic-runtime": "<runtime>", "/adamic-compiler": "<compiler>"}

// An objectPlace is one spelling of a per-build place as a compile's -ffile-prefix-map gives it: what the map rewrites
// it to in what the object records, and what the key writes it as.
type objectPlace struct {
	spelling, mapped, placeholder string
}

// objectPlaces are the per-build places arguments map out of the object (objectPlaceholders), every spelling a map gives,
// in the order clang tries its macro prefix maps, a std::map sorted by std::greater: a spelling another begins with comes
// after it, so the longer one wins.
func objectPlaces(arguments []string) []objectPlace {
	var places []objectPlace
	for _, argument := range arguments {
		mapping, found := strings.CutPrefix(argument, "-ffile-prefix-map=")
		if !found {
			continue
		}
		// clang splits a map at its first =, so a spelling never holds one.
		spelling, mapped, _ := strings.Cut(mapping, "=")
		if placeholder, perBuild := objectPlaceholders[mapped]; perBuild && spelling != "" {
			places = append(places, objectPlace{spelling, mapped, placeholder})
		}
	}
	sort.Slice(places, func(i, j int) bool { return places[i].spelling > places[j].spelling })
	return places
}

// objectPlaceheld writes every spelling of places in a string as its placeholder.
func objectPlaceheld(places []objectPlace) *strings.Replacer {
	var pairs []string
	for _, place := range places {
		pairs = append(pairs, place.spelling, place.placeholder)
	}
	return strings.NewReplacer(pairs...)
}

// objectMarkersMapped is preprocessed with the path of each line marker rewritten as the compile's prefix maps rewrite
// a path in the object. clang 20 applies no prefix map to -E's line markers, in a release build or a Debug one (Golem,
// clang 20.1.8), only to __FILE__, whose expansion is mapped already. A marker is `# <line> "<path>"` and its flags, 1 to
// 4, at a line's start; only the start of its path is rewritten, its line and flags kept exactly. The cases:
//   - main.c: "main.c" in a release or sanitized build, untouched, and in a Debug build, which compiles it by its absolute
//     path, "<build directory>/main.c", rewritten to "/adamic-build/main.c". A Debug build's own #line resets name it
//     "main.c", untouched.
//   - The runtime's headers: "runtime/adamic.h", through the build directory's link, untouched, and in a Debug build,
//     which searches the link by its absolute path, "<build directory>/runtime/adamic.h", rewritten to
//     "/adamic-build/runtime/adamic.h". One under the runtime archive's own directory, which no include reaches today,
//     would be rewritten to /adamic-runtime.
//   - clang's own headers (stddef.h, stdbool.h) under its installation's resource directory, rewritten to
//     "/adamic-compiler/lib/clang/<version>/include/...".
//   - The C library's headers (/usr/include), "<built-in>", "<command line>", and the TypeScript files a Debug build's
//     #line directives name, untouched: none is a per-build place.
//
// A path clang escaped, holding a quote or a backslash, may match no spelling and stay as it is, which only keys apart.
func objectMarkersMapped(preprocessed []byte, places []objectPlace) []byte {
	var mapped bytes.Buffer
	mapped.Grow(len(preprocessed))
	for line := range bytes.Lines(preprocessed) {
		mapped.Write(objectMarkerMapped(line, places))
	}
	return mapped.Bytes()
}

// objectMarkerMapped is line with its path's start rewritten by the first of places it begins with, when line is a
// marker; any other line is itself.
func objectMarkerMapped(line []byte, places []objectPlace) []byte {
	rest, marker := bytes.CutPrefix(line, []byte("# "))
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if !marker || digits == 0 || !bytes.HasPrefix(rest[digits:], []byte(` "`)) {
		return line
	}
	start := len(line) - len(rest) + digits + 2
	for _, place := range places {
		if bytes.HasPrefix(line[start:], []byte(place.spelling)) {
			return slices.Concat(line[:start], []byte(place.mapped), line[start+len(place.spelling):])
		}
	}
	return line
}

// An objectRead is one file main.c's preprocessing read, as its dependency list names it, by the inode and change time
// it had once the preprocessing was done.
type objectRead struct {
	name    string
	inode   uint64
	changed int64
}

// objectReadOf is the file name names, from directory when it is relative, by its inode and change time now.
func objectReadOf(directory string, name string) (objectRead, error) {
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(directory, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return objectRead{}, err
	}
	inode, changed, ok := fileChange(info)
	if !ok {
		return objectRead{}, fmt.Errorf("%s's change time can't be read here", name)
	}
	return objectRead{name: name, inode: inode, changed: changed}, nil
}

// An objectKeying is main.c's key, read before its compile, and what the recheck after the compile reads again: the
// preprocessing, its bytes and every file it read.
type objectKeying struct {
	key string
	// compile is the compile keyed, whose directory, program and environment every preprocessing runs with.
	compile *exec.Cmd
	// preprocess are the preprocessing's arguments but its outputs: the compile's flags, -E and main.c.
	preprocess   []string
	preprocessed []byte
	reads        []objectRead
}

// objectKeyer is linkOne's compileObject for a build that reads the environment: the object cache for this process's
// trust (objectTrust), in the build cache's two object directories, or none where the trust is neither main nor
// candidate, or where the unit cache is off (unitCacheOff, its variable or a mismatch's mark), which turns this cache off
// with it.
func objectKeyer() func(compile *exec.Cmd, source string, compileFlags []string, program string, jobserver *Jobserver) error {
	main, err := cacheDirectory("native-objects-main")
	if err != nil {
		return nil
	}
	candidate, err := cacheDirectory("native-objects-candidate")
	if err != nil {
		return nil
	}
	units, err := cacheDirectory("native-units")
	if err != nil {
		return nil
	}
	cache, on := objectCacheFor(objectTrust(), main, candidate, units)
	if !on {
		return nil
	}
	if why, off := unitCacheOff(units); off {
		fmt.Fprintf(os.Stderr, "native: main.c compiles uncached: %s\n", why)
		return nil
	}
	return cache.compile
}

// An objectCache is where main.c's objects are read and written under one trust, and units the unit cache's directory,
// whose mark a mismatch leaves (markUnitCacheOff), turning both caches off.
type objectCache struct {
	reads  []string
	writes string
	units  string
}

// objectCacheFor is the object cache trust reads and writes, of the directories main and candidate, and whether trust
// names one.
func objectCacheFor(trust string, main string, candidate string, units string) (objectCache, bool) {
	switch trust {
	case "main":
		return objectCache{reads: []string{main}, writes: main, units: units}, true
	case "candidate":
		return objectCache{reads: []string{main, candidate}, writes: candidate, units: units}, true
	}
	return objectCache{}, false
}

// compile makes main.o in the compile's directory: copied from the first of the cache's directories holding an object
// under main.c's key, or compiled, and once its recheck passed, stored in the directory the cache writes. A compile the
// key can't account for compiles uncached, and says why. With unitCheckVariable set, every hit compiles too and the two
// objects are compared: a mismatch prints UnitMismatchPrefix's line, marks the unit cache off, and is the cold object.
func (cache objectCache) compile(compile *exec.Cmd, source string, compileFlags []string, program string, jobserver *Jobserver) error {
	keying, err := newObjectKeying(compile, source, compileFlags, program, jobserver)
	if errors.Is(err, errUncachedObject) {
		fmt.Fprintln(os.Stderr, err)
		clang, err := compileMain(compile, jobserver)
		if err != nil {
			return err
		}
		buildLog("cCompile", "main.c", "", "off", clang)
		return nil
	}
	if err != nil {
		return err
	}
	return cache.compileKeyed(keying, compile, jobserver)
}

// compileKeyed is compile once main.c is keyed, a step of its own so a test can move an input between the key and the
// compile.
func (cache objectCache) compileKeyed(keying *objectKeying, compile *exec.Cmd, jobserver *Jobserver) error {
	object := filepath.Join(compile.Dir, "main.o")
	for _, directory := range cache.reads {
		if found, ok := cachedFile(filepath.Join(directory, keying.key, "main.o")); ok {
			return cache.hit(keying.key, found, object, compile, jobserver)
		}
	}
	clang, err := compileMain(compile, jobserver)
	if err != nil {
		return err
	}
	// clang compiled it, so it is a miss whether or not it's stored.
	buildLog("cCompile", "main.c", keying.key, "miss", clang)
	if differs := keying.recheck(jobserver); differs != "" {
		fmt.Fprintf(os.Stderr, "native: main.c's object isn't stored: %s\n", differs)
		return nil
	}
	return cache.store(keying.key, object)
}

// hit copies the object found under key to object, never linking the cache's file in place, unless unitCheckVariable
// asks every hit compiled again: then compile makes object, and a cached object other than it is a mismatch.
func (cache objectCache) hit(key string, found string, object string, compile *exec.Cmd, jobserver *Jobserver) error {
	cached, err := os.ReadFile(found)
	if err != nil {
		return err
	}
	if os.Getenv(unitCheckVariable) == "" {
		var copying time.Duration
		if err := timed(&copying, func() error { return os.WriteFile(object, cached, 0o644) }); err != nil {
			return err
		}
		buildLog("cCompile", "main.c", key, "hit", copying)
		return nil
	}
	clang, err := compileMain(compile, jobserver)
	if err != nil {
		return err
	}
	cold, err := os.ReadFile(object)
	if err != nil {
		return err
	}
	if !bytes.Equal(cached, cold) {
		mismatch := unitMismatch{Unit: "main.c", Key: key, Cached: fmt.Sprintf("%x", sha256.Sum256(cached)), Cold: fmt.Sprintf("%x", sha256.Sum256(cold)), At: time.Now().UTC()}
		// The cold bytes were compiled from main.c's sources, so they're its object, and the build goes on with them.
		marked := "marked off at " + filepath.Join(cache.units, unitCacheOffName)
		if err := markUnitCacheOff(cache.units, mismatch); err != nil {
			marked = "its mark couldn't be written: " + err.Error()
		}
		fmt.Fprintln(os.Stderr, unitMismatchLine(mismatch)+"; the unit cache is "+marked)
		noteUnitMismatch(mismatch)
		buildLog("cCompile", "main.c", key, "miss", clang)
		return nil
	}
	buildLog("cCompile", "main.c", key, "hit", clang)
	return nil
}

// store publishes object under key in the directory the cache writes, made whole before it's in place, as compileUnitAs
// publishes a unit; an entry another process published first is kept.
func (cache objectCache) store(key string, object string) error {
	contents, err := os.ReadFile(object)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cache.writes, 0o755); err != nil {
		return err
	}
	published, err := os.MkdirTemp(cache.writes, ".publish-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(published)
	if err := os.WriteFile(filepath.Join(published, "main.o"), contents, 0o644); err != nil {
		return err
	}
	entry := filepath.Join(cache.writes, key)
	if err := os.Rename(published, entry); err != nil {
		if _, ok := cachedFile(filepath.Join(entry, "main.o")); !ok {
			return err
		}
	}
	return nil
}

// newObjectKeying reads the key of compile, main.c's compile, which holds source and is compiled with compileFlags by
// the name program, or errUncachedObject naming what the key can't account for: a source that reads the time it's
// compiled, what a split unit's key refuses of the flags, the environment and clang, or a preprocessing that fails,
// which leaves the compile to fail as it would. It runs main.c's preprocessing in the compile's directory, writing main.i
// and main.d there.
func newObjectKeying(compile *exec.Cmd, source string, compileFlags []string, program string, jobserver *Jobserver) (*objectKeying, error) {
	if macro := unitTime.FindString(source); macro != "" {
		return nil, fmt.Errorf("%w: main.c reads %s", errUncachedObject, macro)
	}
	if err := unitFlagsRefusal(compileFlags); err != nil {
		return nil, objectRefusal(err)
	}
	environmentParts, err := unitEnvironmentParts(compile.Env)
	if err != nil {
		return nil, objectRefusal(err)
	}
	places := objectPlaces(compile.Args[1:])
	placeheld := objectPlaceheld(places)
	// clang's build is read under the compile's flags as the key writes them, and its environment but for PWD, the build
	// directory, so a process reads it once and not once a build.
	var flags []string
	for _, flag := range compileFlags {
		flags = append(flags, placeheld.Replace(flag))
	}
	environment := slices.DeleteFunc(slices.Clone(compile.Env), func(setting string) bool { return strings.HasPrefix(setting, "PWD=") })
	clang, err := unitClang(compile.Path, flags, environment)
	if err != nil {
		return nil, objectRefusal(err)
	}
	keying := &objectKeying{compile: compile, preprocess: append(slices.Clone(compileFlags), "-E", program)}
	preprocessed, preprocessing, err := keying.run(jobserver, "main.i", "-MD", "-MF", "main.d")
	if err != nil {
		return nil, fmt.Errorf("%w: its preprocessing failed: %v", errUncachedObject, err)
	}
	keying.preprocessed = preprocessed
	rules, err := os.ReadFile(filepath.Join(compile.Dir, "main.d"))
	if err != nil {
		return nil, fmt.Errorf("%w: its preprocessing's dependencies: %v", errUncachedObject, err)
	}
	// A name the make rules spell in a way dependencies doesn't read ($$, \#) fails its stat, and refuses. Each file is
	// stated before it's hashed, so bytes that change after the hash move the change time the recheck looks at again.
	var files []string
	seen := map[string]bool{}
	for _, name := range dependencies(rules) {
		if seen[name] {
			continue
		}
		seen[name] = true
		read, err := objectReadOf(compile.Dir, name)
		if err != nil {
			return nil, fmt.Errorf("%w: a file its preprocessing read: %v", errUncachedObject, err)
		}
		keying.reads = append(keying.reads, read)
		path := name
		if !filepath.IsAbs(path) {
			path = filepath.Join(compile.Dir, path)
		}
		digest, err := fileDigest(path)
		if err != nil {
			return nil, fmt.Errorf("%w: a file its preprocessing read: %v", errUncachedObject, err)
		}
		files = append(files, "file "+placeheld.Replace(name)+" "+digest)
	}
	// v1: keyed by main.c preprocessed, every file it read by its bytes, in the dependency list's order, and the compile's
	// arguments, every per-build place a placeholder (#2r9xhqs).
	parts := []string{"adamic-objects-v1", "os " + goruntime.GOOS, "arch " + goruntime.GOARCH}
	parts = append(parts, clang.parts...)
	parts = append(parts, environmentParts...)
	for _, argument := range compile.Args[1:] {
		parts = append(parts, "argument "+placeheld.Replace(argument))
	}
	parts = append(parts, files...)
	parts = append(parts, fmt.Sprintf("preprocessed %x", sha256.Sum256(objectMarkersMapped(preprocessed, places))))
	keying.key = partsKey(parts)
	buildLog("cPreprocess", "main.c", keying.key, "off", preprocessing)
	return keying, nil
}

// objectRefusal is err, when it is a split unit key's refusal (errUncachedUnit), as main.c's, and any other error as it
// is.
func objectRefusal(err error) error {
	if !errors.Is(err, errUncachedUnit) {
		return err
	}
	return fmt.Errorf("%w: %s", errUncachedObject, strings.TrimPrefix(err.Error(), errUncachedUnit.Error()+": "))
}

// run runs main.c's preprocessing, writing output and whatever more arguments ask, in the compile's directory under
// exactly its environment, and is the bytes written to output and how long clang ran, timed inside its jobserver.Hold so
// a token's wait isn't counted.
func (keying *objectKeying) run(jobserver *Jobserver, output string, arguments ...string) ([]byte, time.Duration, error) {
	command := clangIn(keying.compile.Dir, keying.compile.Path, slices.Concat(keying.preprocess, arguments, []string{"-o", output})...)
	command.Env = keying.compile.Env
	var combined []byte
	var clang time.Duration
	if err := jobserver.Hold(func() error {
		return timed(&clang, func() (err error) { combined, err = command.CombinedOutput(); return err })
	}); err != nil {
		return nil, clang, fmt.Errorf("%w\n%s", err, combined)
	}
	preprocessed, err := os.ReadFile(filepath.Join(keying.compile.Dir, output))
	return preprocessed, clang, err
}

// recheck runs main.c's preprocessing again, after its compile, to a second file, and looks again at every file the
// first read: "" when the bytes are the first's and every file has the inode and change time it had, or what differs,
// in which case the compile may have read other bytes than the key did, and its object may not be stored. The second
// preprocessing writes no dependency list: the first's names every file to look at again, and a file the second reads
// that the first didn't changes the bytes.
func (keying *objectKeying) recheck(jobserver *Jobserver) string {
	var differences []string
	preprocessed, rechecking, err := keying.run(jobserver, "main.recheck.i")
	switch {
	case err != nil:
		differences = append(differences, fmt.Sprintf("its preprocessing failed after its compile: %v", err))
	case !bytes.Equal(preprocessed, keying.preprocessed):
		differences = append(differences, "main.c preprocessed to other bytes after its compile than before it")
	}
	timed(&rechecking, func() error {
		for _, read := range keying.reads {
			if again, err := objectReadOf(keying.compile.Dir, read.name); err != nil || again != read {
				differences = append(differences, read.name+" changed between main.c's key and its recheck (its inode or change time moved)")
			}
		}
		return nil
	})
	buildLog("cRecheck", "main.c", keying.key, "off", rechecking)
	return strings.Join(differences, "; ")
}
