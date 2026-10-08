package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The replay keeps a typed case's captured compiler options (#0jkpds7). The upstream corpus can't show it on
// its own: both engines read the one tsconfig the replay writes, and none of today's typed cases changes its
// findings between upstream's options and strict alone. What typeaware-17 found was a JavaScript case. Under
// strict alone the program doesn't take a .js file, so Go's oracle refuses the case (its tsconfig names one
// file and matches none) and the comparison is never reached. Under upstream's allowJs it runs. This holds
// typedReplayConfig, the one place TestRulesAgree builds a typed case's tsconfig, to both answers.
func TestTypedReplayKeepsUpstreamCompilerOptions(t *testing.T) {
	t.Parallel()
	oracle := goOracle(t)
	source := filepath.Join(t.TempDir(), "script.js")
	if err := os.WriteFile(source, []byte("const flag = true;\nif (flag === true) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	row := source + "\t@typescript-eslint/no-unnecessary-boolean-literal-compare"
	replay := func(program typedProgram, found bool) error {
		config := filepath.Join(t.TempDir(), "tsconfig.json")
		if err := os.WriteFile(config, []byte(typedReplayConfig(program, found, source)), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := run("", nil, oracle, "--manifest", manifest(t, []string{"program " + config, row}), "--count")
		return err
	}
	captured := typedProgram{CompilerOptions: `{"strict":true,"allowJs":true,"checkJs":true,"types":[]}`}
	if !strings.Contains(typedReplayConfig(captured, true, source), `"allowJs":true`) {
		t.Fatalf("the replay dropped the captured options: %s", typedReplayConfig(captured, true, source))
	}
	if err := replay(captured, true); err != nil {
		t.Fatalf("Go refused a JavaScript case under its captured allowJs: %v", err)
	}
	if err := replay(captured, false); err == nil {
		t.Fatal("Go ran a JavaScript case under strict alone, so this test can't see a dropped allowJs")
	}
}
