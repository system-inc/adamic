// Records raw dependency responses, never lint findings, for the Node oracle.
package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/system-inc/adamic/bridge/tsgo/checker"
)

func main() {
	manifest, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	files := []string{}
	for _, file := range strings.Split(string(manifest), "\n") {
		if file != "" {
			files = append(files, file)
		}
	}
	program, err := checker.Open(os.Args[1], files)
	if err != nil {
		panic(err)
	}
	data := map[string]string{}
	for _, file := range files {
		source, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		for _, question := range []string{"preference-structure", "react-hir\nplain", "react-hir\nmemo"} {
			wire, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", question)
			if err != nil {
				panic(err)
			}
			data[file+"\x00"+question] = wire
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
		panic(err)
	}
}
