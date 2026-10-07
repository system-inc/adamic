package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// TestNativeReleaseFlagsAgreeWithNode is a selectable shipping-build lane. Build and its runtime
// library use the options traced from adamic build and its actual flag function. Any
// LTO flags added there apply here without maintaining another flag list. Native observations are
// always fresh; the existing Node cache has the gate's uncached bypass.
func TestNativeReleaseFlagsAgreeWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_RELEASE_LANE") != "1" {
		t.Skip("set ADAMIC_RELEASE_LANE=1 for the full shipping-build lane")
	}
	t.Parallel()
	releaseFixturesAgreeWithNode(t, false)
}

// This optional control measures the existing release variant on the same finishing fixtures.
// ADAMIC_GATE_UNCACHED=1 makes both sides build and execute every binary anew.
func TestNativeExistingReleaseAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_RELEASE_MEASURE") != "1" {
		t.Skip("set ADAMIC_RELEASE_MEASURE=1 for the timing control")
	}
	t.Parallel()
	releaseFixturesAgreeWithNode(t, true)
}

func releaseFixturesAgreeWithNode(t *testing.T, existing bool) {
	t.Helper()
	options := traceReleaseOptions(t, releaseFunction(t, "cmd/adamic", "build"))
	if existing {
		options = traceReleaseOptions(t, releaseFunction(t, "internal/oracle", "releasedUncached"))
	}
	t.Logf("build-flags: %s", strings.Join(native.Flags(options), " "))
	for _, fixture := range fixtures {
		if !fixture.lowers {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			if fixture.checked {
				// The source lacks Adamic's inserted check. Its reference is the backend on Node,
				// as in the ordinary oracle, rather than the source that runs past that check.
				expected = onJavaScriptBackend(t, program)
			}
			if expected.exitCode != 0 {
				t.Fatalf("finishing fixtures only: Node exit %d", expected.exitCode)
			}
			var actual run
			if existing {
				actual = released(t, program)
			} else {
				// Use the options traced from the command, even if the oracle later deliberately
				// selects different options. Build supplies the runtime archive and link flags.
				binary := filepath.Join(t.TempDir(), "release")
				if err := native.Build(native.C(program), binary, options); err != nil {
					t.Fatal(err)
				}
				actual = execute(t, binary)
			}
			if difference := disagreement(expected, actual); difference != "" {
				t.Errorf("shipped release: %s\nNode: exit %d, stdout %q, stderr %q\nrelease: exit %d, stdout %q, stderr %q",
					difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
		})
	}
}
