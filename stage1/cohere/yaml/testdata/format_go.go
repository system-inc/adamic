// Public Go cohere file formatter, with Prettier defaults as in the JSON slice.
package main

import (
	"fmt"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"os"
	"strings"
)

func main() {
	source, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var out strings.Builder
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		text := unescape.Replace(parts[1])
		formatted, err := (native.Formatter{Options: formatoptions.PrettierDefaults()}).Format("input.yaml", text)
		if err != nil {
			fmt.Fprintf(&out, "error\t%s\n", escape.Replace(err.Error()))
		} else {
			fmt.Fprintf(&out, "ok\t%s\n", escape.Replace(formatted))
		}
	}
	if _, err := os.Stdout.WriteString(out.String()); err != nil {
		panic(err)
	}
}
