package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestMixedArrayElementStorageRemainsRefused(t *testing.T) {
	t.Parallel()
	source := "type Target=readonly number[]|readonly string[];const values:Target='x'.length===1?[7]:['ok'];console.log(`${values[0]}`);"
	path := filepath.Join(t.TempDir(), "storage.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), loaded)
	var unsupported *NotYet
	if !errors.As(err, &unsupported) || !strings.Contains(unsupported.What, "where an array goes") {
		t.Fatalf("mixed physical storage must retain its array-reader refusal: %v", err)
	}
}
