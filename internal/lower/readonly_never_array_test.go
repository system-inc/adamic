package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestWritableNeverArrayRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/writable_never_array_refused.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a never[] seen as writable T[], which can write T into a shared never[]" || refused.Fix != "declare the result readonly T[], or return a fresh [] (adamic/invariant-mutable)" || !strings.HasSuffix(refused.Where, ":3:12") {
		t.Fatalf("want pinned writable never[] refusal, got %v", err)
	}
}
