package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGeneratorParameterBorrowMutant(t *testing.T) {
	t.Parallel()
	generatorBorrowMutant(t, false)
}
func TestGeneratorLocalBorrowMutant(t *testing.T) {
	t.Parallel()
	generatorBorrowMutant(t, true)
}
func generatorBorrowMutant(t *testing.T, local bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generators/owned.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	changed := 0
	if !local {
		pattern := regexp.MustCompile(`(->slots\[[0-9]+\]\.reference = )adamic_retain\((adamic_local_[0-9]+_value)\);`)
		changed = len(pattern.FindAllStringIndex(code, -1))
		code = pattern.ReplaceAllString(code, `${1}${2}; /* mutant: frame borrows caller parameter */`)
	} else {
		declarations := regexp.MustCompile(`(?m)^static adamic_value adamic_function_[0-9]+_generator_resume\([^\n]+\) \{`).FindAllStringIndex(code, -1)
		for i := len(declarations) - 1; i >= 0; i-- {
			start := declarations[i][0]
			end := len(code)
			if n := strings.Index(code[declarations[i][1]:], "\nstatic "); n >= 0 {
				end = declarations[i][1] + n
			}
			body := code[start:end]
			allocation := regexp.MustCompile(`adamic_array \* (adamic_temporary_[0-9]+) = adamic_array_new\(`).FindStringSubmatch(body)
			if allocation == nil {
				continue
			}
			pattern := regexp.MustCompile(`(?m)^(\s*[^\n;]+->reference = ` + regexp.QuoteMeta(allocation[1]) + `;)$`)
			if !pattern.MatchString(body) {
				continue
			}
			changed++
			body = pattern.ReplaceAllString(body, `${1}`+"\n adamic_release("+allocation[1]+"); /* mutant: borrow across yield */")
			code = code[:start] + body + code[end:]
		}
	}
	if changed == 0 {
		t.Fatal("mutation did not reach owned frame storage")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	if got.exitCode == 0 || !strings.Contains(string(got.stderr), "AddressSanitizer: heap-use-after-free") {
		t.Fatalf("ASan did not catch frame borrow: exit=%d stderr=%s", got.exitCode, got.stderr)
	}
	t.Logf("ASan caught %d frame storage mutations", changed)
}
