package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Each build has its inputs beside its closure. Until internal/buildcache is
// available on this base, the existing harness shares it within the run. This
// helper never creates a persistent package cache.
type jsxProductInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

// Include source dependencies as files, not just the entry point. A future
// content-keyed Product must invalidate when an imported source or tool changes.
func jsxInputFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	files := map[string]bool{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(path) {
			case ".go", ".ts", ".a", ".c", ".h", ".mod", ".sum", ".work", ".json":
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				files[absolute] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	result := make([]string, 0, len(files))
	for path := range files {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func jsxOverlayProduct(t *testing.T, inputs jsxProductInputs, build func(dir string) error) string {
	t.Helper()
	value := shared(fmt.Sprintf("jsx product %+v", inputs), func(value *sharedValue) {
		value.path, value.err = os.MkdirTemp(sharedDirectory, "jsx-product-")
		if value.err != nil {
			return
		}
		started := time.Now()
		value.err = build(value.path)
		t.Logf("build %s cold wall: %s", inputs.Name, time.Since(started))
	})
	if value.err != nil {
		t.Fatal(value.err)
	}
	return value.path
}

func jsxGoOracle(t *testing.T, name, root, side, virtualName string) string {
	t.Helper()
	inputs := jsxProductInputs{Name: name, Files: jsxInputFiles(t, side, filepath.Join(repository, "cohere")), Flags: []string{"build", "-overlay"}, Toolchain: runtime.Version()}
	directory := jsxOverlayProduct(t, inputs, func(dir string) error {
		virtual := filepath.Join(root, virtualName)
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		command := jsxDeadlineCommand(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = root
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("build %s: %w\n%s\n%s", name, err, &stdout, &stderr)
		}
		if diagnostics := commandDiagnostics("go", stderr.Bytes()); len(diagnostics) != 0 {
			return fmt.Errorf("build %s: %s", name, diagnostics)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

// The registry can grow without cohere's pin moving. Capture ordinals and
// temporary directories are transport paths, never identities. Locate each
// upstream test's repository-relative file and fingerprint its own fixture/mode.
func jsxStableCaseKeys(t *testing.T, paths []string) map[string]string {
	t.Helper()
	selected := map[string]bool{}
	for _, path := range paths {
		selected[path] = true
	}
	rows := upstream(t)
	requiredRules := map[string]bool{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 {
			t.Fatalf("malformed capture row %q", row)
		}
		if selected[fields[0]] {
			requiredRules[fields[1]] = true
		}
	}
	descriptors := prepareRegistry(t, ".")
	origins := map[string]string{}
	packages := map[string]map[string]string{}
	for _, d := range descriptors {
		if !requiredRules[d.Name] {
			continue
		}
		if packages[d.UpstreamPackage] == nil {
			declarations := map[string]string{}
			relative := filepath.Join("cohere/internal/lint/rules", d.UpstreamPackage)
			files, err := filepath.Glob(filepath.Join(repository, relative, "*_test.go"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				parsed, err := goparser.ParseFile(token.NewFileSet(), file, nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				for _, declaration := range parsed.Decls {
					fn, ok := declaration.(*ast.FuncDecl)
					if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
						continue
					}
					if _, exists := declarations[fn.Name.Name]; exists {
						t.Fatalf("ambiguous upstream test %s", fn.Name.Name)
					}
					declarations[fn.Name.Name] = filepath.ToSlash(filepath.Join(relative, filepath.Base(file)))
				}
			}
			packages[d.UpstreamPackage] = declarations
		}
		matches := map[string]bool{}
		for name, file := range packages[d.UpstreamPackage] {
			if strings.HasPrefix(name, d.UpstreamTest) {
				matches[file] = true
			}
		}
		if len(matches) != 1 {
			t.Fatalf("upstream prefix %s/%s must identify one fixture file, got %v", d.UpstreamPackage, d.UpstreamTest, matches)
		}
		origin := ""
		for file := range matches {
			origin = file
		}
		if origin == "" {
			t.Fatalf("upstream test %s/%s has no source file", d.UpstreamPackage, d.UpstreamTest)
		}
		origins[d.Name] = origin + "\x00" + d.UpstreamTest
	}
	keys := map[string]string{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 {
			t.Fatalf("malformed capture row %q", row)
		}
		path := fields[0]
		if !selected[path] {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		components := strings.Split(filepath.ToSlash(path), "/")
		fixture := ""
		for i, component := range components {
			if strings.HasPrefix(component, "case-") {
				if _, err := strconv.Atoi(strings.TrimPrefix(component, "case-")); err == nil {
					fixture = strings.Join(components[i+1:], "/")
					break
				}
			}
		}
		if fixture == "" {
			t.Fatalf("capture path has no fixture name: %s", path)
		}
		mode := append(append([]string(nil), fields[1:]...), "whole", "jsx-recovery")
		identity, err := json.Marshal([]any{origins[fields[1]], fixture, string(source), mode})
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(identity)
		keys[path] = origins[fields[1]] + "\x00" + fixture + "\x00" + fmt.Sprintf("%x", sum)
	}
	if len(keys) != len(paths) {
		t.Fatalf("stable keys cover %d cases, want live enumeration %d", len(keys), len(paths))
	}
	return keys
}

func jsxCaseShard(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(testJsxLintTreesShards))
}

func jsxTreeShards(paths []string, identities ...map[string]string) ([][]string, error) {
	ordered := append([]string(nil), paths...)
	sort.Strings(ordered)
	shards := make([][]string, testJsxLintTreesShards)
	for _, path := range ordered {
		key := path // Helper-test fixtures are already repository-relative identities.
		if len(identities) > 0 {
			var exists bool
			key, exists = identities[0][path]
			if !exists || key == "" {
				return nil, fmt.Errorf("missing stable case key %q", path)
			}
		} else if filepath.IsAbs(path) || strings.HasPrefix(filepath.Clean(path), "..") {
			return nil, fmt.Errorf("case %q needs a repository-relative stable identity", path)
		}
		shard := jsxCaseShard(key)
		shards[shard] = append(shards[shard], path)
	}
	if err := jsxValidateUnion(paths, shards); err != nil {
		return nil, err
	}
	return shards, nil
}

func jsxValidateUnion(paths []string, shards [][]string) error {
	if len(shards) != testJsxLintTreesShards {
		return fmt.Errorf("enumerated %d shards, want %d", len(shards), testJsxLintTreesShards)
	}
	if len(paths) == 0 {
		return fmt.Errorf("JSX corpus is empty")
	}
	expected := map[string]bool{}
	for _, path := range paths {
		if path == "" || expected[path] {
			return fmt.Errorf("empty or repeated unsplit case id %q", path)
		}
		expected[path] = true
	}
	seen := map[string]bool{}
	count := 0
	for i, cases := range shards {
		for _, path := range cases {
			if !expected[path] {
				return fmt.Errorf("shard-%03d has unknown case id %q", i, path)
			}
			if seen[path] {
				return fmt.Errorf("shard-%03d repeats case id %q", i, path)
			}
			seen[path] = true
			count++
		}
	}
	if count != len(paths) || len(seen) != len(expected) {
		return fmt.Errorf("shard union has %d cases and %d ids, want %d", count, len(seen), len(paths))
	}
	for path := range expected {
		if !seen[path] {
			return fmt.Errorf("shard union missing case id %q", path)
		}
	}
	return nil
}

func jsxShardSelection(value string) (func(int) bool, error) {
	if value == "" {
		return func(int) bool { return true }, nil
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q: want i/n", value)
	}
	index, first := strconv.Atoi(fields[0])
	count, second := strconv.Atoi(fields[1])
	if first != nil || second != nil || count <= 0 || count > testJsxLintTreesShards || index < 0 || index >= count {
		return nil, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return func(shard int) bool { return shard%count == index }, nil
}

func jsxPlantDisagreement(output []byte) []byte {
	// Keep case numbering and change only the first case's tree.
	newline := bytes.IndexByte(output, '\n')
	if newline < 0 {
		panic("tree output has no case header")
	}
	result := append([]byte(nil), output[:newline+1]...)
	result = append(result, []byte("planted tree disagreement\n")...)
	return append(result, output[newline+1:]...)
}

func jsxCheckTree(t *testing.T, side string, got, want []byte) {
	t.Helper()
	if diff := difference(got, want); diff != "" {
		t.Fatalf("%s: %s", side, diff)
	}
}

func TestJsxLintTreesShardCoverage(t *testing.T) {
	paths := make([]string, 8*testJsxLintTreesShards+3)
	for i := range paths {
		paths[i] = fmt.Sprintf("case-%03d.tsx", i)
	}
	shards, err := jsxTreeShards(paths)
	if err != nil {
		t.Fatal(err)
	}
	owners := func(groups [][]string) map[string]int {
		result := map[string]int{}
		for shard, cases := range groups {
			for _, path := range cases {
				result[path] = shard
			}
		}
		return result
	}
	original := owners(shards)
	expanded, err := jsxTreeShards(append([]string{"000-new-case.tsx"}, paths...))
	if err != nil {
		t.Fatal(err)
	}
	for path, shard := range original {
		if owners(expanded)[path] != shard {
			t.Fatalf("adding a case moved %s", path)
		}
	}
	relocated, err := jsxTreeShards([]string{"/new-temp/case-999/fixture.tsx"}, map[string]string{"/new-temp/case-999/fixture.tsx": paths[0]})
	if err != nil {
		t.Fatal(err)
	}
	if owners(relocated)["/new-temp/case-999/fixture.tsx"] != original[paths[0]] {
		t.Fatal("capture renumbering moved a stable identity")
	}
	if _, err := jsxTreeShards(nil); err == nil {
		t.Fatal("empty corpus accepted")
	}

	for _, fault := range []string{"missing", "repeated", "unknown", "shard count"} {
		t.Run(fault, func(t *testing.T) {
			broken := make([][]string, len(shards))
			for i := range shards {
				broken[i] = append([]string(nil), shards[i]...)
			}
			switch fault {
			case "missing":
				broken[0] = broken[0][1:]
			case "repeated":
				broken[0] = append(broken[0], broken[1][0])
			case "unknown":
				broken[0][0] = "unknown.tsx"
			case "shard count":
				broken = broken[:len(broken)-1]
			}
			if err := jsxValidateUnion(paths, broken); err == nil {
				t.Fatalf("%s union fault survived", fault)
			}
		})
	}
	for n := 1; n <= testJsxLintTreesShards; n++ {
		counts := make([]int, len(shards))
		for i := 0; i < n; i++ {
			selectShard, err := jsxShardSelection(fmt.Sprintf("%d/%d", i, n))
			if err != nil {
				t.Fatal(err)
			}
			for shard := range shards {
				if selectShard(shard) {
					counts[shard]++
				}
			}
		}
		for shard, count := range counts {
			if count != 1 {
				t.Fatalf("%d boxes select shard-%03d %d times", n, shard, count)
			}
		}
	}
	for _, invalid := range []string{"1", "-1/16", "16/16", "0/0", "0/17", "x/16", "0/16/1"} {
		if _, err := jsxShardSelection(invalid); err == nil {
			t.Fatalf("invalid selection %q accepted", invalid)
		}
	}
}

// Products are read-only inputs. Overlay Go builds are run-local because
// buildcache refuses overlays; non-Go products use the shared build cache.
func jsxProduct(t *testing.T, inputs jsxProductInputs, build func(string) error) string {
	t.Helper()
	files := make([]string, len(inputs.Files))
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for i, file := range inputs.Files {
		absolute, pathErr := filepath.Abs(file)
		if pathErr != nil {
			t.Fatal(pathErr)
		}
		files[i], err = filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
	}
	return buildcache.Product(t, buildcache.Inputs{Name: inputs.Name, Files: files, Flags: inputs.Flags, Toolchain: []string{inputs.Toolchain}}, build)
}
func jsxTreeNative(t *testing.T, entry string, sanitize bool) string {
	t.Helper()
	inputs := jsxProductInputs{Name: "jsx-tree-lowered", Files: append(jsxInputFiles(t, entry, filepath.Join(repository, "stage1"), filepath.Join(repository, "internal"), filepath.Join(repository, "cohere")), filepath.Join(packageDirectory, "jsx_shards_test.go"), filepath.Join(repository, "go.mod"), filepath.Join(repository, "go.work")), Toolchain: runtime.Version()}
	lowered := jsxProduct(t, inputs, func(dir string) error {
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(lowered)), 0644)
	})
	source := filepath.Join(lowered, "program.c")
	inputs = jsxProductInputs{Name: "jsx-tree-native", Files: append(append([]string(nil), inputs.Files...), jsxInputFiles(t, filepath.Join(repository, "internal/native/runtime"))...), Flags: native.Flags(native.Options{Sanitize: sanitize}), Toolchain: string(execute(t, "", "clang", "--version").output)}
	built := jsxProduct(t, inputs, func(dir string) error {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})
	})
	return filepath.Join(built, "native")
}

var jsxTreeState struct {
	sync.Mutex
	shards                            [][]string
	oracle, binary, directory, runner string
}

// Capture ordinals are not identities. Stable keys hash the repository-relative
// upstream file and the fixture's source/mode, so corpus additions cannot move
// existing cases. The union is always checked against the live enumeration.
func jsxPrepareTrees(t *testing.T, products bool) ([][]string, string, string, string, string) {
	t.Helper()
	jsxTreeState.Lock()
	defer jsxTreeState.Unlock()
	if jsxTreeState.shards == nil {
		paths := jsxTreeSources(t)
		shards, err := jsxTreeShards(paths, jsxStableCaseKeys(t, paths))
		if err != nil {
			t.Fatal(err)
		}
		jsxTreeState.shards = shards
		t.Logf("live union: %d cases in %d shards", len(paths), len(shards))
	}
	if products && jsxTreeState.binary == "" {
		root, _ := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
		side, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go"))
		jsxTreeState.oracle = jsxGoOracle(t, "jsx-parser", root, side, "adamic_jsx_oracle.go")
		jsxTreeState.directory, _ = filepath.Abs(filepath.Join(repository, "stage1/typescript/parser"))
		jsxTreeState.runner, _ = filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		jsxTreeState.binary = jsxTreeNative(t, filepath.Join(jsxTreeState.directory, "main.ts"), true)
	}
	return jsxTreeState.shards, jsxTreeState.oracle, jsxTreeState.binary, jsxTreeState.directory, jsxTreeState.runner
}

func jsxTreeSources(t *testing.T) []string {
	t.Helper()
	rows := upstream(t)
	all := make([]string, len(rows))
	for i, row := range rows {
		all[i] = strings.Split(row, "\t")[0]
	}
	root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	side, _ := filepath.Abs("testdata/jsx_inventory.go")
	oracle := jsxGoOracle(t, "jsx-membership", root, side, "adamic_jsx_inventory.go")
	output := execute(t, "", oracle, manifest(t, all)).output
	membership := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 2 || (fields[1] != "0" && fields[1] != "1") {
			t.Fatalf("malformed JSX membership %q", line)
		}
		if _, exists := membership[fields[0]]; exists {
			t.Fatalf("repeated JSX membership %q", fields[0])
		}
		membership[fields[0]] = fields[1] == "1"
	}
	if len(membership) != len(all) {
		t.Fatalf("membership count %d, want live %d", len(membership), len(all))
	}
	var paths []string
	for _, path := range all {
		jsx, exists := membership[path]
		if !exists {
			t.Fatalf("missing membership %q", path)
		}
		if jsx {
			paths = append(paths, path)
		}
	}
	return paths
}

func TestJsxLintTreesUnion(t *testing.T) {
	t.Parallel()
	// The planner sees declarations, so missing or extra entry points must fail.
	declarations, err := goparser.ParseFile(token.NewFileSet(), "jsx_integration_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]bool{}
	for _, declaration := range declarations.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "TestJsxLintTrees_") {
			entries[fn.Name.Name] = true
		}
	}
	if len(entries) != testJsxLintTreesShards {
		t.Fatalf("declared %d shards, want %d", len(entries), testJsxLintTreesShards)
	}
	for i := 0; i < testJsxLintTreesShards; i++ {
		name := fmt.Sprintf("TestJsxLintTrees_%03d", i)
		if !entries[name] {
			t.Fatalf("missing top-level shard %s", name)
		}
	}
	shards, _, _, _, _ := jsxPrepareTrees(t, false)
	var paths []string
	for _, cases := range shards {
		paths = append(paths, cases...)
	}
	if err := jsxValidateUnion(jsxTreeSources(t), shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("live union: %d cases", len(paths))
}

func jsxRunTreeShard(t *testing.T, i int) {
	t.Helper()
	t.Parallel()
	if os.Getenv("ADAMIC_JSX_SHARD_CHILD") == "1" {
		paths := make([]string, testJsxLintTreesShards)
		for j := range paths {
			paths[j] = fmt.Sprintf("case-%03d.tsx", j)
		}
		shards, err := jsxTreeShards(paths)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range shards[i] {
			want := []byte("case 0\nJsxElement " + path + "\n")
			got := want
			if path == "case-003.tsx" {
				got = jsxPlantDisagreement(got)
			}
			jsxCheckTree(t, "native", got, want)
		}
		return
	}
	selected, err := jsxShardSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	if !selected(i) {
		return
	}
	shards, oracle, binary, directory, runner := jsxPrepareTrees(t, true)
	cases := shards[i]
	if len(cases) == 0 {
		return
	}
	started := time.Now()
	path := manifest(t, cases)
	want := execute(t, "", oracle, "--manifest", path, "--whole", "--jsx-recovery")
	node := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--whole")
	jsxCheckTree(t, "Node", node.output, want.output)
	got := execute(t, "", binary, "--manifest", path, "--whole")
	if os.Getenv("ADAMIC_JSX_TREE_DISAGREEMENT") == "1" && i == 3 {
		got.output = jsxPlantDisagreement(got.output)
	}
	jsxCheckTree(t, "native", got.output, want.output)
	t.Logf("%d cases, logic wall %s, %d identical whole-tree bytes", len(cases), time.Since(started), len(want.output))
}

func TestJsxLintTreesShardDisagreement(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := jsxDeadlineCommand(t, executable, "-test.run=^TestJsxLintTrees_[0-9]{3}$", "-test.v", "-test.timeout=75s")
	command.Env = append(os.Environ(), "ADAMIC_JSX_SHARD_CHILD=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("planted disagreement survived:\n%s", output)
	}
	var failed []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestJsxLintTrees_") {
			failed = append(failed, strings.Fields(line)[2])
		}
	}
	expected := fmt.Sprintf("TestJsxLintTrees_%03d", jsxCaseShard("case-003.tsx"))
	if len(failed) != 1 || failed[0] != expected || !bytes.Contains(output, []byte("planted tree disagreement")) {
		t.Fatalf("want only %s, got %v:\n%s", expected, failed, output)
	}
	t.Logf("planted disagreement caught only by %s", expected)
}

// Cancel the entire child process group, including any spawned compilers.
// This uses Go's context deadline and needs no external timeout executable.
func jsxDeadlineCommand(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	return command
}
