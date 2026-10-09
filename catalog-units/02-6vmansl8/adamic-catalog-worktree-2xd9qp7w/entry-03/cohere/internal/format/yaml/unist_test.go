package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/yaml/unist"
)

// The tree loader. The Go parser is being ported in parallel, so these tests print trees the real one
// builds: Node runs yaml-unist-parser's parse(text, { uniqueKeys: false }) from the fork's node_modules,
// deletes root.comments as parser-yaml.js does, and writes the tree as JSON. _parent is non-enumerable,
// so JSON leaves it out and the loader sets Parent the way defineParents does. Offsets come back in
// UTF-16 units and are converted to bytes against the text, the conversion layer 3 does once.

// forkRoot is the Prettier fork whose node_modules the oracle imports from (internal/corpus): a skip naming
// COHERE_PRETTIER_FORK when it is unset, and a failure when it is set but lacks node_modules/yaml-unist-parser.
func forkRoot(t testing.TB) string {
	t.Helper()
	corpus.Resolve(t, "prettier-fork:node_modules/yaml-unist-parser")
	return corpus.PrettierFork.Root(t)
}

const parseScript = `
import { parse } from "FORK/node_modules/yaml-unist-parser/dist/index.mjs";
import { readFileSync } from "node:fs";

const inputs = JSON.parse(readFileSync(0, "utf8"));
const outputs = inputs.map((text) => {
  try {
    const root = parse(text, { uniqueKeys: false });
    delete root.comments;
    return { root };
  } catch (error) {
    return { error: String(error && error.message) };
  }
});
process.stdout.write(JSON.stringify(outputs));
`

// jsonPoint and jsonNode are the tree as Node writes it. Unknown fields fail the decode, so a field the
// parser adds that unist.Node does not carry is noticed rather than dropped.
type jsonPoint struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

type jsonNode struct {
	Type     string `json:"type"`
	Position struct {
		Start jsonPoint `json:"start"`
		End   jsonPoint `json:"end"`
	} `json:"position"`
	Children            []*jsonNode `json:"children"`
	Value               string      `json:"value"`
	Tag                 *jsonNode   `json:"tag"`
	Anchor              *jsonNode   `json:"anchor"`
	MiddleComments      []*jsonNode `json:"middleComments"`
	LeadingComments     []*jsonNode `json:"leadingComments"`
	TrailingComment     *jsonNode   `json:"trailingComment"`
	EndComments         []*jsonNode `json:"endComments"`
	Chomping            string      `json:"chomping"`
	Indent              *int        `json:"indent"`
	IndicatorComment    *jsonNode   `json:"indicatorComment"`
	DirectivesEndMarker bool        `json:"directivesEndMarker"`
	DocumentEndMarker   bool        `json:"documentEndMarker"`
	Name                string      `json:"name"`
	Parameters          []string    `json:"parameters"`
}

type parseResult struct {
	Root  json.RawMessage `json:"root"`
	Error string          `json:"error"`
}

// parsed is one input's tree, or the parser's error.
type parsed struct {
	root *unist.Node
	err  string
}

// parseTrees parses every text with the real parser, a few Node processes for all of them.
func parseTrees(t testing.TB, texts []string) []parsed {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the parser oracle needs it")
	}
	script := filepath.Join(t.TempDir(), "parse.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(parseScript, "FORK", forkRoot(t))), 0o644); err != nil {
		t.Fatal(err)
	}

	results := make([]parsed, 0, len(texts))
	const batchBytes = 4 << 20
	for start := 0; start < len(texts); {
		end, size := start, 0
		for end < len(texts) && (end == start || size+len(texts[end]) <= batchBytes) {
			size += len(texts[end])
			end++
		}
		results = append(results, parseBatch(t, script, texts[start:end])...)
		start = end
	}
	return results
}

func parseBatch(t testing.TB, script string, texts []string) []parsed {
	t.Helper()
	encoded, err := json.Marshal(texts)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--stack-size=65500", script)
	command.Stdin = bytes.NewReader(encoded)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		t.Fatalf("the parser oracle failed: %v\n%s", err, stderr.String())
	}
	var raw []parseResult
	if err := json.Unmarshal(output, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) != len(texts) {
		t.Fatalf("the parser oracle returned %d results for %d inputs", len(raw), len(texts))
	}
	results := make([]parsed, len(texts))
	for index, result := range raw {
		if result.Root == nil {
			results[index] = parsed{err: result.Error}
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(result.Root))
		decoder.DisallowUnknownFields()
		var root jsonNode
		if err := decoder.Decode(&root); err != nil {
			t.Fatalf("decoding the tree of %q: %v", texts[index], err)
		}
		results[index] = parsed{root: loadTree(&root, newUnitOffsets(texts[index]), nil)}
	}
	return results
}

// unitOffsets maps a UTF-16 unit index to its byte offset in the text.
type unitOffsets []int

func newUnitOffsets(text string) unitOffsets {
	offsets := make(unitOffsets, 0, len(text)+1)
	for index, character := range text {
		offsets = append(offsets, index)
		if character > 0xFFFF {
			// The low surrogate of a pair sits at the pair's start too; no position upstream records
			// falls between the two units of one character.
			offsets = append(offsets, index)
		}
	}
	return append(offsets, len(text))
}

func (offsets unitOffsets) byteOffset(unit int) int {
	if unit < 0 || unit >= len(offsets) {
		panic(fmt.Sprintf("offset %d is outside the text's %d units", unit, len(offsets)-1))
	}
	return offsets[unit]
}

// loadTree converts the JSON tree, setting Parent as defineParents does.
func loadTree(node *jsonNode, offsets unitOffsets, parent *unist.Node) *unist.Node {
	if node == nil {
		return nil
	}
	loaded := &unist.Node{
		NodeType: node.Type,
		Position: unist.Position{
			Start: unist.Point{Line: node.Position.Start.Line, Column: node.Position.Start.Column, Offset: offsets.byteOffset(node.Position.Start.Offset)},
			End:   unist.Point{Line: node.Position.End.Line, Column: node.Position.End.Column, Offset: offsets.byteOffset(node.Position.End.Offset)},
		},
		Parent:              parent,
		Value:               node.Value,
		Chomping:            node.Chomping,
		Indent:              node.Indent,
		DirectivesEndMarker: node.DirectivesEndMarker,
		DocumentEndMarker:   node.DocumentEndMarker,
		Name:                node.Name,
		Parameters:          node.Parameters,
	}
	list := func(nodes []*jsonNode) []*unist.Node {
		if nodes == nil {
			return nil
		}
		loadedNodes := make([]*unist.Node, len(nodes))
		for index, child := range nodes {
			loadedNodes[index] = loadTree(child, offsets, loaded)
		}
		return loadedNodes
	}
	loaded.Children = list(node.Children)
	loaded.Anchor = loadTree(node.Anchor, offsets, loaded)
	loaded.Tag = loadTree(node.Tag, offsets, loaded)
	loaded.LeadingComments = list(node.LeadingComments)
	loaded.MiddleComments = list(node.MiddleComments)
	loaded.IndicatorComment = loadTree(node.IndicatorComment, offsets, loaded)
	loaded.TrailingComment = loadTree(node.TrailingComment, offsets, loaded)
	loaded.EndComments = list(node.EndComments)
	return loaded
}

// normalizeInput is what Prettier's core does to a file before parsing: remove a byte order mark and
// normalize line endings to \n.
func normalizeInput(text string) (string, bool) {
	hasByteOrderMark := strings.HasPrefix(text, "\ufeff")
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	return text, hasByteOrderMark
}

// TestLoaderConvertsOffsets checks the loader on astral and multi-byte text, since every position the
// printer slices by goes through it.
func TestLoaderConvertsOffsets(t *testing.T) {
	t.Parallel()
	text := "é: 🍐\n'中': \"x\"\n"
	trees := parseTrees(t, []string{text})
	if trees[0].err != "" {
		t.Fatal(trees[0].err)
	}
	var values []string
	var walk func(node *unist.Node)
	walk = func(node *unist.Node) {
		switch node.NodeType {
		case "plain", "quoteSingle", "quoteDouble":
			values = append(values, text[node.Position.Start.Offset:node.Position.End.Offset])
		}
		for _, child := range node.Children {
			if child.Parent != node {
				t.Errorf("%s's parent is not set", child.NodeType)
			}
			walk(child)
		}
	}
	walk(trees[0].root)
	expected := []string{"é", "🍐", "'中'", "\"x\""}
	if strings.Join(values, "|") != strings.Join(expected, "|") {
		t.Errorf("scalars sliced by the loaded positions are %q, want %q", values, expected)
	}
	if !utf8.ValidString(text) {
		t.Fatal("the fixture is not UTF-8")
	}
}
