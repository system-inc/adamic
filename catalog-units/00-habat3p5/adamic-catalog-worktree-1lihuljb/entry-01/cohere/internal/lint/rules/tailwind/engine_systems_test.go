package tailwind

import (
	"testing"

	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
)

// engineSystem is one design system an assertion about the engine runs on.
type engineSystem struct {
	name string
	load func(t *testing.T) DesignSystemResult
}

// engineSystems are the design systems every assertion about the engine runs on (#f598zk0).
//
// The public one is the independent theme in this package's testdata over the vendored tailwindcss,
// so the assertion runs on every machine and on CI. ahra's is the live one these tests were written
// against, opt-in through its corpus: a synthetic theme proves the engine on the shapes we thought of,
// and the real one is where the shapes we did not think of live.
var engineSystems = []engineSystem{
	{name: "public", load: publicDesignSystem},
	{name: "ahra", load: classOrderLiveRepositorySystem},
}

// forEachEngineSystem runs an engine assertion once per design system, each as its own subtest, so a
// machine without the ahra corpus skips that subtest by name and still runs the public one.
func forEachEngineSystem(t *testing.T, run func(t *testing.T, designSystem DesignSystemResult)) {
	t.Helper()
	for _, system := range engineSystems {
		t.Run(system.name, func(t *testing.T) {
			t.Parallel()
			run(t, system.load(t))
		})
	}
}

// publicDesignSystem is the independent theme's design system, with its table, in the shape the live
// helpers return.
func publicDesignSystem(t *testing.T) DesignSystemResult {
	t.Helper()
	system := independentLiveSystem(t)
	return DesignSystemResult{System: system, Table: tailwindengine.NewTable(system)}
}
