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
	sources = append(sources, "0123456789", "7777", "3777", "4000", "999999999999999999999999999999999999999999", "1048575", "1048576", "é123😀777", "123x45", "00000000123")
	write := func(source string, offset int) {
		raw := make([]int, len(source))
		for i, b := range []byte(source) {
			raw[i] = int(b)
		}
		must(encode.Encode([]any{raw, offset}))
		if mode == "decimal" {
			v, w := rx.AdamicWave8Decimal(source, offset)
			fmt.Fprintln(expected, v)
			fmt.Fprintln(expected, w)
		} else {
			v, w := rx.AdamicWave8Octal(source, offset)
			fmt.Fprintln(expected, v)
			fmt.Fprintln(expected, w)
		}
		count++
	}
	if mode != "covers" {
		for _, source := range sources {
			for i := 0; i < len(source); i++ {
				write(source, i)
			}
			if mode == "decimal" {
				write(source, len(source))
			}
		}
		for first := 0; first < 256; first++ {
			for _, suffix := range []string{"", "00", "77", "88", "7777", "012345678901234567890"} {
				write(string([]byte{byte(first)})+suffix, 0)
			}
		}
	} else {
		seenQueries := map[[4]int64]bool{}
		query := func(kind uint8, lo, hi, r rune) {
			key := [4]int64{int64(kind), int64(lo), int64(hi), int64(r)}
			if seenQueries[key] {
				return
			}
			seenQueries[key] = true
			must(encode.Encode([]any{kind, lo, hi, r}))
			fmt.Fprintln(expected, rx.AdamicWave8Covers(kind, lo, hi, r))
			count++
		}
		for _, source := range sources {
			for _, r := range source {
				for kind := uint8(0); kind < 6; kind++ {
					query(kind, r, r+2, r)
					query(kind, r, r+2, r+2)
					query(kind, r, r+2, r-1)
					query(kind, r, r+2, r+3)
				}
			}
		}
		for kind := uint8(0); kind < 6; kind++ {
			for _, r := range []rune{-2147483648, -1, 0, 1, 0xD800, 0x10FFFF, 2147483647} {
				query(kind, r, r, r)
				query(kind, 10, 1, r)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d %s queries\n", fixtureCount, count, mode)
}
