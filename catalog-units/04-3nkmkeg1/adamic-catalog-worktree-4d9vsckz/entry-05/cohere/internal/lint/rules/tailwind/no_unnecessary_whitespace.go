package tailwind

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func messageUnnecessaryWhitespace() rule.Message {
	return rule.Message{
		Id: "unnecessaryWhitespace",
		Description: "This class string carries whitespace that does nothing: a leading or trailing " +
			"run, or more than one space between two classes. It is invisible in the rendered output " +
			"and it makes the string a worse thing to diff and search, because two literals that mean " +
			"the same list stop being the same text.",
	}
}

// NoUnnecessaryWhitespaceOptions lets a project name the surfaces that carry class strings.
type NoUnnecessaryWhitespaceOptions struct {
	TailwindLocationOptions
	TailwindClassLiteralOptions
	// AllowMultiline is upstream's `allowMultiline`, on by default: a run holding a newline keeps it
	// and its indentation, losing only the spaces before the newline. Off, it is whitespace like any
	// other, shortened to one space between classes and removed at the edges.
	AllowMultiline *bool `json:"allowMultiline"`
}

// NoUnnecessaryWhitespace reports padding and doubled spaces inside a class string.
//
//	valid:   <div className="flex items-center" />
//	valid:   <div className={`flex ${extra} gap-2`} />
//	invalid: <div className="flex  items-center" />
//	invalid: <div className=" flex" />
//	invalid: <div className="flex " />
//
// The defect is cosmetic in the browser and not cosmetic in the repository. Class strings are the
// one part of the codebase with no compiler, so the only way anyone finds them is by reading and
// searching text. Two literals meaning the same list should be the same bytes, otherwise a search
// for one spelling misses the other, a diff shows a change where nothing changed, and the reviewer
// spends attention on whitespace instead of on the classes.
//
// Runs over every static run of class text, which for a template means each piece between the
// interpolations. That is what makes it safe next to a hole: whitespace adjacent to an interpolation
// is load-bearing, because it is the only thing keeping the substituted value from fusing with the
// class beside it. Measured against upstream: `flex  ${x}` becomes `flex ${x}`, keeping exactly one
// space, and never `flex${x}`.
//
// A class list written across lines keeps its lines. Upstream's `allowMultiline` is on by default,
// and with it a run holding a newline loses only the spaces before the newline. This rule used to
// shorten every run to its first character and drop the runs at the edges, which flattened a
// multi-line template literal to one line and left a doubled tab as a tab where upstream writes one
// space. Read from upstream 4.7.0's source when the option was ported (#gj5nm6e), and corrected to it.
//
// Fixable and safe to fix unattended, because the repair is defined entirely by what the string
// already says. Whitespace-only becomes empty, matching upstream, rather than being deleted: the
// attribute is the author's and this rule has no opinion about whether it should exist.
var NoUnnecessaryWhitespace = rule.Rule{
	Name: "better-tailwindcss/no-unnecessary-whitespace",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		surfaces := DefaultClassLiteralSurfaces()
		allowMultiline := true
		if configured, isConfigured := rule.OptionsAs[NoUnnecessaryWhitespaceOptions](options); isConfigured {
			if configured.AllowMultiline != nil {
				allowMultiline = *configured.AllowMultiline
			}
			surfaces = configured.ClassLiteralSurfaces()
		}

		reader := surfaces.ReaderFor(ctx.FileCache)

		report := func(node *ast.Node) {
			for _, segment := range reader.ClassSegmentsIn(node) {
				_, needsTidying := tidyWhitespace(segment, allowMultiline)
				if !needsTidying {
					continue
				}

				// A segment whose source is not its decoded value is reported without a fix. Writing the
				// decoded text back over the source would turn an escaped quote or backtick into a
				// bare one, the file would stop parsing, and the engine would refuse every fix in it.
				fixes, scoped := whitespaceFixes(ctx.SourceFile.Text(), segment, allowMultiline)
				if !scoped {
					fixes = nil
				}

				ctx.Report(rule.Diagnostic{
					Range:      segment.Range,
					Message:    messageUnnecessaryWhitespace(),
					SourceFile: ctx.SourceFile,
					Fixes:      fixes,
				})
			}
		}

		listeners := rule.Listeners{}
		for _, kind := range ListenerKinds() {
			listeners[kind] = report
		}
		return listeners
	},
}

// tidyWhitespace returns what a segment should say, and whether that differs from what it says.
//
// Each run of whitespace is judged on its own, by upstream's `lintLiterals` in its own order
// (separatorRepair). Comparing the rewritten text against the original, rather than pattern-matching
// for defects, is what keeps the finding and the fix from disagreeing: the rule reports exactly when
// it has something different to propose.
func tidyWhitespace(segment ClassSegment, allowMultiline bool) (string, bool) {
	runs := whitespaceRunsOf(segment.Text)
	classCount := 0
	for _, run := range runs {
		if !run.separator {
			classCount++
		}
	}

	var builder strings.Builder
	for index, run := range runs {
		if !run.separator {
			builder.WriteString(run.text)
			continue
		}
		replacement, _ := separatorRepair(run.text, index == 0, index == len(runs)-1, classCount, segment, allowMultiline)
		builder.WriteString(replacement)
	}

	tidied := builder.String()
	return tidied, tidied != segment.Text
}

// separatorRepair is what one run of whitespace should become, and whether that is a change. It is
// upstream's `lintLiterals`, decision for decision, with a hole beside a run standing in for its
// braces and its concatenation:
//
//   - A segment of nothing but whitespace, with no hole on either side, loses all of it.
//   - With allowMultiline, a run holding a newline loses only the spaces in front of the newline,
//     and keeps the newline and the indentation after it. This comes before the edge cases, so a
//     class list that opens and closes on its own lines keeps them.
//   - Between two classes, or at an edge with a hole beside it, a run longer than one character
//     becomes one space. Shortened, never synthesized: `px-${size}` has no whitespace at the hole and
//     gets none, since inventing one would turn `px-4` into two class names.
//   - At an outer edge with no hole beside it, the run goes.
//
// A hole beside a run is upstream's braces for a template's own text, and the edge rule
// `holeEdges` computes for a string inside another template's hole, which upstream reads as a plain
// string and trims, fusing `flex${open ? ' block' : ""}` into `flexblock`. Keeping one space there is
// deliberate; see holeEdges.
//
// Lengths count characters as upstream's JavaScript does, and every character isSpace accepts is one
// byte, so a byte length is that count.
func separatorRepair(run string, first bool, last bool, classCount int, segment ClassSegment, allowMultiline bool) (string, bool) {
	if classCount == 0 && !segment.LeadingHole && !segment.TrailingHole {
		return "", true
	}

	if allowMultiline && strings.Contains(run, "\n") {
		stripped := strings.TrimLeft(run, " ")
		return stripped, stripped != run
	}

	between := (!first && !last) ||
		(segment.LeadingHole && first && !last) ||
		(segment.TrailingHole && last && !first) ||
		(segment.LeadingHole && segment.TrailingHole)
	keep := (first && segment.LeadingHole) || (last && segment.TrailingHole)
	if between || keep {
		if len(run) <= 1 {
			return run, false
		}
		return " ", true
	}

	return "", true
}

// whitespaceRun is one run of a segment's text: a class, or the whitespace between classes.
type whitespaceRun struct {
	text      string
	separator bool
}

// whitespaceRunsOf splits a segment's decoded text into its classes and separators, in order.
func whitespaceRunsOf(text string) []whitespaceRun {
	runs := []whitespaceRun{}
	start := 0
	for start < len(text) {
		separator := isSpace(rune(text[start]))
		end := start + 1
		for end < len(text) && isSpace(rune(text[end])) == separator {
			end++
		}
		runs = append(runs, whitespaceRun{text: text[start:end], separator: separator})
		start = end
	}
	return runs
}

// whitespaceFixes is `tidyWhitespace`'s repair as one edit per run, touching only whitespace where it
// can.
//
// Rewriting the whole segment claimed every class in it too, so on a literal another rule was also
// fixing, one of the two was refused and waited a pass. So each repair is the smallest edit that
// reaches it: a run that goes is deleted, a run shortened to a space that already starts with one
// keeps that first byte and loses the rest, and a newline run loses only its leading spaces. Every
// class byte and those first separator bytes stay unclaimed, which is what `no-duplicate-classes`
// and `enforce-consistent-class-order` edit (class_tokens.go). Only a run longer than one character
// that starts with a tab or a newline, outside allowMultiline, is rewritten whole.
//
// Applied together the edits produce exactly `tidyWhitespace`'s text. False when the segment's source
// is not its decoded value, and the caller reports without a fix.
func whitespaceFixes(sourceText string, segment ClassSegment, allowMultiline bool) ([]rule.Fix, bool) {
	tokens, tokenized := classTokensIn(sourceText, segment.Range, segment.Text)
	if !tokenized {
		return nil, false
	}
	classCount := len(classesOf(tokens))

	fixes := []rule.Fix{}
	for index, token := range tokens {
		if !token.Separator {
			continue
		}
		replacement, changes := separatorRepair(token.Text, index == 0, index == len(tokens)-1, classCount, segment, allowMultiline)
		if !changes {
			continue
		}
		start, end := token.Range.Pos(), token.Range.End()
		switch {
		case replacement == "":
			fixes = append(fixes, rule.ReplaceRange(token.Range, ""))
		case replacement == " " && token.Text[0] == ' ':
			// The first space stays, so the byte the duplicate rule claims is never claimed here.
			fixes = append(fixes, rule.ReplaceRange(core.NewTextRange(start+1, end), ""))
		case strings.HasSuffix(token.Text, replacement):
			// The leading part goes and the rest, already what it should be, is left alone: a newline
			// run keeps the newline and its indentation.
			fixes = append(fixes, rule.ReplaceRange(core.NewTextRange(start, end-len(replacement)), ""))
		default:
			fixes = append(fixes, rule.ReplaceRange(token.Range, replacement))
		}
	}
	return fixes, true
}
