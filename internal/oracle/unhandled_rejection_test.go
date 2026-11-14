package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var unhandledRejectionFixtures = []struct {
	name, stdout, message string
}{
	{"handled_later", "handled\n", ""},
}

func unhandledRejectionPath(name string) string {
	return "internal/oracle/testdata/unhandled_rejection/" + name + ".a"
}

func init() {
	// Keep the handled-later agreement and its ordinary count record. The old
	// void views are now exact refusal pins; the runtime backstop needs its own harness.
	for _, fixture := range unhandledRejectionFixtures {
		path := unhandledRejectionPath(fixture.name)
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) {
			return counted(t, path, false, nil, false, false)
		}})
	}
}

// Omitting await's handled bit reds
// TestUnhandledRejectionAgreesWithNode/handled_later/sanitized.
func TestUnhandledRejectionAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range unhandledRejectionFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, unhandledRejectionPath(fixture.name)))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(fixture.stdout)}
			if fixture.message != "" {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: Error: " + fixture.message + "\n")
			}
			truth := onNode(t, path)
			if difference := disagreement(want, truth); difference != "" {
				t.Fatalf("source Node: %s: exit %d stdout %q stderr %q", difference, truth.exitCode, truth.stdout, truth.stderr)
			}
			for _, backend := range []string{"sanitized", "release", "javascript"} {
				t.Run(backend, func(t *testing.T) {
					var got run
					switch backend {
					case "sanitized":
						var binary string
						got, binary = natively(t, program)
						if want.exitCode == 0 {
							if report := leaks(t, program, binary); report != "" {
								t.Fatal(report)
							}
						}
					case "release":
						got = released(t, program)
					case "javascript":
						got = onJavaScriptBackend(t, program)
					}
					if difference := disagreement(truth, got); difference != "" {
						t.Fatalf("%s: exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
					}
				})
			}
		})
	}
}

var unhandledRejectionRefusals = []struct {
	name, stdout, message, location string
}{
	{"void_view", "", "rejected", "unhandled_void_view.a:6:25"},
	{"chained", "", "chained", "unhandled_chained.a:9:25"},
	{"timer", "timer\n", "timer rejection", "unhandled_timer.a:7:25"},
	{"two", "", "first", "unhandled_two.a:6:40"},
}

func unhandledRejectionRefusedPath(name string) string {
	return "internal/oracle/testdata/async_review/refused/unhandled_" + name + ".a"
}

func init() {
	for _, fixture := range unhandledRejectionRefusals {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{unhandledRejectionRefusedPath(fixture.name), false, false})
	}
}

// Preserve the original source programs and Node's outputs after :355 refuses
// their promise-returning function seen through a void view (#8j8ecw0).
func TestUnhandledRejectionRefused(t *testing.T) {
	t.Parallel()
	for _, fixture := range unhandledRejectionRefusals {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, unhandledRejectionRefusedPath(fixture.name)))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			want := run{stdout: []byte(fixture.stdout), stderr: []byte("adamic: panic: Error: " + fixture.message + "\n"), exitCode: 70}
			if difference := disagreement(want, truth); difference != "" {
				t.Fatalf("Node witness: %s: %+v", difference, truth)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			const what = "an unawaited async task: a Promise<void> at the initializer site through () seen as void, a type that names no Promise, so whoever holds it drops the task"
			const fix = "await the task inside a function that returns void, or keep the Promise in the type it's seen as"
			if !errors.As(err, &refused) || refused.What != what || refused.Fix != fix || !strings.HasSuffix(refused.Where, fixture.location) {
				t.Fatalf("want exact :355 refusal at %s: %v", fixture.location, err)
			}
			t.Log(err)
		})
	}
}
