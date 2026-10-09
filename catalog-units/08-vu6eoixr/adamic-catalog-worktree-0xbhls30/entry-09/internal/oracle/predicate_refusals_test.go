package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var predicateMiscompileCases = []struct {
	name, output, diagnostic string
	pending                  bool
}{
	{"p05_helper_second_parameter", "number NaN\nnumber 42\n", "type predicate whose return is not proven", false},
	{"p06_helper_second_parameter_undefined", "undefined? abc\nno\nundefined? undefined\nyes\n", "type predicate whose return is not proven", false},
	{"p12_find_undefined_target", "found undefined\n", "find predicate result representation", true},
	{"p13_filter_wrong_parameter_helper", "3\nitem a\nitem undefined\nitem bc\n", "unproven predicate argument", false},
	{"p14_overload_optional_chain_container", "claimed box: box a\ndirect: box\nclaimed box: undefined\ndirect: undefined\n", "", false},
	{"p21_find_inferred_number", "found 2\na|bc\n", "find predicate result representation", true},
	{"p22_filter_inferred_narrowing", "a|bc\n1,2\ntotal 6\n", "filter predicate element representation", true},
	{"p23_find_declared_predicate", "found 2\n", "find predicate result representation", true},
	{"p24_filter_proven_number", "2\n2\n2,4\n", "filter predicate element representation", true},
	{"p26_assert_nested_if_without_else", "value abc\nvalue undefined\n", "type predicate whose return is not proven", false},
	{"p27_overload_parameter_rebound", "text a\ntext 41\n", "", false},
}

func TestPredicateMiscompileRefusals(t *testing.T) {
	t.Parallel()
	baseline := os.Getenv("PREDICATE_MISCOMPILE_BASELINE") != ""
	for _, probe := range predicateMiscompileCases {
		t.Run(probe.name, func(t *testing.T) {
			relative := "internal/oracle/testdata/predicate_refusals/oct8_predicates_" + probe.name + ".a"
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.output || len(node.stderr) != 0 {
				t.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
			}
			t.Logf("Node: exit=%d stdout=%q", node.exitCode, node.stdout)
			program, err := lowered(t, path)
			if baseline {
				t.Logf("lower: %v", err)
			} else if probe.diagnostic != "" {
				var refused *lower.Refused
				var pending *lower.NotYet
				if (probe.pending && !errors.As(err, &pending)) || (!probe.pending && !errors.As(err, &refused)) || err == nil || !strings.Contains(err.Error(), probe.diagnostic) || !strings.Contains(err.Error(), filepath.Base(path)+":") {
					t.Fatalf("want path-bearing %q refusal/pending, got %v", probe.diagnostic, err)
				}
				t.Log(err)
				return
			}
			if err != nil {
				if baseline {
					return
				}
				t.Fatal(err)
			}
			js := onJavaScriptBackend(t, program)
			t.Logf("javascript: exit=%d stdout=%q stderr=%q", js.exitCode, js.stdout, js.stderr)
			release := released(t, program)
			t.Logf("release: exit=%d stdout=%q stderr=%q", release.exitCode, release.stdout, release.stderr)
			sanitized, _ := natively(t, program)
			t.Logf("sanitized: exit=%d stdout=%q stderr=%q", sanitized.exitCode, sanitized.stdout, sanitized.stderr)
			if baseline {
				return
			}
			prefix := "text a\n"
			message := "overload 1 of isText result: predicate x is false"
			if probe.name == "p14_overload_optional_chain_container" {
				prefix = "claimed box: box a\ndirect: box\n"
			}
			for name, got := range map[string]run{"javascript": js, "release": release, "sanitized": sanitized} {
				if got.exitCode != 70 || string(got.stdout) != prefix || string(got.stderr) != "adamic: panic: "+message+"\n" {
					t.Errorf("%s: want named predicate check, exit=%d stdout=%q stderr=%q", name, got.exitCode, got.stdout, got.stderr)
				}
			}
			checked, unobservable := 2, 0
			if probe.name == "p14_overload_optional_chain_container" {
				checked, unobservable = 1, 1
			}
			if program.PredicateChecks.Checked != checked || program.PredicateChecks.Unobservable != unobservable {
				t.Errorf("direction report: %+v", program.PredicateChecks)
			}
		})
	}
}

func init() {
	for _, name := range []string{"p14_overload_optional_chain_container", "p27_overload_parameter_rebound"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/predicate_refusals/oct8_predicates_" + name + ".a", true, true})
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/predicate_refusals/representation_controls.a", true, false})
}
