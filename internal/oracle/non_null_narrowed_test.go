package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"number_minimal", "number", "boolean", "string", "reference", "maybe_number"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_narrowed_" + name + ".a", true, false})
	}
}

// Generated locals, globals and temporaries have program-wide unique names, including
// across region variants. Check their declared C types before clang sees the artifact.
var numericCDeclaration = regexp.MustCompile(`\b(?:double|bool|int64_t)\s+(adamic_(?:local|global|temporary)_[A-Za-z0-9_]+)\b`)
var rightNullComparison = regexp.MustCompile(`\b(adamic_(?:local|global|temporary)_[A-Za-z0-9_]+)\b\s*\)*\s*(?:==|!=)\s*NULL\b`)
var leftNullComparison = regexp.MustCompile(`\bNULL\s*(?:==|!=)\s*\(*\s*(adamic_(?:local|global|temporary)_[A-Za-z0-9_]+)\b`)
var numericMemberNullComparison = regexp.MustCompile(`(?:\.number\s*\)*\s*(?:==|!=)\s*NULL\b|\bNULL\s*(?:==|!=)\s*[^;\n]*\.number\b)`)
var cQuotedText = regexp.MustCompile(`(?s)"(?:\\.|[^"\\])*"|/\*.*?\*/|//[^\n]*`)

func numericNullComparisons(source string) []string {
	source = cQuotedText.ReplaceAllString(source, "")
	numeric := map[string]bool{}
	for _, declaration := range numericCDeclaration.FindAllStringSubmatch(source, -1) {
		numeric[declaration[1]] = true
	}
	var failures []string
	for _, pattern := range []*regexp.Regexp{rightNullComparison, leftNullComparison} {
		for _, comparison := range pattern.FindAllStringSubmatch(source, -1) {
			if numeric[comparison[1]] {
				failures = append(failures, comparison[0])
			}
		}
	}
	failures = append(failures, numericMemberNullComparison.FindAllString(source, -1)...)
	return failures
}

func TestOracleCDoesNotCompareNumbersWithNull(t *testing.T) {
	t.Parallel()
	paths := []string{}
	seen := map[string]bool{}
	for _, fixture := range fixtures {
		if fixture.lowers && !seen[fixture.path] {
			paths = append(paths, fixture.path)
			seen[fixture.path] = true
		}
	}
	for _, fixture := range inputFixtures {
		if !seen[fixture.path] {
			paths = append(paths, fixture.path)
			seen[fixture.path] = true
		}
	}
	for _, fixturePath := range paths {
		t.Run(fixturePath, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixturePath))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if failures := numericNullComparisons(native.C(program)); len(failures) != 0 {
				t.Fatalf("numeric C values compared with NULL: %v", failures)
			}
		})
	}
}

// Restore the former lowering around the narrowing helper's plain-number result.
// The artifact assertion catches the pointer test before any clang invocation.
func TestNonNullNarrowedPointerTestMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_narrowed_number_minimal.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	if failures := numericNullComparisons(source); len(failures) != 0 {
		t.Fatal(failures)
	}
	node := onNode(t, path)
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(node, baseline); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(node, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	message := len(program.Strings)
	program.Strings = append(program.Strings, "non-null assertion failed: value! is null or undefined")
	changes := 0
	for i := range program.Functions {
		if program.Functions[i].Name != "narrowed" {
			continue
		}
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
			if returned, ok := node.(ir.Return); ok {
				if _, narrowed := returned.Value.(ir.Call); narrowed && returned.Value.Type() == ir.Number {
					returned.Value = ir.Coalesce{Value: returned.Value, Panic: ir.StringConstant{Index: message}, Of: ir.Number}
					changes++
					return returned
				}
			}
			return node
		})
	}
	if changes != 1 {
		t.Fatalf("pointer-test mutant changed %d returns", changes)
	}
	failures := numericNullComparisons(native.C(program))
	if len(failures) == 0 {
		t.Fatal("restored pointer test escaped the C artifact assertion")
	}
	t.Logf("restored pointer test caught before clang: %s", strings.Join(failures, ", "))
}

func TestNumericNullComparisonAssertion(t *testing.T) {
	t.Parallel()
	for _, comparison := range []string{"adamic_temporary_1 != NULL", "(adamic_temporary_1) == NULL", "NULL != adamic_temporary_1", "adamic_temporary_2.number != NULL"} {
		source := fmt.Sprintf("double adamic_temporary_1 = 0; adamic_value adamic_temporary_2 = {0}; if (%s) {}", comparison)
		if len(numericNullComparisons(source)) != 1 {
			t.Fatalf("missed %s", comparison)
		}
	}
	source := `double adamic_temporary_1 = 0; adamic_heap *adamic_temporary_2 = NULL; if (adamic_temporary_2 != NULL) {} const char *text = "adamic_temporary_1 != NULL";`
	if failures := numericNullComparisons(source); len(failures) != 0 {
		t.Fatal(failures)
	}
}
