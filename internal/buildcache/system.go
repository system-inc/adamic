package buildcache

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"syscall"
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
		index.owners = map[string]string{}
		lists, _ := filepath.Glob(filepath.Join(dpkgDirectory, "info", "*.list"))
		sort.Strings(lists)
		for _, list := range lists {
			name := strings.TrimSuffix(filepath.Base(list), ".list")
			file, err := os.Open(list)
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
