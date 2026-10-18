package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Each subprocess owns its stdout/stderr; other driver tests capture those globals.
func TestExplainChecksOutput(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"overload_some_empty", "overload_some_false_read", "overload_some_true_only", "overload_erased", "overload_assertion"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			want, err := os.ReadFile("testdata/checks/" + fixture + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			path := "internal/lower/testdata/predicates/" + fixture + ".a"
			for _, backend := range []string{"c", "js", "build"} {
				t.Run(backend, func(t *testing.T) {
					t.Parallel()
					arguments := []string{backend, path, "--explain-checks"}
					if backend == "build" {
						arguments = append(arguments, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
					}
					command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
					command.Dir = repository
					command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1", "XDG_CACHE_HOME="+t.TempDir())
					var output, diagnostic bytes.Buffer
					command.Stdout, command.Stderr = &output, &diagnostic
					if err := command.Run(); err != nil {
						t.Fatalf("%s: %v, %s", backend, err, diagnostic.Bytes())
					}
					if !bytes.Equal(diagnostic.Bytes(), want) {
						t.Fatalf("%s explanation:\n%s\nwant:\n%s", backend, diagnostic.Bytes(), want)
					}
					if backend != "build" && output.Len() == 0 {
						t.Fatalf("%s lost generated source", backend)
					}
					if backend == "build" && output.Len() != 0 {
						t.Fatalf("build wrote report to stdout: %s", output.Bytes())
					}
				})
			}
		})
	}
}

// Not parallel: run writes process-wide stdout/stderr and native.runtimeBuilds in the driver process.
func TestExplainChecksDriver(t *testing.T) {
	if os.Getenv("ADAMIC_EXPLAIN_TEST_DRIVER") != "1" {
		return
	}
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	t.Fatal("missing driver arguments")
}

func explainCheckedWritesOutput(t *testing.T, fixture struct{ name, expression string }) {
	t.Helper()
	source, err := os.ReadFile("../../stage3/checked-writes/" + fixture.name + ".ts")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), fixture.name+".ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"c", "js", "build"} {
		arguments := []string{backend, path, "--explain-checks"}
		if backend == "build" {
			arguments = append(arguments, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
		}
		command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
		command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
		var output, diagnostic bytes.Buffer
		command.Stdout, command.Stderr = &output, &diagnostic
		if err = command.Run(); err != nil {
			t.Fatalf("%s: %v %s", backend, err, diagnostic.Bytes())
		}
		report := diagnostic.String()
		if !strings.Contains(report, ": checked write: "+fixture.expression+" against actual field contract\n") || !strings.Contains(report, "adamic: write checks: checked 1\n") {
			t.Fatalf("%s lost checked write explanation: %s", backend, report)
		}
	}
}

func TestExplainCheckedWritesOutput_emit_node_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"emit-node-misfit", "view.emitNode"})
}

func TestExplainCheckedWritesOutput_number_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"number-misfit", "view.count"})
}

func TestExplainCheckedWritesOutput_flow_node_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"flow-node-misfit", "view.node"})
}

func TestExplainCheckedWritesOutput_container_number_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"container-number-misfit", "values[0]"})
}

func TestExplainCheckedWritesOutput_container_map_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"container-map-misfit", "values[value]"})
}

func TestExplainCheckedWritesOutput_container_fill_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"container-fill-misfit", "values[]"})
}

func TestExplainCheckedWritesOutput_optional_boolean_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"optional-boolean-misfit", "view.multiLine"})
}

func TestExplainCheckedWritesOutput_overload_result_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"overload-result-misfit", "createTempVariable(true).emitNode"})
}

func TestExplainCheckedWritesOutput_overload_callback_misfit(t *testing.T) {
	t.Parallel()
	explainCheckedWritesOutput(t, struct{ name, expression string }{"overload-callback-misfit", "factory(true).emitNode"})
}
