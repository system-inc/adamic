// Run from the repository root: go run ./cloud/reports/ownership-query-lint/testdata on after
// Profiles go to /tmp/ownership-query; loading and explicit GC are outside lowering.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func main() {
	if len(os.Args) != 3 || (os.Args[1] != "on" && os.Args[1] != "off") {
		panic("usage: profile on|off label")
	}
	if err := os.MkdirAll("/tmp/ownership-query", 0755); err != nil {
		panic(err)
	}
	enabled := os.Args[1] == "on"
	for i := 0; i < 3; i++ {
		checked, err := load.Load([]string{"stage1/cohere/lint/main.ts"})
		if err != nil {
			panic(err)
		}
		checked.EnableTSGo()
		runtime.GC()
		f, err := os.Create(fmt.Sprintf("/tmp/ownership-query/%s-%s-%d.cpu", os.Args[2], os.Args[1], i))
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		start := time.Now()
		program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: enabled})
		elapsed := time.Since(start)
		pprof.StopCPUProfile()
		if closeErr := f.Close(); closeErr != nil {
			panic(closeErr)
		}
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s %s %d %.3f ms functions=%d locals=%d\n", os.Args[2], os.Args[1], i, float64(elapsed.Microseconds())/1000, len(program.Functions), len(program.Locals))
	}
}
