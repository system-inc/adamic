package program

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

// The same tree records the same inputs whatever order the parallel loaders ask in (#zc57tqg). On excalidraw,
// which builds asked about a package's directory, and whether with a trailing slash, depended on which loader
// resolved the package first, so the record varied between builds of one tree. Here every order of one set of
// questions must give one answer: a probe with a present path beneath it is left out, a probe with none is
// kept, and the two spellings of a directory are one path.
func TestTheRecordIsTheSameWhateverOrderTheQuestionsCameIn(t *testing.T) {
	t.Parallel()
	type question struct {
		path              string
		present, depended bool
	}
	questions := []question{
		{"/project/node_modules/react-dom/", true, false},
		{"/project/node_modules/react-dom", true, false},
		{"/project/node_modules/react-dom/package.json", true, true},
		{"/project/node_modules/@firebase/app/", true, false},
		{"/project/node_modules/@firebase/app/dist/index.d.ts", true, true},
		{"/project/node_modules/lonely/", true, false},
		{"/project/node_modules/missing", false, false},
		{"/project/source/index.ts", true, true},
	}
	want := func() [3][]string {
		recorder := NewInputRecorder()
		for _, asked := range questions {
			recorder.note(asked.path, asked.present, asked.depended)
		}
		present, absent, probed := recorder.Inputs()
		return [3][]string{present, absent, probed}
	}()
	if !reflect.DeepEqual(want[2], []string{"/project/node_modules/lonely"}) {
		t.Fatalf("probed %v, want only the probe with nothing recorded beneath it, without its slash", want[2])
	}
	for _, path := range want[0] {
		if path == "/project/node_modules/react-dom" || path == "/project/node_modules/@firebase/app" {
			t.Errorf("%s has a recorded path beneath it and was recorded too", path)
		}
	}

	random := rand.New(rand.NewPCG(1, 2))
	for range 200 {
		shuffled := append([]question{}, questions...)
		random.Shuffle(len(shuffled), func(left, right int) { shuffled[left], shuffled[right] = shuffled[right], shuffled[left] })
		// A loader that resolved from its cache asks only some of the questions: drop the redundant probes at
		// random, as a build whose loaders raced differently would.
		recorder := NewInputRecorder()
		for _, asked := range shuffled {
			if !asked.depended && asked.present && asked.path != "/project/node_modules/lonely/" && random.IntN(2) == 0 {
				continue
			}
			recorder.note(asked.path, asked.present, asked.depended)
		}
		present, absent, probed := recorder.Inputs()
		if got := [3][]string{present, absent, probed}; !reflect.DeepEqual(got, want) {
			t.Fatalf("an order the loaders could have asked in recorded\n  %v\nwhere another recorded\n  %v", got, want)
		}
	}
}
