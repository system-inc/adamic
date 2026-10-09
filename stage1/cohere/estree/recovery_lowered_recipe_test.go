package estree

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testRecoveryLoweredRecipeShards = 1

// ADAMIC_TEST_SHARD=i/n selects shards by index modulo n; unset runs all.
// Independently lower the same source at different private paths, so removing
// those paths from the product key is justified by the actual C and JS bytes.
func TestRecoveryLoweredRecipe(t *testing.T) {
	setup := beginRecoverySetup(t)
	defer setup.report(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "(scanner.flags & 1) === 0 &&\n        (token"
	copied := mutantPort(t, "sourceLookahead.ts", anchor, anchor)
	var products []string
	for i, path := range []string{main, copied} {
		inputs := recoveryCacheInputs(t, path)
		inputs.Name = []string{"estree-lowering-proof-a", "estree-lowering-proof-b"}[i]
		inputs.Files = append(inputs.Files, "stage1/cohere/estree/recovery_lowered_recipe_test.go")
		products = append(products, recoveryProduct(t, setup, inputs, func(dir string) error {
			program, err := load.Load([]string{path})
			if err != nil {
				return err
			}
			ir, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(ir)), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(ir)), 0644)
		}))
	}
	shards := partitionRecovery(t, recoveryCases("lowering-recipe", []string{"independent-source-paths"}), testRecoveryLoweredRecipeShards)
	runRecoveryShards(t, shards, func(t *testing.T, _ int, _ []recoveryCase) {
		for _, name := range []string{"port.c", "port.mjs"} {
			a, err := os.ReadFile(filepath.Join(products[0], name))
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(products[1], name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(a, b) {
				t.Fatalf("%s differs across source paths", name)
			}
		}
	})
}
