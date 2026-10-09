package mdast

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/system-inc/cohere/internal/format/oracletest"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// The tree oracle: the embedded markdown bundle's own parser, run by goja, the exact code the format
// oracle runs. Every node of the Go tree must carry the same type, the same fields with the same values
// (null distinct from absent), and the same positions, offsets converted from UTF-16 to bytes.
//
// The fixtures' trees are recorded (see oracletest) as the JSON text upstream's parser serializes, so no
// JavaScript runs for them and the runtime is built only when they are recorded again. The corpus reads
// files off disk, so its oracle stays live, with no golden.

type treeOracle struct {
	golden  *oracletest.Golden
	once    sync.Once
	runtime *goja.Runtime
	err     error
}

// newTreeOracle answers from the test's golden.
func newTreeOracle(t *testing.T) *treeOracle {
	t.Helper()
	return &treeOracle{golden: oracletest.Open(t, t.Name())}
}

// newLiveTreeOracle answers from the bundle's parser on every call.
func newLiveTreeOracle(t *testing.T) *treeOracle {
	t.Helper()
	oracle := &treeOracle{}
	oracle.once.Do(oracle.load)
	if oracle.err != nil {
		t.Fatal(oracle.err)
	}
	return oracle
}

func (oracle *treeOracle) load() {
	bundles, err := prettier.Bundles()
	if err != nil {
		oracle.err = err
		return
	}
	runtime := goja.New()
	for _, name := range []string{"standalone.js", "plugins/markdown.js"} {
		if _, err := runtime.RunString(string(bundles.Files[name])); err != nil {
			oracle.err = fmt.Errorf("evaluating %s: %w", name, err)
			return
		}
	}
	oracle.runtime = runtime
}

// parse returns upstream's tree as generic JSON, with `data` (hast hints for HTML output) dropped.
func (oracle *treeOracle) parse(text string) (any, error) {
	var serialized string
	var err error
	if oracle.golden == nil {
		serialized, err = oracle.serialize(text)
	} else {
		serialized, err = oracle.golden.Answer(oracletest.Key("tree", text), func() (string, error) {
			oracle.once.Do(oracle.load)
			if oracle.err != nil {
				return "", oracle.err
			}
			return oracle.serialize(text)
		})
	}
	if err != nil {
		return nil, err
	}
	var tree any
	if err := json.Unmarshal([]byte(serialized), &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

// serialize is upstream's tree for text, as the JSON text the bundle's parser and JSON.stringify give.
func (oracle *treeOracle) serialize(text string) (string, error) {
	oracle.runtime.Set("__text", text)
	value, err := oracle.runtime.RunString(`JSON.stringify(prettierPlugins.markdown.parsers.markdown.parse(__text),
		(key, value) => key === "data" ? undefined : value)`)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

// convertOffsets rewrites every position offset in upstream's tree from UTF-16 units to bytes.
func convertOffsets(tree any, offsets []int) {
	switch typed := tree.(type) {
	case map[string]any:
		for key, value := range typed {
			if key == "offset" {
				if number, isNumber := value.(float64); isNumber {
					typed[key] = float64(offsets[int(number)])
				}
				continue
			}
			convertOffsets(value, offsets)
		}
	case []any:
		for _, value := range typed {
			convertOffsets(value, offsets)
		}
	}
}

// describe renders a Go node with exactly the keys upstream's node of that type has.
func describe(node *Node) map[string]any {
	description := map[string]any{"type": node.NodeType}
	if node.Position != nil {
		description["position"] = map[string]any{
			"start": describePoint(node.Position.Start),
			"end":   describePoint(node.Position.End),
		}
	}
	if node.IsParent {
		children := make([]any, len(node.Children))
		for index, child := range node.Children {
			children[index] = describe(child)
		}
		description["children"] = children
	}
	if node.IsLiteral {
		if node.ValueNull {
			description["value"] = nil
		} else {
			description["value"] = node.Value
		}
	}
	optional := func(key string, value *string) {
		if value == nil {
			description[key] = nil
		} else {
			description[key] = *value
		}
	}
	switch node.NodeType {
	case "heading":
		description["depth"] = float64(node.Depth)
	case "code":
		optional("lang", node.Lang)
		optional("meta", node.Meta)
	case "math":
		optional("meta", node.Meta)
	case "definition":
		description["identifier"] = node.Identifier
		optional("label", node.Label)
		optional("title", node.Title)
		description["url"] = node.URL
	case "link":
		optional("title", node.Title)
		description["url"] = node.URL
	case "image":
		optional("title", node.Title)
		description["url"] = node.URL
		optional("alt", node.Alt)
	case "linkReference":
		description["identifier"] = node.Identifier
		optional("label", node.Label)
		description["referenceType"] = node.ReferenceType
	case "imageReference":
		optional("alt", node.Alt)
		description["identifier"] = node.Identifier
		optional("label", node.Label)
		description["referenceType"] = node.ReferenceType
	case "list":
		description["ordered"] = node.Ordered
		if node.Start == nil {
			description["start"] = nil
		} else {
			description["start"] = float64(*node.Start)
		}
		description["spread"] = node.Spread
	case "listItem":
		description["spread"] = node.Spread
		if node.Checked == nil {
			description["checked"] = nil
		} else {
			description["checked"] = *node.Checked
		}
	case "footnoteReference", "footnoteDefinition":
		description["identifier"] = node.Identifier
		optional("label", node.Label)
	case "table":
		align := make([]any, len(node.Align))
		for index, value := range node.Align {
			if value != "" {
				align[index] = value
			}
		}
		description["align"] = align
	case "frontMatter":
		frontMatter := node.FrontMatter
		description["language"] = frontMatter.Language
		optional("explicitLanguage", frontMatter.ExplicitLanguage)
		description["value"] = frontMatter.Value
		description["startDelimiter"] = frontMatter.StartDelimiter
		description["endDelimiter"] = frontMatter.EndDelimiter
		description["raw"] = frontMatter.Raw
	}
	return description
}

func describePoint(point Point) map[string]any {
	return map[string]any{"line": float64(point.Line), "column": float64(point.Column), "offset": float64(point.Offset)}
}

// normalizeUpstream drops the fields of upstream's front matter node that are parser bookkeeping (its
// own start and end, which the position restates).
func normalizeUpstream(tree any) {
	switch typed := tree.(type) {
	case map[string]any:
		if typed["type"] == "frontMatter" {
			delete(typed, "start")
			delete(typed, "end")
		}
		for _, value := range typed {
			normalizeUpstream(value)
		}
	case []any:
		for _, value := range typed {
			normalizeUpstream(value)
		}
	}
}

// compareTree reports the first difference between upstream's tree and the Go tree for a text.
func compareTree(oracle *treeOracle, text string) string {
	expected, err := oracle.parse(text)
	if err != nil {
		return "upstream failed: " + err.Error()
	}
	convertOffsets(expected, byteOffsets(text, nil))
	normalizeUpstream(expected)

	root, err := ParseMarkdown(text, nil)
	if err != nil {
		return "the port failed: " + err.Error()
	}
	// Round-trip through JSON so both sides hold the same Go types.
	encoded, _ := json.Marshal(describe(root))
	var actual any
	_ = json.Unmarshal(encoded, &actual)

	return firstTreeDifference("root", expected, actual)
}

func firstTreeDifference(path string, expected any, actual any) string {
	switch want := expected.(type) {
	case map[string]any:
		got, isMap := actual.(map[string]any)
		if !isMap {
			return fmt.Sprintf("%s: want an object, got %v", path, actual)
		}
		keys := map[string]bool{}
		for key := range want {
			keys[key] = true
		}
		for key := range got {
			keys[key] = true
		}
		sorted := make([]string, 0, len(keys))
		for key := range keys {
			sorted = append(sorted, key)
		}
		sort.Strings(sorted)
		// Type first, so a wrong node reads as a wrong node, not as a field mismatch.
		sort.SliceStable(sorted, func(left, right int) bool { return sorted[left] == "type" && sorted[right] != "type" })
		for _, key := range sorted {
			wantValue, wantPresent := want[key]
			gotValue, gotPresent := got[key]
			if wantPresent != gotPresent {
				return fmt.Sprintf("%s.%s: present upstream %v, in the port %v (want %v, got %v)", path, key, wantPresent, gotPresent, wantValue, gotValue)
			}
			if difference := firstTreeDifference(path+"."+key, wantValue, gotValue); difference != "" {
				return difference
			}
		}
		return ""
	case []any:
		got, isSlice := actual.([]any)
		if !isSlice {
			return fmt.Sprintf("%s: want a list, got %v", path, actual)
		}
		for index := 0; index < len(want) || index < len(got); index++ {
			if index >= len(want) || index >= len(got) {
				return fmt.Sprintf("%s: want %d items, got %d", path, len(want), len(got))
			}
			if difference := firstTreeDifference(fmt.Sprintf("%s[%d]", path, index), want[index], got[index]); difference != "" {
				return difference
			}
		}
		return ""
	default:
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Sprintf("%s: want %#v, got %#v", path, expected, actual)
		}
		return ""
	}
}

var treeFixtures = []string{
	"",
	"a",
	"a\nb\n\nc",
	"# x",
	"---\ntitle: x\n---\n\nbody",
	"---\nt: é😀\n---\n\nbody é😀 x",
	"+++\na = 1\n+++\n",
	"é😀 a\n\nb",
	"***",
	"a  \nb\\\nc",
}

func TestTreesMatchUpstream(t *testing.T) {
	t.Parallel()
	oracle := newTreeOracle(t)
	for _, text := range treeFixtures {
		if difference := compareTree(oracle, text); difference != "" {
			t.Errorf("%q: %s", text, difference)
		}
	}
}

// TestTreeOracleCanFail: the comparison must see a wrong node and a wrong offset.
func TestTreeOracleCanFail(t *testing.T) {
	t.Parallel()
	oracle := newTreeOracle(t)
	expected, err := oracle.parse("*a*")
	if err != nil {
		t.Fatal(err)
	}
	root, _ := ParseMarkdown("**a**", nil)
	encoded, _ := json.Marshal(describe(root))
	var actual any
	_ = json.Unmarshal(encoded, &actual)
	if firstTreeDifference("root", expected, actual) == "" {
		t.Fatal("no difference found between the trees of *a* and **a**")
	}
}

// TestCorpusTreesMatchUpstream compares every .md file under COHERE_MARKDOWN_CORPUS.
func TestCorpusTreesMatchUpstream(t *testing.T) {
	t.Parallel()
	root := os.Getenv("COHERE_MARKDOWN_CORPUS")
	if root == "" {
		t.Skip("set COHERE_MARKDOWN_CORPUS to a directory to compare every .md file under it")
	}
	// The live oracle, not a golden: the corpus is whatever markdown is on disk.
	oracle := newLiveTreeOracle(t)
	total, failures := 0, 0
	byDifference := map[string]int{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".git") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(strings.ReplaceAll(string(source), "\r\n", "\n"), "\r", "\n")
		text = strings.TrimPrefix(text, "\uFEFF")
		total++
		if difference := compareTree(oracle, text); difference != "" {
			failures++
			byDifference[treeDifferenceClass(difference)]++
			if failures <= 30 {
				t.Errorf("%s: %s", path, difference)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	classes := make([]string, 0, len(byDifference))
	for class := range byDifference {
		classes = append(classes, class)
	}
	sort.Slice(classes, func(left, right int) bool { return byDifference[classes[left]] > byDifference[classes[right]] })
	for _, class := range classes {
		t.Logf("%5d  %s", byDifference[class], class)
	}
	t.Logf("%d of %d files match upstream's tree", total-failures, total)
}

// treeDifferenceClass groups differences by the field path without indexes.
func treeDifferenceClass(difference string) string {
	head := difference
	if colon := strings.Index(head, ":"); colon >= 0 {
		head = head[:colon]
	}
	var builder strings.Builder
	depth := 0
	for _, character := range head {
		switch {
		case character == '[':
			depth++
		case character == ']':
			depth--
		case depth == 0:
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
