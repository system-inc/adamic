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
	dir, err := os.MkdirTemp("", "defend-iterator-collection-")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "main.a")
	source := "const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const kept:(number|undefined)[]=[...source];"
	if len(os.Args) > 1 {
		source = "const source={[Symbol.iterator](){return{next(){return{value:1,done:true};}}}};const kept=Array.from<number|undefined>(source);"
	}
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		panic(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		panic(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	fmt.Printf("program_nil=%t error_type=%T error=%v\n", program == nil, err, err)
}
