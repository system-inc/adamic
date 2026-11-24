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

	// Loom's jobserver bounds every clang this command's native build runs.
	_ "github.com/system-inc/adamic/internal/ccjobserver"
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
	written, err := writeReport(output, relativeSites(filepath.Dir(asked.root), lowered.Traps), lowered.TrapRewritesSkipped)
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

// relativeSites writes each trap's site relative to the program's common directory: the deepest directory holding
// the root and every site, so a row reads the same on every box. Rooted at src/tsc/tsc.ts, tsc's sites in
// src/compiler read compiler/checker.ts:..., where the root's own directory would leave them absolute (Correctness on
// #nm7z8fv). When the only directory they share is the filesystem's root, or a site has no path, sites stay as they
// came.
func relativeSites(rootDirectory string, traps []ir.Trap) []ir.Trap {
	base, err := filepath.Abs(rootDirectory)
	if err != nil {
		return traps
	}
	for _, trap := range traps {
		if path, _, found := strings.Cut(trap.Site, ":"); found && filepath.IsAbs(path) {
			base = commonDirectory(base, filepath.Dir(path))
		}
	}
	if filepath.Dir(base) == base {
		return traps
	}
	placed := make([]ir.Trap, len(traps))
	for index, trap := range traps {
		placed[index] = trap
		path, position, found := strings.Cut(trap.Site, ":")
		if !found || !filepath.IsAbs(path) {
			continue
		}
		if relative, err := filepath.Rel(base, path); err == nil {
			placed[index].Site = filepath.ToSlash(relative) + ":" + position
		}
	}
	return placed
}

// commonDirectory is the deepest directory holding both of two clean absolute directories.
func commonDirectory(first, second string) string {
	for {
		if relative, err := filepath.Rel(first, second); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return first
		}
		parent := filepath.Dir(first)
		if parent == first {
			return first
		}
		first = parent
	}
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

// findingsRule opens traps.json and traps.md, so nobody files a wall from a row raw.
const findingsRule = "leads only; reproduce without FirstLinkTraps before filing"

// report is traps.json: every trap in lower's order, and the counts row 1 of #bcbp175 reads.
type report struct {
	// Findings says what a row may be used for. A trap-mode run's output is a lead, never a miscompile finding,
	// until it's reproduced as an ordinary program with no trap mode (Compiler's condition 3 on #ht9e83d).
	Findings string `json:"findings"`

	Traps    []ir.Trap      `json:"traps"`
	Total    int            `json:"total"`
	ByPhase  map[string]int `json:"by_phase"`
	ByKind   map[string]int `json:"by_kind"`
	ByReason map[string]int `json:"by_reason"`

	// Unplaced counts the stops no unit holds, which trap nothing: a measurement gap.
	Unplaced int `json:"unplaced"`

	// Suspect counts the traps met after a recovered Go panic, or after a whole-program pass skipped a rewrite, in
	// the same lowering: each may be lowering's own echo rather than a wall (ir.Trap.Suspect).
	Suspect int `json:"suspect"`

	// Echo counts a whole-program pass's stops inside a unit trapped already, whose leftover state may be what it
	// stopped at, so they're no walls of their own (ir.Trap.Echo).
	Echo int `json:"echo"`

	// RewritesSkipped names each whole-program pass whose rewriting past a stop the lowering skipped, in order
	// (ir.Program.TrapRewritesSkipped). Every trap after the first is Suspect.
	RewritesSkipped []string `json:"rewrites_skipped"`
}

func summarize(traps []ir.Trap, skipped []string) report {
	counted := report{Findings: findingsRule, RewritesSkipped: skipped, Traps: traps, Total: len(traps), ByPhase: map[string]int{}, ByKind: map[string]int{}, ByReason: map[string]int{}}
	if counted.Traps == nil {
		counted.Traps = []ir.Trap{}
	}
	if counted.RewritesSkipped == nil {
		counted.RewritesSkipped = []string{}
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
		if trap.Echo {
			counted.Echo++
		}
	}
	return counted
}

// writeReport writes traps.json and traps.md under directory and says where.
func writeReport(directory string, traps []ir.Trap, skipped []string) ([]string, error) {
	counted := summarize(traps, skipped)
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
	fmt.Fprintf(&text, "Findings: %s. A trap-mode run's output is a lead, never a miscompile finding, until it's reproduced as an ordinary program.\n\n# First-link traps\n\n%d traps", counted.Findings, counted.Total)
	if counted.Unplaced > 0 {
		fmt.Fprintf(&text, ", %d of them unplaced: stops no unit holds, which trap nothing, a measurement gap", counted.Unplaced)
	}
	if counted.Suspect > 0 {
		fmt.Fprintf(&text, ", %d of them suspect: met after a recovered Go panic or a skipped rewrite, so each may be lowering's own echo", counted.Suspect)
	}
	if counted.Echo > 0 {
		fmt.Fprintf(&text, ", %d of them echoes: a pass's stop inside a unit trapped already, no wall of its own", counted.Echo)
	}
	text.WriteString(".\n")
	if len(counted.RewritesSkipped) > 0 {
		fmt.Fprintf(&text, "\nRewrites skipped past a stop, in order: %s. Every trap after the first is suspect.\n", strings.Join(counted.RewritesSkipped, ", "))
	}
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
