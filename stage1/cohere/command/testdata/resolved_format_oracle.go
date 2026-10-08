package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
)

type row struct{ Directory, Name, Source, Config string }

func main() {
	data, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	var rows []row
	if e = json.Unmarshal(data, &rows); e != nil {
		panic(e)
	}
	for _, r := range rows {
		resolution, e := formatoptions.Resolve(r.Directory)
		kind, text := "Ok", ""
		if e != nil {
			kind, text = "Refused", e.Error()
		} else {
			formatter := native.Formatter{Options: resolution.Options}
			current := r.Source
			for pass := 1; pass <= 3; pass++ {
				formatted, e := formatter.Format(filepath.Join(r.Directory, r.Name), current)
				if e != nil {
					kind, text = "Error", e.Error()
					break
				}
				text = formatted
				if formatted == current {
					break
				}
				current = formatted
			}
		}
		encoded, e := json.Marshal(text)
		if e != nil {
			panic(e)
		}
		if _, e = os.Stdout.WriteString(kind + ":" + string(encoded) + "\n"); e != nil {
			panic(e)
		}
	}
}
