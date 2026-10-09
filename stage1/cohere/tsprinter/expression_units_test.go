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
		seen := make([]int, 65)
		for box := 0; box < boxes; box++ {
			selected, err := selectExpressionUnits(fmt.Sprintf("%d/%d", box, boxes), 65)
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
		if _, err := selectExpressionUnits(value, 65); err == nil {
			t.Fatalf("invalid selector accepted: %s", value)
		}
	}
	selected, err := selectExpressionUnits("", 65)
	if err != nil {
		t.Fatal(err)
	}
	for unit, runs := range selected {
		if !runs {
			t.Fatalf("unset selector omitted shard-%03d", unit)
		}
	}
}

func expressionUnits(t *testing.T, path, want string, count int) (*corpusPlan, [][]byte) {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(strings.TrimSuffix(path, ".txt") + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []printerCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	rows := bytes.Split(bytes.TrimSuffix(text, []byte("\n")), []byte("\n"))
	answers := bytes.Split(bytes.TrimSuffix([]byte(want), []byte("\n")), []byte("\n"))
	if len(rows) != len(cases) || len(answers) != len(cases) || len(cases) == 0 {
		t.Fatalf("expression enumeration: %d cases, %d rows, %d answers", len(cases), len(rows), len(answers))
	}
	// Every pinned upstream record must occur in the unsplit enumeration. Each
	// unit then checks the exact independently pinned outcome for its own cases.
	for _, record := range upstreamOutcomes(t) {
		if record.Family != "expressions" {
			continue
		}
		found := false
		for _, item := range cases {
			if strings.HasSuffix(item.Label, record.Label) && item.Source == record.Source && item.Want == record.Go {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("recorded upstream case absent: %s", record.Label)
		}
	}
	plan := &corpusPlan{labels: make([]string, len(cases))}
	weights := make([]int, len(cases))
	for i, item := range cases {
		plan.labels[i] = fmt.Sprintf("case-%06d %s", i, item.Label)
		weights[i] = len(rows[i])
	}
	directory := t.TempDir()
	var outputs [][]byte
	for number, indices := range assignCorpus(weights, count) {
		shard := corpusShard{indices: indices, text: filepath.Join(directory, fmt.Sprintf("shard-%03d.txt", number)), specs: filepath.Join(directory, fmt.Sprintf("shard-%03d.json", number))}
		var protocol, expected bytes.Buffer
		subset := make([]printerCase, 0, len(indices))
		for _, i := range indices {
			protocol.Write(rows[i])
			protocol.WriteByte('\n')
			expected.Write(answers[i])
			expected.WriteByte('\n')
			subset = append(subset, cases[i])
		}
		encoded, err := json.Marshal(subset)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.text, protocol.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.specs, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		plan.shards = append(plan.shards, shard)
		outputs = append(outputs, expected.Bytes())
	}
	if err := expressionUnion(plan, outputs, []byte(want)); err != nil {
		t.Fatal(err)
	}
	gapData, err := os.ReadFile(filepath.Dir(path) + "/gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var gaps []printerCase
	if err := json.Unmarshal(gapData, &gaps); err != nil {
		t.Fatal(err)
	}
	gapAnswers, err := os.ReadFile(filepath.Dir(path) + "/gap-answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) != bytes.Count(gapAnswers, []byte("\n")) {
		t.Fatal("gap enumeration/answer count differs")
	}
	whole := &corpusPlan{labels: append([]string(nil), plan.labels...), shards: append([]corpusShard(nil), plan.shards...)}
	gapIDs := make([]int, len(gaps))
	for i, item := range gaps {
		gapIDs[i] = len(cases) + i
		whole.labels = append(whole.labels, fmt.Sprintf("case-%06d gap %s", gapIDs[i], item.Label))
	}
	whole.shards = append(whole.shards, corpusShard{indices: gapIDs})
	wholeOutputs := append(append([][]byte(nil), outputs...), gapAnswers)
	if err := expressionUnion(whole, wholeOutputs, append([]byte(want), gapAnswers...)); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique expression case ids in %d units; %d gap ids in one unit; total %d", len(cases), len(plan.shards), len(gaps), len(cases)+len(gaps))
	return plan, outputs
}

func expressionUnion(plan *corpusPlan, outputs [][]byte, want []byte) error {
	// Position is part of the case id: generated cases may share a label.
	seen := make(map[int]bool, len(plan.labels))
	count := 0
	for number, shard := range plan.shards {
		for _, id := range shard.indices {
			if id < 0 || id >= len(plan.labels) || seen[id] {
				return fmt.Errorf("shard-%03d: missing, extra or repeated case id %d", number, id)
			}
			seen[id] = true
			count++
		}
	}
	if count != len(plan.labels) {
		return fmt.Errorf("union: %d case ids, want %d", count, len(plan.labels))
	}
	for id := range plan.labels {
		if !seen[id] {
			return fmt.Errorf("union: missing case id %d", id)
		}
	}
	merged, err := mergeCorpus(plan, outputs)
	if err != nil {
		return err
	}
	if !bytes.Equal(merged, want) {
		return fmt.Errorf("union differs from unsplit answers")
	}
	return nil
}

func expressionDisagreement(plan *corpusPlan, number int, want []byte, result run) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("shard-%03d: exit %d stderr %s", number, result.exitCode, result.stderr)
	}
	if bytes.Equal(result.stdout, want) {
		return nil
	}
	actual, expected := bytes.Split(result.stdout, []byte("\n")), bytes.Split(want, []byte("\n"))
	for i := 0; i < max(len(actual), len(expected)); i++ {
		if i < len(actual) && i < len(expected) && bytes.Equal(actual[i], expected[i]) {
			continue
		}
		label := "answer count"
		if i < len(plan.shards[number].indices) {
			label = plan.labels[plan.shards[number].indices[i]]
		}
		return fmt.Errorf("shard-%03d %s: %s", number, label, firstDifference(string(result.stdout), string(want)))
	}
	return fmt.Errorf("shard-%03d: byte disagreement", number)
}

// A real oracle process emits a planted wrong answer; only its owning unit
// catches it through the same comparison used by the complete expression test.
func TestExpressionUnitPlantedDisagreement(t *testing.T) {
	t.Parallel()
	plan := &corpusPlan{labels: []string{"case-0", "case-1", "case-2", "case-3"}}
	for _, ids := range assignCorpus([]int{1, 1, 1, 1}, 2) {
		plan.shards = append(plan.shards, corpusShard{indices: ids})
	}
	catches := 0
	for number, shard := range plan.shards {
		var want, actual strings.Builder
		for _, id := range shard.indices {
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
		plan := &corpusPlan{labels: []string{"first", "second"}, shards: []corpusShard{{indices: ids}}}
		if expressionUnion(plan, [][]byte{[]byte("ok\nok\n")}, []byte("ok\nok\n")) == nil {
			t.Fatalf("invalid union accepted: %v", ids)
		}
	}
}

// Include transitive implementation sources and embedded files. Broad directory
// inputs deliberately invalidate a future product hash on any dependency edit.
func expressionInputFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	seen := map[string]bool{}
	for _, root := range roots {
		absolute, err := filepath.Abs(root)
		if err != nil {
			t.Fatal(err)
		}
		err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Name() == ".git" {
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
