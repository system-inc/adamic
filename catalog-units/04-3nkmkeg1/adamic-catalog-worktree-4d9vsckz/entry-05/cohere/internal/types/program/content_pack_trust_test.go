package program

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A missed run's check stats every input, and the content pack validates against those stats rather than taking
// each again (#kdee854). That is sound only through the run-cache clock, so this follows the whole chain: the
// check notes a file's stat; the file changes; a pack trusting the check serves the bytes the check saw, where a
// pack statting afresh refuses them; and recording that run is refused, since the change postdates the run's
// clock, so the next run reads the new bytes and nothing replays the old verdict.
func TestAPackTrustsTheCheckStatsAndTheRunClockRefusesAChangeAfterThem(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	cache := filepath.Join(root, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "a.ts")
	if err := os.WriteFile(file, []byte("export const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A previous run read the file and kept it in the pack, and recorded it as an input.
	identity, statted := statIdentity(file)
	if !statted {
		t.Skip("this platform gives the content pack no identity, so it serves nothing")
	}
	recording, _ := OpenContentPack(cache)
	recording.noteRead(file, identity, true, "export const a = 1;\n")
	if err := recording.Save(); err != nil {
		t.Fatal(err)
	}
	recorded, err := RecordRunCache("key", []string{file}, nil, nil, nil, []byte("verdict"), 0, time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	// This run: the clock starts, then the check stats every input and notes them.
	readSince := time.Now()
	snapshot := NewStatSnapshot()
	if err := recorded.CheckNoting("key", snapshot); err != nil {
		t.Fatalf("an unchanged input missed: %v", err)
	}
	if _, noted := snapshot.identityOf(file); !noted {
		t.Fatal("the check noted no stat for the input, so trusting it proves nothing")
	}

	// The file changes after the check, at the same size.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(file, []byte("export const a = 2;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	trusting, _ := OpenContentPack(cache)
	trusting.TrustStats(snapshot)
	if served, ok := trusting.serve(file); !ok || served != "export const a = 1;\n" {
		t.Errorf("a pack trusting the check served %q (%v), want the bytes the check saw", served, ok)
	}
	statting, _ := OpenContentPack(cache)
	if served, ok := statting.serve(file); ok {
		t.Errorf("a pack statting afresh served %q, so the snapshot was not what decided above", served)
	}

	if _, err := RecordRunCache("key", []string{file}, nil, nil, nil, []byte("verdict"), 0, readSince); err == nil ||
		!strings.Contains(err.Error(), "changed after the run began reading") {
		t.Errorf("a run that read the file before its change was recorded anyway (%v), so the next run could replay it", err)
	}
}
