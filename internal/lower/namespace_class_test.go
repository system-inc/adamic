package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNamespaceClassEarlyConstructionStaysLoud(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		"function make(){return new Debug.Mapper();} const value=make(); namespace Debug {export class Mapper {kind=0;}}",
	)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "namespace read before runtime initialization") {
		test.Fatalf("early constructor read trusted: %v", err)
	}
}
