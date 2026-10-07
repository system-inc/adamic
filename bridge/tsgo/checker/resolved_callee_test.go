package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
	"testing"
)

func TestResolvedCalleeFacts(t *testing.T) {
	if uint32(checker.TypeFlagsNever) != 262144 {
		t.Fatal("pinned Never bit changed")
	}
	t.Parallel()
	program, file := wave20NextProgram(t)
	source := program.Compiler.GetSourceFile(file)
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	count := 0
	bodySeen := false
	ambientSeen := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "CallExpression", "resolved-callee")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			signature := c.GetResolvedSignature(node)
			if signature == nil {
				if len(fields) != 3 || fields[2] != "0" {
					t.Fatal(fields)
				}
			} else {
				if fields[2] != "1" || fields[3] != strconv.FormatUint(uint64(c.GetReturnTypeOfSignature(signature).Flags()), 10) {
					t.Fatal("return flags", fields)
				}
				decl := signature.Declaration()
				if decl == nil {
					if fields[4] != "0" {
						t.Fatal(fields)
					}
				} else {
					f := ast.GetSourceFileOfNode(decl)
					if fields[4] != "1" || fields[5] != f.FileName() || fields[7] != strings.TrimPrefix(decl.Kind.String(), "Kind") || fields[8] != strconv.Itoa(decl.Pos()) || fields[9] != strconv.Itoa(decl.End()) || fields[10] != strconv.FormatUint(uint64(ast.GetFunctionFlags(decl)), 10) {
						t.Fatal("resolved declaration", fields)
					}
					if fields[11] != "0" {
						bodySeen = true
					} else {
						ambientSeen = true
					}
				}
			}
			count++
			if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "CallExpression", "resolved-callee\nextra"); err == nil {
				t.Fatal("suffix accepted")
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if count < 5 || !bodySeen || !ambientSeen {
		t.Fatal("missing signature variants")
	}
	if _, err := program.Inspect(file, uint64(source.Pos()), uint64(source.End()), "SourceFile", "resolved-callee"); err == nil {
		t.Fatal("noncall accepted")
	}
}
