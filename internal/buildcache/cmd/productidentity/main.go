// productidentity is the check that gates the shared tier of internal/buildcache: every product the gate builds
// through buildcache.Product must come out of the new path byte-identical to the fresh build main's code makes.
//
// It runs TestProduct_ units (by default stage 0, the checker and the oracle; see -packages and -run) four ways, each in its own empty local cache, against stores this
// command serves itself on 127.0.0.1 from a scratch directory (never a shared or remote store):
//
//   - fresh: ADAMIC_BUILD_CACHE=off, main's uncached mode, which builds every product into a new directory as main's
//     code does. Run as main's gate (ADAMIC_BUILD_STORE_TRUST=main), so each product it built is published to the
//     local "main" store, which is how this command reads what the fresh build made.
//   - cold: the cache on, an empty local cache and an empty local "shared" store: every product misses, is built and
//     published as a candidate.
//   - warm: another empty local cache, the "shared" store, no write credential: every product must be fetched.
//   - warm-main: another empty local cache reading the "main" store: what a candidate fetches from main's gate.
//
// One line per product follows, with its content digest in each mode (sha256 over every file's path, sha256 and
// mode, the form a manifest stores) and the seconds each mode's census recorded. Any difference names the files and
// their differing byte offsets, and a Mach-O's offsets are labelled when they fall in LC_UUID or the code signature.
// The command exits 1 when a product differs, is missing from a mode, isn't fetched warm, or a unit fails, so it
// can't pass on a partial run. With -flip it then flips one byte of a stored blob and requires the warm run of the
// package holding it to fail as cache poisoning with nothing served.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const token = "product-identity-local-token"

var modes = []string{"fresh", "cold", "warm", "warm-main"}

func main() {
	// By default the three products the first cut gates on (@system_adamic, Oct 9): stage 0, the checker archive and
	// the oracle, as bridge/tsgo builds them. -packages "" -run "" takes every TestProduct_ in the repository.
	packagesFlag := flag.String("packages", "./bridge/tsgo", "comma-separated package directories; empty means every package declaring a TestProduct_")
	runFlag := flag.String("run", "^TestProduct_(Stage0|TsgoArchive|Oracle)$", "only products whose test name matches this regular expression")
	scratchFlag := flag.String("scratch", "", "an empty directory for caches, stores and logs (required)")
	parallelFlag := flag.Int("parallel", 4, "go test -parallel for each package")
	timeoutFlag := flag.Duration("timeout", 60*time.Minute, "go test -timeout for each package in each mode")
	flipFlag := flag.Bool("flip", false, "after the comparison, flip one byte of a stored blob and require the warm run to refuse it")
	elsewhereFlag := flag.Bool("elsewhere", false, "after the comparison, require another checkout at another path and commit to agree with main's build, with "+buildFlags+" and (the mutant) without")
	flag.Parse()
	if *scratchFlag == "" {
		fail("-scratch is required")
	}
	scratch, err := filepath.Abs(*scratchFlag)
	check(err)
	if entries, err := os.ReadDir(scratch); err == nil && len(entries) > 0 {
		fail("%s is not empty: every mode needs empty caches and stores", scratch)
	}
	check(os.MkdirAll(scratch, 0o755))
	root, err := repositoryRoot()
	check(err)

	units, err := productUnits(root, *packagesFlag, *runFlag)
	check(err)
	if len(units) == 0 {
		fail("no TestProduct_ units selected")
	}
	total := 0
	for _, unit := range units {
		total += len(unit.tests)
	}
	fmt.Printf("products: %d TestProduct_ units in %d packages\n", total, len(units))

	server := &store{directory: filepath.Join(scratch, "stores")}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	check(err)
	go http.Serve(listener, server)
	address := "http://" + listener.Addr().String()
	tokenPath := filepath.Join(scratch, "publish-token")
	check(os.WriteFile(tokenPath, []byte(token+"\n"), 0o600))

	results := map[string]*modeResult{}
	healthy := true
	for _, mode := range modes {
		store := "shared"
		if mode == "fresh" || mode == "warm-main" {
			store = "main"
		}
		result := &modeResult{mode: mode, label: mode, store: store, cache: filepath.Join(scratch, mode, "cache"), log: filepath.Join(scratch, mode, "builds.log")}
		results[mode] = result
		for _, unit := range units {
			started := time.Now()
			outcome, err := result.run(root, scratch, address, tokenPath, unit, *parallelFlag, *timeoutFlag)
			check(err)
			fmt.Printf("ran %s %s: %d passed, %d failed, %d skipped, %.1fs\n", mode, unit.directory, outcome.passed, len(outcome.failed), outcome.skipped, time.Since(started).Seconds())
			if len(outcome.failed) > 0 {
				healthy = false
				fmt.Printf("FAIL %s %s: %s (log %s)\n", mode, unit.directory, strings.Join(outcome.failed, " "), outcome.log)
			}
		}
	}
	if !compare(server, results) {
		healthy = false
	}
	for _, conflict := range server.conflicts {
		healthy = false
		fmt.Printf("FAIL store: %s\n", conflict)
	}
	if *flipFlag {
		if !flip(root, scratch, address, server, results, units, *parallelFlag, *timeoutFlag) {
			healthy = false
		}
	}
	if *elsewhereFlag {
		if !elsewhere(root, scratch, address, tokenPath, server, units, *parallelFlag, *timeoutFlag) {
			healthy = false
		}
	}
	if !healthy {
		fmt.Println("product identity: FAIL")
		os.Exit(1)
	}
	fmt.Println("product identity: ok")
}

// A package's TestProduct_ units, as go test -list names them (the way the gate finds them).
type productUnit struct {
	directory string
	tests     []string
}

func productUnits(root, packages, run string) ([]productUnit, error) {
	var directories []string
	if packages != "" {
		directories = strings.Split(packages, ",")
	} else {
		// Every package directory whose tests declare a product; go test -list then confirms the names.
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "testdata" || entry.Name() == "cohere" && filepath.Dir(path) == root) {
				return filepath.SkipDir
			}
			if entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Contains(content, []byte("\nfunc TestProduct_")) {
				directory := "./" + filepath.ToSlash(mustRelative(root, filepath.Dir(path)))
				if len(directories) == 0 || directories[len(directories)-1] != directory {
					directories = append(directories, directory)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	var filter *regexp.Regexp
	if run != "" {
		filter = regexp.MustCompile(run)
	}
	var units []productUnit
	for _, directory := range directories {
		command := exec.Command("go", "test", "-list", "^TestProduct_", directory)
		command.Dir = root
		output, err := command.Output()
		if err != nil {
			return nil, fmt.Errorf("go test -list %s: %v", directory, err)
		}
		unit := productUnit{directory: directory}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "TestProduct_") && (filter == nil || filter.MatchString(line)) {
				unit.tests = append(unit.tests, line)
			}
		}
		if len(unit.tests) > 0 {
			units = append(units, unit)
		}
	}
	return units, nil
}

// modeResult is one way of running the product units: mode says how (fresh, cold, warm, warm-main), store which of
// this command's stores it reads or writes, label where its logs and temporary files go, and extra any environment
// it adds (the GOFLAGS the checkout comparison varies).
type modeResult struct {
	mode, label, store, cache, log string
	extra                          map[string]string
}

type runOutcome struct {
	passed, skipped int
	failed          []string
	log             string
	output          string
}

// run runs one package's product units in this mode, every store address and credential pointed at this command's
// own server, so nothing reaches a real store whatever the environment held.
func (m *modeResult) run(root, scratch, address, tokenPath string, unit productUnit, parallel int, timeout time.Duration) (runOutcome, error) {
	return m.runWith(root, scratch, address, tokenPath, unit, parallel, timeout, m.cache, m.log)
}

func (m *modeResult) runWith(root, scratch, address, tokenPath string, unit productUnit, parallel int, timeout time.Duration, cache, census string) (runOutcome, error) {
	temporary := filepath.Join(scratch, m.label, "tmp")
	for _, directory := range []string{cache, temporary} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return runOutcome{}, err
		}
	}
	environment := map[string]string{
		"ADAMIC_BUILD_CACHE_DIR":   cache,
		"ADAMIC_BUILD_LOG":         census,
		"ADAMIC_BUILD_AUDIT":       "0",
		"ADAMIC_BUILD_CACHE":       "",
		"ADAMIC_BUILD_STORE_TRUST": "",
		// No credential by default: a path that doesn't exist.
		"ADAMIC_BUILD_STORE_TOKEN": filepath.Join(scratch, "no-token"),
		// The off mode never deletes what it built, so it builds under the scratch directory.
		"TMPDIR": temporary,
	}
	switch m.mode {
	case "fresh":
		environment["ADAMIC_BUILD_CACHE"] = "off"
		environment["ADAMIC_BUILD_STORE_TRUST"] = "main"
		environment["ADAMIC_BUILD_STORE_TOKEN"] = tokenPath
		environment["ADAMIC_BUILD_STORE"] = address + "/" + m.store
		environment["ADAMIC_BUILD_STORE_WRITE"] = address + "/" + m.store + "/write"
	case "cold":
		environment["ADAMIC_BUILD_STORE_TOKEN"] = tokenPath
		environment["ADAMIC_BUILD_STORE"] = address + "/" + m.store
		environment["ADAMIC_BUILD_STORE_WRITE"] = address + "/" + m.store + "/write"
	case "warm", "warm-main":
		environment["ADAMIC_BUILD_STORE"] = address + "/" + m.store
		environment["ADAMIC_BUILD_STORE_WRITE"] = address + "/" + m.store + "/write"
	}
	for name, value := range m.extra {
		environment[name] = value
	}
	log := filepath.Join(scratch, m.label, strings.ReplaceAll(strings.TrimPrefix(unit.directory, "./"), "/", "_")+".jsonl")
	if cache != m.cache {
		log = strings.TrimSuffix(log, ".jsonl") + "-flipped.jsonl"
	}
	arguments := []string{"test", "-json", "-count=1", "-run", "^(" + strings.Join(unit.tests, "|") + ")$", "-parallel", strconv.Itoa(parallel), "-timeout", timeout.String(), unit.directory}
	command := exec.Command("go", arguments...)
	command.Dir = root
	command.Env = withEnvironment(os.Environ(), environment)
	output, _ := command.CombinedOutput()
	if err := os.WriteFile(log, output, 0o644); err != nil {
		return runOutcome{}, err
	}
	outcome := runOutcome{log: log, output: string(output)}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 1<<20), 1<<26)
	ended := map[string]bool{}
	for scanner.Scan() {
		var event struct{ Action, Test string }
		if json.Unmarshal(scanner.Bytes(), &event) != nil || event.Test == "" {
			continue
		}
		switch event.Action {
		case "pass":
			outcome.passed++
			ended[event.Test] = true
		case "fail":
			outcome.failed = append(outcome.failed, event.Test)
			ended[event.Test] = true
		case "skip":
			outcome.skipped++
			ended[event.Test] = true
		}
	}
	// A unit with no result (a build failure, a timeout's panic) is a failure too.
	for _, test := range unit.tests {
		if !ended[test] {
			outcome.failed = append(outcome.failed, test+"(no result)")
		}
	}
	return outcome, nil
}

func withEnvironment(base []string, overrides map[string]string) []string {
	var environment []string
	for _, entry := range base {
		name, _, _ := strings.Cut(entry, "=")
		if _, overridden := overrides[name]; !overridden {
			environment = append(environment, entry)
		}
	}
	names := make([]string, 0, len(overrides))
	for name := range overrides {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		environment = append(environment, name+"="+overrides[name])
	}
	return environment
}

// The products each mode holds, by key, with each one's content: a fresh product from the manifest main's gate
// published, a cached one from its directory in that mode's cache.
type product struct {
	name  string
	files []productFile
	// shape is the directory's full listing (directories, symlinks, files and their executable bit), which a manifest can't carry.
	shape string
	// content reads one file's bytes, for offsets when two modes differ.
	content func(path string) ([]byte, error)
}

type productFile struct {
	path, sha256 string
	mode         uint32
}

func (p product) digest() string {
	hash := sha256.New()
	for _, file := range p.files {
		fmt.Fprintf(hash, "%s %s %o\n", file.path, file.sha256, file.mode)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func compare(server *store, results map[string]*modeResult) bool {
	products := map[string]map[string]product{}
	fresh, err := server.products("main", "build")
	check(err)
	products["fresh"] = fresh
	for _, mode := range modes[1:] {
		cached, err := cachedProducts(results[mode].cache)
		check(err)
		products[mode] = cached
	}
	seconds := map[string]map[string]string{}
	outcomes := map[string]map[string][]string{}
	for _, mode := range modes {
		seconds[mode], outcomes[mode] = census(results[mode].log)
	}
	keys := map[string]bool{}
	for _, mode := range modes {
		for key := range products[mode] {
			keys[key] = true
		}
	}
	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return nameOf(products, sorted[i])+sorted[i] < nameOf(products, sorted[j])+sorted[j]
	})
	healthy := true
	identical := 0
	for _, key := range sorted {
		name := nameOf(products, key)
		var line strings.Builder
		fmt.Fprintf(&line, "product %s %s", strings.ReplaceAll(name, " ", "_"), key[:12])
		digests := map[string]string{}
		for _, mode := range modes {
			digest := "missing"
			if found, ok := products[mode][key]; ok {
				digest = found.digest()
				digests[mode] = digest
				digest = digest[:16]
			}
			fmt.Fprintf(&line, " %s=%s", mode, digest)
		}
		line.WriteString(" seconds")
		for _, mode := range modes {
			value := seconds[mode][key[:12]]
			if mode == "fresh" {
				value = seconds[mode][strings.ReplaceAll(name, " ", "_")]
			}
			if value == "" {
				value = "-"
			}
			fmt.Fprintf(&line, " %s=%s", mode, value)
		}
		var problems []string
		for _, mode := range modes {
			if _, ok := digests[mode]; !ok {
				problems = append(problems, "missing from "+mode)
			}
		}
		for _, mode := range []string{"warm", "warm-main"} {
			for _, outcome := range outcomes[mode][key[:12]] {
				if outcome != "fetched" && outcome != "hit" {
					problems = append(problems, mode+" "+outcome+", not fetched")
				}
			}
		}
		reference := ""
		for _, mode := range modes {
			if digest, ok := digests[mode]; ok {
				if reference == "" {
					reference = mode
				} else if digest != digests[reference] {
					problems = append(problems, mode+" differs from "+reference+": "+difference(products[reference][key], products[mode][key]))
				}
			}
		}
		// A directory's shape (empty directories, executable bits) can't travel through a manifest: a fetched product
		// must still have the shape the built one had.
		if cold, ok := products["cold"][key]; ok {
			for _, mode := range []string{"warm", "warm-main"} {
				if fetched, ok := products[mode][key]; ok && fetched.shape != cold.shape {
					problems = append(problems, mode+"'s directory shape differs from cold's: "+shapeDifference(cold.shape, fetched.shape))
				}
			}
		}
		if len(problems) == 0 {
			identical++
			line.WriteString(" identical")
		} else {
			healthy = false
			line.WriteString(" DIFFERS: " + strings.Join(problems, "; "))
		}
		fmt.Println(line.String())
	}
	fmt.Printf("compared %d products: %d identical in all four modes, %d not\n", len(sorted), identical, len(sorted)-identical)
	if len(sorted) == 0 {
		healthy = false
	}
	return healthy
}

func nameOf(products map[string]map[string]product, key string) string {
	for _, mode := range modes {
		if found, ok := products[mode][key]; ok && found.name != "" {
			return found.name
		}
	}
	return "unnamed"
}

// census reads a mode's build log: seconds by key prefix (and by name, since the off mode's lines carry no key), and
// each key's outcomes.
func census(path string) (map[string]string, map[string][]string) {
	seconds, outcomes := map[string]string{}, map[string][]string{}
	content, _ := os.ReadFile(path)
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 5 || fields[0] != "build" {
			continue
		}
		name, key, outcome, value := fields[1], fields[2], fields[3], fields[4]
		index := key
		if outcome == "off" {
			index = name
		}
		if seconds[index] == "" {
			seconds[index] = value
		} else {
			seconds[index] += "+" + value
		}
		outcomes[key] = append(outcomes[key], outcome)
	}
	return seconds, outcomes
}

var keyPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func cachedProducts(cache string) (map[string]product, error) {
	products := map[string]product{}
	entries, err := os.ReadDir(cache)
	if err != nil {
		return products, nil
	}
	for _, entry := range entries {
		if !entry.IsDir() || !keyPattern.MatchString(entry.Name()) {
			continue
		}
		directory := filepath.Join(cache, entry.Name())
		found, err := readDirectory(directory)
		if err != nil {
			return nil, err
		}
		if described, err := os.ReadFile(directory + ".inputs"); err == nil {
			first, _, _ := strings.Cut(string(described), "\n")
			found.name = strings.TrimPrefix(first, "name ")
		}
		products[entry.Name()] = found
	}
	return products, nil
}

func readDirectory(directory string) (product, error) {
	var found product
	var shape []string
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative := filepath.ToSlash(mustRelative(directory, path))
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		// The kind and, for a file, whether it's executable: what the store keeps (store.go writes 0755 or 0644). The rest
		// of a mode is the building user's umask (a gate box's 002 made 0775 where a fetch makes 0755, Oct 9 17:35Z).
		kind := info.Mode().Type().String()
		if info.Mode().IsRegular() {
			kind = "-rw-r--r--"
			if info.Mode()&0o111 != 0 {
				kind = "-rwxr-xr-x"
			}
		}
		shape = append(shape, relative+" "+kind)
		if !info.Mode().IsRegular() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		// The mode as a manifest stores it, so the content digest of a directory and of a manifest compare.
		mode := uint32(0o644)
		if info.Mode()&0o111 != 0 {
			mode = 0o755
		}
		found.files = append(found.files, productFile{path: relative, sha256: hex.EncodeToString(sum[:]), mode: mode})
		return nil
	})
	sort.Slice(found.files, func(i, j int) bool { return found.files[i].path < found.files[j].path })
	found.shape = strings.Join(shape, "\n")
	found.content = func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join(directory, filepath.FromSlash(path)))
	}
	return found, err
}

func shapeDifference(want, got string) string {
	wantLines, gotLines := map[string]bool{}, map[string]bool{}
	for _, line := range strings.Split(want, "\n") {
		wantLines[line] = true
	}
	for _, line := range strings.Split(got, "\n") {
		gotLines[line] = true
	}
	var differences []string
	for line := range wantLines {
		if !gotLines[line] {
			differences = append(differences, "built has "+line)
		}
	}
	for line := range gotLines {
		if !wantLines[line] {
			differences = append(differences, "fetched has "+line)
		}
	}
	sort.Strings(differences)
	if len(differences) > 6 {
		differences = append(differences[:6], fmt.Sprintf("and %d more", len(differences)-6))
	}
	return strings.Join(differences, ", ")
}

// difference names the files two products disagree on, with the byte offsets where they differ.
func difference(want, got product) string {
	wanted := map[string]productFile{}
	for _, file := range want.files {
		wanted[file.path] = file
	}
	var parts []string
	seen := map[string]bool{}
	for _, file := range got.files {
		seen[file.path] = true
		other, ok := wanted[file.path]
		switch {
		case !ok:
			parts = append(parts, file.path+" only in the second")
		case other.mode != file.mode:
			parts = append(parts, fmt.Sprintf("%s mode %o, then %o", file.path, other.mode, file.mode))
		case other.sha256 != file.sha256:
			parts = append(parts, file.path+" "+offsets(want, got, file.path))
		}
	}
	for _, file := range want.files {
		if !seen[file.path] {
			parts = append(parts, file.path+" only in the first")
		}
	}
	return strings.Join(parts, ", ")
}

func offsets(want, got product, path string) string {
	first, err := want.content(path)
	if err != nil {
		return "(unreadable: " + err.Error() + ")"
	}
	second, err := got.content(path)
	if err != nil {
		return "(unreadable: " + err.Error() + ")"
	}
	var ranges [][2]int
	for index := 0; index < min(len(first), len(second)); index++ {
		if first[index] == second[index] {
			continue
		}
		if len(ranges) > 0 && ranges[len(ranges)-1][1] >= index-16 {
			ranges[len(ranges)-1][1] = index
		} else {
			ranges = append(ranges, [2]int{index, index})
		}
	}
	regions := machoRegions(first)
	var described []string
	for index, span := range ranges {
		if index == 12 {
			described = append(described, fmt.Sprintf("and %d more ranges", len(ranges)-12))
			break
		}
		label := ""
		for _, region := range regions {
			if span[0] >= region.start && span[1] < region.end {
				label = " (" + region.name + ")"
			}
		}
		described = append(described, fmt.Sprintf("0x%x-0x%x%s", span[0], span[1], label))
	}
	summary := fmt.Sprintf("differs: %d and %d bytes", len(first), len(second))
	if len(described) > 0 {
		summary += ", at " + strings.Join(described, " ")
	}
	return summary
}

type region struct {
	name       string
	start, end int
}

// machoRegions finds LC_UUID's bytes and the code signature's in a Mach-O, the two places a macOS link may differ
// between two links of the same objects (#cchsq45).
func machoRegions(content []byte) []region {
	file, err := macho.NewFile(bytes.NewReader(content))
	if err != nil {
		return nil
	}
	defer file.Close()
	var regions []region
	offset := 32
	if file.Magic == macho.Magic32 {
		offset = 28
	}
	for _, load := range file.Loads {
		raw := load.Raw()
		if len(raw) < 8 {
			break
		}
		command := file.ByteOrder.Uint32(raw[0:4])
		switch command {
		case 0x1b: // LC_UUID
			regions = append(regions, region{"LC_UUID", offset + 8, offset + len(raw)})
		case 0x1d: // LC_CODE_SIGNATURE
			start := int(file.ByteOrder.Uint32(raw[8:12]))
			size := int(file.ByteOrder.Uint32(raw[12:16]))
			regions = append(regions, region{"code signature", start, start + size})
			// The load command itself carries the signature's offset and size.
			regions = append(regions, region{"LC_CODE_SIGNATURE", offset, offset + len(raw)})
		}
		offset += len(raw)
	}
	return regions
}

// flip proves the warm path refuses a stored blob that isn't its hash, on a real product: one byte of one blob in the
// shared store is flipped (its length kept, so only the hash can tell), and the warm run of a package holding that
// product, in a new empty cache, must fail as cache poisoning with nothing of that key in the cache.
func flip(root, scratch, address string, server *store, results map[string]*modeResult, units []productUnit, parallel int, timeout time.Duration) bool {
	stored, err := server.products("shared", "build-candidate")
	check(err)
	// The package whose warm run was quickest, and a product of it.
	warm := results["warm"]
	type candidate struct {
		unit    productUnit
		key     string
		seconds float64
	}
	var chosen *candidate
	for _, unit := range units {
		log := filepath.Join(scratch, "warm", strings.ReplaceAll(strings.TrimPrefix(unit.directory, "./"), "/", "_")+".jsonl")
		content, err := os.ReadFile(log)
		if err != nil {
			continue
		}
		elapsed := packageElapsed(content)
		keys := keysFetchedBy(content)
		for _, key := range keys {
			for full := range stored {
				if strings.HasPrefix(full, key) && (chosen == nil || elapsed < chosen.seconds) {
					chosen = &candidate{unit, full, elapsed}
				}
			}
		}
	}
	if chosen == nil {
		fmt.Println("FAIL flip: no warm-fetched product found to flip (the census lines aren't in the test output)")
		return false
	}
	victim := stored[chosen.key]
	// The product's largest file: its point, and never an empty blob with no byte to flip.
	var file productFile
	var content []byte
	for _, candidate := range victim.files {
		held, err := os.ReadFile(filepath.Join(server.directory, "shared", "blobs", candidate.sha256))
		check(err)
		if len(held) > len(content) {
			file, content = candidate, held
		}
	}
	if len(content) == 0 {
		fmt.Printf("FAIL flip: product %s has no file with a byte to flip\n", chosen.key[:12])
		return false
	}
	blob := filepath.Join(server.directory, "shared", "blobs", file.sha256)
	position := len(content) / 2
	content[position] ^= 0x01
	check(os.WriteFile(blob, content, 0o644))
	cache := filepath.Join(scratch, "warm-flipped", "cache")
	check(os.MkdirAll(cache, 0o755))
	outcome, err := warm.runWith(root, scratch, address, filepath.Join(scratch, "no-token"), chosen.unit, parallel, timeout, cache, filepath.Join(scratch, "warm-flipped", "builds.log"))
	check(err)
	_, served := os.Stat(filepath.Join(cache, chosen.key))
	poisoned := strings.Contains(outcome.output, "cache poisoning") && strings.Contains(outcome.output, file.sha256)
	fmt.Printf("flip: byte %d of %s (%s, blob %s) in product %s %s; warm run of %s: %d passed, %d failed; poisoning reported %t; served %t\n",
		position, file.path, victim.name, file.sha256[:12], chosen.key[:12], victim.name, chosen.unit.directory, outcome.passed, len(outcome.failed), poisoned, served == nil)
	if !poisoned || len(outcome.failed) == 0 || served == nil {
		fmt.Println("FAIL flip: a flipped blob was not refused")
		return false
	}
	return true
}

// keysFetchedBy is every key a go test -json log's census lines name (each Product call logs its line).
func keysFetchedBy(output []byte) []string {
	var keys []string
	pattern := regexp.MustCompile(`build \S+ ([0-9a-f]{12}) (fetched|hit) `)
	for _, match := range pattern.FindAllSubmatch(output, -1) {
		keys = append(keys, string(match[1]))
	}
	return keys
}

func packageElapsed(output []byte) float64 {
	elapsed := 1e9
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 1<<20), 1<<26)
	for scanner.Scan() {
		var event struct {
			Action, Test string
			Elapsed      float64
		}
		if json.Unmarshal(scanner.Bytes(), &event) == nil && event.Test == "" && (event.Action == "pass" || event.Action == "fail") {
			elapsed = event.Elapsed
		}
	}
	return elapsed
}

// store serves stores the way Loom's does, from a scratch directory: /<store>/blobs/<sha256> and
// /<store>/refs/<namespace>/<key> read with no credential, written under /<store>/write with the bearer token. A blob
// must be its hash, a ref must name a held blob, and a ref never changes: a second, different product for a key is a
// conflict, which this command fails on (a build that isn't reproducible, or a key that isn't honest).
var storeName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

type store struct {
	directory string
	mutex     sync.Mutex
	conflicts []string
}

func (s *store) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	parts := strings.SplitN(strings.TrimPrefix(request.URL.Path, "/"), "/", 2)
	if len(parts) != 2 || !storeName.MatchString(parts[0]) || strings.Contains(request.URL.Path, "..") {
		http.NotFound(writer, request)
		return
	}
	name, object := parts[0], parts[1]
	if request.Method == http.MethodPut {
		s.write(writer, request, name, strings.TrimPrefix(object, "write/"))
		return
	}
	content, err := os.ReadFile(filepath.Join(s.directory, name, filepath.FromSlash(object)))
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	writer.Write(content)
}

func (s *store) write(writer http.ResponseWriter, request *http.Request, name, object string) {
	if request.Header.Get("Authorization") != "Bearer "+token {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return
	}
	content, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	path := filepath.Join(s.directory, name, filepath.FromSlash(object))
	switch {
	case strings.HasPrefix(object, "blobs/"):
		if sum := sha256.Sum256(content); "blobs/"+hex.EncodeToString(sum[:]) != object {
			http.Error(writer, "a blob must be its hash", http.StatusBadRequest)
			return
		}
	case strings.HasPrefix(object, "refs/"):
		if _, err := os.Stat(filepath.Join(s.directory, name, "blobs", string(content))); err != nil {
			http.Error(writer, "a ref must name a held blob", http.StatusConflict)
			return
		}
		if held, err := os.ReadFile(path); err == nil {
			if !bytes.Equal(held, content) {
				s.conflicts = append(s.conflicts, fmt.Sprintf("%s %s was published as manifest %s, then as %s", name, object, held, content))
				http.Error(writer, "a ref never changes", http.StatusConflict)
				return
			}
			writer.WriteHeader(http.StatusOK)
			return
		}
	default:
		http.NotFound(writer, request)
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusCreated)
}

// products reads every product a store's namespace refers to, from its manifest.
func (s *store) products(name, namespace string) (map[string]product, error) {
	products := map[string]product{}
	references, err := os.ReadDir(filepath.Join(s.directory, name, "refs", namespace))
	if errors.Is(err, fs.ErrNotExist) {
		return products, nil
	} else if err != nil {
		return nil, err
	}
	blobs := filepath.Join(s.directory, name, "blobs")
	for _, reference := range references {
		sum, err := os.ReadFile(filepath.Join(s.directory, name, "refs", namespace, reference.Name()))
		if err != nil {
			return nil, err
		}
		encoded, err := os.ReadFile(filepath.Join(blobs, strings.TrimSpace(string(sum))))
		if err != nil {
			return nil, err
		}
		var stored struct {
			Key, Name string
			Files     []struct {
				Path, SHA256 string
				Mode         uint32
			}
		}
		if err := json.Unmarshal(encoded, &stored); err != nil {
			return nil, err
		}
		found := product{name: stored.Name, content: func(path string) ([]byte, error) {
			for _, file := range stored.Files {
				if file.Path == path {
					return os.ReadFile(filepath.Join(blobs, file.SHA256))
				}
			}
			return nil, fs.ErrNotExist
		}}
		for _, file := range stored.Files {
			found.files = append(found.files, productFile{path: file.Path, sha256: file.SHA256, mode: file.Mode})
		}
		sort.Slice(found.files, func(i, j int) bool { return found.files[i].path < found.files[j].path })
		products[reference.Name()] = found
	}
	return products, nil
}

// repositoryRoot is the nearest directory above the working directory holding adamic's go.mod, as buildcache finds it.
func repositoryRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		content, err := os.ReadFile(filepath.Join(directory, "go.mod"))
		if err == nil && strings.HasPrefix(string(content), "module github.com/system-inc/adamic\n") {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", errors.New("no adamic go.mod above the working directory")
		}
		directory = parent
	}
}

func mustRelative(root, path string) string {
	relative, _ := filepath.Rel(root, path)
	return relative
}

func check(err error) {
	if err != nil {
		fail("%v", err)
	}
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "productidentity: "+format+"\n", arguments...)
	os.Exit(2)
}

// buildFlags is what makes a Go product's bytes a function of its key (@system_adamic, Oct 9): without them a binary
// stamps the commit it was built at and the path of the checkout, inputs no key sees. cloud/setup.sh exports them.
const buildFlags = "-buildvcs=false -trimpath"

// elsewhere proves a candidate at another commit and another checkout path gets main's product: a second checkout of
// the same commit, at a different path, with one unrelated commit on top, builds every product fresh and fetches main's
// (warm-main), and the two must be the same bytes as main's own fresh build. It runs twice: with buildFlags, where the
// three must agree and the fetch must hit main's key, and without them (GOFLAGS=-buildvcs=auto, which also overrides a
// -trimpath from the go env file), the mutant, where the fresh builds must differ, or the first run proves nothing.
// On macOS a product holding an archive is reported, not required: macOS's ar stamps the archive's symbol table with
// the time it ran, so two builds of an archive differ there anywhere (byte-identical with ZERO_AR_DATE=1).
func elsewhere(root, scratch, address, tokenPath string, server *store, units []productUnit, parallel int, timeout time.Duration) bool {
	if status, err := gitOutput(root, "status", "--porcelain", "--untracked-files=no", "--ignore-submodules=all"); err != nil || status != "" {
		fmt.Printf("FAIL elsewhere: the checkout must be committed and clean to be cloned exactly (%v): %s\n", err, status)
		return false
	}
	head, err := gitOutput(root, "rev-parse", "HEAD")
	check(err)
	other := filepath.Join(scratch, "elsewhere", "another-checkout-path", "adamic")
	check(os.MkdirAll(filepath.Dir(other), 0o755))
	for _, arguments := range [][]string{
		{"clone", "-q", "--shared", "--no-checkout", root, other},
		{"-C", other, "checkout", "-q", "--detach", head},
		{"-C", other, "-c", "user.name=productidentity", "-c", "user.email=productidentity@localhost", "commit", "-q", "--no-verify", "--allow-empty", "-m", "An unrelated commit"},
	} {
		if output, err := exec.Command("git", arguments...).CombinedOutput(); err != nil {
			fmt.Printf("FAIL elsewhere: git %s: %v\n%s", strings.Join(arguments, " "), err, output)
			return false
		}
	}
	// The cohere submodule's files, linked in (copied across file systems), without its git directory.
	check(linkTree(filepath.Join(root, "cohere"), filepath.Join(other, "cohere")))
	otherHead, err := gitOutput(other, "rev-parse", "HEAD")
	check(err)
	fmt.Printf("elsewhere: %s at %s, another checkout at %s at %s\n", root, head[:10], other, otherHead[:10])

	healthy := true
	for _, variant := range []struct {
		name, goFlags string
		agree         bool
	}{{"with-flags", buildFlags, true}, {"without-flags", "-buildvcs=auto", false}} {
		extra := map[string]string{"GOFLAGS": variant.goFlags}
		base := filepath.Join("elsewhere", variant.name)
		runs := []struct {
			directory string
			mode      *modeResult
		}{
			{root, &modeResult{mode: "fresh", label: base + "/main-fresh", store: "main-" + variant.name, extra: extra}},
			{other, &modeResult{mode: "fresh", label: base + "/other-fresh", store: "other-" + variant.name, extra: extra}},
			{other, &modeResult{mode: "warm-main", label: base + "/other-warm-main", store: "main-" + variant.name, extra: extra}},
		}
		for _, run := range runs {
			run.mode.cache = filepath.Join(scratch, run.mode.label, "cache")
			run.mode.log = filepath.Join(scratch, run.mode.label, "builds.log")
			for _, unit := range units {
				started := time.Now()
				outcome, err := run.mode.run(run.directory, scratch, address, tokenPath, unit, parallel, timeout)
				check(err)
				fmt.Printf("ran %s %s in %s: %d passed, %d failed, %.1fs\n", variant.name, run.mode.label, run.directory, outcome.passed, len(outcome.failed), time.Since(started).Seconds())
				if len(outcome.failed) > 0 {
					healthy = false
					fmt.Printf("FAIL elsewhere %s %s: %s (log %s)\n", variant.name, run.mode.label, strings.Join(outcome.failed, " "), outcome.log)
				}
			}
		}
		mainFresh, err := server.products("main-"+variant.name, "build")
		check(err)
		otherFresh, err := server.products("other-"+variant.name, "build")
		check(err)
		otherWarm, err := cachedProducts(runs[2].mode.cache)
		check(err)
		_, warmOutcomes := census(runs[2].mode.log)
		byName := func(products map[string]product) map[string]string {
			keys := map[string]string{}
			for key, found := range products {
				keys[found.name] = key
			}
			return keys
		}
		otherKeys := byName(otherFresh)
		names := make([]string, 0, len(mainFresh))
		for _, found := range mainFresh {
			names = append(names, found.name)
		}
		sort.Strings(names)
		if len(names) == 0 {
			healthy = false
			fmt.Printf("FAIL elsewhere %s: main's fresh build published nothing\n", variant.name)
		}
		for _, name := range names {
			publishedKey := byName(mainFresh)[name]
			published := mainFresh[publishedKey]
			line := fmt.Sprintf("elsewhere %s %s main=%s %s", variant.name, strings.ReplaceAll(name, " ", "_"), publishedKey[:12], published.digest()[:16])
			otherKey, ok := otherKeys[name]
			if !ok {
				healthy = false
				fmt.Println(line + " FAIL: the other checkout's fresh build didn't publish it")
				continue
			}
			other := otherFresh[otherKey]
			line += fmt.Sprintf(" other=%s %s", otherKey[:12], other.digest()[:16])
			fetched, wasFetched := otherWarm[publishedKey]
			if wasFetched {
				line += " warm-main=" + fetched.digest()[:16] + " (" + strings.Join(warmOutcomes[publishedKey[:12]], ",") + ")"
			} else {
				line += " warm-main=not fetched (the other checkout's key differs)"
			}
			exempt := runtime.GOOS == "darwin" && holdsArchive(published)
			same := published.digest() == other.digest()
			switch {
			case exempt:
				line += " not required on darwin: macOS's ar stamps an archive's symbol table with the time it ran"
				if !same {
					line += "; differs: " + difference(published, other)
				}
			case variant.agree && same && wasFetched && fetched.digest() == published.digest():
				line += " identical"
			case variant.agree:
				healthy = false
				line += " FAIL"
				if !same {
					line += ": the other checkout's fresh build differs: " + difference(published, other)
				}
				if !wasFetched {
					line += ": the other checkout didn't fetch main's product, so its key holds the checkout's path or commit"
				}
			case same:
				healthy = false
				line += " FAIL: without " + buildFlags + " the two checkouts still agree, so the run with them proves nothing"
			default:
				line += " differs, as it must without " + buildFlags + ": " + difference(published, other)
			}
			fmt.Println(line)
		}
	}
	return healthy
}

func holdsArchive(found product) bool {
	for _, file := range found.files {
		if strings.HasSuffix(file.path, ".a") {
			return true
		}
	}
	return false
}

func gitOutput(directory string, arguments ...string) (string, error) {
	output, err := exec.Command("git", append([]string{"-C", directory}, arguments...)...).Output()
	return strings.TrimSpace(string(output)), err
}

// linkTree fills target with source's files, hard-linked where it can and copied where it can't, leaving out .git.
func linkTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		destination := filepath.Join(target, mustRelative(source, path))
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			return os.MkdirAll(destination, info.Mode().Perm()|0o700)
		case info.Mode()&fs.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, destination)
		}
		if os.Link(path, destination) == nil {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, content, info.Mode().Perm())
	})
}
