package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p51_wrong.a", true, true})
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
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/lower/testdata/fx6/p49.a", true, true})
}

func TestFX6P35(t *testing.T) {
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

func TestFX6P36(t *testing.T) {
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

func TestFX6P37(t *testing.T) {
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

func TestFX6P38(t *testing.T) {
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

func TestFX6P51(t *testing.T) {
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

func TestFX6P53(t *testing.T) {
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

// A wrong view is required to stop before reading its payload, even though
// the unchecked source on Node continues. The diagnostic pins the view boundary.
func fx6CheckedStop(t *testing.T, name, message string) {
	t.Helper()
	path, err := filepath.Abs("../../internal/lower/testdata/fx6/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", source.exitCode, source.stdout, source.stderr)
	if source.exitCode != 0 {
		t.Fatal("Node control failed")
	}
	expected := run{exitCode: 70, stderr: []byte(message + "\n")}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(expected, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestFX6P49(t *testing.T) {
	t.Parallel()
	fx6CheckedStop(t, "p49", "adamic: panic: field read failed: value (field value) matches no member of Target; expected Target, found object")
}

func TestFX6P51Wrong(t *testing.T) {
	t.Parallel()
	fx6CheckedStop(t, "p51_wrong", "adamic: panic: field read failed: value (field value) expected Target, found function with incompatible parameter representations")
}
