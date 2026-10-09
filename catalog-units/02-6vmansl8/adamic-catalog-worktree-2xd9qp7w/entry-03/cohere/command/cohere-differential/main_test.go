package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/differential"
	"github.com/system-inc/cohere/internal/lint/configuration"
)

// The control precondition has to be able to reject, or it is the thing it guards against.
//
// A control that misses is supposed to mean the harness is blind. A control written where its rule
// cannot fire misses identically, so the evidence and the defect are indistinguishable and the
// reader attributes the failure to the wrong thing. The check exists to make that case impossible,
// which means a check that cannot reject is worse than no check: it adds the appearance of a
// guarantee to a run that has none.
func TestAControlWhereItsRuleCannotFireIsRefused(t *testing.T) {
	t.Parallel()
	lintConfig := configFrom(t, map[string]any{
		"rules":          map[string]any{"nexus/consistency-no-enum": "error"},
		"ignorePatterns": []string{"ignored-here/**"},
	})

	// The good case first, so a check that rejects everything cannot pass this test.
	lintable := differential.Control{
		Name:         "shared-enum",
		RelativePath: filepath.Join("linted", "Planted.ts"),
		Rule:         "consistency-no-enum",
	}
	if err := checkControlsAreLintable(lintConfig, []differential.Control{lintable}, nil); err != nil {
		t.Fatalf("a control on a linted path with an enabled rule must be accepted: %v", err)
	}

	ignoredPath := lintable
	ignoredPath.RelativePath = filepath.Join("ignored-here", "Planted.ts")
	err := checkControlsAreLintable(lintConfig, []differential.Control{ignoredPath}, nil)
	if err == nil {
		t.Fatal("a control on an ignored path must be refused: it would miss for a reason unrelated to the harness")
	}
	// The message has to name the pattern. "Refused" alone sends the reader looking at the harness,
	// which is the confusion this whole check exists to prevent.
	if !strings.Contains(err.Error(), "ignored-here/**") {
		t.Fatalf("the refusal must name the ignore pattern that would silence it, got %v", err)
	}

	unconfiguredRule := lintable
	unconfiguredRule.Rule = "import-require-path-alias"
	err = checkControlsAreLintable(lintConfig, []differential.Control{unconfiguredRule}, nil)
	if err == nil {
		t.Fatal("a control naming a rule the config does not enable must be refused")
	}
	if !strings.Contains(err.Error(), "import-require-path-alias") {
		t.Fatalf("the refusal must name the rule that would not fire, got %v", err)
	}
}

// An override that scopes a rule off for one directory is invisible from the base rule list.
//
// This is the case a check written against `lintConfig.Rules` alone would sail past: the rule is
// enabled at the base level, the path is not ignored, and the control would still be silent because
// an override turned that rule off for exactly that directory.
func TestAnOverrideThatScopesTheRuleOffIsCaught(t *testing.T) {
	t.Parallel()
	lintConfig := configFrom(t, map[string]any{
		"rules": map[string]any{"nexus/consistency-no-enum": "error"},
		"overrides": []any{map[string]any{
			"files": []string{"generated/**/*.ts"},
			"rules": map[string]any{"nexus/consistency-no-enum": "off"},
		}},
	})

	control := differential.Control{
		Name:         "shared-enum",
		RelativePath: filepath.Join("generated", "Planted.ts"),
		Rule:         "consistency-no-enum",
	}
	if err := checkControlsAreLintable(lintConfig, []differential.Control{control}, nil); err == nil {
		t.Fatal("a control under an override that turns its rule off must be refused, even though the base config enables the rule")
	}

	// The same rule and the same config, outside the override's scope, must still be accepted.
	elsewhere := control
	elsewhere.RelativePath = filepath.Join("source", "Planted.ts")
	if err := checkControlsAreLintable(lintConfig, []differential.Control{elsewhere}, nil); err != nil {
		t.Fatalf("the override must not reject a control outside its scope: %v", err)
	}
}

// The control declarations name bare rules and the config keys carry plugin prefixes.
//
// Without the lookup that reconciles them, every control would be refused as unconfigured and this
// guard would block every run, which is a failure that at least announces itself. Tested anyway
// because the opposite fix, dropping the check, is the tempting one.
func TestABareRuleNameFindsItsPluginPrefixedConfigKey(t *testing.T) {
	t.Parallel()
	lintConfig := configFrom(t, map[string]any{
		"rules": map[string]any{"nexus/consistency-no-enum": "error"},
	})

	if got := pluginQualified(lintConfig, "consistency-no-enum"); got != "nexus/consistency-no-enum" {
		t.Fatalf("pluginQualified(%q) = %q, want the config's own key", "consistency-no-enum", got)
	}
	// A rule the config never mentions returns the bare name, so Enabled reports unconfigured rather
	// than this helper deciding the control is valid.
	if got := pluginQualified(lintConfig, "never-configured"); got != "never-configured" {
		t.Fatalf("an unconfigured rule must come back unchanged, got %q", got)
	}
}

// configFrom writes a lint config to disk and loads it through the real loader.
//
// Built through configuration.Load rather than by hand so the test exercises the same parsing, override
// ordering, and root resolution the command does. A hand-built Config would test this guard against
// a shape the loader never produces.
func configFrom(t *testing.T, raw map[string]any) *configuration.Config {
	t.Helper()

	directory := t.TempDir()
	path := filepath.Join(directory, ".oxlintrc.json")
	contents, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := configuration.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}
