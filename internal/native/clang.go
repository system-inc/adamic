package native

import (
	"os/exec"
	goruntime "runtime"
	"sync/atomic"
)

// All builds in this process share the CPU budget, including runtime compilation,
// preprocessing, compiler probes and linking. Per-build worker limits still apply.
var clangTokens = make(chan struct{}, goruntime.NumCPU())
var clangActive atomic.Int64
var clangPeak atomic.Int64

type pooledClang struct{ *exec.Cmd }

func clangCommand(name string, args ...string) *pooledClang {
	return &pooledClang{exec.Command(name, args...)}
}

func (c *pooledClang) acquire() func() {
	clangTokens <- struct{}{}
	active := clangActive.Add(1)
	for peak := clangPeak.Load(); active > peak; peak = clangPeak.Load() {
		if clangPeak.CompareAndSwap(peak, active) {
			break
		}
	}
	return func() { clangActive.Add(-1); <-clangTokens }
}

func (c *pooledClang) Output() ([]byte, error) {
	release := c.acquire()
	defer release()
	return c.Cmd.Output()
}

func (c *pooledClang) CombinedOutput() ([]byte, error) {
	release := c.acquire()
	defer release()
	return c.Cmd.CombinedOutput()
}
