package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

type querySet struct {
	Queries []string `json:"queries"`
}
type options struct {
	TabWidth        int    `json:"tabWidth"`
	UseTabs         bool   `json:"useTabs"`
	Semi            bool   `json:"semi"`
	SingleQuote     bool   `json:"singleQuote"`
	PrintWidth      int    `json:"printWidth"`
	TrailingComma   string `json:"trailingComma"`
	BracketSpacing  bool   `json:"bracketSpacing"`
	BracketSameLine bool   `json:"bracketSameLine"`
	ArrowParens     string `json:"arrowParens"`
	EndOfLine       string `json:"endOfLine"`
}
type success struct {
	Kind                string   `json:"kind"`
	Options             options  `json:"options"`
	Source              string   `json:"source"`
	HouseIgnore         []string `json:"houseIgnore"`
	HouseIgnoreDeclared bool     `json:"houseIgnoreDeclared"`
	IgnorePatterns      []string `json:"ignorePatterns"`
}
type failure struct {
	Kind     string `json:"kind"`
	Message  string `json:"message"`
	Status   int    `json:"status"`
	Leftover bool   `json:"leftover"`
	Stderr   string `json:"stderr"`
	Stdout   string `json:"stdout"`
}

func main() {
	data, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	var sets []querySet
	if e = json.Unmarshal(data, &sets); e != nil {
		panic(e)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	for _, set := range sets {
		resolver := formatoptions.NewResolver()
		for _, query := range set.Queries {
			resolution, e := resolver.Resolve(filepath.Dir(query))
			if e != nil {
				message := e.Error()
				if e = encoder.Encode(failure{"Refused", message, 1, errors.Is(e, formatoptions.ErrPrettierConfigRemains), message + "\n", ""}); e != nil {
					panic(e)
				}
			} else {
				v := resolution.Options
				hi, ip := resolution.HouseIgnore, resolution.IgnorePatterns
				if hi == nil {
					hi = []string{}
				}
				if ip == nil {
					ip = []string{}
				}
				if e = encoder.Encode(success{"Ok", options{v.TabWidth, v.UseTabs, v.Semi, v.SingleQuote, v.PrintWidth, v.TrailingComma, v.BracketSpacing, v.BracketSameLine, v.ArrowParens, v.EndOfLine}, resolution.Source, hi, resolution.HouseIgnoreDeclared, ip}); e != nil {
					panic(e)
				}
			}
		}
	}
}
