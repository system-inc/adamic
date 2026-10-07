// adamic-reduce makes a failing program as small as it can while it still fails the same way, by
// fuzz's shrinker over TypeScript's own parse of it: whole declarations, statements, class members,
// arguments and elements go, compound statements give way to what they hold, and expressions give
// way to simpler ones of the same type, until nothing more goes. A candidate is kept only when it
// checks as the original did, comes to the same verdict, and fails with the same signature.
//
//	adamic-reduce program.a                           reduce the program's first failure
//	adamic-reduce -signature crash:'a store' -o small.a program.a
//	adamic-reduce -root ../other program.ts           against another checkout's compiler
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/fuzz"
)

func main() {
	os.Exit(run())
}

func run() int {
	root := flag.String("root", ".", "the checkout whose compiler the program fails: its adamic command, runtime and oracle/node.mjs")
	written := flag.String("signature", "", "what to keep, as <kind>:<text>, kind one of "+strings.Join(fuzz.SignatureKinds, ", ")+" (default: the program's first failure)")
	output := flag.String("o", "", "where the reduced program is written (default: stdout)")
	parallel := flag.Int("parallel", runtime.NumCPU(), "how many candidates at once")
	budget := flag.Int("budget", 5000, "the most candidates tried")
	work := flag.String("work", filepath.Join(os.TempDir(), "adamic-reduce"), "where candidates are built and run")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: adamic-reduce [-signature <kind>:<text>] [-o out.a] [-root <checkout>] <program.a|.ts>")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 || (!strings.HasSuffix(flag.Arg(0), ".a") && !strings.HasSuffix(flag.Arg(0), ".ts")) {
		flag.Usage()
		return 2
	}
	path := flag.Arg(0)
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	started := time.Now()
	checkout, err := fuzz.Prepare(*root, filepath.Join(*work, "checkout"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	name := "program" + filepath.Ext(path)

	original := checkout.Observe(string(source), name, filepath.Join(*work, "original"), "")
	var signature fuzz.Signature
	if *written == "" {
		signature, err = fuzz.Derive(original)
		if err != nil {
			fmt.Fprintf(os.Stderr, "adamic-reduce: %s: %s\n", path, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "adamic-reduce: signature %s\n", signature)
	} else {
		signature, err = fuzz.ParseSignature(*written)
		if err != nil {
			fmt.Fprintln(os.Stderr, "adamic-reduce:", err)
			return 2
		}
		if !original.Has(signature) {
			fmt.Fprintf(os.Stderr, "adamic-reduce: %s doesn't fail with %s: %s %v\n", path, signature, original.Verdict, original.Lines)
			return 1
		}
	}
	// A refusal or a crash needs only the compiler; anything else runs three ways.
	observe := func(candidate string, slot int) fuzz.Observation {
		return checkout.Observe(candidate, name, filepath.Join(*work, "candidates", fmt.Sprintf("slot%d", slot)), signature.Kind)
	}
	reduced := fuzz.Reduce(path, string(source), signature, original, observe, *parallel, *budget)

	if *output == "" {
		fmt.Print(reduced.Source)
	} else if err := os.WriteFile(*output, []byte(reduced.Source), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "adamic-reduce: %d lines, %d bytes to %d lines, %d bytes; %d candidates in %s\n",
		lines(string(source)), len(source), lines(reduced.Source), len(reduced.Source), reduced.Tried, time.Since(started).Round(100*time.Millisecond))
	return 0
}

func lines(text string) int {
	count := strings.Count(text, "\n")
	if text != "" && !strings.HasSuffix(text, "\n") {
		count++
	}
	return count
}
