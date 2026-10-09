package unused_exports

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module_roots"
	"github.com/system-inc/cohere/internal/types/program"
)

// Run produces the whole unused report: what nothing references, and what nothing can reach.
//
// `deep` asks for the transitive closure on top of the flat answers. It is a parameter rather than
// always-on because the closure makes a strictly stronger claim that fails differently: see the
// note on `computeIslands`.
//
// The two halves are computed together and printed together because they read as one question to a
// person — "what did I write that is not doing anything" — while being two analyses with different
// evidence behind them. The report keeps them in separate sections precisely so the reader can trust
// them differently, which is the honest presentation of two claims of unequal strength.
func Run(ctx context.Context, graph *program.Graph, files []*ast.SourceFile, deep bool) (*Report, error) {
	roots := module_roots.Build(files)

	report, err := FindUnreferenced(ctx, graph, files, roots, deep)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		report.Unreachable = append(report.Unreachable, FindUnreachable(file)...)
	}

	return report, nil
}

// Write prints the report.
//
// The summary line comes first and states the population as well as the findings, because a count
// with no denominator cannot be told apart from a broken probe. A report saying "0 unused" over a
// tree of 3,400 files is far more likely to be a broken analysis than a clean codebase, and the
// only thing that separates those two readings is the denominator printed beside the zero.
func Write(out io.Writer, report *Report, showIntentional bool) {
	fmt.Fprintf(out, "\nunused: a report, not a gate — nothing here fails a build\n")

	intentionalFiles := 0
	for _, file := range report.Files {
		if file.Intentional.Present {
			intentionalFiles++
		}
	}
	intentionalExports := 0
	for _, finding := range report.Unreferenced {
		if finding.Intentional.Present {
			intentionalExports++
		}
	}

	writeIslandSection(out, report, showIntentional)
	writeFileSection(out, report, showIntentional, intentionalFiles)
	writeExportSection(out, report, showIntentional, intentionalExports)
	writeUnreachableSection(out, report)

	// The coverage line, in the same spirit as the lint phase's. It states what was looked at, not
	// only what was found, so a run that analyzed nothing cannot read like a run that found nothing.
	fmt.Fprintf(out,
		"  looked at %d files (%d spared by a framework or dynamic-load convention) and %d exported declarations\n",
		report.FilesAnalyzed, report.RootedFiles, report.ExportsAnalyzed,
	)
	if report.DeclarationsAnalyzed > 0 {
		// The closure's own denominator, plus the depth it reached. A closure that stopped at depth 1
		// on a real tree has failed to propagate rather than found a flat codebase, and printing the
		// number is what makes that visible instead of quietly doubling the report.
		//
		// The depth is "at least", not an invariant. It is the longest chain THIS traversal happened
		// to follow, and a worklist reaches a symbol by whichever path it pops first, so the number
		// moves run to run — measured on the ahra tree: 8, 9 and 10 across three runs whose findings
		// were byte-identical at 1042 clusters and 2051 declarations every time. The set is the
		// result; the depth is a liveness signal about the traversal, and saying so stops a reader
		// from reading a moving number as an unstable analysis.
		fmt.Fprintf(out,
			"  the closure walked %d declarations and reached a depth of at least %d\n",
			report.DeclarationsAnalyzed, report.ClosurePasses,
		)
	}
}

// writeIslandSection prints the clusters, largest first.
//
// The cluster is the point. A flat list prints an abandoned export, the helper only it called, and
// the table only that helper read as three unrelated lines in three places, and the reader has to
// reassemble them. Printing them as one block with a size is what turns the report from a list into
// the "oh hey, this was written and never got used" it exists to produce.
func writeIslandSection(out io.Writer, report *Report, showIntentional bool) {
	if len(report.Islands) == 0 {
		return
	}

	clusters := 0
	declarations := 0
	intentional := 0
	for _, island := range report.Islands {
		marked := false
		for _, member := range island.Members {
			if member.Intentional.Present {
				marked = true
			}
		}
		if marked {
			intentional++
			if !showIntentional {
				continue
			}
		}
		clusters++
		declarations += len(island.Members)
	}

	fmt.Fprintf(out,
		"\n  islands — %d clusters, %d declarations, dead together (transitive; --unused-deep)\n",
		clusters, declarations,
	)
	if intentional > 0 {
		fmt.Fprintf(out, "    %d clusters hold a %s and are not listed\n", intentional, intentionalMarker)
	}

	for _, island := range report.Islands {
		marked := false
		for _, member := range island.Members {
			if member.Intentional.Present {
				marked = true
			}
		}
		if marked && !showIntentional {
			continue
		}

		// A one-member island is not a cluster and saying "1 declaration" about it adds nothing. The
		// header earns its line only when there is a group to name.
		if len(island.Members) == 1 {
			member := island.Members[0]
			fmt.Fprintf(out, "    %s — %s %s%s\n",
				member.FileName, member.Kind, member.Name, intentionalSuffix(member.Intentional))
			continue
		}

		head := island.Members[0]
		fmt.Fprintf(out, "    %d declarations, entered at %s %s (%s)\n",
			len(island.Members), head.Kind, head.Name, head.FileName)
		for _, member := range island.Members {
			visibility := "private"
			if member.Exported {
				visibility = "exported"
			}
			fmt.Fprintf(out, "      %s %s %s — %s%s\n",
				visibility, member.Kind, member.Name, member.FileName, intentionalSuffix(member.Intentional))
		}
	}
}

func writeFileSection(out io.Writer, report *Report, showIntentional bool, intentional int) {
	if len(report.Files) == 0 {
		return
	}
	sort.Slice(report.Files, func(left, right int) bool {
		return report.Files[left].FileName < report.Files[right].FileName
	})

	fmt.Fprintf(out, "\n  files whose every export is unreferenced — %s\n", countPhrase(len(report.Files), intentional))
	for _, file := range report.Files {
		if file.Intentional.Present && !showIntentional {
			continue
		}
		marker := ""
		if file.Intentional.Present {
			marker = intentionalSuffix(file.Intentional)
		}
		fmt.Fprintf(out, "    %s — %d exports, none used%s\n", file.FileName, file.Exports, marker)
	}
}

func writeExportSection(out io.Writer, report *Report, showIntentional bool, intentional int) {
	if len(report.Unreferenced) == 0 {
		return
	}
	sort.Slice(report.Unreferenced, func(left, right int) bool {
		if report.Unreferenced[left].FileName != report.Unreferenced[right].FileName {
			return report.Unreferenced[left].FileName < report.Unreferenced[right].FileName
		}
		return report.Unreferenced[left].Name < report.Unreferenced[right].Name
	})

	fmt.Fprintf(out, "\n  exports nothing references — %s\n", countPhrase(len(report.Unreferenced), intentional))
	for _, finding := range report.Unreferenced {
		if finding.Intentional.Present && !showIntentional {
			continue
		}
		fmt.Fprintf(out, "    %s — %s %s%s\n",
			finding.FileName, finding.Kind, finding.Name, intentionalSuffix(finding.Intentional))
	}
}

func writeUnreachableSection(out io.Writer, report *Report) {
	if len(report.Unreachable) == 0 {
		return
	}
	sort.Slice(report.Unreachable, func(left, right int) bool {
		if report.Unreachable[left].FileName != report.Unreachable[right].FileName {
			return report.Unreachable[left].FileName < report.Unreachable[right].FileName
		}
		return report.Unreachable[left].Range.Position < report.Unreachable[right].Range.Position
	})

	fmt.Fprintf(out, "\n  statements nothing can reach — %d\n", len(report.Unreachable))
	for _, finding := range report.Unreachable {
		fmt.Fprintf(out, "    %s:%d — %s\n", finding.FileName, finding.Range.Position, finding.Cause)
	}
}

// countPhrase reads `18 unused, 4 declared intentional` rather than 22 undifferentiated lines, so
// the reader can see at a glance how much of the report is already judged.
func countPhrase(total int, intentional int) string {
	if intentional == 0 {
		return fmt.Sprintf("%d", total)
	}
	return fmt.Sprintf("%d, %d declared intentional with %s", total-intentional, intentional, intentionalMarker)
}

func intentionalSuffix(reason intentionalReason) string {
	if !reason.Present {
		return ""
	}
	if reason.Text == "" {
		return fmt.Sprintf("  [%s]", intentionalMarker)
	}
	return fmt.Sprintf("  [%s: %s]", intentionalMarker, reason.Text)
}
