package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// The composer oracle: the fork's own eemeli/yaml, run by Node, parsing each input with
// `new Parser(lineCounter.addNewLine).parse(text)` and composing the tokens with
// `new Composer({ keepSourceTokens: true, uniqueKeys: false, lineCounter, merge: true })
// .compose(tokens, true, text.length)`, exactly as yaml-unist-parser does. Every document is serialized
// whole, node by node at every depth: the node's kind and class, range, srcToken, anchor, tag, flow,
// scalar type, format, minFractionDigits, source and value, comments, spaceBefore, items, and a pair's
// key and value; the document's contents, range, comments, directives, and every error and warning with
// its name, code, message and position. The Go port composes the same text and serializes to the same
// JSON, field for field in the same order, and the two strings must be identical.
//
// A srcToken is compared by identity: both sides number every token and collection item of the CST in
// the same depth-first order, and a node's srcToken is written as that number with its type and offset.
// A number is written as its IEEE 754 bits, so -0, the infinities and every last bit compare exactly.

// forkRoot is the Prettier fork whose node_modules the oracle imports from (internal/corpus): a skip naming
// COHERE_PRETTIER_FORK when it is unset, and a failure when it is set but lacks node_modules/yaml/dist/compose/composer.js.
func forkRoot(t *testing.T) string {
	t.Helper()
	corpus.Resolve(t, "prettier-fork:node_modules/yaml/dist/compose/composer.js")
	return corpus.PrettierFork.Root(t)
}

const oracleScript = `
import { Parser, Composer, LineCounter, isAlias, isMap, isSeq, isScalar, isPair } from "FORK/node_modules/yaml/dist/index.js";
import { readFileSync } from "node:fs";

const float = new Float64Array(1);
const bits = new BigUint64Array(float.buffer);
function number(value) {
  if (Number.isNaN(value)) return "NaN";
  float[0] = value;
  return bits[0].toString(16);
}

function value(v) {
  if (v === null) return null;
  switch (typeof v) {
    case "string": return v;
    case "boolean": return v;
    case "number": return { number: number(v) };
    case "symbol": return { symbol: v.description };
    case "undefined": return { undefined: true };
  }
  if (v instanceof Date) return { date: number(v.getTime()) };
  if (v instanceof Uint8Array) return { bytes: Buffer.from(v).toString("hex") };
  return { other: Object.prototype.toString.call(v) };
}

function run(text) {
  try {
    const lineCounter = new LineCounter();
    const parser = new Parser(lineCounter.addNewLine);
    const composer = new Composer({ keepSourceTokens: true, uniqueKeys: false, lineCounter, merge: true });
    const tokens = [...parser.parse(text)];
    const ids = new Map();
    const visitToken = (token) => {
      if (!token) return;
      ids.set(token, ids.size);
      if (Array.isArray(token.start)) token.start.forEach(visitToken);
      else if (token.start) visitToken(token.start);
      if (token.value) visitToken(token.value);
      if (token.items) for (const item of token.items) {
        ids.set(item, ids.size);
        item.start.forEach(visitToken);
        if (item.key) visitToken(item.key);
        if (item.sep) item.sep.forEach(visitToken);
        if (item.value) visitToken(item.value);
      }
      if (token.props) token.props.forEach(visitToken);
      if (token.end) token.end.forEach(visitToken);
    };
    tokens.forEach(visitToken);
    const tokenId = (token) => {
      if (!ids.has(token)) return "foreign " + token.type + "@" + token.offset;
      return "#" + ids.get(token) + (token.type ? " " + token.type + "@" + token.offset : " item");
    };
    const node = (n) => {
      if (n === null || n === undefined) return null;
      const o = {};
      o.kind = isAlias(n) ? "Alias" : isMap(n) ? "Map" : isSeq(n) ? "Seq" : isScalar(n) ? "Scalar" : isPair(n) ? "Pair" : "Unknown";
      o.class = n.constructor.name;
      if (n.range !== undefined) o.range = n.range;
      if (n.srcToken !== undefined) o.srcToken = tokenId(n.srcToken);
      if (n.anchor !== undefined) o.anchor = n.anchor;
      if (n.tag !== undefined) o.tag = n.tag;
      if (n.flow !== undefined) o.flow = n.flow;
      if (n.type !== undefined) o.type = n.type;
      if (n.format !== undefined) o.format = n.format;
      if (n.minFractionDigits !== undefined) o.minFractionDigits = n.minFractionDigits;
      if (n.source !== undefined) o.source = n.source;
      if (isScalar(n)) o.value = value(n.value);
      if (n.comment !== undefined && n.comment !== null) o.comment = n.comment;
      if (n.commentBefore !== undefined && n.commentBefore !== null) o.commentBefore = n.commentBefore;
      if (n.spaceBefore !== undefined) o.spaceBefore = n.spaceBefore;
      if (n.items !== undefined) o.items = n.items.map(node);
      if (isPair(n)) { o.key = node(n.key); o.value = node(n.value); }
      return o;
    };
    const error = (e) => ({ name: e.name, code: e.code, message: e.message, pos: e.pos });
    const docs = [];
    for (const doc of composer.compose(tokens, true, text.length)) {
      const d = doc.directives;
      docs.push({
        contents: node(doc.contents),
        range: doc.range,
        comment: doc.comment,
        commentBefore: doc.commentBefore,
        directives: {
          docStart: d.docStart,
          docEnd: d.docEnd,
          yaml: { explicit: d.yaml.explicit, version: d.yaml.version },
          tags: Object.entries(d.tags).sort((a, b) => (a[0] < b[0] ? -1 : a[0] > b[0] ? 1 : 0)),
        },
        errors: doc.errors.map(error),
        warnings: doc.warnings.map(error),
      });
    }
    return JSON.stringify(docs);
  } catch (error) {
    return "throw " + error;
  }
}

const inputs = JSON.parse(readFileSync(0, "utf8"));
process.stdout.write(JSON.stringify(inputs.map(run)));
`

// upstreamOutputs runs the oracle over texts, one Node process per batch of a few megabytes.
func upstreamOutputs(t *testing.T, texts []string) []string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the composer oracle needs it")
	}
	root := forkRoot(t)
	script := filepath.Join(t.TempDir(), "oracle.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(oracleScript, "FORK", root)), 0o644); err != nil {
		t.Fatal(err)
	}
	var outputs []string
	for start := 0; start < len(texts); {
		end, size := start, 0
		for end < len(texts) && (end == start || size+len(texts[end]) < 4<<20) {
			size += len(texts[end])
			end++
		}
		outputs = append(outputs, runOracle(t, script, texts[start:end])...)
		start = end
	}
	return outputs
}

func runOracle(t *testing.T, script string, texts []string) []string {
	t.Helper()
	encoded, err := json.Marshal(texts)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", script)
	command.Stdin = strings.NewReader(string(encoded))
	output, err := command.Output()
	if err != nil {
		if exitError, isExit := err.(*exec.ExitError); isExit {
			t.Fatalf("the composer oracle failed: %v\n%s", err, exitError.Stderr)
		}
		t.Fatal(err)
	}
	var outputs []string
	if err := json.Unmarshal(output, &outputs); err != nil {
		t.Fatal(err)
	}
	if len(outputs) != len(texts) {
		t.Fatalf("the composer oracle returned %d outputs for %d inputs", len(outputs), len(texts))
	}
	return outputs
}

// composeText parses and composes text as yaml-unist-parser does, returning the tokens too.
func composeText(text string) ([]*cst.Token, []*Document) {
	units := utf16.Encode([]rune(text))
	lineCounter := cst.NewLineCounter()
	tokens := slices.Collect(cst.NewParser(lineCounter.AddNewLine).Parse(units, false))
	documents := slices.Collect(NewComposer(UnistParserOptions()).Compose(tokens, true, len(units)))
	return tokens, documents
}

// portOutput composes like the oracle and serializes like it.
func portOutput(text string) (output string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			output = fmt.Sprintf("throw %v", recovered)
		}
	}()
	tokens, documents := composeText(text)
	writer := &jsonWriter{ids: map[any]int{}}
	for _, token := range tokens {
		writer.number(token)
	}
	writer.builder.WriteByte('[')
	for index, document := range documents {
		if index > 0 {
			writer.builder.WriteByte(',')
		}
		writer.document(document)
	}
	writer.builder.WriteByte(']')
	return writer.builder.String()
}

type jsonWriter struct {
	builder strings.Builder
	ids     map[any]int
}

// number numbers a token and its descendants in the oracle's order: start, value, items (each item,
// then its start, key, sep and value), props, end.
func (writer *jsonWriter) number(token *cst.Token) {
	if token == nil {
		return
	}
	writer.ids[token] = len(writer.ids)
	for _, each := range token.Start {
		writer.number(each)
	}
	writer.number(token.FlowStart)
	writer.number(token.Value)
	for _, item := range token.Items {
		writer.ids[item] = len(writer.ids)
		for _, each := range item.Start {
			writer.number(each)
		}
		writer.number(item.Key)
		for _, each := range item.Sep {
			writer.number(each)
		}
		writer.number(item.Value)
	}
	for _, each := range token.Props {
		writer.number(each)
	}
	for _, each := range token.End {
		writer.number(each)
	}
}

func (writer *jsonWriter) raw(text string) { writer.builder.WriteString(text) }

func (writer *jsonWriter) key(first *bool, name string) {
	if !*first {
		writer.builder.WriteByte(',')
	}
	*first = false
	writer.builder.WriteString(`"` + name + `":`)
}

func (writer *jsonWriter) text(text string) { writeJSONString(&writer.builder, stringToUnits(text)) }

func (writer *jsonWriter) ints(values []int) {
	if values == nil {
		writer.raw("null")
		return
	}
	writer.builder.WriteByte('[')
	for index, value := range values {
		if index > 0 {
			writer.builder.WriteByte(',')
		}
		writer.raw(strconv.Itoa(value))
	}
	writer.builder.WriteByte(']')
}

func numberBits(value float64) string {
	if math.IsNaN(value) {
		return `"NaN"`
	}
	return `"` + strconv.FormatUint(math.Float64bits(value), 16) + `"`
}

func (writer *jsonWriter) value(value any) {
	switch v := value.(type) {
	case nil:
		writer.raw("null")
	case string:
		writer.text(v)
	case bool:
		writer.raw(strconv.FormatBool(v))
	case float64:
		writer.raw(`{"number":` + numberBits(v) + `}`)
	case *MergeKey:
		writer.raw(`{"symbol":"<<"}`)
	case Date:
		writer.raw(`{"date":` + numberBits(float64(v)) + `}`)
	case Binary:
		writer.raw(fmt.Sprintf(`{"bytes":"%x"}`, []byte(v)))
	default:
		writer.raw(fmt.Sprintf(`{"other":"%T"}`, v))
	}
}

func (writer *jsonWriter) srcToken(token *cst.Token) {
	id, found := writer.ids[token]
	if !found {
		writer.text(fmt.Sprintf("foreign %s@%d", token.Type, token.Offset))
		return
	}
	writer.text(fmt.Sprintf("#%d %s@%d", id, token.Type, token.Offset))
}

func (writer *jsonWriter) node(node *Node) {
	if node == nil {
		writer.raw("null")
		return
	}
	first := true
	writer.builder.WriteByte('{')
	writer.key(&first, "kind")
	writer.text(string(node.Kind))
	writer.key(&first, "class")
	writer.text(node.Class)
	if node.Range != nil {
		writer.key(&first, "range")
		writer.ints(node.Range)
	}
	if node.SrcToken != nil {
		writer.key(&first, "srcToken")
		writer.srcToken(node.SrcToken)
	} else if node.SrcItem != nil {
		writer.key(&first, "srcToken")
		if id, found := writer.ids[node.SrcItem]; found {
			writer.text(fmt.Sprintf("#%d item", id))
		} else {
			writer.text("foreign item")
		}
	}
	if node.HasAnchor() {
		writer.key(&first, "anchor")
		writer.text(node.Anchor)
	}
	if node.Tag != "" {
		writer.key(&first, "tag")
		writer.text(node.Tag)
	}
	if node.Flow {
		writer.key(&first, "flow")
		writer.raw("true")
	}
	if node.Type != "" {
		writer.key(&first, "type")
		writer.text(node.Type)
	}
	if node.Format != "" {
		writer.key(&first, "format")
		writer.text(node.Format)
	}
	if node.MinFractionDigits != 0 {
		writer.key(&first, "minFractionDigits")
		writer.raw(strconv.Itoa(node.MinFractionDigits))
	}
	if node.HasSource() {
		writer.key(&first, "source")
		writer.text(node.Source)
	}
	if node.Kind == KindScalar {
		writer.key(&first, "value")
		writer.value(node.ScalarValue)
	}
	if node.Comment != "" {
		writer.key(&first, "comment")
		writer.text(node.Comment)
	}
	if node.CommentBefore != "" {
		writer.key(&first, "commentBefore")
		writer.text(node.CommentBefore)
	}
	if node.SpaceBefore {
		writer.key(&first, "spaceBefore")
		writer.raw("true")
	}
	if node.Kind == KindMap || node.Kind == KindSeq {
		writer.key(&first, "items")
		writer.builder.WriteByte('[')
		for index, item := range node.Items {
			if index > 0 {
				writer.builder.WriteByte(',')
			}
			writer.node(item)
		}
		writer.builder.WriteByte(']')
	}
	if node.Kind == KindPair {
		writer.key(&first, "key")
		writer.node(node.Key)
		writer.key(&first, "value")
		writer.node(node.Value)
	}
	writer.builder.WriteByte('}')
}

func (writer *jsonWriter) nullableText(text string) {
	if text == "" {
		writer.raw("null")
		return
	}
	writer.text(text)
}

func (writer *jsonWriter) errors(errors []*YAMLError) {
	writer.builder.WriteByte('[')
	for index, yamlError := range errors {
		if index > 0 {
			writer.builder.WriteByte(',')
		}
		first := true
		writer.builder.WriteByte('{')
		writer.key(&first, "name")
		writer.text(yamlError.Name)
		writer.key(&first, "code")
		writer.text(yamlError.Code)
		writer.key(&first, "message")
		writer.text(yamlError.Message)
		writer.key(&first, "pos")
		writer.ints(yamlError.Pos[:])
		writer.builder.WriteByte('}')
	}
	writer.builder.WriteByte(']')
}

func (writer *jsonWriter) document(document *Document) {
	first := true
	writer.builder.WriteByte('{')
	writer.key(&first, "contents")
	writer.node(document.Contents)
	writer.key(&first, "range")
	writer.ints(document.Range)
	writer.key(&first, "comment")
	writer.nullableText(document.Comment)
	writer.key(&first, "commentBefore")
	writer.nullableText(document.CommentBefore)
	writer.key(&first, "directives")
	directives := document.Directives
	inner := true
	writer.builder.WriteByte('{')
	writer.key(&inner, "docStart")
	if directives.DocStart {
		writer.raw("true")
	} else {
		writer.raw("null")
	}
	writer.key(&inner, "docEnd")
	writer.raw(strconv.FormatBool(directives.DocEnd))
	writer.key(&inner, "yaml")
	writer.raw(`{"explicit":` + strconv.FormatBool(directives.YAML.Explicit) + `,"version":`)
	writer.text(directives.YAML.Version)
	writer.raw("}")
	writer.key(&inner, "tags")
	handles := make([]string, 0, len(directives.Tags))
	for handle := range directives.Tags {
		handles = append(handles, handle)
	}
	slices.SortFunc(handles, func(a string, b string) int {
		return slices.Compare(stringToUnits(a), stringToUnits(b))
	})
	writer.builder.WriteByte('[')
	for index, handle := range handles {
		if index > 0 {
			writer.builder.WriteByte(',')
		}
		writer.builder.WriteByte('[')
		writer.text(handle)
		writer.builder.WriteByte(',')
		writer.text(directives.Tags[handle])
		writer.builder.WriteByte(']')
	}
	writer.builder.WriteByte(']')
	writer.builder.WriteByte('}')
	writer.key(&first, "errors")
	writer.errors(document.Errors)
	writer.key(&first, "warnings")
	writer.errors(document.Warnings)
	writer.builder.WriteByte('}')
}

// compareOutputs reports each input whose port output differs from upstream's, up to limit of them.
func compareOutputs(t *testing.T, label string, texts []string, limit int) (failures int) {
	t.Helper()
	expected := upstreamOutputs(t, texts)
	threw, withErrors, withWarnings := 0, 0, 0
	for index, text := range texts {
		actual := portOutput(text)
		if actual != expected[index] {
			failures++
			if failures <= limit {
				t.Errorf("%s %q: %s", label, truncateInput(text), firstDifference(expected[index], actual))
			}
		}
		switch {
		case strings.HasPrefix(expected[index], "throw"):
			threw++
		case strings.Contains(expected[index], `"name":"YAMLParseError"`):
			withErrors++
		}
		if strings.Contains(expected[index], `"name":"YAMLWarning"`) {
			withWarnings++
		}
	}
	t.Logf("%s: of %d inputs, upstream threw on %d, reported errors on %d and warnings on %d",
		label, len(texts), threw, withErrors, withWarnings)
	return failures
}

func firstDifference(expected string, actual string) string {
	index := 0
	for index < len(expected) && index < len(actual) && expected[index] == actual[index] {
		index++
	}
	from := max(index-160, 0)
	return fmt.Sprintf("differs at byte %d\n  want …%s\n  got  …%s", index,
		expected[from:min(index+120, len(expected))], actual[from:min(index+120, len(actual))])
}

func truncateInput(input string) string {
	if len(input) > 80 {
		return input[:80] + "…"
	}
	return input
}

// normalize makes text what both sides read: invalid UTF-8 becomes U+FFFD per byte on both.
func normalize(text string) string {
	return string([]rune(text))
}

// TestDocumentsMatchUpstream runs every fixture through both sides.
func TestDocumentsMatchUpstream(t *testing.T) {
	t.Parallel()
	texts := make([]string, 0, len(fixtures))
	for _, fixture := range fixtures {
		texts = append(texts, normalize(fixture))
	}
	failures := compareOutputs(t, "fixture", texts, 60)
	t.Logf("%d of %d fixtures match upstream", len(texts)-failures, len(texts))
}

// TestGeneratedDocumentsMatchUpstream compares inputs drawn from YAML's indicators, properties, scalar
// styles, spaces, tabs and line breaks, where the hand-written fixtures cannot reach every combination.
func TestGeneratedDocumentsMatchUpstream(t *testing.T) {
	t.Parallel()
	count := 30000
	if testing.Short() {
		count = 3000
	}
	failures := compareOutputs(t, "generated", generatedInputs(count), 40)
	t.Logf("%d of %d generated inputs match upstream", count-failures, count)
}

// TestOracleCanFail proves the comparison sees a difference: documents composed from one input compared
// against upstream's for another must differ.
func TestOracleCanFail(t *testing.T) {
	t.Parallel()
	expected := upstreamOutputs(t, []string{"a: b", "a: 1", "[a]"})
	if portOutput("a:  b") == expected[0] {
		t.Fatal("the comparison found no difference between a: b and a:  b")
	}
	if portOutput("a: 1.0") == expected[1] {
		t.Fatal("the comparison found no difference between the values of a: 1 and a: 1.0")
	}
	if portOutput("[a ]") == expected[2] {
		t.Fatal("the comparison found no difference between [a] and [a ]")
	}
}

// TestYamlTestSuiteMatchesUpstream compares every case of the yaml-test-suite package in the fork.
func TestYamlTestSuiteMatchesUpstream(t *testing.T) {
	t.Parallel()
	root := forkRoot(t)
	suite := filepath.Join(root, "node_modules", "yaml-test-suite", "index.js")
	if _, err := os.Stat(suite); err != nil {
		t.Skip("the fork has no yaml-test-suite package")
	}
	script := filepath.Join(t.TempDir(), "suite.mjs")
	source := `import suite from "SUITE";
const inputs = [];
for (const test of suite) for (const testCase of test.cases) if (typeof testCase.yaml === "string") inputs.push(testCase.yaml);
process.stdout.write(JSON.stringify(inputs));`
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(source, "SUITE", suite)), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("node", script).Output()
	if err != nil {
		t.Fatalf("listing the yaml-test-suite failed: %v", err)
	}
	var texts []string
	if err := json.Unmarshal(output, &texts); err != nil {
		t.Fatal(err)
	}
	failures := compareOutputs(t, "suite", texts, 40)
	t.Logf("%d of %d yaml-test-suite cases match upstream", len(texts)-failures, len(texts))
}

// TestCorpusMatchesUpstream compares every .yaml and .yml file under COHERE_YAML_CORPUS, a measuring run.
func TestCorpusMatchesUpstream(t *testing.T) {
	t.Parallel()
	root := os.Getenv("COHERE_YAML_CORPUS")
	if root == "" {
		t.Skip("set COHERE_YAML_CORPUS to a directory to compare every .yaml and .yml file under it")
	}
	var paths, texts []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".git") {
			return filepath.SkipDir
		}
		extension := strings.ToLower(filepath.Ext(path))
		if !entry.IsDir() && entry.Type().IsRegular() && (extension == ".yaml" || extension == ".yml") {
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			paths = append(paths, path)
			texts = append(texts, normalize(string(source)))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := upstreamOutputs(t, texts)
	failures, withErrors := 0, 0
	for index, text := range texts {
		if strings.Contains(expected[index], `"name":"YAMLParseError"`) {
			withErrors++
		}
		if actual := portOutput(text); actual != expected[index] {
			failures++
			if failures <= 40 {
				t.Errorf("%s: %s", paths[index], firstDifference(expected[index], actual))
			}
		}
	}
	t.Logf("%d of %d files match upstream's documents (%d of them with errors)", len(texts)-failures, len(texts), withErrors)
}
