// Package buildcache is the one home for build products tests need: the checker archive, stage 0, Go oracle
// binaries, a port's native build. Each product is addressed by what can change it, built once per address and
// reused, so a test's time starts after the fetch (@system_adamic's ruling on the 30 s rule, Oct 8).
//
// A product's address is its name key (its name, flags and tool reports) and the read set its last traced build
// measured, valued on the tree at hand (reads.go, #vt46geg): a change misses only the products that read it. A product
// is found by the cache's read sets, which a runner's cache holds as Loom's tree index unpacked them. Only Workshop's
// tree builder (ADAMIC_BUILD_STORE=traced, under cmd/traced) builds traced, and only a settled traced build is placed
// under its key (settle.go). Every other miss builds as before, for this machine alone: keyed by the declared Files,
// never found by a read set, and never failing for want of a trace.
//
// A product on disk is always complete: it is filled in a private directory and placed whole by rename, so a failed
// build leaves nothing. ADAMIC_BUILD_CACHE=off builds every time into a fresh directory: the uncached proof mode,
// which is what lands main. ADAMIC_BUILD_CACHE=read never builds: a runner's mode, whose products and read sets arrive
// built (ErrNotBuilt). The cache is local to the machine (ADAMIC_BUILD_CACHE_DIR, or the user cache directory's
// adamic-build), and a miss there fetches from the shared store by the keys its read sets give (store.go).
package buildcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Product returns the directory holding the product built from inputs, failing the test if it can't be built.
func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {
	t.Helper()
	return product(t, inputs, build, nil)
}

// LeanProduct is Product for a build that takes nothing from its caller but inputs: everything it uses is a file
// inputs.Files names or what it reads while it builds. Under the trace its read set is its own (#mmpt615): what its
// process and the processes it starts read while it builds, and every file inputs.Files names, never what the test
// read before it, which a test's shards differ by. Only callers buildcache reviews make that promise
// (leanCallers, held by a test): a closure over a test's own bytes would be keyed without them.
func LeanProduct(t testing.TB, inputs Inputs, build func(directory string) error) string {
	t.Helper()
	return product(t, inputs, build, leanCalled(1))
}

// LeanGet is Get for a lean build (LeanProduct).
func LeanGet(inputs Inputs, build func(directory string) error) (string, error) {
	directory, _, err := get(inputs, build, leanCalled(1))
	return directory, err
}

// A leanBuild is a build Settle keys by its own window (LeanProduct). A lean build's window counts a file its test wrote
// before the window as an input, never as the build's own output.
//
// caller is the lean caller whose package holds the build's code, as an import path (leanCalled), or "" when it is
// buildcache's own: Settle keys a lean build by the recipe of the code that makes it (leanRecipeOf, #earlycut), never by
// the recipe of the test binary that settled it.
type leanBuild struct {
	caller string
}

// leanCalled is the lean build its caller asks for, skip frames above leanCalled's own: built from buildcache's code, or
// from the package that called, when it's one of leanCallers. Any other package's build isn't lean (nil): its code is
// unreviewed, so it keeps the settling binary's whole recipe.
func leanCalled(skip int) *leanBuild {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return nil
	}
	return leanBuildOf(functionPackage(runtime.FuncForPC(pc).Name()))
}

// leanBuildOf is the lean build caller, a package's import path, asks for: buildcache's own, a reviewed lean caller's
// (leanCallers), or none.
func leanBuildOf(caller string) *leanBuild {
	switch {
	case os.Getenv("ADAMIC_BUILD_LEAN") == "off":
		// Loom's switch while a lean mismatch is open (#earlycut's sampled key check): every product keys by the whole
		// recipe, read once at a tree's start, since build-trees sets it there.
		return nil
	case caller == buildcachePackage:
		return &leanBuild{}
	case slices.Contains(leanCallers, strings.TrimPrefix(caller, modulePath+"/")):
		return &leanBuild{caller: caller}
	}
	return nil
}

// modulePath is the repository's module, the prefix of every package of it as go names one.
const modulePath = "github.com/system-inc/adamic"

// functionPackage is the import path of the package defining the function a symbol names, as the runtime spells it
// (github.com/system-inc/adamic/internal/native/runtimeproducts.Build.func1).
func functionPackage(symbol string) string {
	slash := strings.LastIndex(symbol, "/")
	if dot := strings.Index(symbol[slash+1:], "."); dot >= 0 {
		return symbol[:slash+1+dot]
	}
	return symbol
}

// leanCallers are the packages outside buildcache that may build lean (LeanProduct, LeanGet), each reviewed to take
// nothing from its caller but its Inputs: native's runtime archives, whose sources are the runtime's files.
var leanCallers = []string{"internal/native/runtimeproducts"}

func product(t testing.TB, inputs Inputs, build func(directory string) error, lean *leanBuild) string {
	t.Helper()
	directory, line, err := get(inputs, build, lean)
	if line != "" {
		t.Log(line)
	}
	if errors.Is(err, ErrNotBuilt) {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	return directory
}

// ErrNotBuilt is a product missing under ADAMIC_BUILD_CACHE=read, the mode of a machine that runs tests and never
// builds: Workshop's loom build-tree runs every TestProduct_ test and ships the cache it filled, and a runner reads it.
// A miss there is a product no TestProduct_ test builds, or a key this machine computes differently from Workshop, and
// the error names the product, its key and every input that keyed it, so either reads as what it is.
var ErrNotBuilt = errors.New("not built")

// Get is Product for TestMain and tools.
func Get(inputs Inputs, build func(directory string) error) (string, error) {
	directory, _, err := get(inputs, build, nil)
	return directory, err
}

// ErrUntraced is a product Workshop's tree builder (ADAMIC_BUILD_STORE=traced) would build outside cmd/traced: what it
// builds is for every runner, so it must be keyed by what its build read, and building untraced there is refused.
var ErrUntraced = errors.New("ADAMIC_BUILD_STORE=traced builds only under cmd/traced")

// get finds or builds inputs' product; a lean build takes nothing from its caller (LeanProduct), so Settle keys it by its
// own window, and nil is any other.
func get(inputs Inputs, build func(directory string) error, lean *leanBuild) (string, string, error) {
	started := time.Now()
	root, err := repositoryRoot()
	if err != nil {
		return "", "", err
	}
	// A declared File the machine's own key can't hash is refused here, before the build takes a path: a traced build
	// keys by what it reads and never hashes its declared Files, so an absolute one was green in Loom's trees and red on
	// every untraced build after it landed (lint's jsx owners, #j9st2ff; markdownblocks' answers, #y6ff7ea; #r7cb4bd).
	for _, name := range inputs.Files {
		if err := declaredFile(name); err != nil {
			return "", "", err
		}
	}
	var nameKey string
	keying(func() { nameKey = NameKey(root, inputs) })
	// Every build site, uncached, traced or for this machine alone, builds through relocating, and every directory
	// returned is held under its name key: a product names itself and every other product by name key (relative.go).
	build = relocating(inputs, nameKey, build)
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		directory, err := os.MkdirTemp("", "adamic-build-")
		if err != nil {
			return "", "", err
		}
		if traceDirectory() == "" {
			if err = build(directory); err != nil {
				return "", "", err
			}
			return holding(nameKey, directory), record(inputs.Name, "uncached", "off", started), nil
		}
		// Traced and uncached (main's own gate on Workshop): built fresh, and settled from where it was built.
		_, id, err := tracedBuild(root, inputs, directory, build)
		if err != nil {
			return "", "", err
		}
		journalBuilt(root, inputs, nameKey, id, directory, lean)
		return holding(nameKey, directory), record(inputs.Name, "uncached", "off", started), nil
	}
	cache, err := cacheDirectory()
	if err != nil {
		return "", "", err
	}
	var described string
	var looked found
	keying(func() {
		described = describe(root, inputs)
		looked, err = productFor(root, cache, nameKey, 0, nil)
	})
	if err != nil {
		return "", "", err
	}
	// Only the cache's own read sets give a product's keys: the store holds products by key, never read sets, so a name
	// key the cache holds no read set for has nothing to fetch (Oct 14: asking the store for read sets was 96.7% of its
	// billed reads, every one a 404).
	if looked.directory != "" {
		listed, err := auditListed(nameKey)
		if err == nil && listed {
			err = auditHit(looked.key, inputs.Name, looked.directory, build)
		}
		if err != nil {
			return "", "", err
		}
		journalFound(nameKey, inputs.Name, looked.directory, looked.upstreams)
		rerecord(cache, nameKey, inputs.Name, looked.refound)
		return holding(nameKey, looked.directory), record(inputs.Name, looked.key, "hit", started), nil
	}
	if storeAddress() != "" {
		for index, key := range looked.keys {
			outcome := "fetched"
			product := filepath.Join(cache, key)
			_, err := filled(product, described, func(scratch string) error {
				if err := fetch(key, scratch); err != nil {
					return err
				}
				// An audit rebuilds, so it runs only where building may: under a trace.
				if traceDirectory() != "" && auditing() {
					outcome = "audited"
					return audit(key, inputs.Name, scratch, build)
				}
				return nil
			})
			if errors.Is(err, errNotStored) {
				note("store %s %s: %v", inputs.Name, key[:12], err)
				continue
			}
			if err != nil {
				return "", "", err
			}
			journalFound(nameKey, inputs.Name, product, looked.keyUpstreams[index])
			rerecord(cache, nameKey, inputs.Name, looked.keyRefound[index])
			return holding(nameKey, product), record(inputs.Name, key, outcome, started), nil
		}
	}
	switch {
	case os.Getenv("ADAMIC_BUILD_CACHE") == "read":
		return "", record(inputs.Name, nameKey, "missing", started), fmt.Errorf("product %s (name key %s) is %w: %s, in %s, and ADAMIC_BUILD_CACHE=read reads products and never builds them; its package's TestProduct_ test builds it, on Workshop under a trace. Its inputs, as this machine keys them:\n%s", inputs.Name, nameKey[:12], ErrNotBuilt, looked.why, cache, described)
	case traceDirectory() != "":
		// Built here under the trace, and found by every process under it, until Settle keys it by what it read.
		product := filepath.Join(pendingDirectory(cache), nameKey)
		var id string
		built, err := filled(product, described, func(scratch string) (err error) {
			_, id, err = tracedBuild(root, inputs, scratch, build)
			return err
		})
		if err != nil {
			return "", "", err
		}
		outcome := "hit"
		if built {
			journalBuilt(root, inputs, nameKey, id, product, lean)
			outcome = "miss"
		} else {
			journal(journalEntry{Event: "found", NameKey: nameKey, Directory: product})
		}
		return holding(nameKey, product), record(inputs.Name, nameKey, outcome, started), nil
	case treeBuilder():
		return "", record(inputs.Name, nameKey, "untraced", started), fmt.Errorf("product %s (name key %s): %s, and %w, so a build's reads become its key", inputs.Name, nameKey[:12], looked.why, ErrUntraced)
	}
	// Untraced, as every machine but Workshop's tree builder builds: for this machine alone, keyed by the declared
	// Files and never found by a read set.
	var key string
	keying(func() { key, err = Key(root, inputs) })
	if err != nil {
		return "", "", err
	}
	product := filepath.Join(cache, "local", key)
	built, err := filled(product, described, build)
	if err != nil {
		return "", "", err
	}
	outcome := "hit"
	if built {
		outcome = "miss"
	}
	// Another process finds this product by its name key alone, through the inputs that key it here.
	if err = pointLocal(cache, nameKey, inputs); err != nil {
		note("local name %s %s: %v", inputs.Name, nameKey[:12], err)
	}
	return holding(nameKey, product), record(inputs.Name, key, outcome, started), nil
}

// tracedBuild builds into directory between markers naming the build, and returns its name key and id.
func tracedBuild(root string, inputs Inputs, directory string, build func(directory string) error) (string, string, error) {
	var nameKey string
	keying(func() { nameKey = NameKey(root, inputs) })
	id, err := traced(nameKey, func() error { return build(directory) })
	return nameKey, id, err
}

// journalBuilt journals a traced build where its product now is, for Settle to key and place, and whether it is lean.
func journalBuilt(root string, inputs Inputs, nameKey, id, directory string, lean *leanBuild) {
	entry := journalEntry{Event: "built", Build: id, NameKey: nameKey, Directory: directory, Declared: inputs.Files, Root: root, Lean: lean != nil}
	if lean != nil {
		entry.LeanCaller = lean.caller
	}
	keying(func() { entry.Name, entry.Inputs = portable(root, inputs.Name), describe(root, inputs) })
	journal(entry)
}

// filled makes target by fill when it isn't there: into a private directory beside it, then whole by rename, so a
// product on disk is always complete and a failed fill leaves nothing; one filler per target across every process on
// the machine, the rest waiting to find it filled. It says whether this call filled it.
func filled(target, inputs string, fill func(scratch string) error) (bool, error) {
	if _, err := os.Stat(target); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return false, err
	}
	unlock, err := lockTarget(target)
	if err != nil {
		return false, err
	}
	defer unlock()
	if _, err = os.Stat(target); err == nil {
		return false, nil
	}
	name := filepath.Base(target)
	scratch, err := os.MkdirTemp(filepath.Dir(target), ".building-"+name[:min(12, len(name))]+"-")
	if err != nil {
		return false, err
	}
	if err = fill(scratch); err != nil {
		os.RemoveAll(scratch)
		return false, err
	}
	if err = os.WriteFile(target+".inputs", []byte(inputs), 0o644); err != nil {
		os.RemoveAll(scratch)
		return false, err
	}
	if err = os.Rename(scratch, target); err != nil {
		os.RemoveAll(scratch)
		// A product another placer put there whole is the product: found, not a failure (#bcae1e2).
		if _, found := os.Stat(target); found == nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// lockTarget takes target's lock, which every process that places a product at target holds from its look to its
// rename: filled while it fills, and Settle while it places a settled build (placeSettled). Without it, a fetch and a
// settlement of one key each saw it missing, and the one whose rename lost failed its test with "file exists" while the
// product sat there (#bcae1e2). It returns the release.
func lockTarget(target string) (func(), error) {
	lock, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		lock.Close()
		return nil, err
	}
	return func() {
		syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		lock.Close()
	}, nil
}

// keyVersion is how buildcache builds and holds what Key addresses, the one input no product declares: a change to what
// a product's bytes hold for the same inputs moves it, so no key names two products. v2 is relocating's: a product
// names itself and every other product by name key (relative.go), so the same inputs made other bytes than v1's
// (Workshop, 02:50Z Oct 11: seven products a v1 tree published under one key, landable-6 built again, different). v3 is
// unstamped: Loom's unit environment builds with -buildvcs=false since 05:01Z Oct 11, and a hand-rolled product (a go
// build, a c-archive) keys the flags it declares, never GOFLAGS, so its v2 keys still named the stamped bytes trees had
// published before (landable-9b: 54 refs/action conflicts, each against a ref written at 04:14Z). v4 is placeless: a
// native product's debug info no longer names the random directory it was built in (native.Build and the runtime map it,
// #tsn1wp8), so its bytes are the same build to build, and differ from every v3 product that named its directory
// (landable-11: lint-profile-profiled and yaml-scalars-native, against refs landable-10 wrote). v5 is git's: a declared
// file is valued as a read set values it, by its git blob id and exec bit, a clean tracked file by the id git's index
// holds without reading it (#r47fdtt: a full hit of cssstrings' Setup spent 15 to 19 s hashing declared trees, all of
// cohere 5.7 s, in every process that keyed it). A product's bytes are v4's, but every declared key moves, so each
// misses once, under a version that says why.
const keyVersion = "v5"

// Key is the product's address: a hash of every input, each length-prefixed so no two inputs run together.
func Key(root string, inputs Inputs) (string, error) {
	return keyAt(keyVersion, root, inputs)
}

// keyAt is Key under the given version, so a test can key the same inputs under an older one.
func keyAt(version, root string, inputs Inputs) (string, error) {
	hash := sha256.New()
	field := func(kind, value string) {
		fmt.Fprintf(hash, "%s %d\n%s\n", kind, len(value), value)
	}
	field("buildcache", version)
	field("name", portable(root, inputs.Name))
	for _, name := range inputs.Files {
		if err := hashPath(hash, root, name, field); err != nil {
			return "", err
		}
	}
	for _, flag := range inputs.Flags {
		field("flag", portable(root, flag))
	}
	for _, tool := range inputs.Toolchain {
		field("tool", portable(root, tool))
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// portable is a name, a flag or a tool report as every machine holding the same tree spells it, so a product Workshop built
// is the key a runner asks for (#t37sw0f: runners missed every product because keys named Workshop's machine). The
// repository's absolute path becomes <repository> and the home directory ~, wherever either appears as a whole path,
// so a value keeps what it names inside the tree and loses where the tree was checked out. Go's parallel share
// (-p=N, or -p N, which Workshop's GOFLAGS carries as the core count it computed) is dropped: it changes how many
// packages build at once, never what they build. Everything else stays verbatim, so a flag that changes the output
// still changes the key.
func portable(root, value string) string {
	for _, place := range places(root) {
		value = replacePath(value, place.path, place.name)
	}
	return withoutParallelShare(value)
}

// places are the directories portable rewrites, longest first, so a checkout inside the home directory becomes
// <repository> rather than ~/....: the repository, the build cache (a key that names another product by its path,
// such as estree's main=<build cache>/<key>/main.ts, names it by its key), the home directory, and each as its symbolic
// links resolve (macOS spells /tmp as /private/tmp). A product this process holds anywhere else is <build cache>/<name
// key> too, as Relative names it: one built outside the cache (uncached, or its scratch directory while it builds), and the
// copy Resolved made of one, whose name holds a hash of this machine's roots (#sgemgmn: estree's misc port, keyed by
// its resolved main.ts, missed on every checkout but the one that built it).
func places(root string) []place {
	var found []place
	seen := map[string]bool{}
	add := func(path, name string, resolution func(string) string) {
		if !filepath.IsAbs(path) || filepath.Clean(path) == string(filepath.Separator) {
			return
		}
		for _, candidate := range []string{filepath.Clean(path), resolution(path)} {
			if candidate != "" && !seen[candidate] {
				seen[candidate] = true
				found = append(found, place{candidate, name})
			}
		}
	}
	if root != "" {
		add(root, "<repository>", resolved)
	}
	if cache, err := cacheLocation(); err == nil {
		add(cache, "<build cache>", resolved)
	}
	if home, err := os.UserHomeDir(); err == nil {
		add(home, "~", resolved)
	}
	for _, products := range []*sync.Map{&productDirectories, &resolvedCopies} {
		products.Range(func(directory, key any) bool {
			add(directory.(string), "<build cache>/"+key.(string), productResolved)
			return true
		})
	}
	sort.SliceStable(found, func(i, j int) bool { return len(found[i].path) > len(found[j].path) })
	return found
}

// productResolutions holds what each product directory places has met resolves to. places runs for every key a process
// computes, over every product it holds, and a resolution is a stat per path element, each one a stop under a trace
// (home-6's dry run, Oct 11: bridge/tsgo's BridgeInputProofs spent its whole 10 minutes resolving product directories
// for the keys it checks). A product directory doesn't move once a process holds it, so each resolves once.
var productResolutions sync.Map

func productResolved(directory string) string {
	if value, ok := productResolutions.Load(directory); ok {
		return value.(string)
	}
	value := resolved(directory)
	productResolutions.Store(directory, value)
	return value
}

// parallelShare is go's -p flag as a token of its own: after the start, a space, a quote or an equals sign (GOFLAGS=-p=8),
// and before the end, a space or a quote.
var parallelShare = regexp.MustCompile(`(^|[\s'"=])-p(?:=|[ \t]+)[0-9]+([\s'"]|$)`)

func withoutParallelShare(value string) string {
	for {
		match := parallelShare.FindStringSubmatchIndex(value)
		if match == nil {
			return value
		}
		left, right := value[match[2]:match[3]], value[match[4]:match[5]]
		// What bounded the token survives, once: "a -p=8 b" is "a b", "a -p=8" is "a", "GOFLAGS=-p=8 -trimpath" is
		// "GOFLAGS=-trimpath", "'-p=8'" is "''". A quote or an equals sign is kept; two spaces become one.
		leftMark, rightMark := strings.TrimSpace(left) != "", strings.TrimSpace(right) != ""
		keep := ""
		switch {
		case leftMark && rightMark:
			keep = left + right
		case leftMark:
			keep = left
		case rightMark:
			keep = right
		case left != "" && right != "":
			keep = left
		}
		value = value[:match[0]] + keep + value[match[1]:]
	}
}

// declaredFile refuses a declared File that names no path of the repository by its root-relative name: an absolute
// path, or one that leaves the root.
func declaredFile(name string) error {
	if filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..") {
		return fmt.Errorf("input %q must be inside the repository, relative to its root", name)
	}
	return nil
}

func hashPath(hash io.Writer, root, name string, field func(kind, value string)) error {
	if err := declaredFile(name); err != nil {
		return err
	}
	start := filepath.Join(root, name)
	// WalkDir visits in lexical order, so the same tree always hashes the same.
	return filepath.WalkDir(start, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		// Git's own record is never an input: a .git directory, or a submodule's .git file, whose content is the path of
		// its git directory (gitdir: /home/ahra/...), so a checkout, a worktree and an unpacked source of one tree key
		// the same (#smkk3et).
		if entry.Name() == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if unkeyed(relative) {
			return nil
		}
		// Under a named directory, only what the tree carries: a checkout's untracked files (a build's leftovers, a
		// stray note) are not in the source a runner unpacks, so they never key a product. A path named in Files is
		// hashed whether or not git tracks it.
		if path != start && !Tracked(root, path, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			field("directory", relative)
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			field("symlink", relative+" -> "+target)
		case info.Mode().IsRegular():
			// As a read set values it: a clean tracked file by the id git's index holds, unread, anything else hashed as
			// git hashes a blob, so a checkout and an unpacked source key the same.
			value, err := gitValue(root, path, info)
			if err != nil {
				return err
			}
			field("file", relative+" "+value)
		default:
			return fmt.Errorf("input %s is neither a file, a directory nor a symlink", relative)
		}
		return nil
	})
}

// unkeyed is a file no key reads even when named: go.work.sum, go's record of the checksums of modules the workspace
// fetched, which go writes as a run needs one (Workshop's proof, Oct 10: a checkout gained one mid-run and every key
// naming it moved) and which changes no build's output, since go.mod and go.work pin what it verifies.
func unkeyed(name string) bool {
	return path.Base(name) == "go.work.sum"
}

// trackedFiles is what git tracks in one repository: each tracked path (a file, a symbolic link, or a submodule's
// gitlink), relative to it, and every directory holding one. all is set when git couldn't say, and then everything
// counts, as in a tree with no git at all.
type trackedFiles struct {
	paths, directories map[string]bool
	all                bool
	// objects are git's ids of the files the index (or an unpacked source's manifest) holds, by path, with their modes;
	// dirty are those whose working copy differs from the index when taken, so a read set values a clean file by its id
	// without reading it. taken is when git was asked, or when the manifest's listing was written.
	objects map[string]gitObject
	dirty   map[string]bool
	taken   time.Time
}

type gitObject struct{ mode, id string }

// repositories holds each directory's trackedFiles once per process, or nil for a directory that isn't a repository's.
var repositories sync.Map

// Tracked says whether name, under root, is in the source the tree carries: an entry of the nearest repository above
// it at or under root (git ls-files there, so a submodule answers for its own files), or anything at all when there is
// none, as in an unpacked source. Key asks it of everything under a directory it hashes; a test that lists a tree's
// files itself into Files asks it too, so a checkout's leftovers never key a product (#smkk3et).
func Tracked(root, name string, directory bool) bool {
	for at := filepath.Dir(name); inside(root, at) || at == root; at = filepath.Dir(at) {
		files := repositoryAt(at)
		if files == nil {
			if at == root {
				break
			}
			continue
		}
		if files.all {
			return true
		}
		relative, err := filepath.Rel(at, name)
		if err != nil {
			return true
		}
		relative = filepath.ToSlash(relative)
		// Inside a submodule whose files are there without its git directory (productidentity -elsewhere links cohere's
		// in), the repository above knows the submodule only as a gitlink, so everything under it counts.
		for parent := path.Dir(relative); parent != "."; parent = path.Dir(parent) {
			if files.paths[parent] {
				return true
			}
		}
		return files.paths[relative] || directory && files.directories[relative]
	}
	return true
}

// statGit is os.Lstat, of a directory's .git: a test watches it.
var statGit = os.Lstat

func repositoryAt(directory string) *trackedFiles {
	if found, ok := repositories.Load(directory); ok {
		return found.(*trackedFiles)
	}
	var files *trackedFiles
	// Whether directory holds .git is a look at .git, which no product's read set may hold, so it is keying, as the
	// listing is: a build that asks Tracked of a directory never reads .git (#1p6sent, a traced walk of the tree).
	repository := false
	keying(func() {
		_, err := statGit(filepath.Join(directory, ".git"))
		repository = err == nil
	})
	if repository {
		files = &trackedFiles{paths: map[string]bool{}, directories: map[string]bool{}, objects: map[string]gitObject{}, dirty: map[string]bool{}, taken: time.Now()}
		var listed, dirty []byte
		var err error
		// git reads .git, which no product's read set may hold: listing what the tree carries is keying. It runs with
		// toolchain's git environment, so nothing of the home (a ~/.gitconfig's filter, fsmonitor or autocrlf) shapes
		// what it says the tree holds.
		keying(func() {
			if listed, err = toolchain.GitCommand("-C", directory, "ls-files", "-s", "-v", "-z").Output(); err == nil {
				dirty, err = toolchain.GitCommand("-C", directory, "diff-files", "--name-only", "-z").Output()
			}
		})
		if err != nil {
			files.all = true
		}
		for _, name := range strings.Split(string(dirty), "\x00") {
			files.dirty[name] = true
		}
		for _, line := range strings.Split(string(listed), "\x00") {
			// <tag> <mode> <id> <stage>\t<path>: a tag other than H (assume-unchanged in lower case, S for
			// skip-worktree) is a file git stops comparing with its index, so its id may be stale and it is hashed.
			fields, name, ok := strings.Cut(line, "\t")
			if !ok || name == "" {
				continue
			}
			if parts := strings.Fields(fields); len(parts) == 4 {
				files.objects[name] = gitObject{mode: parts[1], id: parts[2]}
				if parts[0] != "H" {
					files.dirty[name] = true
				}
			}
			files.paths[name] = true
			for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
				files.directories[parent] = true
			}
		}
	} else if listing, taken, ok := manifestListing(directory); ok {
		// A Loom runner's source has no .git: build-tree ships git's answers for each repository in a manifest, so the
		// tracked files are its listing and nothing else, never the npm packages or the manifest the source carries
		// beside them (#59paqn3). A file is valued by the blob id the listing holds for it, unread, as a checkout's
		// clean file is by its index's, when it was last modified a second before the listing was written: the source
		// was unpacked from the commit the listing names (Loom checks every file against it before it ships one), so
		// only a change since, which moves the file's time past the listing's, can make the two differ, and such a file
		// is hashed (#szswjtc: a runner's lookup read and hashed every file of a product's read set, 8 to 27 s a hit).
		// Either way a file's value is its git blob id, so the key is a checkout's.
		files = &trackedFiles{paths: map[string]bool{}, directories: map[string]bool{}, objects: map[string]gitObject{}, dirty: map[string]bool{}, taken: taken}
		for _, line := range strings.Split(string(listing), "\x00") {
			// <mode> <type> <id>\t<path>, as git ls-tree -r -z --full-tree HEAD writes it. A submodule's entry (a
			// commit) names its repository's commit, not a file's content, so only a blob is a file's id.
			fields, name, ok := strings.Cut(line, "\t")
			if !ok || name == "" {
				continue
			}
			if parts := strings.Fields(fields); len(parts) == 3 && parts[1] == "blob" {
				files.objects[name] = gitObject{mode: parts[0], id: parts[2]}
			}
			files.paths[name] = true
			for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
				files.directories[parent] = true
			}
		}
	}
	found, _ := repositories.LoadOrStore(directory, files)
	return found.(*trackedFiles)
}

// trackedManifestDirectory is where a source with no .git carries git's answers about each of its repositories, at the
// source's root: the tree's own listing at .tracked/files, a submodule's at .tracked/<its path>/files (Loom's
// builder/tracked.go writes them, `git ls-tree -r -z --full-tree HEAD` byte for byte).
const trackedManifestDirectory = ".tracked"

// manifestListing is directory's repository's listing from the manifest at or above it, and when it was written (its
// modification time), when the manifest records directory as a repository; a directory it doesn't record isn't one.
func manifestListing(directory string) ([]byte, time.Time, bool) {
	for root := directory; ; root = filepath.Dir(root) {
		if info, err := os.Stat(filepath.Join(root, trackedManifestDirectory)); err == nil && info.IsDir() {
			relative, err := filepath.Rel(root, directory)
			if err != nil || relative == trackedManifestDirectory || strings.HasPrefix(relative, trackedManifestDirectory+string(filepath.Separator)) {
				return nil, time.Time{}, false
			}
			file, err := os.Open(filepath.Join(root, trackedManifestDirectory, relative, "files"))
			if err != nil {
				return nil, time.Time{}, false
			}
			defer file.Close()
			info, err := file.Stat()
			if err != nil {
				return nil, time.Time{}, false
			}
			listing, err := io.ReadAll(file)
			return listing, info.ModTime(), err == nil
		}
		if filepath.Dir(root) == root {
			return nil, time.Time{}, false
		}
	}
}

var tools sync.Map

// Tool names a tool by its own report, such as Tool("clang", "--version"), running it once per process. A tool
// that can't run is named by its error, so the build that needs it fails rather than reusing another's product.
//
// The report names what the tool is, never where or how it was found, so one tool on two machines reports the same
// (#t37sw0f). It is the tool's standard output (its standard error only when it writes nothing else, as clang -v
// does), so a note on standard error, such as go's "go: downloading go1.27.1" the first time a machine meets the
// tree's release, never enters a key. The tool is labelled by its file name, not the path it was called by, and where
// it is installed reads as <name>: toolReport drops InstalledDir and installedAs rewrites the install prefix.
// Tool("go", "env", names...) reports each variable as goSetting keys it (gobuild.go): the release a GOTOOLCHAIN
// selects rather than the policy that selected it, the workspace's content rather than its path, and no locations.
//
// A bare name is the program toolchain.Path resolves it to (under ADAMIC_TOOLS first, as a box's PATH has it), the one a
// build asking toolchain runs, and it runs under toolchain.Env, the environment that build runs its tools with: run
// under the caller's whole environment, the report could name another tool than the one that built (Tests, from
// Lint's review of Compiler's rows, Oct 12: native's clang --version; refusalrewrite's go, whose GOTOOLCHAIN a go env
// file set for the key and GOENV=off took from the build).
func Tool(name string, arguments ...string) string {
	command := strings.Join(append([]string{name}, arguments...), " ")
	entry, _ := tools.LoadOrStore(command, &toolRun{})
	report := entry.(*toolRun)
	// Parallel tests asking at once wait on the first one's run rather than each running the tool.
	report.once.Do(func() {
		label := strings.Join(append([]string{filepath.Base(name)}, arguments...), " ")
		keying(func() { report.value = toolValue(name, label, arguments, toolchain.Env()) })
	})
	return report.value
}

// A tool busy as it is executed is asked again, up to toolBusyAttempts times toolBusyWait apart.
const (
	toolBusyAttempts = 50
	toolBusyWait     = 20 * time.Millisecond
)

// A toolRun is one tool's report, run once however many tests ask for it at the same time.
type toolRun struct {
	once  sync.Once
	value string
}

// toolValue runs a tool for Tool: what it reads is the key's, never a build's. environment is the whole environment it
// runs with.
func toolValue(name, label string, arguments, environment []string) string {
	var value string
	if filepath.Base(name) == "go" && len(arguments) > 0 && arguments[0] == "env" {
		settings, _, err := goSettings("", environment, arguments[1:])
		value = label + ": " + strings.Join(settings, "\n")
		if err != nil {
			value += " (" + err.Error() + ")"
		}
	} else {
		var output, diagnostics bytes.Buffer
		path, err := toolchain.Path(name)
		if err != nil {
			// As a run that couldn't start reports it: no output, then the error.
			return label + ":  (" + err.Error() + ")"
		}
		// A tool written a moment ago can be busy for a moment more (ETXTBSY): a process forked meanwhile, by a test
		// running beside this one, holds the writer's descriptor until it execs, and exec refuses a file open for
		// writing. That says nothing of the tool, so it is asked again rather than named by the error and its path
		// (landable-8, on Cloud: TestAToolsReportNamesTheToolNotItsPlace under a loaded unit).
		for attempt := 0; ; attempt++ {
			output.Reset()
			diagnostics.Reset()
			run := exec.Command(path, arguments...)
			run.Env = environment
			run.Stdout, run.Stderr = &output, &diagnostics
			if err = run.Run(); !errors.Is(err, syscall.ETXTBSY) || attempt == toolBusyAttempts {
				break
			}
			time.Sleep(toolBusyWait)
		}
		report := output.String()
		if strings.TrimSpace(report) == "" {
			report = diagnostics.String()
		}
		value = label + ": " + toolReport(installedAs(path, report))
		if err != nil {
			value += " (" + err.Error() + ")"
		}
	}
	return value
}

// toolReport is a tool's report without the lines that say only where it is installed: clang's --version ends
// with InstalledDir, which differs by home (/home/ahra, /home/cloud, /root on a Codex instance) for the same clang,
// and would key one product apart on every machine (Loom, Oct 9: a key that depends on the home directory is a
// reproducibility bug). Everything that names what the tool is (its version, target and thread model) stays.
func toolReport(output string) string {
	lines := []string{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "InstalledDir:") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// installedAs rewrites the tool's install prefix, the directory above the bin it runs from, to <name> in its report:
// clang -print-resource-dir says /opt/adamic-tools/llvm/lib/clang/20 on a cloud box and
// /home/ahra/adamic-tools/llvm/lib/clang/20 on Workshop for one clang, and both read <clang>/lib/clang/20. path is the
// program that ran, as toolchain.Path resolved it.
//
// The longer prefix is rewritten first. A runner runs $ADAMIC_TOOLS/bin/clang, a link to adamic-tools/llvm/bin/clang,
// and its prefix /home/ahra/adamic-tools is the start of the one clang reports, so rewriting it first left
// <clang>/llvm/bin where Workshop, running ~/.adamic-tools/bin/clang to the same clang, said <clang>/bin, and keyed
// every product that names clang apart (#94aa17s: typeaware's volume-type-symbol-native, 9c677aae81cf on Home and
// Cloud, 8b42c20f9b4a on Workshop).
func installedAs(path, report string) string {
	path, _ = filepath.Abs(path)
	prefixes := []string{}
	for _, executable := range []string{path, resolved(path)} {
		bin := filepath.Dir(executable)
		if executable == "" || filepath.Base(bin) != "bin" || filepath.Dir(bin) == string(filepath.Separator) {
			continue
		}
		prefixes = append(prefixes, filepath.Dir(bin))
	}
	sort.SliceStable(prefixes, func(i, j int) bool { return len(prefixes[i]) > len(prefixes[j]) })
	for _, prefix := range prefixes {
		report = replacePath(report, prefix, "<"+filepath.Base(path)+">")
	}
	return report
}

func cacheDirectory() (string, error) {
	directory, err := cacheLocation()
	if err != nil {
		return "", err
	}
	return directory, os.MkdirAll(directory, 0o755)
}

// describe is a product's inputs as its key reads them (its name, each flag and tool portable), beside it in the cache, so the
// same product's two descriptions from two machines diff to the input that keyed them apart.
func describe(root string, inputs Inputs) string {
	lines := []string{"name " + portable(root, inputs.Name)}
	for _, name := range inputs.Files {
		if !unkeyed(name) {
			lines = append(lines, "file "+name)
		}
	}
	for _, flag := range inputs.Flags {
		lines = append(lines, "flag "+portable(root, flag))
	}
	for _, tool := range inputs.Toolchain {
		lines = append(lines, "tool "+portable(root, tool))
	}
	return strings.Join(lines, "\n") + "\n"
}

// note appends a line about the store to $ADAMIC_BUILD_LOG when set, beside the census lines.
func note(format string, arguments ...any) {
	if path := os.Getenv("ADAMIC_BUILD_LOG"); path != "" {
		if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			fmt.Fprintf(file, format+"\n", arguments...)
			file.Close()
		}
	}
}

// record is the product's census line, 'build <name> <key12> hit|fetched|audited|miss|missing|off <seconds>', also
// appended to $ADAMIC_BUILD_LOG when set, so the gate can count every build as its own unit. missing is read mode's
// miss, which built nothing.
func record(name, key, outcome string, started time.Time) string {
	line := fmt.Sprintf("build %s %s %s %.2f", strings.ReplaceAll(name, " ", "_"), key[:min(12, len(key))], outcome, time.Since(started).Seconds())
	if path := os.Getenv("ADAMIC_BUILD_LOG"); path != "" {
		if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			file.WriteString(line + "\n")
			file.Close()
		}
	}
	return line
}
