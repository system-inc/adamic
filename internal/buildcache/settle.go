package buildcache

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Settle turns a traced run into keyed products (#vt46geg). The trace is strace -f's output (cmd/traced's flags:
// every call naming a path, directory listings and forks, failed calls too), read to its end; the journal is what the
// run's processes wrote under directory. Each build is given what its process and every process it started read,
// outside keying, up to the build's end: the build's own reads and whatever the test read before it (bytes a test
// hands its build), and the reads of builds running beside it in the same process, a superset Kirk accepted. Each
// read is classed, and a build is refused, named, when one can't be keyed: an untracked file, .git, content in the
// home directory, a write in the tree, a path outside every known place, a lost trace. A refused build is never
// placed, recorded or published: it is Loom's to fix, never the author's red. The rest are keyed by their read sets
// and the recipe (the test binary's linked packages), placed under their keys, recorded beside them as read sets,
// and published when the shared store is on.
//
// A read set is valued when the build settles, not when it read, so the tree must be the tree it read: before is
// TreeState when the run began, and a run whose tree state differs at its end (a commit, a checkout, an edit, a
// submodule moved) refuses every build, as does a run any of whose processes wrote into the tree.
func Settle(directory string, trace io.Reader, before string) (Settlement, error) {
	root, err := repositoryRoot()
	if err != nil {
		return Settlement{}, err
	}
	cache, err := cacheDirectory()
	if err != nil {
		return Settlement{}, err
	}
	facts, err := settleFacts(root, cache, directory)
	if err != nil {
		return Settlement{}, err
	}
	state := newTraceState(facts)
	if err = state.read(trace); err != nil {
		return Settlement{}, err
	}
	entries, err := readJournal(directory)
	if err != nil {
		return Settlement{}, err
	}
	if after, err := TreeState(); err != nil || after != before {
		state.lost = append(state.lost, "the tree changed during the run (git HEAD, status or a submodule), so its reads can't be valued now")
	}
	for _, relative := range slices.Sorted(maps.Keys(state.treeWrites)) {
		state.lost = append(state.lost, "the run wrote into the tree (tree write: "+relative+")")
	}
	return state.settle(entries), nil
}

// TreeState is the repository's state as git sees it, and the machine's packages: HEAD, every file's status
// (untracked ones too, but go.work.sum, which go writes as it runs), each submodule's HEAD and status, the content of
// every modified and untracked file in each (status names a dirty file, never what it holds, so a file dirty before the
// run and edited again during it would read the same), and dpkg's record of what is installed (an upgrade during the
// run). cmd/traced takes it before and after a run.
func TreeState() (string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return "", err
	}
	var state strings.Builder
	keying(func() {
		git := func(directory string, arguments ...string) string {
			if err != nil {
				return ""
			}
			var output []byte
			if output, err = exec.Command("git", append([]string{"-C", directory}, arguments...)...).Output(); err != nil {
				err = fmt.Errorf("git -C %s %s: %v", directory, strings.Join(arguments, " "), err)
			}
			return string(output)
		}
		keep := func(output string) {
			for _, line := range strings.Split(output, "\n") {
				if !strings.HasSuffix(line, "go.work.sum") {
					state.WriteString(line + "\n")
				}
			}
		}
		repositories := []string{root}
		keep(git(root, "rev-parse", "HEAD"))
		submodules := git(root, "submodule", "status", "--recursive")
		keep(submodules)
		for _, line := range strings.Split(submodules, "\n") {
			if fields := strings.Fields(line); len(fields) >= 2 {
				repositories = append(repositories, filepath.Join(root, fields[1]))
			}
		}
		for _, repository := range repositories {
			if _, missing := os.Stat(filepath.Join(repository, ".git")); missing != nil {
				continue
			}
			state.WriteString("repository " + strings.TrimPrefix(repository, root) + "\n")
			keep(git(repository, "status", "--porcelain=v2", "-uall"))
			for _, name := range strings.Split(git(repository, "ls-files", "-m", "-o", "--exclude-standard", "-z"), "\x00") {
				if name == "" || unkeyed(name) {
					continue
				}
				if content, readError := os.ReadFile(filepath.Join(repository, name)); readError == nil {
					state.WriteString("worktree " + name + " " + blobID(content, 40) + "\n")
				} else {
					state.WriteString("worktree " + name + " unreadable\n")
				}
			}
		}
		if status, readError := os.ReadFile(filepath.Join(dpkgDirectory, "status")); readError == nil {
			sum := sha256.Sum256(status)
			state.WriteString("dpkg status " + hex.EncodeToString(sum[:]) + "\n")
		}
	})
	return state.String(), err
}

// A Settlement is what Settle did with each build, in the order the builds ended.
type Settlement struct {
	Settled []string
	Refused []string
}

// Read kinds, as bits: what a key must name about a path.
const (
	readContent = 1 << iota
	readExec
	readStat
	readAbsent
	readListing
	readWrite
)

type traceFacts struct {
	root, cache, trace, goRoot, goCache, goModules, tools, home string
	roots, caches, temporary, kernel, system                    []string
	searched, executables                                       map[string]bool
	// ran is every program a process of the run executed, as the trace names it.
	ran map[string]bool
}

// settleFacts are the places a read can be in, as this machine (the one that ran the trace) spells them.
func settleFacts(root, cache, trace string) (*traceFacts, error) {
	facts := &traceFacts{root: root, cache: cache, trace: trace, tools: toolsRoot(), searched: map[string]bool{}, executables: map[string]bool{}, ran: map[string]bool{}}
	// A path a descriptor decodes to is resolved; one a call names isn't. The tree and the cache are matched either way.
	facts.roots, facts.caches = variants(root), variants(cache)
	output, err := exec.Command("go", "env", "-json", "GOROOT", "GOCACHE", "GOMODCACHE").Output()
	if err != nil {
		return nil, fmt.Errorf("go env: %v", err)
	}
	var values map[string]string
	if err = json.Unmarshal(output, &values); err != nil {
		return nil, err
	}
	facts.goRoot, facts.goCache, facts.goModules = values["GOROOT"], values["GOCACHE"], values["GOMODCACHE"]
	facts.home, _ = os.UserHomeDir()
	for _, place := range []string{os.TempDir(), "/tmp", "/var/tmp"} {
		facts.temporary = append(facts.temporary, variants(place)...)
	}
	facts.kernel = []string{"/proc", "/sys", "/dev", "/run"}
	for _, place := range systemPlaces {
		facts.system = append(facts.system, variants(place)...)
	}
	for _, entry := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.IsAbs(entry) {
			facts.searched[filepath.Clean(entry)] = true
		}
	}
	return facts, nil
}

// inTree is name relative to the tree, when it is in it.
func (facts *traceFacts) inTree(name string) (string, bool) {
	for _, root := range facts.roots {
		if name == root {
			return ".", true
		}
		if under(name, root) {
			return strings.TrimPrefix(name, root+"/"), true
		}
	}
	return "", false
}

func isGit(relative string) bool {
	return relative == ".git" || strings.HasPrefix(relative, ".git/") || strings.Contains(relative, "/.git/") || strings.HasSuffix(relative, "/.git")
}

func variants(place string) []string {
	if real := resolved(place); real != "" && real != place {
		return []string{place, real}
	}
	return []string{place}
}

func under(path, place string) bool {
	return place != "" && (path == place || strings.HasPrefix(path, strings.TrimSuffix(place, "/")+"/"))
}

// classed is one read as a key names it: kind is an entry kind (content, listing, exists, above, home, product,
// tool, toolset, go, filesystem), a refusal (untracked, git, home, tree write, temp, other, local product), or "" for a read no key
// needs (the kernel, the machine's system files, Go's own cache, the build cache's bookkeeping).
type classed struct {
	kind, detail string
}

const refusal = "refuse:"

func (facts *traceFacts) classify(name string, kinds int) []classed {
	if found, through := facts.throughLink(name, kinds); through {
		return found
	}
	return []classed{facts.classifyPath(name, kinds)}
}

// throughLink classes a path in the tree that passes through a symbolic link: the first link, read for where it
// points, and what the path resolves to, classed in its place (a target outside the tree is classed as what it is).
func (facts *traceFacts) throughLink(name string, kinds int) ([]classed, bool) {
	for _, root := range facts.roots {
		if !under(name, root) || name == root {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(name, root+"/"), "/")
		for index := range parts {
			relative := strings.Join(parts[:index+1], "/")
			info, err := os.Lstat(filepath.Join(root, relative))
			if err != nil {
				return nil, false
			}
			if info.Mode()&fs.ModeSymlink == 0 {
				continue
			}
			found := []classed{{"link", relative}}
			if !Tracked(facts.root, filepath.Join(facts.root, relative), false) {
				found = []classed{{refusal + "untracked", relative}}
			}
			if real := resolved(name); real != "" && real != name {
				found = append(found, facts.classify(real, kinds)...)
			}
			return found, true
		}
		return nil, false
	}
	return nil, false
}

func (facts *traceFacts) classifyPath(name string, kinds int) classed {
	if kinds == readAbsent && facts.caseProbe(name) {
		return classed{"filesystem", "case"}
	}
	looked := kinds&^(readStat|readAbsent) == 0
	for _, root := range facts.roots {
		if !under(name, root) {
			continue
		}
		relative := "."
		if name != root {
			relative = strings.TrimPrefix(name, root+"/")
		}
		if unkeyed(relative) {
			return classed{}
		}
		if isGit(relative) {
			return classed{refusal + "git", relative}
		}
		// A write here never reaches a read set: any write into the tree refuses the whole run (Settle).
		// A path that was there when it was read (opened, listed or stat'ed: anything but a lookup that missed) must be
		// one the tree carries, whether or not it is still there: an installed node_modules is untracked, never absent.
		path := filepath.Join(facts.root, relative)
		if kinds&^readAbsent != 0 && relative != "." && !Tracked(facts.root, path, false) && !Tracked(facts.root, path, true) {
			return classed{refusal + "untracked", relative}
		}
		switch {
		case kinds&(readContent|readExec) != 0:
			return classed{"content", relative}
		case kinds&readListing != 0:
			return classed{"listing", relative}
		}
		return classed{"exists", relative}
	}
	for _, cache := range facts.caches {
		if !under(name, cache) {
			continue
		}
		if name == cache {
			return classed{}
		}
		parts := strings.Split(strings.TrimPrefix(name, cache+"/"), "/")
		switch {
		// Named under the cache as the journal names it, whichever spelling the trace used.
		case isKey(parts[0]):
			return classed{"product", filepath.Join(facts.cache, parts[0])}
		case parts[0] == "pending" && len(parts) >= 3 && isKey(parts[2]):
			return classed{"product", filepath.Join(facts.cache, "pending", parts[1], parts[2])}
		case parts[0] == "local" && len(parts) >= 2 && isKey(parts[1]):
			return classed{refusal + "local product", name}
		}
		return classed{}
	}
	switch {
	case under(name, facts.trace) || facts.executables[name] || under(name, traceMarker[:len(traceMarker)-1]):
		return classed{}
	case under(name, facts.goRoot):
		return classed{"go", "release"}
	case under(name, facts.goCache):
		return classed{}
	case under(name, facts.goModules):
		return classed{"module", name}
	case under(name, facts.tools):
		relative := strings.TrimPrefix(strings.TrimPrefix(name, facts.tools), "/")
		if relative == "" {
			return classed{}
		}
		if kinds&readExec != 0 {
			return classed{"tool", relative}
		}
		return classed{"toolset", strings.Split(relative, "/")[0]}
	}
	if looked {
		// A lookup that only asks whether a path exists: a search along PATH (a location, never keyed: what is found
		// is keyed by its report), a path in a directory above the tree (node's node_modules search, realpath's lstat of
		// each parent), which every machine answers of its own tree, or a walk down to another known place.
		if facts.searched[filepath.Dir(name)] {
			return classed{}
		}
		for _, root := range facts.roots {
			if under(root, filepath.Dir(name)) {
				relative, _ := filepath.Rel(root, name)
				return classed{"above", filepath.ToSlash(relative)}
			}
		}
		for _, place := range append(append([]string{facts.goRoot, facts.goCache, facts.goModules, facts.tools, facts.trace}, facts.caches...), facts.temporary...) {
			if place != "" && under(place, name) {
				return classed{}
			}
		}
	}
	for _, place := range facts.kernel {
		if under(name, place) {
			return classed{}
		}
	}
	for _, place := range facts.system {
		if under(name, place) {
			return classed{"system", name}
		}
	}
	for _, place := range facts.temporary {
		if under(name, place) {
			// What the build's own processes made there never reaches here (record leaves it out).
			if kinds&(readContent|readExec) != 0 {
				return classed{refusal + "temp not made by this run", name}
			}
			return classed{}
		}
	}
	if name == "/" {
		return classed{}
	}
	if under(name, facts.home) {
		if looked {
			return classed{"home", "~/" + strings.TrimPrefix(strings.TrimPrefix(name, facts.home), "/")}
		}
		return classed{refusal + "home", name}
	}
	return classed{refusal + "other", name}
}

// caseProbe says whether a lookup that missed was a program asking whether the file system is case-sensitive:
// typescript-go's osvfs stats its own executable with every letter's case swapped as it starts (os.go's
// fileSystemCaseSensitivity), in every process that links it, and the answer is the machine's, not a path's. Only a
// swap of a program the run executed is one; any other path in another case is the path it names.
func (facts *traceFacts) caseProbe(name string) bool {
	swapped := swapCase(name)
	return swapped != name && (facts.ran[swapped] || facts.executables[swapped])
}

// swapCase is name with each letter's case swapped, as osvfs swaps it.
func swapCase(name string) string {
	return strings.Map(func(r rune) rune {
		if upper := unicode.ToUpper(r); upper != r {
			return upper
		}
		return unicode.ToLower(r)
	}, name)
}

func isKey(name string) bool {
	if len(name) != 64 {
		return false
	}
	for _, character := range name {
		if !('0' <= character && character <= '9' || 'a' <= character && character <= 'f') {
			return false
		}
	}
	return true
}

// traceState follows a trace: which thread belongs to which process, each process's parent and working directory,
// which threads are keying, every process's reads outside keying, and each build's window.
type traceState struct {
	facts      *traceFacts
	process    map[int]int
	parent     map[int]int
	directory  map[int]string
	keying     map[int]int
	keyingTree map[int]bool
	reads      map[int]map[string]*pathReads
	writers    map[string][]writer
	treeWrites map[string]bool
	sequence   int
	waiting    map[int][]traceCall
	unfinished map[int]string
	open       map[string]int
	ended      map[string]map[string]int
	order      []string
	lost       []string
	started    bool
}

// A pathReads is a process's reads of one path: their kinds, and when it first read the path (0 for a path it only wrote).
type pathReads struct {
	kinds, firstRead int
}

// A writer is a process that wrote a path, and when it first did; it stays a writer after it ends.
type writer struct {
	process, sequence int
}

type traceCall struct {
	pid                    int
	call, arguments, value string
	opened, errno          string
}

func newTraceState(facts *traceFacts) *traceState {
	state := &traceState{facts: facts, process: map[int]int{}, parent: map[int]int{}, directory: map[int]string{}, keying: map[int]int{},
		keyingTree: map[int]bool{}, reads: map[int]map[string]*pathReads{}, writers: map[string][]writer{}, treeWrites: map[string]bool{}, waiting: map[int][]traceCall{},
		unfinished: map[int]string{}, open: map[string]int{}, ended: map[string]map[string]int{}}
	return state
}

var (
	traceLine   = regexp.MustCompile(`^(\d+)\s+(.*)$`)
	resumedLine = regexp.MustCompile(`^<\.\.\. (\w+) resumed>\s?(.*)$`)
	callLine    = regexp.MustCompile(`^(\w+)\((.*)\)\s+=\s+(-?\d+|\?)(?:<([^>]*)>)?(?:\s+(E[A-Z0-9]+))?`)
	exitLine    = regexp.MustCompile(`^\+\+\+ (?:exited with \d+|killed by \w+(?: \(core dumped\))?) \+\+\+$`)
	directoryFD = regexp.MustCompile(`^(AT_FDCWD|-?\d+)(?:<([^>]*)>)?,\s*`)
	quoted      = regexp.MustCompile(`^"((?:[^"\\]|\\.)*)"`)
	fdPath      = regexp.MustCompile(`^-?\d+<([^>]*)>`)
)

var (
	atCalls = set("openat", "openat2", "newfstatat", "fstatat64", "statx", "faccessat", "faccessat2", "readlinkat", "execveat", "mkdirat",
		"unlinkat", "renameat", "renameat2", "utimensat", "fchmodat", "fchownat", "mknodat", "linkat", "name_to_handle_at", "inotify_add_watch")
	plainCalls = set("open", "creat", "stat", "lstat", "stat64", "lstat64", "access", "readlink", "execve", "chdir", "mkdir", "rmdir", "unlink",
		"rename", "truncate", "chmod", "chown", "lchown", "utime", "utimes", "link", "statfs", "getxattr", "lgetxattr", "listxattr", "llistxattr",
		"chroot", "mknod", "acct", "swapon")
	statCalls = set("newfstatat", "fstatat64", "statx", "stat", "lstat", "stat64", "lstat64", "faccessat", "faccessat2", "access", "readlink",
		"readlinkat", "statfs", "getxattr", "lgetxattr", "listxattr", "llistxattr")
	writeCalls = set("mkdir", "mkdirat", "rmdir", "unlink", "unlinkat", "rename", "renameat", "renameat2", "truncate", "chmod", "fchmodat",
		"chown", "lchown", "fchownat", "utime", "utimes", "utimensat", "link", "linkat", "symlink", "symlinkat", "mknod", "mknodat")
	forkCalls = set("clone", "clone3", "fork", "vfork")
)

func set(names ...string) map[string]bool {
	found := map[string]bool{}
	for _, name := range names {
		found[name] = true
	}
	return found
}

func (state *traceState) read(trace io.Reader) error {
	scanner := bufio.NewScanner(trace)
	scanner.Buffer(make([]byte, 1<<20), 1<<26)
	for scanner.Scan() {
		state.line(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for pid, calls := range state.waiting {
		if len(calls) > 0 {
			state.lost = append(state.lost, fmt.Sprintf("a lost trace: process %d's %d calls, which the trace never placed under a parent", pid, len(calls)))
		}
	}
	for id := range state.open {
		state.lost = append(state.lost, "a lost trace: build "+id+" never ended in the trace")
	}
	sort.Strings(state.lost)
	return nil
}

func (state *traceState) line(text string) {
	match := traceLine.FindStringSubmatch(text)
	if match == nil {
		return
	}
	pid, _ := strconv.Atoi(match[1])
	rest := match[2]
	if exitLine.MatchString(rest) {
		state.exited(pid)
		return
	}
	if strings.HasSuffix(rest, "<unfinished ...>") {
		state.unfinished[pid] = strings.TrimSpace(strings.TrimSuffix(rest, "<unfinished ...>"))
		return
	}
	if resumed := resumedLine.FindStringSubmatch(rest); resumed != nil {
		start, ok := state.unfinished[pid]
		if !ok {
			return
		}
		delete(state.unfinished, pid)
		tail := resumed[2]
		if strings.HasPrefix(tail, ")") || strings.HasPrefix(tail, ",") {
			rest = start + tail
		} else {
			rest = start + " " + tail
		}
	}
	call := callLine.FindStringSubmatch(rest)
	if call == nil {
		return
	}
	state.handle(traceCall{pid: pid, call: call[1], arguments: call[2], value: call[3], opened: call[4], errno: call[5]})
}

func (state *traceState) handle(call traceCall) {
	if !state.started {
		// The first process in the trace is the command strace ran, in the directory it ran from.
		state.started = true
		state.process[call.pid] = call.pid
		state.reads[call.pid] = map[string]*pathReads{}
		state.directory[call.pid], _ = os.Getwd()
	}
	process, known := state.process[call.pid]
	if !known {
		state.waiting[call.pid] = append(state.waiting[call.pid], call)
		return
	}
	if forkCalls[call.call] {
		child, err := strconv.Atoi(call.value)
		if err != nil || child <= 0 {
			return
		}
		if strings.Contains(call.arguments, "CLONE_THREAD") {
			state.process[child] = process
		} else {
			state.process[child] = child
			state.parent[child] = process
			state.directory[child] = state.directory[process]
			state.reads[child] = map[string]*pathReads{}
			state.keyingTree[child] = state.keying[call.pid] > 0 || state.keyingTree[process]
		}
		waiting := state.waiting[child]
		delete(state.waiting, child)
		for _, earlier := range waiting {
			state.handle(earlier)
		}
		return
	}
	var name string
	kind := 0
	switch {
	case call.call == "fchdir":
		if match := fdPath.FindStringSubmatch(call.arguments); match != nil && call.value == "0" {
			state.directory[process] = match[1]
		}
		return
	case call.call == "getdents64" || call.call == "getdents":
		match := fdPath.FindStringSubmatch(call.arguments)
		if match == nil {
			return
		}
		name, kind = filepath.Clean(match[1]), readListing
	default:
		var rest string
		name, rest = target(call.call, call.arguments, state.directory[process])
		if name == "" {
			return
		}
		if strings.HasPrefix(name, traceMarker) {
			state.marker(call.pid, process, strings.TrimPrefix(name, traceMarker))
			return
		}
		failed := call.value == "-1"
		if failed && call.errno != "ENOENT" && call.errno != "ENOTDIR" {
			return
		}
		if call.call == "chdir" {
			if !failed {
				state.directory[process] = name
			}
			return
		}
		kind = readKind(call.call, rest, failed)
		if kind == readWrite || kind&readWrite != 0 {
			if second := secondTarget(call.call, rest, state.directory[process]); second != "" {
				state.record(call.pid, process, second, readWrite)
			}
		}
		if call.call == "symlink" || call.call == "symlinkat" {
			// The first argument is what the link says, not a path the call touches.
			return
		}
		// What an open resolved to, as the kernel names the descriptor: a read through a symbolic link reads its target.
		if opened := filepath.Clean(call.opened); call.opened != "" && filepath.IsAbs(opened) && opened != name && !strings.HasSuffix(opened, " (deleted)") {
			state.record(call.pid, process, opened, kind)
		}
	}
	state.record(call.pid, process, name, kind)
}

func readKind(call, rest string, failed bool) int {
	switch {
	case failed:
		return readAbsent
	case call == "execve" || call == "execveat":
		return readExec
	case statCalls[call]:
		return readStat
	case writeCalls[call]:
		return readWrite
	case call == "open" || call == "openat" || call == "openat2" || call == "creat":
		switch {
		case call == "creat" || strings.Contains(rest, "O_WRONLY") || strings.Contains(rest, "O_CREAT") || strings.Contains(rest, "O_TRUNC"):
			return readWrite
		case strings.Contains(rest, "O_DIRECTORY"):
			return readStat
		case strings.Contains(rest, "O_RDWR"):
			return readContent | readWrite
		}
		return readContent
	}
	return readStat
}

func (state *traceState) record(pid, process int, name string, kind int) {
	if kind&readExec != 0 {
		state.facts.ran[name] = true
	}
	if state.keying[pid] > 0 || state.keyingTree[process] {
		return
	}
	state.sequence++
	if kind&readWrite != 0 {
		if !slices.ContainsFunc(state.writers[name], func(other writer) bool { return other.process == process }) {
			state.writers[name] = append(state.writers[name], writer{process, state.sequence})
		}
		// A write into the tree by any process of the run means what was read isn't the tree that is keyed.
		if relative, ok := state.facts.inTree(name); ok && !unkeyed(relative) && !isGit(relative) {
			state.treeWrites[relative] = true
		}
	}
	if state.reads[process] == nil {
		state.reads[process] = map[string]*pathReads{}
	}
	found := state.reads[process][name]
	if found == nil {
		found = &pathReads{}
		state.reads[process][name] = found
	}
	found.kinds |= kind
	if kind&^readWrite != 0 && found.firstRead == 0 {
		found.firstRead = state.sequence
	}
}

func (state *traceState) marker(pid, process int, marker string) {
	event, id, _ := strings.Cut(marker, "/")
	switch event {
	case "key-begin":
		state.keying[pid]++
	case "key-end":
		if state.keying[pid] > 0 {
			state.keying[pid]--
		}
	case "build-begin":
		state.open[id] = process
	case "build-end":
		owner, ok := state.open[id]
		if !ok {
			return
		}
		delete(state.open, id)
		// Everything this process and the processes it started read so far, which is everything up to the build's end,
		// but what they wrote before reading it: that is the build's own output, never an input. Only the build's own
		// processes make output for it; what a sibling or a process above it wrote is an input like any other (a
		// sibling's product, a generated file). A product is always a product, whoever wrote it.
		snapshot := map[string]int{}
		for other, reads := range state.reads {
			if !state.descends(other, owner) {
				continue
			}
			for name, found := range reads {
				kinds := found.kinds
				if kinds&^readWrite != 0 && !underAny(name, state.facts.caches) && state.madeBy(owner, name, found.firstRead) {
					kinds &= readWrite
				}
				if kinds != 0 {
					snapshot[name] |= kinds
				}
			}
		}
		state.ended[id] = snapshot
		state.order = append(state.order, id)
	}
}

// madeBy says whether owner or a process it started wrote name, or a directory above it, before first reading it.
func (state *traceState) madeBy(owner int, name string, firstRead int) bool {
	for at := name; at != "/" && at != "."; at = filepath.Dir(at) {
		for _, wrote := range state.writers[at] {
			if wrote.sequence < firstRead && state.descends(wrote.process, owner) {
				return true
			}
		}
	}
	return false
}

func (state *traceState) descends(process, ancestor int) bool {
	for steps := 0; steps < 1<<16; steps++ {
		if process == ancestor {
			return true
		}
		parent, ok := state.parent[process]
		if !ok {
			return false
		}
		process = parent
	}
	return false
}

// exited folds an ended process's reads into its nearest live ancestor, which every build that could include them
// includes too, and forgets its threads, so a reused pid is placed by its own fork.
func (state *traceState) exited(pid int) {
	process, ok := state.process[pid]
	delete(state.process, pid)
	delete(state.keying, pid)
	if !ok || process != pid {
		return
	}
	reads := state.reads[pid]
	delete(state.reads, pid)
	delete(state.directory, pid)
	for ancestor, ok := state.parent[pid]; ok; ancestor, ok = state.parent[ancestor] {
		if into, live := state.reads[ancestor]; live {
			for name, found := range reads {
				kept := into[name]
				if kept == nil {
					into[name] = found
					continue
				}
				kept.kinds |= found.kinds
				if found.firstRead != 0 && (kept.firstRead == 0 || found.firstRead < kept.firstRead) {
					kept.firstRead = found.firstRead
				}
			}
			return
		}
	}
}

// target is the absolute path a call names and the rest of its arguments, or "".
func target(call, arguments, directory string) (string, string) {
	base := directory
	if atCalls[call] {
		match := directoryFD.FindStringSubmatch(arguments)
		if match == nil {
			return "", arguments
		}
		if match[2] != "" {
			base = match[2]
		}
		arguments = arguments[len(match[0]):]
	} else if !plainCalls[call] && call != "symlink" && call != "symlinkat" {
		return "", arguments
	}
	match := quoted.FindStringSubmatch(arguments)
	if match == nil {
		return "", arguments
	}
	name := unescape(match[1])
	rest := arguments[len(match[0]):]
	if name == "" && strings.Contains(rest, "AT_EMPTY_PATH") {
		if base == "" {
			return "", rest
		}
		return filepath.Clean(base), rest
	}
	if !filepath.IsAbs(name) {
		if base == "" {
			return "", rest
		}
		name = filepath.Join(base, name)
	}
	return filepath.Clean(name), rest
}

// secondTarget is the path a rename, link or symlink writes: its second path argument.
func secondTarget(call, rest, directory string) string {
	switch call {
	case "rename", "link", "symlink":
		rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest), ","))
		match := quoted.FindStringSubmatch(rest)
		if match == nil {
			return ""
		}
		name := unescape(match[1])
		if !filepath.IsAbs(name) && directory != "" {
			name = filepath.Join(directory, name)
		}
		return filepath.Clean(name)
	case "renameat", "renameat2", "linkat", "symlinkat":
		rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest), ","))
		name, _ := target("openat", rest, directory)
		return name
	}
	return ""
}

func unescape(text string) string {
	if !strings.Contains(text, `\`) {
		return text
	}
	if value, err := strconv.Unquote(`"` + text + `"`); err == nil {
		return value
	}
	return text
}

// readJournal is every line the run's processes journaled.
func readJournal(directory string) ([]journalEntry, error) {
	file, err := os.Open(filepath.Join(directory, "journal.jsonl"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var entries []journalEntry
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<26)
	for scanner.Scan() {
		var entry journalEntry
		if json.Unmarshal(scanner.Bytes(), &entry) == nil {
			entries = append(entries, entry)
		}
	}
	return entries, scanner.Err()
}

func (state *traceState) settle(entries []journalEntry) Settlement {
	facts := state.facts
	var settlement Settlement
	products := map[string]string{}
	processes := map[int]journalEntry{}
	built := map[string]journalEntry{}
	for _, entry := range entries {
		switch entry.Event {
		case "process":
			processes[entry.Process] = entry
			if entry.Executable != "" {
				facts.executables[entry.Executable] = true
			}
		case "found", "built":
			products[entry.Directory] = entry.NameKey
		}
		if entry.Event == "built" {
			built[entry.Build] = entry
		}
	}
	refuse := func(entry journalEntry, reason string) {
		settlement.Refused = append(settlement.Refused, fmt.Sprintf("refused %s (name key %s): %s", entry.Name, entry.NameKey[:min(12, len(entry.NameKey))], reason))
		// Never placed: a pending product that can't be keyed leaves the cache.
		if under(entry.Directory, filepath.Join(facts.cache, "pending")) {
			os.RemoveAll(entry.Directory)
			os.Remove(entry.Directory + ".inputs")
			os.Remove(entry.Directory + ".lock")
		}
	}
	recipes := map[string]recipe{}
	settled := map[string]bool{}
	for _, id := range state.order {
		entry, ok := built[id]
		if !ok {
			continue
		}
		settled[id] = true
		if len(state.lost) > 0 {
			refuse(entry, "the whole run can't be keyed: "+strings.Join(state.lost, "; "))
			continue
		}
		set, reason := state.measured(entry, state.ended[id], products)
		if reason == "" {
			own := processes[entry.Process]
			name := recipeName(own)
			if _, ok := recipes[name]; !ok {
				recipes[name] = recipeFor(facts.root, own)
			}
			set.Entries, reason = withRecipe(set.Entries, recipes[name])
		}
		if reason != "" {
			refuse(entry, reason)
			continue
		}
		set.Trace = filepath.Base(facts.trace)
		settingsOverride = recipes[recipeName(processes[entry.Process])].settings
		set, evaluated, product, err := recordUnion(facts.root, facts.cache, entry.NameKey, entry.Name, set, func(key string) string {
			return placeSettled(facts.cache, key, entry)
		})
		settingsOverride = ""
		if err != nil {
			refuse(entry, err.Error())
			continue
		}
		products[product] = entry.NameKey
		line := fmt.Sprintf("settled %s %s: %d reads", entry.Name, evaluated.key[:12], len(set.Entries))
		if set.Drift != nil {
			line += fmt.Sprintf("; declared Files: %d reads undeclared, %d declared unread", set.Drift.Undeclared, set.Drift.Unread)
		}
		// Only Workshop's tree builder publishes: the product, then its read set, so a set never names a product the
		// store doesn't hold.
		if publishing() {
			if err = publish(evaluated.key, entry.Name, product); err == nil {
				err = publishReads(facts.cache, entry.NameKey)
			}
			if err != nil {
				line += "; publish failed: " + err.Error()
			}
		}
		settlement.Settled = append(settlement.Settled, line)
	}
	for id, entry := range built {
		if !settled[id] {
			refuse(entry, "a lost trace: its build's markers aren't in the trace")
		}
	}
	os.Remove(filepath.Join(facts.cache, "pending", filepath.Base(facts.trace)))
	sort.Strings(settlement.Refused)
	return settlement
}

// measured is a build's reads as a read set, or the reason it can't be one.
func (state *traceState) measured(entry journalEntry, reads map[string]int, products map[string]string) (readSet, string) {
	facts := state.facts
	refused := map[string][]string{}
	entries := map[readEntry]bool{}
	modules, sums := false, false
	var treeReads []string
	for name, kinds := range reads {
		if kinds == readWrite && !underAny(name, facts.roots) {
			continue
		}
		for _, found := range facts.classify(name, kinds) {
			switch {
			case found.kind == "":
			case strings.HasPrefix(found.kind, refusal):
				refused[strings.TrimPrefix(found.kind, refusal)] = append(refused[strings.TrimPrefix(found.kind, refusal)], found.detail)
			case found.kind == "module":
				modules = true
			case found.kind == "product":
				if found.detail == entry.Directory {
					continue
				}
				nameKey, ok := products[found.detail]
				if !ok {
					refused["unknown product"] = append(refused["unknown product"], found.detail)
					continue
				}
				entries[readEntry{Kind: "product", Path: nameKey}] = true
			case found.kind == "go":
				entries[readEntry{Kind: "go", Path: "release"}] = true
			case found.kind == "system":
				// Valued by systemValue: a file dpkg installed, unchanged, by its package and version.
				entries[readEntry{Kind: "system", Path: found.detail}] = true
			default:
				entries[readEntry{Kind: found.kind, Path: found.detail}] = true
				if found.kind == "content" || found.kind == "listing" {
					treeReads = append(treeReads, found.detail)
				}
				if path.Base(found.detail) == "go.sum" {
					sums = true
				}
			}
		}
	}
	if modules && !sums {
		refused["module without go.sum"] = append(refused["module without go.sum"], "a module read, and no go.sum in what the build read")
	}
	if len(refused) > 0 {
		var classes []string
		for class, paths := range refused {
			sort.Strings(paths)
			paths = slices.Compact(paths)
			shown := paths
			if len(shown) > 6 {
				shown = append(shown[:6:6], fmt.Sprintf("and %d more", len(paths)-6))
			}
			classes = append(classes, class+": "+strings.Join(shown, ", "))
		}
		sort.Strings(classes)
		return readSet{}, "it read what no key can name, " + strings.Join(classes, "; ")
	}
	set := readSet{}
	for entry := range entries {
		set.Entries = append(set.Entries, entry)
	}
	set.Drift = declaredDrift(entry.Declared, treeReads)
	return set, ""
}

// recordUnion keys set on this tree, places the product by placing it under the key, and records the set first among
// nameKey's, all under the read sets' lock. An earlier set that still values as it was recorded came from a build of
// this same tree: its reads are ones this build could have made too (a build's reads can vary from run to run, and
// two settlements can race), so the set recorded and keyed is their union, and the earlier set gives way to it.
func recordUnion(root, cache, nameKey, name string, set readSet, placing func(key string) string) (readSet, evaluation, string, error) {
	var evaluated evaluation
	var product string
	err := withReads(cache, nameKey, func(file *readsFile) error {
		var kept []readSet
		for _, other := range file.Sets {
			if before, err := evaluate(root, cache, nameKey, other, 0); err == nil && len(differing(other, before)) == 0 {
				set.Entries = sortedEntries(set.Entries, other.Entries)
				continue
			}
			kept = append(kept, other)
		}
		var err error
		if evaluated, err = evaluate(root, cache, nameKey, set, 0); err != nil {
			return err
		}
		for index := range set.Entries {
			set.Entries[index].Value = shortValue(evaluated.values[index])
		}
		set.Key, set.Recorded = evaluated.key, now()
		product = placing(evaluated.key)
		file.Name, file.Sets = name, append([]readSet{set}, kept...)
		return nil
	})
	return set, evaluated, product, err
}

// placeSettled puts a settled build's product under its key and returns where it is: a product already there (the same
// bytes, from a build of the same reads) stays and the build's copy goes; an uncached build's directory on another
// file system stays where it was built.
func placeSettled(cache, key string, entry journalEntry) string {
	product := filepath.Join(cache, key)
	pending := under(entry.Directory, filepath.Join(cache, "pending"))
	switch _, err := os.Stat(product); {
	case err == nil:
		if pending {
			os.RemoveAll(entry.Directory)
		}
	case os.Rename(entry.Directory, product) != nil:
		return entry.Directory
	}
	if pending {
		os.Remove(entry.Directory + ".inputs")
		os.Remove(entry.Directory + ".lock")
	}
	os.WriteFile(product+".inputs", []byte("name key "+entry.NameKey+"\n"+entry.Inputs), 0o644)
	return product
}

func underAny(name string, places []string) bool {
	for _, place := range places {
		if under(name, place) {
			return true
		}
	}
	return false
}

// declaredDrift holds the declared Files against what the build read in the tree.
func declaredDrift(declared, reads []string) *drift {
	if len(declared) == 0 {
		return nil
	}
	covered := func(name, entry string) bool {
		return name == entry || strings.HasPrefix(name, strings.TrimSuffix(entry, "/")+"/")
	}
	result := &drift{}
	read := map[string]bool{}
	for _, name := range reads {
		declaredHere := false
		for _, entry := range declared {
			if covered(name, entry) {
				declaredHere, read[entry] = true, true
			}
		}
		if !declaredHere && name != "." {
			result.Undeclared++
			if len(result.UndeclaredPaths) < 20 {
				result.UndeclaredPaths = append(result.UndeclaredPaths, name)
			}
		}
	}
	for _, entry := range declared {
		if !read[entry] && !unkeyed(entry) {
			result.Unread++
			if len(result.UnreadPaths) < 20 {
				result.UnreadPaths = append(result.UnreadPaths, entry)
			}
		}
	}
	sort.Strings(result.UndeclaredPaths)
	return result
}

func recipeName(process journalEntry) string {
	return process.Package + " " + strings.Join(process.Settings, " ")
}

// recipe is the code a build ran in process, which strace can't see: every file of every package in the tree that the
// process's binary links (go list -deps, with -test for a test binary), each package's directory listing, each
// module's go.mod and go.sum, and the workspace file (Kirk's decision 1: the linked packages, never coverage).
type recipe struct {
	entries  []readEntry
	settings string
	reason   string
}

// recipeFor is recipeOf; a test in a tree of its own, which go list can't load as adamic, names its recipe itself.
var recipeFor = recipeOf

func recipeOf(root string, process journalEntry) recipe {
	pkg, test := strings.CutSuffix(process.Package, ".test")
	if pkg == "" || pkg == "command-line-arguments" {
		return recipe{reason: fmt.Sprintf("its process's binary (%s) names no package, so the code it ran can't be keyed", process.Executable)}
	}
	arguments := []string{"list", "-deps", "-json"}
	if test {
		arguments = append(arguments, "-test")
	}
	environment := os.Environ()
	for _, setting := range process.Settings {
		key, value, _ := strings.Cut(setting, "=")
		switch key {
		case "-tags":
			arguments = append(arguments, "-tags="+value)
		case "GOOS", "GOARCH", "CGO_ENABLED", "GOAMD64", "GOARM64", "GOEXPERIMENT":
			environment = append(environment, key+"="+value)
		}
	}
	command := exec.Command("go", append(arguments, pkg)...)
	command.Dir, command.Env = root, environment
	output, err := command.Output()
	if err != nil {
		return recipe{reason: fmt.Sprintf("go list of its binary's package %s: %v", pkg, err)}
	}
	entries := map[readEntry]bool{}
	var untracked []string
	add := func(kind, name string) {
		relative, err := filepath.Rel(root, name)
		if err != nil || strings.HasPrefix(relative, "..") || unkeyed(relative) {
			return
		}
		info, err := os.Stat(name)
		if err == nil && !Tracked(root, name, info.IsDir()) {
			untracked = append(untracked, filepath.ToSlash(relative))
			return
		}
		entries[readEntry{Kind: kind, Path: filepath.ToSlash(relative)}] = true
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var listed struct {
			Dir                                                                        string
			Standard                                                                   bool
			Module                                                                     *struct{ GoMod string }
			GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
			IgnoredGoFiles, TestGoFiles, XTestGoFiles                                  []string
		}
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return recipe{reason: "go list: " + err.Error()}
		}
		if listed.Standard || !inside(root, listed.Dir) {
			continue
		}
		add("listing", listed.Dir)
		for _, names := range [][]string{listed.GoFiles, listed.CgoFiles, listed.CFiles, listed.CXXFiles, listed.HFiles, listed.SFiles, listed.SysoFiles,
			listed.EmbedFiles, listed.IgnoredGoFiles, listed.TestGoFiles, listed.XTestGoFiles} {
			for _, name := range names {
				add("content", filepath.Join(listed.Dir, name))
			}
		}
		if listed.Module != nil && inside(root, listed.Module.GoMod) {
			for _, name := range existing(filepath.Dir(listed.Module.GoMod), "go.mod", "go.sum") {
				add("content", filepath.Join(filepath.Dir(listed.Module.GoMod), name))
			}
		}
	}
	for _, name := range existing(root, "go.work") {
		add("content", filepath.Join(root, name))
	}
	if len(untracked) > 0 {
		sort.Strings(untracked)
		return recipe{reason: "its binary links files git doesn't track: " + strings.Join(untracked[:min(6, len(untracked))], ", ")}
	}
	result := recipe{settings: settingsValue(process.Settings)}
	for entry := range entries {
		result.entries = append(result.entries, entry)
	}
	return result
}

// withRecipe adds the recipe to a build's entries, in the order every machine reads them: by path, then kind.
func withRecipe(entries []readEntry, code recipe) ([]readEntry, string) {
	if code.reason != "" {
		return nil, code.reason
	}
	return sortedEntries(entries, code.entries, []readEntry{{Kind: "settings", Path: "binary"}}), ""
}
