//go:build ignore

// Build the existing shared parser under ASan/UBSan and emit its JavaScript.
package main

import (
	"context"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
)

func main() {
	program, err := load.Load([]string{"stage1/typescript/parser/main.ts"})
	if err != nil {
		panic(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		panic(err)
	}
	if err := native.Build(native.C(lowered), "/tmp/lint-wave1-09-parser", native.Options{Sanitize: true}); err != nil {
		panic(err)
	}
	if err := os.WriteFile("/tmp/lint-wave1-09-parser.mjs", []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		panic(err)
	}
}
