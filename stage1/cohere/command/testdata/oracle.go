package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"os"
)

func main() {
	var rows []struct {
		Name   string
		Source string
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	formatter := native.Formatter{Options: formatoptions.PrettierDefaults()}
	out := make([]struct {
		Text  string
		Error string
	}, len(rows))
	for i, row := range rows {
		current := row.Source
		for pass := 1; pass <= 3; pass++ {
			text, e := formatter.Format(row.Name, current)
			if e != nil {
				out[i].Error = e.Error()
				break
			}
			out[i].Text = text
			if text == current {
				break
			}
			current = text
		}
	}
	if err = json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}
