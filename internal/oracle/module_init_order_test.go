package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"stage3/drivers/parser/native-enum-map.a",
		"internal/oracle/testdata/module_init_order/enum.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestModuleInitOrder(t *testing.T) {
	// Cohere's cycle runner is process-wide; these probes run sequentially.
	for _, probe := range []struct {
		name, path string
		early      bool
	}{
		{"map-before-enum", "stage3/drivers/parser/native-enum-map.a", false},
		{"cycle", "internal/oracle/testdata/module_init_order/enum.a", false},
		{"early-read", "internal/oracle/testdata/module_init_order/early.a", true},
		{"construction-reads-enum", "internal/oracle/testdata/module_init_order/construction.a", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, probe.path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			compiled, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			if probe.early {
				// TypeScript enums use hoisted var storage. Node throws TypeError,
				// while the ruling requires Adamic's loud binding-readiness stop.
				if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError: Cannot read properties of undefined (reading 'Zero')") {
					t.Fatalf("Node: %+v", truth)
				}
				want := run{stderr: []byte("adamic: panic: ReferenceError: Cannot access 'Pending' before initialization\n"), exitCode: 70}
				for _, result := range []run{compiled, backend, released(t, program)} {
					if diff := disagreement(want, result); diff != "" {
						t.Fatalf("%s: %+v", diff, result)
					}
				}
			} else {
				output := "0\n"
				if probe.name == "cycle" {
					output = "map\nenum\n0\n"
				}
				if truth.exitCode != 0 || string(truth.stdout) != output {
					t.Fatalf("Node: %+v", truth)
				}
				if strings.Contains(native.C(program), "ReferenceError: Cannot access 'Pending'") {
					t.Fatal("proven enum read retained a readiness check")
				}
				for _, result := range []run{compiled, backend, released(t, program)} {
					if diff := disagreement(truth, result); diff != "" {
						t.Fatalf("%s: %+v", diff, result)
					}
				}
			}
			t.Logf("Node stdout=%q stderr=%q exit=%d; native stdout=%q stderr=%q exit=%d", truth.stdout, truth.stderr, truth.exitCode, compiled.stdout, compiled.stderr, compiled.exitCode)
		})
	}
}
