package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/fresh_map_returns.a", true, false})
}

func TestFreshMapSharedReturnsRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"shared", "mixed", "stored"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/fresh_map_"+name+"_refused.a"))
			if err != nil {
				t.Fatal(err)
			}
			observation := onNode(t, path)
			if observation.exitCode != 0 || string(observation.stdout) != "1\n" || len(observation.stderr) != 0 {
				t.Fatalf("Node: %+v", observation)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "a never[] seen as writable T[][], which can write T into a shared never[]" || refused.Fix != "declare the result readonly T[], or return a fresh [] (adamic/invariant-mutable)" {
				t.Fatalf("want pinned shared-array refusal, got %v", err)
			}
			where := ":2:47"
			if name == "stored" {
				where = ":3:12"
			}
			if !strings.HasSuffix(refused.Where, where) {
				t.Fatalf("wrong refusal site: %v", refused)
			}
			t.Log(err)
		})
	}
}

func TestFreshMapMixedReturnRefused(t *testing.T) {
	// The every-path mutant must lose this refusal, independently of unrelated sites.
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/fresh_map_mixed_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "shared never[]") {
		t.Fatalf("want mixed callback refused, got %v", err)
	}
}
