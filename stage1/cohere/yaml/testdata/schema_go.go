// This command only exposes Go cohere's scalar resolver through a source overlay.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/cohere/internal/format/yaml/compose"
)

func main() {
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	number := 0
	tags := []*compose.Tag{}
	for _, version := range []string{"1.2", "1.1"} {
		options := compose.UnistParserOptions()
		options.Version = version
		for document := range compose.NewComposer(options).Compose(nil, true, 0) {
			for _, tag := range document.Schema.Tags {
				if tag.Test != nil {
					tags = append(tags, tag)
				}
			}
		}
	}
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		text := unescape.Replace(parts[1])
		// ECMAScript trim, including BOM and excluding NEL.
		text = strings.TrimFunc(text, func(r rune) bool {
			return r == '\ufeff' || r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' || r == ' ' || r == '\u00a0' || r == '\u1680' || r >= '\u2000' && r <= '\u200a' || r == '\u2028' || r == '\u2029' || r == '\u202f' || r == '\u205f' || r == '\u3000'
		})
		fmt.Fprintf(out, "%d|", number)
		number++
		for _, tag := range tags {
			if tag.Test.MatchString(text) {
				fmt.Fprint(out, "1")
			} else {
				fmt.Fprint(out, "0")
			}
		}
		fmt.Fprintln(out)
	}
}
