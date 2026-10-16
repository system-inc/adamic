package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProduction()
	os.Exit(m.Run())
}
func TestProduct_MarkdownInlineGoOverlay(t *testing.T) { t.Parallel(); Printer(t) }
