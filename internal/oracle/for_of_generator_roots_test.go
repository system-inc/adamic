package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var generatorRootFixtures = []struct{ name, stdout string }{
	{"map", "2\n3\n"},
	{"flat_map", "1\n2\n2\n3\n"},
	{"defined", "2\n"},
}

// Latent replay enters generator bodies, but ordinary .a lowering must retain
// the suspended-frame refusal until ownership and cancellation are designed.
func TestForOfGeneratorRootsRemainRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range generatorRootFixtures {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_of_generator_roots", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.stdout || len(observed.stderr) != 0 {
				t.Fatalf("Node: %+v; want stdout %q", observed, probe.stdout)
			}
			program, err := lowered(t, path)
			var refused *lower.Refused
			if program != nil || !errors.As(err, &refused) || refused.What != "a generator function" {
				t.Fatalf("want generator ownership refusal and no IR, got program %v, error %v", program != nil, err)
			}
		})
	}
}
