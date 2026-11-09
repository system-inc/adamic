package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"delete", "foreach", "effects", "result"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/library_optional_receiver_" + name + ".a", true, false,
		})
	}
}

// A selectable behavioral pin for the argument-before-guard production mutant.
// The fixture families also hold all four witnesses to native, JavaScript and WASI.
func TestOptionalReceiverArgumentEffects(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_optional_receiver_effects.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		t.Fatal("native observation unavailable")
	}
	for name, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: %s; Node stdout=%q; backend stdout=%q stderr=%q", name, difference, truth.stdout, got.stdout, got.stderr)
		}
	}
}
