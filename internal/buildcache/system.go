package buildcache

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func packages() *packageIndex {
	found, _ := packageIndexes.LoadOrStore(dpkgDirectory, &packageIndex{})
	return found.(*packageIndex)
}

func (index *packageIndex) owner(path string) (string, bool) {
	index.ownersOnce.Do(func() {
		index.owners = packageOwners(dpkgDirectory, loaderArchitecture)
	})
	owner, ok := index.owners[path]
	return owner, ok
}

// packageOwners is each file's package in the dpkg database at directory, read from its info/*.list in sorted order.
// Where two architectures' twins of one package list a path (a Mac guest holding libc6-dev:amd64 and libc6-dev:arm64,
// #qmjv4z4), the twin of architecture, the build's loader's, owns it, as the house's own box would say, never whichever
// list sorted last: the twins' headers are the same bytes, and the owner's md5sums still decide whether the file is
// valued by the package at all (systemValue). Any other path keeps the last list naming it, so a box with one owner per
// path values as it did.
func packageOwners(directory, architecture string) map[string]string {
	owners := map[string]string{}
	lists, _ := filepath.Glob(filepath.Join(directory, "info", "*.list"))
	sort.Strings(lists)
	loaderSuffix := ":" + debianArchitecture(architecture)
	for _, list := range lists {
		name := strings.TrimSuffix(filepath.Base(list), ".list")
		file, err := os.Open(list)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if held, found := owners[scanner.Text()]; found && strings.HasSuffix(held, loaderSuffix) && !strings.HasSuffix(name, loaderSuffix) &&
				packageName(held) == packageName(name) {
				continue
			}
			owners[scanner.Text()] = name
		}
		file.Close()
	}
	return owners
}

// packageName is an owner's package without its architecture: libc6-dev for libc6-dev:arm64.
func packageName(owner string) string {
	name, _, _ := strings.Cut(owner, ":")
	return name
}

// debianArchitecture is Go's name for an architecture as dpkg spells it, where the two differ.
func debianArchitecture(architecture string) string {
	if spelled, ok := map[string]string{"386": "i386"}[architecture]; ok {
		return spelled
	}
	return architecture
}

func (index *packageIndex) version(name string) (string, bool) {
	index.versionsOnce.Do(func() {
		index.versions, index.conffiles = map[string]string{}, map[string]string{}
		file, err := os.Open(filepath.Join(dpkgDirectory, "status"))
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
		if content, err := os.ReadFile(filepath.Join(dpkgDirectory, "info", owner+".md5sums")); err == nil {
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

type contentStamp struct {
	path, kind        string
	inode, size, time int64
}

var systemContents sync.Map

// systemValue is a machine file as a build reads it: absent, a link by where it points, a directory by its names, a
// file a package installed and nobody changed since by the package and its version, any other file by its content.
func systemValue(path string) string {
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
	if owner, recorded := packageOwner(path); owner != "" {
		if sum := packages().installed(owner, recorded); sum != "" && hashed("md5") == sum {
			return "package " + owner + " " + packageVersion(owner)
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
