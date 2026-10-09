// adamic-fuzz generates random valid Adamic programs and holds each one to the oracle: its source on
// Node, native under ASan and UBSan, and the JavaScript backend on Node. Any disagreement or
// sanitizer finding is shrunk to a minimal program.
//
//	adamic-fuzz -seed 1 -count 1000             fuzz this checkout from seed 1
//	adamic-fuzz -root ../old -seed 1 -count 100 fuzz another checkout, an old commit say
//	adamic-fuzz -seed 42 -print                 print the program seed 42 makes
//
// Program n of a run is made from seed+n, so any finding reproduces with -seed alone and -count 1.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/system-inc/adamic/internal/fuzz"
)

func main() {
	os.Exit(run())
}

func run() int {
	root := flag.String("root", ".", "the checkout under test: its adamic command, runtime and oracle/node.mjs")
	seed := flag.Uint64("seed", 1, "the first program's seed")
	count := flag.Int("count", 100, "how many programs, from seed on")
	parallel := flag.Int("parallel", 4, "how many programs at once")
	work := flag.String("work", filepath.Join(os.TempDir(), "adamic-fuzz"), "where programs are built and run")
	shrink := flag.Bool("shrink", true, "shrink each finding to a minimal program")
	findings := flag.String("findings", "", "where shrunk findings are written (default: under -work)")
	print := flag.Bool("print", false, "print the program -seed makes, and stop")
	verbose := flag.Bool("v", false, "say what each program came to")
	without := flag.String("without", "", "features to leave out, by name, comma-separated (fuzz.Features), to stay inside what an older stage 0 lowered")
	with := flag.String("with", "", "opt-in features to put in, by name, comma-separated (fuzz.OptIn): shapes stage 0 is known to get wrong today")
	try := flag.String("try", "", "run one program file three ways, print what each did, and stop")
	flag.Parse()

	var leftOut []string
	if *without != "" {
		leftOut = strings.Split(*without, ",")
		for _, feature := range leftOut {
			if !slices.Contains(fuzz.Features, feature) {
				fmt.Fprintf(os.Stderr, "adamic-fuzz: no feature %q; the features are %s\n", feature, strings.Join(fuzz.Features, ", "))
				return 2
			}
		}
	}
	var putIn []string
	if *with != "" {
		putIn = strings.Split(*with, ",")
		for _, feature := range putIn {
			if !slices.Contains(fuzz.OptIn, feature) {
				fmt.Fprintf(os.Stderr, "adamic-fuzz: no opt-in feature %q; they are %s\n", feature, strings.Join(fuzz.OptIn, ", "))
				return 2
			}
		}
	}
	generate := func(seed uint64) *fuzz.Program {
		return fuzz.GenerateFeatures(seed, leftOut, putIn)
	}
	if *print {
		fmt.Print(generate(*seed).Source())
		return 0
	}
	if *findings == "" {
		*findings = filepath.Join(*work, "findings")
	}
	if err := os.RemoveAll(filepath.Join(*work, "programs")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	checkout, err := fuzz.Prepare(*root, filepath.Join(*work, "checkout"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *try != "" {
		source, err := os.ReadFile(*try)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		outcome := checkout.Try(string(source), filepath.Join(*work, "programs", "try"))
		fmt.Printf("%s %s\n", outcome.Verdict, outcome.Key)
		for _, way := range []struct {
			name string
			run  fuzz.Run
		}{{"node", outcome.Node}, {"native", outcome.Native}, {"backend", outcome.Backend}} {
			fmt.Printf("--- %s: exit %d\n%s", way.name, way.run.ExitCode, way.run.Stdout)
			if len(way.run.Stderr) > 0 {
				fmt.Printf("--- %s stderr:\n%s", way.name, way.run.Stderr)
			}
		}
		if outcome.Verdict == fuzz.Finding {
			return 1
		}
		return 0
	}

	var mutex sync.Mutex
	verdicts := map[fuzz.Verdict]int{}
	reasons := map[string]int{}
	var found []string
	var next atomic.Uint64
	started := time.Now()
	var group sync.WaitGroup
	for worker := range *parallel {
		group.Add(1)
		go func() {
			defer group.Done()
			for {
				index := next.Add(1) - 1
				if index >= uint64(*count) {
					return
				}
				programSeed := *seed + index
				directory := filepath.Join(*work, "programs", fmt.Sprintf("worker%d", worker))
				program := generate(programSeed)
				outcome := checkout.Try(program.Source(), directory)
				mutex.Lock()
				verdicts[outcome.Verdict]++
				if outcome.Verdict != fuzz.Agreed {
					reasons[string(outcome.Verdict)+": "+outcome.Key]++
				}
				if *verbose || outcome.Verdict == fuzz.Finding {
					fmt.Printf("seed %d: %s %s\n", programSeed, outcome.Verdict, outcome.Key)
				}
				mutex.Unlock()
				if outcome.Verdict != fuzz.Finding {
					continue
				}
				fmt.Printf("seed %d:\n%s\n", programSeed, indent(outcome.Detail))
				path := filepath.Join(*findings, fmt.Sprintf("seed%d.a", programSeed))
				if *shrink {
					shrunk := fuzz.Shrink(program, outcome.Key, func(candidate *fuzz.Program) fuzz.Outcome {
						return checkout.Try(candidate.Source(), filepath.Join(*work, "programs", fmt.Sprintf("worker%d-shrink", worker)))
					})
					program = shrunk
				}
				if err := fuzz.WriteFinding(path, program, programSeed, leftOut, putIn, outcome.Key); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				mutex.Lock()
				found = append(found, fmt.Sprintf("seed %d: %s -> %s", programSeed, outcome.Key, path))
				mutex.Unlock()
			}
		}()
	}
	group.Wait()

	fmt.Printf("\n%d programs from seed %d in %s on %s\n", *count, *seed, time.Since(started).Round(time.Second), checkout.Root)
	for _, verdict := range []fuzz.Verdict{fuzz.Agreed, fuzz.Finding, fuzz.Checked, fuzz.NotYet, fuzz.Invalid, fuzz.Unfit} {
		fmt.Printf("  %-8s %d\n", verdict, verdicts[verdict])
	}
	var keys []string
	for key := range reasons {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left, right int) bool { return reasons[keys[left]] > reasons[keys[right]] })
	for _, key := range keys {
		fmt.Printf("  %5d  %s\n", reasons[key], key)
	}
	for _, line := range found {
		fmt.Println(line)
	}
	if len(found) > 0 {
		return 1
	}
	return 0
}

func indent(text string) string {
	return "  " + strings.ReplaceAll(text, "\n", "\n  ")
}
