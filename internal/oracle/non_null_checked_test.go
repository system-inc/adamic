package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"narrowed", "present", "undefined", "null", "initializer", "let_initializer"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_checked_" + name + ".ts", true, name != "narrowed" && name != "present"})
	}
}

func TestCheckedNonNullTypeScript(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"narrowed", "present", "undefined", "null", "initializer", "let_initializer"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/non_null_checked_"+name+".ts"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)

			if loweredError != nil {
				test.Fatal(loweredError)
			}
			counts := program.NonNullChecks
			if name == "narrowed" {
				if counts.Proven != 1 || counts.Checked != 0 {
					test.Fatalf("want proven 1 checked 0, got %#v", counts)
				}
			} else if counts.Proven != 0 || counts.Checked != 1 {
				test.Fatalf("want proven 0 checked 1, got %#v", counts)
			}
			node := onNode(test, path)
			want := node
			if name == "undefined" || name == "null" || (name == "initializer" || name == "let_initializer") {
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
				if name == "initializer" || name == "let_initializer" {
					line = 3
					column = 26
				}
				if difference := disagreement(run{stdout: []byte(source)}, node); difference != "" {
					test.Fatal("Node: " + difference)
				}
				site := counts.Sites[0]
				suffix := ":" + strconv.Itoa(line) + ":" + strconv.Itoa(column)
				if !strings.HasSuffix(site.Where, suffix) || site.Expression != "value!" {
					test.Fatalf("wrong assertion site %#v, want %s", site, suffix)
				}
				want = run{stdout: []byte("before\n"), stderr: []byte(
					"adamic: panic: non-null assertion failed at " + path + suffix + ": value! is null or undefined\n",
				), exitCode: 70}
			}
			native, _ := nativelyUncached(test, program)
			for _, got := range []run{native, onJavaScriptBackend(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf("%s: got %#v want %#v", difference, got, want)
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
					test.Fatalf("no-check mutant changed %d checks", changes)
				}
				mutant, _ := nativelyUncached(test, program)
				if mutant.exitCode != 0 || disagreement(want, mutant) == "" {
					test.Fatalf("no-check mutant escaped: %#v", mutant)
				}
				test.Logf("no-check mutant caught: stdout %q exit %d", mutant.stdout, mutant.exitCode)
			}
			if name == "narrowed" {
				changes := 0
				for i := range program.Functions {
					if program.Functions[i].Name == "narrowed" {
						program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
							if returned, ok := node.(ir.Return); ok {
								if call, ok := returned.Value.(ir.Call); ok &&
									call.Returns == ir.Number &&
									len(call.Arguments) == 1 &&
									call.Arguments[0].Type() == ir.Union {
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
					test.Fatalf("representation mutant changed %d narrowed returns", changes)
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
					test.Fatal("representation-changing narrowing escaped the return representation assertion")
				}
				test.Log("look-through mutant caught before either backend: union return where number is required")
			}
		})
	}
}

func TestCheckedNonNullAdamicRefusal(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/non_null_refused/non_null_refuse_checked.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	_, loweredError := lowered(test, path)
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		refused.What != "the non-null assertion !" ||
		refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
		test.Fatalf("wrong .a refusal: %v", loweredError)
	}
}

// Measure the new rows independently without rewriting the area's existing counts.
func TestCheckedNonNullCounts(test *testing.T) {
	test.Parallel()
	recorded, recordedCountsError := recordedCounts()
	if recordedCountsError != nil {
		test.Fatal(recordedCountsError)
	}
	for _, name := range []string{"narrowed", "present", "undefined", "null", "initializer", "let_initializer"} {
		path := "internal/oracle/testdata/non_null_checked_" + name + ".ts"
		_, measure := counted(test, path, false, nil, false, false)
		row := measure()
		test.Log(row)
		if recorded[path] != row {
			test.Errorf("missing recorded row: %s", row)
		}
	}
	for _, operand := range []string{"undefined", "null"} {
		path := "internal/oracle/testdata/non_null_refuse_" + operand + ".ts"
		_, measure := counted(test, path, false, nil, false, false)
		row := measure()
		test.Log(row)
		if recorded[path] != row {
			test.Errorf("missing recorded row: %s", row)
		}
	}
}
