package oracle

import "fmt"

// Only these exact fixtures have ruled differences. Each side remains pinned:
// membership never excuses an arbitrary exit, diagnostic or output change.
var ruledBackendDivergences = []struct {
	fixture    string
	ruling     string
	reason     string
	native     run
	javascript run
}{
	{
		fixture:    "internal/oracle/testdata/4ddd17f_weak_single_narrowed.a",
		ruling:     "@system_adamic, 2026-10-08 18:23, task #z1vjxxd",
		reason:     "Native weak reads stop after the last strong owner frees the target; V8 keeps the JavaScript target alive. JavaScript must not emulate reference counting.",
		native:     run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: a weak reference was read after what it pointed to was freed\n"), exitCode: 70},
		javascript: run{stdout: []byte("before\nafter: nn\n"), exitCode: 0},
	},
	{
		fixture:    "internal/oracle/testdata/d96d304_try_stack.a",
		ruling:     "@system_adamic, task #z1vjxxd, chain step 24",
		reason:     "Native stack guards name the exhausted function; JavaScript preserves V8's unnamed overflow diagnostic. Both stops are terminal.",
		native:     run{stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded in depth\n"), exitCode: 70},
		javascript: run{stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded\n"), exitCode: 70},
	},
}

// backendDisagreement applies a named ruling or the ordinary byte-exact comparison.
// Every unlisted inserted check must agree on both sides.
func backendDisagreement(fixture string, javascript, native run) string {
	for _, entry := range ruledBackendDivergences {
		if entry.fixture != fixture {
			continue
		}
		if difference := disagreement(entry.native, native); difference != "" {
			return fmt.Sprintf("ruled native outcome: %s (%s)", difference, entry.ruling)
		}
		if difference := disagreement(entry.javascript, javascript); difference != "" {
			return fmt.Sprintf("ruled JavaScript outcome: %s (%s)", difference, entry.ruling)
		}
		return ""
	}
	return disagreement(javascript, native)
}

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/4ddd17f_weak_single_narrowed.a",
		"internal/oracle/testdata/d96d304_try_stack.a",
		"internal/oracle/testdata/backend_stop_catches.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, path != "internal/oracle/testdata/backend_stop_catches.a"})
	}
}
