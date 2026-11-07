package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func programCycleRefusalParity(test *testing.T, source string) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	var baseline *Refused
	for _, enabled := range []bool{false, true} {
		loaded, loadError := load.Load([]string{path})
		if loadError != nil {
			test.Fatal(loadError)
		}
		_, lowerWithOptionsError := LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: enabled})
		var refused *Refused
		if !errors.As(lowerWithOptionsError, &refused) {
			test.Fatalf("ProgramRegion=%t: got %v, want Refused", enabled, lowerWithOptionsError)
		}
		if !enabled {
			baseline = refused
		} else if refused.What != baseline.What || refused.Fix != baseline.Fix {
			test.Fatalf("Program changed refusal: got %#v, want %#v", refused, baseline)
		}
	}
}

func TestProgramRecordCycleRefusalMatchesDefault(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`interface R {[key:string]:R} const r:R={}; r['self']=r;`,
		`interface Node {links:Record<string,Node>} const a:Node={links:{}}; a.links['self']=a;`,
	} {
		programCycleRefusalParity(test, source)
	}
}

func TestProgramPromiseCycleRefusalMatchesDefault(test *testing.T) {
	test.Parallel()
	programCycleRefusalParity(test, `interface Box { promise: Promise<Box> | undefined; }
function stash(box: Box, promise: Promise<Box>): void { box.promise = promise; }
const box: Box = {promise: undefined};
const promise = Promise.resolve(box);
stash(box,promise);`)
}
