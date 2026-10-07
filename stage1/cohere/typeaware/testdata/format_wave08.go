// Built as an overlay inside cohere. The CLI on this pinned version does not
// discover .a files; invoke its TypeScript printer on their unchanged paths.
package main

import (
	"fmt"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/javascript"
	"os"
	"path/filepath"
)

func main() {
	for _, path := range os.Args[1:] {
		settings, err := formatoptions.Resolve(filepath.Dir(path))
		if err != nil {
			panic(err)
		}
		if filepath.Ext(path) != ".a" {
			panic("expected Adamic source")
		}
		source, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		formatted, err := javascript.Format(path, string(source), settings.Options, nil)
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
