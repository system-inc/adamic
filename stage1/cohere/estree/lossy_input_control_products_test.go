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
	return threePortMutation{
		name: "lossy-input-control",
		file: "pipeline.ts",
		from: `if(text.includes('\ufffd'))`,
		to:   `if(false)`,
	}
}

var lossyInputControlShared struct {
	once                   sync.Once
	oracle, source, native string
}

func lossyInputControlPrepare(test *testing.T) {
	test.Helper()
	lossyInputControlShared.once.Do(func() {
		lossyInputControlShared.oracle = scalarEdgeOracle(test)
		mutation := lossyInputControlMutation()
		lowered := threePortLoweredProduct(test, mutation)
		// node reads the copy whose imports name this checkout; the product names it <repository> (#tqrqx60).
		lossyInputControlShared.source = resolvedPort(test, lowered, filepath.Join("source", "main.ts"))
		lossyInputControlShared.native = threePortNativeProduct(test, mutation)
	})
	if lossyInputControlShared.native == "" {
		test.Fatal("lossy-input preparation failed")
	}
}

func TestProduct_LossyInputControlLowered(test *testing.T) {
	test.Parallel()
	threePortLoweredProduct(test, lossyInputControlMutation())
}
func TestProduct_LossyInputControlNative(test *testing.T) {
	test.Parallel()
	threePortNativeProduct(test, lossyInputControlMutation())
}

func TestLossyInputControl(test *testing.T) {
	test.Parallel()
	// No setup deadline: prepare once before timing only this unit's work.
	lossyInputControlPrepare(test)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	path := filepath.Join(test.TempDir(), "malformed.ts")
	if writeError := os.WriteFile(path, []byte{47, 47, 240, 144, 128, 10, 120, 59}, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	want := threePortExecute(test, ctx, lossyInputControlShared.oracle, path)
	for name, got := range map[string][]byte{
		"Node": threePortExecute(
			test,
			ctx,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(root(test), "oracle/node.mjs"),
			lossyInputControlShared.source,
			path,
		),
		"native": threePortExecute(test, ctx, lossyInputControlShared.native, path),
	} {
		if d := firstDifference(want, got); d == "" {
			test.Fatal(name + " lossy-input mutant survived")
		} else {
			test.Log(name + ": " + d)
		}
	}
}
