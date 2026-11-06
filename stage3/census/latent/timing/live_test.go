package timing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLatentUnitTimesVisibleBeforePassReturns20pb346(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "latent-times.jsonl")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	timer := NewUnitTimer(output)
	firstDone, resume, runDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(runDone)
		timer.Start("entry.ts", "entry.ts:1:1")()
		close(firstDone)
		<-resume
		timer.Start("entry.ts", "entry.ts:2:1")()
	}()
	defer func() { close(resume); <-runDone }()
	<-firstDone
	select {
	case <-runDone:
		t.Fatal("pass returned before observation")
	default:
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var row UnitTime
	if err := json.Unmarshal(data, &row); err != nil {
		t.Fatalf("unit was not visible before pass return: %q: %v", data, err)
	}
	if row.Where != "entry.ts:1:1" || row.File != "entry.ts" {
		t.Fatalf("wrong live record: %+v", row)
	}
}
