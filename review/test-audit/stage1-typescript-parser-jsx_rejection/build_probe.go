package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"time"
)

func main() {
	started := time.Now()
	program, err := load.Load([]string{"stage1/typescript/parser/main.ts"})
	if err != nil {
		panic(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		panic(err)
	}
	code := native.C(ir)
	lowered := time.Since(started).Seconds()
	target := os.Args[1]
	if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		panic(err)
	}
	buildStart := time.Now()
	err = native.Build(code, target, native.Options{Sanitize: true})
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(map[string]any{"lower_seconds": lowered, "native_seconds": time.Since(buildStart).Seconds(), "total_seconds": time.Since(started).Seconds(), "binary": target})
	fmt.Println(string(data))
}
