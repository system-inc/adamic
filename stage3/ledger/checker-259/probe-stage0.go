// Tooling for the checker ledger. This loads witnesses without invoking lowering.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/load"
)

func main() {
	names := os.Args[1:]
	whole := len(names) > 0 && names[0] == "--whole"
	if whole {
		names = names[1:]
	}
	for _, name := range names {
		paths := []string{name}
		if whole {
			paths = names
			name = "<whole>"
		}
		_, err := load.Load(paths)
		result := struct {
			File        string   `json:"file"`
			Diagnostics []string `json:"diagnostics"`
			Error       string   `json:"error,omitempty"`
		}{File: name, Diagnostics: []string{}}
		if err != nil {
			if check, ok := err.(*load.CheckError); ok {
				result.Diagnostics = check.Diagnostics
			} else {
				result.Error = err.Error()
			}
		}
		data, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		if whole {
			break
		}
	}
}
