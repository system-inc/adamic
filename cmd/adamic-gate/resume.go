package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"syscall"
)

type prediction struct {
	Shard         int
	Seconds       float64
	ParentSeconds float64
	Largest       unit
	Unknown       int
}

func predictions(p plan, weights map[string]float64) []prediction {
	result := make([]prediction, p.Count)
	childSeconds := map[string]float64{}
	known := map[string]bool{}
	for _, u := range p.Units {
		if strings.Contains(u.Test, "/") {
			parent := u.Package + "::" + strings.Split(u.Test, "/")[0]
			childSeconds[parent] += u.Seconds
			if _, ok := known[parent]; !ok {
				known[parent] = true
			}
			if _, ok := weights[u.key()]; !ok {
				known[parent] = false
			}
		}
	}
	// Sequential audited parents repeat their work outside t.Run on each worker.
	// For parallel parents the sum exceeds the parent span, so no residual is inferred.
	seen := map[string]bool{}
	for i := range result {
		result[i].Shard = i
	}
	for _, u := range p.Units {
		r := &result[u.Shard]
		r.Seconds += u.Seconds
		if _, ok := weights[u.key()]; !ok {
			r.Unknown++
		}
		if r.Largest.Package == "" ||
			u.Seconds > r.Largest.Seconds ||
			(u.Seconds == r.Largest.Seconds && u.key() < r.Largest.key()) {
			r.Largest = u
		}
		if strings.Contains(u.Test, "/") {
			parent := u.Package + "::" + strings.Split(u.Test, "/")[0]
			invocation := fmt.Sprintf(
				"%d::%s::%s",
				u.Shard,
				parent,
				strings.Join(strings.Split(u.Test, "/")[:len(strings.Split(u.Test, "/"))-1], "/"),
			)
			if !seen[invocation] && known[parent] {
				residual := weights[parent] - childSeconds[parent]
				if residual > 0 {
					r.ParentSeconds += residual
					r.Seconds += residual
				}
			}
			seen[invocation] = true
		}
	}
	return result
}
func digestJSON(value any) string {
	b, _ := json.Marshal(value)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func planDigest(p plan) string  { p.Digest = ""; return digestJSON(p) }
func shortHash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:16]) }
func fileDigest(path string) (string, error) {
	f, openError := os.Open(path)
	if openError != nil {
		return "", openError
	}
	defer f.Close()
	h := sha256.New()
	if _, copyError := io.Copy(h, f); copyError != nil {
		return "", copyError
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func pathDigest(path string) (string, error) { return digestPath(path, map[string]bool{}) }
func digestPath(path string, active map[string]bool) (string, error) {
	absolute, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		return "", absolutePathError
	}
	if active[absolute] {
		return "", fmt.Errorf("cyclic resume input %s", path)
	}
	active[absolute] = true
	defer delete(active, absolute)
	info, statError := os.Lstat(path)
	if os.IsNotExist(statError) {
		return "missing", nil
	}
	if statError != nil {
		return "", statError
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, readLinkError := os.Readlink(path)
		if readLinkError != nil {
			return "", readLinkError
		}
		resolved, resolveSymlinksError := filepath.EvalSymlinks(path)
		if os.IsNotExist(resolveSymlinksError) {
			return digestJSON([]string{target, "dangling"}), nil
		}
		if resolveSymlinksError != nil {
			return "", resolveSymlinksError
		}
		answer, digestPathError := digestPath(resolved, active)
		return digestJSON([]string{target, answer}), digestPathError
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("unsupported resume input %s", path)
		}
		data, fileDigestError := fileDigest(path)
		if fileDigestError != nil {
			return "", fileDigestError
		}
		stat := info.Sys().(*syscall.Stat_t)
		return digestJSON([]any{uint32(info.Mode()), stat.Uid, stat.Gid, data}), nil
	}
	entries, readDirectoryError := os.ReadDir(path)
	if readDirectoryError != nil {
		return "", readDirectoryError
	}
	stat := info.Sys().(*syscall.Stat_t)
	records := []string{digestJSON([]any{uint32(info.Mode()), stat.Uid, stat.Gid})}
	for _, e := range entries {
		key, digestPathError := digestPath(filepath.Join(path, e.Name()), active)
		if digestPathError != nil {
			return "", digestPathError
		}
		records = append(records, e.Name(), key)
	}
	return digestJSON(records), nil
}

// Resume is explicit evidence reuse. Hash private environment values; never print them.
// ADAMIC paths may point to mutable external corpora, binaries, or npm installations.
func executionIdentity() (string, error) {
	env := append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
	effective := map[string]string{}
	for _, v := range env {
		k, value, _ := strings.Cut(v, "=")
		effective[k] = value
	}
	env = nil
	external := map[string]string{}
	for k, v := range effective {
		env = append(env, k+"="+v)
		if strings.HasPrefix(k, "ADAMIC_") && v != "" {
			if _, statError := os.Lstat(v); statError == nil {
				d, pathDigestError := pathDigest(v)
				if pathDigestError != nil {
					return "", pathDigestError
				}
				external[k] = d
			} else if !os.IsNotExist(statError) {
				return "", statError
			}
		}
	}
	sort.Strings(env)
	defaultWidth, pathDigestError := pathDigest("/tmp/adamic-markdown-width")
	if pathDigestError != nil {
		return "", pathDigestError
	}
	external["default-width"] = defaultWidth
	tools := map[string]string{}
	executable, executableError := os.Executable()
	if executableError != nil {
		return "", executableError
	}
	gateHash, fileDigestError := fileDigest(executable)
	if fileDigestError != nil {
		return "", fileDigestError
	}
	tools["gate"] = gateHash
	for _, tool := range []string{"go", "clang", "node", "gofmt"} {
		path, lookPathError := exec.LookPath(tool)
		if lookPathError != nil {
			return "", lookPathError
		}
		path, resolveSymlinksError := filepath.EvalSymlinks(path)
		if resolveSymlinksError != nil {
			return "", resolveSymlinksError
		}
		hash, fileDigestError := fileDigest(path)
		if fileDigestError != nil {
			return "", fileDigestError
		}
		tools[tool] = path + ":" + hash
	}
	goEnv, outputError := output("go", "env", "-json")
	if outputError != nil {
		return "", outputError
	}
	var values map[string]string
	if decodeError := json.Unmarshal([]byte(goEnv), &values); decodeError != nil {
		return "", decodeError
	}
	tools["go-env"] = stableGoEnvironment(values)
	for _, name := range []string{"compile", "link", "asm", "cgo"} {
		path := filepath.Join(values["GOTOOLDIR"], name)
		hash, fileDigestError := fileDigest(path)
		if fileDigestError != nil {
			return "", fileDigestError
		}
		tools[name] = hash
	}
	resource, outputError := output("clang", "-print-resource-dir")
	if outputError != nil {
		return "", outputError
	}
	hash, pathDigestError := pathDigest(resource)
	if pathDigestError != nil {
		return "", pathDigestError
	}
	tools["clang-resources"] = hash
	return executionKey(env, tools, external), nil
}

func stableGoEnvironment(values map[string]string) string {
	copy := map[string]string{}
	for key, value := range values {
		copy[key] = value
	}
	// Go creates this random work-directory remapping even for `go env`.
	// The actual TMPDIR and all compiler flags remain independently keyed.
	copy["GOGCCFLAGS"] = regexp.MustCompile(`go-build[0-9]+=/tmp/go-build`).ReplaceAllString(
		copy["GOGCCFLAGS"],
		"go-build=/tmp/go-build",
	)
	return digestJSON(copy)
}

func executionKey(env []string, tools, external map[string]string) string {
	return digestJSON(struct {
		Env             []string
		Tools, External map[string]string
	}{env, tools, external})
}

type resumeState struct {
	PlanDigest, Context string
	Index               int
}
type packageEvidence struct {
	Key, Package, LogDigest, StderrDigest string
	Invocations                           []invocation
}

func checkpointKey(s resumeState, pkg string, patterns []string) string {
	return digestJSON(struct {
		State    resumeState
		Package  string
		Patterns []string
		Flags    []string
	}{s, pkg, patterns, []string{"-count=1", "-json", "-timeout", "60m", "ADAMIC_GATE_UNCACHED=1"}})
}
func testArgs(pkg, pattern string) []string {
	return []string{"test", "-count=1", "-json", "-timeout", "60m", "-run", pattern, pkg}
}
func appendFile(out io.Writer, path string) error {
	f, openError := os.Open(path)
	if openError != nil {
		return openError
	}
	defer f.Close()
	_, copyError := io.Copy(out, f)
	return copyError
}
func atomicJSON(path string, v any) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".checkpoint-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	err = writeJSON(f, v)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	if renameError := os.Rename(tmp, path); renameError != nil {
		return renameError
	}
	d, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
func completePackage(dir string, e packageEvidence, p plan, index int, pkg string) error {
	results, lines, _, readLogError := readLog(filepath.Join(dir, "test.jsonl"))
	if readLogError != nil {
		return readLogError
	}
	runs, readRunsError := readRuns(filepath.Join(dir, "test.jsonl"))
	if readRunsError != nil {
		return readRunsError
	}
	units := []unit{}
	for _, u := range p.Units {
		if u.Package == pkg && u.Shard == index {
			units = append(units, u)
		}
	}
	selectors := patterns(units)
	if len(e.Invocations) != len(selectors) {
		return errors.New("package invocation count differs")
	}
	for i, call := range e.Invocations {
		if call.Package != pkg || !call.Uncached || !reflect.DeepEqual(call.Args, testArgs(pkg, selectors[i])) {
			return errors.New("package invocation inputs differ")
		}
	}
	produced := map[string]bool{}
	problems := validateResults(p, index, results, produced, map[string]int{})
	problems = append(problems, validateResults(p, index, runs, map[string]bool{}, map[string]int{})...)
	for _, u := range units {
		if !produced[u.key()] {
			problems = append(problems, "planned unit has no terminal event: "+u.key())
		}
	}
	starts, finishes := map[string]int{}, map[string]int{}
	packageTerminals := 0
	for _, r := range runs {
		starts[r.key()]++
	}
	for _, r := range results {
		if r.Package != pkg {
			problems = append(problems, "foreign package in checkpoint")
		}
		if r.Test == "" {
			packageTerminals++
		} else {
			finishes[r.key()]++
		}
	}
	if packageTerminals != len(e.Invocations) {
		problems = append(problems, "package log has no complete invocation terminal events")
	}
	if !reflect.DeepEqual(starts, finishes) {
		problems = append(problems, "package run/terminal events differ")
	}
	for _, line := range lines {
		m := cacheHits.FindStringSubmatch(line)
		if m == nil || m[2] != "0" {
			problems = append(problems, "checkpoint has cache hits or malformed cache line")
		}
	}
	// A complete failed package remains a failed checkpoint; never turn it green on resume.
	var structural []string
	for _, problem := range problems {
		if !strings.HasPrefix(problem, "package failed:") {
			structural = append(structural, problem)
		}
	}
	if len(structural) > 0 {
		return errors.New(strings.Join(structural, "; "))
	}
	return nil
}
func loadPackage(dir, key string, p plan, index int, pkg string) (packageEvidence, bool, error) {
	var e packageEvidence
	path := filepath.Join(dir, "complete.json")
	if _, statError := os.Stat(path); os.IsNotExist(statError) {
		return e, false, nil
	} else if statError != nil {
		return e, false, statError
	}
	if loadJsonError := loadJSON(path, &e); loadJsonError != nil {
		return e, false, loadJsonError
	}
	if e.Key != key || e.Package != pkg {
		return e, false, errors.New("package resume key differs")
	}
	log, fileDigestError := fileDigest(filepath.Join(dir, "test.jsonl"))
	if fileDigestError != nil {
		return e, false, fileDigestError
	}
	stderr, fileDigestError := fileDigest(filepath.Join(dir, "test.stderr"))
	if fileDigestError != nil {
		return e, false, fileDigestError
	}
	if log != e.LogDigest || stderr != e.StderrDigest {
		return e, false, errors.New("package checkpoint checksum differs")
	}
	if completePackageError := completePackage(dir, e, p, index, pkg); completePackageError != nil {
		return e, false, completePackageError
	}
	return e, true, nil
}

// The aggregate is convenient for consumers, but completed package files are the
// durable evidence. Verify both representations, including their terminal events.
func validatePackageFiles(dir string, p plan, s summary) []string {
	var state resumeState
	if loadJsonError := loadJSON(filepath.Join(dir, "run.json"), &state); loadJsonError != nil {
		return []string{"missing package checkpoint state: " + loadJsonError.Error()}
	}
	if state.PlanDigest != p.Digest || state.Index != s.Index || state.Context == "" {
		return []string{"package checkpoint state differs from plan"}
	}
	byPackage := map[string][]unit{}
	for _, u := range p.Units {
		if u.Shard == s.Index {
			byPackage[u.Package] = append(byPackage[u.Package], u)
		}
	}
	packages := []string{}
	for pkg := range byPackage {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	h := sha256.New()
	var calls []invocation
	var problems []string
	for _, pkg := range packages {
		packageDir := filepath.Join(dir, "packages", shortHash(pkg))
		e, found, loadPackageError := loadPackage(
			packageDir,
			checkpointKey(state, pkg, patterns(byPackage[pkg])),
			p,
			s.Index,
			pkg,
		)
		if loadPackageError != nil || !found {
			problems = append(
				problems,
				fmt.Sprintf("shard %d package %s has no valid complete log: %v", s.Index, pkg, loadPackageError),
			)
			continue
		}
		calls = append(calls, e.Invocations...)
		if appendFileError := appendFile(h, filepath.Join(packageDir, "test.jsonl")); appendFileError != nil {
			problems = append(problems, appendFileError.Error())
		}
	}
	digest, fileDigestError := fileDigest(filepath.Join(dir, "test.jsonl"))
	if fileDigestError != nil || digest != hex.EncodeToString(h.Sum(nil)) {
		problems = append(problems, "aggregate differs from complete package logs")
	}
	if !reflect.DeepEqual(calls, s.Invocations) {
		problems = append(problems, "summary invocation evidence differs from package logs")
	}
	return problems
}

// The lock excludes an old live runner. Only delete this package's recorded scratch.
type scratchRecord struct{ Key, Path string }

func cleanupPackageScratch(dir, key, root string) error {
	path := filepath.Join(dir, "scratch.json")
	if _, statError := os.Stat(path); os.IsNotExist(statError) {
		return nil
	} else if statError != nil {
		return statError
	}
	var record scratchRecord
	if loadJsonError := loadJSON(path, &record); loadJsonError != nil {
		return loadJsonError
	}
	absolute, absolutePathError := filepath.Abs(record.Path)
	if absolutePathError != nil {
		return absolutePathError
	}
	if record.Key != key || filepath.Dir(absolute) != root || !strings.HasPrefix(filepath.Base(absolute), "package-") {
		return errors.New("refusing unowned package scratch path")
	}
	if removeError := os.RemoveAll(absolute); removeError != nil {
		return removeError
	}
	return os.Remove(path)
}
