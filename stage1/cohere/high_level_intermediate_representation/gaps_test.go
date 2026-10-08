package high_level_intermediate_representation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Roadmap step 28: assert both meaning and the exact pending compiler refusal.
// A successful lowering fails this test so the ruling and certificate are revisited.
func TestOptionalBooleanGapStandsWhereGapsMdSays(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs("testdata/optional-boolean-gap.a")
	if err != nil {
		t.Fatal(err)
	}
	got := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", path)
	if string(got) != "true\n" {
		t.Fatalf("Node output %q; REPORT.md records true\\n", got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil {
		t.Fatal("this gap lowers now: update REPORT.md and recertify the complete native corpus")
	}
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("stage 0 refuses another way now: %v", err)
	}
	const want = "a field of type boolean | undefined"
	if notYet.What != want {
		t.Fatalf("lower.NotYet.What = %q, want %q", notYet.What, want)
	}
}

// The complete Go census, independent of the port's admission grammar, is the
// source of truth for semantic gap impact. Save keys and test-call provenance.
func recordOptionalBooleanImpact(t *testing.T, destination string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(destination, "records.json"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Key      string   `json:"key"`
		Calls    []string `json:"calls"`
		Dump     string   `json:"dump"`
		Excluded string   `json:"excluded"`
	}
	var census struct{ Records []record }
	if err := json.Unmarshal(data, &census); err != nil {
		t.Fatal(err)
	}
	type affected struct {
		Key   string   `json:"key"`
		Calls []string `json:"calls"`
	}
	matches := regexp.MustCompile(`(?m)^terminal [0-9]+ Optional `)
	total := 0
	rows := []affected{}
	for _, r := range census.Records {
		original := false
		for _, call := range r.Calls {
			if !strings.HasPrefix(call, "probe:") {
				original = true
			}
		}
		if !original {
			continue
		}
		total++
		if matches.MatchString(r.Dump) {
			rows = append(rows, affected{r.Key, r.Calls})
		}
	}
	if total != 1465 || len(rows) != 74 {
		t.Fatalf("optional boolean impact changed: %d affected / %d originals; revisit REPORT.md", len(rows), total)
	}
	out, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "optional-boolean-impact.json"), append(out, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("optional boolean semantic impact: %d/%d original graphs; %d do not read that field", len(rows), total, total-len(rows))
}

func TestStaticConstructorSpreadGapStandsWhereGapsMdSays(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs("testdata/static-constructor-spread-gap.a")
	if err != nil {
		t.Fatal(err)
	}
	got := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", path)
	if string(got) != "0\n" {
		t.Fatalf("Node output %q, want 0\\n", got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil {
		t.Fatal("static constructor spread lowers now: remove REPORT.md gap 2 and recertify CloneFunction natively")
	}
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("gap refuses another way now: %v", err)
	}
	const want = "spreading in a program with static constructor objects"
	if notYet.What != want {
		t.Fatalf("lower.NotYet.What = %q, want %q", notYet.What, want)
	}
}
