package rule

import (
	"sort"
	"sync"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
)

// designSystemRecordings keeps every recording file system DesignSystemFS handed out, by the program it was
// handed out for, so the findings cache can learn what the design system read without importing the reader
// (#35nqkwc).
//
// Keyed per program rather than in one slot, because "not loaded" has to be evidence. A slot that another
// program could take would answer "not loaded" for a program whose rules did read the stylesheets, and the
// cache would record their findings as independent of every stylesheet: replayed over any edit to one.
// Production walks one program at a time, but parallel tests do not, and neither would a future concurrent
// graph (@system_cohere_lint, reviewing 5d0cc66).
//
// Weakly, so a program nothing else holds is not kept alive by its entry here; entries for collected programs
// are swept on the next registration.
var designSystemRecordings struct {
	sync.Mutex
	byProgram map[weak.Pointer[compiler.Program]][]*RecordingFS
}

func recordDesignSystemFS(program *compiler.Program, recorder *RecordingFS) {
	designSystemRecordings.Lock()
	defer designSystemRecordings.Unlock()
	if designSystemRecordings.byProgram == nil {
		designSystemRecordings.byProgram = map[weak.Pointer[compiler.Program]][]*RecordingFS{}
	}
	for key := range designSystemRecordings.byProgram {
		if key.Value() == nil {
			delete(designSystemRecordings.byProgram, key)
		}
	}
	key := weak.Make(program)
	designSystemRecordings.byProgram[key] = append(designSystemRecordings.byProgram[key], recorder)
}

// DesignSystemReads is everything the design system read for program, and whether it was loaded at all. False
// means no rule asked for it on this program, so nothing they reported can depend on a stylesheet.
//
// The reads of every load are merged, a path seen present in any keeping present, so a second load that
// asked about more than the first widens the set rather than replacing it.
func DesignSystemReads(program *compiler.Program) ([]FileRead, bool) {
	designSystemRecordings.Lock()
	defer designSystemRecordings.Unlock()
	recorders := designSystemRecordings.byProgram[weak.Make(program)]
	if len(recorders) == 0 {
		return nil, false
	}
	merged := map[string]FileRead{}
	for _, recorder := range recorders {
		for _, read := range recorder.Reads() {
			if earlier, seen := merged[read.Path]; seen && (earlier.Present || !read.Present) {
				continue
			}
			merged[read.Path] = read
		}
	}
	reads := make([]FileRead, 0, len(merged))
	for _, read := range merged {
		reads = append(reads, read)
	}
	sort.Slice(reads, func(first, second int) bool { return reads[first].Path < reads[second].Path })
	return reads, true
}
