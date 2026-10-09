package estree

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func lossyInputControlMutation() threePortMutation {
	return threePortMutation{name: "lossy-input-control", file: "pipeline.ts", from: `if(text.includes('\ufffd'))`, to: `if(false)`}
}

var lossyInputControlShared struct {
	once                   sync.Once
	oracle, source, native string
}

func lossyInputControlPrepare(t *testing.T) {
	t.Helper()
	lossyInputControlShared.once.Do(func() {
		lossyInputControlShared.oracle = scalarEdgeOracle(t)
		mutation := lossyInputControlMutation()
		lowered := threePortLoweredProduct(t, mutation)
		lossyInputControlShared.source = filepath.Join(lowered, "source", "main.ts")
		lossyInputControlShared.native = threePortNativeProduct(t, mutation)
	})
	if lossyInputControlShared.native == "" {
		t.Fatal("lossy-input preparation failed")
	}
}

func TestProduct_LossyInputControlLowered(t *testing.T) {
	t.Parallel()
	threePortLoweredProduct(t, lossyInputControlMutation())
}
func TestProduct_LossyInputControlNative(t *testing.T) {
	t.Parallel()
	threePortNativeProduct(t, lossyInputControlMutation())
}

func TestLossyInputControl(t *testing.T) {
	t.Parallel()
	// No setup deadline: prepare once before timing only this unit's work.
	lossyInputControlPrepare(t)
	script := mutantEmittedProduct(t, lossyInputControlShared.source)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "malformed.ts")
	if err := os.WriteFile(path, []byte{47, 47, 240, 144, 128, 10, 120, 59}, 0644); err != nil {
		t.Fatal(err)
	}
	want := threePortExecute(t, ctx, lossyInputControlShared.oracle, path)
	for name, got := range map[string][]byte{
		"Node":    threePortExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), lossyInputControlShared.source, path),
		"native":  threePortExecute(t, ctx, lossyInputControlShared.native, path),
		"emitted": mutantEmittedOutput(t, lossyInputControlShared.source, script, path),
	} {
		if d := firstDifference(want, got); d == "" {
			t.Fatal(name + " lossy-input mutant survived")
		} else {
			t.Log(name + ": " + d)
		}
	}
}
