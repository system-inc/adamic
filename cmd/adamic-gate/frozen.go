package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type sourceRecord struct{ Name, Index, Digest string }
type inputRecord struct{ Kind, Digest string }
type frozenIdentity struct {
	Version       int
	SourceFiles   []sourceRecord
	Tools         map[string]string
	GoEnvironment map[string]string
	Inputs        map[string]inputRecord
}

func sourceRecords() ([]sourceRecord, error) {
	var records []sourceRecord
	var walk func(string) error
	walk = func(directory string) error {
		tracked, err := output("git", "-C", directory, "ls-files", "-s", "-z")
		if err != nil {
			return err
		}
		for _, line := range strings.Split(tracked, "\x00") {
			metadata, name, ok := strings.Cut(line, "\t")
			if !ok {
				continue
			}
			name = filepath.Join(directory, name)
			var digest string
			if strings.HasPrefix(metadata, "160000 ") {
				if _, err := os.Stat(filepath.Join(name, ".git")); err != nil {
					return fmt.Errorf("source submodule %s is uninitialized", name)
				}
				digest, err = output("git", "-C", name, "rev-parse", "HEAD")
			} else if strings.HasPrefix(metadata, "120000 ") {
				var target string
				target, err = os.Readlink(name)
				digest = digestJSON(target)
			} else {
				digest, err = fileDigest(name)
			}
			if err != nil {
				return fmt.Errorf("source %s: %w", name, err)
			}
			records = append(records, sourceRecord{Name: filepath.ToSlash(name), Index: metadata, Digest: digest})
			if strings.HasPrefix(metadata, "160000 ") {
				if err := walk(name); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk("."); err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	return records, nil
}

func toolIdentity() (map[string]string, error) {
	tools := map[string]string{}
	for _, name := range []string{"go", "clang", "node"} {
		args := []string{"--version"}
		if name == "go" {
			args = []string{"version"}
		}
		version, err := output(name, args...)
		if err != nil {
			return nil, err
		}
		path, err := exec.LookPath(name)
		if err != nil {
			return nil, err
		}
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			return nil, err
		}
		digest, err := fileDigest(path)
		if err != nil {
			return nil, err
		}
		tools[name] = strings.Split(version, "\n")[0]
		tools[name+"-binary"] = digest
	}
	resource, err := output("clang", "-print-resource-dir")
	if err != nil {
		return nil, err
	}
	digest, err := inputDigest(resource)
	if err != nil {
		return nil, err
	}
	tools["clang-resources"] = digest
	return tools, nil
}

func planningGoEnvironment() (map[string]string, error) {
	text, err := output("go", "env", "-json")
	if err != nil {
		return nil, err
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(text), &values); err != nil {
		return nil, err
	}
	root, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, key := range []string{"GOOS", "GOARCH", "GOAMD64", "GOARM", "GOARM64", "GOEXPERIMENT", "CGO_ENABLED", "GOFLAGS", "GOTOOLCHAIN", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOMOD", "GOWORK"} {
		value := values[key]
		if key == "GOMOD" || key == "GOWORK" {
			if value == root || strings.HasPrefix(value, root+string(os.PathSeparator)) {
				value = "$REPO" + strings.TrimPrefix(value, root)
			}
		}
		result[key] = value
	}
	return result, nil
}

// Hash answer-bearing bytes and permissions, not clone-specific .git pack layouts.
// A Git input also records its checked-out commit. Read fresh on every validation.
func inputDigest(root string) (string, error) {
	h := sha256.New()
	active := map[string]bool{}
	var walk func(string, string) error
	walk = func(path, name string) error {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if active[absolute] {
			return fmt.Errorf("cyclic input %s", path)
		}
		active[absolute] = true
		defer delete(active, absolute)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%d:%s:%d:", len(name), name, info.Mode())
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "%d:%s", len(target), target)
			next := target
			if !filepath.IsAbs(next) {
				next = filepath.Join(filepath.Dir(path), target)
			}
			return walk(next, name+"->")
		}
		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
				commit, err := output("git", "-C", path, "rev-parse", "HEAD")
				if err != nil {
					return err
				}
				fmt.Fprint(h, commit)
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.Name() == ".git" {
					continue
				}
				if err := walk(filepath.Join(path, entry.Name()), name+"/"+entry.Name()); err != nil {
					return err
				}
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported input %s", path)
		}
		fmt.Fprintf(h, "%d:", info.Size())
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(h, f)
		return errors.Join(copyErr, f.Close())
	}
	if err := walk(root, ""); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func operationalInput(name string) bool {
	return name == "ADAMIC_TOOLS" || name == "ADAMIC_SETUP_REPOSITORY" || name == "ADAMIC_GATE_UNCACHED" || name == "ADAMIC_GATE_PLAN" || name == "ADAMIC_GATE_LAUNCH" || strings.HasPrefix(name, "ADAMIC_GATE_FLEET_")
}

func planningInputs(p plan, index int) (map[string]inputRecord, error) {
	values := map[string]string{}
	for _, name := range []string{"NODE_OPTIONS", "NODE_PATH", "LANG", "LC_ALL", "TZ", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", "CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH"} {
		values[name] = os.Getenv(name)
	}
	for _, entry := range os.Environ() {
		name, value, _ := strings.Cut(entry, "=")
		if (strings.HasPrefix(name, "ADAMIC_") || name == "WASI_SYSROOT") && !operationalInput(name) {
			values[name] = value
		}
	}
	if p.Archive != nil && (index < 0 || p.Archive.Shard == index) {
		if os.Getenv(archiveVariable) == "" {
			return nil, fmt.Errorf("frozen plan input %s for %s is missing; provision all planning inputs", archiveVariable, archiveUnit)
		}
		values[archiveVariable] = os.Getenv(archiveVariable)
	} else {
		delete(values, archiveVariable)
	}
	// The default width oracle is used when no explicit dependency path is set.
	if values["ADAMIC_MARKDOWNWIDTH_DEPS"] == "" {
		values["default-markdown-width"] = "/tmp/adamic-markdown-width"
	}
	for _, variable := range []string{"GOFLAGS", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS"} {
		for index, token := range strings.Fields(os.Getenv(variable)) {
			path := token
			if _, value, ok := strings.Cut(token, "="); ok {
				path = value
			}
			for _, prefix := range []string{"-I", "-L", "-B"} {
				if strings.HasPrefix(path, prefix) {
					path = strings.TrimPrefix(path, prefix)
					break
				}
			}
			if _, err := os.Stat(path); err == nil && path != "" {
				values[fmt.Sprintf("%s-file-%d", variable, index)] = path
			}
		}
	}
	result := map[string]inputRecord{}
	for name, value := range values {
		record := inputRecord{Kind: "literal", Digest: digestJSON(value)}
		if value != "" {
			path := value
			if name == "WASI_SYSROOT" {
				path = filepath.Dir(filepath.Dir(path))
			}
			if info, err := os.Stat(path); err == nil {
				record.Kind = "file"
				if info.IsDir() {
					record.Kind = "directory"
				}
				record.Digest, err = inputDigest(path)
				if err != nil {
					return nil, fmt.Errorf("input %s: %w", name, err)
				}
			} else if !os.IsNotExist(err) {
				return nil, fmt.Errorf("input %s: %w", name, err)
			}
		}
		result[name] = record
	}
	return result, nil
}

func makeFrozenPlan(count int) (plan, error) {
	p, err := makePlan(count)
	if err != nil {
		return p, err
	}
	return freezePlan(p)
}

func freezePlan(p plan) (plan, error) {
	identity := &frozenIdentity{Version: 1}
	var err error
	identity.SourceFiles, err = sourceRecords()
	if err != nil {
		return p, err
	}
	identity.Tools, err = toolIdentity()
	if err != nil {
		return p, err
	}
	identity.GoEnvironment, err = planningGoEnvironment()
	if err != nil {
		return p, err
	}
	identity.Inputs, err = planningInputs(p, -1)
	if err != nil {
		return p, err
	}
	if identity.Tools["go"] != p.GoVersion || identity.Tools["node"] != p.NodeVersion {
		return p, errors.New("toolchain changed during planning")
	}
	source, err := sourceIdentity()
	if err != nil {
		return p, err
	}
	if source != p.Source {
		return p, errors.New("source changed during planning")
	}
	p.Frozen = identity
	p.Digest = planDigest(p)
	return p, nil
}

func consumePlan(path string, count, index int) (plan, error) {
	if path == "" {
		return makePlan(count)
	}
	var p plan
	if err := loadJSON(path, &p); err != nil {
		return p, err
	}
	if err := validateFrozenPlan(p, count, index); err != nil {
		return p, err
	}
	return p, nil
}

func validateFrozenPlan(p plan, count, index int) error {
	if p.Version != 1 || p.Count < 1 || p.Count != count || !fullCommit.MatchString(p.Commit) || p.Frozen == nil || p.Frozen.Version != 1 {
		return errors.New("invalid frozen plan version, count or identity")
	}
	if p.Digest == "" || p.Digest != planDigest(p) {
		return errors.New("frozen plan digest differs")
	}
	if p.GoVersion != p.Frozen.Tools["go"] || p.NodeVersion != p.Frozen.Tools["node"] {
		return errors.New("frozen plan go/node version metadata differs")
	}
	commit, err := output("git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if commit != p.Commit {
		return fmt.Errorf("frozen plan commit differs: checkout %s, plan %s", commit, p.Commit)
	}
	actual, err := sourceRecords()
	if err != nil {
		return err
	}
	expectedFiles := map[string]sourceRecord{}
	for _, record := range p.Frozen.SourceFiles {
		expectedFiles[record.Name] = record
	}
	for _, record := range actual {
		if record != expectedFiles[record.Name] {
			return fmt.Errorf("frozen plan source %s differs", record.Name)
		}
		delete(expectedFiles, record.Name)
	}
	if len(expectedFiles) > 0 {
		names := []string{}
		for name := range expectedFiles {
			names = append(names, name)
		}
		sort.Strings(names)
		return fmt.Errorf("frozen plan source %s is missing", names[0])
	}
	status, err := output("git", "status", "--porcelain", "--untracked-files=normal")
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("frozen plan requires clean checkout: %s", status)
	}
	source, err := sourceIdentity()
	if err != nil {
		return err
	}
	if source != p.Source {
		return errors.New("frozen plan source digest differs")
	}
	modules, err := inspectSubmodules(".", p.Commit)
	if err != nil {
		return err
	}
	if err := requirePinnedSubmodules(modules); err != nil {
		return err
	}
	tools, err := toolIdentity()
	if err != nil {
		return err
	}
	for _, name := range []string{"go", "go-binary", "clang", "clang-binary", "clang-resources", "node", "node-binary"} {
		if tools[name] != p.Frozen.Tools[name] {
			return fmt.Errorf("frozen plan tool %s differs", name)
		}
	}
	environment, err := planningGoEnvironment()
	if err != nil {
		return err
	}
	for name, value := range environment {
		if value != p.Frozen.GoEnvironment[name] {
			return fmt.Errorf("frozen plan Go input %s differs", name)
		}
	}
	if !reflect.DeepEqual(environment, p.Frozen.GoEnvironment) {
		return errors.New("frozen plan Go input inventory differs")
	}
	inputs, err := planningInputs(p, index)
	if err != nil {
		return err
	}
	expected := map[string]inputRecord{}
	for name, value := range p.Frozen.Inputs {
		if name == archiveVariable && index >= 0 && (p.Archive == nil || p.Archive.Shard != index) {
			continue
		}
		expected[name] = value
	}
	for name, value := range expected {
		if inputs[name] != value {
			return fmt.Errorf("frozen plan input %s differs", name)
		}
	}
	for name := range inputs {
		if _, ok := expected[name]; !ok {
			return fmt.Errorf("frozen plan unexpected input %s", name)
		}
	}
	if err := validateAffinity(p); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, u := range p.Units {
		if u.Shard < 0 || u.Shard >= p.Count || seen[u.key()] {
			return fmt.Errorf("invalid frozen unit %s", u.key())
		}
		seen[u.key()] = true
	}
	for _, c := range p.Complements {
		if c.Shard < 0 || c.Shard >= p.Count {
			return fmt.Errorf("invalid frozen complement %s", c.Parent)
		}
	}
	if len(p.Shards) != p.Count {
		return errors.New("frozen plan shard predictions incomplete")
	}
	for index, prediction := range p.Shards {
		if prediction.Shard != index {
			return errors.New("frozen plan shard prediction index differs")
		}
		packages := assignedPackages(p, index)
		names := []string{}
		for pkg := range packages {
			names = append(names, pkg)
		}
		sort.Strings(names)
		total := 0.0
		for _, pkg := range names {
			seconds, ok := p.PackageSeconds[packagePredictionKey(index, pkg)]
			if !ok || seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
				return fmt.Errorf("frozen plan package prediction %s missing or invalid", pkg)
			}
			total += seconds
		}
		if math.Abs(total-prediction.Seconds) > 1e-7*math.Max(1, math.Abs(prediction.Seconds)) {
			return fmt.Errorf("frozen plan shard %d package predictions differ", index)
		}
	}
	return nil
}
