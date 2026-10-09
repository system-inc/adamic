package markdown

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/table.js.

type tableCell struct {
	text  string
	width int
}

func printTable(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	settings := settingsOf(options)
	layout := doc.Options{PrintWidth: settings.printWidth, TabWidth: settings.tabWidth, UseTabs: settings.useTabs}

	var columnMaxWidths []int
	contents := printing.Map(path, func(path *astPath, _ int, _ any) []tableCell {
		return printing.Map(path, func(path *astPath, columnIndex int, _ any) tableCell {
			text := doc.Print(print(nil, nil), layout)
			width := doc.StringWidth(text)
			for len(columnMaxWidths) <= columnIndex {
				columnMaxWidths = append(columnMaxWidths, 3) // minimum width = 3 (---, :--, :-:, --:)
			}
			columnMaxWidths[columnIndex] = max(columnMaxWidths[columnIndex], width)
			return tableCell{text: text, width: width}
		}, "children")
	}, "children")

	alignAt := func(index int) string {
		if index < len(node.Align) {
			return node.Align[index]
		}
		return ""
	}

	printAlign := func(isCompact bool) []string {
		var aligns []string
		for index, width := range columnMaxWidths {
			if index >= len(contents[0]) {
				// The header row must match the delimiter row in the number of cells. If not, a table
				// will not be recognized: https://github.github.com/gfm/#example-203
				continue
			}
			align := alignAt(index)
			first, last := "-", "-"
			if align == "center" || align == "left" {
				first = ":"
			}
			if align == "center" || align == "right" {
				last = ":"
			}
			middle := "-"
			if !isCompact {
				middle = strings.Repeat("-", width-2)
			}
			aligns = append(aligns, first+middle+last)
		}
		return aligns
	}

	printRow := func(columns []tableCell, isCompact bool) []string {
		row := make([]string, len(columns))
		for columnIndex, cell := range columns {
			if isCompact {
				row[columnIndex] = cell.text
				continue
			}
			spaces := columnMaxWidths[columnIndex] - cell.width
			align := alignAt(columnIndex)
			before := 0
			if align == "right" {
				before = spaces
			} else if align == "center" {
				before = spaces / 2
			}
			after := spaces - before
			row[columnIndex] = strings.Repeat(" ", before) + cell.text + strings.Repeat(" ", after)
		}
		return row
	}

	printTableContents := func(isCompact bool) doc.Doc {
		rows := [][]string{printRow(contents[0], isCompact), printAlign(isCompact)}
		for _, rowContents := range contents[1:] {
			rows = append(rows, printRow(rowContents, isCompact))
		}
		lines := make([]doc.Doc, len(rows))
		for index, columns := range rows {
			lines[index] = doc.Text("| " + strings.Join(columns, " | ") + " |")
		}
		return doc.Join(doc.HardlineWithoutBreakParent, lines)
	}

	alignedTable := printTableContents(false)
	if settings.proseWrap != "never" {
		return doc.Concat{doc.BreakParent, alignedTable}
	}

	// Only if the --prose-wrap never is set and it exceeds the print width.
	compactTable := printTableContents(true)
	return doc.Concat{doc.BreakParent, doc.NewGroup(doc.NewIfBreak(compactTable, alignedTable, nil), doc.GroupOptions{})}
}
