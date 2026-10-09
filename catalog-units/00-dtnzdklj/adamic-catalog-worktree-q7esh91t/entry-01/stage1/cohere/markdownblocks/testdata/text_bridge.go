package markdown

import (
	"fmt"
	"strings"
)

func AdamicTextFixture(text string) string {
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	fields := []string{}
	flag := func(value bool) int {
		if value {
			return 1
		}
		return 0
	}
	for _, node := range splitText(text, nil) {
		fields = append(fields, fmt.Sprintf("%s\t%s\t%d,%d,%d\t%s", node.NodeType, node.Kind, flag(node.IsCJ), flag(node.HasLeadingPunctuation), flag(node.HasTrailingPunctuation), escape.Replace(node.Value)))
	}
	return escape.Replace(strings.Join(fields, "\n"))
}
