package markdown

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func AdamicPathFacts(path string) {
	data, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	integer := func(s string) int {
		n, e := strconv.Atoi(s)
		if e != nil {
			panic(e)
		}
		return n
	}
	var nodes []*Node
	var children [][]int
	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Split(line, "\t")
		switch f[0] {
		case "A":
			nodes = nil
			children = nil
		case "N":
			nodes = append(nodes, &Node{NodeType: f[1], IsParent: integer(f[3])&2 != 0})
			ids := []int{}
			if f[14] != "" {
				for _, id := range strings.Split(f[14], ",") {
					ids = append(ids, integer(id))
				}
			}
			children = append(children, ids)
		case "F":
			for i, n := range nodes {
				if n.IsParent {
					n.Children = []*Node{}
					for _, id := range children[i] {
						n.Children = append(n.Children, nodes[id])
					}
				}
			}
			fmt.Fprintln(output, adamicAstEscape(adamicPathObserve(nodes[0])))
		case "":
		default:
			panic("unknown path fact")
		}
	}
}
