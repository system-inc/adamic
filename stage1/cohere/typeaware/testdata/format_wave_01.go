// Runs the pinned native formatter over physical .a sources using virtual parser names.
package main

import (
	"fmt"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"os"
	"path/filepath"
)

func main() {
	for _, path := range os.Args[1:] {
		source, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		resolution, err := formatoptions.Resolve(filepath.Dir(path))
		if err != nil {
			panic(err)
		}
		formatter := native.Formatter{Options: resolution.Options}
		formatted, err := formatter.Format(path+".ts", string(source))
		if err != nil {
			panic(err)
		}
		if formatted != string(source) {
			if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
				panic(err)
			}
			fmt.Println(path)
		}
	}
}
