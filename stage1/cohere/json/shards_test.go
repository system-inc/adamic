package json

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// Bound both case count and text bytes; oversized corpus files stand alone.
// These are deterministic bounds, never selected from run-time measurements.
func jsonPortShards(cases []textCase) []nativeChunk {
	var shards []nativeChunk
	for start := 0; start < len(cases); {
		end, bytes := start, 0
		for end < len(cases) && end-start < 16 {
			size := len(cases[end].Text)
			if end > start && bytes+size > 128*1024 {
				break
			}
			bytes += size
			end++
		}
		shards = append(shards, nativeChunk{start: start, end: end})
		start = end
	}
	return shards
}

func jsonPortUnion(cases []textCase, shards []nativeChunk) error {
	expected := make(map[string]textCase, len(cases))
	for _, item := range cases {
		if _, exists := expected[item.Name]; exists {
			return fmt.Errorf("unsplit enumeration repeats case id %q", item.Name)
		}
		expected[item.Name] = item
	}
	seen := make(map[string]bool, len(cases))
	count := 0
	for ordinal, shard := range shards {
		if shard.start < 0 || shard.end > len(cases) || shard.end <= shard.start {
			return fmt.Errorf("shard-%04d invalid range [%d:%d]", ordinal, shard.start, shard.end)
		}
		for _, item := range cases[shard.start:shard.end] {
			if seen[item.Name] {
				return fmt.Errorf("shard-%04d repeats case id %q", ordinal, item.Name)
			}
			if want, exists := expected[item.Name]; !exists || want != item {
				return fmt.Errorf("shard-%04d unexpected case id %q", ordinal, item.Name)
			}
			seen[item.Name] = true
			count++
		}
	}
	if count != len(cases) {
		return fmt.Errorf("shard union has %d cases, unsplit enumeration has %d", count, len(cases))
	}
	for id := range expected {
		if !seen[id] {
			return fmt.Errorf("shard union missing case id %q", id)
		}
	}
	return nil
}

func jsonPortSelection(value string) (int, int, error) {
	if value == "" {
		return 0, 1, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) == 2 {
		index, a := strconv.Atoi(parts[0])
		count, b := strconv.Atoi(parts[1])
		if a == nil && b == nil && count > 0 && index >= 0 && index < count {
			return index, count, nil
		}
	}
	return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD=%q: expected zero-based i/n with 0 <= i < n", value)
}

func TestJSONPortShardUnion(t *testing.T) {
	t.Parallel()
	cases := make([]textCase, 39)
	for index := range cases {
		cases[index] = textCase{Name: fmt.Sprintf("case-%d.json", index), Text: "{}"}
	}
	cases[17].Text = strings.Repeat(" ", 128*1024+1)
	shards := jsonPortShards(cases)
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct {
		name   string
		shards []nativeChunk
	}{
		{"missing", shards[1:]},
		{"repeated", append(append([]nativeChunk(nil), shards...), shards[0])},
		{"out of bounds", []nativeChunk{{start: 0, end: len(cases) + 1}}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if err := jsonPortUnion(cases, mutant.shards); err == nil {
				t.Fatal("invalid union survived")
			}
		})
	}
	for _, value := range []string{"0/0", "-1/2", "2/2", "bad", "0/2/3"} {
		if _, _, err := jsonPortSelection(value); err == nil {
			t.Fatalf("invalid selector %q survived", value)
		}
	}
	for boxes := 1; boxes <= 7; boxes++ {
		seen := make([]int, len(shards))
		for box := 0; box < boxes; box++ {
			i, n, err := jsonPortSelection(fmt.Sprintf("%d/%d", box, boxes))
			if err != nil {
				t.Fatal(err)
			}
			for ordinal := range shards {
				if ordinal%n == i {
					seen[ordinal]++
				}
			}
		}
		for ordinal, count := range seen {
			if count != 1 {
				t.Fatalf("shard-%04d assigned %d times", ordinal, count)
			}
		}
	}
}

func TestJSONPortShardDisagreement(t *testing.T) {
	t.Parallel()
	cases := make([]textCase, 33)
	answers := make([]answer, 33)
	for index := range cases {
		cases[index] = textCase{Name: fmt.Sprintf("case-%d.json", index), Text: "{}"}
		answers[index] = answer{Output: "{}"}
	}
	const planted = 17
	shards := jsonPortShards(cases)
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	caught := 0
	for ordinal, shard := range shards {
		name := fmt.Sprintf("shard-%04d", ordinal)
		_, expected := protocol(cases[shard.start:shard.end], answers[shard.start:shard.end])
		observed := append([]answer(nil), answers[shard.start:shard.end]...)
		holds := planted >= shard.start && planted < shard.end
		if holds {
			observed[planted-shard.start].Output = "{X}"
		}
		_, got := protocol(cases[shard.start:shard.end], observed)
		err := comparisonError(name, run{stdout: []byte(got)}, expected, cases[shard.start:shard.end])
		if (err != nil) != holds {
			t.Fatalf("%s planted case ownership disagrees with comparison: %v", name, err)
		}
		if err != nil {
			caught++
			if !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), cases[planted].Name) {
				t.Fatalf("missing shard or case name: %v", err)
			}
			t.Logf("caught planted disagreement only in %s: %v", name, err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards", caught)
	}
}

// Decode the Go driver's escaped protocol for optional per-case calibration.
func jsonPortAnswers(protocol string, count int) ([]answer, error) {
	lines := strings.Split(strings.TrimSuffix(protocol, "\n"), "\n")
	if len(lines) != count {
		return nil, fmt.Errorf("Go answered %d/%d cases", len(lines), count)
	}
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	answers := make([]answer, count)
	for index, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Go case %d malformed answer", index)
		}
		value := unescape.Replace(parts[1])
		switch parts[0] {
		case "ok":
			answers[index].Output = value
		case "error":
			answers[index].Error = value
		default:
			return nil, fmt.Errorf("Go case %d unknown answer kind %q", index, parts[0])
		}
	}
	return answers, nil
}
