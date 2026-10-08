package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"narrowed", "present", "undefined", "null", "let_initializer"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_checked_" + name + ".ts", true, name != "narrowed" && name != "present"})
	}
}

func TestCheckedNonNullTypeScript(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"narrowed", "present", "undefined", "null", "initializer", "let_initializer"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_checked_"+name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if name == "initializer" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || refused.What != "var" || refused.Fix != "use const or let" {
					t.Fatalf("want area's var refusal, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			counts := program.NonNullChecks
			if name == "narrowed" {
				if counts.Proven != 1 || counts.Checked != 0 {
					t.Fatalf("want proven 1 checked 0, got %#v", counts)
				}
			} else if counts.Proven != 0 || counts.Checked != 1 {
				t.Fatalf("want proven 0 checked 1, got %#v", counts)
			}
			node := onNode(t, path)
			want := node
			if name == "undefined" || name == "null" || name == "let_initializer" {
				source := "before\nundefined\n"
				if name == "undefined" {
					source = "before\ntrue\n"
				}
				column := 24
				line := 3
				if name == "null" {
					source = "before\ntrue\n"
					column = 24
				}
				if name == "let_initializer" {
					line = 3
					column = 26
				}
				if difference := disagreement(run{stdout: []byte(source)}, node); difference != "" {
					t.Fatal("Node: " + difference)
				}
				site := counts.Sites[0]
				suffix := ":" + strconv.Itoa(line) + ":" + strconv.Itoa(column)
				if !strings.HasSuffix(site.Where, suffix) || site.Expression != "value!" {
					t.Fatalf("wrong assertion site %#v, want %s", site, suffix)
				}
				want = run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed at " + path + suffix + ": value! is null or undefined\n"), exitCode: 70}
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: got %#v want %#v", difference, got, want)
				}
			}
			if name == "undefined" {
				changes := 0
				mutate := func(node any) any {
					if value, ok := node.(ir.Coalesce); ok && value.Panic != nil {
						changes++
						return value.Value
					}
					return node
				}
				for i := range program.Functions {
					program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, mutate)
				}
				if changes != 1 {
					t.Fatalf("no-check mutant changed %d checks", changes)
				}
				mutant, _ := nativelyUncached(t, program)
				if mutant.exitCode != 0 || disagreement(want, mutant) == "" {
					t.Fatalf("no-check mutant escaped: %#v", mutant)
				}
				t.Logf("no-check mutant caught: stdout %q exit %d", mutant.stdout, mutant.exitCode)
			}
			if name == "narrowed" {
				changes := 0
				for i := range program.Functions {
					if program.Functions[i].Name == "narrowed" {
						program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
							if returned, ok := node.(ir.Return); ok {
								if call, ok := returned.Value.(ir.Call); ok && call.Returns == ir.Number && len(call.Arguments) == 1 && call.Arguments[0].Type() == ir.Union {
									returned.Value = call.Arguments[0]
									changes++
									return returned
								}
							}
							return node
						})
					}
				}
				if changes != 1 {
					t.Fatalf("representation mutant changed %d narrowed returns", changes)
				}
				caught := false
				for _, function := range program.Functions {
					if function.Name == "narrowed" {
						mutateReadiness(function.Body, func(node any) any {
							if returned, ok := node.(ir.Return); ok && returned.Value.Type() != function.Returns {
								caught = true
							}
							return node
						})
					}
				}
				if !caught {
					t.Fatal("representation-changing narrowing escaped the return representation assertion")
				}
				t.Log("look-through mutant caught before either backend: union return where number is required")
			}
		})
	}
}

func TestCheckedNonNullAdamicRefusal(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_refuse_checked.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "the non-null assertion !" || refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
		t.Fatalf("wrong .a refusal: %v", err)
	}
}

// Measure this unit's new rows independently while the refusal-table worker retires
// historical .a rows that the new source-mode policy no longer admits.
func TestCheckedNonNullCounts(t *testing.T) {
	t.Parallel()
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"narrowed", "present", "undefined", "null", "let_initializer"} {
		path := "internal/oracle/testdata/non_null_checked_" + name + ".ts"
		row := counted(t, path, false, nil, false, false)
		t.Log(row)
		if !strings.Contains(string(recorded), row+"\n") {
			t.Errorf("missing recorded row: %s", row)
		}
	}
}
