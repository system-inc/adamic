package oracle

import "testing"

func init() {
	for _, path := range []string{
		"stage3/parser-ahead/factories/already-complete.a",
		"stage3/parser-ahead/factories/complete.a",
		"stage3/parser-ahead/factories/escaped.a",
		"stage3/parser-ahead/rulings/factory-unset.a",
		"stage3/parser-ahead/rulings/factory-use.a",
		"stage3/parser-ahead/rulings/own-key-order.a",
		"stage3/parser-ahead/rulings/node-array-keys.a",
		"stage3/parser-ahead/rulings/node-array-json.a",
		"stage3/parser-ahead/rulings/node-array-length.a",
		"stage3/parser-ahead/rulings/node-array-unset.a",
		"stage3/parser-ahead/rulings/node-array-records.a",
		"stage3/parser-ahead/node-array/fields.a",
		"stage3/parser-ahead/rulings/null-placeholder-pending.a",
	} {
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
	}
}

// Each leaf measures one parser fixture; only TestCountsAreRecorded writes the ledger.
func assertParserAheadCount(t *testing.T, path string) {
	t.Helper()
	_, measure := counted(t, path, false, nil, false, false)
	row := measure()
	recorded, err := readCountRow(countsDirectory, "runtime", path)
	if err != nil {
		t.Fatalf("missing counts %s: %v", row, err)
	}
	if recorded != row {
		t.Fatalf("recorded %s, measured %s", recorded, row)
	}
}

func TestParserAheadCountsFactoriesAlreadyComplete(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/factories/already-complete.a")
}

func TestParserAheadCountsFactoriesComplete(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/factories/complete.a")
}

func TestParserAheadCountsFactoriesEscaped(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/factories/escaped.a")
}

func TestParserAheadCountsRulingsFactoryUnset(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/factory-unset.a")
}

func TestParserAheadCountsRulingsFactoryUse(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/factory-use.a")
}

func TestParserAheadCountsRulingsOwnKeyOrder(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/own-key-order.a")
}

func TestParserAheadCountsRulingsNodeArrayKeys(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/node-array-keys.a")
}

func TestParserAheadCountsRulingsNodeArrayJson(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/node-array-json.a")
}

func TestParserAheadCountsRulingsNodeArrayLength(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/node-array-length.a")
}

func TestParserAheadCountsRulingsNodeArrayUnset(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/node-array-unset.a")
}

func TestParserAheadCountsRulingsNodeArrayRecords(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/node-array-records.a")
}

func TestParserAheadCountsRulingsNullPlaceholderPending(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/null-placeholder-pending.a")
}

func TestParserAheadCountsNodeArrayFields(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/node-array/fields.a")
}
