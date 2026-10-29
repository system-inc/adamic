package timing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"
)

func TestLatentUnitTimerWritesSeparateSequentialRecords(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	timer := NewUnitTimer(&output)
	timer.Start("first.ts", "first.ts:1:1")()
	func() {
		defer timer.Start("second.ts", "second.ts:2:1")()
		defer func() { _ = recover() }()
		panic("lowering failure")
	}()
	decoder := json.NewDecoder(&output)
	for index, name := range []string{"first.ts", "second.ts"} {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		wall, wallOK := record["wall_seconds"].(float64)
		cpu, cpuOK := record["cpu_seconds"].(float64)
		where := []string{"first.ts:1:1", "second.ts:2:1"}[index]
		if len(record) != 4 || record["file"] != name || record["where"] != where || !wallOK || !cpuOK || wall <= 0 || cpu < 0 {
			t.Fatalf("invalid unit time: %+v", record)
		}
	}
	var extra UnitTime
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("extra record: %v", err)
	}
	var disabled *UnitTimer
	disabled.Start("untimed.ts", "untimed.ts:1:1")()
}
