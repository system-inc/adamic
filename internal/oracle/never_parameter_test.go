package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestNeverParameterRuling(t *testing.T) {
	t.Parallel()
	const storage = "a value of type never, never (no value has it)"
	const abi = "an override with a different native representation for parameter"
	const condition = "a call whose argument for a never parameter isn't a rulings.md:76 never read"
	const variance = "adamic/contravariant-override"
	const bivariant = "tsc relates a method's parameters both ways"
	for _, pin := range []struct {
		name, stdout, location, reason string
		refused                        bool
	}{
		{"n01", "empty string", "7:2", abi, false},
		{"n02", "literal string", "5:2", bivariant, true},
		{"n03", "never1 string", "7:26", bivariant, true},
		{"n04", "sn", "", "", false},
		{"n05", "never1 string", "4:20", "adamic/no-unchecked-cast", true},
		{"n07", "dead string", "8:2", abi, false},
		{"n08", "empty", "7:2", variance, true},
		{"n09", "empty number", "7:2", variance, true},
		{"n10", "dead", "9:18", bivariant, true},
		{"n11", "dead", "8:2", variance, true},
		{"n12", "empty", "7:2", variance, true},
		{"callback", "received", "1:17", storage, false},
		{"dot-call", "sn", "1:17", storage, false},
		{"never-result", "fail", "8:21", condition, false},
		{"self-reference", "sn", "1:15", storage, false},
		{"stored", "sn", "1:17", storage, false},
		{"switch", "string number", "", "", false},
		{"default-optional", "sn", "", "", false},
		{"ternary", "hello bool", "", "", false},
		{"r_increment", "sn", "1:15", storage, false},
		{"r_assign", "sn", "1:15", storage, false},
		{"r_logical", "sn", "1:15", storage, false},
		{"r_destructure", "sn", "1:15", storage, false},
		{"r_for_of", "sn", "1:15", storage, false},
		{"r_nested_decrement", "sn", "1:15", storage, false},
		{"r_second", "sn", "", "", false},
		{"r_second_unchecked", "sn", "10:18", condition, false},
		{"c01", "caught assertNever reached with object", "3:52", "adamic/no-unchecked-cast", true},
		{"c02", "caught assertNever reached with object", "4:7", "error TS2339: Property 'captureStackTrace' does not exist", false},
		{"c02_node", "caught assertNever reached with object", "5:1", "node:globals.ErrorConstructor.captureStackTrace written", false},
		{"c03", "getFunction saw none", "14:8", "error TS2339: Property 'captureStackTrace' does not exist", false},
		{"c03_node", "getFunction saw none", "6:1", "node:globals.ErrorConstructor.prepareStackTrace", false},
		{"f_alias", "sn", "6:6", "captureStackTrace through a receiver other than the global Error", false},
		{"f_namespace_alias", "sn", "6:6", "captureStackTrace through a receiver other than the global Error", false},
		{"f_arguments", "sn", "6:3", "arguments other than a read of arguments.length", true},
		{"f_apply", "sn", "10:22", storage, false},
		{"f_call", "sn", "10:22", storage, false},
		{"f_called_marker", "caught assertNever reached with undefined", "3:22", storage, false},
		{"f_compare", "sn", "11:22", storage, false},
		{"f_cycle_only", "sn", "11:22", storage, false},
		{"f_default", "sn", "10:22", storage, false},
		{"f_destructured", "sn", "11:22", storage, false},
		{"f_length", "sn", "11:22", storage, false},
		{"f_name", "sn", "11:22", storage, false},
		{"f_stored", "sn", "11:22", storage, false},
		{"f_typeof", "sn", "11:22", storage, false},
		{"f_cycle", "sn", "", "", false},
		{"f_debug", "sn", "", "", false},
		{"f_function", "sn", "", "", false},
		{"f_nullish", "sn", "", "", false},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			file := pin.name + ".ts"
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_parameter", file))
			if err != nil {
				t.Fatal(err)
			}
			checkerStop := strings.HasPrefix(pin.reason, "error TS")
			var truth run
			if checkerStop {
				// The Node cache keys source inputs through the Adamic loader.
				truth = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
			} else {
				truth = onNode(t, path)
			}
			if truth.exitCode != 0 || string(truth.stdout) != pin.stdout+"\n" {
				t.Fatalf("Node exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			if checkerStop {
				// The original stack-hook fixtures have no Node import.
				// Their twins below pin the lowering refusals with Node types.
				_, err := load.Load([]string{path})
				var check *load.CheckError
				if !errors.As(err, &check) || !strings.Contains(err.Error(), file+":"+pin.location+": "+pin.reason) {
					t.Fatalf("want a checker stop at %s:%s with %q, got %v", file, pin.location, pin.reason, err)
				}
				return
			}
			program, err := lowered(t, path)
			if pin.location != "" {
				var refused *lower.Refused
				var notYet *lower.NotYet
				if err == nil || !strings.Contains(err.Error(), file+":"+pin.location+":") ||
					!strings.Contains(err.Error(), pin.reason) ||
					pin.refused && !errors.As(err, &refused) || !pin.refused && !errors.As(err, &notYet) {
					t.Fatalf("want a stop at %s:%s with %q, got %v", file, pin.location, pin.reason, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if pin.name == "n04" {
				var bodyTrap, callerTrap bool
				for _, message := range program.Strings {
					bodyTrap = bodyTrap || strings.Contains(message, "unreachable read of x at "+path+":2:28")
					callerTrap = callerTrap || strings.Contains(message, "unreachable read of v at "+path+":7:16")
				}
				if !bodyTrap || !callerTrap {
					t.Fatalf("expected a trap on the never body read and the caller argument, got %v, %v", bodyTrap, callerTrap)
				}
			}
			for _, backend := range []struct {
				name string
				got  run
			}{{"JavaScript", onJavaScriptBackend(t, program)}, {"native release", released(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Errorf("%s: %s", backend.name, difference)
				}
			}
		})
	}
}
