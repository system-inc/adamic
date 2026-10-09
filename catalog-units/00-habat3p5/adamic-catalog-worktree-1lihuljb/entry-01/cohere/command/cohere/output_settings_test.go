package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestTheOutputSettingIsHonored runs the default view, with no flag, over a project whose settings set
// "output": { "phases": true }: the footer leads its parentheses with the phase timings. The same project
// without the setting does not, which is what makes the first half mean something. A key the block does
// not know, and the block in a file another extends, are refused by name.
func TestTheOutputSettingIsHonored(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	run := func(settings map[string]string) (string, int) {
		t.Helper()
		root := t.TempDir()
		files := map[string]string{"tsconfig.json": fixScopeTsconfig, "Clean.ts": "export const clean = 1;\n"}
		for name, contents := range settings {
			files[name] = contents
		}
		writeTree(t, root, files)
		// The default view, so not runCohere, which asks for --verbose.
		command := exec.Command(binary, "--no-fix", "--no-format", "--no-cache")
		command.Dir = root
		output, err := command.CombinedOutput()
		if exitError, isExit := err.(*exec.ExitError); isExit {
			return string(output), childExitCode(t, exitError)
		}
		return string(output), 0
	}
	footerOf := func(output string) string {
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, "✓ 💎 ") || strings.HasPrefix(line, "✗ ☠️ ") {
				return line
			}
		}
		return ""
	}

	withPhases, code := run(map[string]string{"CohereSettings.json": `{"output":{"phases":true}}`})
	if footer := footerOf(withPhases); code != 0 || !strings.Contains(footer, "(🕸 ") {
		t.Errorf("phases: true did not put the timings in the footer (exit %d):\n%s", code, withPhases)
	}
	without, code := run(map[string]string{"CohereSettings.json": `{}`})
	if footer := footerOf(without); code != 0 || footer == "" || strings.Contains(footer, "🕸") {
		t.Errorf("a project with no output setting printed the timings, or no footer (exit %d):\n%s", code, without)
	}

	unknown, code := run(map[string]string{"CohereSettings.json": `{"output":{"phase":true}}`})
	if code == 0 || !strings.Contains(unknown, `"output" block`) {
		t.Errorf("an output key the block does not know was not refused (exit %d):\n%s", code, unknown)
	}
	inherited, code := run(map[string]string{
		"CohereSettings.json": `{"extends":"./Base.json"}`,
		"Base.json":           `{"output":{"phases":true}}`,
	})
	if code == 0 || !strings.Contains(inherited, `declares "output"`) {
		t.Errorf("an output block in a file another extends was not refused (exit %d):\n%s", code, inherited)
	}
}
