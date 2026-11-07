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
	file, parseFileError := parser.ParseFile(fs, "overlay.go", source, parser.ParseComments)
	if parseFileError != nil {
		return nil, parseFileError
	}
	for _, spec := range append(file.Imports[:0:0], file.Imports...) {
		path, unquoteError := strconv.Unquote(spec.Path.Value)
		if unquoteError != nil {
			return nil, unquoteError
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
	if nodeError := format.Node(&output, fs, file); nodeError != nil {
		return nil, nodeError
	}
	return output.Bytes(), nil
}

func main() {
	for _, path := range os.Args[1:] {
		source, readError := os.ReadFile(path)
		if readError != nil {
			panic(readError)
		}
		output, cleanError := clean(source)
		if cleanError != nil {
			panic(cleanError)
		}
		if writeError := os.WriteFile(path, output, 0600); writeError != nil {
			panic(writeError)
		}
	}
}
