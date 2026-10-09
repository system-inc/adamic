package printing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/doc"
)

/*
 * The print core against upstream's, through a toy language.
 *
 * No real printer exists yet to test the core through, and comment attachment is where a port drifts
 * silently: a comment on the wrong node prints in a plausible wrong place. So this generates a small
 * language, parenthesised lists of atoms with block and line comments at random gaps, and runs the same
 * AST and the same tiny printer through upstream's real printAstToDoc in node and through this core.
 * Two things are compared: the formatted text, and for every comment where it attached (leading,
 * trailing or dangling, and on which node). The second localizes a failure the first only reports.
 */

type toyNode struct {
	Kind                    string     `json:"type"`
	From                    int        `json:"start"`
	To                      int        `json:"end"`
	Value                   string     `json:"value,omitempty"`
	Label                   *toyNode   `json:"label,omitempty"`
	Children                []*toyNode `json:"children,omitempty"`
	CommentFields[*toyNode] `json:"-"`
}

func (node *toyNode) Type() string { return node.Kind }
func (node *toyNode) Field(name string) any {
	switch name {
	case "children":
		return node.Children
	case "label":
		if node.Label == nil {
			return nil
		}
		return node.Label
	case "comments":
		return node.Comments
	}
	return nil
}
func (node *toyNode) CommentData() *CommentFields[*toyNode] { return &node.CommentFields }

// toySource builds text and AST together, so positions are exact by construction.
type toySource struct {
	random   *rand.Rand
	text     strings.Builder
	comments []*toyNode
	counter  int
}

// gap writes one to three items. Runs matter: two comments on one line are what the same-line scans in
// isOwnLineComment and isEndOfLineComment exist for, a blank line next to a comment is what the
// blank-line hardlines exist for, and a comma is a gap character that is not whitespace or "(", which is
// what makes breakTies decide anything. A first version wrote one item per gap, and five mutations of
// the core survived it.
func (source *toySource) gap() {
	for count := 1 + source.random.Intn(3); count > 0; count-- {
		source.gapItem()
	}
}

func (source *toySource) gapItem() {
	switch source.random.Intn(9) {
	case 0:
		source.counter++
		start := source.text.Len()
		source.text.WriteString(fmt.Sprintf("/*b%d*/", source.counter))
		source.comments = append(source.comments, &toyNode{Kind: "Block", From: start, To: source.text.Len()})
		source.text.WriteString([]string{" ", "\n", ""}[source.random.Intn(3)])
	case 1:
		source.counter++
		start := source.text.Len()
		source.text.WriteString(fmt.Sprintf("//l%d", source.counter))
		source.comments = append(source.comments, &toyNode{Kind: "Line", From: start, To: source.text.Len()})
		source.text.WriteString("\n")
	case 2:
		source.text.WriteString("\n")
	case 3:
		source.text.WriteString("\n\n")
	case 4:
		source.text.WriteString(", ")
	case 5:
		source.text.WriteString(" ")
	}
}

func (source *toySource) node(depth int) *toyNode {
	if depth > 3 || source.random.Intn(3) == 0 {
		start := source.text.Len()
		value := []string{"a", "bb", "ccc", "identifier", "x"}[source.random.Intn(5)]
		source.text.WriteString(value)
		return &toyNode{Kind: "Atom", From: start, To: source.text.Len(), Value: value}
	}
	start := source.text.Len()
	source.text.WriteString("(")
	node := &toyNode{Kind: "List", From: start}
	// The label is written before the children and listed after them in the visitor keys, so position
	// order and key order disagree, which is what sortedChildNodes' sort is for.
	if source.random.Intn(3) == 0 {
		labelStart := source.text.Len()
		source.text.WriteString("#tag")
		node.Label = &toyNode{Kind: "Atom", From: labelStart, To: source.text.Len(), Value: "#tag"}
		source.text.WriteString(" ")
	}
	count := source.random.Intn(4)
	for index := 0; index < count; index++ {
		source.gap()
		if index > 0 {
			source.text.WriteString(" ")
		}
		node.Children = append(node.Children, source.node(depth+1))
	}
	source.gap()
	source.text.WriteString(")")
	node.To = source.text.Len()
	return node
}

func toyProgram(random *rand.Rand) (*toyNode, []*toyNode, string) {
	source := &toySource{random: random}
	program := &toyNode{Kind: "Program"}
	count := 1 + random.Intn(3)
	for index := 0; index < count; index++ {
		source.gap()
		program.Children = append(program.Children, source.node(0))
		source.text.WriteString("\n")
	}
	source.gap()
	program.To = source.text.Len()
	return program, source.comments, source.text.String()
}

func toyPrinter() *Printer[*toyNode] {
	printer := &Printer[*toyNode]{}
	printer.LocStart = func(node *toyNode) int { return node.From }
	printer.LocEnd = func(node *toyNode) int { return node.To }
	printer.VisitorKeys = func(node *toyNode) []string {
		switch node.Kind {
		case "List":
			return []string{"children", "label"}
		case "Program":
			return []string{"children"}
		}
		return nil
	}
	printer.CanAttachComment = func(node *toyNode, _ []*toyNode) bool { return node.Kind == "Atom" || node.Kind == "List" }
	printer.IsBlockComment = func(comment *toyNode) bool { return comment.Kind == "Block" }
	// A handler, so the hooks are exercised and placement is observable: an end-of-line block comment
	// inside a list dangles on that list. Without one, a mutation that misclassified end-of-line comments
	// survived, because default attachment sends both placements to the same node; the JavaScript
	// handlers branch on placement, so the toy has to as well.
	printer.HandleComments.EndOfLine = func(context *CommentContext[*toyNode]) bool {
		if context.HasEnclosing && context.Enclosing.Kind == "List" && context.Comment.Kind == "Block" {
			AddDanglingComment(context.Enclosing, context.Comment, "")
			return true
		}
		return false
	}
	printer.PrintComment = func(path *AstPath[*toyNode], options *Options[*toyNode]) doc.Doc {
		comment, _ := path.Node()
		return doc.Text(options.OriginalText[comment.From:comment.To])
	}
	printer.Print = func(path *AstPath[*toyNode], options *Options[*toyNode], print PrintFunc, _ any) doc.Doc {
		node, _ := path.Node()
		children := Map(path, func(*AstPath[*toyNode], int, any) doc.Doc { return print(nil, nil) }, "children")
		switch node.Kind {
		case "Program":
			return doc.Concat{doc.Join(doc.Hardline, children), PrintDanglingComments(path, options, DanglingOptions[*toyNode]{}), doc.Hardline}
		case "List":
			var label doc.Doc = doc.Text("")
			if node.Label != nil {
				label = doc.Concat{print("label", nil), doc.Text(" ")}
			}
			if len(children) == 0 {
				return doc.Concat{doc.Text("("), label, PrintDanglingComments(path, options, DanglingOptions[*toyNode]{}), doc.Text(")")}
			}
			return doc.NewGroup(doc.Concat{
				doc.Text("("), label,
				doc.NewIndent(doc.Concat{doc.Softline, doc.Join(doc.LineDoc, children)}),
				PrintDanglingComments(path, options, DanglingOptions[*toyNode]{}),
				doc.Softline,
				doc.Text(")"),
			}, doc.GroupOptions{})
		default:
			return doc.Text(node.Value)
		}
	}
	return printer
}

const toyUpstreamScript = `
import { printAstToDoc } from "./src/main/ast-to-doc.js";
import { printDanglingComments } from "./src/main/comments/print.js";
import { addDanglingComment } from "./src/main/comments/utilities.js";
import { builders as b, printer as docPrinter } from "./src/document/public.js";
process.stdin.setEncoding("utf8");
let input = ""; process.stdin.on("data", (chunk) => input += chunk);
process.stdin.on("end", async () => {
  const results = [];
  for (const { ast, comments, text, width } of JSON.parse(input)) {
    ast.comments = comments;
    const visitorKeys = (node) => node.type === "List" ? ["children", "label"] : node.type === "Program" ? ["children"] : [];
    const printer = {
      features: {},
      canAttachComment: (node) => node.type === "Atom" || node.type === "List",
      isBlockComment: (comment) => comment.type === "Block",
      handleComments: {
        endOfLine(comment) {
          if (comment.enclosingNode?.type === "List" && comment.type === "Block") {
            addDanglingComment(comment.enclosingNode, comment);
            return true;
          }
          return false;
        },
      },
      printComment: (path, options) => options.originalText.slice(path.node.start, path.node.end),
      print(path, options, print) {
        const node = path.node;
        // omitempty drops an empty children array from the JSON, so a missing one means none.
        const children = node.children ? path.map(print, "children") : [];
        if (node.type === "Program") return [b.join(b.hardline, children), printDanglingComments(path, options), b.hardline];
        if (node.type === "List") {
          const label = node.label ? [print("label"), " "] : "";
          if (children.length === 0) return ["(", label, printDanglingComments(path, options), ")"];
          return b.group(["(", label, b.indent([b.softline, b.join(b.line, children)]), printDanglingComments(path, options), b.softline, ")"]);
        }
        return node.value;
      },
    };
    const options = {
      printer, originalText: text, locStart: (n) => n.start, locEnd: (n) => n.end,
      getVisitorKeys: visitorKeys, embeddedLanguageFormatting: "off", printWidth: width, tabWidth: 2,
    };
    try {
      const doc = await printAstToDoc(ast, options);
      const formatted = docPrinter.printDocToString(doc, { printWidth: width, tabWidth: 2, useTabs: false, endOfLine: "lf" }).formatted;
      const attached = [];
      const walk = (node) => {
        for (const comment of node.comments ?? []) attached.push([comment.start, comment.leading ? "leading" : comment.trailing ? "trailing" : "dangling", node.type, node.start]);
        if (node.label) walk(node.label);
        for (const child of node.children ?? []) walk(child);
      };
      walk(ast);
      attached.sort((x, y) => x[0] - y[0]);
      results.push({ formatted, attached });
    } catch (error) {
      results.push({ error: String(error) });
    }
  }
  process.stdout.write(JSON.stringify(results));
});
`

type toyCase struct {
	Ast      *toyNode   `json:"ast"`
	Comments []*toyNode `json:"comments"`
	Text     string     `json:"text"`
	Width    int        `json:"width"`
}

type toyResult struct {
	Formatted string  `json:"formatted"`
	Attached  [][]any `json:"attached"`
	Error     string  `json:"error"`
}

func attachedSummary(root *toyNode) [][]any {
	// Empty, not nil: a nil slice encodes as null and upstream's empty array as [], and every comment-free
	// program read as a mismatch until this said so.
	attached := [][]any{}
	var walk func(*toyNode)
	walk = func(node *toyNode) {
		for _, comment := range node.Comments {
			kind := "dangling"
			if comment.Leading {
				kind = "leading"
			} else if comment.Trailing {
				kind = "trailing"
			}
			attached = append(attached, []any{float64(comment.From), kind, node.Kind, float64(node.From)})
		}
		if node.Label != nil {
			walk(node.Label)
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	for left := 1; left < len(attached); left++ {
		for right := left; right > 0 && attached[right][0].(float64) < attached[right-1][0].(float64); right-- {
			attached[right], attached[right-1] = attached[right-1], attached[right]
		}
	}
	return attached
}

// TestCoreAgreesWithUpstream is the differential. Off unless COHERE_PRETTIER_FORK names the fork.
func TestCoreAgreesWithUpstream(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)
	const count = 3000
	random := rand.New(rand.NewSource(20261002))
	cases := make([]toyCase, count)
	for index := range cases {
		ast, comments, text := toyProgram(random)
		cases[index] = toyCase{Ast: ast, Comments: comments, Text: text, Width: 6 + random.Intn(30)}
	}
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}

	command := exec.Command("node", "--input-type=module", "-e", toyUpstreamScript)
	command.Dir = root
	command.Stdin = bytes.NewReader(encoded)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	var upstream []toyResult
	if err := json.Unmarshal(stdout.Bytes(), &upstream); err != nil {
		t.Fatal(err)
	}

	// The Go side gets fresh ASTs, decoded from the same JSON, so neither side sees the other's
	// attachment state.
	var fresh []toyCase
	if err := json.Unmarshal(encoded, &fresh); err != nil {
		t.Fatal(err)
	}

	commentsTotal, upstreamErrors, textMismatches, attachMismatches := 0, 0, 0, 0
	for index, testCase := range fresh {
		commentsTotal += len(testCase.Comments)
		if upstream[index].Error != "" {
			upstreamErrors++
			if upstreamErrors <= 3 {
				t.Logf("upstream refused case %d: %s", index, upstream[index].Error)
			}
			continue
		}
		options := &Options[*toyNode]{Printer: toyPrinter(), OriginalText: testCase.Text, EmbeddedLanguageFormatting: "off"}
		printed, err := PrintAstToDoc(testCase.Ast, testCase.Comments, options)
		if err != nil {
			t.Errorf("case %d: %v\ntext %q", index, err, testCase.Text)
			continue
		}
		formatted := doc.Print(printed, doc.Options{PrintWidth: testCase.Width, TabWidth: 2})

		attachedGo, _ := json.Marshal(attachedSummary(testCase.Ast))
		attachedUpstream, _ := json.Marshal(upstream[index].Attached)
		if string(attachedGo) != string(attachedUpstream) {
			attachMismatches++
			if attachMismatches <= 3 {
				t.Errorf("case %d attachment\ntext %q\nwant %s\ngot  %s", index, testCase.Text, attachedUpstream, attachedGo)
			}
		}
		if formatted != upstream[index].Formatted {
			textMismatches++
			if textMismatches <= 3 {
				t.Errorf("case %d output\ntext %q\nwant %q\ngot  %q", index, testCase.Text, upstream[index].Formatted, formatted)
			}
		}
	}
	t.Logf("%d programs, %d comments, %d upstream refused, %d attachment mismatches, %d output mismatches",
		count, commentsTotal, upstreamErrors, attachMismatches, textMismatches)
	if upstreamErrors > count/10 {
		t.Errorf("upstream refused %d of %d programs, so the comparison is thinner than it looks", upstreamErrors, count)
	}
}
