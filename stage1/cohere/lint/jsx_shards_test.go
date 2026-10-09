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
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
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

// Each build has its inputs beside its closure. Non-Go products and the
// completed setup use internal/buildcache; overlay Go builds are shared within
// the setup child because buildcache refuses overlay builds.
type jsxProductInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

// Include source dependencies as files, not just the entry point, so products
// invalidate when an imported source or tool changes.
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
	rows := jsxUpstream(t)
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
	t.Parallel()
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

// Every consumer prepares the immutable capture, build, and inventory product
// once per process, before starting its own case deadline.
type jsxTreeBundle struct {
	Paths  []string
	Shards [][]string
}

var jsxReadyDirectory string
var jsxPrepareOnce sync.Once

func jsxSetupInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "jsx-tree-setup-v1",
		Files:     []string{"cohere", "stage1/cohere/lint", "stage1/typescript/parser", "internal", "oracle", "go.mod", "go.work"},
		Flags:     []string{"whole", "jsx-recovery", "sanitize=address,undefined", "shards=16", "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH, "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "CC=" + os.Getenv("CC"), "CGO_CFLAGS=" + os.Getenv("CGO_CFLAGS"), "CGO_LDFLAGS=" + os.Getenv("CGO_LDFLAGS")},
		Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version"), buildcache.Tool("node", "--version")},
	}
}

func TestJsxLintTrees_Setup(t *testing.T) {
	t.Parallel()
	jsxPrepareTrees(t)
}

func jsxPrepareTrees(t *testing.T) string {
	t.Helper()
	jsxPrepareOnce.Do(func() {
		jsxReadyDirectory = buildcache.Product(t, jsxSetupInputs(), func(directory string) error {
			paths := jsxTreeSources(t)
			shards, err := jsxTreeShards(paths, jsxStableCaseKeys(t, paths))
			if err != nil {
				return err
			}
			root, _ := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
			side, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go"))
			oracle := jsxGoOracle(t, "jsx-parser", root, side, "adamic_jsx_oracle.go")
			parserDirectory, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser"))
			binary := jsxTreeNative(t, filepath.Join(parserDirectory, "main.ts"), true)
			copied := map[string]string{}
			for _, path := range paths {
				components := strings.Split(filepath.ToSlash(path), "/")
				relative := ""
				for i, component := range components {
					if strings.HasPrefix(component, "case-") {
						relative = filepath.Join(append([]string{"cases"}, components[i:]...)...)
						break
					}
				}
				if relative == "" {
					return fmt.Errorf("case has no capture directory: %s", path)
				}
				target := filepath.Join(directory, relative)
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return err
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if err := os.WriteFile(target, data, 0444); err != nil {
					return err
				}
				copied[path] = relative
			}
			bundle := jsxTreeBundle{Shards: make([][]string, len(shards))}
			for _, path := range paths {
				bundle.Paths = append(bundle.Paths, copied[path])
			}
			for i, cases := range shards {
				for _, path := range cases {
					bundle.Shards[i] = append(bundle.Shards[i], copied[path])
				}
			}
			for name, source := range map[string]string{"parser-oracle": oracle, "native": binary} {
				data, err := os.ReadFile(source)
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(directory, name), data, 0555); err != nil {
					return err
				}
			}
			data, err := json.Marshal(bundle)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "bundle.json"), data, 0444)
		})
	})
	// A failed builder may have called Fatal inside Once; do not let another
	// parallel consumer mistake its empty result for a prepared product.
	if jsxReadyDirectory == "" {
		t.Fatal("shared JSX preparation failed")
	}
	return jsxReadyDirectory
}

func jsxFetchTrees(t *testing.T) (jsxTreeBundle, string) {
	t.Helper()
	directory := jsxPrepareTrees(t)
	data, err := os.ReadFile(filepath.Join(directory, "bundle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var bundle jsxTreeBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	if err := jsxValidateUnion(bundle.Paths, bundle.Shards); err != nil {
		t.Fatal(err)
	}
	for i, path := range bundle.Paths {
		bundle.Paths[i] = filepath.Join(directory, path)
	}
	for i, cases := range bundle.Shards {
		for j, path := range cases {
			bundle.Shards[i][j] = filepath.Join(directory, path)
		}
	}
	return bundle, directory
}

func jsxTreeSources(t *testing.T) []string {
	t.Helper()
	rows := jsxUpstream(t)
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
	bundle, _ := jsxFetchTrees(t)
	shards := bundle.Shards
	var paths []string
	for _, cases := range shards {
		paths = append(paths, cases...)
	}
	if err := jsxValidateUnion(bundle.Paths, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("live union: %d cases", len(paths))
}

func jsxRunTreeShard(t *testing.T, i int) {
	t.Helper()
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
	bundle, products := jsxFetchTrees(t)
	oracle, binary := filepath.Join(products, "parser-oracle"), filepath.Join(products, "native")
	directory, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser"))
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	cases := bundle.Shards[i]
	if len(cases) == 0 {
		return
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	path := manifest(t, cases)
	want := jsxCaseExecute(t, ctx, "", oracle, "--manifest", path, "--whole", "--jsx-recovery")
	node := jsxCaseExecute(t, ctx, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--whole")
	jsxCheckTree(t, "Node", node.output, want.output)
	got := jsxCaseExecute(t, ctx, "", binary, "--manifest", path, "--whole")
	if os.Getenv("ADAMIC_JSX_TREE_DISAGREEMENT") == "1" && i == 3 {
		got.output = jsxPlantDisagreement(got.output)
	}
	jsxCheckTree(t, "native", got.output, want.output)
	if time.Since(started) > 60*time.Second {
		t.Fatalf("cooked: shard-%03d exceeded its 60-second case budget", i)
	}
	t.Logf("%d cases, logic wall %s, %d identical whole-tree bytes", len(cases), time.Since(started), len(want.output))
}

func TestJsxLintTreesShardDisagreement(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := jsxDeadlineCommand(t, executable, "-test.run=^TestJsxLintTrees_[0-9]{3}$", "-test.v", "-test.timeout=90s")
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
	return jsxContextCommand(t, context.Background(), name, args...)
}

func jsxContextCommand(t *testing.T, parent context.Context, name string, args ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
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

func jsxCaseExecute(t *testing.T, ctx context.Context, directory, name string, args ...string) execution {
	t.Helper()
	command := jsxContextCommand(t, ctx, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	if ctx.Err() != nil {
		t.Fatalf("cooked: case deadline exceeded: %s %v", name, args)
	}
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, time.Since(started)}
}

func jsxUpstream(t *testing.T) []string {
	t.Helper()
	value := shared("jsx-setup-upstream", func(value *sharedValue) {
		directory, err := os.MkdirTemp(sharedDirectory, "jsx-upstream-")
		if err != nil {
			value.err = err
			return
		}
		value.rows, value.err = jsxCaptureUpstream(".", directory)
	})
	if value.err != nil {
		t.Fatal(value.err)
	}
	return append([]string(nil), value.rows...)
}

func jsxCaptureRun(directory string, environment []string, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	if directory != "" {
		command.Env = append(command.Env, "PWD="+directory)
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	output, err := os.CreateTemp(sharedDirectory, "jsx-capture-stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

// Capture uses the package's complete upstream protocol, with an owned runner
// so all spawned Go compilers inherit the setup context and group cancellation.
func jsxCaptureUpstream(sourceRoot, directory string) ([]string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, err
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, err
	}
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, err
	}
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := jsxCaptureRun(root, environment, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=90s"); err != nil {
			return nil, err
		}
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, err
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			return nil, err
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				return nil, err
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}
