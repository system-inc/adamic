package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNamespaceClassEarlyConstructionStaysLoud(t *testing.T) {
	_, err := lowerSource(t, "function make(){return new Debug.Mapper();} const value=make(); namespace Debug {export class Mapper {kind=0;}}")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "before all runtime namespaces") {
		t.Fatalf("early constructor read trusted: %v", err)
	}
}
