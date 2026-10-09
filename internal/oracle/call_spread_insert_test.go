package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/call_spread_insert.a", true, false})
}

func TestCallSpreadInsert(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_spread_insert.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	got, binary := natively(t, p)
	for name, value := range map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, p)} {
		if diff := disagreement(truth, value); diff != "" {
			t.Fatalf("%s %s: %#v vs %#v", name, diff, value, truth)
		}
	}
	if report := leaks(t, p, binary); report != "" {
		t.Fatal(report)
	}
}

// The mutant repeats the spread-producing call and releases its discarded result.
func TestCallSpreadEvaluatedTwiceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_spread_insert.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(p)
	// Keep ownership balanced so only the Node comparison can catch repeated evaluation.
	mutated := strings.ReplaceAll(code, "= adamic_function_0_source();", "= (adamic_release(adamic_function_0_source()), adamic_function_0_source());")
	if mutated == code {
		t.Fatal("mutation did not apply")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutated, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %#v", got)
	}
	if diff := disagreement(onNode(t, path), got); diff != "stdout differs" {
		t.Fatalf("mutant caught by %q", diff)
	}
}
