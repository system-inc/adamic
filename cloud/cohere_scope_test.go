package cloud

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// scopeKinds are the kinds docs/cohere-scope.md may give an ignored glob: a tree of inputs, a committed record, or a
// generated file, the last naming its generator in the row.
var scopeKinds = map[string]bool{"input": true, "record": true, "generated": true}

// scopeRow is one row of docs/cohere-scope.md's table: `| `glob` | kind | why | read by |`, with the padding cohere's
// format adds to align the columns.
var scopeRow = regexp.MustCompile("^\\| `([^`]+)` +\\| ([a-z]+) +\\| ([^|]+) \\| ([^|]+) \\|$")

// checkCohereScope holds CohereSettings.json's ignorePatterns and docs/cohere-scope.md's table to one another, so no
// glob leaves cohere's judgment without a written reason (Adamic's ruling on #h2c34ad): every glob is in both, once,
// each row has a known kind, a reason and a reader, and no glob is a bare extension.
func checkCohereScope(settings []byte, document string) error {
	var parsed struct {
		IgnorePatterns []string `json:"ignorePatterns"`
	}
	if decodeError := json.Unmarshal(settings, &parsed); decodeError != nil {
		return fmt.Errorf("CohereSettings.json: %v", decodeError)
	}
	documented := map[string]bool{}
	for _, line := range strings.Split(document, "\n") {
		// The ignore table comes first. A `##` section after it, such as the trees out of tsconfig's plain check,
		// holds tables of its own that aren't ignores.
		if strings.HasPrefix(line, "## ") {
			break
		}
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		row := scopeRow.FindStringSubmatch(line)
		if row == nil {
			return fmt.Errorf("docs/cohere-scope.md: %q isn't a row of glob, kind, why and read by", line)
		}
		glob, kind := row[1], row[2]
		if documented[glob] {
			return fmt.Errorf("docs/cohere-scope.md lists %s twice", glob)
		}
		if !scopeKinds[kind] {
			return fmt.Errorf("docs/cohere-scope.md gives %s the kind %q, not input, record or generated", glob, kind)
		}
		if strings.TrimSpace(row[3]) == "" || strings.TrimSpace(row[4]) == "" {
			return fmt.Errorf("docs/cohere-scope.md gives %s no reason or no reader", glob)
		}
		documented[glob] = true
	}
	inSettings := map[string]bool{}
	for _, glob := range parsed.IgnorePatterns {
		if inSettings[glob] {
			return fmt.Errorf("CohereSettings.json ignores %s twice", glob)
		}
		inSettings[glob] = true
		if strings.HasPrefix(glob, "**/*.") || strings.HasPrefix(glob, "*.") {
			return fmt.Errorf(
				"CohereSettings.json ignores %s, a bare extension, which would hide hand-written source",
				glob,
			)
		}
		if !documented[glob] {
			return fmt.Errorf("CohereSettings.json ignores %s, which docs/cohere-scope.md gives no reason for", glob)
		}
	}
	for glob := range documented {
		if !inSettings[glob] {
			return fmt.Errorf(
				"docs/cohere-scope.md gives a reason for %s, which CohereSettings.json doesn't ignore",
				glob,
			)
		}
	}
	return nil
}

func TestCohereScopeEveryIgnoreHasItsReason(test *testing.T) {
	test.Parallel()
	// Reads: CohereSettings.json
	settings, readError := os.ReadFile(filepath.Join("..", "CohereSettings.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	// Reads: docs/cohere-scope.md
	document, readError := os.ReadFile(filepath.Join("..", "docs", "cohere-scope.md"))
	if readError != nil {
		test.Fatal(readError)
	}
	if checkCohereScopeError := checkCohereScope(settings, string(document)); checkCohereScopeError != nil {
		test.Fatal(checkCohereScopeError)
	}
}

// Each planted pair breaks one rule and must be refused, so the check above could have failed.
func TestCohereScopeRefusesAnUnpairedIgnore(test *testing.T) {
	test.Parallel()
	row := "| `**/testdata/**` | input | Test inputs. | Their tests. |\n"
	good := `{"ignorePatterns": ["**/testdata/**"]}`
	if checkCohereScopeError := checkCohereScope([]byte(good), row); checkCohereScopeError != nil {
		test.Fatalf("a good pair was refused: %v", checkCohereScopeError)
	}
	// The same row aligned the way cohere's format pads a table is the same row.
	padded := "| `**/testdata/**`   | input     | Test inputs.   | Their tests.   |\n"
	if checkCohereScopeError := checkCohereScope([]byte(good), padded); checkCohereScopeError != nil {
		test.Fatalf("a padded row was refused: %v", checkCohereScopeError)
	}
	// A later section's table is not read as ignores, so its three-column rows neither fail nor count.
	later := row + "\n## Out of tsconfig\n\n| `internal/oracle/testdata` | Why. | Its test. |\n"
	if checkCohereScopeError := checkCohereScope([]byte(good), later); checkCohereScopeError != nil {
		test.Fatalf("a later section's table was read as ignores: %v", checkCohereScopeError)
	}
	for name, planted := range map[string][2]string{
		"no reason":        {`{"ignorePatterns": ["**/testdata/**", "stage3/x.json"]}`, row},
		"no setting":       {`{"ignorePatterns": []}`, row},
		"bare extension":   {`{"ignorePatterns": ["**/*.json"]}`, "| `**/*.json` | record | Records. | Readers. |\n"},
		"unknown kind":     {good, "| `**/testdata/**` | convenience | Quiet. | Nobody. |\n"},
		"empty reason":     {good, "| `**/testdata/**` | input |   | Their tests. |\n"},
		"listed twice":     {good, row + row},
		"ignored twice":    {`{"ignorePatterns": ["**/testdata/**", "**/testdata/**"]}`, row},
		"malformed row":    {good, "| `**/testdata/**` | input |\n"},
		"settings refused": {`{"ignorePatterns": "**/testdata/**"}`, row},
	} {
		if checkCohereScopeError := checkCohereScope([]byte(planted[0]), planted[1]); checkCohereScopeError == nil {
			test.Errorf("%s: the pair was accepted", name)
		}
	}
}
