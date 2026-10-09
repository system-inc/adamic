package lower

import (
	"errors"
	"os"
	"testing"
)

func TestConstructorCacheOneArgumentRemainsNotYet(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/constructor_cache_one_argument.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a lexical constructor-cache initializer needing additional captures" {
		t.Fatalf("want additional-capture NotYet for a one-argument initializer, got %v", err)
	}
}
