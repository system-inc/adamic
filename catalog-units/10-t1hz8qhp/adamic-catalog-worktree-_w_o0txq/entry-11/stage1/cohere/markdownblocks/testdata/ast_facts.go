package markdown

import (
	"bufio"
	"fmt"
	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
	"os"
	"strconv"
	"strings"
)

// AdamicASTFacts runs the real Go preprocess on the same initial AST transport as Node/native.
func AdamicASTFacts(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	unescape := func(text string) string {
		var result strings.Builder
		for i := 0; i < len(text); i++ {
			if text[i] != '\\' {
				result.WriteByte(text[i])
				continue
			}
			i++
			switch text[i] {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			default:
				result.WriteByte(text[i])
			}
		}
		return result.String()
	}
	integer := func(text string) int {
		n, e := strconv.Atoi(text)
		if e != nil {
			panic(e)
		}
		return n
	}
	var source string
	tab := 4
	var tree []*Node
	var children [][]int
	var positions map[int]*mdast.Position
	var offsets []int
	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Split(line, "\t")
		switch f[0] {
		case "A":
			source = unescape(f[1])
			tab = integer(f[2])
			tree = nil
			children = nil
			positions = map[int]*mdast.Position{}
			offsets = make([]int, 0, len(source)+1)
			for byteOffset, point := range source {
				offsets = append(offsets, byteOffset)
				if point > 0xffff {
					offsets = append(offsets, byteOffset)
				}
			}
			offsets = append(offsets, len(source))
		case "N":
			mask := integer(f[3])
			n := &Node{NodeType: f[1], Value: unescape(f[2]), IsParent: mask&2 != 0, Ordered: mask&4 != 0, IsIndented: mask&8 != 0, IsAligned: mask&16 != 0, IsLiteral: mask&64 != 0, IsCJ: mask&128 != 0, HasLeadingPunctuation: mask&256 != 0, HasTrailingPunctuation: mask&512 != 0, Kind: f[13]}
			if mask&1 != 0 {
				raw := unescape(f[4])
				n.Raw = &raw
			}
			if mask&32 != 0 {
				alt := unescape(f[5])
				n.OriginalAltText = &alt
			}
			if mask&1024 != 0 {
				id := integer(f[6])
				if positions[id] == nil {
					positions[id] = &mdast.Position{Start: mdast.Point{Offset: offsets[integer(f[7])], Line: integer(f[9]), Column: integer(f[11])}, End: mdast.Point{Offset: offsets[integer(f[8])], Line: integer(f[10]), Column: integer(f[12])}}
				}
				n.Position = positions[id]
			}
			ids := []int{}
			if f[14] != "" {
				for _, id := range strings.Split(f[14], ",") {
					ids = append(ids, integer(id))
				}
			}
			tree = append(tree, n)
			children = append(children, ids)
		case "F":
			for i, n := range tree {
				if n.IsParent {
					n.Children = []*Node{}
					for _, id := range children[i] {
						n.Children = append(n.Children, tree[id])
					}
				}
			}
			result := func() (answer string) {
				defer func() {
					if value := recover(); value != nil {
						answer = "E\t" + fmt.Sprint(value)
					}
				}()
				nodes := &arena.Arena[Node]{}
				return adamicAstSerialize(preprocess(tree[0], source, tab, nodes), source)
			}()
			fmt.Fprintln(output, adamicAstEscape(result))
		case "":
		default:
			panic("unknown AST fact")
		}
	}
}
