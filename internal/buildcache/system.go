package buildcache

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/system-inc/adamic/internal/buildcache/location"
)

// What a build reads of the machine itself (#vt46geg, review 3): /usr, /lib, /bin, /etc. A linker from binutils, a
// libc header, the dynamic loader's cache all change what a build makes, and no tool report covers them all (clang's
// --version says nothing of libc6-dev), so they are keyed, never ignored, and a runner whose machine differs misses
// loudly rather than reading a product its machine wouldn't make. A file a package installed is keyed by the package
// and its version (dpkg's record), so one entry covers the hundreds of files a compile reads from it and the key moves
// when apt upgrades it; anything else (a symbolic link /etc/alternatives keeps, the loader's cache, a file no package
// owns) by its content, hashed once per path, inode, modification time and size. Ownership alone isn't trusted (review
// 2): a package's file is valued by the package only while its content is what dpkg installed (its md5sums, or the
// conffile hash in status), and by its content otherwise, so a file edited in place is keyed as edited.
//
// Not keyed: the kernel's places (/proc, /sys, /dev, /run), and Go's build cache, whose entries are addressed by the
// hashes of their own inputs (the Go files and toolchain a read set already names), so keying them would key every
// build on what was cached before it.
var systemPlaces = []string{"/etc", "/usr", "/lib", "/lib32", "/lib64", "/libx32", "/bin", "/sbin"}

// dpkgDirectory is dpkg's database: info/*.list (each package's files) and status (each package's version).
var dpkgDirectory = "/var/lib/dpkg"

// packageIndex is one dpkg database read: owners (each file's package, read only by Settle) and versions (each
// package's, read by every lookup), each once.
type packageIndex struct {
	ownersOnce, versionsOnce sync.Once
	owners, versions         map[string]string
	// conffiles are configuration files' md5s from status; sums are each package's md5sums, read when first asked.
	conffiles map[string]string
	sums      sync.Map
}

var packageIndexes sync.Map

// openPackageFile opens a file of dpkg's database: os.Open, a var so a test can see each one, and whether it ran as
// keying.
var openPackageFile = os.Open

func packages() *packageIndex {
	found, _ := packageIndexes.LoadOrStore(dpkgDirectory, &packageIndex{})
	return found.(*packageIndex)
}

func (index *packageIndex) owner(path string) (string, bool) {
	index.ownersOnce.Do(func() {
		index.owners = map[string]string{}
		lists, _ := filepath.Glob(filepath.Join(dpkgDirectory, "info", "*.list"))
		sort.Strings(lists)
		for _, list := range lists {
			name := strings.TrimSuffix(filepath.Base(list), ".list")
			file, err := openPackageFile(list)
			if err != nil {
				continue
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				index.owners[scanner.Text()] = name
			}
			file.Close()
		}
	})
	owner, ok := index.owners[path]
	return owner, ok
}

func (index *packageIndex) version(name string) (string, bool) {
	index.versionsOnce.Do(func() {
		index.versions, index.conffiles = map[string]string{}, map[string]string{}
		file, err := openPackageFile(filepath.Join(dpkgDirectory, "status"))
		if err != nil {
			return
		}
		defer file.Close()
		var name, architecture, version string
		flush := func() {
			if name != "" && version != "" {
				index.versions[name] = version
				index.versions[name+":"+architecture] = version
			}
			name, architecture, version = "", "", ""
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1<<20), 1<<24)
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case line == "":
				flush()
			case strings.HasPrefix(line, "Package: "):
				name = strings.TrimPrefix(line, "Package: ")
			case strings.HasPrefix(line, "Architecture: "):
				architecture = strings.TrimPrefix(line, "Architecture: ")
			case strings.HasPrefix(line, "Version: "):
				version = strings.TrimPrefix(line, "Version: ")
			case strings.HasPrefix(line, " /"):
				// A Conffiles line: " /etc/path md5".
				if fields := strings.Fields(line); len(fields) >= 2 {
					index.conffiles[fields[0]] = fields[1]
				}
			}
		}
		flush()
	})
	version, ok := index.versions[name]
	return version, ok
}

// installed is the md5 dpkg recorded for a file it installed: from the package's md5sums, or its conffile hash.
func (index *packageIndex) installed(owner, path string) string {
	index.version("")
	if sum, ok := index.conffiles[path]; ok {
		return sum
	}
	found, ok := index.sums.Load(owner)
	if !ok {
		sums := map[string]string{}
		if file, err := openPackageFile(filepath.Join(dpkgDirectory, "info", owner+".md5sums")); err == nil {
			content, _ := io.ReadAll(file)
			file.Close()
			for _, line := range strings.Split(string(content), "\n") {
				// "<md5>  <path without its leading slash>"
				if sum, name, ok := strings.Cut(line, "  "); ok {
					sums["/"+name] = sum
				}
			}
		}
		found, _ = index.sums.LoadOrStore(owner, sums)
	}
	return found.(map[string]string)[path]
}

// packageOwner is the package that installed a file, by the path read, the path it resolves to, or either with or
// without /usr in front (dpkg records /lib/... where a merged /usr reads /usr/lib/..., and the other way round).
func packageOwner(path string) (string, string) {
	index := packages()
	candidates := []string{path}
	if real := resolved(path); real != "" {
		candidates = append(candidates, real)
	}
	for _, candidate := range append([]string{}, candidates...) {
		if rest, ok := strings.CutPrefix(candidate, "/usr"); ok {
			candidates = append(candidates, rest)
		} else {
			candidates = append(candidates, "/usr"+candidate)
		}
	}
	for _, candidate := range candidates {
		if owner, ok := index.owner(candidate); ok {
			return owner, candidate
		}
	}
	return "", ""
}

func packageVersion(name string) string {
	if version, ok := packages().version(name); ok {
		return "version " + version
	}
	return "not installed"
}

// ldCache is the dynamic loader's cache, which ldconfig builds from every library the machine holds (a var so a test
// can value a cache of its own as the machine's).
var ldCache = "/etc/ld.so.cache"

// wslLibraries is where WSL mounts the Windows GPU driver's libraries, which ldconfig puts in ldCache: the driver is
// Windows', each box's own version, and no build loads it (a build that opens one refuses its product, classifyPath).
const wslLibraries = "/usr/lib/wsl"

// machineZone are the files that hold the zone a box was set to (Settle refuses a build that reads one, #cbw5jtf).
var machineZone = []string{"/etc/localtime", "/etc/timezone"}

// loaderDirectories are the dynamic loader's own search directories, which it tries for a soname its cache doesn't name
// (glibc's system directories on Debian and Ubuntu, with their glibc-hwcaps subdirectories): a library there is found
// whether or not the cache names it.
var loaderDirectories = []string{"/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu", "/lib", "/usr/lib", "/lib64", "/usr/lib64"}

// loaderFlags are the flags an entry must have, exactly, for the build's own loader to take it, by the architecture
// the build runs as: glibc's _dl_cache_check_flags, which on x86_64 is `(flags) == 0x303` (FLAG_ELF_LIBC6 with
// FLAG_X8664_LIB64, sysdeps/x86_64/dl-cache.h). An architecture not here has no rule this knows, and its cache doesn't
// read.
var loaderFlags = map[string]int32{"amd64": 0x303}

// loaderArchitecture is the architecture the build's loader runs as: this binary's own, an amd64 build under Rosetta in
// an arm64 guest included (a var so a test can ask as another).
var loaderArchitecture = runtime.GOARCH

// ldCacheValue is the loader's cache as what only it can find: the soname map's entries the build's own loader takes
// (loaderFlags), outside loaderDirectories and wslLibraries, each by flags, soname and path. A library in a loader
// directory is read by its own path when it's loaded, and a soname the cache misses is looked for in each loader
// directory, each a read of its own, so neither needs the cache's word; a library anywhere else (ld.so.conf.d's
// /usr/local/lib) is found through the cache alone, and keyed by it. So two boxes that installed different libraries
// in the loader's directories, or run different Windows GPU drivers, value it the same (#6eexfyv, measured Oct 12:
// Cloud's cache maps about 120 more libraries, QEMU's, and Server's maps libnvidia-ngx). An entry for another
// architecture is one this loader never takes: an arm64 Mac guest's ldconfig flags its own aarch64 libraries 0xa03,
// and an amd64 build there sees none of them (#q4z4cmj, Oct 17: every one of Marin's and Zach's guests' 456 and 439
// entries). It reads glibc's format, a header "glibc-ld.so.cache1.1" (48 bytes) and 24-byte entries whose string
// offsets are from the file's start; a cache in another format, one that doesn't read, or one for an architecture
// loaderFlags doesn't know, is false, and systemValue then calls it unreadable, with its reason and its content's hash,
// never a value a cache that reads could have.
func ldCacheValue(content []byte) (string, bool) {
	return ldCacheValueAs(content, loaderArchitecture)
}

// ldCacheValueAs is ldCacheValue for the loader of architecture: false for one loaderFlags has no rule for, never a
// guess at its flags (no flag at all would keep nothing, and value every cache as the empty one).
func ldCacheValueAs(content []byte, architecture string) (string, bool) {
	wanted, known := loaderFlags[architecture]
	if !known {
		return "", false
	}
	return ldCacheValueFor(content, wanted)
}

// ldCacheValueFor is ldCacheValue for a loader that takes entries flagged exactly wanted.
func ldCacheValueFor(content []byte, wanted int32) (string, bool) {
	const magic, header, entrySize = "glibc-ld.so.cache1.1", 48, 24
	if len(content) < header || string(content[:len(magic)]) != magic {
		return "", false
	}
	word := func(at int) uint32 {
		return uint32(content[at]) | uint32(content[at+1])<<8 | uint32(content[at+2])<<16 | uint32(content[at+3])<<24
	}
	text := func(at uint32) (string, bool) {
		if int(at) >= len(content) {
			return "", false
		}
		end := int(at)
		for end < len(content) && content[end] != 0 {
			end++
		}
		return string(content[at:end]), end < len(content)
	}
	count := int(word(len(magic)))
	if count < 0 || header+count*entrySize > len(content) {
		return "", false
	}
	lines := make([]string, 0, count)
	for index := range count {
		at := header + index*entrySize
		soname, ok := text(word(at + 4))
		if !ok {
			return "", false
		}
		library, ok := text(word(at + 8))
		if !ok {
			return "", false
		}
		flags := int32(word(at))
		if flags != wanted || library == wslLibraries || strings.HasPrefix(library, wslLibraries+"/") || inLoaderDirectory(library) {
			continue
		}
		lines = append(lines, fmt.Sprintf("%d %s %s", flags, soname, library))
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return "ld.so.cache " + hex.EncodeToString(sum[:]), true
}

// inLoaderDirectory says whether library sits directly in one of loaderDirectories or a glibc-hwcaps directory under one.
func inLoaderDirectory(library string) bool {
	directory := filepath.Dir(library)
	if parent, hwcaps := filepath.Split(filepath.Dir(directory)); hwcaps == "glibc-hwcaps" {
		directory = filepath.Clean(parent)
	}
	return slices.Contains(loaderDirectories, directory)
}

// houseValuesVariable is ADAMIC_HOUSE_VALUES: the house's own values for the machine's files, handed by Loom to every
// key maker on a house box (#shadow): a unit, build-tree's traced runs, Settle, the key check and the lean check, all
// alike. Handed, a machine file is valued as the house stamped it, and a file the stamp doesn't name by its content, so
// no key reads dpkg's database: under a unit that read reached the test's own set, and a traced run reading it can't
// be narrowed, since the database is in no stamp. Unhanded (a developer's machine, a box outside the house), a machine
// file is valued through dpkg as before, and those keys are that machine's alone.
const houseValuesVariable = "ADAMIC_HOUSE_VALUES"

// houseValuesBound is the most a handed file may hold: the stamp is about 80 KB, and a place's members a few thousand
// paths more.
const houseValuesBound = 16 << 20

// A houseStamp is a handed file's values: each machine file the house stamps by its own line, and each place it stamps by
// one digest, read once a process.
type houseStamp struct {
	paths  map[string]string
	places []housePlace
}

// A housePlace is a set of machine files the house stamps by one digest (#shadow, planner.HousePlace): the paths its
// stamper wrote a line for (members), and every path under a directory it holds whole (held) but one under a name it
// sets aside on some box (except). The stamper writes all of it into the handed file, so adamic reads the place's rule
// from the house, never a copy of the planner's that could drift from it.
type housePlace struct {
	name, digest string
	held, except []string
	members      map[string]bool
}

// covers says whether path is the place's: a member, or under a held directory and under none it sets aside, a probe
// that missed inside a held directory among them.
func (place housePlace) covers(path string) bool {
	under := func(directory string) bool { return path == directory || strings.HasPrefix(path, directory+"/") }
	if slices.ContainsFunc(place.except, under) {
		return false
	}
	return place.members[path] || slices.ContainsFunc(place.held, under)
}

// A handedHouse is one handed file's stamp, read once a process, or why the file is refused.
type handedHouse struct {
	once  sync.Once
	stamp *houseStamp
	err   error
}

var handedHouses sync.Map

// houseValues is the handed file's stamp, nil when none is handed, or why it's refused.
func houseValues() (*houseStamp, error) {
	path := os.Getenv(houseValuesVariable)
	if path == "" {
		return nil, nil
	}
	found, _ := handedHouses.LoadOrStore(path, &handedHouse{})
	house := found.(*handedHouse)
	house.once.Do(func() {
		keying(func() { house.stamp, house.err = readHouseValues(path) })
	})
	return house.stamp, house.err
}

// readHouseValues reads the file the house handed by its handle: opened never through a link, and refused unless it's a
// regular file the code under test can't have written (houseOwned), in a directory this process can't write, whose
// bytes hash to the sha256 its name carries (root names each copy <sha256>.json by the digest a signed manifest names,
// launch.HouseValuesDirectory), so nothing the reader trusts rests on how a user namespace shows the file's owner. A
// stamp naming no path is refused too, never read as one that values everything by content, and so is a place that
// isn't whole: a name not "place:", no digest, or a path that isn't absolute and clean.
func readHouseValues(path string) (*houseStamp, error) {
	descriptor, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("house values %s: %w", path, err)
	}
	file := os.NewFile(uintptr(descriptor), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("house values %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("house values %s: not a regular file (%s)", path, info.Mode().Type())
	}
	if why := houseOwned(info); why != "" {
		return nil, fmt.Errorf("house values %s: %s", path, why)
	}
	if why := houseDirectoryWritable(filepath.Dir(path)); why != "" {
		return nil, fmt.Errorf("house values %s: %s", path, why)
	}
	content, err := io.ReadAll(io.LimitReader(file, houseValuesBound+1))
	if err != nil {
		return nil, fmt.Errorf("house values %s: %w", path, err)
	}
	if len(content) > houseValuesBound {
		return nil, fmt.Errorf("house values %s: over %d bytes", path, houseValuesBound)
	}
	named, digest := strings.TrimSuffix(filepath.Base(path), ".json"), sha256.Sum256(content)
	if named != hex.EncodeToString(digest[:]) {
		return nil, fmt.Errorf("house values %s: its bytes hash to %x, not the sha256 its name carries", path, digest)
	}
	// The stamp's box and moment are never read: the values are the house's, alike on every box its check passed.
	var stamp struct {
		Paths  map[string]string `json:"paths"`
		Places map[string]struct {
			Digest  string   `json:"digest"`
			Held    []string `json:"held"`
			Except  []string `json:"except"`
			Members []string `json:"members"`
		} `json:"places"`
	}
	if err := json.Unmarshal(content, &stamp); err != nil {
		return nil, fmt.Errorf("house values %s: %w", path, err)
	}
	if len(stamp.Paths) == 0 {
		return nil, fmt.Errorf("house values %s: names no path", path)
	}
	handed := &houseStamp{paths: stamp.Paths}
	for _, name := range slices.Sorted(maps.Keys(stamp.Places)) {
		written := stamp.Places[name]
		if !strings.HasPrefix(name, "place:") || written.Digest == "" {
			return nil, fmt.Errorf("house values %s: place %q has no name or no digest", path, name)
		}
		place := housePlace{name: name, digest: written.Digest, held: written.Held, except: written.Except, members: map[string]bool{}}
		for _, member := range written.Members {
			place.members[member] = true
		}
		for _, named := range slices.Concat(written.Held, written.Except, written.Members) {
			if !filepath.IsAbs(named) || filepath.Clean(named) != named {
				return nil, fmt.Errorf("house values %s: place %s names %q, not an absolute, clean path", path, name, named)
			}
		}
		handed.places = append(handed.places, place)
	}
	return handed, nil
}

// houseOwners are the host's owners a handed file may have: root alone (Node's ruling on #shadow), since build-tree's
// product tests run as the builder's own user, and a file the code under test can write must never value a key. A var,
// so a test can hand a file of its own.
var houseOwners = []uint32{0}

// uidMapFile and overflowUidFile say how the host's owners show in this process's user namespace: a unit's test
// commands run in one mapping only the serving user (runner/modulecache_linux.go), where host root shows as the
// kernel's overflow uid. Vars, so a test can hand a namespace of its own.
var uidMapFile, overflowUidFile = "/proc/self/uid_map", "/proc/sys/kernel/overflowuid"

// ownersHere are the owners a handed file may show here: each of houseOwners as this namespace maps it, and the overflow
// uid when one of them isn't mapped (Node's ruling on #4wh8ybq). With no map to read (not Linux) the namespace is the
// host's. A namespace mapping its root to another host uid shows root's files as the overflow uid, never as 0, so a
// unit that is root inside can't pass a file of its own.
func ownersHere() ([]uint32, error) {
	content, err := os.ReadFile(uidMapFile)
	if errors.Is(err, fs.ErrNotExist) {
		return houseOwners, nil
	}
	if err != nil {
		return nil, err
	}
	var owners []uint32
	unmapped := false
	for _, owner := range houseOwners {
		mapped := false
		for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
			var inside, outside, count uint64
			if _, err := fmt.Sscan(line, &inside, &outside, &count); err != nil {
				return nil, fmt.Errorf("%s: %q: %w", uidMapFile, line, err)
			}
			if uint64(owner) >= outside && uint64(owner)-outside < count {
				owners, mapped = append(owners, uint32(inside+uint64(owner)-outside)), true
				break
			}
		}
		unmapped = unmapped || !mapped
	}
	if unmapped {
		overflow, err := os.ReadFile(overflowUidFile)
		if err != nil {
			return nil, err
		}
		var uid uint32
		if _, err := fmt.Sscan(string(overflow), &uid); err != nil {
			return nil, fmt.Errorf("%s: %w", overflowUidFile, err)
		}
		owners = append(owners, uid)
	}
	return owners, nil
}

// houseOwned is why a handed file's owner or mode refuses it, "" when neither does.
func houseOwned(info os.FileInfo) string {
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Sprintf("writable by group or other (%s)", info.Mode().Perm())
	}
	system, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "its owner can't be read"
	}
	owners, err := ownersHere()
	if err != nil {
		return fmt.Sprintf("its namespace's owners can't be read: %v", err)
	}
	if !slices.Contains(owners, system.Uid) {
		return fmt.Sprintf("owned by uid %d, not root", system.Uid)
	}
	return ""
}

// writeAccess is access(2)'s `W_OK`.
const writeAccess = 0x2

// houseDirectoryWritable is why a handed file's directory refuses it: one this process can write, where the code under
// test could have put the file, or one it can't ask about. "" when it can't write it.
func houseDirectoryWritable(directory string) string {
	switch err := syscall.Access(directory, writeAccess); {
	case err == nil:
		return fmt.Sprintf("its directory %s can be written by this process", directory)
	case errors.Is(err, syscall.EACCES), errors.Is(err, syscall.EROFS), errors.Is(err, syscall.EPERM):
		return ""
	default:
		return fmt.Sprintf("its directory %s can't be asked about: %v", directory, err)
	}
}

// systemEntryValue is a system entry's value: as the handed house values give it (houseValues), else systemValue's.
// A path a place covers is the place's, by its name and digest, first: the planner records a reach into a place as the
// place's one line, never a line a path, so a key and a narrow set part on the same files. Then a path the stamp names,
// by its line; then anything else by its content, never through dpkg.
func systemEntryValue(path string) (string, error) {
	stamp, err := houseValues()
	if err != nil {
		return "", err
	}
	if stamp == nil {
		return systemValue(path), nil
	}
	for _, place := range stamp.places {
		if place.covers(path) {
			return "house " + place.name + " " + place.digest, nil
		}
	}
	if value, ok := stamp.paths[path]; ok {
		return "house " + value, nil
	}
	return machineValue(path, false), nil
}

type contentStamp struct {
	path, kind        string
	inode, size, time int64
}

var systemContents sync.Map

// systemValue is a machine file as a build reads it: absent, a link by where it points, a directory by its names, a
// file a package installed and nobody changed since by the package and its version, any other file by its content.
func systemValue(path string) string {
	return machineValue(path, true)
}

// machineValue is systemValue, through dpkg's database only when packaged: unpackaged, every file is valued by its
// content, the way a file the handed house values don't name is (systemEntryValue).
func machineValue(path string, packaged bool) string {
	info, err := os.Lstat(path)
	switch {
	case err != nil:
		return "absent"
	case path == ldCache && info.Mode().IsRegular():
		// The loader's cache is valued only as the loader reads it. One that won't read that way is unreadable, a miss,
		// and never the empty cache's value: the loader may read what this can't (a newer format), and a library found
		// only through it would go unkeyed. Two that don't read value alike only when their bytes are the same, so its
		// reason and its content's hash are both in the value.
		content, err := os.ReadFile(path)
		if err != nil {
			return "unreadable: " + err.Error()
		}
		value, ok := ldCacheValue(content)
		if !ok {
			reason := "not glibc's cache format as this reads it"
			if _, known := loaderFlags[loaderArchitecture]; !known {
				reason = "no loader rule for " + loaderArchitecture
			}
			return fmt.Sprintf("unreadable: %s, content %x", reason, sha256.Sum256(content))
		}
		return value
	case info.Mode()&os.ModeSymlink != 0:
		target, _ := os.Readlink(path)
		return "link " + target
	case info.IsDir():
		entries, err := os.ReadDir(path)
		if err != nil {
			return "unreadable: " + err.Error()
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		sum := sha256.Sum256([]byte(strings.Join(names, "\n")))
		return "directory " + hex.EncodeToString(sum[:])
	case !info.Mode().IsRegular():
		return "other " + info.Mode().Type().String()
	}
	hashed := func(kind string) string {
		stamp := contentStamp{path: path, kind: kind, size: info.Size(), time: info.ModTime().UnixNano()}
		if system, ok := info.Sys().(*syscall.Stat_t); ok {
			stamp.inode = int64(system.Ino)
		}
		if value, ok := systemContents.Load(stamp); ok {
			return value.(string)
		}
		// Kept across processes in the build cache too, under the stamp: a file changed in place gets a new stamp. Read
		// where this process writes, then in the build cache (location.ReadPlaces), and kept where it writes.
		name := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d\x00%d\x00%d", stamp.path, stamp.kind, stamp.inode, stamp.size, stamp.time)))
		var kept string
		if output, err := outputLocation(); err == nil {
			kept = filepath.Join(output, "machine", hex.EncodeToString(name[:]))
			for _, place := range location.ReadPlaces(kept) {
				if value, err := os.ReadFile(place); err == nil && len(value) > 0 {
					systemContents.Store(stamp, string(value))
					return string(value)
				}
			}
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "unreadable: " + err.Error()
		}
		var value string
		if kind == "md5" {
			value = fmt.Sprintf("%x", md5.Sum(content))
		} else {
			value = fmt.Sprintf("%x", sha256.Sum256(content))
		}
		systemContents.Store(stamp, value)
		if kept != "" && os.MkdirAll(filepath.Dir(kept), 0o755) == nil {
			if temporary, err := os.CreateTemp(filepath.Dir(kept), ".value-"); err == nil {
				temporary.WriteString(value)
				temporary.Close()
				os.Rename(temporary.Name(), kept)
			}
		}
		return value
	}
	if packaged {
		if owner, recorded := packageOwner(path); owner != "" {
			if sum := packages().installed(owner, recorded); sum != "" && hashed("md5") == sum {
				return "package " + owner + " " + packageVersion(owner)
			}
		}
	}
	return fmt.Sprintf("file %t %s", info.Mode()&0o111 != 0, hashed("sha256"))
}

// settingsOverride, while Settle keys a build, is the build settings of the binary that built it; otherwise a key is
// valued with this binary's own (a runner runs Workshop's binaries, built alike).
var settingsOverride string

// binarySettings is what the binary was built with that changes what its code does (review 5): tags, ldflags,
// gcflags, -race, -asan, -msan, GOEXPERIMENT, CGO_ENABLED, the target, and the rest of the build info's settings but
// the commit, sorted and portable.
func binarySettings() string {
	if settingsOverride != "" {
		return settingsOverride
	}
	var settings []string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			settings = append(settings, setting.Key+"="+setting.Value)
		}
	}
	return settingsValue(settings)
}

func settingsValue(settings []string) string {
	var kept []string
	for _, setting := range settings {
		if !strings.HasPrefix(setting, "vcs") {
			root, _ := repositoryRoot()
			kept = append(kept, portable(root, setting))
		}
	}
	sort.Strings(kept)
	return strings.Join(kept, "\n")
}
