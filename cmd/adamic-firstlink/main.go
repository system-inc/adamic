// adamic-firstlink is the first link's measurement command (#bcbp175), and the only door to lower's first-link trap
// mode (lower.Options.FirstLinkTraps, #ht9e83d): every unit that stops lowers as a loud trap instead of stopping the
// program, so one run of a whole program reports every stop, and then emits, builds and links it anyway.
//
//	adamic-firstlink [--module-format esm|commonjs-bundle] <root.a|root.ts> <out-dir> [--link]
//
// It writes into out-dir, and nowhere else:
//
//	traps.json  every trap, with counts by phase, by kind and by reason
//	traps.md    the same as tables, phase by phase
//	c/main.c    the trap-mode program's C
//	js/main.js  its JavaScript twin
//	firstlink   with --link, the native binary
//
// and prints each path it wrote, one a line. A trap is data, never a failure, so it exits 0 when its report is
// written; 1 when the program doesn't check or a phase fails, naming the phase; 2 for a usage error.
//
// The C it writes is a trap-mode program: a stopped unit runs as a stub that prints `adamic: first-link trap:` and
// exits 75. So nothing it writes is ever a product a gate, the oracle or the agreement run reads, and it isn't part
// of adamic's usage, docs/cli.md or adamic's exit codes.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

const usage = "usage: adamic-firstlink [--module-format esm|commonjs-bundle] <root.a|root.ts> <out-dir> [--link]"

// moduleFormats are the words --module-format takes, as adamic's own command reads them.
var moduleFormats = map[string]lower.ModuleFormat{"esm": lower.ModuleFormatEcmaScript, "commonjs-bundle": lower.ModuleFormatCommonJSBundle}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// request is one run's words, read by parse.
type request struct {
	root, output string
	format       lower.ModuleFormat
	link         bool
}

// parse reads the arguments: an optional --module-format before the root, the root and the output directory, and an
// optional --link after them. ok is false for anything else.
func parse(arguments []string) (request, bool) {
	var asked request
	if len(arguments) >= 2 && arguments[0] == "--module-format" {
		format, known := moduleFormats[arguments[1]]
		if !known {
			return asked, false
		}
		asked.format, arguments = format, arguments[2:]
	}
	if len(arguments) == 3 && arguments[2] == "--link" {
		asked.link, arguments = true, arguments[:2]
	}
	if len(arguments) != 2 || arguments[0] == "" || arguments[1] == "" || strings.HasPrefix(arguments[0], "-") || strings.HasPrefix(arguments[1], "-") {
		return asked, false
	}
	asked.root, asked.output = arguments[0], arguments[1]
	return asked, true
}

func run(arguments []string, stdout, stderr io.Writer) int {
	asked, ok := parse(arguments)
	if !ok {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	output, err := filepath.Abs(asked.output)
	if err != nil {
		fmt.Fprintf(stderr, "adamic-firstlink: the output directory: %v\n", err)
		return 2
	}
	program, err := load.Load([]string{asked.root})
	if err != nil {
		var checkError *load.CheckError
		if errors.As(err, &checkError) {
			for _, diagnostic := range checkError.Diagnostics {
				fmt.Fprintln(stderr, diagnostic)
			}
		}
		fmt.Fprintf(stderr, "adamic-firstlink: check: %v\n", err)
		return 1
	}
	lowered, err := lower.LowerWithOptions(context.Background(), program, lower.Options{ModuleFormat: asked.format, FirstLinkTraps: true})
	if err != nil {
		// Trap mode lowers every stop as a trap, so an error here is one the mode couldn't place.
		fmt.Fprintf(stderr, "adamic-firstlink: lower: %v\n", err)
		return 1
	}
	written, err := writeReport(output, lowered.Traps)
	if err != nil {
		fmt.Fprintf(stderr, "adamic-firstlink: report: %v\n", err)
		return 1
	}
	source := native.C(lowered)
	for _, file := range []struct{ name, text string }{{"c/main.c", source}, {"js/main.js", javascript.JavaScript(lowered)}} {
		path, err := writeUnder(output, file.name, file.text)
		if err != nil {
			fmt.Fprintf(stderr, "adamic-firstlink: emit: %v\n", err)
			return 1
		}
		written = append(written, path)
	}
	if asked.link {
		binary := filepath.Join(output, "firstlink")
		if err := native.Build(source, binary, native_runtime.Options{}); err != nil {
			fmt.Fprintf(stderr, "adamic-firstlink: build and link: %v\n", err)
			for _, path := range written {
				fmt.Fprintln(stdout, path)
			}
			return 1
		}
		written = append(written, binary)
	}
	for _, path := range written {
		fmt.Fprintln(stdout, path)
	}
	return 0
}

// writeUnder writes text to name under directory, making its directories, and says where it wrote. A name that
// would leave directory is refused, so the command writes nowhere its caller didn't name.
func writeUnder(directory, name, text string) (string, error) {
	path := filepath.Join(directory, name)
	if relative, err := filepath.Rel(directory, path); err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s isn't under %s", name, directory)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(text), 0o644)
}

// report is traps.json: every trap in lower's order, and the counts row 1 of #bcbp175 reads.
type report struct {
	Traps    []ir.Trap `json:"traps"`
	Total    int                `json:"total"`
	ByPhase  map[string]int     `json:"by_phase"`
	ByKind   map[string]int     `json:"by_kind"`
	ByReason map[string]int     `json:"by_reason"`
	// Unplaced counts the traps lower couldn't place in a unit: a pass stopped at a site with no unit, and was
	// skipped from there on, so what it would have found after that site is a measurement gap.
	Unplaced int `json:"unplaced"`
	// Suspect counts the traps met after a recovered Go panic in the same lowering, which may be that panic's echo
	// rather than walls of their own (ir.Trap.Suspect).
	Suspect int `json:"suspect"`
}

func summarize(traps []ir.Trap) report {
	counted := report{Traps: traps, Total: len(traps), ByPhase: map[string]int{}, ByKind: map[string]int{}, ByReason: map[string]int{}}
	if counted.Traps == nil {
		counted.Traps = []ir.Trap{}
	}
	for _, trap := range traps {
		counted.ByPhase[trap.Phase]++
		counted.ByKind[trap.Kind]++
		counted.ByReason[trap.Kind+": "+trap.Reason]++
		if trap.Unit == "" {
			counted.Unplaced++
		}
		if trap.Suspect {
			counted.Suspect++
		}
	}
	return counted
}

// writeReport writes traps.json and traps.md under directory and says where.
func writeReport(directory string, traps []ir.Trap) ([]string, error) {
	counted := summarize(traps)
	data, err := json.MarshalIndent(counted, "", "  ")
	if err != nil {
		return nil, err
	}
	jsonPath, err := writeUnder(directory, "traps.json", string(data)+"\n")
	if err != nil {
		return nil, err
	}
	markdownPath, err := writeUnder(directory, "traps.md", markdown(counted))
	if err != nil {
		return nil, err
	}
	return []string{jsonPath, markdownPath}, nil
}

// markdown is traps.md: the totals, then one table a phase, each trap's reason counted with its first site.
func markdown(counted report) string {
	var text strings.Builder
	fmt.Fprintf(&text, "# First-link traps\n\n%d traps", counted.Total)
	if counted.Unplaced > 0 {
		fmt.Fprintf(&text, ", %d of them unplaced: a pass stopped where lower couldn't name a unit and was skipped from there on, so what it would have found after is a measurement gap", counted.Unplaced)
	}
	if counted.Suspect > 0 {
		fmt.Fprintf(&text, ", %d of them suspect: met after a recovered Go panic in the same lowering, so each may be that panic's echo", counted.Suspect)
	}
	text.WriteString(".\n")
	phases := make([]string, 0, len(counted.ByPhase))
	for phase := range counted.ByPhase {
		phases = append(phases, phase)
	}
	sort.Strings(phases)
	for _, phase := range phases {
		fmt.Fprintf(&text, "\n## %s, %d\n\n| count | kind | reason | first site |\n|---:|---|---|---|\n", phase, counted.ByPhase[phase])
		type row struct {
			kind, reason, site string
			count              int
		}
		rows := map[string]*row{}
		var order []string
		for _, trap := range counted.Traps {
			if trap.Phase != phase {
				continue
			}
			key := trap.Kind + "\x00" + trap.Reason
			if rows[key] == nil {
				rows[key] = &row{kind: trap.Kind, reason: trap.Reason, site: trap.Site}
				order = append(order, key)
			}
			rows[key].count++
		}
		sort.SliceStable(order, func(i, j int) bool { return rows[order[i]].count > rows[order[j]].count })
		for _, key := range order {
			fmt.Fprintf(&text, "| %d | %s | %s | %s |\n", rows[key].count, rows[key].kind, cell(rows[key].reason), cell(rows[key].site))
		}
	}
	return text.String()
}

// cell keeps a table cell on its row.
func cell(text string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ").Replace(text)
}
