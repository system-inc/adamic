package oracle

import (
	"os"
	"strings"
	"testing"
)

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
		"stage3/parser-ahead/node-array/fields.a",
		"stage3/parser-ahead/rulings/null-placeholder-pending.a",
	} {
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) string { return counted(t, path, false, nil, false, false) }})
	}
}

// Each leaf measures one parser fixture; only TestCountsAreRecorded writes the ledger.
func assertParserAheadCount(t *testing.T, path string) {
	t.Helper()
	row := counted(t, path, false, nil, false, false)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	key := strings.Split(row, " | ")[0] + " | "
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, key) {
			if line != row {
				t.Fatalf("recorded %s, measured %s", line, row)
			}
			return
		}
	}
	t.Fatalf("missing counts %s", row)
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

func TestParserAheadCountsRulingsNullPlaceholderPending(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/rulings/null-placeholder-pending.a")
}

func TestParserAheadCountsNodeArrayFields(t *testing.T) {
	t.Parallel()
	assertParserAheadCount(t, "stage3/parser-ahead/node-array/fields.a")
}
