package main

import (
	"context"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
)

func main() {
	program, err := load.Load([]string{os.Args[1]})
	if err != nil {
		panic(err)
	}
	program.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[2], []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		panic(err)
	}
}
