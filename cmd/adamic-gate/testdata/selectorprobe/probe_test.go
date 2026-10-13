package selectorprobe

import "testing"

func TestParent(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"internal/oracle/testdata/a.a", "internal/oracle/testdata/aXa", "internal/oracle/testdata/b+.a", "internal/load/testdata/0.1/compile/main.a", "internal/load/testdata/0X1/compile/extra.a", "internal/load/testdata/0X1/compile/main.a"} {
		t.Run(name, func(t *testing.T) { t.Parallel() })
	}
}
func TestParentExtra(t *testing.T) {}
func TestOther(t *testing.T)       {}
