// lint-shards runs the stage 1 lint driver over a manifest in N processes and prints what one process
// would (stage1/cohere/lint/shards). N defaults to the machine's cores.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/system-inc/adamic/stage1/cohere/lint/shards"
)

func main() {
	count := flag.Int("n", runtime.GOMAXPROCS(0), "number of shard processes")
	countOnly := flag.Bool("count", false, "print the total findings count only")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: lint-shards [-n N] [-count] <native lint binary> <manifest>")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	output, runError := shards.Run(flag.Arg(0), flag.Arg(1), *count, *countOnly)
	if runError != nil {
		fmt.Fprintln(os.Stderr, runError)
		os.Exit(1)
	}
	if _, writeError := os.Stdout.Write(output); writeError != nil {
		fmt.Fprintln(os.Stderr, writeError)
		os.Exit(1)
	}
}
