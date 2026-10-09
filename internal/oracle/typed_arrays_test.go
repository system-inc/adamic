package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"uint8array", "uint16array", "uint16_stop", "int32array", "float64array", "views", "order", "primes", "primes_large", "stats", "workers_stats", "stop"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/typed_arrays_" + name + ".a", true, name == "stop" || name == "uint16_stop",
		})
	}
}

// This intentional difference has independent pins for both backends and Node.
// Removing the check in both backends must still fail, even though they agree.
func TestTypedArrayWriteStopIsPinned(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"stop", "uint16_stop"} {
		t.Run(kind, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_"+kind+".a"))
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
		})
	}
}

// Original platform functions run on Node independently of the typed-array port.
// This pins the port's arithmetic, not just native agreement with its own source.
func TestTypedArrayWorkersStatsMatchesPlatforms(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_workers_stats.a"))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_arrays_workers_stats_reference.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, reference)
	if expected.exitCode != 0 {
		t.Fatalf("platforms Node reference: %#v", expected)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := natively(t, p)
	for name, observed := range map[string]run{"source": onNode(t, path), "native": native, "release": released(t, p), "JavaScript": onJavaScriptBackend(t, p)} {
		if difference := disagreement(expected, observed); difference != "" {
			t.Errorf("%s: %s: exit %d stdout %q stderr %q", name, difference, observed.exitCode, observed.stdout, observed.stderr)
		}
	}
}
