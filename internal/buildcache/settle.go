package buildcache

import (
	"bufio"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Settle turns a traced run into keyed products (#vt46geg). The trace is strace -f's output (cmd/traced's flags:
// every call naming a path, directory listings and forks, failed calls too), read to its end; the journal is what the
// run's processes wrote under directory. Each build is given what its process and every process it started read,
// outside keying, up to the build's end: the build's own reads and whatever the test read before it (bytes a test
// hands its build), and the reads of builds running beside it in the same process, a superset Kirk accepted. Each
// read is classed, and a build is refused, named, when one can't be keyed: an untracked file, .git, content in the
// home directory, a write in the tree, a path outside every known place, a lost trace. A refused build is never
// placed or recorded: it is Loom's to fix, never the author's red. The rest are keyed by their read sets and the recipe
// (the test binary's linked packages), placed under their keys, and recorded beside them as read sets, which Loom ships
// with the cache.
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
	settledTree.Store(&sync.Map{})
	defer settledTree.Store(nil)
	settlement := state.settle(entries)
	settlement.CommandExited, settlement.CommandStatus = state.commandStatus >= 0, state.commandStatus
	return settlement, nil
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
			if output, err = toolchain.GitCommand(append([]string{"-C", directory}, arguments...)...).Output(); err != nil {
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

// A Settlement is what Settle did with each build, in the order the builds ended: a line each for people, and a
// Report each for Loom's tree builder.
type Settlement struct {
	Settled []string
	Refused []string
	Reports []Report
	// CommandExited says the trace holds the command's own exit, with CommandStatus its status.
	CommandExited bool
	CommandStatus int
	// TestReads is the test read set's line (#sfqdxvy), "test read set <key>: <n> reads" or "no test read set: <why>",
	// when ADAMIC_TRACE_TEST_KEY asked for one; empty otherwise.
	TestReads string
}

// A Report is one product's outcome as cmd/traced writes it to ADAMIC_TRACE_REPORT, one JSON line each, for Loom's
// tree builder: settled under Key (with how many reads its set names), found already built under Key, or refused for
// Reason. When the whole run can't
// be keyed (a lost trace, the tree changed or written), one line more comes first, with outcome lost, its Reason and
// no product; every product of that run is refused too.
type Report struct {
	Name    string `json:"name,omitempty"`
	NameKey string `json:"nameKey,omitempty"`
	Key     string `json:"key,omitempty"`
	Outcome string `json:"outcome"`
	Reason  string `json:"reason,omitempty"`
	Reads   int    `json:"reads,omitempty"`
	// Window is how a settled build's reads were taken (#mmpt615's census): BuildWindowed, BuildLeanSuperset or
	// BuildSuperset.
	Window string `json:"window,omitempty"`
	// Lean is a product keyed by the lean recipe (leanRecipeOf), settled or found under a lean set (leanSet): the products
	// Loom's sampled key check draws from, hits as well as builds.
	Lean bool `json:"lean,omitempty"`
}

// How a settled build's reads were taken: a lean build (LeanProduct) by its own window and declared files; a lean build
// whose window the trace couldn't place, and every other build, by everything its process read up to its end, the
// superset Kirk accepted.
const (
	BuildWindowed     = "windowed"
	BuildLeanSuperset = "lean superset"
	BuildSuperset     = "superset"
)

// Report outcomes.
const (
	ReportSettled = "settled"
	ReportRefused = "refused"
	ReportFound   = "found"
	ReportLost    = "lost"
	// ReportReadSet is the run's test read set (#sfqdxvy), recorded under NameKey (ADAMIC_TRACE_TEST_KEY) with its
	// marker under Key: Loom vouches for the test with it only when cmd/find finds it under that very key.
	ReportReadSet = "readSet"
)

// WriteReport writes reports to file, one JSON line each, replacing what it held.
func WriteReport(file string, reports []Report) error {
	var lines []byte
	for _, report := range reports {
		line, err := json.Marshal(report)
		if err != nil {
			return err
		}
		lines = append(append(lines, line...), '\n')
	}
	return os.WriteFile(file, lines, 0o644)
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
	root, cache, trace, goRoot, goCache, goModules, goTelemetry, nativeRuntimes, tools, home string
	roots, caches, temporary, kernel, system, buildLogs                                      []string
	searched, executables                                                                    map[string]bool
	// ran is every program a process of the run executed, as the trace names it; opened every file a process of it read the
	// content of, which a case probe may ask about too (caseProbe).
	ran, opened map[string]bool
	// goCaches are goCache's spellings, as written and resolved, each without its leading slash (goCacheJoined).
	goCaches []string
	// cache is where the run wrote products, and caches is every spelling of every directory its lookups read them in:
	// the output directory and, when that is another, the build cache (location.ReadPlaces). cacheRoots is, for each
	// spelling, the directory it spells, as the journal names it.
	cacheRoots map[string]string
}

// settleFacts are the places a read can be in, as this machine (the one that ran the trace) spells them.
func settleFacts(root, cache, trace string) (*traceFacts, error) {
	tools, err := toolsRoot()
	if err != nil {
		return nil, err
	}
	facts := &traceFacts{root: root, cache: cache, trace: trace, tools: tools, searched: map[string]bool{}, executables: map[string]bool{}, ran: map[string]bool{}, opened: map[string]bool{},
		cacheRoots: map[string]string{}}
	// A path a descriptor decodes to is resolved; one a call names isn't. The tree and the cache are matched either way.
	// A read in the build cache through an output directory of this run's own is the cache's, as a read in that
	// directory is: placed in the one, classified in either.
	facts.roots = variants(root)
	for _, place := range location.ReadPlaces(cache) {
		for _, spelling := range variants(place) {
			facts.caches = append(facts.caches, spelling)
			facts.cacheRoots[spelling] = place
		}
	}
	// Where a product's go keeps things: toolchain's go, under the environment GoBuild gives it.
	output, err := toolchain.Command("go", "env", "-json", "GOROOT", "GOCACHE", "GOMODCACHE").Output()
	if err != nil {
		return nil, fmt.Errorf("go env: %v", err)
	}
	var values map[string]string
	if err = json.Unmarshal(output, &values); err != nil {
		return nil, err
	}
	facts.goRoot, facts.goCache, facts.goModules = values["GOROOT"], values["GOCACHE"], values["GOMODCACHE"]
	for _, spelling := range variants(facts.goCache) {
		if trimmed := strings.Trim(filepath.ToSlash(spelling), "/"); trimmed != "" {
			facts.goCaches = append(facts.goCaches, trimmed)
		}
	}
	// Go's telemetry, which every go command reads and counts into as it runs (its upload days, its counters), and which
	// GOENV=off doesn't turn off: go's own bookkeeping, as its build cache is, never an input to what it builds.
	if configuration, err := os.UserConfigDir(); err == nil {
		facts.goTelemetry = filepath.Join(configuration, "go", "telemetry")
	}
	// The native runtime cache (internal/native's cachedRuntime): each runtime under the address of everything its
	// compile read (runtimeKey: its sources, flags, compiler and version, platform, and the system headers it read), as
	// Go's build cache addresses an action, so what a build reads there is no entry.
	if user, err := os.UserCacheDir(); err == nil {
		facts.nativeRuntimes = filepath.Join(user, "adamic", "runtime")
	}
	facts.home, _ = os.UserHomeDir()
	// The census log buildcache appends a line to for every product it builds or finds (ADAMIC_BUILD_LOG): its own
	// bookkeeping, as the trace's journal is, never a product's input (Planner's shape B proof on Cloud, Oct 11: a build
	// tree's log under the home directory refused two estree products).
	// The trace's report (ADAMIC_TRACE_REPORT) is the same kind of bookkeeping, and a path every run names afresh: a
	// build that looks at the ADAMIC_ locations (Relative) looked at it (#n9jc61n, Planner's skip proof on Home, Oct 11:
	// markdown-list-layout-shared-v3 keyed its own run's logs/product-388.trace.jsonl).
	for _, variable := range []string{"ADAMIC_BUILD_LOG", "ADAMIC_TRACE_REPORT"} {
		if log := os.Getenv(variable); log != "" {
			facts.buildLogs = append(facts.buildLogs, variants(filepath.Clean(log))...)
		}
	}
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

// cacheRoot is the directory spelling, one of caches, spells, as the journal names it: the run's cache when facts
// don't know it.
func (facts *traceFacts) cacheRoot(spelling string) string {
	if root, ok := facts.cacheRoots[spelling]; ok {
		return root
	}
	return facts.cache
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
// tool, toolset, go, filesystem, absolute), a refusal (untracked, git, home, tree write, temp, other, local product), or "" for a read no key
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
			// A link inside a project's installed node_modules (npm's .bin) is the install's, which nodeInstallValue values
			// with each link by where it points; what it resolves to outside that node_modules is classed in its place, as
			// any read is (#4gp2ytz: stage3/api's .bin/tsc and .bin/tsserver refused the adapted tree as untracked).
			if project, ok := facts.nodeInstallOf(relative); ok {
				found := []classed{{nodeInstallKind, project}}
				install := filepath.Join(root, filepath.FromSlash(project), "node_modules")
				if real := resolved(name); real != "" && !under(real, install) && !under(real, resolved(install)) {
					found = append(found, facts.classify(real, kinds)...)
				}
				return found, true
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

// nodeInstallKind is a read set's entry for a project's installed node_modules, by the project's directory: npm's
// install, which Loom places in the tree from the project's lockfile (NodeProject), valued whole by nodeInstallValue
// (#qx8fkaf, Node's ruling). Any read under it, a lookup that missed included, is a read of the install.
const nodeInstallKind = "node install"

// nodeInstallOf is the project whose installed node_modules relative is in, or is: the directory above relative's first
// node_modules, when git tracks that directory's package-lock.json and not its node_modules. Its first, never a nested
// one: node_modules/a/node_modules is the same install, made by the same npm ci.
func (facts *traceFacts) nodeInstallOf(relative string) (string, bool) {
	parts := strings.Split(relative, "/")
	index := slices.Index(parts, "node_modules")
	if index < 0 {
		return "", false
	}
	project := "."
	if index > 0 {
		project = strings.Join(parts[:index], "/")
	}
	directory := filepath.Join(facts.root, filepath.FromSlash(project))
	if Tracked(facts.root, filepath.Join(directory, "node_modules"), true) || !Tracked(facts.root, filepath.Join(directory, "package-lock.json"), false) {
		return "", false
	}
	return project, true
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
		if unkeyed(relative) || kinds == readAbsent && facts.goCacheJoined(relative) {
			return classed{}
		}
		if isGit(relative) {
			return classed{refusal + "git", relative}
		}
		if project, ok := facts.nodeInstallOf(relative); ok {
			return classed{nodeInstallKind, project}
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
		// A lookup in the cache that missed read nothing, as a miss in Go's build cache reads nothing: namedProduct asking
		// whether a product is pending under this trace, or one whose build never finished (Planner's step 5 tree on Home,
		// Oct 11: three bridge-test-stage0 builds refused for "unknown product" at a pending path that wasn't there).
		if kinds == readAbsent {
			return classed{}
		}
		parts := strings.Split(strings.TrimPrefix(name, cache+"/"), "/")
		switch {
		// Named under the directory it was read in as the journal names it, whichever spelling the trace used: the output
		// directory's or the build cache's, where the lookup that found it found it.
		case isKey(parts[0]):
			return classed{"product", filepath.Join(facts.cacheRoot(cache), parts[0])}
		case parts[0] == "pending" && len(parts) >= 3 && isKey(parts[2]):
			return classed{"product", filepath.Join(facts.cacheRoot(cache), "pending", parts[1], parts[2])}
		case parts[0] == "local" && len(parts) >= 2 && isKey(parts[1]):
			return classed{refusal + "local product", name}
		}
		return classed{}
	}
	switch {
	case under(name, facts.trace) || facts.executables[name] || under(name, traceMarker[:len(traceMarker)-1]) || slices.Contains(facts.buildLogs, name):
		return classed{}
	case under(name, facts.goRoot):
		return classed{"go", "release"}
	case under(name, facts.goCache), under(name, facts.goTelemetry), under(name, facts.nativeRuntimes):
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
		aboveTree := slices.ContainsFunc(facts.roots, func(root string) bool { return under(root, name) })
		// A step of a walk down to a known place (realpath's lstat of each parent of the build cache) is the place's,
		// never keyed, unless it leads to the tree too, which above keys: where a run's cache sits isn't an input (#n9jc61n,
		// Planner's skip proof on Home, Oct 11: 12 products re-keyed between two runs of one tree, each naming its own
		// run's directory, ../trees-2 against ../trees-3).
		for _, place := range append(append([]string{facts.goRoot, facts.goCache, facts.goModules, facts.goTelemetry, facts.nativeRuntimes, facts.tools, facts.trace}, facts.caches...), facts.temporary...) {
			if place != "" && under(place, name) && !aboveTree {
				return classed{}
			}
		}
		// A lookup that climbs to home or past it lands where every machine answers alike wherever its tree sits, so it
		// is keyed where it lands, below: home's, the machine's C stack (/lib, /usr) or an absolute path. Keyed relative
		// to the tree, ../../../../lib is /lib at Workshop's depth and a path under ~/loom-serve at a runner's unit
		// source, so every product whose build looked up the C stack missed on a runner (Builder's dry B2, Oct 12: 29 of
		// 39 product packages moved in the key check). The tree's own parents, and what sits between home and the tree,
		// stay relative.
		landed := !aboveTree && facts.landsWhereMachinesAgree(filepath.Dir(name))
		for _, root := range facts.roots {
			if under(root, filepath.Dir(name)) && !landed {
				relative, _ := filepath.Rel(root, name)
				return classed{"above", filepath.ToSlash(relative)}
			}
		}
		// A path beside one of the build cache's parents (node's node_modules and package.json search, walking up from a
		// product it runs) is keyed as above keys the tree's, relative to the cache: each run's cache sits elsewhere,
		// and what is beside it is the same.
		for _, cache := range facts.caches {
			if under(cache, filepath.Dir(name)) && !landed {
				relative, _ := filepath.Rel(cache, name)
				return classed{"above cache", filepath.ToSlash(relative)}
			}
		}
		// Deeper below a directory of the cache's chain that's under the home, the same: tsc's @types search, climbing the
		// parents of a product in a cache under the home, missed <the cache's parent>/node_modules/@types, which named the
		// tree's own directory, so two checkouts keyed the adapted tree apart (#4gp2ytz, Compiler's Golem harness on
		// 1412e316ed). Valued by whether it's there alone; a listing or a content read there is still the home's, refused.
		// Only a miss: a path that's there keeps its own class, system, home or any other, since a test's temporary
		// directories are its cache's siblings and on a runner they sit under the home, where this rule took an lstat
		// of the machine's usr/cc ahead of its system class (chg_gp66hxpv's red, Tests' read).
		for _, cache := range facts.caches {
			if kinds != readAbsent {
				break
			}
			for at := filepath.Dir(filepath.Dir(name)); at != "/" && at != "."; at = filepath.Dir(at) {
				if !under(cache, at) {
					continue
				}
				if at != facts.home && under(at, facts.home) {
					relative, _ := filepath.Rel(cache, name)
					return classed{"above cache", filepath.ToSlash(relative)}
				}
				break
			}
		}
	}
	// A directory on the build cache's own chain of parents, listed or opened to list (Go's os.Open of a directory has no
	// O_DIRECTORY, so it reads as content), is keyed as a lookup beside the cache's parents is: by its place relative to
	// the cache and whether it's there, never by the names it lists, so two boxes whose homes hold different things
	// settle to one key (#4gp2ytz: esbuild's resolver lists every parent of where it runs up to /, looking for package.json,
	// tsconfig.json and node_modules, and every product it bundled refused under a cache in the home). Only that chain:
	// a listing of any other directory under the home is still refused, and what the resolver acts on, a file it opens or
	// a directory it lists off the chain, is read as itself and refused there (Tests' bounds, Builder's ruling).
	if name != "/" && kinds&^(readStat|readAbsent|readListing|readContent) == 0 && kinds&(readListing|readContent) != 0 && isDirectory(name) {
		for _, cache := range facts.caches {
			if under(cache, name) && name != cache {
				relative, _ := filepath.Rel(cache, name)
				return classed{"above cache", filepath.ToSlash(relative)}
			}
		}
	}
	for _, place := range facts.kernel {
		if under(name, place) {
			return classed{}
		}
	}
	// The Windows GPU driver's libraries, WSL's mount (#6eexfyv): a build that opens one depends on its box's Windows
	// driver, which no key names, so it refuses; the loader's cache is valued without them for every other build.
	if under(name, wslLibraries) && kinds&(readContent|readExec) != 0 {
		return classed{refusal + "the Windows GPU driver", name}
	}
	// The machine's time zone (#cbw5jtf): /etc/localtime links to whatever zone the box was set to, which no key names
	// (Server's Los Angeles against the house's Denver, Oct 12: every product a traced Go oracle built missed on Server).
	// A unit's TZ=UTC reaches every tool (toolchain carries it), so a build that still reads it refuses; only a look at it
	// is keyed, as a system path. The zone database itself (/usr/share/zoneinfo) is tzdata's, valued by its package.
	if slices.Contains(machineZone, name) && kinds&(readContent|readExec) != 0 {
		return classed{refusal + "the machine's time zone", name}
	}
	for _, place := range facts.system {
		if under(name, place) {
			// A machine directory a build only stat'd or opened to search in is keyed by whether and as what it exists,
			// never by its names: those are whatever the machine installed (Cloud's /usr/lib/x86_64-linux-gnu holds QEMU's
			// 235 more), so a native product missed on every box but the one that built it (#vdha67a, measured Oct 12).
			// What the build found or missed in it is keyed path by path; a directory it listed stays keyed by its names.
			if kinds&readListing == 0 && isDirectory(name) {
				return classed{"system directory", name}
			}
			return classed{"system", name}
		}
	}
	// A path that resolves through a symbolic link is what it resolves to (~/.adamic-tools, the link PATH runs clang
	// through, is ~/adamic-tools, the tools directory): its target is classed in its place, as strictly as if it had
	// been named. A machine file's link is keyed as itself (the system places, above), never resolved.
	if real := resolved(name); real != "" && real != name {
		return facts.classifyPath(real, kinds)
	}
	for _, place := range facts.temporary {
		if under(name, place) {
			// What the build's own processes made there never reaches here (record leaves it out). A directory opened
			// without O_DIRECTORY reads as content, but a directory has none (what it lists is getdents', its own read):
			// go's os.RemoveAll opens the temporary directory so to remove what the build made in it.
			if kinds&(readContent|readExec) != 0 && !isDirectory(name) {
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
	// A path anywhere else that was only looked for (a module resolver's virtual root, /adamic-prelude; clang's search for
	// a GCC installation, /opt/rh) is keyed by whether and as what it exists, the same on every machine that answers
	// alike; reading what one holds is still refused.
	if looked {
		return classed{"absolute", name}
	}
	return classed{refusal + "other", name}
}

// goCacheJoined says whether a path in the tree holds Go's build cache's own path below one of its directories: go list
// -test looks for the test main it cached at the package's directory joined to the cache's absolute path, a lookup
// that always misses (strace of go1.27.1 on Cloud, Oct 13: openat of <package>/home/ahra/.../gc/bb/...-d, ENOENT). It is
// go's bookkeeping in its own cache, never a tree path: classed as one, every product keyed where its builder's Go cache
// sat (#pqq9z5q: exists cmd/adamic/<GOCACHE>/28/...-d and bridge/tsgo's, 313 entries of Server's run).
func (facts *traceFacts) goCacheJoined(relative string) bool {
	for _, cache := range facts.goCaches {
		if strings.Contains("/"+relative+"/", "/"+cache+"/") {
			return true
		}
	}
	return false
}

// machine says whether name is in one of the system places, the machine's own files, keyed as themselves.
func (facts *traceFacts) machine(name string) bool {
	for _, place := range facts.system {
		if under(name, place) {
			return true
		}
	}
	return false
}

// landsWhereMachinesAgree says whether a directory a lookup looked in is the file system's root, the home directory or
// one of home's parents: what is there is the machine's, not the tree's place on it.
func (facts *traceFacts) landsWhereMachinesAgree(directory string) bool {
	return directory == "/" || directory == facts.home || under(facts.home, directory)
}

// caseProbe says whether a lookup that missed was a program asking whether the file system is case-sensitive:
// typescript-go's osvfs stats its own executable with every letter's case swapped as it starts (os.go's
// fileSystemCaseSensitivity), in every process that links it, and the answer is the machine's, not a path's. TypeScript's
// JavaScript sys asks the same of its own __filename, typescript.js, a file node read rather than ran (#4gp2ytz:
// stage3/api's, keyed as an absolute path naming the checkout, so the adapted tree keyed apart at two checkout paths).
// Only a swap of a program the run executed or a file it read is one; any other path in another case is the path it
// names.
func (facts *traceFacts) caseProbe(name string) bool {
	swapped := swapCase(name)
	return swapped != name && (facts.ran[swapped] || facts.executables[swapped] || facts.opened[swapped])
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
	// lastReads is each ended build's paths by when they were last read in its window, and endedAt when it ended.
	lastReads map[string]map[string]int
	endedAt   map[string]int
	// windowOpened is when each build's window opened, the sequence of the last read before it, and windowOwner the
	// process it is open for; lastOpened is the newest opening. A lean build's window is its reads (#mmpt615), and what
	// its processes wrote before it opened is never its own output (madeBy), so a writer is kept once a process for each
	// opening (record).
	windowOpened, windowOwner map[string]int
	lastOpened                int
	// windowLate are the windows the trace opened late, their build-begin from a thread it placed after the marker,
	// with no process the build's id names to open them for at the marker (openEarly): their owner's other threads may
	// have read meanwhile, outside them, so a lean build's reads are its superset. misowned are those opened at the
	// marker for the process the id names, whose marker's thread the trace then placed in another (Mentat's M2 and S1).
	windowLate, misowned map[string]bool
	order                []string
	lost                 []string
	started              bool
	// command is the command strace ran, the trace's first process; commandStatus its exit status, -1 until it exits.
	command, commandStatus int
	// runReads is everything the run read, the command's reads and every process's it started, taken as the command
	// exits (snapshot), and runLast when each path was last read (#sfqdxvy). Nil until it exits.
	runReads, runLast map[string]int
	// outside is each process's reads made outside every open build window, and windows each build's reads inside its
	// window, by its owner and the processes it started. runOutside and runOutsideLast are the run's reads outside every
	// window, taken as runReads is: the test read set's. A build's reads are its product's, never the test's, so a test
	// that built its products and one that found them record one set (#mmpt615, Node's ruling: the 111 test sets
	// #s57yq6y's A/B moved differed only by their products' reads, a build's content reads in one run, the test's own
	// stat of the same files in the other).
	outside                    map[int]map[string]*pathReads
	windows                    map[string]map[string]*pathReads
	runOutside, runOutsideLast map[string]int
}

// A pathReads is a process's reads of one path: their kinds, and when it first and last read the path (0 for a path it
// only wrote).
type pathReads struct {
	kinds, firstRead, lastRead int
}

// A writer is a process that wrote a path, and when it first did; it stays a writer after it ends.
type writer struct {
	process, sequence int
}

type traceCall struct {
	pid                    int
	call, arguments, value string
	opened, errno          string
	// queued is a call of a thread the trace hadn't placed when it printed it, handled once its thread is placed.
	queued bool
}

func newTraceState(facts *traceFacts) *traceState {
	state := &traceState{facts: facts, commandStatus: -1, process: map[int]int{}, parent: map[int]int{}, directory: map[int]string{}, keying: map[int]int{},
		keyingTree: map[int]bool{}, reads: map[int]map[string]*pathReads{}, writers: map[string][]writer{}, treeWrites: map[string]bool{}, waiting: map[int][]traceCall{},
		unfinished: map[int]string{}, open: map[string]int{}, ended: map[string]map[string]int{}, lastReads: map[string]map[string]int{}, endedAt: map[string]int{},
		outside: map[int]map[string]*pathReads{}, windows: map[string]map[string]*pathReads{}, windowOpened: map[string]int{}, windowOwner: map[string]int{},
		windowLate: map[string]bool{}, misowned: map[string]bool{}}
	return state
}

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
	number, rest, ok := splitTraceLine(text)
	if !ok {
		return
	}
	pid, _ := strconv.Atoi(number)
	if isExit(rest) {
		if code, exited := strings.CutPrefix(rest, "+++ exited with "); exited && state.started && pid == state.command {
			state.commandStatus, _ = strconv.Atoi(strings.TrimSuffix(code, " +++"))
		}
		state.exited(pid)
		return
	}
	if strings.HasSuffix(rest, "<unfinished ...>") {
		state.unfinished[pid] = strings.TrimSpace(strings.TrimSuffix(rest, "<unfinished ...>"))
		return
	}
	if tail, resumed := resumedTail(rest); resumed {
		start, ok := state.unfinished[pid]
		if !ok {
			return
		}
		delete(state.unfinished, pid)
		if strings.HasPrefix(tail, ")") || strings.HasPrefix(tail, ",") {
			rest = start + tail
		} else {
			rest = start + " " + tail
		}
	}
	call, ok := parseCall(rest)
	if !ok {
		return
	}
	state.handle(traceCall{pid: pid, call: call.call, arguments: call.arguments, value: call.value, opened: call.opened, errno: call.errno})
}

func (state *traceState) handle(call traceCall) {
	if !state.started {
		// The first process in the trace is the command strace ran, in the directory it ran from.
		state.started = true
		state.command = call.pid
		state.process[call.pid] = call.pid
		state.reads[call.pid] = map[string]*pathReads{}
		state.directory[call.pid], _ = os.Getwd()
	}
	process, known := state.process[call.pid]
	if !known {
		state.openEarly(call)
		call.queued = true
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
		if decoded, ok := descriptorPath(call.arguments); ok && call.value == "0" {
			state.directory[process] = unescape(decoded)
		}
		return
	case call.call == "getdents64" || call.call == "getdents":
		decoded, ok := descriptorPath(call.arguments)
		if !ok {
			return
		}
		name, kind = filepath.Clean(unescape(decoded)), readListing
	default:
		var rest string
		name, rest = target(call.call, call.arguments, state.directory[process])
		if name == "" {
			return
		}
		if strings.HasPrefix(name, traceMarker) {
			state.marker(call.pid, process, strings.TrimPrefix(name, traceMarker), call.queued)
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
		// Not for a machine file, which is keyed as itself and valued through its link (classifyPath): WSL makes
		// /etc/resolv.conf a link to /mnt/wsl/resolv.conf, a place no key can name, and every traced store lookup's DNS
		// read refused its product (Workshop, Oct 12: 866 of 866 in the smoke's tree).
		if opened := filepath.Clean(unescape(call.opened)); call.opened != "" && filepath.IsAbs(opened) && opened != name && !strings.HasSuffix(opened, " (deleted)") && !state.facts.machine(name) {
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
	if kind&readContent != 0 {
		state.facts.opened[name] = true
	}
	if state.keying[pid] > 0 || state.keyingTree[process] {
		return
	}
	state.sequence++
	if kind&readWrite != 0 {
		// A process's first write since the newest window opened, so a window sees whether its processes wrote a path after
		// it opened (madeBy).
		if !slices.ContainsFunc(state.writers[name], func(other writer) bool { return other.process == process && other.sequence > state.lastOpened }) {
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
	noteRead(state.reads[process], name, kind, state.sequence)
	// Its place: inside every open build window whose owner is process or started it, else outside them all.
	inside := false
	for id, owner := range state.open {
		if state.descends(process, owner) {
			inside = true
			noteRead(state.windows[id], name, kind, state.sequence)
		}
	}
	if !inside {
		if state.outside[process] == nil {
			state.outside[process] = map[string]*pathReads{}
		}
		noteRead(state.outside[process], name, kind, state.sequence)
	}
}

// noteRead adds a read of name, of kind, made at sequence, to reads.
func noteRead(reads map[string]*pathReads, name string, kind, sequence int) {
	found := reads[name]
	if found == nil {
		found = &pathReads{}
		reads[name] = found
	}
	found.kinds |= kind
	if kind&^readWrite != 0 {
		if found.firstRead == 0 {
			found.firstRead = sequence
		}
		found.lastRead = sequence
	}
}

func (state *traceState) marker(pid, process int, marker string, queued bool) {
	event, id, _ := strings.Cut(marker, "/")
	switch event {
	case "key-begin":
		state.keying[pid]++
	case "key-end":
		if state.keying[pid] > 0 {
			state.keying[pid]--
		}
	case "build-begin":
		if _, opened := state.windows[id]; opened {
			// Opened at the marker, for the process the id names (openEarly): the thread it came from is that process's.
			if owner := state.windowOwner[id]; owner != process {
				state.windowLate[id] = true
				if _, open := state.open[id]; open {
					state.open[id], state.windowOwner[id] = process, process
				} else {
					state.misowned[id] = true
				}
			}
			return
		}
		state.openWindow(id, process)
		if queued {
			state.windowLate[id] = true
		}
	case "build-end":
		owner, ok := state.open[id]
		if !ok {
			return
		}
		delete(state.open, id)
		snapshot, last := state.snapshot(owner)
		state.ended[id], state.lastReads[id], state.endedAt[id] = snapshot, last, state.sequence
		state.order = append(state.order, id)
	}
}

// openWindow opens build id's window for owner, now.
func (state *traceState) openWindow(id string, owner int) {
	state.open[id] = owner
	state.windows[id] = map[string]*pathReads{}
	state.windowOpened[id], state.windowOwner[id], state.lastOpened = state.sequence, owner, state.sequence
}

// openEarly opens a build's window at its build-begin marker when the trace hasn't placed the marker's thread yet
// (strace -f often prints a new thread's first call before its parent's clone returns), for the process the build's id
// names, when the trace knows that process: the reads its placed threads make before the thread is placed are the
// window's, never outside it (Mentat's M2). A process it doesn't know yet has no placed thread to read meanwhile, and its
// marker opens the window when its thread is placed, late (windowLate).
func (state *traceState) openEarly(call traceCall) {
	name, _ := target(call.call, call.arguments, "")
	marker, ok := strings.CutPrefix(name, traceMarker)
	if !ok {
		return
	}
	event, id, _ := strings.Cut(marker, "/")
	if _, opened := state.windows[id]; event != "build-begin" || opened {
		return
	}
	if owner, ok := buildProcess(id); ok && state.process[owner] == owner {
		state.openWindow(id, owner)
	}
}

// buildProcess is the process a build's id names (traced): its middle field, the pid of the process that built it.
func buildProcess(id string) (int, bool) {
	fields := strings.Split(id, ".")
	if len(fields) != 3 {
		return 0, false
	}
	pid, err := strconv.Atoi(fields[1])
	return pid, err == nil && pid > 0
}

// snapshot is everything owner and the processes it started read so far, by kinds, with when each path was last read,
// but what they wrote before reading it: that is the owner's own output, never an input. Only the owner's own processes
// make output for it; what a sibling or a process above it wrote is an input like any other (a sibling's product, a
// generated file). A product is always a product, whoever wrote it.
func (state *traceState) snapshot(owner int) (map[string]int, map[string]int) {
	return state.snapshotOf(state.reads, owner)
}

// snapshotOf is snapshot over processes' reads of one place, every read or those outside every build window.
func (state *traceState) snapshotOf(processes map[int]map[string]*pathReads, owner int) (map[string]int, map[string]int) {
	snapshot, last := map[string]int{}, map[string]int{}
	for other, reads := range processes {
		if !state.descends(other, owner) {
			continue
		}
		for name, found := range reads {
			kinds := found.kinds
			if kinds&^readWrite != 0 && !underAny(name, state.facts.caches) && state.madeBy(owner, name, 0, found.firstRead) {
				kinds &= readWrite
			}
			if kinds != 0 {
				snapshot[name] |= kinds
				last[name] = max(last[name], found.lastRead)
			}
		}
	}
	return snapshot, last
}

// windowSnapshot is snapshot over build id's window alone: what its owner and the processes it started read while the
// build was open, by kinds, with when each path was last read, but what they wrote after the window opened and before
// reading it. A write before the window is never the build's own output: the reads behind those bytes are outside the
// window, so the file reads by content (Mentat's M1: a test that writes a fixed path from its own case and builds from
// it).
func (state *traceState) windowSnapshot(id string) (map[string]int, map[string]int) {
	owner, opened := state.windowOwner[id], state.windowOpened[id]
	snapshot, last := map[string]int{}, map[string]int{}
	for name, found := range state.windows[id] {
		kinds := found.kinds
		if kinds&^readWrite != 0 && !underAny(name, state.facts.caches) && state.madeBy(owner, name, opened, found.firstRead) {
			kinds &= readWrite
		}
		if kinds != 0 {
			snapshot[name] |= kinds
			last[name] = max(last[name], found.lastRead)
		}
	}
	return snapshot, last
}

// leanReads is a lean build's reads (#mmpt615, Node's ruling): its window, and every file its declared Files name, as the
// tree carries them, each directory by its listing and every file under it by its content, read or not. A lean build
// may use its declared files' bytes taken before its window (native's runtime archives snapshot the runtime's files as
// the request is made, and compile the copies), so they are its reads whatever the window holds. A declared path that
// isn't there is read as absent. One it can't read whole is the reason it can't be keyed, never dropped: a file or a
// directory the tree doesn't carry, or one that can't be read (Mentat's S2, Node's must), since the build may have
// taken its bytes before its window, and the key would lack them.
func (state *traceState) leanReads(entry journalEntry, window, last map[string]int, end int) (map[string]int, map[string]int, string) {
	reads, lastReads := maps.Clone(window), maps.Clone(last)
	add := func(name string, kind int) {
		reads[name] |= kind
		lastReads[name] = max(lastReads[name], end)
	}
	root := cmp.Or(entry.Root, state.facts.root)
	for _, declared := range entry.Declared {
		start := declared
		if !filepath.IsAbs(start) {
			start = filepath.Join(root, filepath.FromSlash(declared))
		}
		err := filepath.WalkDir(start, func(name string, found fs.DirEntry, err error) error {
			switch {
			case err != nil && name == start && (errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)):
				add(name, readAbsent)
				return nil
			case err != nil:
				return err
			case found.IsDir() && found.Name() == ".git":
				return filepath.SkipDir
			case found.IsDir():
				if name != start && !Tracked(root, name, true) {
					return fmt.Errorf("%s is a directory the tree doesn't carry", name)
				}
				add(name, readListing)
			case Tracked(root, name, false):
				add(name, readContent)
			case !unkeyed(name):
				return fmt.Errorf("%s is a file the tree doesn't carry", name)
			}
			return nil
		})
		if err != nil {
			return nil, nil, fmt.Sprintf("declared %s can't be read whole: %v", declared, err)
		}
	}
	return reads, lastReads, ""
}

// madeBy says whether owner or a process it started wrote name, or a directory above it, after since and before first
// reading it.
func (state *traceState) madeBy(owner int, name string, since, firstRead int) bool {
	for at := name; at != "/" && at != "."; at = filepath.Dir(at) {
		for _, wrote := range state.writers[at] {
			if since < wrote.sequence && wrote.sequence < firstRead && state.descends(wrote.process, owner) {
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
	if ok && process == pid && pid == state.command && state.started {
		// The whole run's reads, and those outside every build window, the test read set's, before the command's own
		// fold away with it.
		state.runReads, state.runLast = state.snapshot(pid)
		state.runOutside, state.runOutsideLast = state.snapshotOf(state.outside, pid)
	}
	delete(state.process, pid)
	delete(state.keying, pid)
	if !ok || process != pid {
		return
	}
	reads, outside := state.reads[pid], state.outside[pid]
	delete(state.reads, pid)
	delete(state.outside, pid)
	delete(state.directory, pid)
	// A process that outlives the command has no live ancestor left to fold into: its reads join the run's, since the
	// test read set is everything the run read, whenever it read it (Mentat's review of 8.8).
	defer func() {
		if state.runReads == nil || pid == state.command {
			return
		}
		for ancestor, ok := state.parent[pid]; ok; ancestor, ok = state.parent[ancestor] {
			if _, live := state.reads[ancestor]; live {
				return
			}
		}
		state.joinRun(state.runReads, state.runLast, reads)
		state.joinRun(state.runOutside, state.runOutsideLast, outside)
	}()
	for ancestor, ok := state.parent[pid]; ok; ancestor, ok = state.parent[ancestor] {
		if into, live := state.reads[ancestor]; live {
			merge(into, reads)
			if state.outside[ancestor] == nil {
				state.outside[ancestor] = map[string]*pathReads{}
			}
			merge(state.outside[ancestor], outside)
			return
		}
	}
}

// merge folds reads into into, as an ended process's reads join its nearest live ancestor's.
func merge(into, reads map[string]*pathReads) {
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
		kept.lastRead = max(kept.lastRead, found.lastRead)
	}
}

// joinRun adds reads to a run's, run and last, but what the command or a process it started wrote before reading it:
// the run's own output, never an input.
func (state *traceState) joinRun(run, last map[string]int, reads map[string]*pathReads) {
	for name, found := range reads {
		kinds := found.kinds
		if kinds&^readWrite != 0 && !underAny(name, state.facts.caches) && state.madeBy(state.command, name, 0, found.firstRead) {
			kinds &= readWrite
		}
		if kinds != 0 {
			run[name] |= kinds
			last[name] = max(last[name], found.lastRead)
		}
	}
}

// target is the absolute path a call names and the rest of its arguments, or "".
func target(call, arguments, directory string) (string, string) {
	base := directory
	if atCalls[call] {
		decoded, length, ok := directoryArgument(arguments)
		if !ok {
			return "", arguments
		}
		if decoded != "" {
			base = unescape(decoded)
		}
		arguments = arguments[length:]
	} else if !plainCalls[call] && call != "symlink" && call != "symlinkat" {
		return "", arguments
	}
	quotedName, length, ok := quotedArgument(arguments)
	if !ok {
		return "", arguments
	}
	name := unescape(quotedName)
	rest := arguments[length:]
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
		quotedName, _, ok := quotedArgument(rest)
		if !ok {
			return ""
		}
		name := unescape(quotedName)
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

// unescape is a path as strace spells it, quoted or decoded from a descriptor (--decode-fds=path), back in its bytes:
// strace writes every byte outside printable ASCII, and the quotes and angle brackets around it, as an octal escape
// (walking/\303\251.txt is walking/é.txt), which git ls-files -z never does.
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
	if len(state.lost) > 0 {
		settlement.Reports = append(settlement.Reports, Report{Outcome: ReportLost, Reason: "the whole run can't be keyed: " + strings.Join(state.lost, "; ")})
	}
	refuse := func(entry journalEntry, reason string) {
		settlement.Refused = append(settlement.Refused, fmt.Sprintf("refused %s (name key %s): %s", entry.Name, entry.NameKey[:min(12, len(entry.NameKey))], reason))
		settlement.Reports = append(settlement.Reports, Report{Name: entry.Name, NameKey: entry.NameKey, Outcome: ReportRefused, Reason: reason})
		// Never placed: a pending product that can't be keyed leaves the cache.
		if under(entry.Directory, filepath.Join(facts.cache, "pending")) {
			os.RemoveAll(entry.Directory)
			os.Remove(entry.Directory + ".inputs")
			os.Remove(entry.Directory + ".lock")
		}
	}
	// When each product this trace built ended its build: a read of it before then was its own build writing it.
	productEnds := map[string]int{}
	for id, entry := range built {
		if end, ended := state.endedAt[id]; ended {
			productEnds[entry.Directory] = end
		}
	}
	recipes := map[string]recipe{}
	settled := map[string]bool{}
	// keyed is each build that settled, by its id, with its product's name key: a test read set stands it in for its window.
	keyed := map[string]string{}
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
		if state.misowned[id] {
			refuse(entry, fmt.Sprintf("a lost trace: its build's window opened for process %d, which its id names, and the trace placed its marker in another", state.windowOwner[id]))
			continue
		}
		reads, lastReads, window := state.ended[id], state.lastReads[id], BuildSuperset
		if entry.Lean {
			window = BuildLeanSuperset
			if !state.windowLate[id] {
				inside, last := state.windowSnapshot(id)
				var why string
				if reads, lastReads, why = state.leanReads(entry, inside, last, state.endedAt[id]); why != "" {
					refuse(entry, why)
					continue
				}
				window = BuildWindowed
			}
		}
		set, reason := state.measured(entry, reads, lastReads, products, productEnds)
		if reason == "" {
			reason = refusedHandedRead(entry, set)
		}
		// The code a build ran: a windowed lean build's own when it's buildcache's (leanRecipeOf, #earlycut), every other
		// build's process's binary whole (decision 1). A lean caller's build keeps its window and the whole recipe until its
		// census holds: native's archives reach the jobserver, a package variable written outside its initializer, until it's
		// handed in as a value (#yzwm4rj, Node's reading 1).
		own := processes[entry.Process]
		name := recipeName(own)
		leanRecipe := window == BuildWindowed && entry.LeanCaller == ""
		if leanRecipe {
			name = "lean " + name
		}
		if reason == "" {
			if _, ok := recipes[name]; !ok {
				if leanRecipe {
					recipes[name] = leanRecipeFor(facts.root, entry.LeanCaller, own)
				} else {
					recipes[name] = recipeFor(facts.root, own)
				}
			}
			set.Entries, reason = withRecipe(set.Entries, recipes[name])
		}
		if reason != "" {
			refuse(entry, reason)
			continue
		}
		set.Trace, set.Lean = filepath.Base(facts.trace), leanRecipe
		settingsOverride = recipes[name].settings
		set, evaluated, product, err := recordUnion(facts.root, facts.cache, entry.NameKey, entry.Name, set, func(key string) string {
			return placeSettled(facts.cache, key, entry)
		})
		settingsOverride = ""
		if err != nil {
			refuse(entry, err.Error())
			continue
		}
		products[product] = entry.NameKey
		keyed[id] = entry.NameKey
		line := fmt.Sprintf("settled %s %s: %d reads, %s", entry.Name, evaluated.key[:12], len(set.Entries), window)
		if set.Drift != nil {
			line += fmt.Sprintf("; declared Files: %d reads undeclared, %d declared unread", set.Drift.Undeclared, set.Drift.Unread)
		}
		settlement.Settled = append(settlement.Settled, line)
		settlement.Reports = append(settlement.Reports, Report{Name: entry.Name, NameKey: entry.NameKey, Key: evaluated.key, Outcome: ReportSettled, Reads: len(set.Entries), Window: window, Lean: set.Lean})
	}
	for id, entry := range built {
		if !settled[id] {
			refuse(entry, "a lost trace: its build's markers aren't in the trace")
		}
	}
	if nameKey := os.Getenv("ADAMIC_TRACE_TEST_KEY"); nameKey != "" {
		state.recordTestReads(&settlement, nameKey, processes, recipes, products, productEnds, keyed)
	}
	// What the run found already built, under its key in the cache: a test's products are these and the ones it
	// settled (Planner's step 5 proof on Home, Oct 11: a tree that took a test's products from the census lost every
	// one found). One this trace built itself is pending until it settles, and its settlement is its line.
	names, reported := map[string]string{}, map[string]bool{}
	for _, entry := range entries {
		if entry.Name != "" && entry.NameKey != "" {
			names[entry.NameKey] = entry.Name
		}
	}
	for _, entry := range entries {
		key := filepath.Base(entry.Directory)
		if entry.Event != "found" || !isKey(key) || !slices.Contains(facts.caches, filepath.Dir(entry.Directory)) || reported[entry.NameKey+key] {
			continue
		}
		reported[entry.NameKey+key] = true
		settlement.Reports = append(settlement.Reports, Report{Name: names[entry.NameKey], NameKey: entry.NameKey, Key: key, Outcome: ReportFound, Lean: leanSet(facts.cache, entry.NameKey, key)})
	}
	os.Remove(filepath.Join(facts.cache, "pending", filepath.Base(facts.trace)))
	sort.Strings(settlement.Refused)
	return settlement
}

// measured is a build's reads as a read set, or the reason it can't be one.
func (state *traceState) measured(entry journalEntry, reads, lastReads map[string]int, products map[string]string, productEnds map[string]int) (readSet, string) {
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
				// A product this trace built is an input only if its build had ended before this build read it: read
				// while it was still being built, it was its own build writing it, which a superset window (a sibling in
				// the same process) also holds, and two siblings holding each other would be a cycle no key can value
				// (Home's dry run, Oct 11: bridge-test-api and tsgo-asan.a, each refused for reading the other).
				if end, builtHere := productEnds[found.detail]; builtHere && lastReads[name] <= end {
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
// two settlements can race), so the set recorded and keyed is their union, and the earlier set gives way to it. The
// union is lean only when every set in it is: a lean set folded with a whole one keys by the whole recipe's reads too.
func recordUnion(root, cache, nameKey, name string, set readSet, placing func(key string) string) (readSet, evaluation, string, error) {
	var evaluated evaluation
	var product string
	err := withReads(cache, nameKey, func(file *readsFile) error {
		var kept []readSet
		for _, other := range file.Sets {
			if before, err := evaluate(root, cache, nameKey, other, 0, nil); err == nil && len(differing(other, before)) == 0 {
				set.Entries, set.Lean = sortedEntries(set.Entries, other.Entries), set.Lean && other.Lean
				continue
			}
			kept = append(kept, other)
		}
		var err error
		if evaluated, err = evaluate(root, cache, nameKey, set, 0, nil); err != nil {
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

// leanSet says whether the set nameKey's read sets in cache store under key is lean: what a found product's report
// says, so Loom's sampled key check draws from lean products a run found as well as the ones it built.
func leanSet(cache, nameKey, key string) bool {
	file, err := loadReads(cache, nameKey)
	if err != nil {
		return false
	}
	for _, set := range file.Sets {
		if set.Key == key {
			return set.Lean
		}
	}
	return false
}

// placeSettled puts a settled build's product under its key and returns where it is: a product already there (the same
// bytes, from a build of the same reads) stays and the build's copy goes; an uncached build's directory on another
// file system stays where it was built. It looks and places under the key's lock, as filled does, so a lookup fetching
// the same key never loses its rename to this one, nor this one to it (#bcae1e2). cache is where the run writes: a
// product already under the key in the build cache it reads through an output directory of its own is placed too, and
// stays where it is; the lock, a rename and the rewritten <key>.inputs are the output directory's alone.
func placeSettled(cache, key string, entry journalEntry) string {
	product := filepath.Join(cache, key)
	pending := under(entry.Directory, filepath.Join(cache, "pending"))
	unlock, err := lockTarget(product)
	if err != nil {
		return entry.Directory
	}
	defer unlock()
	placed, already := existingPlace(product)
	switch {
	case already:
		if pending {
			os.RemoveAll(entry.Directory)
		}
	case os.Rename(entry.Directory, product) != nil:
		return entry.Directory
	default:
		placed = product
	}
	if pending {
		os.Remove(entry.Directory + ".inputs")
		os.Remove(entry.Directory + ".lock")
	}
	writeWhole(product+".inputs", []byte("name key "+entry.NameKey+"\n"+entry.Inputs))
	return placed
}

// writeWhole writes content to path through a temporary file beside it, whose name starts with a dot as every scratch
// name in the cache does, renamed into place, so a reader or a copy of the directory never meets half of it. A failure
// leaves path as it was and is its error.
func writeWhole(path string, content []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-")
	if err != nil {
		return err
	}
	if _, err = temporary.Write(content); err == nil {
		err = temporary.Close()
	} else {
		temporary.Close()
	}
	if err == nil {
		err = os.Chmod(temporary.Name(), 0o644)
	}
	if err == nil {
		err = os.Rename(temporary.Name(), path)
	}
	if err != nil {
		os.Remove(temporary.Name())
	}
	return err
}

// TestReadSetName is a test read set's name in its read sets file, and so in a Finding (#sfqdxvy).
const TestReadSetName = "test read set"

// productKeyKind is a test read set's entry for a product the run used, by its name key, valued by the key the product
// is found under: a product's own sets name what they read by the bytes of it (kind product), a test set by its key.
const productKeyKind = "product key"

// recordTestReads records the whole run's reads as the test read set under nameKey, the name key Loom's tree builder
// names in ADAMIC_TRACE_TEST_KEY (#sfqdxvy), and gives settlement its line, and its report line when it's recorded:
// outcome readSet, the name key and the marker's key, which Loom keeps to vouch with this set and no other. A product test that only finds products another
// test built has no build, so no product's set holds the reads that decided the names it asked for; this set does: what
// the command and every process it started read, keying aside and not what the run wrote before reading it (snapshot),
// classed and keyed as a product's set is, with the recipe of the processes that journaled. A build's own reads aren't
// the test's: each build that settled (keyed) stands in for its window by its product's key (#mmpt615). Recorded by recordUnion
// beside a marker product under its key, so cmd/find finds it exactly when every value the run read is what it was. Only
// a run that can stand for the test's next run is recorded: nothing lost, the command exited 0, and every process that
// journaled ran one recipe; any other records nothing, and its test runs.
func (state *traceState) recordTestReads(settlement *Settlement, nameKey string, processes map[int]journalEntry, recipes map[string]recipe, products map[string]string, productEnds map[string]int, keyed map[string]string) {
	settlement.TestReads = state.testReads(settlement, nameKey, processes, recipes, products, productEnds, keyed)
}

// testReads is recordTestReads' line, after recording the set and its report line when it can.
func (state *traceState) testReads(settlement *Settlement, nameKey string, processes map[int]journalEntry, recipes map[string]recipe, products map[string]string, productEnds map[string]int, keyed map[string]string) string {
	facts := state.facts
	none := func(why string) string { return "no test read set: " + why }
	switch {
	case !nameKeyPattern.MatchString(nameKey):
		return none(fmt.Sprintf("ADAMIC_TRACE_TEST_KEY %q isn't a name key (64 lowercase hex digits)", nameKey))
	case len(state.lost) > 0:
		return none("the whole run can't be keyed: " + strings.Join(state.lost, "; "))
	case state.commandStatus != 0:
		return none(fmt.Sprintf("the command didn't exit 0 (%d), and a failing run isn't the run a skip would stand in for", state.commandStatus))
	case state.runReads == nil:
		return none("the command's exit isn't in the trace")
	}
	name := ""
	for _, pid := range slices.Sorted(maps.Keys(processes)) {
		switch own := recipeName(processes[pid]); {
		case name == "":
			name = own
		case own != name:
			return none("its processes ran different binaries, " + name + " and " + own)
		}
	}
	if name == "" {
		return none("no process of the run keyed a product, so no recipe names the code it ran")
	}
	if _, ok := recipes[name]; !ok {
		for _, process := range processes {
			recipes[name] = recipeFor(facts.root, process)
			break
		}
	}
	// The run's reads outside every build window, and a window's reads only when its build didn't settle, since then no
	// key stands for them (#mmpt615, Node's ruling).
	reads, last := maps.Clone(state.runOutside), maps.Clone(state.runOutsideLast)
	for id, window := range state.windows {
		if _, settled := keyed[id]; !settled {
			state.joinRun(reads, last, window)
		}
	}
	set, reason := state.measured(journalEntry{Name: TestReadSetName, NameKey: nameKey}, reads, last, products, productEnds)
	if reason == "" {
		// Every product the run used is named by its key: one it found, as the run read it, and one it built, for the
		// window the set leaves out. Its key, never its bytes: a window holds every read its owner's process made while
		// it was open, a parallel subtest's too, and the product's set is a superset of them, so only its key moves with
		// each file any of them read (Node's guard on #mmpt615: keys complete before minimal).
		for index := range set.Entries {
			if set.Entries[index].Kind == "product" {
				set.Entries[index].Kind = productKeyKind
			}
		}
		for _, built := range keyed {
			set.Entries = append(set.Entries, readEntry{Kind: productKeyKind, Path: built})
		}
		set.Entries, reason = withRecipe(set.Entries, recipes[name])
	}
	if reason != "" {
		return none(reason)
	}
	set.Trace = filepath.Base(facts.trace)
	settingsOverride = recipes[name].settings
	set, evaluated, _, err := recordUnion(facts.root, facts.cache, nameKey, TestReadSetName, set, func(key string) string {
		return placeTestMarker(facts.cache, key, nameKey)
	})
	settingsOverride = ""
	if err != nil {
		return none(err.Error())
	}
	settlement.Reports = append(settlement.Reports, Report{Name: TestReadSetName, NameKey: nameKey, Key: evaluated.key, Outcome: ReportReadSet, Reads: len(set.Entries)})
	return fmt.Sprintf("test read set %s: %d reads", evaluated.key[:12], len(set.Entries))
}

// placeTestMarker is a test read set's marker product under key: a directory naming the set it stands for, so a lookup
// finds the set when its key names something in the cache, as a product's is found. Nothing reads it. Its file goes by a
// temporary file and a rename, as a read sets file does: a seeded cache links a seed's files, and a write in place would
// write the seed's (Mentat's re-review).
func placeTestMarker(cache, key, nameKey string) string {
	marker := filepath.Join(cache, key)
	if err := os.MkdirAll(marker, 0o755); err != nil {
		return marker
	}
	partial, err := os.CreateTemp(marker, ".test-read-set-")
	if err != nil {
		return marker
	}
	_, err = partial.WriteString(nameKey + "\n")
	if closeErr := partial.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(partial.Name(), filepath.Join(marker, "test-read-set"))
	}
	if err != nil {
		os.Remove(partial.Name())
	}
	return marker
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
	// An entry covers a read that is it, or that is under it (a read under "dir" or "dir/"). Each read climbs its own
	// directories through these, rather than meeting every entry: home-6's dry run (Oct 11) had tsprinter's mutant
	// products, 76,958 reads against 75,696 declared entries each, settling for over 20 minutes.
	exactly, under := map[string][]string{}, map[string][]string{}
	for _, entry := range declared {
		exactly[entry] = append(exactly[entry], entry)
		under[strings.TrimSuffix(entry, "/")] = append(under[strings.TrimSuffix(entry, "/")], entry)
	}
	result := &drift{}
	read := map[string]bool{}
	for _, name := range reads {
		var covering []string
		covering = append(covering, exactly[name]...)
		for at := name; ; {
			parent := path.Dir(at)
			if parent == at || parent == "." {
				break
			}
			if parent == "/" {
				// "dir/" with dir empty: what an empty or "/" entry covers, every absolute read.
				covering = append(covering, under[""]...)
				break
			}
			covering = append(covering, under[parent]...)
			at = parent
		}
		declaredHere := len(covering) > 0
		for _, entry := range covering {
			read[entry] = true
		}
		if !declaredHere && name != "." {
			result.Undeclared++
			result.UndeclaredPaths = append(result.UndeclaredPaths, name)
		}
	}
	for _, entry := range declared {
		if !read[entry] && !unkeyed(entry) {
			result.Unread++
			result.UnreadPaths = append(result.UnreadPaths, entry)
		}
	}
	result.UndeclaredPaths, result.UnreadPaths = driftSample(result.UndeclaredPaths), driftSample(result.UnreadPaths)
	return result
}

// driftSample is the first 20 of paths in order, sorted before it is cut: the reads come from a map, so a sample taken
// first and sorted after named another 20 each run, and the read set recorded it (Loom, Oct 11).
func driftSample(paths []string) []string {
	sort.Strings(paths)
	if len(paths) > 20 {
		paths = paths[:20:20]
	}
	return paths
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

// makeFilesOnly keys this package, in every recipe but its own tests', by its make files alone (make_*.go, isMakeFile)
// and its listing (#a328f0h, Loom's ruling 45922): every product's process links buildcache, so keying all of it moved
// every product's key at every change to how products are keyed, traced, settled or stored, which never changes a
// product's bytes (54% of the rekeys over main's 200 landings of Oct 11 to 13, Adam's seed 48a9e045). The make files are
// what decides a product's bytes, and TestTheMakeFilesUseNothingFromTheMachinery holds them to using nothing the
// machinery defines, so a change outside them can't move a product's bytes. The listing still names every file, so a
// file added or removed moves every key. Off is decision 1 for every package.
const makeFilesOnly = true

// buildcachePackage is this package's import path, as go list names it.
const buildcachePackage = "github.com/system-inc/adamic/internal/buildcache"

// isMakeFile says whether name, a file of this package, is one of the files that decide a product's bytes (make_*.go),
// the only ones a product's key reads under makeFilesOnly.
func isMakeFile(name string) bool {
	return strings.HasPrefix(name, "make_") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

// IsMachinery says whether file, slash-separated from the tree's root, is part of the build machinery: a file of this
// package that a product keyed by recipeOf leaves out of its key (every one but its make files, under makeFilesOnly),
// though a lean product, keyed by leanRecipeOf, keys them all; or anything under its cmd/ (traced, which decides what a
// read set records, and find, the skip), which no product's process links. Calling a file some products key machinery
// only refuses more seeds, the safe side. A product's key never says which machinery made it, so Loom seeds a tree from another only
// when the diff between them names none of it (#kjgmzt8, Node's ruling): the same machinery, from the same inputs the
// keys name.
func IsMachinery(file string) bool {
	directory := strings.TrimPrefix(buildcachePackage, "github.com/system-inc/adamic/")
	if strings.HasPrefix(file, directory+"/cmd/") {
		return true
	}
	parent, name := path.Split(file)
	return makeFilesOnly && parent == directory+"/" && !isMakeFile(name)
}

func recipeOf(root string, process journalEntry) recipe {
	pkg, test := strings.CutSuffix(process.Package, ".test")
	if pkg == "" || pkg == "command-line-arguments" {
		return recipe{reason: fmt.Sprintf("its process's binary (%s) names no package, so the code it ran can't be keyed", process.Executable)}
	}
	return recipeListed(root, process, []recipeListing{{pkg: pkg, test: test, deps: true}})
}

// leanRecipeFor is leanRecipeOf; a test in a tree of its own names a lean build's recipe itself, as recipeFor does.
var leanRecipeFor = leanRecipeOf

// leanRecipeOf is a lean build's recipe (#earlycut, Node's approved design): the code that makes it, never the code of
// the test binary that settled it, which keyed every native runtime archive and every go oracle built in a test linking
// the compiler on the whole compiler. It's buildcache and each package of the tree buildcache links, all their non-test
// files (buildcache whole, Tests' ruling of Oct 16: goProduct's plan runs a go build, and
// they're no make file), and a lean caller's own package (caller, leanCallers), never what the caller links: what its
// build calls there is in the Files it declares, which the window keys as read (leanReads), and the caller's census holds
// the build's reach to them (runtimeproducts' TestTheRuntimeBuildReachesOnlyItsDeclaredFiles). Under process's build
// settings, as the binary was built, and refused as recipeOf refuses when process names no package.
func leanRecipeOf(root, caller string, process journalEntry) recipe {
	if pkg := strings.TrimSuffix(process.Package, ".test"); pkg == "" || pkg == "command-line-arguments" {
		return recipe{reason: fmt.Sprintf("its process's binary (%s) names no package, so the code it ran can't be keyed", process.Executable)}
	}
	listings := []recipeListing{{pkg: buildcachePackage, deps: true, whole: true}}
	if caller != "" && caller != buildcachePackage {
		listings = append(listings, recipeListing{pkg: caller})
	}
	return recipeListed(root, process, listings)
}

// A recipeListing is one package a recipe keys, as go list names it: with every package of the tree it links (deps), and
// its test files with it, when it's a test binary's package (test). Buildcache is keyed by its make files alone
// (makeFilesOnly), unless it's the package under test or the listing keys it whole (whole, a lean build's).
type recipeListing struct {
	pkg               string
	test, deps, whole bool
}

// recipeListed is the recipe of listings, under process's build settings: every file of each listed package of the tree,
// each package's directory listing, each module's go.mod and go.sum, the modules the main module replaces with a
// directory of the tree, and the workspace file.
func recipeListed(root string, process journalEntry, listings []recipeListing) recipe {
	// The binary's own build settings over toolchain's environment, the one GoBuild's go builds under.
	var tags, environment []string
	for _, setting := range process.Settings {
		key, value, _ := strings.Cut(setting, "=")
		switch key {
		case "-tags":
			tags = append(tags, "-tags="+value)
		case "GOOS", "GOARCH", "CGO_ENABLED", "GOAMD64", "GOARM64", "GOEXPERIMENT":
			environment = append(environment, key+"="+value)
		}
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
	for _, listing := range listings {
		pkg, test := listing.pkg, listing.test
		arguments := []string{"list", "-json"}
		if listing.deps {
			arguments = append(arguments, "-deps")
		}
		if test {
			arguments = append(arguments, "-test")
		}
		command := toolchain.Command("go", append(append(arguments, tags...), pkg)...)
		command.Dir, command.Env = root, toolchain.Env(environment...)
		output, err := command.Output()
		if err != nil {
			return recipe{reason: fmt.Sprintf("go list of its binary's package %s: %v", pkg, err)}
		}
		if reason := recipeEntries(root, pkg, test, listing.whole, output, add); reason != "" {
			return recipe{reason: reason}
		}
	}
	// A module the main module replaces with a directory of the tree selects versions through its go.mod, whether or not
	// the binary links a package of it (the 17 cohere/TypeScript-shim modules, #c2chyab): only the main module's and the
	// workspace's replace directives apply, so those name every such go.mod.
	replaced, err := localReplacements(root, root)
	if err != nil {
		return recipe{reason: err.Error()}
	}
	for _, directory := range replaced {
		for _, name := range existing(directory, "go.mod", "go.sum") {
			add("content", filepath.Join(directory, name))
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

// recipeEntries adds to a recipe, through add, what output, go list's JSON of pkg (a test binary's when test), names of
// the tree: each package's directory, its files, and its module's go.mod and go.sum, buildcache's every non-test file
// when whole. It's the reason the listing can't be read, or "".
func recipeEntries(root, pkg string, test, whole bool, output []byte, add func(kind, name string)) string {
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var listed struct {
			ImportPath, Dir                                                            string
			Standard                                                                   bool
			Module                                                                     *struct{ GoMod string }
			GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
			IgnoredGoFiles, TestGoFiles, XTestGoFiles                                  []string
		}
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return "go list: " + err.Error()
		}
		if listed.Standard || !inside(root, listed.Dir) {
			continue
		}
		add("listing", listed.Dir)
		// Only the package under test's test files are compiled into its binary (its own and its external test
		// package's): a dependency's _test.go never is, and keying it moved every product whose binary linked that
		// package at every test edit there (#c2chyab: internal/buildcache/precise_test.go was in all 808 product
		// tests' read sets). A dependency's ignored files stay, since a build-constraint edit can compile one in, but
		// never its ignored test files, another platform's, which no edit compiles into a binary that isn't its own.
		name, _, _ := strings.Cut(listed.ImportPath, " ")
		under := test && (name == pkg || name == pkg+"_test")
		ignored := listed.IgnoredGoFiles
		if !under {
			ignored = slices.DeleteFunc(slices.Clone(ignored), func(file string) bool { return strings.HasSuffix(file, "_test.go") })
		}
		files := [][]string{listed.GoFiles, listed.CgoFiles, listed.CFiles, listed.CXXFiles, listed.HFiles, listed.SFiles, listed.SysoFiles,
			listed.EmbedFiles, ignored}
		if under {
			files = append(files, listed.TestGoFiles, listed.XTestGoFiles)
		}
		// This package, linked by every product's process, is keyed by its make files alone (makeFilesOnly), unless it is
		// the package under test or a lean build's recipe keys it whole.
		makeOnly := makeFilesOnly && name == buildcachePackage && !under && !whole
		for _, names := range files {
			for _, name := range names {
				if makeOnly && !isMakeFile(name) {
					continue
				}
				// go list -test names the test main it generated by its absolute place in go's own cache: go's, made from
				// the test files keyed here, and outside the tree, so add leaves it out. Joined to the package's directory
				// it was a tree path naming the cache's location, and every product keyed where its builder's Go cache sat
				// (#pqq9z5q: 1,271 of 1,271 in Server's comparison against Workshop).
				if !filepath.IsAbs(name) {
					name = filepath.Join(listed.Dir, name)
				}
				add("content", name)
			}
		}
		if listed.Module != nil && inside(root, listed.Module.GoMod) {
			for _, name := range existing(filepath.Dir(listed.Module.GoMod), "go.mod", "go.sum") {
				add("content", filepath.Join(filepath.Dir(listed.Module.GoMod), name))
			}
		}
	}
	return ""
}

// localReplacements are the directories each place's go.mod and go.work replace modules with, absolute, read from the
// files alone, never the module graph or the network. A place without one names none.
func localReplacements(root string, places ...string) ([]string, error) {
	var directories []string
	add := func(place string, replaced []string) {
		for _, path := range replaced {
			if directory := filepath.Join(place, path); !slices.Contains(directories, directory) {
				directories = append(directories, directory)
			}
		}
	}
	for _, place := range places {
		if len(existing(place, "go.mod")) > 0 {
			content, err := os.ReadFile(filepath.Join(place, "go.mod"))
			if err != nil {
				return nil, err
			}
			replaced, err := goModReplacements(content)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", filepath.Join(place, "go.mod"), err)
			}
			add(place, replaced)
		}
		if len(existing(place, "go.work")) > 0 {
			// Named, so neither GOWORK nor GOFLAGS can point the edit elsewhere; a runner's go answers it (Loom's
			// stand-in allows go work edit -json, as goWorkspace asks it).
			output, err := toolchain.CommandIn(root, "go", "work", "edit", "-json", filepath.Join(place, "go.work")).Output()
			if err != nil {
				return nil, fmt.Errorf("go work edit: %v", err)
			}
			var file struct {
				Replace []struct {
					New struct{ Path, Version string }
				}
			}
			if err := json.Unmarshal(output, &file); err != nil {
				return nil, fmt.Errorf("go work edit: %v", err)
			}
			var replaced []string
			for _, replacement := range file.Replace {
				if replacement.New.Version == "" {
					replaced = append(replaced, replacement.New.Path)
				}
			}
			add(place, replaced)
		}
	}
	return directories, nil
}

// goModReplacements is go.mod's local replacements as go reads them: each replace directive's new path, in file order,
// for every replacement that names no version (a directory, as go requires of one). It reads the file as
// golang.org/x/mod/modfile lexes it (Go's own cmd/go vendors it): tokens split at spaces, a // comment to the line's
// end outside a string, a string quoted with " (escapes) or ` (raw) on one line, ( ) [ ] { } and , each a token, and
// => a token like any other word; a directive is one line, or a verb's ( block of lines up to its ). Of two directives
// replacing one module (its path and version), the later stands where it is and the earlier goes, as go mod edit's
// clean-up keeps them (modfile's removeDups): lap 4's go.mod replaces shim/parser twice. It runs no go command: a Loom
// runner's stand-in go refuses go mod edit, and every unit keying a GoBuild product asked it (lap 4, 27 packages
// broken at fetch).
func goModReplacements(content []byte) ([]string, error) {
	type replacement struct{ old, path string }
	var replacements []replacement
	block := ""
	for number, line := range strings.Split(string(content), "\n") {
		tokens, err := goModTokens(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", number+1, err)
		}
		var arguments []goModToken
		switch {
		case len(tokens) == 0:
			continue
		case block != "" && tokens[0].is(")"):
			block = ""
			continue
		case block != "":
			if block != "replace" {
				continue
			}
			arguments = tokens
		case len(tokens) == 3 && tokens[1].is("(") && tokens[2].is(")"):
			// An empty block on one line: replace ().
			continue
		case len(tokens) == 2 && tokens[1].is("("):
			block = tokens[0].text
			continue
		case tokens[0].text == "replace" && !tokens[0].quoted:
			arguments = tokens[1:]
		default:
			continue
		}
		arrow := slices.IndexFunc(arguments, func(token goModToken) bool { return token.text == "=>" && !token.quoted })
		if arrow < 1 || arrow > 2 || len(arguments) < arrow+2 || len(arguments) > arrow+3 {
			return nil, fmt.Errorf("line %d: a replace go doesn't read: %s", number+1, strings.TrimSpace(line))
		}
		var old []string
		for _, token := range arguments[:arrow] {
			old = append(old, token.text)
		}
		path := ""
		if len(arguments) == arrow+2 {
			path = arguments[arrow+1].text
		}
		replacements = append(replacements, replacement{strings.Join(old, " "), path})
	}
	var local []string
	for index, each := range replacements {
		later := slices.ContainsFunc(replacements[index+1:], func(other replacement) bool { return other.old == each.old })
		if each.path != "" && !later {
			local = append(local, each.path)
		}
	}
	return local, nil
}

// A goModToken is one token of a go.mod line: its text, unquoted when it was a string, and whether it was one, since a
// quoted ( or => is a path, never the punctuation or the arrow.
type goModToken struct {
	text        string
	quoted      bool
	punctuation bool
}

// is says whether the token is the punctuation mark mark.
func (token goModToken) is(mark string) bool {
	return token.punctuation && token.text == mark
}

// goModTokens is one go.mod line's tokens.
func goModTokens(line string) ([]goModToken, error) {
	var tokens []goModToken
	for index := 0; index < len(line); {
		switch c := line[index]; {
		case c == ' ' || c == '\t' || c == '\r':
			index++
		case strings.HasPrefix(line[index:], "//"):
			return tokens, nil
		case strings.ContainsRune("()[]{},", rune(c)):
			tokens = append(tokens, goModToken{text: string(c), punctuation: true})
			index++
		case c == '"' || c == '`':
			end := index + 1
			for end < len(line) && line[end] != c {
				if c == '"' && line[end] == '\\' {
					end++
				}
				end++
			}
			if end >= len(line) {
				return nil, fmt.Errorf("a string that doesn't end on its line")
			}
			unquoted, err := strconv.Unquote(line[index : end+1])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, goModToken{text: unquoted, quoted: true})
			index = end + 1
		default:
			end := index
			for end < len(line) && !strings.ContainsRune(" \t\r()[]{},", rune(line[end])) && !strings.HasPrefix(line[end:], "//") {
				end++
			}
			tokens = append(tokens, goModToken{text: line[index:end]})
			index = end
		}
	}
	return tokens, nil
}

// withRecipe adds the recipe to a build's entries, in the order every machine reads them: by path, then kind.
func withRecipe(entries []readEntry, code recipe) ([]readEntry, string) {
	if code.reason != "" {
		return nil, code.reason
	}
	return sortedEntries(entries, code.entries, []readEntry{{Kind: "settings", Path: "binary"}}), ""
}

// refusedHandedRead is why a LeanProductOf build's set can't be kept (#zb9268t, Node's and Tests' rulings), or "": the
// window recorded an absolute read at or under one of its handed paths or its process's temporary directory. Settle's
// window is the reader that records each read, and the census never sees a path, so this is where the invariant holds:
// every read under a handed path is valued by its content (a product by its name key, an install by Loom's manifest, a
// file of the tree by content), and one keyed by its spelling would key per box and per run. Every other build is
// unaffected (entry.Temporary is empty).
func refusedHandedRead(entry journalEntry, set readSet) string {
	if entry.Temporary == "" {
		return ""
	}
	roots := append(slices.Clone(entry.Handed), entry.Temporary)
	for _, read := range set.Entries {
		if read.Kind != "absolute" {
			continue
		}
		for _, root := range roots {
			if read.Path == root || strings.HasPrefix(read.Path, root+string(filepath.Separator)) {
				return fmt.Sprintf("its window read %s under %s by its spelling, so its key would move per box and per run: a handed build reads there only what is valued by content", read.Path, root)
			}
		}
	}
	return ""
}
