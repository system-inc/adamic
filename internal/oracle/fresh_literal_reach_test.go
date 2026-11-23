package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A walk that asks what a value seen as a type may really be relates each object literal by its regular widened
// type (lower's relatedShape), so a fresh literal with a field the type lacks is still found (#sbvnh96,
// #yfmm1vv). On main, a literal cast to a narrower interface and stored into what its extra field holds leaked
// natively, since the cycle finder never followed that field. These agree with Node on every backend,
// sanitized, and leak nothing. The variable form, whose regular type the finder always had, is their control.
// Mutant (scratch-cy): relatedShape returning the fresh type again turns the three cast rows red under leak
// checking and leaves the variable row green.
func TestFreshLiteralCyclesAreCollected(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, node string }{
		{"cycle_cast_push.ts", "1 c\n"},
		{"cycle_cast_map.ts", "1 m\n"},
		{"cycle_cast_field.ts", "f\n"},
		{"cycle_variable_push.ts", "1 v\n"},
		// A cast literal with a plain extra field reaches no toJSON, so it still prints both fields.
		{"json_cast_extra_field.ts", "{\"b\":2,\"a\":1}\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/fresh_literal_reach/" + probe.file
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != probe.node {
				t.Fatalf("source Node: exit %d, stdout %q, want %q", truth.exitCode, truth.stdout, probe.node)
			}
			generatorAgrees(t, relative)
		})
	}
}

// JSON.stringify through a view that can reach a literal with its own toJSON stops at compile time, whether the
// literal is cast to the view or kept in a variable first (#yfmm1vv). On main the cast form reached run time,
// where the JavaScript backend printed nothing and native stopped. Node's own answer is pinned beside each.
// Mutant: the fresh type related again turns the cast row red, since it then lowers.
func TestFreshLiteralToJSONReachStops(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"json_cast_to_json.ts", "json_variable_to_json.ts"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/fresh_literal_reach", file))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "\"custom\"\n" {
				t.Fatalf("source Node: exit %d, stdout %q, want %q", truth.exitCode, truth.stdout, "\"custom\"\n")
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), file) || !strings.Contains(err.Error(), "which has a toJSON member") {
				t.Fatalf("want NotYet naming the toJSON member, got %v", err)
			}
		})
	}
}
