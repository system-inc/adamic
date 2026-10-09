package doc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
)

/*
 * The layout algorithm against upstream's printDocToString.
 *
 * One generator produces a random doc *spec*, and two constructors build it: buildSpec here into Go
 * docs, and a node script into upstream's own builders from src/document. Both print it, and the bytes
 * must be equal. Building from one spec on both sides is the point: a Go doc and a JavaScript doc
 * written separately could each be wrong in the same way, and the comparison would agree.
 *
 * The generator reaches the cases a printer port gets wrong quietly: groups that fit and break,
 * conditional groups, fills, IfBreak and IndentIfBreak on group ids printed before, after or never,
 * line suffixes flushed by boundaries and by the end of the document, trim, literal lines under
 * markAsRoot, align by width, string and dedent, groups shared by two parents, and text containing wide
 * characters, emoji, trailing whitespace and the empty string.
 */

// spec is a doc in a form both languages can build. Only the fields a kind uses are set.
type spec struct {
	Kind     string  `json:"k"`
	Text     string  `json:"v,omitempty"`
	Children []*spec `json:"c,omitempty"`
	Number   int     `json:"n,omitempty"`
	Flag     bool    `json:"f,omitempty"`
	ID       string  `json:"id,omitempty"`
	Ref      int     `json:"r,omitempty"`
}

type specGenerator struct {
	random   *rand.Rand
	groupIDs []string
	shared   int
	depth    int
}

var specWords = []string{"a", "bc", "def", "word", "longer", "identifier", "x", "", "中文", "😀", "trailing  ", "tab\t", "é"}

func (generator *specGenerator) text() *spec {
	return &spec{Kind: "text", Text: specWords[generator.random.Intn(len(specWords))]}
}

func (generator *specGenerator) line() *spec {
	return &spec{Kind: []string{"line", "softline", "hardline", "literalline", "hardlineWithoutBreakParent"}[generator.random.Intn(5)]}
}

func (generator *specGenerator) groupID() string {
	if len(generator.groupIDs) == 0 || generator.random.Intn(3) == 0 {
		id := fmt.Sprintf("g%d", len(generator.groupIDs))
		generator.groupIDs = append(generator.groupIDs, id)
		return id
	}
	return generator.groupIDs[generator.random.Intn(len(generator.groupIDs))]
}

func (generator *specGenerator) node() *spec {
	generator.depth++
	defer func() { generator.depth-- }()
	if generator.depth > 6 {
		if generator.random.Intn(3) == 0 {
			return generator.line()
		}
		return generator.text()
	}

	switch generator.random.Intn(17) {
	case 0, 1:
		return generator.text()
	case 2:
		return generator.line()
	case 3, 4:
		count := 1 + generator.random.Intn(4)
		children := make([]*spec, count)
		for index := range children {
			children[index] = generator.node()
		}
		return &spec{Kind: "concat", Children: children}
	case 5:
		return &spec{Kind: "indent", Children: []*spec{generator.node()}}
	case 6:
		switch generator.random.Intn(4) {
		case 0:
			return &spec{Kind: "alignWidth", Number: generator.random.Intn(5) - 1, Children: []*spec{generator.node()}}
		case 1:
			return &spec{Kind: "alignString", Text: []string{"> ", "- ", "  ", "", "→ ", "中 "}[generator.random.Intn(6)], Children: []*spec{generator.node()}}
		case 2:
			return &spec{Kind: "markAsRoot", Children: []*spec{generator.node()}}
		default:
			return &spec{Kind: "dedentToRoot", Children: []*spec{generator.node()}}
		}
	case 7, 8:
		group := &spec{Kind: "group", Flag: generator.random.Intn(6) == 0, Children: []*spec{generator.node()}}
		if generator.random.Intn(3) == 0 {
			group.ID = generator.groupID()
		}
		if generator.random.Intn(5) == 0 {
			generator.shared++
			group.Ref = generator.shared
			return &spec{Kind: "concat", Children: []*spec{group, {Kind: "sharedGroup", Ref: group.Ref}}}
		}
		return group
	case 9:
		count := 2 + generator.random.Intn(2)
		states := make([]*spec, count)
		for index := range states {
			states[index] = generator.node()
		}
		return &spec{Kind: "conditionalGroup", Children: states, Flag: generator.random.Intn(6) == 0}
	case 10:
		count := 1 + generator.random.Intn(5)
		parts := make([]*spec, 0, 2*count)
		for index := 0; index < count; index++ {
			if index > 0 {
				parts = append(parts, &spec{Kind: []string{"line", "softline", "hardline"}[generator.random.Intn(3)]})
			}
			parts = append(parts, generator.node())
		}
		return &spec{Kind: "fill", Children: parts}
	case 11:
		ifBreak := &spec{Kind: "ifBreak", Children: []*spec{generator.node(), generator.node()}}
		if generator.random.Intn(2) == 0 {
			ifBreak.ID = generator.groupID()
		}
		return ifBreak
	case 12:
		return &spec{Kind: "indentIfBreak", ID: generator.groupID(), Flag: generator.random.Intn(2) == 0, Children: []*spec{generator.node()}}
	case 13:
		return &spec{Kind: "lineSuffix", Children: []*spec{generator.node()}}
	case 14:
		return &spec{Kind: []string{"lineSuffixBoundary", "breakParent", "trim"}[generator.random.Intn(3)]}
	case 15:
		return &spec{Kind: "label", Text: "member-chain", Children: []*spec{generator.node()}}
	default:
		return generator.text()
	}
}

// buildSpec constructs the Go doc for a spec.
func buildSpec(root *spec) Doc { return buildSpecAs(root, false) }

// buildSpecAs is buildSpec, building each concat as a *Sequence of its parts when asSequences is set.
func buildSpecAs(root *spec, asSequences bool) Doc {
	ids := map[string]*GroupID{}
	groups := map[int]*Group{}
	id := func(name string) *GroupID {
		if name == "" {
			return nil
		}
		if ids[name] == nil {
			ids[name] = NewGroupID(name)
		}
		return ids[name]
	}
	var build func(*spec) Doc
	children := func(node *spec) []Doc {
		built := make([]Doc, len(node.Children))
		for index, child := range node.Children {
			built[index] = build(child)
		}
		return built
	}
	build = func(node *spec) Doc {
		switch node.Kind {
		case "text":
			return Text(node.Text)
		case "line":
			return LineDoc
		case "softline":
			return Softline
		case "hardline":
			return Hardline
		case "literalline":
			return Literalline
		case "hardlineWithoutBreakParent":
			return HardlineWithoutBreakParent
		case "concat":
			if asSequences {
				return &Sequence{Parts: children(node)}
			}
			return Concat(children(node))
		case "indent":
			return NewIndent(build(node.Children[0]))
		case "alignWidth":
			return NewAlign(node.Number, build(node.Children[0]))
		case "alignString":
			return AlignWithString(node.Text, build(node.Children[0]))
		case "markAsRoot":
			return MarkAsRoot(build(node.Children[0]))
		case "dedentToRoot":
			return DedentToRoot(build(node.Children[0]))
		case "group":
			group := NewGroup(build(node.Children[0]), GroupOptions{ID: id(node.ID), ShouldBreak: node.Flag})
			if node.Ref != 0 {
				groups[node.Ref] = group
			}
			return group
		case "sharedGroup":
			return groups[node.Ref]
		case "conditionalGroup":
			return ConditionalGroup(children(node), GroupOptions{ShouldBreak: node.Flag})
		case "fill":
			return NewFill(children(node))
		case "ifBreak":
			return NewIfBreak(build(node.Children[0]), build(node.Children[1]), id(node.ID))
		case "indentIfBreak":
			return NewIndentIfBreak(build(node.Children[0]), id(node.ID), node.Flag)
		case "lineSuffix":
			return NewLineSuffix(build(node.Children[0]))
		case "lineSuffixBoundary":
			return LineSuffixBoundary
		case "breakParent":
			return BreakParent
		case "trim":
			return Trim
		case "label":
			return NewLabel(node.Text, build(node.Children[0]))
		}
		panic("unknown spec kind " + node.Kind)
	}
	return build(root)
}

// upstreamScript builds each spec with upstream's builders and prints it with upstream's printer.
const upstreamScript = `
import { builders as b, printer } from "./src/document/public.js";
process.stdin.setEncoding("utf8");
let input = ""; process.stdin.on("data", (chunk) => input += chunk);
process.stdin.on("end", () => {
  const results = JSON.parse(input).map(({ spec, options }) => {
    const ids = new Map(), groups = new Map();
    const id = (name) => { if (!name) return undefined; if (!ids.has(name)) ids.set(name, Symbol(name)); return ids.get(name); };
    const build = (node) => {
      const kids = () => (node.c || []).map(build);
      switch (node.k) {
        case "text": return node.v || "";
        case "line": return b.line;
        case "softline": return b.softline;
        case "hardline": return b.hardline;
        case "literalline": return b.literalline;
        case "hardlineWithoutBreakParent": return b.hardlineWithoutBreakParent;
        case "concat": return kids();
        case "indent": return b.indent(build(node.c[0]));
        case "alignWidth": return b.align(node.n || 0, build(node.c[0]));
        case "alignString": return b.align(node.v || "", build(node.c[0]));
        case "markAsRoot": return b.markAsRoot(build(node.c[0]));
        case "dedentToRoot": return b.dedentToRoot(build(node.c[0]));
        case "group": { const g = b.group(build(node.c[0]), { id: id(node.id), shouldBreak: !!node.f }); if (node.r) groups.set(node.r, g); return g; }
        case "sharedGroup": return groups.get(node.r);
        case "conditionalGroup": return b.conditionalGroup(kids(), { shouldBreak: !!node.f });
        case "fill": return b.fill(kids());
        case "ifBreak": return b.ifBreak(build(node.c[0]), build(node.c[1]), { groupId: id(node.id) });
        case "indentIfBreak": return b.indentIfBreak(build(node.c[0]), { groupId: id(node.id), negate: !!node.f });
        case "lineSuffix": return b.lineSuffix(build(node.c[0]));
        case "lineSuffixBoundary": return b.lineSuffixBoundary;
        case "breakParent": return b.breakParent;
        case "trim": return b.trim;
        case "label": return b.label(node.v, build(node.c[0]));
      }
      throw new Error("unknown spec kind " + node.k);
    };
    try {
      return { formatted: printer.printDocToString(build(spec), { ...options, endOfLine: "lf" }).formatted };
    } catch (error) {
      return { error: String(error) };
    }
  });
  process.stdout.write(JSON.stringify(results));
});
`

type printCase struct {
	Spec    *spec          `json:"spec"`
	Options map[string]any `json:"options"`
}

type printResultJSON struct {
	Formatted string `json:"formatted"`
	Error     string `json:"error"`
}

// boundaryCases are docs placed exactly at the print width, which random generation reaches too rarely.
//
// They exist because a mutation moving fits()'s pending space onto an empty string survived 5,000
// random docs: the difference only shows when a line, an empty string and a break sit at exactly the
// width. Each width is tried at the boundary and one either side.
func boundaryCases() ([]*spec, []Options) {
	var specs []*spec
	var options []Options
	text := func(value string) *spec { return &spec{Kind: "text", Text: value} }
	for width := 3; width <= 8; width++ {
		word := strings.Repeat("w", width)
		shapes := []*spec{
			{Kind: "concat", Children: []*spec{{Kind: "group", Children: []*spec{{Kind: "concat", Children: []*spec{text(word), {Kind: "line"}, text("")}}}}, {Kind: "hardline"}}},
			{Kind: "concat", Children: []*spec{{Kind: "group", Children: []*spec{{Kind: "concat", Children: []*spec{text(word), {Kind: "line"}, text(""), text("x")}}}}, {Kind: "hardline"}}},
			{Kind: "concat", Children: []*spec{{Kind: "group", Children: []*spec{{Kind: "concat", Children: []*spec{text(word[:width-1]), {Kind: "line"}, text("")}}}}, {Kind: "softline"}, text("y")}},
			{Kind: "fill", Children: []*spec{text(word), {Kind: "line"}, text(""), {Kind: "line"}, text("z")}},
		}
		for _, shape := range shapes {
			for _, printWidth := range []int{width - 1, width, width + 1} {
				specs = append(specs, shape)
				options = append(options, Options{PrintWidth: printWidth, TabWidth: 2})
			}
		}
	}
	return specs, options
}

// TestPrintAgreesWithUpstream is the differential. Off unless COHERE_PRETTIER_FORK names the fork.
func TestPrintAgreesWithUpstream(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)
	const count = 5000
	random := rand.New(rand.NewSource(20261001))
	cases := make([]printCase, count)
	options := make([]Options, count)
	for index := range cases {
		generator := &specGenerator{random: random}
		options[index] = Options{PrintWidth: 8 + random.Intn(40), TabWidth: []int{2, 4}[random.Intn(2)], UseTabs: random.Intn(4) == 0}
		cases[index] = printCase{
			Spec:    generator.node(),
			Options: map[string]any{"printWidth": options[index].PrintWidth, "tabWidth": options[index].TabWidth, "useTabs": options[index].UseTabs},
		}
	}

	boundarySpecs, boundaryOptions := boundaryCases()
	for index, boundary := range boundarySpecs {
		options = append(options, boundaryOptions[index])
		cases = append(cases, printCase{Spec: boundary, Options: map[string]any{"printWidth": boundaryOptions[index].PrintWidth, "tabWidth": 2, "useTabs": false}})
	}

	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--input-type=module", "-e", upstreamScript)
	command.Dir = root
	command.Stdin = bytes.NewReader(encoded)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	var upstream []printResultJSON
	if err := json.Unmarshal(stdout.Bytes(), &upstream); err != nil {
		t.Fatal(err)
	}

	mismatches, upstreamErrors, rewrites := 0, 0, 0
	for index, testCase := range cases {
		if upstream[index].Error != "" {
			upstreamErrors++
			if upstreamErrors <= 3 {
				t.Logf("upstream refused case %d: %s", index, upstream[index].Error)
			}
			continue
		}
		got := Print(buildSpec(testCase.Spec), options[index])
		if strings.Contains(got, "\n") {
			rewrites++
		}
		if got != upstream[index].Formatted {
			mismatches++
			if mismatches <= 5 {
				specJSON, _ := json.Marshal(testCase.Spec)
				t.Errorf("case %d (%+v)\nspec %s\nwant %q\ngot  %q", index, options[index], specJSON, upstream[index].Formatted, got)
			}
		}
	}
	t.Logf("%d docs printed (%d random, %d at the width boundary), %d upstream refused, %d with line breaks, %d mismatches",
		len(cases), count, len(boundarySpecs), upstreamErrors, rewrites, mismatches)
	if upstreamErrors > count/10 {
		t.Errorf("upstream refused %d of %d generated docs; the generator is producing invalid docs and the comparison is thinner than it looks", upstreamErrors, count)
	}
}
