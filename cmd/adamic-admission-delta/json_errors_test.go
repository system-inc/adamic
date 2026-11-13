package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The four ruled Error writes bind the exact fixture bytes to the shared stop and lifting task.
func TestRuntimeNotYetJSONErrorRows(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	list, err := readRuntimeNotYetStops(filepath.Join(root, "cloud/admission-corpus/runtime-notyet-stops.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"direct", "view", "logical", "message_absent"} {
		path := "internal/oracle/testdata/json_errors/" + name + ".a"
		source, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		got, ok := list.Stops[contentHash(source)]
		want := runtimeNotYetStop{Path: path, RulingTask: "#z4a9p9e", LiftingTask: "#ejcaz75", Stderr: "adamic: panic: " + ir.JSONErrorWriteStop + "\n"}
		if !ok || got != want {
			t.Errorf("%s: got %+v, listed %v; want %+v", name, got, ok, want)
		}
	}
}
