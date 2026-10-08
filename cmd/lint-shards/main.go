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
	count := flag.Int("n", runtime.NumCPU(), "number of shard processes")
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
	output, err := shards.Run(flag.Arg(0), flag.Arg(1), *count, *countOnly)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
