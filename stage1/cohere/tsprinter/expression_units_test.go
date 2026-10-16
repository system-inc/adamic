package tsprinter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
)

func expressionSelection(t *testing.T, count int) []bool {
	t.Helper()
	selected, err := selectExpressionUnits(os.Getenv("ADAMIC_TEST_SHARD"), count)
	if err != nil {
		t.Fatal(err)
	}
	return selected
}

func selectExpressionUnits(value string, count int) ([]bool, error) {
	index, boxes := 0, 1
	if value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("ADAMIC_TEST_SHARD=%q: want i/n", value)
		}
		var err error
		index, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		boxes, err = strconv.Atoi(parts[1])
		if err != nil || boxes <= 0 || index < 0 || index >= boxes {
			return nil, fmt.Errorf("ADAMIC_TEST_SHARD=%q: require 0 <= i < n", value)
		}
	}
	selected := make([]bool, count)
	for i := range selected {
		selected[i] = i%boxes == index
	}
	return selected, nil
}

func TestExpressionUnitBoxSelection(t *testing.T) {
	t.Parallel()
	for _, boxes := range []int{1, 2, 7, 65, 100} {
		seen := make([]int, oracles.ExpressionsAgainstGoAndPrettierShards)
		for box := 0; box < boxes; box++ {
			selected, err := selectExpressionUnits(fmt.Sprintf("%d/%d", box, boxes), oracles.ExpressionsAgainstGoAndPrettierShards)
			if err != nil {
				t.Fatal(err)
			}
			for unit, runs := range selected {
				if runs {
					seen[unit]++
				}
			}
		}
		for unit, count := range seen {
			if count != 1 {
				t.Fatalf("shard-%03d assigned %d times across %d boxes", unit, count, boxes)
			}
		}
	}
	for _, value := range []string{"1", "x/2", "0/0", "-1/2", "2/2", "0/x"} {
		if _, err := selectExpressionUnits(value, oracles.ExpressionsAgainstGoAndPrettierShards); err == nil {
			t.Fatalf("invalid selector accepted: %s", value)
		}
	}
	selected, err := selectExpressionUnits("", oracles.ExpressionsAgainstGoAndPrettierShards)
	if err != nil {
		t.Fatal(err)
	}
	for unit, runs := range selected {
		if !runs {
			t.Fatalf("unset selector omitted shard-%03d", unit)
		}
	}
}

func expressionDisagreement(plan *oracles.CorpusPlan, number int, want []byte, result oracles.Run) error {
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		return fmt.Errorf("shard-%03d: exit %d stderr %s", number, result.ExitCode, result.Stderr)
	}
	if bytes.Equal(result.Stdout, want) {
		return nil
	}
	actual, expected := bytes.Split(result.Stdout, []byte("\n")), bytes.Split(want, []byte("\n"))
	for i := 0; i < max(len(actual), len(expected)); i++ {
		if i < len(actual) && i < len(expected) && bytes.Equal(actual[i], expected[i]) {
			continue
		}
		label := "answer count"
		if i < len(plan.Shards[number].Indices) {
			label = plan.Labels[plan.Shards[number].Indices[i]]
		}
		return fmt.Errorf("shard-%03d %s: %s", number, label, firstDifference(string(result.Stdout), string(want)))
	}
	return fmt.Errorf("shard-%03d: byte disagreement", number)
}

// A real oracle process emits a planted wrong answer; only its owning unit
// catches it through the same comparison used by the complete expression test.
func TestExpressionUnitPlantedDisagreement(t *testing.T) {
	t.Parallel()
	plan := &oracles.CorpusPlan{Labels: []string{"case-0", "case-1", "case-2", "case-3"}}
	for _, ids := range oracles.AssignCorpus([]int{1, 1, 1, 1}, 2) {
		plan.Shards = append(plan.Shards, oracles.CorpusShard{Indices: ids})
	}
	catches := 0
	for number, shard := range plan.Shards {
		var want, actual strings.Builder
		for _, id := range shard.Indices {
			want.WriteString("ok\n")
			if id == 2 {
				actual.WriteString("planted\n")
			} else {
				actual.WriteString("ok\n")
			}
		}
		encoded, _ := json.Marshal(actual.String())
		result := executeOne(t, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		err := expressionDisagreement(plan, number, []byte(want.String()), result)
		if err != nil {
			catches++
			if !strings.Contains(err.Error(), "shard-000 case-2") {
				t.Fatal(err)
			}
			t.Logf("planted disagreement caught: %v", err)
		}
	}
	if catches != 1 {
		t.Fatalf("planted case caught by %d units, want exactly one", catches)
	}
}

func TestExpressionUnitUnionRejectsMissingAndRepeated(t *testing.T) {
	t.Parallel()
	for _, ids := range [][]int{{0}, {0, 0}} {
		plan := &oracles.CorpusPlan{Labels: []string{"first", "second"}, Shards: []oracles.CorpusShard{{Indices: ids}}}
		if oracles.ExpressionUnion(plan, [][]byte{[]byte("ok\nok\n")}, []byte("ok\nok\n")) == nil {
			t.Fatalf("invalid union accepted: %v", ids)
		}
	}
}

// Include transitive implementation sources and embedded files. Broad directory
// inputs deliberately invalidate a future product hash on any dependency edit.
func expressionInputFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	seen := map[string]bool{}
	tree, err := filepath.Abs(oracles.Repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range roots {
		absolute, err := filepath.Abs(root)
		if err != nil {
			t.Fatal(err)
		}
		err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Only what the tree carries: a checkout's untracked files aren't in a runner's source (#smkk3et).
			if entry.Name() == ".git" || path != absolute && !buildcache.Tracked(tree, path, entry.IsDir()) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() {
				if entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type().IsRegular() {
				seen[path] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	files := make([]string, 0, len(seen))
	for path := range seen {
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}
