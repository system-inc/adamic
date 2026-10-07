// Command adamic-test262 runs a slice of test262 through stage 0 and through Node, and counts how
// they compare.
//
//	adamic-test262 -test262 /path/to/test262 built-ins/String/prototype/padStart
//	adamic-test262 -adapt -json -test262 /path/to/test262 built-ins/String built-ins/Math
//
// The test262 checkout is not part of this repo: clone it at a pinned commit into a scratch
// directory and pass it. A test is skipped, with a reason, when it needs a feature Adamic refuses
// or has not built, when it is a negative parse or early error, or when it loads a harness file
// the prelude cannot stand in for. Anything else is attempted: one program, the adapted harness
// plus the test, lowered by an isolated persistent compiler worker, linked with clang and run
// natively and with Node. An explicit -root or -compiler-subprocess uses `adamic c`. pass
// means both succeeded and their output matches. fail means they disagree, or native failed where
// Node passed. refused means stage 0 or the checker said no, the reason normalized the way a meter
// groups them (cmd/adamic-meter is not on main; see reason.go). crashed means a signal, a
// sanitizer, a timeout, or the compiler itself failing. The crash record names the test file.
//
// --adapt rewrites test262's spelling in memory only, and counts each rewrite: var to let where
// the meaning does not change, a callback parameter typed the way the call would type it, == and
// != where both sides already have the same type, and throw new Test262Error to throw new Error
// when nothing observes the constructor. The checkout is not modified.
//
// -jobs defaults to GOMAXPROCS. Reports and progress lines retain serial order. The runtime is
// cached per source/toolchain/flag set, with the address and undefined-behavior sanitizers.
// Successful C generation, Node and native observations are cached under the user cache directory,
// with exact program,
// toolchain, adaptation and command identities. ADAMIC_GATE_UNCACHED=1 bypasses all result caches.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/system-inc/adamic/internal/nodepin"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(arguments []string) (exit int) {
	if _, err := nodepin.Check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	flags := flag.NewFlagSet("adamic-test262", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	test262 := flags.String("test262", "", "test262 checkout (a clone at a pinned commit, not part of this repo)")
	root := flags.String("root", ".", "Adamic checkout whose cmd/adamic and runtime are under test")
	work := flags.String("work", "", "scratch directory (default: a directory under the system temp)")
	profilePath := flags.String("profile", "", "write phase durations, cache hits and worker timeline to this JSON file")
	asJSON := flags.Bool("json", false, "write the report as JSON on stdout; the table goes to stderr")
	classifyOnly := flags.Bool("classify-only", false, "classify every test and do not compile or run")
	adapt := flags.Bool("adapt", false, "rewrite test262 style in memory (var to let, callback params, strict equality, Test262Error) and count each rewrite")
	subprocess := flags.Bool("compiler-subprocess", false, "start adamic c for each test (reference path; also used with an explicit -root)")
	jobs := flags.Int("jobs", runtime.GOMAXPROCS(0), "number of concurrent tests")
	limit := flags.Int("limit", 0, "run at most this many attempted tests per filter (0 is all)")
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: adamic-test262 [flags] <filter>...\n\n")
		fmt.Fprintf(os.Stderr, "  filter is a directory under test262's test/, like built-ins/String/prototype/padStart\n\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *jobs < 1 {
		fmt.Fprintln(os.Stderr, "jobs must be positive")
		return 2
	}
	if *test262 == "" || flags.NArg() == 0 {
		flags.Usage()
		return 2
	}
	workDirectory := *work
	if workDirectory == "" {
		var err error
		workDirectory, err = os.MkdirTemp("", "adamic-test262-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	var profile *runProfile
	if *profilePath != "" {
		profile = newRunProfile()
		defer func() {
			if err := profile.write(*profilePath); err != nil {
				fmt.Fprintln(os.Stderr, err)
				exit = 1
			}
		}()
	}
	inProcess := !*subprocess
	flags.Visit(func(flag *flag.Flag) {
		if flag.Name == "root" {
			inProcess = false
		}
	})
	var prepared *engine
	var err error
	if !*classifyOnly {
		prepared, err = prepareMode(*root, *test262, workDirectory, profile, inProcess)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else {
		prepared = &engine{test262: *test262, log: os.Stderr, adapt: *adapt}
	}
	prepared.adapt = *adapt
	prepared.jobs = *jobs
	prepared.inProcess = inProcess
	document := reportDocument{Test262: *test262, Commit: test262Commit(*test262), Adapt: *adapt}
	for _, filter := range flags.Args() {
		report, err := prepared.runFilter(filter, *limit, *classifyOnly)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		document.Filters = append(document.Filters, report)
	}
	if *asJSON {
		printTables(os.Stderr, document)
		if err := writeJSON(os.Stdout, document); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	printTables(os.Stdout, document)
	return 0
}

func printTables(writer io.Writer, document reportDocument) {
	if document.Commit != "" {
		fmt.Fprintf(writer, "test262 %s\n", document.Commit)
	}
	if document.Adapt {
		fmt.Fprintf(writer, "adapt on\n")
	} else {
		fmt.Fprintf(writer, "adapt off\n")
	}
	fmt.Fprintln(writer)
	for _, filter := range document.Filters {
		fmt.Fprintf(writer, "%s\n", filter.Path)
		if filter.Unrun > 0 {
			fmt.Fprintf(writer, "pass %d   fail %d   refused %d   crashed %d   skipped %d   not run %d   total %d\n\n",
				filter.Pass, filter.Fail, filter.Refused, filter.Crashed, filter.Skipped, filter.Unrun, filter.Total)
		} else {
			fmt.Fprintf(writer, "pass %d   fail %d   refused %d   crashed %d   skipped %d   total %d\n\n",
				filter.Pass, filter.Fail, filter.Refused, filter.Crashed, filter.Skipped, filter.Total)
		}
		width := len("directory")
		for _, directory := range filter.Directories {
			if len(directory.Path) > width {
				width = len(directory.Path)
			}
		}
		fmt.Fprintf(writer, "%-*s  %6s  %6s  %8s  %8s  %8s\n", width, "directory", "pass", "fail", "refused", "crashed", "skipped")
		for _, directory := range filter.Directories {
			fmt.Fprintf(writer, "%-*s  %6d  %6d  %8d  %8d  %8d\n", width, directory.Path, directory.Pass, directory.Fail, directory.Refused, directory.Crashed, directory.Skipped)
		}
		fmt.Fprintln(writer)
		printReasons(writer, "refusal reasons", filter.RefusalReasons)
		printReasons(writer, "skip reasons", filter.SkipReasons)
		printReasons(writer, "fail reasons", filter.FailReasons)
		printReasons(writer, "crash reasons", filter.CrashReasons)
		printReasons(writer, "adaptations", filter.Adaptations)
		if len(filter.Passes) > 0 {
			fmt.Fprintf(writer, "passes (%d)\n", len(filter.Passes))
			shown := filter.Passes
			if len(shown) > 40 {
				shown = shown[:40]
			}
			for _, path := range shown {
				fmt.Fprintf(writer, "  %s\n", path)
			}
			if len(filter.Passes) > len(shown) {
				fmt.Fprintf(writer, "  ... and %d more\n", len(filter.Passes)-len(shown))
			}
			fmt.Fprintln(writer)
		}
	}
}

func printReasons(writer io.Writer, title string, reasons []reasonCount) {
	if len(reasons) == 0 {
		return
	}
	fmt.Fprintf(writer, "%s\n", title)
	shown := reasons
	if len(shown) > 15 {
		shown = shown[:15]
	}
	width := 0
	for _, reason := range shown {
		if reason.Count > width {
			width = reason.Count
		}
	}
	digits := len(fmt.Sprintf("%d", width))
	for _, reason := range shown {
		fmt.Fprintf(writer, "  %*d  %s\n", digits, reason.Count, oneLine(reason.Reason))
	}
	if len(reasons) > len(shown) {
		fmt.Fprintf(writer, "  ... and %d more\n", len(reasons)-len(shown))
	}
	fmt.Fprintln(writer)
}

func oneLine(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) > 160 {
		return text[:160] + "..."
	}
	return text
}
