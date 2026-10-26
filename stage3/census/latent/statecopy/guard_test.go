package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatentStatecopySortedMapProgramMembership(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, source := range map[string]string{
		"internal/ir/map.go":               "package ir; type SortedMap[K comparable,V any] map[K]V",
		"internal/lower/lower.go":          "package lower; type lowering struct {plan *programMembership}",
		"internal/lower/program_region.go": "package lower; type programMembership struct {types ir.SortedMap[int, []bool]}",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	defer func() {
		if failure := recover(); failure != nil {
			t.Fatalf("program_region.go: programMembership.types: %v", failure)
		}
	}()
	result := string(generate(root, filepath.Join(root, "internal/lower/lower.go")))
	if !strings.Contains(result, "make(ir.SortedMap[int, []bool]") || !strings.Contains(result, "make([]bool") {
		t.Fatal(result)
	}
}

// The overlay adds only maps of checker-owned nodes and bools to lowering.
// All production state follows the same root and recursive struct decisions as generation.
func TestLatentStatecopyCoversEveryLoweringStateField(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/*.go
	// Reads: internal/ir/*.go
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	g := loadGenerator(root, filepath.Join(root, "internal/lower/lower.go"))
	done := map[string]bool{}
	for {
		progressed := false
		for name := range g.needed {
			if done[name] {
				continue
			}
			done[name] = true
			progressed = true
			structure := g.types[name].(*ast.StructType)
			for _, field := range structure.Fields.List {
				for _, id := range field.Names {
					if name == "lowering" && id.Name == "impliedTargetQuery" && printed(field.Type) == "func(*ast.Node) *checker.Type" {
						continue
					}
					func() {
						defer func() {
							if failure := recover(); failure != nil {
								t.Errorf("%s.%s (%s): %v", name, id.Name, fieldSource(root, name), failure)
							}
						}()
						g.copy(field.Type, "value."+id.Name)
					}()
				}
			}
		}
		if !progressed {
			break
		}
	}
}

func fieldSource(root, name string) string {
	paths, _ := filepath.Glob(filepath.Join(root, "internal/lower/*.go"))
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		positions := token.NewFileSet()
		file, err := parser.ParseFile(positions, path, nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					if spec, ok := spec.(*ast.TypeSpec); ok && spec.Name.Name == name {
						return positions.Position(spec.Pos()).String()
					}
				}
			}
		}
	}
	return name
}
