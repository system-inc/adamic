package main

import (
	"context"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func main() {
	program, err := load.Load([]string{os.Args[1]})
	if err != nil {
		panic(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[3], []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		panic(err)
	}
	if err = native.Build(native.C(lowered), os.Args[2], native.Options{Sanitize: true}); err != nil {
		panic(err)
	}
	fmt.Println("sanitized native and emitted JavaScript built")
}
