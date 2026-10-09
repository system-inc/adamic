package program_test

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// A retired graph records nothing into the types section, by either of the two paths that record (#sm79kfv).
//
// A graph is retired when the fix phase rewrote files and a successor was built from the new bytes. Its
// check is left to finish rather than cancelled, because cancelling the compiler's whole-program check part
// way panics, so the check can end after the successor's has recorded. What it found describes bytes that no
// longer exist, and recording it last would leave the successor's record overwritten. Each path is paired
// with the same call on a live graph, which must record, so a pass here cannot be a path that never records.
func TestARetiredGraphRecordsNothing(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"main.ts":       "import { helper } from './helper';\nexport const value: number = helper();\n",
		"helper.ts":     "export function helper(): number { return 1; }\n",
	})
	ctx := context.Background()
	build := func() *program.Graph {
		t.Helper()
		graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
		if err != nil {
			t.Fatalf("building: %v", err)
		}
		graph.Shapes, _ = graph.Signatures(ctx, program.RecordedRealSignatures(graph))
		return graph
	}
	key := sha256.Sum256([]byte("compiler options"))

	// RecordFull, which a full check records through.
	live := build()
	liveReuse := program.NewTypeDiagnosticsReuse(nil, key)
	live.RecordFull(ctx, liveReuse, live.AllDiagnosticParts(ctx))
	recorded := liveReuse.Recorded()
	if recorded == nil {
		t.Fatal("control: a live graph's full check recorded nothing, so the retired case below proves nothing")
	}
	liveReuse.FinishGlobals(true)

	retired := build()
	retired.Retire()
	retiredReuse := program.NewTypeDiagnosticsReuse(nil, key)
	retired.RecordFull(ctx, retiredReuse, retired.AllDiagnosticParts(ctx))
	if retiredReuse.Recorded() != nil {
		t.Error("RecordFull recorded a retired graph's check")
	}

	// CheckReusing, which a check that replays records through, from the section the live graph left.
	replaying := build()
	replayReuse := program.NewTypeDiagnosticsReuse(recorded, key)
	if _, reused := replaying.CheckReusing(ctx, replayReuse); !reused {
		t.Fatal("control: a live graph did not replay the section it just recorded")
	}
	if replayReuse.Recorded() == nil {
		t.Fatal("control: a live graph's replay recorded nothing, so the retired case below proves nothing")
	}

	retiredReplaying := build()
	retiredReplaying.Retire()
	retiredReplayReuse := program.NewTypeDiagnosticsReuse(recorded, key)
	if _, reused := retiredReplaying.CheckReusing(ctx, retiredReplayReuse); !reused {
		t.Fatal("a retired graph did not replay, so whether it records was never asked")
	}
	if retiredReplayReuse.Recorded() != nil {
		t.Error("CheckReusing recorded a retired graph's check")
	}
}
