// Package oracles owns yaml's compiler-free Go reference products.
// Both test binaries resolve paths from stage1/cohere/yaml.
package oracles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/toolchain"
)

const repository = "../../.."

var oracleProducer bool

// EnableProductProduction is called only by the owner test binary's TestMain.
func EnableProductProduction() { oracleProducer = true }

func oracleProduct(test *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	test.Helper()
	return buildcache.Product(test, inputs, func(directory string) error {
		if !oracleProducer {
			return fmt.Errorf("oracle %s is missing: settle yaml/oracles owners first", inputs.Name)
		}
		return build(directory)
	})
}
func oracleCommand(ctx context.Context, binary string, args ...string) ([]byte, error) {
	command := toolchain.CommandContextIn(ctx, "", binary, args...)
	// The built program inherits the same environment as the original command.
	command.Env = nil
	var out, errOut bytes.Buffer
	command.Stdout, command.Stderr = &out, &errOut
	if err := childguard.Run(command, childguard.Options{Unbounded: true}); err != nil {
		return nil, fmt.Errorf("%s: %w\n%s", binary, err, errOut.Bytes())
	}
	if errOut.Len() != 0 {
		return nil, fmt.Errorf("%s stderr: %s", binary, errOut.Bytes())
	}
	return out.Bytes(), nil
}
func oracleOwner(output string) string {
	switch output {
	case "go-format-oracle-owner-v1":
		return "FormatGo"
	case "go-lexer-oracle-owner-v1":
		return "LexerGo"
	case "go-cst-oracle-owner-v1":
		return "CSTGo"
	case "go-unist-oracle-owner-v1":
		return "UnistGo"
	case "go-width-oracle-owner-v1":
		return "WidthGo"
	case "go-schema-oracle-owner-v1":
		return "SchemaGo"
	case "go-scalars-oracle-owner-v1":
		return "ScalarsGo"
	case "go-props-oracle-owner-v1":
		return "PropsGo"
	case "go-compose-oracle-owner-v1":
		return "ComposeGo"
	default:
		panic("unknown oracle output: " + output)
	}
}

var lexCorpus = []struct {
	checkout, pin string
	roots         []string
}{
	{"cohere", corpusfiles.CohereCommit, []string{".github/workflows"}},
	{
		"cohere/TypeScript",
		corpusfiles.TypeScriptGoCommit,
		[]string{".custom-gcl.yml", ".github", ".golangci.yml", "tools/pipelines"},
	},
}

// LexicalCases enumerates the unchanged comparison corpus.
func LexicalCases(test *testing.T) (string, int, int) {
	test.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		test.Fatal(err)
	}
	var texts []string
	files := 0
	patterns := []string{"*.yaml", "*.yml"}
	var paths []string
	for _, corpus := range lexCorpus {
		paths = append(
			paths,
			corpusfiles.Upstream(test, filepath.Join(root, corpus.checkout), corpus.pin, corpus.roots, patterns)...,
		)
	}
	sort.Strings(paths)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			test.Fatal(err)
		}
		if !utf8.Valid(data) {
			test.Fatalf("invalid UTF-8: %s", path)
		}
		texts = append(texts, string(data))
		files++
	}
	// Each lexical state, separators, directives, tags, markers, chomping and Unicode boundaries.
	texts = append(
		texts,
		"",
		"\ufeff",
		"\ufeffa: b\n",
		"%YAML 1.1 # c\n---\na: b\n...\n",
		"a: |+\n  b\n\n\n",
		"a: |-\n  b\n\n",
		"a: >2-\n  b\n",
		"{a: [1, 2], b: \"q\\\"r\"}\n",
		"a: 'q''r'\n",
		"a: !<tag:example.org,%E2%82%AC> &ref b\nc: *ref\n",
		"a: x\n y\nz: x\n",
		"a: [x\nq: y\n",
		"a: |\n\tb\n",
		"a: \"x\ny\"\n",
		"---x\n...\n",
		"😀: café\n中: x\n",
		"a:\tx\r\n# b\r\n",
		"\x00\n",
	)
	values := []string{
		"x",
		"true",
		"null",
		"1",
		"'q''r'",
		`"a\nb"`,
		"[x,y]",
		"{a:b, c: d}",
		"&a x",
		"*a",
		"!str x",
		"😀",
		"中文",
	}
	for _, value := range values {
		for _, indicator := range []string{"", "- ", "a: ", "? ", "[", "{a: "} {
			texts = append(texts, indicator+value+"\n", indicator+value+" # c\r\n")
		}
	}
	random := rand.New(rand.NewSource(20261006))
	alphabet := []rune("abc012:?!&*[]{}|>+-.,%'\"# \\\t\r\n😀中\ufeff\x00")
	for count := 0; count < 2000; count++ {
		var text strings.Builder
		for size := random.Intn(180); size > 0; size-- {
			text.WriteRune(alphabet[random.Intn(len(alphabet))])
		}
		texts = append(texts, text.String())
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var source strings.Builder
	cases := 0
	for index, text := range texts {
		chunks := []int{0}
		if index >= files {
			chunks = append(chunks, 1, 2, 7)
		}
		for _, chunk := range chunks {
			fmt.Fprintf(&source, "%d\t%s\n", chunk, escape.Replace(text))
			cases++
		}
	}
	path := filepath.Join(test.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(source.String()), 0644); err != nil {
		test.Fatal(err)
	}
	return path, files, cases
}

// CompositionCases enumerates the unchanged comparison corpus.
func CompositionCases(test *testing.T) (string, int, int) {
	path, files, count := LexicalCases(test)
	data, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var extra strings.Builder
	for _, version := range []string{"", "%YAML 1.1\n---\n", "%YAML 1.2\n---\n"} {
		for _, tag := range []string{
			"!!str",
			"!!int",
			"!!bool",
			"!!null",
			"!!timestamp",
			"!!binary",
			"!!merge",
			"!!set",
			"!!pairs",
			"!!omap",
			"!!map",
			"!!seq",
			"!<toString>",
			"!<constructor>",
			"!<__proto__>",
			"!<unknown>",
		} {
			for _, value := range []string{
				"word",
				"2000-01-01",
				"null",
				"<<",
				"{}",
				"[]",
				"{a: null}",
				"{a: 1}",
				"[{a: 1}, {a: 2}]",
				"[{}, word]",
			} {
				extra.WriteString("0\t" + escape.Replace(version+tag+" "+value+"\n") + "\n")
				count++
			}
		}
	}
	for _, text := range []string{
		"%TAG !e! tag:example.com,2020:\n---\n!e!%E4%B8%AD value\n",
		"%TAG !e! tag:example.com,2020:\n---\n!e!%FF value\n",
		"--- a\n...\n# end\n--- b\n",
		"---\n" + strings.Repeat("k", 1025) + ": v\n",
	} {
		extra.WriteString("0\t" + escape.Replace(text) + "\n")
		count++
	}
	target := filepath.Join(test.TempDir(), "compose-cases.txt")
	if err := os.WriteFile(target, append(data, []byte(extra.String())...), 0644); err != nil {
		test.Fatal(err)
	}
	return target, files, count
}

// FormatterCases enumerates the unchanged comparison corpus.
func FormatterCases(test *testing.T) (string, int, int) {
	path, files, count := CompositionCases(test)
	data, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var extra strings.Builder
	add := func(text string) { extra.WriteString("0\t" + escape.Replace(text) + "\n"); count++ }
	for _, unit := range []string{"a", "中", "😀", "é", "👨‍👩‍👦"} {
		for width := 34; width <= 86; width += 2 {
			value := strings.Repeat(unit, width)
			for _, text := range []string{
				"key: [" + value + ", tail]\n",
				"{first: " + value + ", second: [a, b]}\n",
				"[" + value + ", {x: y}, end]\n",
				"\"" + value + "\": [a, b]\n",
			} {
				add(text)
			}
		}
	}
	for _, header := range []string{"|", "|-", "|+", "|2", "|2-", "|+2", ">", ">-", ">+", ">2", ">2-", ">+2"} {
		for _, body := range []string{
			"",
			"\n",
			"  a\n",
			"  a\n\n",
			"  a\n\n\n",
			"  a\n    b\n  c\n",
			"  a  \n  b\t\n",
		} {
			for _, prefix := range []string{"", "key: ", "- "} {
				add(prefix + header + " # header\n" + body)
			}
		}
	}
	for _, text := range []string{
		"# prettier-ignore\nkey:   [a,b]\n",
		"---\n# prettier-ignore\na:    {b: c}\n",
		"a:\n  # prettier-ignore\n  -   {b:  c}\n",
		"? [a,b]\n: {c: d}\n",
		"key: # colon\n  value\n",
		"a: 1\n\n# end\n",
		"[a,\n\n# middle\nb]\n",
		"&a\n# middle\n{b: c}\n",
		"!local &a\n# middle\nvalue\n",
		"a: 'x\\y'\n",
		"a: \"x\\\"y\"\n",
		"a: 'x\ny'\n",
		"a: \"x\\n y\"\n",
		"\ufeffa:  1\r\n",
		"\ufeff  ",
		"  \t\n",
		"",
		"---\n...\n---\n",
	} {
		add(text)
	}
	target := filepath.Join(test.TempDir(), "format-cases.txt")
	if err := os.WriteFile(target, append(data, []byte(extra.String())...), 0644); err != nil {
		test.Fatal(err)
	}
	return target, files, count
}

// Formatter finds the Go reference program owned by this package.
func Formatter(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-format", "format_go.go")
}

// Lexer finds the Go reference program owned by this package.
func Lexer(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-lexer", "lexer_go.go")
}

// ConcreteSyntaxTree finds the Go reference program owned by this package.
func ConcreteSyntaxTree(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-cst", "cst_go.go")
}

// Unist finds the Go reference program owned by this package.
func Unist(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-unist", "unist_go.go")
}

// Width finds the Go reference program owned by this package.
func Width(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-width", "width_go.go")
}

// Schema finds the Go reference program owned by this package.
func Schema(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-schema", "schema_go.go")
}

// Scalars finds the Go reference program owned by this package.
func Scalars(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-scalars", "scalar_go.go", numbersExports(
		test,
		"\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"",
		"scalar_exports.txt",
	)...)
}

// Properties finds the Go reference program owned by this package.
func Properties(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-props", "props_go.go", numbersExports(
		test,
		"\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"",
		"props_exports.txt",
	)...)
}

// Composition finds the Go reference program owned by this package.
func Composition(test *testing.T) string {
	test.Helper()
	return formatterComparison(test, "go-compose", "compose_go.go", numbersExports(
		test,
		"\n\"slices\"\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"",
		"compose_exports.txt",
	)...)
}

func formatterComparisonBuild(test *testing.T, output, adapter string, mutations []buildcache.Mutation) string {
	test.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		test.Fatal(err)
	}
	source, err := filepath.Abs(filepath.Join("testdata", adapter))
	if err != nil {
		test.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{
		"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source},
	})
	if err != nil {
		test.Fatal(err)
	}
	path := filepath.Join(test.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		test.Fatal(err)
	}
	arguments := []string{"-overlay=" + path}
	output += "-oracle-owner-v1"
	if !oracleProducer {
		inputs, err := buildcache.GoBuildInInputs(
			"cohere",
			output,
			"./command/formatter_comparison",
			arguments,
			mutations,
		)
		if err != nil {
			test.Fatal(err)
		}
		directory := buildcache.BuiltElsewhere(
			test,
			output,
			inputs,
			"stage1/cohere/yaml/oracles TestProduct_YAML"+oracleOwner(output),
			func(string) error { return fmt.Errorf("settle the yaml/oracles owner first") },
		)
		return filepath.Join(directory, output)
	}
	if len(mutations) == 0 {
		return buildcache.GoBuildIn(test, "cohere", output, "./command/formatter_comparison", arguments)
	}
	return buildcache.GoBuildInMutated(test, "cohere", output, "./command/formatter_comparison", arguments, mutations)
}

func numbersExports(test *testing.T, imports, exports string) []buildcache.Mutation {
	test.Helper()
	const numbers = "cohere/internal/format/yaml/compose/numbers.go"
	original, err := os.ReadFile(filepath.Join(repository, numbers))
	if err != nil {
		test.Fatal(err)
	}
	appended, err := os.ReadFile(filepath.Join("testdata", exports))
	if err != nil {
		test.Fatal(err)
	}
	last := strings.LastIndex(string(original), "\nfunc ")
	if last < 0 {
		test.Fatalf("%s has no top-level function to append %s after", numbers, exports)
	}
	tail := string(original[last:])
	return []buildcache.Mutation{
		{File: numbers, Before: "import (", After: "import (" + imports},
		{File: numbers, Before: tail, After: tail + string(appended)},
	}
}

var products sync.Map

type productValue struct {
	once sync.Once
	path string
}

func formatterComparison(test *testing.T, output, adapter string, mutations ...buildcache.Mutation) string {
	test.Helper()
	key := fmt.Sprintf("go %s %s %q", output, adapter, mutations)
	stored, _ := products.LoadOrStore(key, &productValue{})
	value := stored.(*productValue)
	value.once.Do(func() { value.path = formatterComparisonBuild(test, output, adapter, mutations) })
	if value.path == "" {
		test.Fatalf("product %s failed earlier in this process", key)
	}
	return value.path
}

func formatterMutantsSetupInputs(test *testing.T, oracle string) buildcache.Inputs {
	test.Helper()
	binary, err := os.ReadFile(oracle)
	if err != nil {
		test.Fatal(err)
	}
	files := []string{"internal/corpusfiles", "stage1/cohere/yaml", "go.mod", "go.work"}
	for _, corpus := range lexCorpus {
		for _, root := range corpus.roots {
			files = append(files, corpus.checkout+"/"+root)
		}
	}
	return buildcache.Inputs{
		Name:  "yaml-formatter-mutants-oracle-and-corpus-v3 oracle-owner-v1",
		Files: files,
		Flags: []string{fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary))},
	}
}

// FormatterMutantAnswers finds the complete formatter-mutant corpus and Go answers.
func FormatterMutantAnswers(test *testing.T) string {
	test.Helper()
	formatterMutantsOracle.once.Do(func() {
		// Shared setup is bounded by the product unit, before shard own work starts.
		ctx := context.Background()
		started := time.Now()
		oracle := Formatter(test)
		product := oracleProduct(test, formatterMutantsSetupInputs(test, oracle), func(directory string) error {
			cases, _, count := FormatterCases(test)
			data, err := os.ReadFile(cases)
			if err != nil {
				return err
			}
			if len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")) != count {
				return fmt.Errorf("corpus count disagrees with live enumeration")
			}
			expected, err := oracleCommand(ctx, oracle, "--cases", cases)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(directory, "cases.txt"), data, 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "expected.txt"), expected, 0644)
		})
		formatterMutantsOracle.product = product
		test.Logf("TestFormatterMutants (setup): %.3fs", time.Since(started).Seconds())

	})
	if formatterMutantsOracle.product == "" {
		test.Fatal("oracle preparation failed")
	}
	return formatterMutantsOracle.product
}

var formatterMutantsOracle struct {
	once    sync.Once
	product string
}
