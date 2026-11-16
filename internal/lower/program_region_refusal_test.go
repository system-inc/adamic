package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func programCycleRefusalParity(t *testing.T, source string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	var baseline *Refused
	for _, enabled := range []bool{false, true} {
		loaded, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		_, err = LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: enabled})
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("ProgramRegion=%t: got %v, want Refused", enabled, err)
		}
		if !enabled {
			baseline = refused
		} else if refused.What != baseline.What || refused.Fix != baseline.Fix {
			t.Fatalf("Program changed refusal: got %#v, want %#v", refused, baseline)
		}
	}
}

// A record cycle lowers in both modes since records joined graph regions (#1vb0x5s): a graph record is
// never a Program member (ir.RecordLiteral has no ProgramRegion stamp), so the Program region changes
// nothing about it. Oracle's TestGraphRecordOnACycle holds both programs to Node and the counted build.
func TestProgramRecordCycleLowersInBothModes(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`interface R {[key:string]:R} const r:R={}; r['self']=r;`,
		`interface Node {links:Record<string,Node>} const a:Node={links:{}}; a.links['self']=a;`,
	} {
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		for _, enabled := range []bool{false, true} {
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: enabled}); err != nil {
				t.Fatalf("ProgramRegion=%t: %v", enabled, err)
			}
		}
	}
}

func TestProgramPromiseCycleRefusalMatchesDefault(t *testing.T) {
	t.Parallel()
	programCycleRefusalParity(t, `interface Box { promise: Promise<Box> | undefined; }
function stash(box: Box, promise: Promise<Box>): void { box.promise = promise; }
const box: Box = {promise: undefined};
const promise = Promise.resolve(box);
stash(box,promise);`)
}
