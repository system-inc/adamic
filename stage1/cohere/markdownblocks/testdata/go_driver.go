package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/markdown"
	"github.com/system-inc/cohere/internal/format/native"
)

type input struct {
	Name string `json:"name"`
	Text string `json:"text"`
}
type output struct {
	Name      string   `json:"name"`
	Auto      string   `json:"auto"`
	Off       string   `json:"off"`
	AutoError string   `json:"autoError,omitempty"`
	OffError  string   `json:"offError,omitempty"`
	Embeds    []string `json:"embeds"`
}

func main() {
	source, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer source.Close()
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 65536), 32*1024*1024)
	encoder := json.NewEncoder(os.Stdout)
	options := formatoptions.Default()
	formatter := native.Formatter{Options: options}
	offOnly := len(os.Args) > 2 && os.Args[2] == "off-only"
	for scanner.Scan() {
		var in input
		if err := json.Unmarshal(scanner.Bytes(), &in); err != nil {
			panic(err)
		}
		result := output{Name: in.Name}
		if !offOnly {
			result.Embeds = markdown.EmbeddedParsers(in.Text)
			result.Auto, err = formatter.Format("fixture.md", in.Text)
			if err != nil {
				result.AutoError = err.Error()
			}
		}
		// The off baseline uses the same real Markdown printer, with no embedding callback.
		text := strings.TrimPrefix(in.Text, "\ufeff")
		text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
		result.Off, err = markdown.Format(text, options, nil)
		if err != nil {
			result.OffError = err.Error()
		}
		if strings.HasPrefix(in.Text, "\ufeff") {
			result.Off = "\ufeff" + result.Off
		}
		if err := encoder.Encode(result); err != nil {
			panic(err)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
