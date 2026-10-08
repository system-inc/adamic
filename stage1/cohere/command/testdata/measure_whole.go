//go:build ignore

// Build the scratch whole program with the existing renderer and checker ABI.
package main

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
)

func main() {
	archive, output := os.Args[2], os.Args[3]
	var c string
	if os.Args[1] == "--c" {
		data, e := os.ReadFile(os.Args[2])
		if e != nil {
			panic(e)
		}
		c = string(data)
		archive, output = os.Args[3], os.Args[4]
	} else {
		p, e := load.Load([]string{os.Args[1]})
		if e != nil {
			panic(e)
		}
		p.EnableTSGo()
		ir, e := lower.Lower(context.Background(), p)
		if e != nil {
			panic(e)
		}
		c, e = native.TSGoC(ir)
		if e != nil {
			panic(e)
		}
	}
	if e := os.WriteFile(output+".c", []byte(c), 0644); e != nil {
		panic(e)
	}
	if e := native.BuildTSGo(c, output, archive, native.Options{}); e != nil {
		panic(e)
	}
	info, e := os.Stat(output)
	if e != nil {
		panic(e)
	}
	fmt.Printf("whole C=%d binary=%d\n", len(c), info.Size())
}
