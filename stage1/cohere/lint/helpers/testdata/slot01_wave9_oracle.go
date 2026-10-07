package main

import (
	"encoding/json"
	"fmt"
	rx "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strconv"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	var paths []string
	data, err := os.ReadFile(os.Args[1])
	must(err)
	must(json.Unmarshal(data, &paths))
	cases, err := os.Create(os.Args[2])
	must(err)
	defer cases.Close()
	expected, err := os.Create(os.Args[3])
	must(err)
	defer expected.Close()
	encode := json.NewEncoder(cases)
	sources := []string{"", `const e=<div href={id} href="later" />;`, `const e=<div href="" HREF="&#47;about" {...props} xlink:href="other" />;`, `const e=<div href href="later" />;`, `const e=<div href={"expression"} />;`}
	fixtureCount := 0
	for _, path := range paths {
		file, err := goparser.ParseFile(token.NewFileSet(), path, nil, 0)
		must(err)
		constants := map[string]goast.Expr{}
		goast.Inspect(file, func(node goast.Node) bool {
			if declaration, ok := node.(*goast.ValueSpec); ok && len(declaration.Names) == len(declaration.Values) {
				for i, name := range declaration.Names {
					constants[name.Name] = declaration.Values[i]
				}
			}
			return true
		})
		var value func(goast.Expr, int) (string, bool)
		value = func(expression goast.Expr, depth int) (string, bool) {
			if depth > 32 {
				return "", false
			}
			switch e := expression.(type) {
			case *goast.BasicLit:
				if e.Kind == token.STRING {
					s, err := strconv.Unquote(e.Value)
					return s, err == nil
				}
			case *goast.BinaryExpr:
				if e.Op == token.ADD {
					a, ok := value(e.X, depth+1)
					b, ok2 := value(e.Y, depth+1)
					return a + b, ok && ok2
				}
			case *goast.Ident:
				if e, ok := constants[e.Name]; ok {
					return value(e, depth+1)
				}
			case *goast.ParenExpr:
				return value(e.X, depth+1)
			}
			return "", false
		}
		found := 0
		seen := map[string]bool{}
		goast.Inspect(file, func(node goast.Node) bool {
			expression, ok := node.(goast.Expr)
			if !ok {
				return true
			}
			source, ok := value(expression, 0)
			if ok {
				if source != "" && !seen[source] {
					sources = append(sources, source)
					seen[source] = true
					found++
				}
				return false // Preserve complete concatenations rather than parsing their fragments.
			}
			return true
		})
		if found == 0 {
			panic("no string inputs in " + path)
		}
		fmt.Fprintf(os.Stderr, "%s: %d string inputs\n", path, found)
		fixtureCount += found
	}

	count := 0
	mode := os.Args[4]
	sources = append(sources, "*", "+", "?", "*?", "+?", "??", "???", "{0}", "{1,}", "{1,2}?", "{01,002}", "{2,1}", "{}", "{,1}", "{,}", "{1,2,3}", "{1", "{1,", "x12}", "é12}", "{１２}", "{999999999999999999999999999999999999999}", "(?=", "(?!", "(?<=", "(?<!", "(?<name>", "(?", "(?<", "(\\?=", "😀(?<=x)", "(?=x)+?", "(?i:)")
	write := func(source string, offset int) {
		raw := make([]int, len(source))
		for i, b := range []byte(source) {
			raw[i] = int(b)
		}
		must(encode.Encode([]any{raw, offset}))
		fmt.Fprintln(expected, rx.AdamicWave9Width(source, offset, mode))
		count++
	}
	for _, source := range sources {
		for i := 0; i <= len(source); i++ {
			write(source, i)
		}
	}
	for first := 0; first < 256; first++ {
		for _, suffix := range []string{"", "12}", "?", "=", "!", "<=", "<!"} {
			write(string([]byte{byte(first)})+suffix, 0)
		}
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d %s queries\n", fixtureCount, count, mode)
}
