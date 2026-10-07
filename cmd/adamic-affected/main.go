// adamic-affected records inputs, never test results for reuse.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const formatVersion = 1

type packageInfo struct {
	ImportPath, Dir                                          string
	GoFiles, CgoFiles, EmbedFiles, TestGoFiles, XTestGoFiles []string
	CFiles, CXXFiles, HFiles, SFiles, SysoFiles              []string
	Module                                                   *struct{ GoMod string }
}
type closure struct {
	Static    map[string]string
	Observed  map[string]string
	Uncertain []string
	Events    string
}
type recordFile struct {
	Version   int
	Commit    string
	Toolchain map[string]string
	Submodule string
	Packages  map[string]closure
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "adamic-affected:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: adamic-affected record -out FILE | select -record FILE")
	}
	rootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return err
	}
	root := strings.TrimSpace(string(rootBytes))
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	out := flags.String("out", "", "record destination outside repository")
	input := flags.String("record", "", "green-main record")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	packages, err := list(root, "./...")
	if err != nil {
		return err
	}
	switch args[0] {
	case "record":
		if *out == "" {
			return fmt.Errorf("-out is required")
		}
		destination, err := filepath.Abs(*out)
		if err != nil {
			return err
		}
		if inside(root, destination) {
			return fmt.Errorf("record must live outside repository")
		}
		status, err := command(root, "git", "status", "--porcelain")
		if err != nil {
			return err
		}
		if len(status) != 0 {
			return fmt.Errorf("record requires a clean checkout")
		}
		commit, err := command(root, "git", "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		main, err := command(root, "git", "rev-parse", "origin/main")
		if err != nil {
			return err
		}
		if !bytes.Equal(commit, main) {
			return fmt.Errorf("record requires HEAD = origin/main")
		}
		tools, err := toolchain(root)
		if err != nil {
			return err
		}
		submodule, err := submoduleIdentity(root)
		if err != nil {
			return err
		}
		record := recordFile{formatVersion, strings.TrimSpace(string(commit)), tools, submodule, map[string]closure{}}
		initial := map[string]string{}
		if err := tree(root, root, initial); err != nil {
			return err
		}
		logs := destination + ".logs"
		if err := os.MkdirAll(logs, 0755); err != nil {
			return err
		}
		logs, err = os.MkdirTemp(logs, "run-")
		if err != nil {
			return err
		}
		for _, pkg := range packages {
			if pkg.ImportPath == "github.com/system-inc/adamic/internal/native" {
				if err := observerCompatible(logs); err != nil {
					return err
				}
				break
			}
		}
		// All package runs are uncached. Publish nothing until every package is green.
		for index, pkg := range packages {
			fmt.Fprintf(os.Stderr, "record %d/%d %s\n", index+1, len(packages), pkg.ImportPath)
			static, err := staticInputs(root, pkg.ImportPath)
			if err != nil {
				return err
			}
			stem := filepath.Join(logs, fmt.Sprintf("%03d", index))
			binary := stem + ".test"
			if err := logged(root, stem+".build.log", "go", "test", "-c", "-o", binary, pkg.ImportPath); err != nil {
				return err
			}
			value := closure{Static: static, Observed: map[string]string{}}
			dependencies, err := list(root, "-deps", "-test", pkg.ImportPath)
			if err != nil {
				return err
			}
			for _, dependency := range dependencies {
				if inside(root, dependency.Dir) && len(dependency.CgoFiles) > 0 {
					value.Uncertain = append(value.Uncertain, "repository cgo inputs may include headers outside Go's file list")
				}
			}
			if _, err := os.Stat(binary); os.IsNotExist(err) {
				record.Packages[pkg.ImportPath] = value
				continue
			} else if err != nil {
				return err
			}
			value.Events = stem + ".jsonl"
			if err := tracedRun(root, pkg, binary, stem, &value); err != nil {
				return err
			}
			for path, after := range value.Observed {
				before, present := initial[path]
				if !present {
					before = "missing"
				}
				if before != after {
					return fmt.Errorf("observed input changed during record: %s", path)
				}
			}
			// Reject inputs changed by tests or a concurrent editor.
			after, err := staticInputs(root, pkg.ImportPath)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(static, after) {
				return fmt.Errorf("static inputs changed during %s", pkg.ImportPath)
			}
			record.Packages[pkg.ImportPath] = value
		}
		after, err := toolchain(root)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(tools, after) {
			return fmt.Errorf("toolchain changed during record")
		}
		afterSubmodule, err := submoduleIdentity(root)
		if err != nil {
			return err
		}
		if submodule != afterSubmodule {
			return fmt.Errorf("submodule changed during record")
		}
		for name, value := range record.Packages {
			current, err := staticInputs(root, name)
			if err != nil || !reflect.DeepEqual(value.Static, current) || changed(root, value.Observed) {
				return fmt.Errorf("inputs changed before publishing record for %s", name)
			}
		}
		status, err = command(root, "git", "status", "--porcelain")
		if err != nil || len(status) != 0 {
			return fmt.Errorf("checkout changed during record")
		}
		encoded, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return err
		}
		temporary, err := os.CreateTemp(filepath.Dir(destination), ".affected-")
		if err != nil {
			return err
		}
		defer os.Remove(temporary.Name())
		if _, err := temporary.Write(encoded); err != nil {
			temporary.Close()
			return err
		}
		if err := temporary.Close(); err != nil {
			return err
		}
		return os.Rename(temporary.Name(), destination)
	case "select":
		if *input == "" {
			return fmt.Errorf("-record is required")
		}
		contents, readErr := os.ReadFile(*input)
		var record recordFile
		invalid := readErr != nil || json.Unmarshal(contents, &record) != nil || record.Version != formatVersion || record.Packages == nil
		tools, toolErr := toolchain(root)
		submodule, subErr := submoduleIdentity(root)
		all := invalid || toolErr != nil || subErr != nil || !reflect.DeepEqual(record.Toolchain, tools) || record.Submodule != submodule
		for _, pkg := range packages {
			value, known := record.Packages[pkg.ImportPath]
			selected := all || !known || len(value.Uncertain) > 0 || value.Static == nil || value.Observed == nil
			if !selected {
				current, err := staticInputs(root, pkg.ImportPath)
				selected = err != nil || !reflect.DeepEqual(value.Static, current) || changed(root, value.Observed)
			}
			if selected {
				fmt.Println(pkg.ImportPath)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown operation %q", args[0])
	}
}
func command(directory, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %v: %w: %s", name, args, err, stderr.String())
	}
	return output, nil
}
func logged(directory, path, name string, args ...string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	cmd.Stdout = file
	cmd.Stderr = file
	cmd.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w; see %s", name, err, path)
	}
	return nil
}
func list(root string, args ...string) ([]packageInfo, error) {
	output, err := command(root, "go", append([]string{"list", "-json"}, args...)...)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	var result []packageInfo
	for {
		var value packageInfo
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ImportPath < result[j].ImportPath })
	return result, nil
}
func digest(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
func inside(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
func inputPath(root, path string) string {
	if inside(root, path) {
		relative, _ := filepath.Rel(root, path)
		return relative
	}
	return path
}
func fingerprint(path string) (string, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "missing", nil
	}
	if err != nil {
		return "", err
	}
	var data bytes.Buffer
	fmt.Fprintf(&data, "%s\x00", info.Mode())
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&data, "link:%s\x00", target)
		resolved, err := filepath.EvalSymlinks(path)
		if os.IsNotExist(err) {
			data.WriteString("dangling")
		} else if err != nil {
			return "", err
		} else {
			value, err := fingerprint(resolved)
			if err != nil {
				return "", err
			}
			data.WriteString(value)
		}
	} else if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			fmt.Fprintf(&data, "%q:%s\x00", entry.Name(), entry.Type())
		}
	} else if info.Mode().IsRegular() {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		data.Write(contents)
	} else {
		return "", fmt.Errorf("unsupported input %s", path)
	}
	return digest(data.Bytes()), nil
}
func add(root, path string, inputs map[string]string) error {
	value, err := fingerprint(path)
	if err != nil {
		return err
	}
	inputs[inputPath(root, path)] = value
	return nil
}
func tree(root, path string, inputs map[string]string) error {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return add(root, path, inputs)
	}
	return filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		return add(root, path, inputs)
	})
}
func staticInputs(root, pkg string) (map[string]string, error) {
	dependencies, err := list(root, "-deps", "-test", pkg)
	if err != nil {
		return nil, err
	}
	inputs := map[string]string{}
	for _, dependency := range dependencies {
		if dependency.Module != nil && dependency.Module.GoMod != "" {
			if err := add(root, dependency.Module.GoMod, inputs); err != nil {
				return nil, err
			}
			if err := add(root, filepath.Join(filepath.Dir(dependency.Module.GoMod), "go.sum"), inputs); err != nil {
				return nil, err
			}
		}
		for _, files := range [][]string{dependency.GoFiles, dependency.CgoFiles, dependency.EmbedFiles, dependency.TestGoFiles, dependency.XTestGoFiles, dependency.CFiles, dependency.CXXFiles, dependency.HFiles, dependency.SFiles, dependency.SysoFiles} {
			for _, name := range files {
				path := name
				if !filepath.IsAbs(path) {
					path = filepath.Join(dependency.Dir, path)
				}
				if err := add(root, path, inputs); err != nil {
					return nil, err
				}
			}
		}
		if err := tree(root, filepath.Join(dependency.Dir, "testdata"), inputs); err != nil {
			return nil, err
		}
	}
	for _, name := range []string{"go.mod", "go.sum", "go.work", "go.work.sum", ".gitignore", ".gitmodules"} {
		if err := add(root, filepath.Join(root, name), inputs); err != nil {
			return nil, err
		}
	}
	workspace, err := command(root, "go", "env", "GOWORK")
	if err != nil {
		return nil, err
	}
	workspacePath := strings.TrimSpace(string(workspace))
	if workspacePath != "" && workspacePath != "off" {
		if err := add(root, workspacePath, inputs); err != nil {
			return nil, err
		}
		if err := add(root, workspacePath+".sum", inputs); err != nil {
			return nil, err
		}
	}
	return inputs, nil
}
func changed(root string, inputs map[string]string) bool {
	for path, before := range inputs {
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		after, err := fingerprint(path)
		if err != nil || after != before {
			return true
		}
	}
	return false
}
func toolchain(root string) (map[string]string, error) {
	values := map[string]string{}
	for _, probe := range [][]string{{"go", "version"}, {"clang", "--version"}, {"node", "--version"}, {"go", "env", "-json", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOEXPERIMENT", "GOTOOLCHAIN", "GOWORK", "GOMOD", "GOROOT", "CC", "CXX", "GOENV", "GOMODCACHE", "GOCACHE"}} {
		output, err := command(root, probe[0], probe[1:]...)
		if err != nil {
			return nil, err
		}
		values[strings.Join(probe, " ")] = string(output)
	}
	// Children inherit the environment, and tests use os.Environ as well as dynamic Getenv.
	// Keep only a digest in the artifact, never credential values.
	environment := os.Environ()
	sort.Strings(environment)
	values["environment SHA256"] = digest([]byte(strings.Join(environment, "\x00")))
	values["uid"] = fmt.Sprint(os.Getuid())
	return values, nil
}
func submoduleIdentity(root string) (string, error) {
	var data bytes.Buffer
	for _, args := range [][]string{{"-C", "cohere", "rev-parse", "HEAD"}, {"-C", "cohere", "diff", "HEAD", "--binary", "--submodule=diff"}, {"submodule", "status", "--recursive"}} {
		output, err := command(root, "git", args...)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&data, "%d:%s", len(output), output)
	}
	// Include untracked bytes, staged changes and nested worktrees, not just git diff's names.
	inputs := map[string]string{}
	if err := tree(root, filepath.Join(root, "cohere"), inputs); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}
	data.Write(encoded)
	return digest(data.Bytes()), nil
}
