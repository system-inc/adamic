package css

import (
	"os"
	"testing"
)

// Not parallel: establishes the shared helpers' css-relative working directory,
// grants this compiler-free binary ownership, and cleans shared build inputs.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	oracleProducer = true
	code := m.Run()
	if cssParserTopRoot != "" {
		if err := os.RemoveAll(cssParserTopRoot); err != nil && code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
