package lower

import (
	"errors"
	"os"
	"testing"
)

func TestConstructorCacheOneArgumentRemainsNotYet(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile("testdata/constructor_cache_one_argument.a")
	if err != nil {
		test.Fatal(err)
	}
	_, err = lowerSource(test, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) ||
		notYet.What != "a lexical constructor-cache initializer needing additional captures" {
		test.Fatalf("want additional-capture NotYet for a one-argument initializer, got %v", err)
	}
}
