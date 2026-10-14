package oracle

import (
	"os"
	"testing"
)

func TestV4ExistingExplicitMethodReceiver(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/namespace_method_receiver.a")
	if err != nil {
		t.Fatal(err)
	}
	v4EscapeSource(t, string(source), "7:99\n11\n")
}
