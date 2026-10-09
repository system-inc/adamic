package yaml

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/system-inc/cohere/internal/format/differential"
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/oracletest"
	"github.com/system-inc/cohere/internal/format/prettier"
	"github.com/system-inc/cohere/internal/format/printing"
)

// The oracle for what prettier.Engine cannot be asked: proseWrap, which formatoptions.Options does not carry
// and every corpus leaves at "preserve", and a file path the engine does not route (.prettierrc). These
// run the same embedded bundles in their own runtime, with the option given.

// Its answers are recorded (see oracletest), so the runtime is built only when they are recorded again.
type optionsOracle struct {
	golden  *oracletest.Golden
	once    sync.Once
	runtime *goja.Runtime
	err     error
}

func newOptionsOracle(t *testing.T) *optionsOracle {
	t.Helper()
	return &optionsOracle{golden: oracletest.Open(t, t.Name())}
}

func (oracle *optionsOracle) load() {
	bundles, err := prettier.Bundles()
	if err != nil {
		oracle.err = err
		return
	}
	runtime := goja.New()
	for _, name := range prettier.BundleFiles {
		if _, err := runtime.RunString(string(bundles.Files[name])); err != nil {
			oracle.err = fmt.Errorf("evaluating %s: %w", name, err)
			return
		}
	}
	oracle.runtime = runtime
}

func (oracle *optionsOracle) format(text string, filePath string, parser string, options formatoptions.Options, proseWrap string) (string, error) {
	key := oracletest.Key("options", text, filePath, parser, fmt.Sprintf("%+v", options), proseWrap)
	return oracle.golden.Answer(key, func() (string, error) {
		oracle.once.Do(oracle.load)
		if oracle.err != nil {
			return "", oracle.err
		}
		return oracle.formatLive(text, filePath, parser, options, proseWrap)
	})
}

func (oracle *optionsOracle) formatLive(text string, filePath string, parser string, options formatoptions.Options, proseWrap string) (string, error) {
	oracle.runtime.Set("__source", text)
	oracle.runtime.Set("__filePath", filePath)
	oracle.runtime.Set("__parser", parser)
	oracle.runtime.Set("__proseWrap", proseWrap)
	oracle.runtime.Set("__printWidth", options.PrintWidth)
	oracle.runtime.Set("__tabWidth", options.TabWidth)
	oracle.runtime.Set("__useTabs", options.UseTabs)
	oracle.runtime.Set("__singleQuote", options.SingleQuote)
	oracle.runtime.Set("__bracketSpacing", options.BracketSpacing)
	oracle.runtime.Set("__trailingComma", options.TrailingComma)
	oracle.runtime.Set("__endOfLine", options.EndOfLine)
	value, err := oracle.runtime.RunString(`prettier.format(__source, {
		filepath: __filePath, parser: __parser, plugins: prettierPlugins, proseWrap: __proseWrap,
		printWidth: __printWidth, tabWidth: __tabWidth, useTabs: __useTabs, singleQuote: __singleQuote,
		bracketSpacing: __bracketSpacing, trailingComma: __trailingComma, endOfLine: __endOfLine,
	})`)
	if err != nil {
		return "", err
	}
	// prettier.format returns a promise; each RunString drains goja's job queue, so pump until settled.
	promise, isPromise := value.Export().(*goja.Promise)
	if !isPromise {
		return value.String(), nil
	}
	for pump := 0; pump < 100000 && promise.State() == goja.PromiseStatePending; pump++ {
		if _, err := oracle.runtime.RunString("0"); err != nil {
			return "", err
		}
	}
	if promise.State() != goja.PromiseStateFulfilled {
		return "", fmt.Errorf("%v", promise.Result())
	}
	return promise.Result().String(), nil
}

// proseWrapFixtures are the inputs whose words the always and never branches join and split: multi-line
// plain and quoted scalars, folded blocks, trailing backslashes, and long lines at a narrow width.
var proseWrapFixtures = []string{
	"a: b\n  c\n\n  d", "a: \"b\n  c\"", "a: 'b\n\n  c'", "a: \"b \\\n  c\"", "a: \"b\\\n  c\"", "a: \"b c \\\n  d\\\n  e\"",
	"a: >\n  x\n  y\n\n  z\n", "a: >\n  x\n    y\n  z\n", "a: >\n  x  \n  y\n", "a: >\n   x\n  y\n", "a: >-\n  x y  z\n",
	"a: >\n  word word word word word word word word word word word word word word word word word word\n",
	"a: word word word word word word word word word word word word word word word word word word word",
	"a: 'word word word word word word word word word word word word word word word word word word word'",
	"? word word word word word word word word word word word word word word word word\n: b",
	"word word word word word word word: word word word word word word word word word word",
	"a: b c\nd: \"e\n  f\"", "- a b\n  c d\n- \"e\n  f\"", "a: |\n  x y\n  z\n", "a: b  c\n  d", "a: \"b\\\n\n  c\"",
	"a: >\n  x\n\n\n  y\n\nb: c", "key: 'it''s a long line of single quoted text that wraps'",
	"a: word  word  word  word  word  word  word  word  word  word  word  word  word  word  word",
	"a: 'word  word  word  word  word  word  word  word  word  word  word  word  word  word  word'",
	"a: >\n  word  word  word  word  word  word  word  word  word  word  word  word  word  word\n",
	"a: >\n  \u00a0x\n  y\n", "a: >\n  x\u00a0\n  y\n", "a: plain\u00a0 b\n  c",
}

func TestProseWrapMatchesOracle(t *testing.T) {
	t.Parallel()
	oracle := newOptionsOracle(t)
	inputs := append(append(append([]string{}, proseWrapFixtures...), formatFixtures...), suiteInputs(t)...)
	texts, byteOrderMarks, trees := parseInputs(t, inputs)
	failures, fullStackFailures, compared := 0, 0, 0
	for _, proseWrap := range []string{"always", "never"} {
		for variantIndex, options := range formatVariants[:3] {
			for index, input := range inputs {
				expected, err := oracle.format(input, "fixture.yaml", "yaml", options, proseWrap)
				if err != nil {
					continue
				}
				compared++
				// The full stack: the Go parser and the printer, from the text.
				if fullStack, err := formatFullStack("fixture.yaml", input, options, proseWrap, nil); err != nil || fullStack != expected {
					fullStackFailures++
					if err != nil {
						t.Errorf("%s, variant %d, %q: Format: %v", proseWrap, variantIndex, input, err)
					} else {
						t.Errorf("%s, variant %d, %q: Format: %s", proseWrap, variantIndex, input, differential.FirstDifference(expected, fullStack))
					}
				}
				actual, err := formatParsed("fixture.yaml", trees[index], texts[index], byteOrderMarks[index], options, proseWrap, nil)
				if err != nil {
					failures++
					t.Errorf("%s, variant %d, %q: %v", proseWrap, variantIndex, input, err)
					continue
				}
				if actual != expected {
					failures++
					t.Errorf("%s, variant %d, %q: %s", proseWrap, variantIndex, input, differential.FirstDifference(expected, actual))
				}
			}
		}
	}
	t.Logf("%d proseWrap comparisons: full stack %d differ, tree loader %d differ", compared, fullStackFailures, failures)
}

// jsonTextToDoc stands in for the native JSON printer: the oracle formats the JSON, and the text comes
// back as a doc with its trailing newline stripped, as textToDoc strips the trailing hardline.
func jsonTextToDoc(t *testing.T, oracle *optionsOracle, options formatoptions.Options) printing.TextToDoc {
	return func(text string, parser string) (doc.Doc, error) {
		if parser != "json" {
			t.Errorf("the YAML printer asked for parser %q", parser)
		}
		formatted, err := oracle.format(text, "embedded.json", "json", options, "preserve")
		if err != nil {
			return nil, err
		}
		return doc.Text(strings.TrimSuffix(formatted, "\n")), nil
	}
}

// TestPrettierRcEmbedsJSON covers embed.js: a .prettierrc, .stylelintrc or .lintstagedrc that is JSON
// prints as JSON, and one that is YAML falls back to the YAML printer; any other name never embeds.
func TestPrettierRcEmbedsJSON(t *testing.T) {
	t.Parallel()
	oracle := newOptionsOracle(t)
	inputs := []string{
		`{"semi":false,"singleQuote":true}`, "{\n  \"a\": [1,2,3]\n}\n", "semi: false\nsingleQuote:   true\n",
		"# c\nsemi: false", "[1,  2]", "{a: 1}", "", "\"a\"",
	}
	fileNames := []string{".prettierrc", "a/.stylelintrc", "a\\.lintstagedrc", "prettierrc", ".prettierrc.yaml", "x.yaml"}
	texts, byteOrderMarks, trees := parseInputs(t, inputs)
	for _, options := range formatVariants[:2] {
		textToDoc := jsonTextToDoc(t, oracle, options)
		for _, fileName := range fileNames {
			for index, input := range inputs {
				expected, err := oracle.format(input, fileName, "yaml", options, "preserve")
				if err != nil {
					if trees[index].err == "" {
						t.Errorf("%s, %q: the oracle failed (%v) but the parser did not", fileName, input, err)
					}
					if _, fullStackErr := formatFullStack(fileName, input, options, "preserve", textToDoc); !printing.IsSyntax(fullStackErr) {
						t.Errorf("%s, %q: the oracle failed (%v) but Format gave %v", fileName, input, err, fullStackErr)
					}
					continue
				}
				if fullStack, err := formatFullStack(fileName, input, options, "preserve", textToDoc); err != nil {
					t.Errorf("%s, %q: Format: %v", fileName, input, err)
				} else if fullStack != expected {
					t.Errorf("%s, %q: Format: %s", fileName, input, differential.FirstDifference(expected, fullStack))
				}
				actual, err := formatParsed(fileName, trees[index], texts[index], byteOrderMarks[index], options, "preserve", textToDoc)
				if err != nil {
					t.Errorf("%s, %q: %v", fileName, input, err)
					continue
				}
				if actual != expected {
					t.Errorf("%s, %q: %s", fileName, input, differential.FirstDifference(expected, actual))
				}
			}
		}
	}
}

// TestPrettierRcWithoutTextToDocPrintsYAML: with no textToDoc the embed fails, and the core prints the
// file as YAML, as upstream does when the JSON format throws.
func TestPrettierRcWithoutTextToDocPrintsYAML(t *testing.T) {
	t.Parallel()
	texts, _, trees := parseInputs(t, []string{`{"a":1}`})
	actual, err := PrintFile(".prettierrc", trees[0].root, texts[0], formatoptions.Default(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if actual != "{ 'a': 1 }\n" {
		t.Fatalf("got %q", actual)
	}
}

func docString(document doc.Doc) string {
	return doc.Print(document, doc.Options{PrintWidth: 120, TabWidth: 4})
}

func stripTrailingHardline(document doc.Doc) doc.Doc { return doc.StripTrailingHardline(document) }
