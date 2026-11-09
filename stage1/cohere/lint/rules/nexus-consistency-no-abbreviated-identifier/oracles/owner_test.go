package oracles

import (
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: changes the working directory and enables production before tests run.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	producer = true
	os.Exit(m.Run())
}

func TestProduct_TypedOwnership(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	oracle := TypedOwnership(t, root)
	info, err := os.Stat(oracle)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatal("typed oracle is not executable")
	}
}
