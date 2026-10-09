package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"time"
)

func main() {
	start := time.Now()
	p, e := load.Load([]string{os.Args[1]})
	if e != nil {
		panic(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		panic(e)
	}
	lowerSeconds := time.Since(start).Seconds()
	start = time.Now()
	e = native.Build(native.C(ir), os.Args[2], native.Options{Sanitize: true})
	if e != nil {
		panic(e)
	}
	b, _ := json.Marshal(map[string]any{"entry": os.Args[1], "load_lower_seconds": lowerSeconds, "native_build_seconds": time.Since(start).Seconds()})
	fmt.Println(string(b))
}
