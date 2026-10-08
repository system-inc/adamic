package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestEvolvingStringMixedWritesStayNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `let value; value = "owned:" + true.toString(); value = {word: "object"};`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a value of type any") {
		t.Fatalf("want incompatible evolving writes unsupported, got %v", err)
	}
}
