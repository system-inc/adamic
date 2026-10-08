package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Nightly or per-main-push shipping oracle. Ordinary oracle/test flags are untouched.
func TestReleaseAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_RELEASE") != "1" {
		t.Skip("set ADAMIC_ORACLE_RELEASE=1 for the shipped release oracle")
	}
	t.Parallel()
	for _, fixture := range fixtures {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want explicit refusal, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if fixture.checked {
				node := want
				want = onJavaScriptBackend(t, program)
				if want.exitCode != 70 || node.exitCode == 70 {
					t.Fatal("inserted-check fixture did not distinguish Node")
				}
			}
			binary := filepath.Join(t.TempDir(), "shipped")
			if err := native.Build(native.C(program), binary, native.Options{Release: true}); err != nil {
				t.Fatal(err)
			}
			got := execute(t, binary)
			if difference := disagreement(want, got); difference != "" {
				t.Errorf("%s\nNode/backend: exit %d stdout %q stderr %q\nshipped: exit %d stdout %q stderr %q", difference, want.exitCode, want.stdout, want.stderr, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}

func TestReleaseOracleCatchesOneByte(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_RELEASE") != "1" {
		t.Skip("set ADAMIC_ORACLE_RELEASE=1")
	}
	t.Parallel()
	path, _ := filepath.Abs(filepath.Join(repository, "dedication/dedication.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	program.Strings[0] += "!"
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Release: true}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(onNode(t, path), execute(t, binary)); difference != "stdout differs" {
		t.Fatalf("shipping oracle mutant escaped: %q", difference)
	}
}
