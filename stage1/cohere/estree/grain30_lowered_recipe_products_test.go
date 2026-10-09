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
	t.Parallel()
	setup := beginRecoverySetup(t)
	defer setup.report(t)
	products := []string{estreeLoweredRecipeProduct(t, 0), estreeLoweredRecipeProduct(t, 1)}
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

func estreeLoweredRecipeProduct(t *testing.T, i int) string {
	t.Helper()
	setup := beginRecoverySetup(t)
	main := estreeMainPath(t)
	path := main
	if i == 1 {
		const anchor = "(scanner.flags & 1) === 0 &&\n        (token"
		path = mutantPort(t, "sourceLookahead.ts", anchor, anchor)
	}
	inputs := recoveryCacheInputs(t, path)
	inputs.Name = []string{"estree-lowering-proof-a", "estree-lowering-proof-b"}[i]
	inputs.Files = append(inputs.Files, "stage1/cohere/estree/grain30_lowered_recipe_products_test.go")
	return recoveryProduct(t, setup, inputs, func(dir string) error {
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
	})
}

func TestProduct_EstreeLoweredRecipeA(t *testing.T) { t.Parallel(); estreeLoweredRecipeProduct(t, 0) }
func TestProduct_EstreeLoweredRecipeB(t *testing.T) { t.Parallel(); estreeLoweredRecipeProduct(t, 1) }
