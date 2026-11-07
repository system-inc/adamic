package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{
		"02_scanner_digits",
		"03_scanner_hoisting",
		"04_scanner_escaped_text",
		"05_scanner_reassigned_text",
		"06_parser_token_state",
		"07_binder_symbol_count",
		"08_checker_symbol_recursion",
		"09_checker_constituent_recursion",
		"10_checker_arrow_cleanup",
		"11_scanner_method_wrapper",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{realNestedFixture(name), true, false})
	}
}

func realNestedFixture(name string) string {
	extension := ".a"
	if name == "09_checker_constituent_recursion" {
		extension = ".ts"
	}
	return "stage3/fixtures/nested-functions/" + name + extension
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// The newly closed recursive constituent gap must retain the source's count,
// not merely compile. This mutant produces valid C and has no leak or sanitizer failure.
func TestClosedNestedConstituentMutant(test *testing.T) {
	test.Parallel()
	path, pathErr := filepath.Abs(
		filepath.Join(repository, "stage3/fixtures/nested-functions/09_checker_constituent_recursion.ts"),
	)
	if pathErr != nil {
		test.Fatal(pathErr)
	}
	expected := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if strings.Contains(function.Name, "getConstituentCount") &&
			!strings.Contains(function.Name, "getConstituentCountOfTypes") {
			function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{}}}
			changed++
		}
	}
	if changed != 1 {
		test.Fatalf("want one constituent implementation, changed %d", changed)
	}
	actual, binary := nativelyUncached(test, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("mutant failed outside Node output: %+v", actual)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			test.Fatalf("%s mutant failed outside output: %+v", name, got)
		}
		if disagreement(expected, got) == "" {
			test.Fatalf("%s incorrect constituent count escaped Node", name)
		}
	}
	test.Log("zero constituent count caught by Node stdout in both backends; valid C, no sanitizer findings or leaks")
}
