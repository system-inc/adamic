package oracle

import "testing"

// The manual lifecycle probe uses Linux ioctl interposition and GNU ld wraps.
// Other platforms keep the unconditional emitted pipe/file/null comparisons.
func TestProduct_StdoutColumnsProbes(t *testing.T) {
	t.Parallel()
	stdoutColumnsProducts(t)
}
