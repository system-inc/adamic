package regex

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Generated table.json and shapes/fixtures.json are fixed by cohere commit
// 7945d102a6c18dd36adf9114a758ce646e8b2359 (testdata/shapes/CENSUS.md).
// Their pinned fixture count guards corpus loss; sample totals are enumerated live.
const testShapeFixturesShards = 107

// ADAMIC_TEST_SHARD=i/n (zero-based i) selects units locally; unset runs all.
// The gate selects stable shard-NNN subtests directly with -run. Build products
// are prepared once and shared until internal/buildcache is available on the base.
func TestShapeFixtures(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("table.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ ID string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	fixturesData, err := os.ReadFile("testdata/shapes/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		ID      string
		Samples []json.RawMessage
	}
	if err := json.Unmarshal(fixturesData, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(rows) != testShapeFixturesShards || len(fixtures) != testShapeFixturesShards {
		t.Fatalf("enumerated %d table rows / %d fixtures, declared %d shards", len(rows), len(fixtures), testShapeFixturesShards)
	}
	// Enumerate unsplit case IDs independently, then count the actual shard plan.
	expected := map[string]bool{}
	sampleTotal := 0
	for _, fixture := range fixtures {
		if fixture.ID == "" || expected[fixture.ID] {
			t.Fatalf("repeated or empty fixture ID %q", fixture.ID)
		}
		expected[fixture.ID] = true
		for sample := range fixture.Samples {
			expected[fmt.Sprintf("%s/sample-%d", fixture.ID, sample)] = true
			sampleTotal++
		}
	}
	shards := make([]int, len(rows))
	actual := map[string]bool{}
	for i, row := range rows {
		shards[i] = i
		if row.ID != fixtures[i].ID {
			t.Fatalf("fixture/table order drift at %d", i)
		}
		ids := []string{row.ID}
		for sample := range fixtures[i].Samples {
			ids = append(ids, fmt.Sprintf("%s/sample-%d", row.ID, sample))
		}
		for _, id := range ids {
			if actual[id] || !expected[id] {
				t.Fatalf("repeated or unexpected shard case %q", id)
			}
			actual[id] = true
		}
	}
	if len(shards) != testShapeFixturesShards || len(actual) != len(expected) {
		t.Fatalf("shard union count differs from unsplit enumeration")
	}
	for id := range expected {
		if !actual[id] {
			t.Fatalf("missing shard case %q", id)
		}
	}
	t.Logf("union: %d fixtures and %d samples, %d unique IDs across %d shards", len(rows), sampleTotal, len(actual), len(shards))
	dir := t.TempDir()
	binary := filepath.Join(dir, "shapes-gate")
	// Inputs: Name=shapes-gate; Files=gate.go and transitive Go dependencies;
	// Flags=go build; Toolchain=Go. No package-local cache.
	product := func(dir string) error {
		command := exec.Command("go", "build", "-o", filepath.Join(dir, "shapes-gate"), "stage1/cohere/lint/regex/testdata/shapes/gate.go")
		command.Dir = repository
		output, err := command.CombinedOutput()
		if err != nil || len(output) != 0 {
			return fmt.Errorf("gate build: %v: %s", err, output)
		}
		return nil
	}
	start := time.Now()
	if err := product(dir); err != nil {
		t.Fatal(err)
	}
	t.Logf("build Go shapes gate wall %.6fs", time.Since(start).Seconds())
	invoke := func(arguments ...string) ([]byte, error) {
		command := exec.Command(binary, arguments...)
		command.Dir = repository
		return command.CombinedOutput()
	}
	products := filepath.Join(dir, "products")
	start = time.Now()
	prepared, err := invoke("-prepare", products)
	if err != nil {
		t.Fatalf("prepare: %v\n%s", err, prepared)
	}
	t.Log(strings.TrimSpace(string(prepared)))
	t.Logf("build all fixture products wall %.6fs", time.Since(start).Seconds())
	allPending := strings.Contains(string(prepared), fmt.Sprintf("PREPARED fixtures=%d pending=%d", len(rows), len(rows)))
	selected, count := shapeShardSelection(t)
	for _, i := range shards {
		if selected >= 0 && i%count != selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %s exceeds 30s", elapsed)
				}
			}()
			t.Logf("fixture %d: %s; %d samples; all original backend checks", i, rows[i].ID, len(fixtures[i].Samples))
			args := []string{"-products", products, "-case", strconv.Itoa(i)}
			if os.Getenv("ADAMIC_REGEX_TRANSLATION_MUTANT") == "1" {
				args = append(args, "-mutant")
			}
			output, err := invoke(args...)
			if err != nil || !strings.Contains(string(output), "PASS fixtures=1 ") {
				t.Fatalf("fixture comparison: %v\n%s", err, output)
			}
			t.Log(strings.TrimSpace(string(output)))
			// Plant the real one-character translation mutant in fixture 0. Every
			// other shard must still pass; only shard-000 owns and catches this mutant.
			t.Run("translation-mutant", func(t *testing.T) {
				output, err := invoke("-products", products, "-case", strconv.Itoa(i), "-mutant")
				if i == 0 {
					if err == nil || !strings.Contains(string(output), rows[0].ID+": source Node fixture mismatch") {
						t.Fatalf("mutant escaped or failed outside comparison: %v\n%s", err, output)
					}
					t.Logf("planted translation mutant caught by shard-000: %s", rows[0].ID)
				} else if err != nil || !strings.Contains(string(output), "PASS fixtures=1 ") {
					t.Fatalf("mutant leaked into shard-%03d: %v\n%s", i, err, output)
				}
			})
			t.Run("native-transition", func(t *testing.T) {
				output, err := invoke("-products", products, "-case", strconv.Itoa(i), "-force-native")
				if allPending && i == 0 {
					if err == nil || !strings.Contains(string(output), "awaits codex/regex-runtime-compiler: forced native requirement caught refusal") {
						t.Fatalf("native refusal control failed: %v\n%s", err, output)
					}
				} else if err != nil || !strings.Contains(string(output), "PASS fixtures=1 ") {
					t.Fatalf("native transition control failed: %v\n%s", err, output)
				}
			})
		})
	}
}

func shapeShardSelection(t *testing.T) (selected, count int) {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return -1, 1
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q: want i/n", value)
	}
	selected, err := strconv.Atoi(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	count, err = strconv.Atoi(parts[1])
	if err != nil || count < 1 || selected < 0 || selected >= count {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return selected, count
}
