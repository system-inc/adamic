package oracle

import (
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeProcessMethodPresence(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"node_process_method_presence.a", "node_process_global_presence.a", "node_process_parser_presence.a"} {
		t.Run(fixture, func(t *testing.T) {
			path, binary, script := sanitized(t, "internal/oracle/testdata/"+fixture)
			truth := onNode(t, path)
			for _, got := range []run{execute(t, binary), onNode(t, script)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			for function := -1; function < len(program.Functions); function++ {
				graph := flow.Build(program, function)
				flow.Construct(graph)
				if violations := flow.VerifySSA(graph); len(violations) != 0 {
					t.Fatalf("flow: %v", violations)
				}
			}
			code := native.C(program)
			anchor := "adamic_node_next_tick_feature()"
			if !strings.Contains(code, anchor) {
				t.Fatal("descriptor feature anchor missing")
			}
			changed := strings.ReplaceAll(code, anchor, "false")
			mutant := filepath.Join(t.TempDir(), "absent")
			if err := native.Build(changed, mutant, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			bad := execute(t, mutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				t.Fatal("absent native descriptor must fail only Node bytes")
			}
			data, err := os.ReadFile(script)
			if err != nil {
				t.Fatal(err)
			}
			jsAnchor := "!!process.nextTick"
			if !strings.Contains(string(data), jsAnchor) {
				t.Fatal("JavaScript descriptor anchor missing")
			}
			jsMutant := filepath.Join(t.TempDir(), "absent.mjs")
			if err := os.WriteFile(jsMutant, []byte(strings.ReplaceAll(string(data), jsAnchor, "false")), 0600); err != nil {
				t.Fatal(err)
			}
			bad = onNode(t, jsMutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				t.Fatal("absent JavaScript descriptor must fail only Node bytes")
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(truth, onWASI(t, code)); difference != "" {
					t.Fatal("WASI: " + difference)
				}
				bad = onWASI(t, changed)
				if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
					t.Fatal("absent WASI descriptor must fail only Node bytes")
				}
			}
			t.Log("presence agrees with Node; absent descriptors run cleanly and differ only in stdout")
		})
	}
}
