package native

import (
	goruntime "runtime"
	"sync/atomic"

	"github.com/system-inc/adamic/internal/toolchain"
)

// All builds share the process CPU budget, including runtime compilation,
// preprocessing, probes and links. A named Loom FIFO further bounds the box.
var clangTokens = make(chan struct{}, max(1, goruntime.GOMAXPROCS(0)))
var clangActive atomic.Int64
var clangPeak atomic.Int64

func acquireClang() func() {
	clangTokens <- struct{}{}
	active := clangActive.Add(1)
	for peak := clangPeak.Load(); active > peak; peak = clangPeak.Load() {
		if clangPeak.CompareAndSwap(peak, active) {
			break
		}
	}
	return func() { clangActive.Add(-1); <-clangTokens }
}
func compilerVersion(compiler string) ([]byte, error) {
	command := toolchain.Command(compiler, "--version")
	var output []byte
	err := withCompilerToken(func() (err error) { output, err = command.CombinedOutput(); return err })
	return output, err
}
