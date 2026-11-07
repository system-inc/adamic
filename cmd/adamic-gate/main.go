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

func (unit unit) key() string { return unit.Package + "::" + unit.Test }

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

func (result result) key() string { return result.Package + "::" + result.Test }

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
	if runError := run(os.Args[1:]); runError != nil {
		fmt.Fprintln(os.Stderr, "adamic-gate:", runError)
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
		if parseError := flags.Parse(args[1:]); parseError != nil {
			return parseError
		}
		plan, makePlanError := makePlan(*count)
		if makePlanError != nil {
			return makePlanError
		}
		return writeJSON(os.Stdout, plan)
	case "shard":
		flags := flag.NewFlagSet("shard", flag.ContinueOnError)
		count := flags.Int("count", 8, "number of shards")
		index := flags.Int("index", -1, "zero-based shard")
		out := flags.String("out", "", "evidence directory")
		resume := flags.Bool("resume", false, "reuse complete package evidence for identical execution inputs")
		scratch := flags.String(
			"scratch",
			"/workspace/adamic-gate-scratch",
			"disk-backed scratch root outside the repository",
		)
		if parseError := flags.Parse(args[1:]); parseError != nil {
			return parseError
		}
		return shard(*index, *count, *out, *scratch, *resume)
	case "merge":
		flags := flag.NewFlagSet("merge", flag.ContinueOnError)
		out := flags.String("out", "merged", "merged evidence directory")
		if parseError := flags.Parse(args[1:]); parseError != nil {
			return parseError
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
		if parseError := flags.Parse(args[1:]); parseError != nil {
			return parseError
		}
		if flags.NArg() != 1 {
			return errors.New("timings [-out file] <go test -json log>")
		}
		results, _, _, readLogError := readLog(flags.Arg(0))
		if readLogError != nil {
			return readLogError
		}
		weights := map[string]float64{}
		for _, result := range results {
			weights[result.key()] = result.Seconds
		}
		return saveJSON(*out, weights)
	}
	return fmt.Errorf("unknown command %q", args[0])
}
func output(name string, args ...string) (string, error) {
	b, combinedOutputError := exec.Command(name, args...).CombinedOutput()
	if combinedOutputError != nil {
		return "", fmt.Errorf("%s %v: %w\n%s", name, args, combinedOutputError, b)
	}
	return strings.TrimSpace(string(b)), nil
}
func writeJSON(writer io.Writer, v any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
func saveJSON(path string, v any) error {
	file, createError := os.Create(path)
	if createError != nil {
		return createError
	}
	writeJsonError := writeJSON(file, v)
	closeError := file.Close()
	return errors.Join(writeJsonError, closeError)
}
func loadJSON(path string, v any) error {
	content, readError := os.ReadFile(path)
	if readError != nil {
		return readError
	}
	return json.Unmarshal(content, v)
}

// Read literal fixture rows from the test's AST, then verify each named input exists. A directory
// glob alone would accidentally include helpers and probes the parent does not execute.
func fixtureRows(file, variable string) ([]string, error) {
	tree, parseFileError := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if parseFileError != nil {
		return nil, parseFileError
	}
	var names []string
	ast.Inspect(tree, func(node ast.Node) bool {
		declaration, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for index, name := range declaration.Names {
			if name.Name != variable {
				continue
			}
			if index >= len(declaration.Values) {
				continue
			}
			literal, ok := declaration.Values[index].(*ast.CompositeLit)
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
				text, unquoteError := strconv.Unquote(value.Value)
				if unquoteError != nil {
					continue
				}
				names = append(names, text)
			}
		}
		return true
	})
	if len(names) == 0 {
		return nil, fmt.Errorf("no fixture rows for %s in %s", variable, file)
	}
	for _, name := range names {
		if _, statError := os.Stat(name); statError != nil {
			return nil, statError
		}
	}
	return names, nil
}

// Literal first fields of the anonymous slice iterated directly by a test. This is deliberately
// restricted to audited independent parents; arbitrary AST inference cannot prove selectability.
func literalChildren(file, parent string, table ...string) ([]string, error) {
	tree, parseFileError := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if parseFileError != nil {
		return nil, parseFileError
	}
	var names []string
	for _, d := range tree.Decls {
		funcDecl, ok := d.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != parent {
			continue
		}
		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			stmt, ok := node.(*ast.RangeStmt)
			if !ok {
				return true
			}
			literal, ok := stmt.X.(*ast.CompositeLit)
			if !ok && len(table) == 1 {
				if id, yes := stmt.X.(*ast.Ident); yes && id.Name == table[0] {
					ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
						assignment, yes := node.(*ast.AssignStmt)
						if !yes {
							return true
						}
						for index, lhs := range assignment.Lhs {
							id, yes := lhs.(*ast.Ident)
							if yes && id.Name == table[0] && index < len(assignment.Rhs) {
								literal, ok = assignment.Rhs[index].(*ast.CompositeLit)
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
			for _, elt := range literal.Elts {
				if label, ok := elt.(*ast.BasicLit); ok && label.Kind == token.STRING {
					name, unquoteError := strconv.Unquote(label.Value)
					if unquoteError == nil {
						names = append(names, strings.ReplaceAll(name, " ", "_"))
					}
					continue
				}
				row, ok := elt.(*ast.CompositeLit)
				if !ok || len(row.Elts) == 0 {
					continue
				}
				lit, ok := row.Elts[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				name, unquoteError := strconv.Unquote(lit.Value)
				if unquoteError == nil {
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
func children(packagePath, parent string) ([]string, error) {
	if strings.HasSuffix(packagePath, "/stage1/cohere/typeaware") && parent == "TestVolumeAgreementAndMutants" {
		return literalChildren("stage1/cohere/typeaware/volume_test.go", parent, "changes")
	}
	if strings.HasSuffix(packagePath, "/internal/oracle") {
		switch parent {
		case "TestInputAgreesWithNode":
			return fixtureRows("internal/oracle/input_test.go", "inputFixtures")
		case "TestFreshWriteProbesStayRefused":
			paths, globError := filepath.Glob("internal/oracle/testdata/fresh_refused/*.a")
			for index := range paths {
				paths[index] = filepath.Base(paths[index])
			}
			if len(paths) < 20 {
				return nil, errors.New("fresh probes disappeared")
			}
			return paths, globError
		}
	}
	if strings.HasSuffix(packagePath, "/stage1/cohere/css") && parent == "TestCSSPrinterAgreesWithGo" {
		return literalChildren("stage1/cohere/css/print_test.go", parent)
	}
	if strings.HasSuffix(packagePath, "/stage1/cohere/lint") {
		switch parent {
		case "TestMutants":
			return literalChildren("stage1/cohere/lint/lint_test.go", parent)
		case "TestVolumeMutants":
			return literalChildren("stage1/cohere/lint/volume_test.go", parent)
		}
	}
	if strings.HasSuffix(packagePath, "/internal/native") {
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
	tracked, outputError := output("git", "ls-files", "-s", "-z")
	if outputError != nil {
		return "", outputError
	}
	hash := sha256.New()
	for _, line := range strings.Split(tracked, "\x00") {
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) != 2 {
			continue
		}
		name := fields[1]
		fmt.Fprintf(hash, "%d:%s", len(line), line)
		if strings.HasPrefix(line, "160000 ") {
			sha, outputError := output("git", "-C", name, "rev-parse", "HEAD")
			if outputError != nil {
				return "", outputError
			}
			fmt.Fprint(hash, sha)
			continue
		}
		var b []byte
		if strings.HasPrefix(line, "120000 ") {
			target, readLinkError := os.Readlink(name)
			if readLinkError != nil {
				return "", readLinkError
			}
			b = []byte(target)
		} else {
			var readError error
			b, readError = os.ReadFile(name)
			if readError != nil {
				return "", readError
			}
		}
		fmt.Fprintf(hash, "%d:", len(b))
		hash.Write(b)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func makePlan(count int) (plan, error) {
	plan := plan{Version: 1, Count: count}
	status, outputError := output("git", "status", "--porcelain", "--untracked-files=normal")
	if outputError != nil {
		return plan, outputError
	}
	if status != "" {
		return plan, errors.New(
			"planning requires a clean checkout, including submodules; keep evidence outside the repository",
		)
	}
	if count < 1 {
		return plan, errors.New("count must be positive")
	}
	plan.Commit, outputError = output("git", "rev-parse", "HEAD")
	if outputError != nil {
		return plan, outputError
	}
	var sourceIdentityError error
	plan.Source, sourceIdentityError = sourceIdentity()
	if sourceIdentityError != nil {
		return plan, sourceIdentityError
	}
	var weights map[string]float64
	if loadJsonError := loadJSON(timingPath, &weights); loadJsonError != nil {
		return plan, loadJsonError
	}
	packages, outputError := output("go", "list", "./...")
	if outputError != nil {
		return plan, outputError
	}
	listing, outputError := output("go", "test", "-json", "-list", ".", "./...")
	if outputError != nil {
		return plan, outputError
	}
	tests := map[string][]string{}
	listed := regexp.MustCompile(`^(Test|Example|Fuzz)\w*$`)
	for _, line := range strings.Split(listing, "\n") {
		var event event
		if decodeError := json.Unmarshal([]byte(line), &event); decodeError != nil {
			return plan, fmt.Errorf("test listing: %w", decodeError)
		}
		name := strings.TrimSpace(event.Output)
		if listed.MatchString(name) {
			tests[event.Package] = append(tests[event.Package], name)
		}
	}
	for _, packagePath := range strings.Split(packages, "\n") {
		for _, testCase := range tests[packagePath] {
			names, childrenError := children(packagePath, testCase)
			if childrenError != nil {
				return plan, childrenError
			}
			if len(names) == 0 {
				plan.Units = append(plan.Units, unit{Package: packagePath, Test: testCase})
			} else {
				for _, name := range names {
					plan.Units = append(plan.Units, unit{Package: packagePath, Test: testCase + "/" + name})
				}
			}
		}
		if len(tests[packagePath]) == 0 {
			plan.Units = append(plan.Units, unit{Package: packagePath})
		}
	}

	loads := make([]float64, count)
	known := []int{}
	seen := map[string]bool{}
	for index := range plan.Units {
		unit := &plan.Units[index]
		if seen[unit.key()] {
			return plan, fmt.Errorf("duplicate planned unit %s", unit.key())
		}
		seen[unit.key()] = true
		if seconds, ok := weights[unit.key()]; ok {
			if seconds < 0 {
				return plan, fmt.Errorf("negative timing %s", unit.key())
			}
			unit.Seconds = seconds
			known = append(known, index)
		} else {
			hash := sha256.Sum256([]byte(unit.key()))
			unit.Shard = int(binary.BigEndian.Uint64(hash[:8]) % uint64(count))
		}
	}
	sort.Slice(known, func(left, right int) bool {
		a, b := plan.Units[known[left]], plan.Units[known[right]]
		if a.Seconds == b.Seconds {
			return a.key() < b.key()
		}
		return a.Seconds > b.Seconds
	})
	for _, index := range known {
		best := 0
		for index := 1; index < count; index++ {
			if loads[index] < loads[best] {
				best = index
			}
		}
		plan.Units[index].Shard = best
		loads[best] += plan.Units[index].Seconds
	}
	sort.Slice(plan.Units, func(left, right int) bool { return plan.Units[left].key() < plan.Units[right].key() })
	plan.Shards = predictions(plan, weights)
	plan.Digest = planDigest(plan)
	return plan, nil
}

// Group only siblings under an identical literal prefix. Alternating whole slash-containing
// patterns would form a Cartesian product in Go's -run parser and execute unplanned tests.
func patterns(units []unit) []string {
	groups := map[string][]string{}
	for _, unit := range units {
		if unit.Test == "" {
			groups[""] = nil
			continue
		}
		parts := strings.Split(unit.Test, "/")
		prefix := strings.Join(parts[:len(parts)-1], "/")
		groups[prefix] = append(groups[prefix], parts[len(parts)-1])
	}
	keys := []string{}
	for key := range groups {
		keys = append(keys, key)
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
		for index := range names {
			names[index] = regexp.QuoteMeta(names[index])
		}
		last := "^(" + strings.Join(names, "|") + ")$"
		if prefix != "" {
			parts := strings.Split(prefix, "/")
			for index := range parts {
				parts[index] = "^" + regexp.QuoteMeta(parts[index]) + "$"
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
	file, openError := os.Open(path)
	if openError != nil {
		return nil, openError
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 64*1024*1024)
	var runs []result
	for scanner.Scan() {
		var event event
		if decodeError := json.Unmarshal(scanner.Bytes(), &event); decodeError != nil {
			return nil, decodeError
		}
		if event.Action == "run" && event.Test != "" {
			runs = append(runs, result{Package: event.Package, Test: event.Test, Action: "run"})
		}
	}
	return runs, scanner.Err()
}
func readLog(path string) ([]result, []string, int, error) {
	file, openError := os.Open(path)
	if openError != nil {
		return nil, nil, 0, openError
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 64*1024*1024)
	outputs := map[string]string{}
	var results []result
	var cache []string
	raw := 0
	for scanner.Scan() {
		var event event
		if decodeError := json.Unmarshal(scanner.Bytes(), &event); decodeError != nil {
			return nil, nil, 0, fmt.Errorf("%s: malformed JSON: %w", path, decodeError)
		}
		key := event.Package + "::" + event.Test
		outputs[key] += event.Output
		for _, line := range strings.Split(event.Output, "\n") {
			if strings.Contains(line, "gate cache:") {
				cache = append(cache, line)
			}
		}
		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			results = append(results, result{
				Package: event.Package,
				Test:    event.Test,
				Action:  event.Action,
				Output:  outputs[key],
				Seconds: event.Elapsed,
				Reason:  skipReason(event.Action, outputs[key]),
			})
			delete(outputs, key)
			if event.Test != "" {
				raw++
			}
		}
	}
	return results, cache, raw, scanner.Err()
}
func disk(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if statfsError := syscall.Statfs(path, &stat); statfsError != nil {
		return 0, statfsError
	}
	if uint64(stat.Type) == 0x01021994 {
		return 0, fmt.Errorf("scratch %s is tmpfs", path)
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}
func tmpfs(path string) bool {
	var statfs_t syscall.Statfs_t
	return syscall.Statfs(path, &statfs_t) == nil && uint64(statfs_t.Type) == 0x01021994
}
func free(path string) uint64 {
	var statfs_t syscall.Statfs_t
	if syscall.Statfs(path, &statfs_t) != nil {
		return 0
	}
	return statfs_t.Bavail * uint64(statfs_t.Bsize)
}
func loadAverage() string {
	content, readError := os.ReadFile("/proc/loadavg")
	if readError != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(content))
}
func buildFlags(commit, before, after string) string {
	get := func(n string, a ...string) string {
		text, outputError := output(n, a...)
		if outputError != nil {
			return "unavailable"
		}
		return strings.Split(text, "\n")[0]
	}
	quota, _ := os.ReadFile("/sys/fs/cgroup/cpu.max")
	return fmt.Sprintf(
		"commit=%s nproc=%s cpu.max=%q go=%q clang=%q node=%q load_before=%q load_after=%q uncached=1 GOFLAGS=%q CGO_ENABLED=%q GOMAXPROCS=%q width_deps=%q",
		commit,
		get("nproc"),
		strings.TrimSpace(string(quota)),
		get("go", "version"),
		get("clang", "--version"),
		get("node", "--version"),
		before,
		after,
		os.Getenv("GOFLAGS"),
		os.Getenv("CGO_ENABLED"),
		os.Getenv("GOMAXPROCS"),
		os.Getenv("ADAMIC_MARKDOWNWIDTH_DEPS"),
	)
}
func shard(index, count int, out, scratch string, resume bool) error {
	started := time.Now()
	beforeLoad := loadAverage()
	if out == "" || index < 0 || index >= count {
		return errors.New("shard requires -out and 0 <= index < count")
	}
	root, absolutePathError := filepath.Abs(scratch)
	if absolutePathError != nil {
		return absolutePathError
	}
	repo, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		return absolutePathError
	}
	if root == repo || strings.HasPrefix(root, repo+string(os.PathSeparator)) {
		return errors.New("scratch must be outside repository corpus")
	}
	if createDirectoryError := os.MkdirAll(root, 0777); createDirectoryError != nil {
		return createDirectoryError
	}
	if changeModeError := os.Chmod(root, 0777|os.ModeSticky); changeModeError != nil {
		return changeModeError
	}
	resolved, resolveSymlinksError := filepath.EvalSymlinks(root)
	if resolveSymlinksError != nil {
		return resolveSymlinksError
	}
	if resolved == repo || strings.HasPrefix(resolved, repo+string(os.PathSeparator)) {
		return errors.New("resolved scratch is inside repository")
	}
	beforeDisk, diskError := disk(root)
	if diskError != nil {
		return diskError
	}
	if setenvError := os.Setenv("TMPDIR", root); setenvError != nil {
		return setenvError
	}
	plan, makePlanError := makePlan(count)
	if makePlanError != nil {
		return makePlanError
	}

	context, executionIdentityError := executionIdentity()
	if executionIdentityError != nil {
		return executionIdentityError
	}
	state := resumeState{PlanDigest: plan.Digest, Context: context, Index: index}
	if _, statError := os.Stat(out); statError == nil {
		if !resume {
			return errors.New("output directory already exists; use -resume for identical inputs")
		}
		var saved resumeState
		if loadJsonError := loadJSON(filepath.Join(out, "run.json"), &saved); loadJsonError != nil {
			return loadJsonError
		}
		if saved != state {
			return errors.New("resume inputs differ: plan, shard, environment, tools, or external inputs changed")
		}
	} else if !os.IsNotExist(statError) {
		return statError
	}
	if createDirectoryError := os.MkdirAll(out, 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	lock, openError := os.OpenFile(filepath.Join(out, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if openError != nil {
		return openError
	}
	defer lock.Close()
	if flockError := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); flockError != nil {
		return errors.New("another runner owns this output directory")
	}
	if atomicJsonError := atomicJSON(filepath.Join(out, "run.json"), state); atomicJsonError != nil {
		return atomicJsonError
	}
	summary := summary{Version: 1, Plan: plan, Index: index, Checks: map[string]string{}, DiskBefore: beforeDisk}
	raw, createError := os.Create(filepath.Join(out, "test.jsonl"))
	if createError != nil {
		return createError
	}
	stderr, createError := os.Create(filepath.Join(out, "test.stderr"))
	if createError != nil {
		raw.Close()
		return createError
	}
	var commandStderr io.Writer = stderr
	runCommand := func(name string, args []string, temporary string, writer io.Writer) int {
		command := exec.Command(name, args...)
		command.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1", "TMPDIR="+temporary)
		command.Stdout = writer
		command.Stderr = commandStderr
		t := time.Now()
		runError := command.Run()
		code := 0
		if runError != nil {
			code = 1
			if exit, ok := runError.(*exec.ExitError); ok {
				code = exit.ExitCode()
			}
			fmt.Fprintf(commandStderr, "%s %v: %v\n", name, args, runError)
		}
		if name == "go" && len(args) > 0 && args[0] == "test" {
			uncached := false
			for _, value := range command.Env {
				if strings.HasPrefix(value, "ADAMIC_GATE_UNCACHED=") {
					uncached = value == "ADAMIC_GATE_UNCACHED=1"
				}
			}
			summary.Invocations = append(
				summary.Invocations,
				invocation{args[len(args)-1], args, uncached, code, time.Since(t).Seconds()},
			)
		}
		return code
	}
	if index == 0 {
		temp, temporaryDirectoryError := os.MkdirTemp(root, "checks-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		os.Chmod(temp, 0755)
		formatPath := filepath.Join(out, "gofmt.log")
		file, createError := os.Create(formatPath)
		if createError != nil {
			return createError
		}
		code := runCommand("gofmt", []string{"-l", "cmd", "internal"}, temp, file)
		file.Close()
		content, _ := os.ReadFile(formatPath)
		if code != 0 || len(content) > 0 {
			summary.Checks["gofmt"] = "fail"
		} else {
			summary.Checks["gofmt"] = "pass"
		}
		vet, createError := os.Create(filepath.Join(out, "vet.log"))
		if createError != nil {
			return createError
		}
		if runCommand("go", []string{"vet", "./..."}, temp, vet) != 0 {
			summary.Checks["vet"] = "fail"
		} else {
			summary.Checks["vet"] = "pass"
		}
		vet.Close()
		if removeError := os.RemoveAll(temp); removeError != nil {
			summary.Errors = append(summary.Errors, removeError.Error())
		}
	}
	byPackage := map[string][]unit{}
	for _, unit := range plan.Units {
		if unit.Shard == index {
			byPackage[unit.Package] = append(byPackage[unit.Package], unit)
		}
	}
	pkgs := []string{}
	for packagePath := range byPackage {
		pkgs = append(pkgs, packagePath)
	}
	sort.Strings(pkgs)
	fmt.Printf(
		"shard %d disk before=%d bytes scratch=%s /tmp_free=%d bytes /tmp_tmpfs=%t\n",
		index,
		beforeDisk,
		root,
		free("/tmp"),
		tmpfs("/tmp"),
	)
	for _, packagePath := range pkgs {
		packageDirectory := filepath.Join(out, "packages", shortHash(packagePath))
		key := checkpointKey(state, packagePath, patterns(byPackage[packagePath]))
		if cleanupPackageScratchError := cleanupPackageScratch(
			packageDirectory,
			key,
			root,
		); cleanupPackageScratchError != nil {
			return cleanupPackageScratchError
		}
		if resume {
			saved, found, loadPackageError := loadPackage(packageDirectory, key, plan, index, packagePath)
			if loadPackageError != nil {
				return loadPackageError
			}
			if found {
				if appendFileError := appendFile(
					raw,
					filepath.Join(packageDirectory, "test.jsonl"),
				); appendFileError != nil {
					return appendFileError
				}
				if appendFileError := appendFile(
					stderr,
					filepath.Join(packageDirectory, "test.stderr"),
				); appendFileError != nil {
					return appendFileError
				}
				summary.Invocations = append(summary.Invocations, saved.Invocations...)
				fmt.Printf("resumed %s from complete log\n", packagePath)
				continue
			}
		}
		if createDirectoryError := os.MkdirAll(packageDirectory, 0755); createDirectoryError != nil {
			return createDirectoryError
		}
		packageLog, err := os.Create(filepath.Join(packageDirectory, "test.jsonl.partial"))
		if err != nil {
			return err
		}
		packageStderr, err := os.Create(filepath.Join(packageDirectory, "test.stderr.partial"))
		if err != nil {
			packageLog.Close()
			return err
		}
		temporary, err := os.MkdirTemp(root, "package-")
		if err != nil {
			return err
		}
		if changeModeError := os.Chmod(temporary, 0755); changeModeError != nil {
			return changeModeError
		}
		if atomicJsonError := atomicJSON(
			filepath.Join(packageDirectory, "scratch.json"),
			scratchRecord{Key: key, Path: temporary},
		); atomicJsonError != nil {
			return atomicJsonError
		}
		first := len(summary.Invocations)
		commandStderr = io.MultiWriter(stderr, packageStderr)
		for _, pattern := range patterns(byPackage[packagePath]) {
			runCommand("go", testArgs(packagePath, pattern), temporary, io.MultiWriter(raw, packageLog))
		}
		commandStderr = stderr
		err = errors.Join(packageLog.Sync(), packageStderr.Sync(), packageLog.Close(), packageStderr.Close())
		if err != nil {
			return err
		}
		if renameError := os.Rename(
			filepath.Join(packageDirectory, "test.jsonl.partial"),
			filepath.Join(packageDirectory, "test.jsonl"),
		); renameError != nil {
			return renameError
		}
		if renameError := os.Rename(
			filepath.Join(packageDirectory, "test.stderr.partial"),
			filepath.Join(packageDirectory, "test.stderr"),
		); renameError != nil {
			return renameError
		}
		checkpoint := packageEvidence{Key: key, Package: packagePath, Invocations: summary.Invocations[first:]}
		if completePackageError := completePackage(
			packageDirectory,
			checkpoint,
			plan,
			index,
			packagePath,
		); completePackageError != nil {
			summary.Errors = append(summary.Errors, completePackageError.Error())
		} else {
			var fileDigestError error
			checkpoint.LogDigest, fileDigestError = fileDigest(filepath.Join(packageDirectory, "test.jsonl"))
			if fileDigestError != nil {
				return fileDigestError
			}
			checkpoint.StderrDigest, fileDigestError = fileDigest(filepath.Join(packageDirectory, "test.stderr"))
			if fileDigestError != nil {
				return fileDigestError
			}
			if atomicJsonError := atomicJSON(
				filepath.Join(packageDirectory, "complete.json"),
				checkpoint,
			); atomicJsonError != nil {
				return atomicJsonError
			}
		}
		if cleanupPackageScratchError := cleanupPackageScratch(
			packageDirectory,
			key,
			root,
		); cleanupPackageScratchError != nil {
			summary.Errors = append(summary.Errors, cleanupPackageScratchError.Error())
		}
		fmt.Printf("finished %s free=%d bytes\n", packagePath, free(root))
	}
	raw.Close()
	stderr.Close()
	afterSource, identityError := sourceIdentity()
	afterCommit, commitError := output("git", "rev-parse", "HEAD")
	status, statusError := output("git", "status", "--porcelain", "--untracked-files=normal")
	if identityError != nil ||
		commitError != nil ||
		statusError != nil ||
		afterSource != plan.Source ||
		afterCommit != plan.Commit ||
		status != "" {
		summary.Errors = append(summary.Errors, "repository changed during shard execution")
	}

	var readLogError error
	summary.Results, summary.CacheLines, _, readLogError = readLog(filepath.Join(out, "test.jsonl"))
	if readLogError != nil {
		summary.Errors = append(summary.Errors, readLogError.Error())
	}
	summary.WallSeconds = time.Since(started).Seconds()
	summary.DiskAfter = free(root)
	summary.BuildFlags = buildFlags(plan.Commit, beforeLoad, loadAverage())
	summary.Pass, summary.Fail, summary.Skip = totals(canonical(summary.Results))
	if saveJsonError := saveJSON(filepath.Join(out, "summary.json"), summary); saveJsonError != nil {
		return saveJsonError
	}
	cache, createError := os.Create(filepath.Join(out, "gate-cache.log"))
	if createError != nil {
		return createError
	}
	for _, line := range summary.CacheLines {
		fmt.Fprintln(cache, line)
	}
	cache.Close()
	fmt.Printf(
		"shard %d pass=%d fail=%d skip=%d wall=%.3fs disk after=%d bytes\n%s\n",
		index,
		summary.Pass,
		summary.Fail,
		summary.Skip,
		summary.WallSeconds,
		summary.DiskAfter,
		summary.BuildFlags,
	)
	for _, call := range summary.Invocations {
		if call.Exit != 0 {
			return errors.New("shard test command failed; evidence retained")
		}
	}
	for _, check := range summary.Checks {
		if check != "pass" {
			return errors.New("shard gate check failed")
		}
	}
	if len(summary.Errors) > 0 {
		return errors.New(strings.Join(summary.Errors, "; "))
	}
	return nil
}
func canonical(results []result) []result {
	byName := map[string]result{}
	rank := map[string]int{"skip": 1, "pass": 2, "fail": 3}
	for _, result := range results {
		old, ok := byName[result.key()]
		if !ok || rank[result.Action] > rank[old.Action] {
			byName[result.key()] = result
		}
	}
	keys := []string{}
	for key := range byName {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out []result
	for _, key := range keys {
		out = append(out, byName[key])
	}
	return out
}
func totals(results []result) (pass, fail, skip int) {
	for _, result := range results {
		if result.Test == "" {
			continue
		}
		switch result.Action {
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
func samePlan(left, right plan) bool {
	x, _ := json.Marshal(left)
	y, _ := json.Marshal(right)
	return string(x) == string(y)
}

// An enumerated child's ancestors necessarily execute in Go. Only these declared ancestors can
// repeat; all work units and all other tests have exactly one owner and one terminal event.
func owns(unit unit, result result) bool {
	return unit.Package == result.Package &&
		(result.Test == unit.Test || (unit.Test != "" && strings.HasPrefix(result.Test, unit.Test+"/")))
}
func ancestor(unit unit, result result) bool {
	return unit.Package == result.Package && result.Test != "" && strings.HasPrefix(unit.Test, result.Test+"/")
}

var cacheHits = regexp.MustCompile(`gate cache: (native|node|probe) hits=(\d+) misses=(\d+)`)

func validateResults(
	expected plan,
	index int,
	results []result,
	produced map[string]bool,
	seenTest map[string]int,
) []string {
	var problems []string
	for _, result := range results {
		if result.Test == "" {
			if result.Action == "fail" {
				problems = append(problems, "package failed: "+result.Package)
			}
			for _, unit := range expected.Units {
				if unit.Test == "" && unit.Package == result.Package && unit.Shard == index {
					produced[unit.key()] = true
				}
			}
			continue
		}
		owner := -1
		isAncestor := false
		for unitIndex, unit := range expected.Units {
			if owns(unit, result) {
				owner = unitIndex
			}
			if ancestor(unit, result) && unit.Shard == index {
				isAncestor = true
			}
		}
		if owner >= 0 {
			unit := expected.Units[owner]
			if unit.Shard != index {
				problems = append(
					problems,
					fmt.Sprintf("test %s ran in shard %d, owner %d", result.key(), index, unit.Shard),
				)
			}
			if previous, ok := seenTest[result.key()]; ok {
				problems = append(
					problems,
					fmt.Sprintf("test %s ran twice in shards %d and %d", result.key(), previous, index),
				)
			}
			seenTest[result.key()] = index
			if result.Test == unit.Test {
				produced[unit.key()] = true
			}
		} else if !isAncestor {
			problems = append(problems, "unplanned test: "+result.key())
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
		var summary summary
		if loadJsonError := loadJSON(filepath.Join(dir, "summary.json"), &summary); loadJsonError != nil {
			return loadJsonError
		}
		summaries = append(summaries, summary)
	}
	expected, makePlanError := makePlan(summaries[0].Plan.Count)
	if makePlanError != nil {
		return makePlanError
	}
	merged := merged{Version: 1, Plan: expected}
	seenShard := map[int]bool{}
	seenTest := map[string]int{}
	seenRuns := map[string]int{}
	produced := map[string]bool{}
	var all []result
	for index, summary := range summaries {
		if summary.Version != 1 || !samePlan(expected, summary.Plan) {
			merged.Errors = append(
				merged.Errors,
				fmt.Sprintf("shard %d plan differs from repository enumeration", summary.Index),
			)
		}
		if summary.Index < 0 || summary.Index >= expected.Count || seenShard[summary.Index] {
			merged.Errors = append(merged.Errors, fmt.Sprintf("invalid or duplicate shard %d", summary.Index))
		}
		seenShard[summary.Index] = true
		if summary.Index == 0 {
			for _, check := range []string{"gofmt", "vet"} {
				if summary.Checks[check] != "pass" {
					merged.Errors = append(merged.Errors, check+" did not pass")
				}
			}
		}
		for _, call := range summary.Invocations {
			if !call.Uncached {
				merged.Errors = append(
					merged.Errors,
					fmt.Sprintf("shard %d invocation was not uncached", summary.Index),
				)
			}
			if call.Exit != 0 {
				merged.Errors = append(merged.Errors, fmt.Sprintf("shard %d command exit %d", summary.Index, call.Exit))
			}
		}
		hasUnits := false
		for _, unit := range expected.Units {
			if unit.Shard == summary.Index {
				hasUnits = true
			}
		}
		if hasUnits && len(summary.Invocations) == 0 {
			merged.Errors = append(merged.Errors, fmt.Sprintf("shard %d has no invocation evidence", summary.Index))
		}
		merged.Errors = append(merged.Errors, summary.Errors...)
		merged.Errors = append(merged.Errors, validatePackageFiles(dirs[index], expected, summary)...)
		results, lines, raw, readLogError := readLog(filepath.Join(dirs[index], "test.jsonl"))
		if readLogError != nil {
			return readLogError
		}
		merged.RawTerminalEvents += raw
		for _, result := range results {
			if result.Test != "" && result.Action == "fail" {
				merged.Errors = append(
					merged.Errors,
					fmt.Sprintf("shard %d test %s failed\n%s", summary.Index, result.key(), result.Output),
				)
			}
		}

		declared := map[string]int{}
		observed := map[string]int{}
		for _, call := range summary.Invocations {
			declared[call.Package]++
			assigned := false
			for _, unit := range expected.Units {
				if unit.Package == call.Package && unit.Shard == summary.Index {
					assigned = true
				}
			}
			if !assigned {
				merged.Errors = append(merged.Errors, "invocation for unassigned package: "+call.Package)
			}
		}
		for _, result := range results {
			if result.Test == "" {
				observed[result.Package]++
			}
		}
		merged.Errors = append(merged.Errors, invocationCountErrors(declared, observed)...)
		for _, line := range lines {
			match := cacheHits.FindStringSubmatch(line)
			if match == nil || match[2] != "0" {
				merged.Errors = append(
					merged.Errors,
					fmt.Sprintf("shard %d refuses cache line: %s", summary.Index, line),
				)
			}
		}
		merged.Errors = append(merged.Errors, validateResults(expected, summary.Index, results, produced, seenTest)...)
		runs, readRunsError := readRuns(filepath.Join(dirs[index], "test.jsonl"))
		if readRunsError != nil {
			return readRunsError
		}
		merged.Errors = append(
			merged.Errors,
			validateResults(expected, summary.Index, runs, map[string]bool{}, seenRuns)...,
		)
		ran := map[string]bool{}
		for _, run := range runs {
			ran[run.key()] = true
		}
		for _, result := range results {
			if result.Test != "" && !ran[result.key()] {
				merged.Errors = append(merged.Errors, "terminal event without run: "+result.key())
			}
		}
		finished := map[string]bool{}
		for _, result := range results {
			if result.Test != "" {
				finished[result.key()] = true
			}
		}
		for _, run := range runs {
			if !finished[run.key()] {
				merged.Errors = append(merged.Errors, "run without terminal event: "+run.key())
			}
		}

		all = append(all, results...)
		merged.WallSeconds = append(merged.WallSeconds, summary.WallSeconds)
		merged.BuildFlags = append(merged.BuildFlags, summary.BuildFlags)
	}
	for index := 0; index < expected.Count; index++ {
		if !seenShard[index] {
			merged.Errors = append(merged.Errors, fmt.Sprintf("missing shard %d", index))
		}
	}
	for _, unit := range expected.Units {
		if !produced[unit.key()] {
			merged.Errors = append(merged.Errors, "planned unit produced no terminal test event: "+unit.key())
		}
	}
	merged.Results = canonical(all)
	merged.Pass, merged.Fail, merged.Skip = totals(merged.Results)
	merged.TestEvents = merged.Pass + merged.Fail + merged.Skip
	merged.Green = len(merged.Errors) == 0 && merged.Fail == 0
	if _, statError := os.Stat(out); statError == nil {
		return errors.New("merged output already exists")
	}
	if createDirectoryError := os.MkdirAll(out, 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	if saveJsonError := saveJSON(filepath.Join(out, "merged.json"), merged); saveJsonError != nil {
		return saveJsonError
	}
	file, createError := os.Create(filepath.Join(out, "test.jsonl"))
	if createError != nil {
		return createError
	}
	for _, dir := range dirs {
		input, openError := os.Open(filepath.Join(dir, "test.jsonl"))
		if openError != nil {
			file.Close()
			return openError
		}
		_, copyError := io.Copy(file, input)
		input.Close()
		if copyError != nil {
			file.Close()
			return copyError
		}
	}
	file.Close()
	verdict := "RED"
	if merged.Green {
		verdict = "GREEN"
	}
	fmt.Printf(
		"%s pass=%d fail=%d skip=%d test_events=%d raw_terminal_events=%d\n",
		verdict,
		merged.Pass,
		merged.Fail,
		merged.Skip,
		merged.TestEvents,
		merged.RawTerminalEvents,
	)
	for _, problem := range merged.Errors {
		fmt.Fprintln(os.Stderr, problem)
	}
	if !merged.Green {
		return errors.New("merged gate is not green")
	}
	return nil
}

// invocationCountErrors names each package whose go test invocations and terminal events disagree, in package order, so
// a record's errors are the same bytes on every run (ranging over the maps would order them by Go's per-run hash seed).
func invocationCountErrors(declared, observed map[string]int) []string {
	var problems []string
	for _, packagePath := range sortedKeys(declared) {
		if observed[packagePath] != declared[packagePath] {
			problems = append(problems, "invocation/terminal count mismatch: "+packagePath)
		}
	}
	for _, packagePath := range sortedKeys(observed) {
		if declared[packagePath] != observed[packagePath] {
			problems = append(problems, "unrecorded package invocation: "+packagePath)
		}
	}
	return problems
}

func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func compare(path, log string) error {
	if info, statError := os.Stat(path); statError == nil && info.IsDir() {
		path = filepath.Join(path, "merged.json")
	}
	var merged merged
	if loadJsonError := loadJSON(path, &merged); loadJsonError != nil {
		return loadJsonError
	}
	whole, _, raw, readLogError := readLog(log)
	if readLogError != nil {
		return readLogError
	}
	whole = canonical(whole)
	for _, result := range whole {
		if result.Test == "" && result.Action == "fail" {
			return fmt.Errorf("unsharded package failed: %s", result.Package)
		}
	}
	a, b := map[string]string{}, map[string]string{}
	for _, result := range merged.Results {
		if result.Test != "" {
			a[result.key()] = result.Action
		}
	}
	for _, result := range whole {
		if result.Test != "" {
			b[result.key()] = result.Action
		}
	}
	keys := map[string]bool{}
	for key := range a {
		keys[key] = true
	}
	for key := range b {
		keys[key] = true
	}
	var diff []string
	for key := range keys {
		if a[key] != b[key] {
			diff = append(diff, fmt.Sprintf("%s merged=%q whole=%q", key, a[key], b[key]))
		}
	}
	sort.Strings(diff)
	pass, fail, skip := totals(whole)
	fmt.Printf(
		"merged pass=%d fail=%d skip=%d events=%d\nwhole pass=%d fail=%d skip=%d events=%d raw_terminal_events=%d\n",
		merged.Pass,
		merged.Fail,
		merged.Skip,
		len(a),
		pass,
		fail,
		skip,
		len(b),
		raw,
	)
	for _, line := range diff {
		fmt.Println(line)
	}
	if !merged.Green || len(diff) > 0 || len(a) != len(b) || raw != len(b) {
		return errors.New("compare failed")
	}
	fmt.Println("diff: empty")
	return nil
}
