package printer

import (
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/graphql/printer/oracles"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

type printerBuildInputs = oracles.Inputs

func printerBuild(test *testing.T, inputs printerBuildInputs, build func(dir string) error) string {
	test.Helper()
	return buildcache.Product(test, printerKey(test, inputs), build)
}

// printerKey is the buildcache inputs printerBuild keys a product by.
func printerKey(test *testing.T, inputs printerBuildInputs) buildcache.Inputs {
	return oracles.Key(test, inputs)
}

// printerGeneratedSums holds each generated input's SHA-256 by path, size and modification time: the 51 MB Go oracle
// keys all four corpora, and a process hashes it once rather than once per corpus, the four at once included.

func printerInputFiles(test *testing.T, roots ...string) []string {
	test.Helper()
	seen := map[string]bool{}
	tree, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, root := range roots {
		absolute, absolutePathError := filepath.Abs(root)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		files, printerWalkError := printerWalk(tree, absolute)
		if printerWalkError != nil {
			test.Fatal(printerWalkError)
		}
		for _, path := range files {
			seen[path] = true
		}
	}
	files := make([]string, 0, len(seen))
	for path := range seen {
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

// Two mutant source products reached through links, as a runner holds them, key distinct lowered products, and each
// keys the same through its link as through its directory, so a real directory's key (Workshop's) doesn't move.
func TestPrinterLoweredKeyFollowsALinkedSource(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	key := func(source string) string {
		test.Helper()
		return buildcache.NameKey(directory, printerKey(test, printerLoweredInputs(test, source)))
	}
	keys := map[string]string{}
	for _, mutant := range []string{"000", "001"} {
		product := filepath.Join(directory, "products", mutant)
		for _, file := range printerPortFiles {
			path := filepath.Join(product, filepath.FromSlash(file))
			if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if writeError := os.WriteFile(path, []byte(file+" of mutant "+mutant), 0o644); writeError != nil {
				test.Fatal(writeError)
			}
		}
		link := filepath.Join(directory, "links", mutant)
		if createDirectoryError := os.MkdirAll(filepath.Dir(link), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if symlinkError := os.Symlink(product, link); symlinkError != nil {
			test.Fatal(symlinkError)
		}
		keys[mutant] = key(filepath.Join(link, "graphql/printer/main.ts"))
		if direct := key(filepath.Join(product, "graphql/printer/main.ts")); direct != keys[mutant] {
			test.Fatalf(
				"mutant %s keys %s through its link and %s through its directory",
				mutant,
				keys[mutant][:12],
				direct[:12],
			)
		}
	}
	if keys["000"] == keys["001"] {
		test.Fatalf("mutants 000 and 001 key one lowered product %s through their links", keys["000"][:12])
	}
	test.Logf("through their links, mutant 000 keys %s and 001 keys %s", keys["000"][:12], keys["001"][:12])
}

// The package's own lowered printer keys the compiler and the port's sources, and nothing else of the tree: no
// directory beside them, so no listing of one, enters its key (#7zhyjz6).
func TestPrinterLoweredKeysOnlyTheCompilerAndThePort(test *testing.T) {
	test.Parallel()
	inputs := printerLoweredInputs(test, printerInTreeSource(test))
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := map[string]bool{}
	for _, input := range compilerinputs.Lowering() {
		want[input] = true
	}
	for _, file := range printerPortFiles {
		want["stage1/cohere/"+file] = true
	}
	for _, file := range printerKey(test, inputs).Files {
		if !want[file] {
			test.Errorf("the lowered printer keys %s, which is neither the compiler nor the port", file)
		}
		delete(want, file)
	}
	for file := range want {
		test.Errorf("the lowered printer doesn't key %s", file)
	}
	if test.Failed() {
		return
	}
	for _, file := range printerPortFiles {
		if _, statError := os.Stat(filepath.Join(root, "stage1/cohere", filepath.FromSlash(file))); statError != nil {
			test.Error(statError)
		}
	}
}

// printerPortFiles is exactly what main.ts imports, read from the sources' own import lines: a new import outside the
// list would be lowered without being keyed, and a file the list names that nothing imports keys what isn't read.
func TestPrinterPortFilesAreMainsImports(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(filepath.Join(repository, "stage1/cohere"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	importLine := regexp.MustCompile(`(?m)^\s*(?:import|export)\b[^;]*?\bfrom\s+'([^']+)'`)
	reached := map[string]bool{}
	pending := []string{"graphql/printer/main.ts"}
	for len(pending) > 0 {
		file := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if reached[file] {
			continue
		}
		reached[file] = true
		source, readError := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if readError != nil {
			test.Fatal(readError)
		}
		for _, match := range importLine.FindAllStringSubmatch(string(source), -1) {
			if !strings.HasPrefix(match[1], ".") {
				continue
			}
			pending = append(pending, filepath.ToSlash(filepath.Join(filepath.Dir(file), match[1])))
		}
	}
	listed := map[string]bool{}
	for _, file := range printerPortFiles {
		listed[file] = true
		if !reached[file] {
			test.Errorf("printerPortFiles names %s, which main.ts doesn't import", file)
		}
	}
	for file := range reached {
		if !listed[file] {
			test.Errorf("main.ts imports %s, which printerPortFiles doesn't name", file)
		}
	}
}

// printerWalk is every regular file under absolute that the tree carries. A product reached through a link, as a
// runner links each one it fetched, is walked where it points (buildcache.ProductRoot), and each file is named under
// absolute as given: walked as the link itself it lists nothing, and mega's run 2 keyed all three mutants' lowered
// products by the repository alone, so they ran one sanitized binary and escaped the oracle.
func printerWalk(tree, absolute string) ([]string, error) {
	start, productRootError := buildcache.ProductRoot(absolute)
	if productRootError != nil {
		return nil, productRootError
	}
	var files []string
	walkError := filepath.WalkDir(start, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		relative, relativePathError := filepath.Rel(start, path)
		if relativePathError != nil {
			return relativePathError
		}
		path = filepath.Join(absolute, relative)
		// Only what the tree carries: a checkout's untracked files aren't in a runner's source (#smkk3et).
		if entry.Name() == ".git" ||
			entry.Name() == "node_modules" ||
			path != absolute && !buildcache.Tracked(tree, path, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	return files, walkError
}

// printerPortFiles are the port's sources by their path under a source root: stage1/cohere for the package's own
// main.ts, a product's directory for a copy, which printerDirectoryAt lays out the same way. main.ts's imports reach
// these and nothing else of the tree (TestPrinterPortFilesAreMainsImports).
var printerPortFiles = []string{
	"graphql/token.ts",
	"graphql/characterClasses.ts",
	"graphql/blockString.ts",
	"graphql/lexer.ts",
	"graphql/parser.ts",
	"json/width.ts",
	"json/widthTables.ts",
	"graphql/printer/doc.ts",
	"graphql/printer/printer.ts",
	"graphql/printer/main.ts",
}

// printerLoweredInputs keys the lowered printer of path, a source root's graphql/printer/main.ts, by what lowering it
// reads: the compiler (compilerinputs.Lowering) and the port's sources under that root. It keyed every file the tree
// carries, by a walk that listed every directory, so a doc-only change under a lint rule's evidence/ moved all ten
// GraphQL printer products (#7zhyjz6).
func printerLoweredInputs(test *testing.T, path string) printerBuildInputs {
	test.Helper()
	tree, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	var files []string
	for _, input := range compilerinputs.Lowering() {
		files = append(files, filepath.Join(tree, filepath.FromSlash(input)))
	}
	for _, file := range printerPortFiles {
		files = append(files, filepath.Join(root, filepath.FromSlash(file)))
	}
	return printerBuildInputs{Name: "lowered GraphQL printer", Files: files, Toolchain: runtime.Version()}
}

func printerLoweredProduct(test *testing.T, path string) string {
	test.Helper()
	return printerBuild(test, printerLoweredInputs(test, path), func(dir string) error {
		start := time.Now()
		program := lowered(test, path)
		test.Logf("lowering wall %.3fs", time.Since(start).Seconds())
		start = time.Now()
		source := native.C(program)
		test.Logf("C emission wall %.3fs", time.Since(start).Seconds())
		if writeError := os.WriteFile(dir+"/port.c", []byte(source), 0644); writeError != nil {
			return writeError
		}
		return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
	})
}

func printerCompiledProduct(test *testing.T, loweredProduct string, options native.Options) string {
	test.Helper()
	data, readError := os.ReadFile(loweredProduct + "/port.c")
	if readError != nil {
		test.Fatal(readError)
	}
	clang := buildcache.Tool("clang", "--version")
	name := "release GraphQL printer"
	if options.Sanitize {
		name = "sanitized GraphQL printer"
	}
	inputs := printerInputFiles(test, filepath.Join(repository, "internal/native"))
	return printerBuild(test, printerBuildInputs{
		Name:      name,
		Files:     append([]string{loweredProduct + "/port.c"}, inputs...),
		Flags:     native.Flags(options),
		Toolchain: clang,
	}, func(dir string) error {
		start := time.Now()
		buildError := native.Build(string(data), dir+"/port", options)
		test.Logf("clang wall %.3fs", time.Since(start).Seconds())
		return buildError
	}) + "/port"
}

type printerProducts struct{ source, backend, sanitized, release string }

func preparePrinterProducts(test *testing.T, path string) printerProducts {
	test.Helper()
	loweredProduct := printerLoweredProduct(test, path)
	sanitized := printerCompiledProduct(test, loweredProduct, native.Options{Sanitize: true})
	release := ""
	if runtime.GOOS == "darwin" {
		release = printerCompiledProduct(test, loweredProduct, native.Options{})
	}
	return printerProducts{
		source:    path,
		backend:   loweredProduct + "/program.mjs",
		sanitized: sanitized,
		release:   release,
	}
}

type printerCase struct{ id, input, want string }
type printerShard struct {
	mode, path string
	cases      []printerCase
}

func enumeratePrinter(test *testing.T, mode, path, want string) []printerCase {
	test.Helper()
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.HasSuffix(string(data), "\n") || !strings.HasSuffix(want, "\n") {
		test.Fatal("corpus protocol lacks final newline")
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(inputs) != len(answers) || len(inputs) == 0 {
		test.Fatalf("%s enumeration: %d inputs, %d answers", mode, len(inputs), len(answers))
	}
	result := make([]printerCase, len(inputs))
	for i := range inputs {
		if !strings.HasPrefix(inputs[i], ">") {
			test.Fatalf("%s case %d lacks input marker", mode, i)
		}
		result[i] = printerCase{id: fmt.Sprintf("%s/case-%06d", mode, i), input: inputs[i], want: answers[i]}
	}
	return result
}

// Validate both count and ID set, and the exact original input/answer pairing.
func printerShardUnion(whole []printerCase, shards []printerShard) error {
	expected := make(map[string]printerCase, len(whole))
	for _, item := range whole {
		if _, exists := expected[item.id]; exists {
			return fmt.Errorf("duplicate unsplit id %s", item.id)
		}
		expected[item.id] = item
	}
	seen := map[string]bool{}
	count := 0
	for number, shard := range shards {
		for _, item := range shard.cases {
			original, exists := expected[item.id]
			if !exists || seen[item.id] || item != original {
				return fmt.Errorf("shard-%03d: missing, changed or repeated id %s", number, item.id)
			}
			seen[item.id] = true
			count++
		}
	}
	if count != len(whole) || len(seen) != len(expected) {
		return fmt.Errorf("shard union %d ids, unsplit %d", count, len(whole))
	}
	return nil
}

func printerShardSelection(value string, count int) ([]bool, error) {
	index, boxes := 0, 1
	if value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("ADAMIC_TEST_SHARD=%q: want i/n", value)
		}
		var parseError error
		index, parseError = strconv.Atoi(parts[0])
		if parseError != nil {
			return nil, parseError
		}
		boxes, parseError = strconv.Atoi(parts[1])
		if parseError != nil || boxes <= 0 || index < 0 || index >= boxes {
			return nil, fmt.Errorf("ADAMIC_TEST_SHARD=%q: require 0 <= i < n", value)
		}
	}
	selected := make([]bool, count)
	for i := range selected {
		selected[i] = i%boxes == index
	}
	return selected, nil
}

func printerShardDisagreement(number int, shard printerShard, result run) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("shard-%03d: exit %d, %s", number, result.exitCode, result.stderr)
	}
	var want strings.Builder
	for _, item := range shard.cases {
		want.WriteString(item.want)
		want.WriteByte('\n')
	}
	if string(result.stdout) == want.String() {
		return nil
	}
	rows := strings.Split(string(result.stdout), "\n")
	for i, item := range shard.cases {
		if i >= len(rows) || rows[i] != item.want {
			return fmt.Errorf(
				"shard-%03d %s: %s",
				number,
				item.id,
				firstDifference(string(result.stdout), want.String()),
			)
		}
	}
	return fmt.Errorf("shard-%03d: extra output or changed final newline", number)
}

func TestPrinterShardUnionRejectsMissingAndRepeated(test *testing.T) {
	test.Parallel()
	whole := []printerCase{
		{id: "defaults/case-000000", input: ">query{a}", want: "ok"},
		{id: "tabs/case-000000", input: ">query{b}", want: "ok"},
	}
	shards := []printerShard{{cases: whole[:1]}, {cases: whole[1:]}}
	if printerShardUnionError := printerShardUnion(whole, shards); printerShardUnionError != nil {
		test.Fatal(printerShardUnionError)
	}
	for _, invalid := range [][]printerShard{
		{{cases: whole[:1]}},
		{{cases: []printerCase{whole[0], whole[0]}}},
		{{cases: []printerCase{{id: whole[0].id, input: "changed", want: "ok"}, whole[1]}}},
	} {
		if printerShardUnionError := printerShardUnion(whole, invalid); printerShardUnionError == nil {
			test.Fatal("invalid union accepted")
		}
	}
}

func TestPrinterShardSelection(test *testing.T) {
	test.Parallel()
	for _, boxes := range []int{1, 2, 7, 100} {
		seen := make([]int, testPrinterAsGoCohereShards)
		for box := 0; box < boxes; box++ {
			selected, printerShardSelectionError := printerShardSelection(fmt.Sprintf("%d/%d", box, boxes), len(seen))
			if printerShardSelectionError != nil {
				test.Fatal(printerShardSelectionError)
			}
			for i, runs := range selected {
				if runs {
					seen[i]++
				}
			}
		}
		for i, count := range seen {
			if count != 1 {
				test.Fatalf("shard-%03d assigned %d times", i, count)
			}
		}
	}
	for _, invalid := range []string{"1", "x/2", "0/0", "-1/2", "2/2", "0/x"} {
		if _, printerShardSelectionError := printerShardSelection(invalid, 4); printerShardSelectionError == nil {
			test.Fatalf("invalid selector accepted: %s", invalid)
		}
	}
	selected, printerShardSelectionError := printerShardSelection("", 4)
	if printerShardSelectionError != nil {
		test.Fatal(printerShardSelectionError)
	}
	for i, runs := range selected {
		if !runs {
			test.Fatalf("unset selector omitted shard-%03d", i)
		}
	}
}
