package main

import (
	"fmt"
	"github.com/system-inc/cohere/internal/format/doc"
	"os"
	"strconv"
	"strings"
)

func main() {
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	decode := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`)
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	integer := func(text string) int {
		value, err := strconv.Atoi(text)
		if err != nil {
			panic(err)
		}
		return value
	}
	var documents []doc.Doc
	groups := map[int]*doc.GroupID{}
	for _, line := range strings.Split(string(source), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if fields[0] == "R" {
			value := doc.Print(documents[integer(fields[1])], doc.Options{PrintWidth: 120, TabWidth: 4})
			if fields[2] == "1" {
				value = "\ufeff" + value
			}
			fmt.Println(encode.Replace(value))
			documents = nil
			groups = map[int]*doc.GroupID{}
			continue
		}
		if fields[0] != "D" {
			panic("canonical opcode")
		}
		kind, text, width, flags, gid := fields[1], decode.Replace(fields[2]), integer(fields[3]), integer(fields[4]), integer(fields[5])
		var children []doc.Doc
		if fields[6] != "" {
			for _, child := range strings.Split(fields[6], ",") {
				children = append(children, documents[integer(child)])
			}
		}
		var id *doc.GroupID
		if gid >= 0 {
			id = groups[gid]
			if id == nil {
				id = doc.NewGroupID("fixture")
				groups[gid] = id
			}
		}
		var built doc.Doc
		switch kind {
		case "t":
			built = doc.Text(text)
		case "a":
			built = doc.Concat(children)
		case "h":
			built = &doc.Line{Soft: flags&1 != 0, Hard: flags&2 != 0, Literal: flags&4 != 0}
		case "i":
			built = doc.NewIndent(children[0])
		case "s":
			built = doc.AlignWithString(text, children[0])
		case "w":
			built = doc.NewAlign(width, children[0])
		case "r":
			built = doc.MarkAsRoot(children[0])
		case "d":
			built = doc.DedentToRoot(children[0])
		case "f":
			built = doc.NewFill(children)
		case "g":
			options := doc.GroupOptions{ID: id, ShouldBreak: flags&1 != 0}
			if flags&2 != 0 {
				options.ExpandedStates = children
			}
			built = doc.NewGroup(children[0], options)
		case "b":
			built = doc.NewIfBreak(children[0], children[1], id)
		case "l":
			built = doc.NewLabel("fixture", children[0])
		case "p":
			built = doc.BreakParent
		case "x":
			built = doc.Trim
		default:
			panic("canonical doc kind")
		}
		documents = append(documents, built)
	}
}
