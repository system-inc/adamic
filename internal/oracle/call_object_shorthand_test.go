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
	}{"internal/oracle/testdata/call_object_shorthand.a", true, false})
}

func TestCallObjectShorthand(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_object_shorthand.a"))
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

func TestCallShorthandWrongBindingMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/call_object_shorthand.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(p)
	mutated := strings.ReplaceAll(code, "= adamic_global_0_status;", "= adamic_global_1_other;")
	if mutated == code {
		t.Fatal("wrong binding mutation did not apply")
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
