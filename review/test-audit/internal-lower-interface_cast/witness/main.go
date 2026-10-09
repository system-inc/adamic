package main

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.MkdirTemp("", "u033-witness-")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "receiver.a")
	source := "const own={value:1,read():number{return this.value;}};console.log(`${own.read()}`);"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		panic(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		panic(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		fmt.Printf("error=%v\n", err)
		return
	}
	for _, function := range program.Functions {
		if function.Name == "object_method" {
			fmt.Printf("name=%s receiver=%t parameters=%d\n", function.Name, function.Receiver, len(function.Parameters))
		}
	}
}
