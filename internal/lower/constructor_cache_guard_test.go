package lower

import (
	"errors"
	"os"
	"testing"
)

func TestConstructorCacheOneArgumentRemainsNotYet(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/constructor_cache_one_argument.a")
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(source))
	var notYet *NotYet
	if !errors.As(lowerSourceError, &notYet) ||
		notYet.What != "a lexical constructor-cache initializer needing additional captures" {
		test.Fatalf("want additional-capture NotYet for a one-argument initializer, got %v", lowerSourceError)
	}
}
