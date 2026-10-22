package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p38_number.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p35.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p36.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p37.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p38.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p51.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p53.a", true, false})

}

func TestOracleViewsViewedProducerMutation(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p35.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsUnrelatedMutableField(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p36.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsUnrelatedUnionField(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p37.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsUnrelatedOptionalStringField(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p38.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsDestructuredScalarUnion(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p51.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsDestructuredCallableUnion(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p53.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestOracleViewsDestructuredInvalidObjectUnionRefuses(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../lower/testdata/fx6_refused/p49.a")
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "true\n" {
		t.Fatalf("Node control: %+v", source)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "a destructured object-union view with an incompatible literal producer" || refused.Fix != "build a member of Target before destructuring field value" || !strings.Contains(refused.Where, "p49.a:") {
		t.Fatalf("want located incompatible-producer refusal with fix, got %v", err)
	}
}

// The stored function takes a string and no member of Target does, so the destructured read is refused where it
// stands, as p49's incompatible literal producer is, rather than compiled to stop at run time where Node prints 5
// (Compiler's ruling (b) on #aecx10a).
func TestOracleViewsDestructuredInvalidCallableRefuses(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/p51_wrong.a")
	if err != nil {
		t.Fatal(err)
	}
	if source := onNode(t, path); source.exitCode != 0 {
		t.Fatalf("Node control failed: %#v", source)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "checked view read of field value with unsupported representation conversion contract" || refused.Fix != "prove or implement the representation conversion contract before reading this field" || !strings.Contains(refused.Where, "p51_wrong.a:7:9") {
		t.Fatalf("want the located representation conversion refusal at 7:9, got %v", err)
	}
}
