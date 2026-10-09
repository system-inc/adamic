package cst

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/corpus"
)

// The CST oracle: the fork's own eemeli/yaml, run by Node, parsing each input with
// `new Parser(lineCounter.addNewLine).parse(text)` the way yaml-unist-parser does, and every token it
// yields serialized whole. The Go port serializes its tokens to the same JSON, and the two strings must be
// identical: every field of every token at every depth, the fields' presence (an absent end against an
// empty one, an absent key against a null one, an indent a token does not have), every offset, and the
// LineCounter's lineStarts. Both sides sort object keys, so key insertion order, which carries no meaning,
// is the only thing not compared.
//
// An input can also be fed in chunks, parse(chunk, true) per chunk and parse('', false) to finish, which
// drives the lexer's incomplete-input paths (setNext and the buffer it keeps).

// forkRoot is the Prettier fork whose node_modules the oracle imports from (internal/corpus): a skip naming
// COHERE_PRETTIER_FORK when it is unset, and a failure when it is set but lacks node_modules/yaml/dist/parse/parser.js.
func forkRoot(t *testing.T) string {
	t.Helper()
	corpus.Resolve(t, "prettier-fork:node_modules/yaml/dist/parse/parser.js")
	return corpus.PrettierFork.Root(t)
}

const oracleScript = `
import { Parser, LineCounter } from "FORK/node_modules/yaml/dist/index.js";
import { readFileSync } from "node:fs";

function canonical(value) {
  if (Array.isArray(value)) return value.map(canonical);
  if (value !== null && typeof value === "object") {
    const sorted = {};
    for (const key of Object.keys(value).sort()) sorted[key] = canonical(value[key]);
    return sorted;
  }
  return value;
}

function run({ text, chunk }) {
  try {
    const lineCounter = new LineCounter();
    const parser = new Parser(lineCounter.addNewLine);
    const tokens = [];
    if (!chunk) {
      for (const token of parser.parse(text)) tokens.push(token);
    } else {
      for (let i = 0; i < text.length; i += chunk)
        for (const token of parser.parse(text.slice(i, i + chunk), true)) tokens.push(token);
      for (const token of parser.parse("", false)) tokens.push(token);
    }
    return JSON.stringify(canonical({ lineStarts: lineCounter.lineStarts, tokens }));
  } catch (error) {
    return "throw";
  }
}

const inputs = JSON.parse(readFileSync(0, "utf8"));
process.stdout.write(JSON.stringify(inputs.map(run)));
`

// oracleInput is one parse: the whole text at once when Chunk is 0, else Chunk code units at a time.
type oracleInput struct {
	Text  string `json:"text"`
	Chunk int    `json:"chunk"`
}

// upstreamOutputs runs the oracle over inputs, one Node process per batch of a few megabytes.
func upstreamOutputs(t *testing.T, inputs []oracleInput) []string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the CST oracle needs it")
	}
	root := forkRoot(t)
	directory := t.TempDir()
	script := filepath.Join(directory, "oracle.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(oracleScript, "FORK", root)), 0o644); err != nil {
		t.Fatal(err)
	}
	var outputs []string
	for start := 0; start < len(inputs); {
		end, size := start, 0
		for end < len(inputs) && (end == start || size+len(inputs[end].Text) < 4<<20) {
			size += len(inputs[end].Text)
			end++
		}
		outputs = append(outputs, runOracle(t, script, inputs[start:end])...)
		start = end
	}
	return outputs
}

func runOracle(t *testing.T, script string, inputs []oracleInput) []string {
	t.Helper()
	encoded, err := json.Marshal(inputs)
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
			t.Fatalf("the CST oracle failed: %v\n%s", err, exitError.Stderr)
		}
		t.Fatal(err)
	}
	var outputs []string
	if err := json.Unmarshal(output, &outputs); err != nil {
		t.Fatal(err)
	}
	if len(outputs) != len(inputs) {
		t.Fatalf("the CST oracle returned %d outputs for %d inputs", len(outputs), len(inputs))
	}
	return outputs
}

// sourceUnits is the text as a JavaScript string holds it.
func sourceUnits(text string) []uint16 {
	return utf16.Encode([]rune(text))
}

// portOutput parses like the oracle and serializes like it.
func portOutput(input oracleInput) (output string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			output = "throw"
		}
	}()
	text := sourceUnits(input.Text)
	lineCounter := NewLineCounter()
	parser := NewParser(lineCounter.AddNewLine)
	tokens := []*Token{}
	if input.Chunk == 0 {
		for token := range parser.Parse(text, false) {
			tokens = append(tokens, token)
		}
	} else {
		for i := 0; i < len(text); i += input.Chunk {
			for token := range parser.Parse(slice(text, i, i+input.Chunk), true) {
				tokens = append(tokens, token)
			}
		}
		for token := range parser.Parse([]uint16{}, false) {
			tokens = append(tokens, token)
		}
	}
	var builder strings.Builder
	builder.WriteString(`{"lineStarts":[`)
	for index, offset := range lineCounter.LineStarts {
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.Itoa(offset))
	}
	builder.WriteString(`],"tokens":`)
	writeTokenList(&builder, tokens)
	builder.WriteByte('}')
	return builder.String()
}

// The serializer writes what JSON.stringify writes for the canonicalized upstream object: keys sorted by
// code unit, absent fields left out, an absent list (nil) never confused with an empty one.

func writeTokenList(builder *strings.Builder, tokens []*Token) {
	if tokens == nil {
		builder.WriteString("null")
		return
	}
	builder.WriteByte('[')
	for index, token := range tokens {
		if index > 0 {
			builder.WriteByte(',')
		}
		writeToken(builder, token)
	}
	builder.WriteByte(']')
}

type objectWriter struct {
	builder *strings.Builder
	first   bool
}

func (writer *objectWriter) key(name string) {
	if !writer.first {
		writer.builder.WriteByte(',')
	}
	writer.first = false
	writer.builder.WriteString(`"` + name + `":`)
}

func writeToken(builder *strings.Builder, token *Token) {
	if token == nil {
		builder.WriteString("null")
		return
	}
	builder.WriteByte('{')
	writer := &objectWriter{builder: builder, first: true}
	isCollection := token.Type == "block-map" || token.Type == "block-seq" || token.Type == "flow-collection"
	if token.End != nil {
		writer.key("end")
		writeTokenList(builder, token.End)
	}
	if token.HasIndent() {
		writer.key("indent")
		builder.WriteString(strconv.Itoa(token.Indent))
	}
	if isCollection || token.Items != nil {
		writer.key("items")
		if token.Items == nil {
			builder.WriteString("null")
		} else {
			builder.WriteByte('[')
			for index, item := range token.Items {
				if index > 0 {
					builder.WriteByte(',')
				}
				writeItem(builder, item)
			}
			builder.WriteByte(']')
		}
	}
	if token.Type == "error" {
		writer.key("message")
		writeString(builder, sourceUnits(token.Message))
	}
	writer.key("offset")
	builder.WriteString(strconv.Itoa(token.Offset))
	if token.Type == "block-scalar" || token.Props != nil {
		writer.key("props")
		writeTokenList(builder, token.Props)
	}
	if token.HasSource() {
		writer.key("source")
		writeString(builder, token.Source)
	}
	if token.Type == "document" {
		writer.key("start")
		writeTokenList(builder, token.Start)
	} else if token.Type == "flow-collection" || token.FlowStart != nil {
		writer.key("start")
		writeToken(builder, token.FlowStart)
	}
	writer.key("type")
	writeString(builder, sourceUnits(token.Type))
	if token.Value != nil {
		writer.key("value")
		writeToken(builder, token.Value)
	}
	builder.WriteByte('}')
}

func writeItem(builder *strings.Builder, item *CollectionItem) {
	builder.WriteByte('{')
	writer := &objectWriter{builder: builder, first: true}
	if item.ExplicitKey {
		writer.key("explicitKey")
		builder.WriteString("true")
	}
	if item.HasKey() {
		writer.key("key")
		writeToken(builder, item.Key)
	} else if item.Key != nil {
		// A key the port holds without upstream's field: shown so the comparison fails.
		writer.key("key!")
		writeToken(builder, item.Key)
	}
	if item.Sep != nil {
		writer.key("sep")
		writeTokenList(builder, item.Sep)
	}
	writer.key("start")
	writeTokenList(builder, item.Start)
	if item.Value != nil {
		writer.key("value")
		writeToken(builder, item.Value)
	}
	builder.WriteByte('}')
}

// writeString is JSON.stringify of a string: the short escapes, other controls as \u00xx, lone
// surrogates as \udxxx, everything else as itself.
func writeString(builder *strings.Builder, units []uint16) {
	builder.WriteByte('"')
	for i := 0; i < len(units); i++ {
		unit := units[i]
		switch {
		case unit == '"':
			builder.WriteString(`\"`)
		case unit == '\\':
			builder.WriteString(`\\`)
		case unit == '\b':
			builder.WriteString(`\b`)
		case unit == '\f':
			builder.WriteString(`\f`)
		case unit == '\n':
			builder.WriteString(`\n`)
		case unit == '\r':
			builder.WriteString(`\r`)
		case unit == '\t':
			builder.WriteString(`\t`)
		case unit < 0x20:
			fmt.Fprintf(builder, `\u%04x`, unit)
		case isHighSurrogate(unit) && i+1 < len(units) && isLowSurrogate(units[i+1]):
			builder.WriteRune(utf16.DecodeRune(rune(unit), rune(units[i+1])))
			i++
		case isHighSurrogate(unit) || isLowSurrogate(unit):
			fmt.Fprintf(builder, `\u%04x`, unit)
		default:
			builder.WriteRune(rune(unit))
		}
	}
	builder.WriteByte('"')
}

// compareOutputs reports each input whose port output differs from upstream's, up to limit of them.
func compareOutputs(t *testing.T, label string, inputs []oracleInput, limit int) (failures int) {
	t.Helper()
	expected := upstreamOutputs(t, inputs)
	threw, errorTokens := 0, 0
	for index, input := range inputs {
		actual := portOutput(input)
		if actual != expected[index] {
			failures++
			if failures <= limit {
				t.Errorf("%s %q (chunk %d): %s", label, truncateInput(input.Text), input.Chunk,
					firstDifference(expected[index], actual))
			}
		}
		if expected[index] == "throw" {
			threw++
		} else if strings.Contains(expected[index], `"type":"error"`) {
			errorTokens++
		}
	}
	// A parse that throws is compared only as having thrown, so say how many did.
	t.Logf("%s: upstream threw on %d of %d and yielded error tokens on %d", label, threw, len(inputs), errorTokens)
	return failures
}

func firstDifference(expected string, actual string) string {
	index := 0
	for index < len(expected) && index < len(actual) && expected[index] == actual[index] {
		index++
	}
	from := max(index-120, 0)
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

// TestTokensMatchUpstream runs every fixture through both sides, whole.
func TestTokensMatchUpstream(t *testing.T) {
	t.Parallel()
	inputs := make([]oracleInput, 0, len(fixtures))
	for _, fixture := range fixtures {
		inputs = append(inputs, oracleInput{Text: normalize(fixture)})
	}
	failures := compareOutputs(t, "fixture", inputs, 60)
	t.Logf("%d of %d fixtures match upstream", len(inputs)-failures, len(inputs))
}

// TestChunkedTokensMatchUpstream feeds every fixture in chunks of 1, 2, 3, 5 and 8 code units, which
// splits lines, tokens and surrogate pairs at every point the lexer has to buffer across.
func TestChunkedTokensMatchUpstream(t *testing.T) {
	t.Parallel()
	var inputs []oracleInput
	for _, fixture := range fixtures {
		for _, chunk := range []int{1, 2, 3, 5, 8} {
			inputs = append(inputs, oracleInput{Text: normalize(fixture), Chunk: chunk})
		}
	}
	failures := compareOutputs(t, "fixture", inputs, 60)
	t.Logf("%d of %d chunked parses match upstream", len(inputs)-failures, len(inputs))
}

// TestGeneratedTokensMatchUpstream compares inputs drawn from YAML's indicator characters, spaces,
// tabs, line breaks and a few letters, where the hand-written fixtures cannot reach every combination.
func TestGeneratedTokensMatchUpstream(t *testing.T) {
	t.Parallel()
	count := 20000
	if testing.Short() {
		count = 2000
	}
	inputs := make([]oracleInput, 0, count)
	for _, text := range generatedInputs(count) {
		inputs = append(inputs, oracleInput{Text: text})
	}
	failures := compareOutputs(t, "generated", inputs, 40)
	t.Logf("%d of %d generated inputs match upstream", len(inputs)-failures, len(inputs))
}

// TestOracleCanFail proves the comparison sees a difference: the tokens of one input compared against
// upstream's for another must differ.
func TestOracleCanFail(t *testing.T) {
	t.Parallel()
	expected := upstreamOutputs(t, []oracleInput{{Text: "a: b"}, {Text: "a: [b]"}})
	if portOutput(oracleInput{Text: "a:  b"}) == expected[0] {
		t.Fatal("the comparison found no difference between the tokens of a: b and a:  b")
	}
	if portOutput(oracleInput{Text: "a: [b,]"}) == expected[1] {
		t.Fatal("the comparison found no difference between the tokens of a: [b] and a: [b,]")
	}
}

// TestYamlTestSuiteMatchesUpstream compares every case of the yaml-test-suite package in the fork, whole
// and in chunks of 3.
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
	var inputs []oracleInput
	for _, text := range texts {
		inputs = append(inputs, oracleInput{Text: text}, oracleInput{Text: text, Chunk: 3})
	}
	failures := compareOutputs(t, "suite", inputs, 40)
	t.Logf("%d of %d yaml-test-suite parses (%d cases, whole and chunked) match upstream",
		len(inputs)-failures, len(inputs), len(texts))
}

// TestCorpusMatchesUpstream compares every .yaml and .yml file under COHERE_YAML_CORPUS, a measuring run.
func TestCorpusMatchesUpstream(t *testing.T) {
	t.Parallel()
	root := os.Getenv("COHERE_YAML_CORPUS")
	if root == "" {
		t.Skip("set COHERE_YAML_CORPUS to a directory to compare every .yaml and .yml file under it")
	}
	var paths []string
	var inputs []oracleInput
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
			inputs = append(inputs, oracleInput{Text: normalize(string(source))})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := upstreamOutputs(t, inputs)
	failures := 0
	for index, input := range inputs {
		if actual := portOutput(input); actual != expected[index] {
			failures++
			if failures <= 40 {
				t.Errorf("%s: %s", paths[index], firstDifference(expected[index], actual))
			}
		}
	}
	t.Logf("%d of %d files match upstream's tokens", len(inputs)-failures, len(inputs))
}
