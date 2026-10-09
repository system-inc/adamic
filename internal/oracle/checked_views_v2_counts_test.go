package oracle

import (
	"path/filepath"
	"testing"
)

func init() { additionalFixtureCounts = append(additionalFixtureCounts, checkedViewsV2Counts) }
func checkedViewsV2Counts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, path := range []string{
		"stage3/interface-downcasts/lane4b/fixtures/diagnostic-boolean.a",
		"stage3/interface-downcasts/lane4b/fixtures/diagnostic-code-wrong.a",
		"stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.a",
		"stage3/interface-downcasts/lane4b/fixtures/diagnostic-wrong.a",
		"stage3/interface-downcasts/lane4b/fixtures/literal-boolean.a",
		"stage3/interface-downcasts/lane4b/fixtures/literal-good.a",
		"stage3/interface-downcasts/lane4b/fixtures/literal-negative-wrong.a",
		"stage3/interface-downcasts/lane4b/fixtures/literal-text-wrong.a",
		"stage3/interface-downcasts/lane4b/fixtures/node-indicator-false.a",
		"stage3/interface-downcasts/lane4b/fixtures/node-indicator-flags-wrong.a",
		"stage3/interface-downcasts/lane4b/fixtures/node-indicator-good.a",
		"stage3/interface-downcasts/lane4b/fixtures/node-indicator-wrong.a",
		"stage3/interface-downcasts/lane4b/fixtures/required-missing.a",
		"stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.a",
		"stage3/interface-downcasts/untagged/fixtures/binding-name-source-nested.a",
		"stage3/interface-downcasts/untagged/fixtures/binding-name-source-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/callable-union-good-number.a",
		"stage3/interface-downcasts/untagged/fixtures/callable-union-good-string.a",
		"stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a",
		"stage3/interface-downcasts/untagged/fixtures/callable-union-optional-boundary.a",
		"stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/class-data-good.a",
		"stage3/interface-downcasts/untagged/fixtures/class-data-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-callback-good.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-callback-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-generic-good.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-generic-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-helper-good.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-helper-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-stored-good.a",
		"stage3/interface-downcasts/untagged/fixtures/flow-stored-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/option-element-source-good.a",
		"stage3/interface-downcasts/untagged/fixtures/option-element-source-nested.a",
		"stage3/interface-downcasts/untagged/fixtures/option-element-source-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/recursive-absent.a",
		"stage3/interface-downcasts/untagged/fixtures/recursive-good.a",
		"stage3/interface-downcasts/untagged/fixtures/recursive-nested.a",
		"stage3/interface-downcasts/untagged/fixtures/recursive-wrong.a",
		"stage3/interface-downcasts/untagged/fixtures/structural-source-good.a",
		"stage3/interface-downcasts/untagged/fixtures/structural-source-nested.a",
		"stage3/interface-downcasts/untagged/fixtures/structural-source-wrong.a",
		"stage3/interface-downcasts/v2/callable-producer-wrong.a",
		"stage3/interface-downcasts/v2/fixed-tuple-good.a",
		"stage3/interface-downcasts/v2/fixed-tuple-wrong.a",
		"stage3/interface-downcasts/v2/fs-option-boxing.a",
		"stage3/interface-downcasts/v2/mixed-read-write.a",
		"stage3/interface-downcasts/v2/null-read-write.a",
		"stage3/interface-downcasts/v2/representation-read-write.a",
		"stage3/interface-downcasts/v2/tuple-identity-wrong.a",
		"stage3/interface-downcasts/v2/undefined-read-write.a",
	} {
		// V4's ruled .a escapes stop before emission, so they no longer have
		// runtime counts. Keep their Node control and path-and-fix refusal here.
		field := ""
		switch path {
		case "stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a":
			field = "view.holder.value"
		case "stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a", "stage3/interface-downcasts/v2/callable-producer-wrong.a":
			field = "view.value"
		}
		if field != "" {
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			v4UnionReadRefusal(t, absolute, field)
			continue
		}
		rows = append(rows, counted(t, path, false, nil, false, false))
	}
	return rows
}
