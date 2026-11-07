package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
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

// Each build has its inputs beside its closure and uses internal/buildcache.
type jsxProductInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
}

// Include source dependencies as files, not just the entry point, so products
// invalidate when an imported source or tool changes.
func jsxInputFiles(test *testing.T, roots ...string) []string {
	test.Helper()
	files := map[string]bool{}
	for _, root := range roots {
		walkError := filepath.WalkDir(root, func(path string, entry fs.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" || lintRuleSubdirectory(path) {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(path) {
			case ".go", ".ts", ".a", ".c", ".h", ".mod", ".sum", ".work", ".json":
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				absolute, absolutePathError := filepath.Abs(path)
				if absolutePathError != nil {
					return absolutePathError
				}
				files[absolute] = true
			}
			return nil
		})
		if walkError != nil {
			test.Fatal(walkError)
		}
	}
	result := make([]string, 0, len(files))
	for path := range files {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

// The registry can grow without cohere's pin moving. Capture ordinals and
// temporary directories are transport paths, never identities. Locate each
// upstream test's repository-relative file and fingerprint its own fixture/mode.
func jsxStableCaseKeys(test *testing.T, paths []string) map[string]string {
	test.Helper()
	selected := map[string]bool{}
	for _, path := range paths {
		selected[path] = true
	}
	rows := jsxUpstream(test)
	requiredRules := map[string]bool{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 {
			test.Fatalf("malformed capture row %q", row)
		}
		if selected[fields[0]] {
			requiredRules[fields[1]] = true
		}
	}
	descriptors := prepareRegistry(test, ".")
	origins := map[string]string{}
	packages := map[string]map[string]string{}
	for _, d := range descriptors {
		if !requiredRules[d.Name] {
			continue
		}
		if packages[d.UpstreamPackage] == nil {
			declarations := map[string]string{}
			relative := filepath.Join("cohere/internal/lint/rules", d.UpstreamPackage)
			// Reads: cohere
			files, globError := filepath.Glob(filepath.Join(repository, relative, "*_test.go"))
			if globError != nil {
				test.Fatal(globError)
			}
			for _, file := range files {
				parsed, parseFileError := goparser.ParseFile(token.NewFileSet(), file, nil, 0)
				if parseFileError != nil {
					test.Fatal(parseFileError)
				}
				for _, declaration := range parsed.Decls {
					fn, ok := declaration.(*ast.FuncDecl)
					if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
						continue
					}
					if _, exists := declarations[fn.Name.Name]; exists {
						test.Fatalf("ambiguous upstream test %s", fn.Name.Name)
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
			test.Fatalf(
				"upstream prefix %s/%s must identify one fixture file, got %v",
				d.UpstreamPackage,
				d.UpstreamTest,
				matches,
			)
		}
		origin := ""
		for file := range matches {
			origin = file
		}
		if origin == "" {
			test.Fatalf("upstream test %s/%s has no source file", d.UpstreamPackage, d.UpstreamTest)
		}
		origins[d.Name] = origin + "\x00" + d.UpstreamTest
	}
	keys := map[string]string{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 {
			test.Fatalf("malformed capture row %q", row)
		}
		path := fields[0]
		if !selected[path] {
			continue
		}
		source, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		components := strings.Split(filepath.ToSlash(path), "/")
		fixture := ""
		for i, component := range components {
			if strings.HasPrefix(component, "case-") {
				if _, parseError := strconv.Atoi(strings.TrimPrefix(component, "case-")); parseError == nil {
					fixture = strings.Join(components[i+1:], "/")
					break
				}
			}
		}
		if fixture == "" {
			test.Fatalf("capture path has no fixture name: %s", path)
		}
		mode := append(append([]string(nil), fields[1:]...), "whole", "jsx-recovery")
		identity, encodeError := json.Marshal([]any{origins[fields[1]], fixture, string(source), mode})
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		sum := sha256.Sum256(identity)
		keys[path] = origins[fields[1]] + "\x00" + fixture + "\x00" + fmt.Sprintf("%x", sum)
	}
	if len(keys) != len(paths) {
		test.Fatalf("stable keys cover %d cases, want live enumeration %d", len(keys), len(paths))
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
	if jsxValidateUnionError := jsxValidateUnion(paths, shards); jsxValidateUnionError != nil {
		return nil, jsxValidateUnionError
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

func jsxCheckTree(test *testing.T, side string, got, want []byte) {
	test.Helper()
	if diff := difference(got, want); diff != "" {
		test.Fatalf("%s: %s", side, diff)
	}
}

func TestJsxLintTreesShardCoverage(test *testing.T) {
	test.Parallel()
	paths := make([]string, 8*testJsxLintTreesShards+3)
	for i := range paths {
		paths[i] = fmt.Sprintf("case-%03d.tsx", i)
	}
	shards, jsxTreeShardsError := jsxTreeShards(paths)
	if jsxTreeShardsError != nil {
		test.Fatal(jsxTreeShardsError)
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
	expanded, jsxTreeShardsError := jsxTreeShards(append([]string{"000-new-case.tsx"}, paths...))
	if jsxTreeShardsError != nil {
		test.Fatal(jsxTreeShardsError)
	}
	for path, shard := range original {
		if owners(expanded)[path] != shard {
			test.Fatalf("adding a case moved %s", path)
		}
	}
	relocated, jsxTreeShardsError := jsxTreeShards(
		[]string{"/new-temp/case-999/fixture.tsx"},
		map[string]string{"/new-temp/case-999/fixture.tsx": paths[0]},
	)
	if jsxTreeShardsError != nil {
		test.Fatal(jsxTreeShardsError)
	}
	if owners(relocated)["/new-temp/case-999/fixture.tsx"] != original[paths[0]] {
		test.Fatal("capture renumbering moved a stable identity")
	}
	if _, jsxTreeShardsError := jsxTreeShards(nil); jsxTreeShardsError == nil {
		test.Fatal("empty corpus accepted")
	}

	for _, fault := range []string{"missing", "repeated", "unknown", "shard count"} {
		test.Run(fault, func(test *testing.T) {
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
			if jsxValidateUnionError := jsxValidateUnion(paths, broken); jsxValidateUnionError == nil {
				test.Fatalf("%s union fault survived", fault)
			}
		})
	}
	for n := 1; n <= testJsxLintTreesShards; n++ {
		counts := make([]int, len(shards))
		for i := 0; i < n; i++ {
			selectShard, jsxShardSelectionError := jsxShardSelection(fmt.Sprintf("%d/%d", i, n))
			if jsxShardSelectionError != nil {
				test.Fatal(jsxShardSelectionError)
			}
			for shard := range shards {
				if selectShard(shard) {
					counts[shard]++
				}
			}
		}
		for shard, count := range counts {
			if count != 1 {
				test.Fatalf("%d boxes select shard-%03d %d times", n, shard, count)
			}
		}
	}
	for _, invalid := range []string{"1", "-1/16", "16/16", "0/0", "0/17", "x/16", "0/16/1"} {
		if _, jsxShardSelectionError := jsxShardSelection(invalid); jsxShardSelectionError == nil {
			test.Fatalf("invalid selection %q accepted", invalid)
		}
	}
}

// Products are read-only inputs, shared across product and shard processes.
func jsxProduct(test *testing.T, inputs jsxProductInputs, build func(string) error) string {
	test.Helper()
	files := make([]string, len(inputs.Files))
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for i, file := range inputs.Files {
		absolute, pathError := filepath.Abs(file)
		if pathError != nil {
			test.Fatal(pathError)
		}
		var relativePathError error
		files[i], relativePathError = filepath.Rel(root, absolute)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
	}
	return buildcache.Product(
		test,
		buildcache.Inputs{Name: inputs.Name, Files: files, Flags: inputs.Flags, Toolchain: inputs.Toolchain},
		build,
	)
}
func jsxTreeInputs(test *testing.T, entry string) jsxProductInputs {
	test.Helper()
	return jsxProductInputs{Name: "jsx-tree-lowered", Files: append(
		jsxInputFiles(
			test,
			entry,
			filepath.Join(repository, "stage1"),
			filepath.Join(repository, "internal"),
			filepath.Join(repository, "cohere"),
		),
		filepath.Join(packageDirectory, "jsx_shards_test.go"),
		filepath.Join(repository, "go.mod"),
		filepath.Join(repository, "go.work"),
	), Toolchain: []string{runtime.Version()}}
}

func jsxTreeLowered(test *testing.T, entry string) string {
	test.Helper()
	return jsxProduct(test, jsxTreeInputs(test, entry), func(dir string) error {
		program, loadError := load.Load([]string{entry})
		if loadError != nil {
			return loadError
		}
		lowered, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		return os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(lowered)), 0644)
	})
}

func jsxTreeNative(test *testing.T, entry string, sanitize bool) string {
	test.Helper()
	source := filepath.Join(jsxTreeLowered(test, entry), "program.c")
	inputs := jsxTreeInputs(test, entry)
	inputs.Name = "jsx-tree-native"
	inputs.Files = append(inputs.Files, jsxInputFiles(test, filepath.Join(repository, "internal/native/runtime"))...)
	// native.Build splits by ADAMIC_NATIVE_SPLIT when its options don't, so the switch is an input.
	inputs.Flags = append(
		native.Flags(native.Options{Sanitize: sanitize}),
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
	)
	inputs.Toolchain = []string{buildcache.Tool("clang", "--version")}
	built := jsxProduct(test, inputs, func(dir string) error {
		data, readError := os.ReadFile(source)
		if readError != nil {
			return readError
		}
		return native.Build(string(data), filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})
	})
	return filepath.Join(built, "native")
}

func jsxParserEntry(test *testing.T) string {
	test.Helper()
	entry, absolutePathError := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/main.ts"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return entry
}

func jsxParserOracle(test *testing.T) string { test.Helper(); return oracles.JSXParserOracle(test) }

func jsxMembershipOracle(test *testing.T) string {
	test.Helper()
	return oracles.JSXMembershipOracle(test)
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
	// The capture's go test and the oracles' go builds run toolchain's go under toolchain.Env, so the key reads go's
	// settings there (TestJsxSetupKeyNamesTheBuildsCompiler), never the caller's CC, CGO_* or GOFLAGS.
	return buildcache.Inputs{
		Name: "jsx-tree-setup-v1",
		Files: []string{
			"cohere",
			"stage1/cohere/lint",
			"stage1/typescript/parser",
			"internal",
			"oracle",
			"go.mod",
			"go.work",
		},
		Flags: []string{
			"whole",
			"jsx-recovery",
			"sanitize=address,undefined",
			"shards=16",
			"GOOS=" + runtime.GOOS,
			"GOARCH=" + runtime.GOARCH,
		},
		Toolchain: []string{
			runtime.Version(),
			buildcache.Tool("go", "version"),
			lintGoSettings(),
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("node", "--version"),
		},
	}
}

// TestJsxSetupKeyNamesTheBuildsCompiler holds the setup key to the compiler its go runs build with. With no CC,
// toolchain names clang, so the key must be the one an explicit CC=clang gives; another compiler must move it. Each
// answer comes from a child of this binary, since Tool reads a tool once per process.
func TestJsxSetupKeyNamesTheBuildsCompiler(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_JSX_SETUP_KEY_CHILD") == "1" {
		inputs := jsxSetupInputs()
		fmt.Printf(
			"jsx-setup-key\t%x\n",
			sha256.Sum256([]byte(strings.Join(append(inputs.Flags, inputs.Toolchain...), "\x00"))),
		)
		return
	}
	key := func(compiler string) string {
		var environment []string
		for _, setting := range os.Environ() {
			if !strings.HasPrefix(setting, "CC=") {
				environment = append(environment, setting)
			}
		}
		environment = append(environment, "ADAMIC_JSX_SETUP_KEY_CHILD=1")
		if compiler != "" {
			environment = append(environment, "CC="+compiler)
		}
		command, ctx := lintBuildPhaseChild(
			test,
			os.Args[0],
			"-test.run=^TestJsxSetupKeyNamesTheBuildsCompiler$",
			"-test.timeout=120s",
		)
		command.Env = environment
		stdout, stderr, captureError := childguard.Capture(ctx, command)
		output := append(stdout, stderr...)
		if captureError != nil {
			test.Fatalf("CC=%s: %v\n%s", compiler, captureError, output)
		}
		for _, line := range strings.Split(string(output), "\n") {
			if value, found := strings.CutPrefix(line, "jsx-setup-key\t"); found {
				return value
			}
		}
		test.Fatalf("CC=%s: the child printed no key:\n%s", compiler, output)
		return ""
	}
	unset, clang, gcc := key(""), key("clang"), key("gcc")
	if unset != clang {
		test.Errorf("with no CC the build runs clang, but the key (%s) isn't CC=clang's (%s)", unset, clang)
	}
	if gcc == clang {
		test.Errorf("CC=gcc keys as clang does (%s)", gcc)
	}
}

func TestJsxLintTrees_Setup(test *testing.T) {
	test.Parallel()
	jsxPrepareTrees(test)
}

func jsxPrepareTrees(test *testing.T) string {
	test.Helper()
	jsxPrepareOnce.Do(func() {
		jsxReadyDirectory = buildcache.Product(test, jsxSetupInputs(), func(directory string) error {
			paths := jsxTreeSources(test)
			shards, jsxTreeShardsError := jsxTreeShards(paths, jsxStableCaseKeys(test, paths))
			if jsxTreeShardsError != nil {
				return jsxTreeShardsError
			}
			oracle := jsxParserOracle(test)
			binary := jsxTreeNative(test, jsxParserEntry(test), true)
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
				if createDirectoryError := os.MkdirAll(filepath.Dir(target), 0755); createDirectoryError != nil {
					return createDirectoryError
				}
				data, readError := os.ReadFile(path)
				if readError != nil {
					return readError
				}
				if writeError := os.WriteFile(target, data, 0444); writeError != nil {
					return writeError
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
				data, readError := os.ReadFile(source)
				if readError != nil {
					return readError
				}
				if writeError := os.WriteFile(filepath.Join(directory, name), data, 0555); writeError != nil {
					return writeError
				}
			}
			data, encodeError := json.Marshal(bundle)
			if encodeError != nil {
				return encodeError
			}
			return os.WriteFile(filepath.Join(directory, "bundle.json"), data, 0444)
		})
	})
	// A failed builder may have called Fatal inside Once; do not let another
	// parallel consumer mistake its empty result for a prepared product.
	if jsxReadyDirectory == "" {
		test.Fatal("shared JSX preparation failed")
	}
	return jsxReadyDirectory
}

func jsxFetchTrees(test *testing.T) (jsxTreeBundle, string) {
	test.Helper()
	directory := jsxPrepareTrees(test)
	data, readError := os.ReadFile(filepath.Join(directory, "bundle.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var bundle jsxTreeBundle
	if decodeError := json.Unmarshal(data, &bundle); decodeError != nil {
		test.Fatal(decodeError)
	}
	if jsxValidateUnionError := jsxValidateUnion(bundle.Paths, bundle.Shards); jsxValidateUnionError != nil {
		test.Fatal(jsxValidateUnionError)
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

func jsxTreeSources(test *testing.T) []string {
	test.Helper()
	rows := jsxUpstream(test)
	all := make([]string, len(rows))
	for i, row := range rows {
		all[i] = strings.Split(row, "\t")[0]
	}
	oracle := jsxMembershipOracle(test)
	output := execute(test, "", oracle, manifest(test, all)).output
	membership := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 2 || (fields[1] != "0" && fields[1] != "1") {
			test.Fatalf("malformed JSX membership %q", line)
		}
		if _, exists := membership[fields[0]]; exists {
			test.Fatalf("repeated JSX membership %q", fields[0])
		}
		membership[fields[0]] = fields[1] == "1"
	}
	if len(membership) != len(all) {
		test.Fatalf("membership count %d, want live %d", len(membership), len(all))
	}
	var paths []string
	for _, path := range all {
		jsx, exists := membership[path]
		if !exists {
			test.Fatalf("missing membership %q", path)
		}
		if jsx {
			paths = append(paths, path)
		}
	}
	return paths
}

func TestJsxLintTreesUnion(test *testing.T) {
	test.Parallel()
	// The planner sees declarations, so missing or extra entry points must fail.
	declarations, parseFileError := goparser.ParseFile(token.NewFileSet(), "jsx_integration_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	entries := map[string]bool{}
	for _, declaration := range declarations.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "TestJsxLintTrees_") {
			entries[fn.Name.Name] = true
		}
	}
	if len(entries) != testJsxLintTreesShards {
		test.Fatalf("declared %d shards, want %d", len(entries), testJsxLintTreesShards)
	}
	for i := 0; i < testJsxLintTreesShards; i++ {
		name := fmt.Sprintf("TestJsxLintTrees_%03d", i)
		if !entries[name] {
			test.Fatalf("missing top-level shard %s", name)
		}
	}
	bundle, _ := jsxFetchTrees(test)
	shards := bundle.Shards
	var paths []string
	for _, cases := range shards {
		paths = append(paths, cases...)
	}
	if jsxValidateUnionError := jsxValidateUnion(bundle.Paths, shards); jsxValidateUnionError != nil {
		test.Fatal(jsxValidateUnionError)
	}
	test.Logf("live union: %d cases", len(paths))
}

func jsxRunTreeShard(test *testing.T, i int) {
	test.Helper()
	if os.Getenv("ADAMIC_JSX_SHARD_CHILD") == "1" {
		paths := make([]string, testJsxLintTreesShards)
		for j := range paths {
			paths[j] = fmt.Sprintf("case-%03d.tsx", j)
		}
		shards, jsxTreeShardsError := jsxTreeShards(paths)
		if jsxTreeShardsError != nil {
			test.Fatal(jsxTreeShardsError)
		}
		for _, path := range shards[i] {
			want := []byte("case 0\nJsxElement " + path + "\n")
			got := want
			if path == "case-003.tsx" {
				got = jsxPlantDisagreement(got)
			}
			jsxCheckTree(test, "native", got, want)
		}
		return
	}
	selected, jsxShardSelectionError := jsxShardSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if jsxShardSelectionError != nil {
		test.Fatal(jsxShardSelectionError)
	}
	if !selected(i) {
		return
	}
	bundle, products := jsxFetchTrees(test)
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
	path := manifest(test, cases)
	want := jsxCaseExecute(test, ctx, "", oracle, "--manifest", path, "--whole", "--jsx-recovery")
	node := jsxCaseExecute(
		test,
		ctx,
		"",
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(directory, "main.ts"),
		"--manifest",
		path,
		"--whole",
	)
	jsxCheckTree(test, "Node", node.output, want.output)
	got := jsxCaseExecute(test, ctx, "", binary, "--manifest", path, "--whole")
	if os.Getenv("ADAMIC_JSX_TREE_DISAGREEMENT") == "1" && i == 3 {
		got.output = jsxPlantDisagreement(got.output)
	}
	jsxCheckTree(test, "native", got.output, want.output)
	if time.Since(started) > 60*time.Second {
		test.Fatalf("cooked: shard-%03d exceeded its 60-second case budget", i)
	}
	test.Logf(
		"%d cases, logic wall %s, %d identical whole-tree bytes",
		len(cases),
		time.Since(started),
		len(want.output),
	)
}

func TestJsxLintTreesShardDisagreement(test *testing.T) {
	test.Parallel()
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	command, ctx := lintBuildPhaseChild(
		test,
		executable,
		"-test.run=^TestJsxLintTrees_[0-9]{3}$",
		"-test.v",
		"-test.timeout=600s",
	)
	command.Env = append(os.Environ(), "ADAMIC_JSX_SHARD_CHILD=1")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if captureError == nil {
		test.Fatalf("planted disagreement survived:\n%s", output)
	}
	var failed []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestJsxLintTrees_") {
			failed = append(failed, strings.Fields(line)[2])
		}
	}
	expected := fmt.Sprintf("TestJsxLintTrees_%03d", jsxCaseShard("case-003.tsx"))
	if len(failed) != 1 || failed[0] != expected || !bytes.Contains(output, []byte("planted tree disagreement")) {
		test.Fatalf("want only %s, got %v:\n%s", expected, failed, output)
	}
	test.Logf("planted disagreement caught only by %s", expected)
}

// jsxContextCommand is lintCommand under a case's own 90-second deadline: the entire child process group is
// cancelled with it, including any spawned compilers, with no external timeout executable. It adds no WaitDelay, for
// lintCommand's reason.
func jsxContextCommand(test *testing.T, parent context.Context, directory, name string, args ...string) *exec.Cmd {
	test.Helper()
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	test.Cleanup(cancel)
	return lintCommand(ctx, directory, name, args...)
}

func jsxCaseExecute(test *testing.T, ctx context.Context, directory, name string, args ...string) execution {
	test.Helper()
	command := jsxContextCommand(test, ctx, directory, name, args...)
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	runError := command.Run()
	if ctx.Err() != nil {
		test.Fatalf("cooked: case deadline exceeded: %s %v", name, args)
	}
	if runError != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return execution{data, time.Since(started)}
}

func jsxUpstream(test *testing.T) []string {
	test.Helper()
	value := shared("jsx-setup-upstream", func(value *sharedValue) {
		directory, temporaryDirectoryError := os.MkdirTemp(sharedDirectory, "jsx-upstream-")
		if temporaryDirectoryError != nil {
			value.err = temporaryDirectoryError
			return
		}
		value.rows, value.err = jsxCaptureUpstream(".", directory)
	})
	if value.err != nil {
		test.Fatal(value.err)
	}
	return append([]string(nil), value.rows...)
}

func jsxCaptureRun(directory string, environment []string, name string, args ...string) ([]byte, error) {
	// toolchain's program for name, under its environment plus the capture's own settings, as the setup key reads it.
	path, pathError := toolchain.Path(name)
	if pathError != nil {
		return nil, pathError
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := exec.CommandContext(ctx, path, args...)
	command.Dir = directory
	command.Env = toolchain.Env(environment...)
	if directory != "" {
		command.Env = append(command.Env, "PWD="+directory)
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	// No WaitDelay, for lintCommand's reason: stdout is a file, and exec reads stderr to its end. go test waits for
	// the compilers and test binaries it starts before it exits, so nothing holds the pipe after it, and the unit's
	// ceiling ends a wait on one that did.
	output, temporaryFileError := os.CreateTemp(sharedDirectory, "jsx-capture-stdout-")
	if temporaryFileError != nil {
		return nil, temporaryFileError
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if runError := command.Run(); runError != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	return os.ReadFile(output.Name())
}

// Capture uses the package's complete upstream protocol, with an owned runner
// so all spawned Go compilers inherit the setup context and group cancellation.
func jsxCaptureUpstream(sourceRoot, directory string) ([]string, error) {
	root, absolutePathError := filepath.Abs(filepath.Join(repository, "cohere"))
	if absolutePathError != nil {
		return nil, absolutePathError
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, readError := os.ReadFile(harness)
	if readError != nil {
		return nil, readError
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if writeError := os.WriteFile(
		side,
		[]byte(strings.Replace(string(data), original, replacement, 1)),
		0644,
	); writeError != nil {
		return nil, writeError
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
		return nil, writeError
	}
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, generateError := registry.Generate(sourceRoot)
	if generateError != nil {
		return nil, generateError
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
		if _, jsxCaptureRunError := jsxCaptureRun(
			root,
			environment,
			"go",
			"test",
			"-overlay="+overlayPath,
			"./internal/lint/rules/"+name,
			"-run",
			"^("+strings.Join(packages[name], "|")+")",
			"-count=1",
			"-timeout=90s",
		); jsxCaptureRunError != nil {
			return nil, jsxCaptureRunError
		}
	}
	files, globError := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if globError != nil {
		return nil, globError
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	// The raw capture is the upstream tests' own record, a file per process named by its pid, its lines in the
	// order parallel tests wrote them: never a product's bytes, which must be the same build to build (landable-15's
	// tree, Oct 11: lint-rules-agree-capture held capture-<pid>.jsonl). The rows and cases made from it are sorted.
	defer os.RemoveAll(capture)
	for _, path := range files {
		data, readError := os.ReadFile(path)
		if readError != nil {
			return nil, readError
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if decodeError := json.Unmarshal(line, &row); decodeError != nil {
				return nil, decodeError
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
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0755); createDirectoryError != nil {
			return nil, createDirectoryError
		}
		if writeError := os.WriteFile(path, []byte(row.Source), 0644); writeError != nil {
			return nil, writeError
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if decodeError := json.Unmarshal(row.Options, &legacy); decodeError != nil {
				return nil, decodeError
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
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" ||
			row.Source == "var a = /" ||
			row.Source == "var a = [/];" ||
			row.Source == "if (/) {}" ||
			row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf(
			"%s\t%s\t%s\t%s\t%t\t%s\t%s",
			path,
			row.Rule,
			legacy.Mode,
			legacy.Null,
			legacy.AllowEmptyCatch,
			string(row.Options),
			mode,
		))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}
