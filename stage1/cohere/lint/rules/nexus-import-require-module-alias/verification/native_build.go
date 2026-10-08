package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

func build() error {
	if len(os.Args) != 4 {
		return fmt.Errorf("usage: native_build <lint directory> <output> <checker archive>")
	}
	directory, output, archive := os.Args[1], os.Args[2], os.Args[3]
	if _, err := registry.Generate(directory); err != nil {
		return err
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		return err
	}
	program.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	source, err := native.TSGoC(lowered)
	if err != nil {
		return err
	}
	if err := os.WriteFile(output+".js", []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		return err
	}
	return native.BuildSplitTSGo(source, output, archive, native.Options{Sanitize: true, Jobs: 1})
}

func main() {
	if err := build(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
