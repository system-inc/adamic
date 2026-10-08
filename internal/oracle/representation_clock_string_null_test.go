package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const clockStringNullFixture = "internal/oracle/testdata/representation_clock_string_null.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{clockStringNullFixture, true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/representation_clock_string_null_observations.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/representation_clock_string_null_stale.a", true, false})
}

func TestClockStringNullTypeofMutant(t *testing.T) {
	t.Run("clock-string-null-typeof-undefined", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(repository, clockStringNullFixture))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		want := onNode(t, path)
		if string(want.stdout) != "string:false:text\nstring:false:\nobject:true:fallback\n" {
			t.Fatalf("unexpected Node observation: %+v", want)
		}
		if restoreNullTypeOf(reflect.ValueOf(program)) == 0 {
			t.Fatal("mutant changed no typeof observations")
		}
		got, binary := natively(t, program)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("mutant must finish cleanly: %+v", got)
		}
		if disagreement(want, got) != "stdout differs" || !strings.Contains(string(got.stdout), "undefined:true:fallback\n") {
			t.Fatalf("stdout must kill mutant: %+v", got)
		}
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
		t.Logf("Node %q; mutant %q", want.stdout, got.stdout)
	})
}
