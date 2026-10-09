package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/housesets"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"github.com/system-inc/cohere/internal/types/program"
)

// Zero config is the house stack (Kirk, 2026-10-04, #bfxz13m): a project with no CohereSettings.json,
// or one whose chain names no `cohere:` set, gets every set we pick, each where the program shows it
// fits, and the run's first line names each set and its evidence so nothing is applied silently.

// loadLintConfig loads the configuration a run lints with: the project's own chain when it names its
// sets, or the house stack detected from graph when it does not.
//
// A `--lint-config` that names a missing file is still an error, as it always was: the person asked for
// that file. Only the default file's absence means zero config.
func loadLintConfig(graph *program.Graph, location projectLocation) (*configuration.Config, error) {
	if !usesHouseSets(location) {
		return configuration.LoadFor(location.LintConfigFileName, registeredRuleNames())
	}
	settings, err := configuration.OwnSettingsOf(location.LintConfigFileName)
	if err != nil {
		return nil, err
	}
	settingsDirectory, err := filepath.Abs(filepath.Dir(location.LintConfigFileName))
	if err != nil {
		return nil, err
	}
	detection := housesets.Detect(graph.ProjectFiles(), location.Root, graph.Program.Host().FS(), settings, settingsDirectory)
	// Settings for rules the house does not apply would be read by nothing, which is the silent drop
	// reading them exists to end (#gj5nm6e). The run says so instead, with the way out.
	if detection.TailwindEntryPoint == "" && tailwind.WritesSettings(settings) {
		return nil, fmt.Errorf("%s writes settings[\"better-tailwindcss\"], and zero config does not apply %s: %s. "+
			"Nothing would read those settings, so name its stylesheet in them as entryPoint, name %q in "+
			"\"extends\", or remove them", location.LintConfigFileName, configuration.TailwindSetName,
			strings.TrimSuffix(detection.TailwindSkipped, ", so its rules are skipped"), configuration.TailwindSetName)
	}
	return configuration.LoadHouse(location.LintConfigFileName, registeredRuleNames(), detection)
}

// usesHouseSets reports whether the run takes the house stack: its settings file is the default one and
// is missing, or names no set cohere carries.
func usesHouseSets(location projectLocation) bool {
	if _, err := os.Stat(location.LintConfigFileName); err != nil && location.LintConfigFileNameGiven {
		return false
	}
	zero, err := configuration.IsZeroConfig(location.LintConfigFileName)
	// A chain that cannot be read is not zero config: LoadFor reads it again and fails the run naming why.
	return err == nil && zero
}

// houseSources is lintConfigSources under zero config: the house sets and the project's own file, when it
// has one, so a cache keyed on them changes when either does.
func houseSources(location projectLocation) []string {
	if !usesHouseSets(location) {
		return lintConfigSources(location.LintConfigFileName)
	}
	sources, err := configuration.HouseSources(location.LintConfigFileName)
	if err != nil {
		return []string{location.LintConfigFileName}
	}
	return sources
}

// writeSetsLine prints the line that says, under zero config, which house sets apply where, and which
// were looked for and not applied, with the evidence for each. Nothing for a configuration that names
// its own sets.
//
// Written to out and not as an invocation line: which sets applied is part of the verdict, true of the
// tree rather than of this run, so a replay prints it too and nothing is ever applied silently.
func writeSetsLine(out io.Writer, lintConfig *configuration.Config) {
	if line := lintConfig.SetsLine(); line != "" {
		fmt.Fprintln(out, line)
	}
}

// errHouseSetsPerFile is --print-config's and --list-rules-enabled's answer under zero config. Which house
// sets a file gets is read from the program, which those two flags do not build, so an answer without it
// would be a guess; --explain builds it and says what ran on one file and why.
var errHouseSetsPerFile = errors.New("this project has no CohereSettings.json that names its sets, so which house sets " +
	"apply to a file is decided from the program per file, and --print-config does not build the program: " +
	"run cohere --explain <file> to see what runs on it and why")
