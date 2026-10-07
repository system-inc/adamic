package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"uint8array", "int32array", "float64array", "views", "order", "primes", "primes_large", "stats", "stop"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/typed_arrays_" + name + ".a", true, name == "stop",
		})
	}
}

// This intentional difference has independent pins for both backends and Node.
// Removing the check in both backends must still fail, even though they agree.
func TestTypedArrayWriteStopIsPinned(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_stop.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	stopped := run{stdout: []byte("before 2 7 undefined\n"), stderr: []byte("adamic: panic: index 2 is outside an array of length 2\n"), exitCode: 70}
	node := run{stdout: []byte("before 2 7 undefined\nafter 2 7 undefined\n")}
	native, _ := natively(t, p)
	for name, observed := range map[string]run{"native": native, "release": released(t, p), "JavaScript": onJavaScriptBackend(t, p)} {
		if difference := disagreement(stopped, observed); difference != "" {
			t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, observed.exitCode, observed.stdout, observed.stderr)
		}
	}
	if observed := onNode(t, path); disagreement(node, observed) != "" {
		t.Errorf("Node no longer silently drops the write: %#v", observed)
	}
}
