package css

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// A shared setup that fails inside its Once is named by every later caller. Each planted failure
// (cssPlantedSetupFailure) runs two tests that share the setup in a fresh process: the first to reach the Once fails
// with the planted error, and the other must fail naming that test, never with a nameless product lookup ("css
// sanitized", "no traced build has recorded what it reads") or "composition setup did not complete" (#qbbmqfs).
func TestCSSSetupFailuresNameTheFirst(t *testing.T) {
	t.Parallel()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, planted := range []struct {
		setup, run, earlier string
		tests               []string
	}{
		{"program", "^(TestProduct_CSSParserParserLowered|TestProduct_CSSParserParserNative)$", "the prepared css parser program -1 failed earlier in this process, in ", []string{"TestProduct_CSSParserParserLowered", "TestProduct_CSSParserParserNative"}},
		{"composition", "^(TestCompositionMatchesGo_000|TestCompositionMatchesGo_001)$", "the composition setup failed earlier in this process, in ", []string{"TestCompositionMatchesGo_000", "TestCompositionMatchesGo_001"}},
	} {
		result := cssParserExecute(t, []string{"ADAMIC_CSS_PLANTED_SETUP_FAILURE=" + planted.setup}, binary, "-test.run="+planted.run, "-test.v", "-test.count=1", "-test.timeout=120s")
		output := string(result.stdout)
		if result.exitCode != 1 || strings.Count(output, "planted: the "+planted.setup+" setup fails") != 1 {
			t.Fatalf("%s: want the planted failure exactly once, exit %d\n%s\n%s", planted.setup, result.exitCode, output, result.stderr)
		}
		for _, nameless := range []string{"css  sanitized", "no traced build has recorded what it reads", "composition setup did not complete"} {
			if strings.Contains(output, nameless) {
				t.Fatalf("%s: a later caller failed with %q\n%s", planted.setup, nameless, output)
			}
		}
		named := regexp.MustCompile(regexp.QuoteMeta(planted.earlier) + `(\w+)`).FindStringSubmatch(output)
		if named == nil {
			t.Fatalf("%s: no later caller named the earlier failure\n%s", planted.setup, output)
		}
		for _, test := range planted.tests {
			if !strings.Contains(output, "--- FAIL: "+test+" (") {
				t.Fatalf("%s: %s didn't fail\n%s", planted.setup, test, output)
			}
		}
		first := named[1]
		if first != planted.tests[0] && first != planted.tests[1] || !strings.Contains(output, "--- FAIL: "+first+" (") {
			t.Fatalf("%s: the later caller named %s, which isn't the test that failed first\n%s", planted.setup, first, output)
		}
		t.Logf("%s: the planted failure in %s, and the later caller named it: %s%s", planted.setup, first, planted.earlier, first)
	}
}
