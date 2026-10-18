package fresh_test

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checkStep21CorpusRefusal(t *testing.T, path string) {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	header := strings.SplitN(string(source), "\n", 2)[0]
	reason, ok := strings.CutPrefix(header, "// a-check: refused ")
	if !ok {
		t.Fatalf("%s lacks its refusal header", path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{absolute})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if err == nil || !strings.HasSuffix(err.Error(), "Adamic 0.1 refuses "+reason) {
		t.Fatalf("%s: lost its exact refusal: %v", path, err)
	}
}
