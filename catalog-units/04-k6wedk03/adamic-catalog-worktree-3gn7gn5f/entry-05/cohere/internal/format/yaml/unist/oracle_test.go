package unist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
)

// The parse oracle: Node runs yaml-unist-parser's parse(text, { uniqueKeys: false }) from the fork's
// node_modules and writes the whole tree as JSON, root.comments included, every object under its
// identity: an id per object (a comment reached from root.comments and from the node it is attached to
// is one object, one id), its _parent's id, and the keys it has. The test parses the same text with
// Parse and walks both trees in step, comparing every field of every node: type, position (offsets
// converted to bytes as the printer's tree loader converts them), value, chomping, indent, markers,
// name, parameters, every child list with null and [] told apart, and the comment slots. The ids prove
// the shape is the same graph: the same node reached twice on one side is the same node reached twice on
// the other, and every Parent is the node upstream's _parent is. Where upstream throws, Parse must fail
// with the same name and message, and for a YAMLSyntaxError the same position.

// forkRoot is the Prettier fork whose node_modules the oracle imports from (internal/corpus): a skip naming
// COHERE_PRETTIER_FORK when it is unset, and a failure when it is set but lacks node_modules/yaml-unist-parser.
func forkRoot(t testing.TB) string {
	t.Helper()
	corpus.Resolve(t, "prettier-fork:node_modules/yaml-unist-parser")
	return corpus.PrettierFork.Root(t)
}

const oracleScript = `
import { parse } from "FORK/node_modules/yaml-unist-parser/dist/index.mjs";
import { readFileSync } from "node:fs";

const inputs = JSON.parse(readFileSync(0, "utf8"));
const outputs = inputs.map((text) => {
  let root;
  try {
    root = parse(text, { uniqueKeys: false });
  } catch (error) {
    return { error: { name: String(error && error.name), message: String(error && error.message), position: error && error.position } };
  }
  const ids = new Map();
  const idOf = (node) => {
    if (!ids.has(node)) ids.set(node, ids.size);
    return ids.get(node);
  };
  const serialize = (node) => {
    if (node === null) return null;
    if (node === undefined) return { undefined: true };
    const parent = node._parent;
    const output = { id: idOf(node), parent: parent === undefined ? -2 : parent === null ? -1 : idOf(parent), keys: Object.keys(node) };
    for (const [key, value] of Object.entries(node)) {
      if (key === "position") output.position = value;
      else if (Array.isArray(value)) output[key] = value.map((each) => (each !== null && typeof each === "object" ? serialize(each) : each));
      else if (value !== null && typeof value === "object") output[key] = serialize(value);
      else output[key] = value;
    }
    return output;
  };
  return { root: serialize(root) };
});
process.stdout.write(JSON.stringify(outputs));
`

type jsonPoint struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

type jsonPosition struct {
	Start jsonPoint `json:"start"`
	End   jsonPoint `json:"end"`
}

// jsonNode is a node as the oracle writes it. Unknown fields fail the decode, so a field upstream adds
// that unist.Node does not carry is noticed rather than dropped.
type jsonNode struct {
	ID        int          `json:"id"`
	Parent    int          `json:"parent"`
	Keys      []string     `json:"keys"`
	Undefined bool         `json:"undefined"`
	Type      string       `json:"type"`
	Position  jsonPosition `json:"position"`

	Children            []*jsonNode `json:"children"`
	Comments            []*jsonNode `json:"comments"`
	Value               *string     `json:"value"`
	Tag                 *jsonNode   `json:"tag"`
	Anchor              *jsonNode   `json:"anchor"`
	MiddleComments      []*jsonNode `json:"middleComments"`
	LeadingComments     []*jsonNode `json:"leadingComments"`
	TrailingComment     *jsonNode   `json:"trailingComment"`
	EndComments         []*jsonNode `json:"endComments"`
	Chomping            *string     `json:"chomping"`
	Indent              *int        `json:"indent"`
	IndicatorComment    *jsonNode   `json:"indicatorComment"`
	DirectivesEndMarker *bool       `json:"directivesEndMarker"`
	DocumentEndMarker   *bool       `json:"documentEndMarker"`
	Name                *string     `json:"name"`
	Parameters          []string    `json:"parameters"`
}

type jsonError struct {
	Name     string        `json:"name"`
	Message  string        `json:"message"`
	Position *jsonPosition `json:"position"`
}

type oracleResult struct {
	Root  json.RawMessage `json:"root"`
	Error *jsonError      `json:"error"`
}

// oracleParse runs the real parser over every text, a Node process per few megabytes.
func oracleParse(t testing.TB, texts []string) []oracleResult {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the parse oracle needs it")
	}
	script := filepath.Join(t.TempDir(), "parse.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(oracleScript, "FORK", forkRoot(t))), 0o644); err != nil {
		t.Fatal(err)
	}
	results := make([]oracleResult, 0, len(texts))
	const batchBytes = 4 << 20
	for start := 0; start < len(texts); {
		end, size := start, 0
		for end < len(texts) && (end == start || size+len(texts[end]) <= batchBytes) {
			size += len(texts[end])
			end++
		}
		encoded, err := json.Marshal(texts[start:end])
		if err != nil {
			t.Fatal(err)
		}
		command := exec.Command("node", "--stack-size=65500", script)
		command.Stdin = bytes.NewReader(encoded)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			t.Fatalf("the parse oracle failed: %v\n%s", err, stderr.String())
		}
		var batch []oracleResult
		if err := json.Unmarshal(output, &batch); err != nil {
			t.Fatal(err)
		}
		if len(batch) != end-start {
			t.Fatalf("the parse oracle returned %d results for %d inputs", len(batch), end-start)
		}
		results = append(results, batch...)
		start = end
	}
	return results
}

// comparison walks the two trees in step.
type comparison struct {
	offsets unitOffsets
	// byID is the port's node for each oracle id, and idOf the oracle id for each port node.
	byID map[int]*Node
	idOf map[*Node]int
	// parents are the port's nodes with the oracle's parent id for each.
	parents     map[*Node]int
	differences []string
}

func (c *comparison) differ(path string, format string, arguments ...any) {
	c.differences = append(c.differences, path+": "+fmt.Sprintf(format, arguments...))
}

// jsonString is the port's string as the oracle's arrives: JSON.stringify escapes a lone surrogate,
// which encoding/json decodes as U+FFFD. The port's strings hold one as WTF-8 (three bytes, ED then A0
// to BF), so those become U+FFFD too; any other invalid byte is left for the comparison to catch.
func jsonString(text string) string {
	if utf8.ValidString(text) {
		return text
	}
	var builder strings.Builder
	for index := 0; index < len(text); {
		if index+2 < len(text) && text[index] == 0xED && text[index+1] >= 0xA0 && text[index+1] <= 0xBF {
			builder.WriteRune(utf8.RuneError)
			index += 3
			continue
		}
		builder.WriteByte(text[index])
		index++
	}
	return builder.String()
}

func (c *comparison) point(path string, name string, mine Point, theirs jsonPoint) {
	expected := Point{Line: theirs.Line, Column: theirs.Column, Offset: c.offsets.byteOffset(theirs.Offset)}
	if mine != expected {
		c.differ(path, "%s is %+v, upstream's %+v (unit offset %d)", name, mine, expected, theirs.Offset)
	}
}

func (c *comparison) list(path string, name string, mine []*Node, theirs []*jsonNode, present bool) {
	if !present {
		if mine != nil {
			c.differ(path, "%s is %d nodes where upstream has no such field", name, len(mine))
		}
		return
	}
	if mine == nil || theirs == nil {
		// encoding/json decodes null as a nil slice and [] as an empty one.
		if (mine == nil) != (theirs == nil) {
			c.differ(path, "%s: one side is null, the other an array (the port's is nil: %v)", name, mine == nil)
		}
		return
	}
	if len(mine) != len(theirs) {
		c.differ(path, "%s has %d nodes, upstream's %d", name, len(mine), len(theirs))
		return
	}
	for index := range mine {
		c.node(fmt.Sprintf("%s.%s[%d]", path, name, index), mine[index], theirs[index])
	}
}

func (c *comparison) node(path string, mine *Node, theirs *jsonNode) {
	if theirs == nil || theirs.Undefined {
		if mine != nil {
			c.differ(path, "is a %s where upstream's is null", mine.NodeType)
		}
		return
	}
	if mine == nil {
		c.differ(path, "is nil where upstream's is a %s", theirs.Type)
		return
	}
	path = path + "<" + theirs.Type + ">"

	// Identity: one object upstream is one node here, and the other way round.
	if known, seen := c.byID[theirs.ID]; seen {
		if known != mine {
			c.differ(path, "upstream reaches its object #%d again, the port a different node", theirs.ID)
		}
	} else {
		c.byID[theirs.ID] = mine
	}
	if known, seen := c.idOf[mine]; seen {
		if known != theirs.ID {
			c.differ(path, "the port reaches one node as upstream's objects #%d and #%d", known, theirs.ID)
		}
	} else {
		c.idOf[mine] = theirs.ID
	}
	c.parents[mine] = theirs.Parent

	has := func(key string) bool { return slices.Contains(theirs.Keys, key) }
	if mine.NodeType != theirs.Type {
		c.differ(path, "type is %q", mine.NodeType)
	}
	c.point(path, "start", mine.Position.Start, theirs.Position.Start)
	c.point(path, "end", mine.Position.End, theirs.Position.End)
	if hasLeadingCommentsField(mine.NodeType) != has("leadingComments") || hasTrailingCommentField(mine.NodeType) != has("trailingComment") ||
		hasChildrenField(mine.NodeType) != has("children") {
		c.differ(path, "the port's field predicates disagree with upstream's keys %v", theirs.Keys)
	}

	stringField := func(name string, mineValue string, theirsValue *string) {
		expected := ""
		if theirsValue != nil {
			expected = *theirsValue
		}
		if jsonString(mineValue) != expected {
			c.differ(path, "%s is %q, upstream's %q", name, mineValue, expected)
		}
	}
	stringField("value", mine.Value, theirs.Value)
	stringField("chomping", mine.Chomping, theirs.Chomping)
	stringField("name", mine.Name, theirs.Name)
	if (mine.Indent == nil) != (theirs.Indent == nil) || mine.Indent != nil && *mine.Indent != *theirs.Indent {
		c.differ(path, "indent differs")
	}
	boolField := func(name string, mineValue bool, theirsValue *bool) {
		if mineValue != (theirsValue != nil && *theirsValue) {
			c.differ(path, "%s is %v", name, mineValue)
		}
	}
	boolField("directivesEndMarker", mine.DirectivesEndMarker, theirs.DirectivesEndMarker)
	boolField("documentEndMarker", mine.DocumentEndMarker, theirs.DocumentEndMarker)
	if has("parameters") != (mine.Parameters != nil) || !slices.Equal(mine.Parameters, theirs.Parameters) {
		c.differ(path, "parameters are %q, upstream's %q", mine.Parameters, theirs.Parameters)
	}

	c.list(path, "children", mine.Children, theirs.Children, has("children"))
	c.node(path+".tag", mine.Tag, theirs.Tag)
	c.node(path+".anchor", mine.Anchor, theirs.Anchor)
	c.list(path, "middleComments", mine.MiddleComments, theirs.MiddleComments, has("middleComments"))
	c.list(path, "leadingComments", mine.LeadingComments, theirs.LeadingComments, has("leadingComments"))
	c.node(path+".indicatorComment", mine.IndicatorComment, theirs.IndicatorComment)
	c.node(path+".trailingComment", mine.TrailingComment, theirs.TrailingComment)
	c.list(path, "endComments", mine.EndComments, theirs.EndComments, has("endComments"))
	c.list(path, "comments", mine.Comments, theirs.Comments, has("comments"))
}

// compareParse compares Parse's result for the text with the oracle's and returns the differences.
func compareParse(text string, expected oracleResult) []string {
	root, err := Parse(text, nil)
	offsets := newUnitOffsets(text, nil)
	if expected.Error != nil {
		if err == nil {
			return []string{fmt.Sprintf("upstream throws %s: %s; the port parses", expected.Error.Name, expected.Error.Message)}
		}
		name, message := "", err.Error()
		var syntaxError *SyntaxError
		var thrown *ThrownError
		switch {
		case errors.As(err, &syntaxError):
			name, message = "YAMLSyntaxError", syntaxError.Message
			if expected.Error.Position == nil {
				return []string{"upstream's YAMLSyntaxError has no position"}
			}
			want := Position{
				Start: Point{Line: expected.Error.Position.Start.Line, Column: expected.Error.Position.Start.Column, Offset: offsets.byteOffset(expected.Error.Position.Start.Offset)},
				End:   Point{Line: expected.Error.Position.End.Line, Column: expected.Error.Position.End.Column, Offset: offsets.byteOffset(expected.Error.Position.End.Offset)},
			}
			if syntaxError.Position != want {
				return []string{fmt.Sprintf("the error's position is %+v, upstream's %+v", syntaxError.Position, want)}
			}
		case errors.As(err, &thrown):
			name, message = thrown.Name, thrown.Message
		}
		if name != expected.Error.Name || message != expected.Error.Message {
			return []string{fmt.Sprintf("the port fails with %s: %s; upstream throws %s: %s", name, message, expected.Error.Name, expected.Error.Message)}
		}
		return nil
	}
	if err != nil {
		return []string{fmt.Sprintf("the port fails (%v); upstream parses", err)}
	}
	decoder := json.NewDecoder(bytes.NewReader(expected.Root))
	decoder.DisallowUnknownFields()
	var theirs jsonNode
	if err := decoder.Decode(&theirs); err != nil {
		return []string{fmt.Sprintf("decoding upstream's tree: %v", err)}
	}
	c := &comparison{offsets: offsets, byID: map[int]*Node{}, idOf: map[*Node]int{}, parents: map[*Node]int{}}
	c.node("root", root, &theirs)
	for node, parentID := range c.parents {
		switch parentID {
		case -1, -2:
			// null (the root), or never defined: nil either way here.
			if node.Parent != nil {
				c.differ(node.NodeType, "has a parent where upstream's has none")
			}
		default:
			if expected, known := c.byID[parentID]; !known || node.Parent != expected {
				c.differ(node.NodeType+"@"+fmt.Sprint(node.Position.Start), "parent is not upstream's _parent")
			}
		}
	}
	return c.differences
}

// compareAll parses every text both ways and reports the first few that differ.
func compareAll(t *testing.T, label string, names []string, texts []string, maxReported int) (failures int, errored int) {
	t.Helper()
	expected := oracleParse(t, texts)
	for index, text := range texts {
		if expected[index].Error != nil {
			errored++
		}
		differences := compareParse(text, expected[index])
		if len(differences) == 0 {
			continue
		}
		failures++
		if failures <= maxReported {
			name := fmt.Sprintf("%q", text)
			if names != nil {
				name = names[index]
			} else if len(name) > 200 {
				name = name[:200] + "..."
			}
			if len(differences) > 5 {
				differences = differences[:5]
			}
			t.Errorf("%s %s:\n  %s", label, name, strings.Join(differences, "\n  "))
		}
	}
	return failures, errored
}
