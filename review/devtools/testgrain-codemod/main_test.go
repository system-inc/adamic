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

func fixture(t *testing.T, name, source string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestPreparationAndShardClock(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"node_table_split_test.go", "node_table_oracle_floor_test.go"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file := fixture(t, name, `package sample
import ("testing"; "time")
func nodeTableSetup(t *testing.T) {
 once.Do(func() { started := time.Now(); prepare(); t.Logf("setup: %s", time.Since(started)) })
}
func nodeTableRunShard(t *testing.T) {
 nodeTableSetup(t)
 started := time.Now()
 timer := time.AfterFunc(time.Minute, func() { panic("cooked: killed") })
 defer timer.Stop()
 work()
 elapsed := time.Since(started)
 t.Logf("cooked=%t", elapsed > time.Minute)
 if elapsed > time.Minute { t.Fatal("cooked: budget") }
}
`)
			rewrite(file)
			raw, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(raw)
			for _, forbidden := range []string{"time.AfterFunc", "time.Now", "cooked", "timer.Stop"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("retained %s: %s", forbidden, text)
				}
			}
			if !strings.Contains(text, "testgrain.Setup(") || !strings.Contains(text, "prepare()") || !strings.Contains(text, "once.Do(") {
				t.Fatalf("lost preparation: %s", text)
			}
			f, err := parser.ParseFile(token.NewFileSet(), file, raw, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				body := syntax(fn.Body)
				if fn.Name.Name == "nodeTableSetupPrepare" && strings.Contains(body, "testgrain.Unit") {
					t.Fatal("setup is on the shard clock")
				}
				if fn.Name.Name == "nodeTableRunShard" {
					setup, unit, work := strings.Index(body, "nodeTableSetup(t)"), strings.Index(body, "testgrain.Unit(t)"), strings.Index(body, "work()")
					if setup < 0 || unit <= setup || work <= unit {
						t.Fatalf("wrong clock boundary: %s", body)
					}
				}
			}
		})
	}
}

func TestCommandKeepsVariadicArguments(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"text_shards_test.go", "text_independent_shards_test.go"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file := fixture(t, name, `package sample
import "testing"
func textBounded(t *testing.T, name string, arguments ...string) any { return nil }
func textPhaseDeadline(t *testing.T) func() { return nil }
`)
			rewrite(file)
			f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			ast.Inspect(f, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "textPhaseDeadline" {
					t.Error("deadline helper survived")
				}
				if call, ok := n.(*ast.CallExpr); ok && syntax(call.Fun) == "testgrain.CommandContext" {
					found = true
					if call.Ellipsis == 0 || len(call.Args) != 4 {
						t.Error("lost variadic forwarding")
					}
				}
				return true
			})
			if !found {
				t.Fatal("command was not delegated")
			}
		})
	}
}

func TestRejectsAlreadyConvertedWithoutWriting(t *testing.T) {
	t.Parallel()
	source := "package sample\nimport \"github.com/system-inc/adamic/internal/testgrain\"\n"
	file := fixture(t, "node_table_split_test.go", source)
	defer func() {
		if recover() == nil {
			t.Error("accepted already converted file")
		}
		got, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != source {
			t.Error("changed rejected file")
		}
	}()
	rewrite(file)
}
