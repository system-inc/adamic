// importclean removes imports left unused by scratch overlay rewrites.
package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func clean(source []byte) ([]byte, error) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "overlay.go", source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	for _, spec := range append(file.Imports[:0:0], file.Imports...) {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		// Blank and dot imports cannot be judged by selector use; preserve their semantics.
		if spec.Name != nil && (spec.Name.Name == "_" || spec.Name.Name == ".") {
			continue
		}
		if !astutil.UsesImport(file, path) {
			name := ""
			if spec.Name != nil {
				name = spec.Name.Name
			}
			astutil.DeleteNamedImport(fs, file, name, path)
		}
	}
	var output bytes.Buffer
	if err := format.Node(&output, fs, file); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func main() {
	for _, path := range os.Args[1:] {
		source, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		output, err := clean(source)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, output, 0600); err != nil {
			panic(err)
		}
	}
}
