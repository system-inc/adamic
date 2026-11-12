package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// graphCheckedViewFixtures are the graph programs holding a checked view, each with its runtime count record
// (testdata/counts/runtime/internal/oracle/testdata/graph_checked_views/).
var graphCheckedViewFixtures = []string{
	"internal/oracle/testdata/graph_checked_views/read.ts",
	"internal/oracle/testdata/graph_checked_views/stored.ts",
	"internal/oracle/testdata/graph_checked_views/tagged.ts",
}

func init() {
	for _, path := range graphCheckedViewFixtures {
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
	}
}

// A graph program holds checked views since slice 3 of #ztpcwrr (#ef727kr): a view owns nothing, so a
// read through it retains as any read does, and a write through it takes writeProperty's graph hold.
// Each fixture puts the viewed objects on a cycle 1,000 times in a function, so every turn's cycle is
// a graph region that must free whole, and holds the answer to Node and the counted build. They're .ts,
// where a downcast stays a checked view (in .a it's refused, #m8dvvmf). Writing a field of a type a
// view checks (Node and Identifier here) is NotYet without a source-slot certificate, and a legacy
// view's field can't hold a reference yet, so each cycle closes through a holder no view checks. The
// write is a view stored into the holder's graph slot, which the graph hold takes as any store.
//
// Mutant: graphAdoptionBoundary's refusal of ViewOrigins and CheckedFields put back stops every
// fixture before it lowers.
func TestGraphCheckedViews(t *testing.T) {
	t.Parallel()
	for _, fixture := range graphCheckedViewFixtures {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatal("not a graph program, so it pins nothing about one")
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, counted)
			if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
				t.Fatal(report)
			}
		})
	}
}
