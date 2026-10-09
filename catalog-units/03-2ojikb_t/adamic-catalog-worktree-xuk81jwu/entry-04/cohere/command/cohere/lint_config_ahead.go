package main

import (
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
)

// lintConfigAhead is the lint config read on its own goroutine while the graph builds (#g3046x5).
//
// Loading the config and checking its settings need nothing from the program for a project whose settings name
// their own sets. On a cold ahra run the two sat on the path between the graph and the walk for 45 to 50ms:
// about 19ms of work measured alone (11.5ms loading, 6.7ms of the first settings check), the rest the main
// goroutine waiting for a core beside the ten others busy then. Begun before the build, they run during it, which
// is bound by the kernel's file opens rather than the cores. Only the selectors, which ask which files an
// override reaches, wait for the program, and they took about 2.5ms.
//
// A zero-config project's sets are detected from the program's files (housesets.Detect), so its config is read
// after the build, as it always was: startLintConfigAhead returns nil for it.
type lintConfigAhead struct {
	done chan struct{}

	// Set before done closes. The errors are kept apart so configureLint reports them in the order it always did:
	// the load's, then the selectors', then the settings'.
	config        *configuration.Config
	loadError     error
	settingsError error
	options       configuration.OptionsRegistry
}

// startLintConfigAhead begins reading the config and returns at once, or returns nil when the config can only be
// read with the program built.
func startLintConfigAhead(location projectLocation) *lintConfigAhead {
	if usesHouseSets(location) {
		return nil
	}
	ahead := &lintConfigAhead{done: make(chan struct{})}
	go func() {
		defer close(ahead.done)
		// What loadLintConfig does for a project that names its sets.
		ahead.config, ahead.loadError = configuration.LoadFor(location.LintConfigFileName, registeredRuleNames())
		if ahead.loadError != nil {
			return
		}
		base := optionsBase(ahead.config, location.Root)
		ahead.settingsError = registry.CheckSettings(base)
		if ahead.settingsError == nil {
			ahead.options = registry.OptionsAt(base)
		}
	}()
	return ahead
}
