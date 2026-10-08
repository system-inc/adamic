// Built by overlay inside cohere, so the benchmark uses its real file wrapper.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	formatter := native.Formatter{Options: formatoptions.Default()}
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		formatted, err := formatter.Format("probe.graphql", unescape.Replace(line[1:]))
		if err != nil {
			fmt.Fprintln(out, "error\t"+escape.Replace(err.Error()))
		} else {
			fmt.Fprintln(out, "ok\t"+escape.Replace(formatted))
		}
	}
}
