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

func predictions(plan plan, weights map[string]float64) []prediction {
	result := make([]prediction, plan.Count)
	childSeconds := map[string]float64{}
	known := map[string]bool{}
	for _, unit := range plan.Units {
		if strings.Contains(unit.Test, "/") {
			parent := unit.Package + "::" + strings.Split(unit.Test, "/")[0]
			childSeconds[parent] += unit.Seconds
			if _, ok := known[parent]; !ok {
				known[parent] = true
			}
			if _, ok := weights[unit.key()]; !ok {
				known[parent] = false
			}
		}
	}
	// Sequential audited parents repeat their work outside t.Run on each worker.
	// For parallel parents the sum exceeds the parent span, so no residual is inferred.
	seen := map[string]bool{}
	for index := range result {
		result[index].Shard = index
	}
	for _, unit := range plan.Units {
		prediction := &result[unit.Shard]
		prediction.Seconds += unit.Seconds
		if _, ok := weights[unit.key()]; !ok {
			prediction.Unknown++
		}
		if prediction.Largest.Package == "" ||
			unit.Seconds > prediction.Largest.Seconds ||
			(unit.Seconds == prediction.Largest.Seconds && unit.key() < prediction.Largest.key()) {
			prediction.Largest = unit
		}
		if strings.Contains(unit.Test, "/") {
			parent := unit.Package + "::" + strings.Split(unit.Test, "/")[0]
			invocation := fmt.Sprintf(
				"%d::%s::%s",
				unit.Shard,
				parent,
				strings.Join(strings.Split(unit.Test, "/")[:len(strings.Split(unit.Test, "/"))-1], "/"),
			)
			if !seen[invocation] && known[parent] {
				residual := weights[parent] - childSeconds[parent]
				if residual > 0 {
					prediction.ParentSeconds += residual
					prediction.Seconds += residual
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
func planDigest(plan plan) string { plan.Digest = ""; return digestJSON(plan) }
func shortHash(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:16])
}
func fileDigest(path string) (string, error) {
	file, openError := os.Open(path)
	if openError != nil {
		return "", openError
	}
	defer file.Close()
	hash := sha256.New()
	if _, copyError := io.Copy(hash, file); copyError != nil {
		return "", copyError
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
	for _, entry := range entries {
		key, digestPathError := digestPath(filepath.Join(path, entry.Name()), active)
		if digestPathError != nil {
			return "", digestPathError
		}
		records = append(records, entry.Name(), key)
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
	for key, value := range effective {
		env = append(env, key+"="+value)
		if strings.HasPrefix(key, "ADAMIC_") && value != "" {
			if _, statError := os.Lstat(value); statError == nil {
				d, pathDigestError := pathDigest(value)
				if pathDigestError != nil {
					return "", pathDigestError
				}
				external[key] = d
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

func checkpointKey(state resumeState, packagePath string, patterns []string) string {
	return digestJSON(struct {
		State    resumeState
		Package  string
		Patterns []string
		Flags    []string
	}{state, packagePath, patterns, []string{"-count=1", "-json", "-timeout", "60m", "ADAMIC_GATE_UNCACHED=1"}})
}
func testArgs(packagePath, pattern string) []string {
	return []string{"test", "-count=1", "-json", "-timeout", "60m", "-run", pattern, packagePath}
}
func appendFile(out io.Writer, path string) error {
	file, openError := os.Open(path)
	if openError != nil {
		return openError
	}
	defer file.Close()
	_, copyError := io.Copy(out, file)
	return copyError
}
func atomicJSON(path string, v any) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".checkpoint-")
	if err != nil {
		return err
	}
	temporary := f.Name()
	defer os.Remove(temporary)
	err = writeJSON(f, v)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	if renameError := os.Rename(temporary, path); renameError != nil {
		return renameError
	}
	d, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
func completePackage(dir string, evidence packageEvidence, plan plan, index int, packagePath string) error {
	results, lines, _, readLogError := readLog(filepath.Join(dir, "test.jsonl"))
	if readLogError != nil {
		return readLogError
	}
	runs, readRunsError := readRuns(filepath.Join(dir, "test.jsonl"))
	if readRunsError != nil {
		return readRunsError
	}
	units := []unit{}
	for _, unit := range plan.Units {
		if unit.Package == packagePath && unit.Shard == index {
			units = append(units, unit)
		}
	}
	selectors := patterns(units)
	if len(evidence.Invocations) != len(selectors) {
		return errors.New("package invocation count differs")
	}
	for index, call := range evidence.Invocations {
		if call.Package != packagePath ||
			!call.Uncached ||
			!reflect.DeepEqual(call.Args, testArgs(packagePath, selectors[index])) {
			return errors.New("package invocation inputs differ")
		}
	}
	produced := map[string]bool{}
	problems := validateResults(plan, index, results, produced, map[string]int{})
	problems = append(problems, validateResults(plan, index, runs, map[string]bool{}, map[string]int{})...)
	for _, unit := range units {
		if !produced[unit.key()] {
			problems = append(problems, "planned unit has no terminal event: "+unit.key())
		}
	}
	starts, finishes := map[string]int{}, map[string]int{}
	packageTerminals := 0
	for _, run := range runs {
		starts[run.key()]++
	}
	for _, result := range results {
		if result.Package != packagePath {
			problems = append(problems, "foreign package in checkpoint")
		}
		if result.Test == "" {
			packageTerminals++
		} else {
			finishes[result.key()]++
		}
	}
	if packageTerminals != len(evidence.Invocations) {
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
func loadPackage(dir, key string, plan plan, index int, packagePath string) (packageEvidence, bool, error) {
	var evidence packageEvidence
	path := filepath.Join(dir, "complete.json")
	if _, statError := os.Stat(path); os.IsNotExist(statError) {
		return evidence, false, nil
	} else if statError != nil {
		return evidence, false, statError
	}
	if loadJsonError := loadJSON(path, &evidence); loadJsonError != nil {
		return evidence, false, loadJsonError
	}
	if evidence.Key != key || evidence.Package != packagePath {
		return evidence, false, errors.New("package resume key differs")
	}
	log, fileDigestError := fileDigest(filepath.Join(dir, "test.jsonl"))
	if fileDigestError != nil {
		return evidence, false, fileDigestError
	}
	stderr, fileDigestError := fileDigest(filepath.Join(dir, "test.stderr"))
	if fileDigestError != nil {
		return evidence, false, fileDigestError
	}
	if log != evidence.LogDigest || stderr != evidence.StderrDigest {
		return evidence, false, errors.New("package checkpoint checksum differs")
	}
	if completePackageError := completePackage(dir, evidence, plan, index, packagePath); completePackageError != nil {
		return evidence, false, completePackageError
	}
	return evidence, true, nil
}

// The aggregate is convenient for consumers, but completed package files are the
// durable evidence. Verify both representations, including their terminal events.
func validatePackageFiles(dir string, plan plan, summary summary) []string {
	var state resumeState
	if loadJsonError := loadJSON(filepath.Join(dir, "run.json"), &state); loadJsonError != nil {
		return []string{"missing package checkpoint state: " + loadJsonError.Error()}
	}
	if state.PlanDigest != plan.Digest || state.Index != summary.Index || state.Context == "" {
		return []string{"package checkpoint state differs from plan"}
	}
	byPackage := map[string][]unit{}
	for _, unit := range plan.Units {
		if unit.Shard == summary.Index {
			byPackage[unit.Package] = append(byPackage[unit.Package], unit)
		}
	}
	packages := []string{}
	for packagePath := range byPackage {
		packages = append(packages, packagePath)
	}
	sort.Strings(packages)
	hash := sha256.New()
	var calls []invocation
	var problems []string
	for _, packagePath := range packages {
		packageDirectory := filepath.Join(dir, "packages", shortHash(packagePath))
		evidence, found, loadPackageError := loadPackage(
			packageDirectory,
			checkpointKey(state, packagePath, patterns(byPackage[packagePath])),
			plan,
			summary.Index,
			packagePath,
		)
		if loadPackageError != nil || !found {
			problems = append(
				problems,
				fmt.Sprintf(
					"shard %d package %s has no valid complete log: %v",
					summary.Index,
					packagePath,
					loadPackageError,
				),
			)
			continue
		}
		calls = append(calls, evidence.Invocations...)
		if appendFileError := appendFile(hash, filepath.Join(packageDirectory, "test.jsonl")); appendFileError != nil {
			problems = append(problems, appendFileError.Error())
		}
	}
	digest, fileDigestError := fileDigest(filepath.Join(dir, "test.jsonl"))
	if fileDigestError != nil || digest != hex.EncodeToString(hash.Sum(nil)) {
		problems = append(problems, "aggregate differs from complete package logs")
	}
	if !reflect.DeepEqual(calls, summary.Invocations) {
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
