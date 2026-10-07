// runtimelibrary prints the path of the cached runtime library a coverage measurement's builds link,
// building it first if it isn't cached: the sanitized runtime, as the fuzzer and the oracle compile
// it, with clang's source-based coverage on (native.Options.Coverage). llvm-cov reads the runtime's
// coverage mapping from that archive's objects. verify/coverage/measure.sh runs it.
//
//	go run ./verify/coverage/runtimelibrary             the sanitized runtime
//	go run ./verify/coverage/runtimelibrary -release    the -O2 one the oracle's release builds link
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/native"
)

func main() {
	release := flag.Bool("release", false, "the release build's runtime (-O2, no sanitizers) instead of the sanitized one")
	flag.Parse()
	library, err := native.RuntimeLibrary("", native.Options{Sanitize: !*release, Coverage: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(library)
}
