package main

import (
	"regexp"
	"strings"
	"testing"
)

// A project with no CohereSettings.json records its runs and replays them, and creating one afterward is
// noticed (#e0tybvv). Zero config is a missing settings file, and the run declared the file as an input it
// read, so recording stat'ed a file that was not there and refused the whole record: no zero-config project
// ever replayed. The file is an input by its absence, since creating one changes what runs.
func TestAZeroConfigProjectReplaysAndNoticesASettingsFileCreated(t *testing.T) {
	t.Parallel()
	fixture := newRunCacheFixture(t, buildCohere(t))
	fixture.remove("CohereSettings.json")
	fixture.write("source/debug.ts", "export function stop(): void {\n    debugger;\n}\n")
	fixture.commit("zero config")

	// A run over files written this second is not recorded, so the second run records and the third replays.
	fixture.run(true, "--no-fix")
	fixture.run(true, "--no-fix")
	replayed, _ := fixture.run(true, "--no-fix")
	if !isRunCacheReplay(replayed) {
		t.Fatalf("an unchanged zero-config tree did not replay:\n%s", replayed)
	}

	// The settings file turns off a rule the house stack reports here.
	fixture.write("CohereSettings.json", `{"rules":{"no-debugger":"off"}}`)
	warm, warmExit := fixture.run(true, "--no-fix")
	cold, coldExit := fixture.run(false, "--no-fix")
	if isRunCacheReplay(warm) {
		t.Fatalf("a run after CohereSettings.json was created replayed the zero-config run:\n%s", warm)
	}
	finding := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - .*$`)
	findings := func(output string) string { return strings.Join(finding.FindAllString(output, -1), "\n") }
	if findings(cold) == findings(replayed) {
		t.Fatalf("the settings file changed nothing a run reports, so the miss above proves nothing:\n--- zero config\n%s\n--- with settings\n%s", replayed, cold)
	}
	if findings(warm) != findings(cold) || warmExit != coldExit {
		t.Errorf("after CohereSettings.json was created the run reported other findings than a run with no cache:\n--- warm\n%s\n--- cold\n%s", warm, cold)
	}
}
