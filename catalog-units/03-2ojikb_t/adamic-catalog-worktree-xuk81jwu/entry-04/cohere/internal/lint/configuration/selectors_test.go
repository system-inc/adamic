package configuration

import (
	"strings"
	"testing"
)

// TestZeroMatchOverrideSelectorIsRefused is the failure this guard exists to expose.
//
// The valid selector beside the miss is load-bearing: a validator that checks the files array as a
// single OR would see modules/** match and let the stale generated-code selector remain silent.
func TestZeroMatchOverrideSelectorIsRefused(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Root: "/repo",
		Overrides: []Override{{
			Files: []string{"modules/**", "**/generated-miss/**/*.ts"},
		}},
	}

	err := configuration.ValidateSelectors([]string{
		"/repo/index.ts",
		"/repo/modules/finance/Sync.ts",
		"/repo/libraries/api/generated/Operations.ts",
	})
	if err == nil {
		t.Fatal("a config with one live selector and one zero-match selector validated, so the live " +
			"one hid the vacuous one")
	}
	if !strings.Contains(err.Error(), "generated-miss") {
		t.Errorf("the error does not name the selector a reader must fix: %v", err)
	}
	if strings.Contains(err.Error(), `"modules/**"`) {
		t.Errorf("the error names a selector that did match: %v", err)
	}
}

// TestEveryMatchingSelectorPasses is the other direction. Without it a validator that rejects every
// override would satisfy the failure case above while making all real configurations unusable.
func TestEveryMatchingSelectorPasses(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Root: "/repo",
		Overrides: []Override{
			{Files: []string{"**/generated/**/*.{ts,tsx}"}},
			{Files: []string{"**/next-env.d.ts"}},
			{Files: []string{"modules/**"}},
		},
	}

	err := configuration.ValidateSelectors([]string{
		"/repo/libraries/api/generated/Operations.ts",
		"/repo/next-env.d.ts",
		"/repo/modules/finance/Sync.ts",
	})
	if err != nil {
		t.Fatalf("the three live selector shapes did not validate against files each one reaches: %v", err)
	}
}

// TestASelectorReachingOnlyIgnoredFilesIsRefused pins the effective population rather than the disk
// population. An override cannot affect a file excluded before override resolution, so matching one
// must not make the selector look alive.
func TestASelectorReachingOnlyIgnoredFilesIsRefused(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Root:           "/repo",
		IgnorePatterns: []string{"generated/**"},
		Overrides: []Override{{
			Files: []string{"generated/**"},
		}},
	}

	err := configuration.ValidateSelectors([]string{
		"/repo/index.ts",
		"/repo/generated/Operations.ts",
	})
	if err == nil {
		t.Fatal("a selector reaching only globally ignored files validated even though no rule can run there")
	}
	if !strings.Contains(err.Error(), "1 non-ignored project files") {
		t.Errorf("the error does not state the population it actually checked: %v", err)
	}
}

// TestAnOverrideWithoutSelectorsIsRefused covers the zero-length form of the same defect. There is
// no pattern to misspell, but the block can never apply and previously disappeared just as quietly.
func TestAnOverrideWithoutSelectorsIsRefused(t *testing.T) {
	t.Parallel()
	configuration := &Config{Overrides: []Override{{}}}

	err := configuration.ValidateSelectors([]string{"index.ts"})
	if err == nil {
		t.Fatal("an override with no files patterns validated even though it cannot select a file")
	}
	if !strings.Contains(err.Error(), "override 0 with no files patterns") {
		t.Errorf("the error does not identify the empty override: %v", err)
	}
}
