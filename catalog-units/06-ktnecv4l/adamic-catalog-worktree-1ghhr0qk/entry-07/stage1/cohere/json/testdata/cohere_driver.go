// Built through an overlay over command/formatter_comparison/main.go in the pinned cohere module.
// This gives the throughput test the same read/format/escaped-stdout protocol on all three sides.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/javascript"
)

func main() {
	source, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var output strings.Builder
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		text, err := javascript.FormatJSON(parts[0], unescape.Replace(parts[1]), formatoptions.PrettierDefaults())
		if err != nil {
			fmt.Fprintf(&output, "error\t%s\n", escape.Replace(err.Error()))
		} else {
			fmt.Fprintf(&output, "ok\t%s\n", escape.Replace(text))
		}
	}
	if _, err := os.Stdout.WriteString(output.String()); err != nil {
		panic(err)
	}
}
