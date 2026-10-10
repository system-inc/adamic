package load

import (
	"path/filepath"
	"testing"
)

func TestCollectionIteratorElementInference(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"iterable", "callback", "entries", "set-copy", "nested-entries"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "overlay-iterators", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Load([]string{path}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCollectionIteratorDoneDiscriminatesReturn(t *testing.T) {
	t.Parallel()
	source := `const map = new Map<string, string>().values().next();
if (!map.done) { const yielded: string = map.value; }
else { const completed: undefined = map.value; }
const set = new Set<string>().values().next();
if (!set.done) { const yielded: string = set.value; }
else { const completed: undefined = set.value; }
`
	if _, err := Load(writeProgram(t, [2]string{"main.a", source})); err != nil {
		t.Fatal(err)
	}
	checkErrors(t, writeProgram(t, [2]string{"main.a", `const value: string = new Set<string>().values().next().value;`}))
}
