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
 * The document utilities against upstream's, src/document/utilities.
 *
 * The same random specs the printer differential uses (printer_test.go) are built on both sides, then
 * each utility is applied and the results compared: the booleans directly, and the docs both
 * structurally (a canonical serialization, so a transform that prints the same but builds a different
 * tree is still caught) and by printing them.
 */

const utilitiesScript = `
import { builders as b, printer } from "./src/document/public.js";
import * as utils from "./src/document/utilities/index.js";
process.stdin.setEncoding("utf8");
let input = ""; process.stdin.on("data", (chunk) => input += chunk);
process.stdin.on("end", () => {
  const results = JSON.parse(input).map(({ spec }) => {
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
    const show = (doc) => {
      if (doc === undefined) return "undefined";
      if (typeof doc === "string") return JSON.stringify(doc);
      if (Array.isArray(doc)) return "[" + doc.map(show).join(",") + "]";
      switch (doc.type) {
        case "group": return "group(" + (doc.id ? doc.id.description : "") + "," + Boolean(doc.break) + "," + show(doc.contents) + (doc.expandedStates ? ",states[" + doc.expandedStates.map(show).join(",") + "]" : "") + ")";
        case "fill": return "fill[" + doc.parts.map(show).join(",") + "]";
        case "if-break": return "ifBreak(" + (doc.groupId ? doc.groupId.description : "") + "," + show(doc.breakContents) + "," + show(doc.flatContents) + ")";
        case "indent": return "indent(" + show(doc.contents) + ")";
        case "align": return "align(" + (typeof doc.n === "number" ? (doc.n === -Infinity ? "root" : doc.n) : typeof doc.n === "string" ? JSON.stringify(doc.n) : "markRoot") + "," + show(doc.contents) + ")";
        case "indent-if-break": return "indentIfBreak(" + doc.groupId.description + "," + Boolean(doc.negate) + "," + show(doc.contents) + ")";
        case "line-suffix": return "lineSuffix(" + show(doc.contents) + ")";
        case "line-suffix-boundary": return "lineSuffixBoundary";
        case "line": return "line(" + Boolean(doc.soft) + "," + Boolean(doc.hard) + "," + Boolean(doc.literal) + ")";
        case "break-parent": return "breakParent";
        case "trim": return "trim";
        case "label": return "label(" + doc.label + "," + show(doc.contents) + ")";
      }
      throw new Error("unknown doc " + JSON.stringify(doc));
    };
    const print = (doc) => { try { return printer.printDocToString(doc, { printWidth: 20, tabWidth: 2, endOfLine: "lf" }).formatted; } catch (error) { return "error: " + error; } };
    try {
      const outcome = {};
      outcome.willBreak = utils.willBreak(build(spec));
      outcome.canBreak = utils.canBreak(build(spec));
      outcome.isEmpty = utils.isEmptyDoc ? utils.isEmptyDoc(build(spec)) : null;
      const cleaned = utils.cleanDoc(build(spec));
      outcome.clean = show(cleaned);
      outcome.cleanPrinted = print(cleaned);
      const stripped = utils.stripTrailingHardline(build(spec));
      outcome.strip = show(stripped);
      outcome.stripPrinted = print(stripped);
      outcome.replaced = show(utils.replaceEndOfLine(build(spec)));
      outcome.removed = show(utils.removeLines ? utils.removeLines(build(spec)) : "");
      return outcome;
    } catch (error) {
      return { error: String(error) };
    }
  });
  process.stdout.write(JSON.stringify(results));
});
`

type utilitiesOutcome struct {
	WillBreak    bool   `json:"willBreak"`
	CanBreak     bool   `json:"canBreak"`
	IsEmpty      *bool  `json:"isEmpty"`
	Clean        string `json:"clean"`
	CleanPrinted string `json:"cleanPrinted"`
	Strip        string `json:"strip"`
	StripPrinted string `json:"stripPrinted"`
	Replaced     string `json:"replaced"`
	Removed      string `json:"removed"`
	Error        string `json:"error"`
}

// showDoc is the Go side of the script's show: a canonical serialization of a doc's structure.
func showDoc(document Doc) string {
	quote := func(text string) string {
		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		encoder.SetEscapeHTML(false)
		_ = encoder.Encode(text)
		return strings.TrimSuffix(buffer.String(), "\n")
	}
	list := func(parts []Doc) string {
		shown := make([]string, len(parts))
		for index, part := range parts {
			shown[index] = showDoc(part)
		}
		return strings.Join(shown, ",")
	}
	idName := func(id *GroupID) string {
		if id == nil {
			return ""
		}
		return id.name
	}
	switch typed := document.(type) {
	case nil:
		return "undefined"
	case Text:
		return quote(string(typed))
	case *Group:
		result := "group(" + idName(typed.ID) + "," + fmt.Sprint(typed.Break) + "," + showDoc(typed.Contents)
		if typed.ExpandedStates != nil {
			result += ",states[" + list(typed.ExpandedStates) + "]"
		}
		return result + ")"
	case *Fill:
		return "fill[" + list(typed.Parts) + "]"
	case *IfBreak:
		return "ifBreak(" + idName(typed.GroupID) + "," + showDoc(typed.BreakContents) + "," + showDoc(typed.FlatContents) + ")"
	case *Indent:
		return "indent(" + showDoc(typed.Contents) + ")"
	case *Align:
		var n string
		switch typed.Kind {
		case AlignWidth:
			n = fmt.Sprint(typed.Width)
		case AlignString:
			n = quote(typed.String)
		case AlignRoot:
			n = "markRoot"
		case AlignDedentToRoot:
			n = "root"
		}
		return "align(" + n + "," + showDoc(typed.Contents) + ")"
	case *IndentIfBreak:
		return "indentIfBreak(" + idName(typed.GroupID) + "," + fmt.Sprint(typed.Negate) + "," + showDoc(typed.Contents) + ")"
	case *LineSuffix:
		return "lineSuffix(" + showDoc(typed.Contents) + ")"
	case lineSuffixBoundaryDoc:
		return "lineSuffixBoundary"
	case *Line:
		return "line(" + fmt.Sprint(typed.Soft) + "," + fmt.Sprint(typed.Hard) + "," + fmt.Sprint(typed.Literal) + ")"
	case breakParentDoc:
		return "breakParent"
	case trimDoc:
		return "trim"
	case *Label:
		return "label(" + typed.Label + "," + showDoc(typed.Contents) + ")"
	}
	if parts, isArray := Parts(document); isArray {
		return "[" + list(parts) + "]"
	}
	panic(fmt.Sprintf("unknown doc %T", document))
}

// TestUtilitiesAgreeWithUpstream is the differential. Off unless COHERE_PRETTIER_FORK names the fork.
func TestUtilitiesAgreeWithUpstream(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)
	const count = 3000
	random := rand.New(rand.NewSource(20261002))
	cases := make([]printCase, count)
	for index := range cases {
		generator := &specGenerator{random: random}
		cases[index] = printCase{Spec: generator.node()}
	}
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--input-type=module", "-e", utilitiesScript)
	command.Dir = root
	command.Stdin = bytes.NewReader(encoded)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	var upstream []utilitiesOutcome
	if err := json.Unmarshal(stdout.Bytes(), &upstream); err != nil {
		t.Fatal(err)
	}

	print := func(document Doc) (result string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				result = fmt.Sprint("error: ", recovered)
			}
		}()
		return Print(document, Options{PrintWidth: 20, TabWidth: 2})
	}

	mismatches, upstreamErrors := 0, 0
	report := func(index int, what string, want any, got any) {
		mismatches++
		if mismatches <= 8 {
			specJSON, _ := json.Marshal(cases[index].Spec)
			t.Errorf("case %d %s\nspec %s\nwant %v\ngot  %v", index, what, specJSON, want, got)
		}
	}
	for index, testCase := range cases {
		want := upstream[index]
		if want.Error != "" {
			upstreamErrors++
			continue
		}
		if got := WillBreak(buildSpec(testCase.Spec)); got != want.WillBreak {
			report(index, "willBreak", want.WillBreak, got)
		}
		if got := CanBreak(buildSpec(testCase.Spec)); got != want.CanBreak {
			report(index, "canBreak", want.CanBreak, got)
		}
		if want.IsEmpty != nil {
			if got := IsEmptyDoc(buildSpec(testCase.Spec)); got != *want.IsEmpty {
				report(index, "isEmptyDoc", *want.IsEmpty, got)
			}
		}
		cleaned := CleanDoc(buildSpec(testCase.Spec))
		if got := showDoc(cleaned); got != want.Clean {
			report(index, "cleanDoc", want.Clean, got)
		} else if got := print(cleaned); got != want.CleanPrinted {
			report(index, "cleanDoc printed", want.CleanPrinted, got)
		}
		stripped := StripTrailingHardline(buildSpec(testCase.Spec))
		if got := showDoc(stripped); got != want.Strip {
			report(index, "stripTrailingHardline", want.Strip, got)
		} else if got := print(stripped); got != want.StripPrinted {
			report(index, "stripTrailingHardline printed", want.StripPrinted, got)
		}
		if got := showDoc(ReplaceEndOfLine(buildSpec(testCase.Spec), nil)); got != want.Replaced {
			report(index, "replaceEndOfLine", want.Replaced, got)
		}
		if got := showDoc(RemoveLines(buildSpec(testCase.Spec))); want.Removed != `""` && got != want.Removed {
			report(index, "removeLines", want.Removed, got)
		}
	}
	t.Logf("%d docs, %d upstream refused, %d mismatches", len(cases), upstreamErrors, mismatches)
	if upstreamErrors > count/10 {
		t.Errorf("upstream refused %d of %d docs; the comparison is thinner than it looks", upstreamErrors, count)
	}
}
