package docsdata

import (
	"os"
	"path/filepath"

	"github.com/system-inc/cohere/internal/benchresults"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/settingsschema"
)

// SourceInputs gathers every input but the binary's help from the module at root: the live registry
// and sets, the files beside them, and the committed examples. The generator adds the help, and
// replaces the examples when it recaptures.
func SourceInputs(root string) (Inputs, error) {
	names := registry.Names()
	inputs := Inputs{
		Registrations: rule.Registered(),
		SetNames:      configuration.SetNames(),
		LoadSet:       func(name string) (*configuration.Config, error) { return configuration.LoadFor(name, names) },
		SetContents:   configuration.SourceContents,
	}
	var err error
	if inputs.SwiftVerdicts, err = os.ReadFile(filepath.Join(root, "swift", "HouseRuleVerdicts.json")); err != nil {
		return inputs, err
	}
	if inputs.SwiftRules, err = os.ReadFile(filepath.Join(root, "swift", "Rules.json")); err != nil {
		return inputs, err
	}
	if inputs.Changelog, err = os.ReadFile(filepath.Join(root, "CHANGELOG.md")); err != nil {
		return inputs, err
	}
	if inputs.SettingsSchemas, err = settingsschema.Build(names); err != nil {
		return inputs, err
	}
	if inputs.Benchmarks, err = benchresults.Read(filepath.Join(root, benchresults.Directory)); err != nil {
		return inputs, err
	}
	inputs.Examples, err = ReadCommittedExamples(root)
	return inputs, err
}
