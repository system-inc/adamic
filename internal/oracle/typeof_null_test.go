package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"e4eec87_v01_typeof_null.a", "e4eec87_v02_typeof_null_literal.a", "e4eec87_w01_uncovered_paths.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Mutate emitted C, so each mutant must still build and run under the ordinary sanitizer flags.
func TestTypeOfNullAndUncoveredMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name        string
		pattern     string
		replacement string
	}{
		{"null becomes undefined", `&adamic_typeof_object`, `&adamic_typeof_undefined`},
		{"boolean undefined narrowing loses absence", `(== NULL \? )\(adamic_maybe_boolean\)\{false, false\}`, `${1}(adamic_maybe_boolean){true, false}`},
		{"boolean undefined boxing loses absence", `(\.present \? )NULL( : \([^\n]*?\)\.boolean \? &adamic_box_true\.heap : &adamic_box_false\.heap)`, `${1}&adamic_box_false.heap${2}`},
		{"pair typeof loses present type", `(\.present \? )&adamic_typeof_(boolean|number)( : &adamic_typeof_undefined)`, `${1}&adamic_typeof_undefined${3}`},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			pattern := regexp.MustCompile(mutant.pattern)
			caught := 0
			for _, fixture := range fixtures {
				if !fixture.lowers || fixture.checked {
					continue
				}
				if mutant.name == "null becomes undefined" && !strings.HasSuffix(fixture.path, "e4eec87_v02_typeof_null_literal.a") {
					continue
				}
				path, err := filepath.Abs(filepath.Join(repository, fixture.path))
				if err != nil {
					t.Fatal(err)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				code := native.C(program)
				changed := pattern.ReplaceAllString(code, mutant.replacement)
				if mutant.name == "boolean undefined narrowing loses absence" {
					// Array element reads also unbox booleans. Restrict this mutant to ir.Union operands.
					narrow := regexp.MustCompile(`(\w+)( == NULL \? )\(adamic_maybe_boolean\)\{false, false\}`)
					changed = narrow.ReplaceAllStringFunc(code, func(match string) string {
						parts := narrow.FindStringSubmatch(match)
						declared := regexp.MustCompile(`adamic_heap \*\s*` + regexp.QuoteMeta(parts[1]) + `\b`)
						if !declared.MatchString(code) {
							return match
						}
						return parts[1] + parts[2] + "(adamic_maybe_boolean){true, false}"
					})
				}
				if changed == code {
					continue
				}
				binary := filepath.Join(t.TempDir(), "mutant")
				if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
				if got.exitCode != 0 || len(got.stderr) != 0 {
					t.Fatalf("%s: mutant must run cleanly: exit %d, stderr %s", fixture.path, got.exitCode, got.stderr)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatalf("%s: mutant leaked: %s", fixture.path, report)
				}
				if difference := disagreement(onNode(t, path), got); difference == "stdout differs" {
					caught++
					t.Logf("%s caught by Node stdout comparison: %q", fixture.path, got.stdout)
				} else if difference != "" {
					t.Fatalf("unexpected comparison: %s", difference)
				}
			}
			if caught != 1 {
				t.Fatalf("want exactly one fixture to catch mutant, got %d", caught)
			}
		})
	}
}
