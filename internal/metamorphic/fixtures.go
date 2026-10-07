package metamorphic

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
)

// Fixtures are the programs TestNativeAgreesWithNode runs that stage 0 lowers, read from the oracle's
// own list in internal/oracle/oracle_test.go under root, sorted by path. Reading the list (with Go's
// parser) rather than the testdata directory keeps to exactly what the oracle holds to Node.
func Fixtures(root string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "internal", "oracle", "oracle_test.go"), nil, 0)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, declaration := range file.Decls {
		general, isGeneral := declaration.(*ast.GenDecl)
		if !isGeneral || general.Tok != token.VAR {
			continue
		}
		for _, specification := range general.Specs {
			value, isValue := specification.(*ast.ValueSpec)
			if !isValue || len(value.Names) != 1 || value.Names[0].Name != "fixtures" || len(value.Values) != 1 {
				continue
			}
			literal, isLiteral := value.Values[0].(*ast.CompositeLit)
			if !isLiteral {
				continue
			}
			for _, element := range literal.Elts {
				entry, isEntry := element.(*ast.CompositeLit)
				if !isEntry || len(entry.Elts) < 2 {
					continue
				}
				path, isPath := entry.Elts[0].(*ast.BasicLit)
				lowers, isLowers := entry.Elts[1].(*ast.Ident)
				if !isPath || path.Kind != token.STRING || !isLowers || lowers.Name != "true" {
					continue
				}
				unquoted, err := strconv.Unquote(path.Value)
				if err != nil {
					return nil, err
				}
				paths = append(paths, unquoted)
			}
		}
	}
	if len(paths) == 0 {
		return nil, errors.New("metamorphic: no fixtures in internal/oracle/oracle_test.go")
	}
	sort.Strings(paths)
	return paths, nil
}

// Stride is count of paths chosen by a fixed stride over them, the first included.
func Stride(paths []string, count int) []string {
	if count <= 0 || count >= len(paths) {
		return paths
	}
	chosen := make([]string, 0, count)
	for index := range count {
		chosen = append(chosen, paths[index*len(paths)/count])
	}
	return chosen
}
