// adamic-gate partitions the repository gate and checks its coverage from raw evidence.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const timingPath = "cmd/adamic-gate/timings.json"

type unit struct {
	Package string
	Test    string
	Shard   int
	Seconds float64
}

func (u unit) key() string { return u.Package + "::" + u.Test }

type plan struct {
	Version int
	Commit  string
	Source  string
	Count   int
	Units   []unit
	Digest  string
	Shards  []prediction
}
type event struct {
	Action, Package, Test, Output string
	Elapsed                       float64
}
type result struct {
	Package, Test, Action, Output string
	Reason                        string `json:",omitempty"`
	Seconds                       float64
}

func (r result) key() string { return r.Package + "::" + r.Test }

type invocation struct {
	Package  string
	Args     []string
	Uncached bool
	Exit     int
	Seconds  float64
}
type summary struct {
	Version               int
	Plan                  plan
	Index                 int
	Invocations           []invocation
	Checks                map[string]string
	Results               []result
	CacheLines            []string
	Pass, Fail, Skip      int
	WallSeconds           float64
	BuildFlags            string
	DiskBefore, DiskAfter uint64
	Errors                []string
}
type merged struct {
	Version           int
	Plan              plan
	Results           []result
	Pass, Fail, Skip  int
	TestEvents        int
	RawTerminalEvents int
	WallSeconds       []float64
	BuildFlags        []string
	Errors            []string
	Green             bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "adamic-gate:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: adamic-gate plan|shard|merge|compare|timings")
	}
	switch args[0] {
	case "plan":
		flags := flag.NewFlagSet("plan", flag.ContinueOnError)
		count := flags.Int("count", 8, "number of shards")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		p, err := makePlan(*count)
		if err != nil {
			return err
		}
		return writeJSON(os.Stdout, p)
	case "shard":
		flags := flag.NewFlagSet("shard", flag.ContinueOnError)
		count := flags.Int("count", 8, "number of shards")
		index := flags.Int("index", -1, "zero-based shard")
		out := flags.String("out", "", "evidence directory")
		resume := flags.Bool("resume", false, "reuse complete package evidence for identical execution inputs")
		scratch := flags.String("scratch", "/workspace/adamic-gate-scratch", "disk-backed scratch root outside the repository")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return shard(*index, *count, *out, *scratch, *resume)
	case "merge":
		flags := flag.NewFlagSet("merge", flag.ContinueOnError)
		out := flags.String("out", "merged", "merged evidence directory")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return merge(flags.Args(), *out)
	case "compare":
		if len(args) != 3 {
			return errors.New("compare <merged directory or JSON> <unsharded JSON log>")
		}
		return compare(args[1], args[2])
	case "timings":
		flags := flag.NewFlagSet("timings", flag.ContinueOnError)
		out := flags.String("out", timingPath, "timing file")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 1 {
			return errors.New("timings [-out file] <go test -json log>")
		}
		r, _, _, err := readLog(flags.Arg(0))
		if err != nil {
			return err
		}
		weights := map[string]float64{}
		for _, v := range r {
			weights[v.key()] = v.Seconds
		}
		return saveJSON(*out, weights)
	}
	return fmt.Errorf("unknown command %q", args[0])
}
func output(name string, args ...string) (string, error) {
	b, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w\n%s", name, args, err, b)
	}
	return strings.TrimSpace(string(b)), nil
}
func writeJSON(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}
func saveJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = writeJSON(f, v)
	closeErr := f.Close()
	return errors.Join(err, closeErr)
}
func loadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// Read literal fixture rows from the test's AST, then verify each named input exists. A directory
// glob alone would accidentally include helpers and probes the parent does not execute.
func fixtureRows(file, variable string) ([]string, error) {
	tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return nil, err
	}
	var names []string
	ast.Inspect(tree, func(n ast.Node) bool {
		declaration, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range declaration.Names {
			if name.Name != variable {
				continue
			}
			if i >= len(declaration.Values) {
				continue
			}
			literal, ok := declaration.Values[i].(*ast.CompositeLit)
			if !ok {
				continue
			}
			for _, element := range literal.Elts {
				row, ok := element.(*ast.CompositeLit)
				if !ok || len(row.Elts) == 0 {
					continue
				}
				value, ok := row.Elts[0].(*ast.BasicLit)
				if !ok || value.Kind != token.STRING {
					continue
				}
				s, e := strconv.Unquote(value.Value)
				if e != nil {
					continue
				}
				names = append(names, s)
			}
		}
		return true
	})
	if len(names) == 0 {
		return nil, fmt.Errorf("no fixture rows for %s in %s", variable, file)
	}
	for _, name := range names {
		if _, err := os.Stat(name); err != nil {
			return nil, err
		}
	}
	return names, nil
}

// Literal first fields of the anonymous slice iterated directly by a test. This is deliberately
// restricted to audited independent parents; arbitrary AST inference cannot prove selectability.
func literalChildren(file, parent string, table ...string) ([]string, error) {
	tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, d := range tree.Decls {
		f, ok := d.(*ast.FuncDecl)
		if !ok || f.Name.Name != parent {
			continue
		}
		ast.Inspect(f.Body, func(n ast.Node) bool {
			r, ok := n.(*ast.RangeStmt)
			if !ok {
				return true
			}
			literal, ok := r.X.(*ast.CompositeLit)
			if !ok && len(table) == 1 {
				if id, yes := r.X.(*ast.Ident); yes && id.Name == table[0] {
					ast.Inspect(f.Body, func(node ast.Node) bool {
						assignment, yes := node.(*ast.AssignStmt)
						if !yes {
							return true
						}
						for i, lhs := range assignment.Lhs {
							id, yes := lhs.(*ast.Ident)
							if yes && id.Name == table[0] && i < len(assignment.Rhs) {
								literal, ok = assignment.Rhs[i].(*ast.CompositeLit)
							}
						}
						return true
					})
				}
			}
			if !ok {
				return true
			}
			if _, ok := literal.Type.(*ast.ArrayType); !ok {
				return true
			}
			for _, e := range literal.Elts {
				if label, ok := e.(*ast.BasicLit); ok && label.Kind == token.STRING {
					name, err := strconv.Unquote(label.Value)
					if err == nil {
						names = append(names, strings.ReplaceAll(name, " ", "_"))
					}
					continue
				}
				row, ok := e.(*ast.CompositeLit)
				if !ok || len(row.Elts) == 0 {
					continue
				}
				s, ok := row.Elts[0].(*ast.BasicLit)
				if !ok || s.Kind != token.STRING {
					continue
				}
				name, e := strconv.Unquote(s.Value)
				if e == nil {
					names = append(names, strings.ReplaceAll(name, " ", "_"))
				}
			}
			return false
		})
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("cannot enumerate %s", parent)
	}
	return names, nil
}
func children(pkg, parent string) ([]string, error) {
	if strings.HasSuffix(pkg, "/cmd/adamic-test262") {
		switch parent {
		case "TestLargeCompilerOutputIsComplete":
			return literalChildren("cmd/adamic-test262/corpus_test.go", parent)
		case "TestParallelCachedMatchesSerial":
			return literalChildren("cmd/adamic-test262/cache_test.go", parent)
		}
	}
	if strings.HasSuffix(pkg, "/stage1/cohere/typeaware") && parent == "TestVolumeAgreementAndMutants" {
		return literalChildren("stage1/cohere/typeaware/volume_test.go", parent, "changes")
	}
	if strings.HasSuffix(pkg, "/internal/oracle") {
		switch parent {
		case "TestNativeAgreesWithNode":
			return fixtureRows("internal/oracle/oracle_test.go", "fixtures")
		case "TestInputAgreesWithNode":
			return fixtureRows("internal/oracle/input_test.go", "inputFixtures")
		case "TestFreshWriteProbesStayRefused":
			paths, err := filepath.Glob("internal/oracle/testdata/fresh_refused/*.a")
			for i := range paths {
				paths[i] = filepath.Base(paths[i])
			}
			if len(paths) < 20 {
				return nil, errors.New("fresh probes disappeared")
			}
			return paths, err
		}
	}
	if strings.HasSuffix(pkg, "/stage1/cohere/css") && parent == "TestCSSPrinterAgreesWithGo" {
		return literalChildren("stage1/cohere/css/print_test.go", parent)
	}
	if strings.HasSuffix(pkg, "/stage1/cohere/lint") {
		switch parent {
		case "TestMutants":
			return literalChildren("stage1/cohere/lint/lint_test.go", parent)
		case "TestVolumeMutants":
			return literalChildren("stage1/cohere/lint/volume_test.go", parent)
		}
	}
	if strings.HasSuffix(pkg, "/internal/native") {
		switch parent {
		case "TestNormalizeMatchesNode":
			return literalChildren("internal/native/normalize_test.go", parent)
		case "TestStringIndexMatchesNode":
			return literalChildren("internal/native/string_index_test.go", parent)
		}
	}
	return nil, nil
}
func sourceIdentity() (string, error) {
	tracked, err := output("git", "ls-files", "-s", "-z")
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, line := range strings.Split(tracked, "\x00") {
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) != 2 {
			continue
		}
		name := fields[1]
		fmt.Fprintf(h, "%d:%s", len(line), line)
		if strings.HasPrefix(line, "160000 ") {
			sha, err := output("git", "-C", name, "rev-parse", "HEAD")
			if err != nil {
				return "", err
			}
			fmt.Fprint(h, sha)
			continue
		}
		var b []byte
		if strings.HasPrefix(line, "120000 ") {
			target, err := os.Readlink(name)
			if err != nil {
				return "", err
			}
			b = []byte(target)
		} else {
			var err error
			b, err = os.ReadFile(name)
			if err != nil {
				return "", err
			}
		}
		fmt.Fprintf(h, "%d:", len(b))
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func makePlan(count int) (plan, error) {
	p := plan{Version: 1, Count: count}
	status, err := output("git", "status", "--porcelain", "--untracked-files=normal")
	if err != nil {
		return p, err
	}
	if status != "" {
		return p, errors.New("planning requires a clean checkout, including submodules; keep evidence outside the repository")
	}
	if count < 1 {
		return p, errors.New("count must be positive")
	}
	p.Commit, err = output("git", "rev-parse", "HEAD")
	if err != nil {
		return p, err
	}
	p.Source, err = sourceIdentity()
	if err != nil {
		return p, err
	}
	var weights map[string]float64
	if err := loadJSON(timingPath, &weights); err != nil {
		return p, err
	}
	packages, err := output("go", "list", "./...")
	if err != nil {
		return p, err
	}
	listing, err := output("go", "test", "-json", "-list", ".", "./...")
	if err != nil {
		return p, err
	}
	tests := map[string][]string{}
	listed := regexp.MustCompile(`^(Test|Example|Fuzz)\w*$`)
	for _, line := range strings.Split(listing, "\n") {
		var e event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return p, fmt.Errorf("test listing: %w", err)
		}
		name := strings.TrimSpace(e.Output)
		if listed.MatchString(name) {
			tests[e.Package] = append(tests[e.Package], name)
		}
	}
	for _, pkg := range strings.Split(packages, "\n") {
		for _, test := range tests[pkg] {
			names, err := children(pkg, test)
			if err != nil {
				return p, err
			}
			if len(names) == 0 {
				p.Units = append(p.Units, unit{Package: pkg, Test: test})
			} else {
				for _, name := range names {
					p.Units = append(p.Units, unit{Package: pkg, Test: test + "/" + name})
				}
			}
		}
		if len(tests[pkg]) == 0 {
			p.Units = append(p.Units, unit{Package: pkg})
		}
	}

	loads := make([]float64, count)
	known := []int{}
	seen := map[string]bool{}
	for i := range p.Units {
		u := &p.Units[i]
		if seen[u.key()] {
			return p, fmt.Errorf("duplicate planned unit %s", u.key())
		}
		seen[u.key()] = true
		if seconds, ok := weights[u.key()]; ok {
			if seconds < 0 {
				return p, fmt.Errorf("negative timing %s", u.key())
			}
			u.Seconds = seconds
			known = append(known, i)
		} else {
			hash := sha256.Sum256([]byte(u.key()))
			u.Shard = int(binary.BigEndian.Uint64(hash[:8]) % uint64(count))
		}
	}
	sort.Slice(known, func(i, j int) bool {
		a, b := p.Units[known[i]], p.Units[known[j]]
		if a.Seconds == b.Seconds {
			return a.key() < b.key()
		}
		return a.Seconds > b.Seconds
	})
	for _, i := range known {
		best := 0
		for j := 1; j < count; j++ {
			if loads[j] < loads[best] {
				best = j
			}
		}
		p.Units[i].Shard = best
		loads[best] += p.Units[i].Seconds
	}
	sort.Slice(p.Units, func(i, j int) bool { return p.Units[i].key() < p.Units[j].key() })
	p.Shards = predictions(p, weights)
	p.Digest = planDigest(p)
	return p, nil
}

// Group only siblings under an identical literal prefix. Alternating whole slash-containing
// patterns would form a Cartesian product in Go's -run parser and execute unplanned tests.
func patterns(units []unit) []string {
	groups := map[string][]string{}
	for _, u := range units {
		if u.Test == "" {
			groups[""] = nil
			continue
		}
		parts := strings.Split(u.Test, "/")
		prefix := strings.Join(parts[:len(parts)-1], "/")
		groups[prefix] = append(groups[prefix], parts[len(parts)-1])
	}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result []string
	for _, prefix := range keys {
		names := groups[prefix]
		if len(names) == 0 {
			result = append(result, "^$")
			continue
		}
		sort.Strings(names)
		for i := range names {
			names[i] = regexp.QuoteMeta(names[i])
		}
		last := "^(" + strings.Join(names, "|") + ")$"
		if prefix != "" {
			parts := strings.Split(prefix, "/")
			for i := range parts {
				parts[i] = "^" + regexp.QuoteMeta(parts[i]) + "$"
			}
			last = strings.Join(parts, "/") + "/" + last
		}
		result = append(result, last)
	}
	return result
}
func skipReason(action, text string) string {
	if action != "skip" {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "===") && !strings.HasPrefix(line, "---") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
func readRuns(path string) ([]result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 65536), 64*1024*1024)
	var runs []result
	for scanner.Scan() {
		var e event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, err
		}
		if e.Action == "run" && e.Test != "" {
			runs = append(runs, result{Package: e.Package, Test: e.Test, Action: "run"})
		}
	}
	return runs, scanner.Err()
}
func readLog(path string) ([]result, []string, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 65536), 64*1024*1024)
	outputs := map[string]string{}
	var results []result
	var cache []string
	raw := 0
	for scanner.Scan() {
		var e event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, nil, 0, fmt.Errorf("%s: malformed JSON: %w", path, err)
		}
		key := e.Package + "::" + e.Test
		outputs[key] += e.Output
		for _, line := range strings.Split(e.Output, "\n") {
			if strings.Contains(line, "gate cache:") {
				cache = append(cache, line)
			}
		}
		if e.Action == "pass" || e.Action == "fail" || e.Action == "skip" {
			results = append(results, result{Package: e.Package, Test: e.Test, Action: e.Action, Output: outputs[key], Seconds: e.Elapsed, Reason: skipReason(e.Action, outputs[key])})
			delete(outputs, key)
			if e.Test != "" {
				raw++
			}
		}
	}
	return results, cache, raw, scanner.Err()
}
func disk(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	if uint64(stat.Type) == 0x01021994 {
		return 0, fmt.Errorf("scratch %s is tmpfs", path)
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}
func tmpfs(path string) bool {
	var s syscall.Statfs_t
	return syscall.Statfs(path, &s) == nil && uint64(s.Type) == 0x01021994
}
func free(path string) uint64 {
	var s syscall.Statfs_t
	if syscall.Statfs(path, &s) != nil {
		return 0
	}
	return s.Bavail * uint64(s.Bsize)
}
func loadAverage() string {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(b))
}
func buildFlags(commit, before, after string) string {
	get := func(n string, a ...string) string {
		s, e := output(n, a...)
		if e != nil {
			return "unavailable"
		}
		return strings.Split(s, "\n")[0]
	}
	quota, _ := os.ReadFile("/sys/fs/cgroup/cpu.max")
	return fmt.Sprintf("commit=%s nproc=%s cpu.max=%q go=%q clang=%q node=%q load_before=%q load_after=%q uncached=1 GOFLAGS=%q CGO_ENABLED=%q GOMAXPROCS=%q width_deps=%q", commit, get("nproc"), strings.TrimSpace(string(quota)), get("go", "version"), get("clang", "--version"), get("node", "--version"), before, after, os.Getenv("GOFLAGS"), os.Getenv("CGO_ENABLED"), os.Getenv("GOMAXPROCS"), os.Getenv("ADAMIC_MARKDOWNWIDTH_DEPS"))
}
func shard(index, count int, out, scratch string, resume bool) error {
	started := time.Now()
	beforeLoad := loadAverage()
	if out == "" || index < 0 || index >= count {
		return errors.New("shard requires -out and 0 <= index < count")
	}
	root, err := filepath.Abs(scratch)
	if err != nil {
		return err
	}
	repo, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	if root == repo || strings.HasPrefix(root, repo+string(os.PathSeparator)) {
		return errors.New("scratch must be outside repository corpus")
	}
	if err := os.MkdirAll(root, 0777); err != nil {
		return err
	}
	if err := os.Chmod(root, 0777|os.ModeSticky); err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if resolved == repo || strings.HasPrefix(resolved, repo+string(os.PathSeparator)) {
		return errors.New("resolved scratch is inside repository")
	}
	beforeDisk, err := disk(root)
	if err != nil {
		return err
	}
	if err := os.Setenv("TMPDIR", root); err != nil {
		return err
	}
	p, err := makePlan(count)
	if err != nil {
		return err
	}

	context, err := executionIdentity()
	if err != nil {
		return err
	}
	state := resumeState{PlanDigest: p.Digest, Context: context, Index: index}
	if _, err := os.Stat(out); err == nil {
		if !resume {
			return errors.New("output directory already exists; use -resume for identical inputs")
		}
		var saved resumeState
		if err := loadJSON(filepath.Join(out, "run.json"), &saved); err != nil {
			return err
		}
		if saved != state {
			return errors.New("resume inputs differ: plan, shard, environment, tools, or external inputs changed")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(out, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return errors.New("another runner owns this output directory")
	}
	if err := atomicJSON(filepath.Join(out, "run.json"), state); err != nil {
		return err
	}
	s := summary{Version: 1, Plan: p, Index: index, Checks: map[string]string{}, DiskBefore: beforeDisk}
	raw, err := os.Create(filepath.Join(out, "test.jsonl"))
	if err != nil {
		return err
	}
	stderr, err := os.Create(filepath.Join(out, "test.stderr"))
	if err != nil {
		raw.Close()
		return err
	}
	var commandStderr io.Writer = stderr
	runCommand := func(name string, args []string, tmp string, w io.Writer) int {
		cmd := exec.Command(name, args...)
		cmd.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1", "TMPDIR="+tmp)
		cmd.Stdout = w
		cmd.Stderr = commandStderr
		t := time.Now()
		err := cmd.Run()
		code := 0
		if err != nil {
			code = 1
			if exit, ok := err.(*exec.ExitError); ok {
				code = exit.ExitCode()
			}
			fmt.Fprintf(commandStderr, "%s %v: %v\n", name, args, err)
		}
		if name == "go" && len(args) > 0 && args[0] == "test" {
			uncached := false
			for _, v := range cmd.Env {
				if strings.HasPrefix(v, "ADAMIC_GATE_UNCACHED=") {
					uncached = v == "ADAMIC_GATE_UNCACHED=1"
				}
			}
			s.Invocations = append(s.Invocations, invocation{args[len(args)-1], args, uncached, code, time.Since(t).Seconds()})
		}
		return code
	}
	if index == 0 {
		temp, err := os.MkdirTemp(root, "checks-")
		if err != nil {
			return err
		}
		os.Chmod(temp, 0755)
		formatPath := filepath.Join(out, "gofmt.log")
		f, err := os.Create(formatPath)
		if err != nil {
			return err
		}
		code := runCommand("gofmt", []string{"-l", "cmd", "internal"}, temp, f)
		f.Close()
		b, _ := os.ReadFile(formatPath)
		if code != 0 || len(b) > 0 {
			s.Checks["gofmt"] = "fail"
		} else {
			s.Checks["gofmt"] = "pass"
		}
		vet, err := os.Create(filepath.Join(out, "vet.log"))
		if err != nil {
			return err
		}
		if runCommand("go", []string{"vet", "./..."}, temp, vet) != 0 {
			s.Checks["vet"] = "fail"
		} else {
			s.Checks["vet"] = "pass"
		}
		vet.Close()
		if err := os.RemoveAll(temp); err != nil {
			s.Errors = append(s.Errors, err.Error())
		}
	}
	byPackage := map[string][]unit{}
	for _, u := range p.Units {
		if u.Shard == index {
			byPackage[u.Package] = append(byPackage[u.Package], u)
		}
	}
	pkgs := []string{}
	for pkg := range byPackage {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	fmt.Printf("shard %d disk before=%d bytes scratch=%s /tmp_free=%d bytes /tmp_tmpfs=%t\n", index, beforeDisk, root, free("/tmp"), tmpfs("/tmp"))
	for _, pkg := range pkgs {
		packageDir := filepath.Join(out, "packages", shortHash(pkg))
		key := checkpointKey(state, pkg, patterns(byPackage[pkg]))
		if err := cleanupPackageScratch(packageDir, key, root); err != nil {
			return err
		}
		if resume {
			saved, found, err := loadPackage(packageDir, key, p, index, pkg)
			if err != nil {
				return err
			}
			if found {
				if err := appendFile(raw, filepath.Join(packageDir, "test.jsonl")); err != nil {
					return err
				}
				if err := appendFile(stderr, filepath.Join(packageDir, "test.stderr")); err != nil {
					return err
				}
				s.Invocations = append(s.Invocations, saved.Invocations...)
				fmt.Printf("resumed %s from complete log\n", pkg)
				continue
			}
		}
		if err := os.MkdirAll(packageDir, 0755); err != nil {
			return err
		}
		packageLog, err := os.Create(filepath.Join(packageDir, "test.jsonl.partial"))
		if err != nil {
			return err
		}
		packageStderr, err := os.Create(filepath.Join(packageDir, "test.stderr.partial"))
		if err != nil {
			packageLog.Close()
			return err
		}
		tmp, err := os.MkdirTemp(root, "package-")
		if err != nil {
			return err
		}
		if err := os.Chmod(tmp, 0755); err != nil {
			return err
		}
		if err := atomicJSON(filepath.Join(packageDir, "scratch.json"), scratchRecord{Key: key, Path: tmp}); err != nil {
			return err
		}
		first := len(s.Invocations)
		commandStderr = io.MultiWriter(stderr, packageStderr)
		for _, pattern := range patterns(byPackage[pkg]) {
			runCommand("go", testArgs(pkg, pattern), tmp, io.MultiWriter(raw, packageLog))
		}
		commandStderr = stderr
		err = errors.Join(packageLog.Sync(), packageStderr.Sync(), packageLog.Close(), packageStderr.Close())
		if err != nil {
			return err
		}
		if err := os.Rename(filepath.Join(packageDir, "test.jsonl.partial"), filepath.Join(packageDir, "test.jsonl")); err != nil {
			return err
		}
		if err := os.Rename(filepath.Join(packageDir, "test.stderr.partial"), filepath.Join(packageDir, "test.stderr")); err != nil {
			return err
		}
		checkpoint := packageEvidence{Key: key, Package: pkg, Invocations: s.Invocations[first:]}
		if err := completePackage(packageDir, checkpoint, p, index, pkg); err != nil {
			s.Errors = append(s.Errors, err.Error())
		} else {
			checkpoint.LogDigest, err = fileDigest(filepath.Join(packageDir, "test.jsonl"))
			if err != nil {
				return err
			}
			checkpoint.StderrDigest, err = fileDigest(filepath.Join(packageDir, "test.stderr"))
			if err != nil {
				return err
			}
			if err := atomicJSON(filepath.Join(packageDir, "complete.json"), checkpoint); err != nil {
				return err
			}
		}
		if err := cleanupPackageScratch(packageDir, key, root); err != nil {
			s.Errors = append(s.Errors, err.Error())
		}
		fmt.Printf("finished %s free=%d bytes\n", pkg, free(root))
	}
	raw.Close()
	stderr.Close()
	afterSource, identityErr := sourceIdentity()
	afterCommit, commitErr := output("git", "rev-parse", "HEAD")
	status, statusErr := output("git", "status", "--porcelain", "--untracked-files=normal")
	if identityErr != nil || commitErr != nil || statusErr != nil || afterSource != p.Source || afterCommit != p.Commit || status != "" {
		s.Errors = append(s.Errors, "repository changed during shard execution")
	}

	s.Results, s.CacheLines, _, err = readLog(filepath.Join(out, "test.jsonl"))
	if err != nil {
		s.Errors = append(s.Errors, err.Error())
	}
	s.WallSeconds = time.Since(started).Seconds()
	s.DiskAfter = free(root)
	s.BuildFlags = buildFlags(p.Commit, beforeLoad, loadAverage())
	s.Pass, s.Fail, s.Skip = totals(canonical(s.Results))
	if err := saveJSON(filepath.Join(out, "summary.json"), s); err != nil {
		return err
	}
	cache, err := os.Create(filepath.Join(out, "gate-cache.log"))
	if err != nil {
		return err
	}
	for _, line := range s.CacheLines {
		fmt.Fprintln(cache, line)
	}
	cache.Close()
	fmt.Printf("shard %d pass=%d fail=%d skip=%d wall=%.3fs disk after=%d bytes\n%s\n", index, s.Pass, s.Fail, s.Skip, s.WallSeconds, s.DiskAfter, s.BuildFlags)
	for _, call := range s.Invocations {
		if call.Exit != 0 {
			return errors.New("shard test command failed; evidence retained")
		}
	}
	for _, check := range s.Checks {
		if check != "pass" {
			return errors.New("shard gate check failed")
		}
	}
	if len(s.Errors) > 0 {
		return errors.New(strings.Join(s.Errors, "; "))
	}
	return nil
}
func canonical(results []result) []result {
	byName := map[string]result{}
	rank := map[string]int{"skip": 1, "pass": 2, "fail": 3}
	for _, r := range results {
		old, ok := byName[r.key()]
		if !ok || rank[r.Action] > rank[old.Action] {
			byName[r.key()] = r
		}
	}
	keys := []string{}
	for k := range byName {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []result
	for _, k := range keys {
		out = append(out, byName[k])
	}
	return out
}
func totals(results []result) (pass, fail, skip int) {
	for _, r := range results {
		if r.Test == "" {
			continue
		}
		switch r.Action {
		case "pass":
			pass++
		case "fail":
			fail++
		case "skip":
			skip++
		}
	}
	return
}
func samePlan(a, b plan) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

// An enumerated child's ancestors necessarily execute in Go. Only these declared ancestors can
// repeat; all work units and all other tests have exactly one owner and one terminal event.
func owns(u unit, r result) bool {
	return u.Package == r.Package && (r.Test == u.Test || (u.Test != "" && strings.HasPrefix(r.Test, u.Test+"/")))
}
func ancestor(u unit, r result) bool {
	return u.Package == r.Package && r.Test != "" && strings.HasPrefix(u.Test, r.Test+"/")
}

var cacheHits = regexp.MustCompile(`gate cache: (native|node|probe) hits=(\d+) misses=(\d+)`)

func validateResults(expected plan, index int, results []result, produced map[string]bool, seenTest map[string]int) []string {
	var problems []string
	for _, r := range results {
		if r.Test == "" {
			if r.Action == "fail" {
				problems = append(problems, "package failed: "+r.Package)
			}
			for _, u := range expected.Units {
				if u.Test == "" && u.Package == r.Package && u.Shard == index {
					produced[u.key()] = true
				}
			}
			continue
		}
		owner := -1
		isAncestor := false
		for j, u := range expected.Units {
			if owns(u, r) {
				owner = j
			}
			if ancestor(u, r) && u.Shard == index {
				isAncestor = true
			}
		}
		if owner >= 0 {
			u := expected.Units[owner]
			if u.Shard != index {
				problems = append(problems, fmt.Sprintf("test %s ran in shard %d, owner %d", r.key(), index, u.Shard))
			}
			if previous, ok := seenTest[r.key()]; ok {
				problems = append(problems, fmt.Sprintf("test %s ran twice in shards %d and %d", r.key(), previous, index))
			}
			seenTest[r.key()] = index
			if r.Test == u.Test {
				produced[u.key()] = true
			}
		} else if !isAncestor {
			problems = append(problems, "unplanned test: "+r.key())
		}
	}
	return problems
}

func merge(dirs []string, out string) error {
	if len(dirs) == 0 {
		return errors.New("merge requires shard directories")
	}
	var summaries []summary
	for _, dir := range dirs {
		var s summary
		if err := loadJSON(filepath.Join(dir, "summary.json"), &s); err != nil {
			return err
		}
		summaries = append(summaries, s)
	}
	expected, err := makePlan(summaries[0].Plan.Count)
	if err != nil {
		return err
	}
	m := merged{Version: 1, Plan: expected}
	seenShard := map[int]bool{}
	seenTest := map[string]int{}
	seenRuns := map[string]int{}
	produced := map[string]bool{}
	var all []result
	for i, s := range summaries {
		if s.Version != 1 || !samePlan(expected, s.Plan) {
			m.Errors = append(m.Errors, fmt.Sprintf("shard %d plan differs from repository enumeration", s.Index))
		}
		if s.Index < 0 || s.Index >= expected.Count || seenShard[s.Index] {
			m.Errors = append(m.Errors, fmt.Sprintf("invalid or duplicate shard %d", s.Index))
		}
		seenShard[s.Index] = true
		if s.Index == 0 {
			for _, check := range []string{"gofmt", "vet"} {
				if s.Checks[check] != "pass" {
					m.Errors = append(m.Errors, check+" did not pass")
				}
			}
		}
		for _, call := range s.Invocations {
			if !call.Uncached {
				m.Errors = append(m.Errors, fmt.Sprintf("shard %d invocation was not uncached", s.Index))
			}
			if call.Exit != 0 {
				m.Errors = append(m.Errors, fmt.Sprintf("shard %d command exit %d", s.Index, call.Exit))
			}
		}
		hasUnits := false
		for _, u := range expected.Units {
			if u.Shard == s.Index {
				hasUnits = true
			}
		}
		if hasUnits && len(s.Invocations) == 0 {
			m.Errors = append(m.Errors, fmt.Sprintf("shard %d has no invocation evidence", s.Index))
		}
		m.Errors = append(m.Errors, s.Errors...)
		m.Errors = append(m.Errors, validatePackageFiles(dirs[i], expected, s)...)
		results, lines, raw, err := readLog(filepath.Join(dirs[i], "test.jsonl"))
		if err != nil {
			return err
		}
		m.RawTerminalEvents += raw
		for _, r := range results {
			if r.Test != "" && r.Action == "fail" {
				m.Errors = append(m.Errors, fmt.Sprintf("shard %d test %s failed\n%s", s.Index, r.key(), r.Output))
			}
		}

		declared := map[string]int{}
		observed := map[string]int{}
		for _, call := range s.Invocations {
			declared[call.Package]++
			assigned := false
			for _, u := range expected.Units {
				if u.Package == call.Package && u.Shard == s.Index {
					assigned = true
				}
			}
			if !assigned {
				m.Errors = append(m.Errors, "invocation for unassigned package: "+call.Package)
			}
		}
		for _, r := range results {
			if r.Test == "" {
				observed[r.Package]++
			}
		}
		for pkg, n := range declared {
			if observed[pkg] != n {
				m.Errors = append(m.Errors, "invocation/terminal count mismatch: "+pkg)
			}
		}
		for pkg, n := range observed {
			if declared[pkg] != n {
				m.Errors = append(m.Errors, "unrecorded package invocation: "+pkg)
			}
		}
		for _, line := range lines {
			match := cacheHits.FindStringSubmatch(line)
			if match == nil || match[2] != "0" {
				m.Errors = append(m.Errors, fmt.Sprintf("shard %d refuses cache line: %s", s.Index, line))
			}
		}
		m.Errors = append(m.Errors, validateResults(expected, s.Index, results, produced, seenTest)...)
		runs, err := readRuns(filepath.Join(dirs[i], "test.jsonl"))
		if err != nil {
			return err
		}
		m.Errors = append(m.Errors, validateResults(expected, s.Index, runs, map[string]bool{}, seenRuns)...)
		ran := map[string]bool{}
		for _, r := range runs {
			ran[r.key()] = true
		}
		for _, r := range results {
			if r.Test != "" && !ran[r.key()] {
				m.Errors = append(m.Errors, "terminal event without run: "+r.key())
			}
		}
		finished := map[string]bool{}
		for _, r := range results {
			if r.Test != "" {
				finished[r.key()] = true
			}
		}
		for _, r := range runs {
			if !finished[r.key()] {
				m.Errors = append(m.Errors, "run without terminal event: "+r.key())
			}
		}

		all = append(all, results...)
		m.WallSeconds = append(m.WallSeconds, s.WallSeconds)
		m.BuildFlags = append(m.BuildFlags, s.BuildFlags)
	}
	for i := 0; i < expected.Count; i++ {
		if !seenShard[i] {
			m.Errors = append(m.Errors, fmt.Sprintf("missing shard %d", i))
		}
	}
	for _, u := range expected.Units {
		if !produced[u.key()] {
			m.Errors = append(m.Errors, "planned unit produced no terminal test event: "+u.key())
		}
	}
	m.Results = canonical(all)
	m.Pass, m.Fail, m.Skip = totals(m.Results)
	m.TestEvents = m.Pass + m.Fail + m.Skip
	m.Green = len(m.Errors) == 0 && m.Fail == 0
	if _, err := os.Stat(out); err == nil {
		return errors.New("merged output already exists")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	if err := saveJSON(filepath.Join(out, "merged.json"), m); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(out, "test.jsonl"))
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		input, err := os.Open(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			f.Close()
			return err
		}
		_, err = io.Copy(f, input)
		input.Close()
		if err != nil {
			f.Close()
			return err
		}
	}
	f.Close()
	verdict := "RED"
	if m.Green {
		verdict = "GREEN"
	}
	fmt.Printf("%s pass=%d fail=%d skip=%d test_events=%d raw_terminal_events=%d\n", verdict, m.Pass, m.Fail, m.Skip, m.TestEvents, m.RawTerminalEvents)
	for _, problem := range m.Errors {
		fmt.Fprintln(os.Stderr, problem)
	}
	if !m.Green {
		return errors.New("merged gate is not green")
	}
	return nil
}
func compare(path, log string) error {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "merged.json")
	}
	var m merged
	if err := loadJSON(path, &m); err != nil {
		return err
	}
	whole, _, raw, err := readLog(log)
	if err != nil {
		return err
	}
	whole = canonical(whole)
	for _, r := range whole {
		if r.Test == "" && r.Action == "fail" {
			return fmt.Errorf("unsharded package failed: %s", r.Package)
		}
	}
	a, b := map[string]string{}, map[string]string{}
	for _, r := range m.Results {
		if r.Test != "" {
			a[r.key()] = r.Action
		}
	}
	for _, r := range whole {
		if r.Test != "" {
			b[r.key()] = r.Action
		}
	}
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	var diff []string
	for k := range keys {
		if a[k] != b[k] {
			diff = append(diff, fmt.Sprintf("%s merged=%q whole=%q", k, a[k], b[k]))
		}
	}
	sort.Strings(diff)
	pass, fail, skip := totals(whole)
	fmt.Printf("merged pass=%d fail=%d skip=%d events=%d\nwhole pass=%d fail=%d skip=%d events=%d raw_terminal_events=%d\n", m.Pass, m.Fail, m.Skip, len(a), pass, fail, skip, len(b), raw)
	for _, line := range diff {
		fmt.Println(line)
	}
	if !m.Green || len(diff) > 0 || len(a) != len(b) || raw != len(b) {
		return errors.New("compare failed")
	}
	fmt.Println("diff: empty")
	return nil
}
