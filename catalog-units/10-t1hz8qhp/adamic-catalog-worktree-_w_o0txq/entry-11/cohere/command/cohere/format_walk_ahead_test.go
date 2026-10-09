package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/native"
)

// countingEngine is a format engine whose walks are counted, so a test can tell a walk taken from ahead from
// one made on the spot.
type countingEngine struct {
	*native.Resolving
	walks int
}

func (engine *countingEngine) Enumerate(root string) (formatfiles.Enumeration, error) {
	engine.walks++
	return engine.Resolving.Enumerate(root)
}

// Not parallel: it sets aheadFormatWalk, the package-level walk the fix phase takes, which a run in this
// package beside it could take or replace.
//
// The format walk begun beside the graph build is the walk the fix phase would have made (#679s763): taken
// for the same engine and root, it equals a walk made on the spot, and it is taken once. A walk of another
// root, or by another engine, is made on the spot.
func TestTheFormatWalkBegunAheadIsTheOneTheFixPhaseWouldMake(t *testing.T) {
	root := t.TempDir()
	for name, contents := range map[string]string{
		"a.ts":          "export const a = 1;\n",
		"notes.md":      "# Notes\n",
		"ignored/b.ts":  "export const b = 2;\n",
		".gitignore":    "ignored/\n",
		"styles/c.css":  "a { color: red; }\n",
		"data/d.json":   "{}\n",
		"other/e.ts":    "export const e = 3;\n",
		"other/f.yaml":  "f: 1\n",
		"other/g.txt":   "not formatted\n",
		"nested/h.tsx":  "export const h = <div />;\n",
		"nested/i.scss": "a { b: c; }\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	resolving, err := native.NewResolving(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { aheadFormatWalk = nil })

	direct, err := resolving.Enumerate(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(direct.Files) < 5 {
		t.Fatalf("the fixture's walk found %d files, too few for a match to prove anything", len(direct.Files))
	}

	engine := &countingEngine{Resolving: resolving}
	startFormatWalkAhead(engine, root, nil)
	taken, err := enumerateFormatTree(engine, root)
	if err != nil {
		t.Fatal(err)
	}
	if engine.walks != 1 {
		t.Fatalf("the engine walked %d times for one walk begun ahead and taken", engine.walks)
	}
	if !reflect.DeepEqual(taken, direct) {
		t.Fatalf("the walk begun ahead differs from a direct walk\nahead:  %+v\ndirect: %+v", taken, direct)
	}
	if _, err := enumerateFormatTree(engine, root); err != nil || engine.walks != 2 {
		t.Errorf("a second call took the same walk again (%d walks)", engine.walks)
	}

	startFormatWalkAhead(engine, root, nil)
	// Waited for, so the counts below never race the walk's goroutine.
	<-aheadFormatWalk.done
	other := &countingEngine{Resolving: resolving}
	if _, err := enumerateFormatTree(other, root); err != nil || other.walks != 1 {
		t.Errorf("another engine took the walk begun for this one (%d walks of its own)", other.walks)
	}
	if _, err := enumerateFormatTree(engine, filepath.Join(root, "other")); err != nil || engine.walks != 4 {
		t.Errorf("a walk of another root was taken from the walk begun ahead (%d walks)", engine.walks)
	}
}
