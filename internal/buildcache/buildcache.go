// Package buildcache is the one home for build products tests need: the checker archive, stage 0, Go oracle
// binaries, a port's native build. Each product is addressed by what can change it, built once per address and
// reused, so a test's time starts after the fetch (@system_adamic's ruling on the 30 s rule, Oct 8).
//
// A product's address is its name key (its name, flags and tool reports) and the read set its last traced build
// measured, valued on the tree at hand (reads.go, #vt46geg): a change misses only the products that read it. A product
// is found by the cache's read sets, then the store's. Only Workshop's tree builder (ADAMIC_BUILD_STORE=traced, under
// cmd/traced) builds traced, and only a settled traced build is placed under its key or published (settle.go). Every
// other miss builds as before, for this machine alone: keyed by the declared Files, never published, never found by a
// read set, and never failing for want of a trace.
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
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Inputs are everything that can change a product. Name is one per build function. Files are repository-relative
// files and directories, hashed by content (a directory by every name and file under it, sorted). Flags are every
// flag and environment value the build reads. Toolchain names the tools, from Tool or runtime.Version().
type Inputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
}

// Product returns the directory holding the product built from inputs, failing the test if it can't be built.
func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {
	t.Helper()
	directory, line, err := get(inputs, build)
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
	directory, _, err := get(inputs, build)
	return directory, err
}

// ErrUntraced is a product Workshop's tree builder (ADAMIC_BUILD_STORE=traced) would build outside cmd/traced: what it
// publishes must be keyed by what its build read, so building untraced there is refused rather than published.
var ErrUntraced = errors.New("ADAMIC_BUILD_STORE=traced builds only under cmd/traced")

func get(inputs Inputs, build func(directory string) error) (string, string, error) {
	started := time.Now()
	if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
		directory, err := os.MkdirTemp("", "adamic-build-")
		if err != nil {
			return "", "", err
		}
		if traceDirectory() == "" {
			if err = build(directory); err != nil {
				return "", "", err
			}
			return directory, record(inputs.Name, "uncached", "off", started), nil
		}
		// Traced and uncached (main's own gate on Workshop): built fresh, and settled from where it was built.
		root, err := repositoryRoot()
		if err != nil {
			return "", "", err
		}
		nameKey, id, err := tracedBuild(root, inputs, directory, build)
		if err != nil {
			return "", "", err
		}
		journalBuilt(root, inputs, nameKey, id, directory)
		return directory, record(inputs.Name, "uncached", "off", started), nil
	}
	root, err := repositoryRoot()
	if err != nil {
		return "", "", err
	}
	cache, err := cacheDirectory()
	if err != nil {
		return "", "", err
	}
	var nameKey, described string
	var looked found
	keying(func() {
		nameKey, described = NameKey(root, inputs), describe(root, inputs)
		looked, err = productFor(root, cache, nameKey, 0)
	})
	if err != nil {
		return "", "", err
	}
	// The store's read sets for this name key join the cache's, so a machine with none of its own finds what the store
	// holds (#vt46geg).
	if looked.directory == "" && storeAddress() != "" {
		if remote := fetchReads(nameKey); len(remote) > 0 {
			if err = mergeReads(cache, nameKey, inputs.Name, remote); err != nil {
				return "", "", err
			}
			keying(func() { looked, err = productFor(root, cache, nameKey, 0) })
			if err != nil {
				return "", "", err
			}
		}
	}
	if looked.directory != "" {
		journal(journalEntry{Event: "found", NameKey: nameKey, Directory: looked.directory})
		return looked.directory, record(inputs.Name, looked.key, "hit", started), nil
	}
	if storeAddress() != "" {
		for _, key := range looked.keys {
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
			journal(journalEntry{Event: "found", NameKey: nameKey, Directory: product})
			return product, record(inputs.Name, key, outcome, started), nil
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
			journalBuilt(root, inputs, nameKey, id, product)
			outcome = "miss"
		} else {
			journal(journalEntry{Event: "found", NameKey: nameKey, Directory: product})
		}
		return product, record(inputs.Name, nameKey, outcome, started), nil
	case publishing():
		return "", record(inputs.Name, nameKey, "untraced", started), fmt.Errorf("product %s (name key %s): %s, and %w, so a build's reads become its key", inputs.Name, nameKey[:12], looked.why, ErrUntraced)
	}
	// Untraced, as every machine but Workshop's tree builder builds: for this machine alone, keyed by the declared
	// Files, never published and never found by a read set.
	var key string
	keying(func() { key, err = Key(root, inputs) })
	if err != nil {
		return "", "", err
	}
	built, err := filled(filepath.Join(cache, "local", key), described, build)
	if err != nil {
		return "", "", err
	}
	outcome := "hit"
	if built {
		outcome = "miss"
	}
	return filepath.Join(cache, "local", key), record(inputs.Name, key, outcome, started), nil
}

// tracedBuild builds into directory between markers naming the build, and returns its name key and id.
func tracedBuild(root string, inputs Inputs, directory string, build func(directory string) error) (string, string, error) {
	var nameKey string
	keying(func() { nameKey = NameKey(root, inputs) })
	id, err := traced(nameKey, func() error { return build(directory) })
	return nameKey, id, err
}

// journalBuilt journals a traced build where its product now is, for Settle to key and place.
func journalBuilt(root string, inputs Inputs, nameKey, id, directory string) {
	entry := journalEntry{Event: "built", Build: id, NameKey: nameKey, Directory: directory, Declared: inputs.Files, Root: root}
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
	lock, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return false, err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return false, err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
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
		return false, err
	}
	return true, nil
}

// Key is the product's address: a hash of every input, each length-prefixed so no two inputs run together.
func Key(root string, inputs Inputs) (string, error) {
	hash := sha256.New()
	field := func(kind, value string) {
		fmt.Fprintf(hash, "%s %d\n%s\n", kind, len(value), value)
	}
	field("buildcache", "v1")
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

type place struct{ path, name string }

// places are the directories portable rewrites, longest first, so a checkout inside the home directory becomes
// <repository> rather than ~/....: the repository, the build cache (a key that names another product by its path,
// such as estree's main=<build cache>/<key>/main.ts, names it by its key), the home directory, and each as its symbolic
// links resolve (macOS spells /tmp as /private/tmp).
func places(root string) []place {
	var found []place
	add := func(path, name string) {
		if !filepath.IsAbs(path) || filepath.Clean(path) == string(filepath.Separator) {
			return
		}
		for _, candidate := range []string{filepath.Clean(path), resolved(path)} {
			if candidate != "" && !slices.ContainsFunc(found, func(other place) bool { return other.path == candidate }) {
				found = append(found, place{candidate, name})
			}
		}
	}
	if root != "" {
		add(root, "<repository>")
	}
	if cache, err := cacheLocation(); err == nil {
		add(cache, "<build cache>")
	}
	if home, err := os.UserHomeDir(); err == nil {
		add(home, "~")
	}
	sort.SliceStable(found, func(i, j int) bool { return len(found[i].path) > len(found[j].path) })
	return found
}

func resolved(path string) string {
	if value, err := filepath.EvalSymlinks(path); err == nil {
		return value
	}
	return ""
}

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

func hashPath(hash io.Writer, root, name string, field func(kind, value string)) error {
	if filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..") {
		return fmt.Errorf("input %q must be inside the repository, relative to its root", name)
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
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(content)
			field("file", fmt.Sprintf("%s %t %x", relative, info.Mode()&0o111 != 0, sum))
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
	// objects are git's ids of the files the index holds, by path, with their modes; dirty are those whose working
	// copy differs from the index when taken, so a read set values a clean file by its id without reading it.
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

func repositoryAt(directory string) *trackedFiles {
	if found, ok := repositories.Load(directory); ok {
		return found.(*trackedFiles)
	}
	var files *trackedFiles
	if _, err := os.Lstat(filepath.Join(directory, ".git")); err == nil {
		files = &trackedFiles{paths: map[string]bool{}, directories: map[string]bool{}, objects: map[string]gitObject{}, dirty: map[string]bool{}, taken: time.Now()}
		var listed, dirty []byte
		var err error
		// git reads .git, which no product's read set may hold: listing what the tree carries is keying.
		keying(func() {
			if listed, err = exec.Command("git", "-C", directory, "ls-files", "-s", "-v", "-z").Output(); err == nil {
				dirty, err = exec.Command("git", "-C", directory, "diff-files", "--name-only", "-z").Output()
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
	}
	found, _ := repositories.LoadOrStore(directory, files)
	return found.(*trackedFiles)
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
func Tool(name string, arguments ...string) string {
	command := strings.Join(append([]string{name}, arguments...), " ")
	if value, ok := tools.Load(command); ok {
		return value.(string)
	}
	label := strings.Join(append([]string{filepath.Base(name)}, arguments...), " ")
	var value string
	keying(func() { value = toolValue(name, label, arguments) })
	tools.Store(command, value)
	return value
}

// toolValue runs a tool for Tool: what it reads is the key's, never a build's.
func toolValue(name, label string, arguments []string) string {
	var value string
	if filepath.Base(name) == "go" && len(arguments) > 0 && arguments[0] == "env" {
		settings, _, err := goSettings("", nil, arguments[1:])
		value = label + ": " + strings.Join(settings, "\n")
		if err != nil {
			value += " (" + err.Error() + ")"
		}
	} else {
		var output, diagnostics bytes.Buffer
		run := exec.Command(name, arguments...)
		run.Stdout, run.Stderr = &output, &diagnostics
		err := run.Run()
		report := output.String()
		if strings.TrimSpace(report) == "" {
			report = diagnostics.String()
		}
		value = label + ": " + toolReport(installedAs(name, report))
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
// /home/ahra/adamic-tools/llvm/lib/clang/20 on Workshop for one clang, and both read <clang>/lib/clang/20.
func installedAs(name, report string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return report
	}
	path, _ = filepath.Abs(path)
	for _, executable := range []string{path, resolved(path)} {
		bin := filepath.Dir(executable)
		if executable == "" || filepath.Base(bin) != "bin" || filepath.Dir(bin) == string(filepath.Separator) {
			continue
		}
		report = replacePath(report, filepath.Dir(bin), "<"+filepath.Base(name)+">")
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

// cacheLocation is where products live on this machine: ADAMIC_BUILD_CACHE_DIR, or the user cache directory's
// adamic-build.
func cacheLocation() (string, error) {
	if directory := os.Getenv("ADAMIC_BUILD_CACHE_DIR"); directory != "" {
		return filepath.Abs(directory)
	}
	user, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(user, "adamic-build"), nil
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
