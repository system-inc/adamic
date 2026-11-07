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
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{
			path,
			func(test *testing.T) (string, func() string) { return counted(test, path, false, nil, false, false) },
		})
	}
}

// Each leaf measures one parser fixture; only TestCountsAreRecorded writes the ledger.
func assertParserAheadCount(test *testing.T, path string) {
	test.Helper()
	_, measure := counted(test, path, false, nil, false, false)
	row := measure()
	recorded, readCountRowError := readCountRow(countsDirectory, "runtime", path)
	if readCountRowError != nil {
		test.Fatalf("missing counts %s: %v", row, readCountRowError)
	}
	if recorded != row {
		test.Fatalf("recorded %s, measured %s", recorded, row)
	}
}

func TestParserAheadCountsFactoriesAlreadyComplete(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/factories/already-complete.a")
}

func TestParserAheadCountsFactoriesComplete(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/factories/complete.a")
}

func TestParserAheadCountsFactoriesEscaped(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/factories/escaped.a")
}

func TestParserAheadCountsRulingsFactoryUnset(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/factory-unset.a")
}

func TestParserAheadCountsRulingsFactoryUse(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/factory-use.a")
}

func TestParserAheadCountsRulingsOwnKeyOrder(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/own-key-order.a")
}

func TestParserAheadCountsRulingsNodeArrayKeys(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/node-array-keys.a")
}

func TestParserAheadCountsRulingsNodeArrayJson(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/node-array-json.a")
}

func TestParserAheadCountsRulingsNodeArrayLength(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/node-array-length.a")
}

func TestParserAheadCountsRulingsNodeArrayUnset(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/node-array-unset.a")
}

func TestParserAheadCountsRulingsNodeArrayRecords(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/node-array-records.a")
}

func TestParserAheadCountsRulingsNullPlaceholderPending(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/rulings/null-placeholder-pending.a")
}

func TestParserAheadCountsNodeArrayFields(test *testing.T) {
	test.Parallel()
	assertParserAheadCount(test, "stage3/parser-ahead/node-array/fields.a")
}
