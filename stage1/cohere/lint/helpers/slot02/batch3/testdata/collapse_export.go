package tailwind

import (
	"sort"
	"strings"
)

func AdamicSlot02MathFunctionName(name string) bool { return isMathFunctionName(name) }

// Probe every consumer input and actual Go value-parser function name, plus all
// accepted names with boundary/case controls derived from Go's own table.
func AdamicSlot02MathNames(texts []string) []string {
	names := map[string]bool{"": true, "url": true, "var": true, "attr": true, "calc(1px)": true, "CALC": true, "世界": true}
	for _, name := range mathFunctions {
		names[name] = true
		for _, suffix := range []string{"x", "-", " ", "(", ")", "\x00", "é", "😀"} {
			names[name+suffix] = true
			names[suffix+name] = true
		}
		for i := 0; i < len(name); i++ {
			names[name[:i]+strings.ToUpper(name[i:i+1])+name[i+1:]] = true
		}
	}
	var walk func([]ValueNode)
	walk = func(nodes []ValueNode) {
		for _, node := range nodes {
			if node.Kind == ValueNodeKindFunction {
				names[node.Value] = true
			}
			walk(node.Nodes)
		}
	}
	for _, text := range texts {
		names[text] = true
		walk(ParseValue(text))
	}
	result := []string{}
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}
