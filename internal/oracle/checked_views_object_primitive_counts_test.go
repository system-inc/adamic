package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

const objectPrimitiveBatchCounts = "../../stage3/interface-downcasts/lane4b/original/batch-counts.json"

// Declaration-backed fixtures live outside the ordinary self-contained fixture
// registry. Keep their measured rows with their pinned declaration evidence.
func objectPrimitiveOriginalCount(t *testing.T, program *ir.Program, fixture string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	name, args := pinnedStack(binary)
	result := execute(t, name, args...)
	match := countsLine.FindSubmatch(result.stderr)
	if match == nil {
		t.Fatalf("missing fixture counts: exit %d stderr %q", result.exitCode, result.stderr)
	}
	regions, merges := "0", "0"
	if graph := graphCountsLine.FindSubmatch(result.stderr); graph != nil {
		regions, merges = string(graph[1]), string(graph[2])
	}
	row := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s/%s", match[1], match[2], match[3], match[4], match[5], match[6], regions, merges)
	records := map[string]string{}
	if data, err := os.ReadFile(checkedViewFixturePath(objectPrimitiveBatchCounts)); err == nil {
		if err := json.Unmarshal(data, &records); err != nil {
			t.Fatal(err)
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if *updateCounts {
		records[fixture] = row
		data, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(objectPrimitiveBatchCounts, append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	} else if records[fixture] != row {
		t.Fatalf("fixture counts changed for %s: recorded %s measured %s; regenerate this suite with -args -update-counts", fixture, records[fixture], row)
	}
	t.Logf("counts %s: %s (allocations/frees/retains/releases/peak/arena/graph regions/merges)", fixture, row)
}

// The complete integration counts update can fail on unrelated fixtures. This
// focused updater changes only this lane's measured row, in registry order.
func TestCheckedViewObjectPrimitiveFixCounts(t *testing.T) {
	for _, name := range []string{"fixes-graph", "fixes-tuple-unread", "fixes-tuple"} {
		t.Run(name, func(t *testing.T) {
			objectPrimitiveFixCount(t, "stage3/interface-downcasts/lane4b/fixtures/"+name+".a")
		})
	}
}

func objectPrimitiveFixCount(t *testing.T, path string) {
	row := counted(t, checkedViewFixturePath(path), false, nil, false, false)
	data, err := os.ReadFile(checkedViewFixturePath(countsPath))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	prefix := "| " + path + " |"
	found := false
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		found = true
		if *updateCounts {
			text = strings.Replace(text, line, row, 1)
		} else if line != row {
			t.Fatalf("graph counts changed: recorded %s measured %s", line, row)
		}
	}
	if !found {
		if !*updateCounts {
			t.Fatal("graph fixture has no recorded row")
		}
		anchor := "| stage3/interface-downcasts/readiness-number-uninitialized.a |"
		if strings.Contains(path, "tuple-unread") {
			anchor = "| stage3/interface-downcasts/lane4b/fixtures/fixes-graph.a |"
		} else if strings.HasSuffix(path, "/fixes-tuple.a") {
			anchor = "| stage3/interface-downcasts/lane4b/fixtures/fixes-tuple-unread.a |"
		}
		start := strings.Index(text, anchor)
		if start < 0 {
			t.Fatal("counts registry anchor drift")
		}
		end := start + strings.Index(text[start:], "\n") + 1
		text = text[:end] + row + "\n" + text[end:]
	}
	if *updateCounts {
		if err := os.WriteFile(countsPath, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Log(row)
}
