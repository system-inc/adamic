package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type definition struct{ File, Text string }
type observation struct {
	Symbol, OldFile, NewFile, OldHash, NewHash string
	Identical                                  bool
	OldText, NewText                           string
}

func command(args ...string) []byte {
	data, e := exec.Command("git", args...).Output()
	if e != nil {
		panic(e)
	}
	return data
}
func receiver(n *ast.FuncDecl) string {
	if n.Recv == nil {
		return ""
	}
	var b bytes.Buffer
	if e := format.Node(&b, token.NewFileSet(), n.Recv.List[0].Type); e != nil {
		panic(e)
	}
	return strings.TrimPrefix(b.String(), "*")
}
func definitions(ref, dir string) map[string]definition {
	out := map[string]definition{}
	for _, file := range strings.Split(string(command("ls-tree", "-r", "--name-only", ref, "--", dir)), "\n") {
		if filepath.Dir(file) != dir || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			continue
		}
		fs := token.NewFileSet()
		tree, e := parser.ParseFile(fs, file, command("show", ref+":"+file), 0)
		if e != nil {
			panic(e)
		}
		for _, d := range tree.Decls {
			n, ok := d.(*ast.FuncDecl)
			if !ok || n.Name.Name == "init" {
				continue
			}
			key := n.Name.Name
			if r := receiver(n); r != "" {
				key = r + "." + key
			}
			var b bytes.Buffer
			if e := format.Node(&b, fs, n); e != nil {
				panic(e)
			}
			if _, exists := out[key]; exists {
				panic("ambiguous " + key)
			}
			out[key] = definition{file, b.String()}
		}
	}
	return out
}
func main() {
	data, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	var symbols []string
	if e = json.Unmarshal(data, &symbols); e != nil {
		panic(e)
	}
	old, new := map[string]map[string]definition{}, map[string]map[string]definition{}
	rows := []observation{}
	for _, symbol := range symbols {
		suffix := strings.TrimPrefix(symbol, "github.com/system-inc/cohere/")
		dot := strings.Index(suffix, ".")
		if dot < 0 {
			panic(symbol)
		}
		dir, name := suffix[:dot], strings.TrimPrefix(suffix[dot+1:], "*")
		if old[dir] == nil {
			old[dir] = definitions("715ba94f3608a6500086b1076ce5cb7e51b836db", dir)
			new[dir] = definitions("7945d102a6c18dd36adf9114a758ce646e8b2359", dir)
		}
		a, ok := old[dir][name]
		if !ok {
			panic("old definition missing " + symbol)
		}
		b, ok := new[dir][name]
		if !ok {
			panic("new definition missing " + symbol)
		}
		row := observation{Symbol: symbol, OldFile: a.File, NewFile: b.File, OldHash: fmt.Sprintf("%x", sha256.Sum256([]byte(a.Text))), NewHash: fmt.Sprintf("%x", sha256.Sum256([]byte(b.Text))), Identical: a.Text == b.Text}
		if !row.Identical {
			row.OldText = a.Text
			row.NewText = b.Text
		}
		rows = append(rows, row)
	}
	data, e = json.MarshalIndent(rows, "", "  ")
	if e != nil {
		panic(e)
	}
	fmt.Println(string(data))
}
