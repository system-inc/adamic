package lower

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const viewElementSource = `interface Base { readonly kind: 'box' | 'other' }
interface Box extends Base { readonly kind: 'box'; readonly count: number }
function read(base: Base): number { const view = base as Box; return view['count']; }
const raw = {kind: 'box' as const, count: 'wrong'}; console.log(String(read(raw)));`

func TestViewStringElementReadControl(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, strings.Replace(viewElementSource, "count: 'wrong'", "count: 2", 1))
}

func TestViewLiteralTypedElementReadControl(t *testing.T) {
	t.Parallel()
	source := strings.Replace(viewElementSource, "return view['count'];", "const key = 'count'; return view[key];", 1)
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: 'wrong'", "count: 2", 1))
}

func TestViewElementFixAgreesWithNode(t *testing.T) {
	t.Parallel()
	source := strings.Replace(viewElementSource, "view['count']", "view.count", 1)
	source = strings.Replace(source, "count: 'wrong'", "count: 2", 1)
	lowersAndAgreesWithNode(t, source)
}

func TestViewElementP05CorrectControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "p05")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: 'seven'", "count: 7", 1))
}

func TestViewElementP06CorrectControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "p06")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: true", "count: 1", 1))
}

func TestViewElementP01CorrectControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "p01")
	lowersAndAgreesWithNode(t, strings.Replace(source, "text: 7", "text: 'seven'", 1))
}

func TestViewElementP07CorrectControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "p07")
	lowersAndAgreesWithNode(t, strings.Replace(source, "text: true", "text: 'seven'", 1))
}

func viewElementFixture(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "view_element_reads", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestViewElementDestructuredControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "destructured")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: 'seven'", "count: 7", 1))
}

func TestViewElementDestructuredUnionControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "destructured_union")
	lowersAndAgreesWithNode(t, strings.Replace(source, "text: true", "text: 'seven'", 1))
}

func TestViewSpreadControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "spread")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: 'seven'", "count: 7", 1))
}

func TestViewSpreadMethodRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Root {readonly kind:1} interface View extends Root {read():number} const raw={kind:1 as const, read(){return 1}}; const base:Root=raw; const view=base as View; const copy={...view}; console.log(Object.keys(copy).join(','));`)
	if err == nil || !strings.Contains(err.Error(), "method without an own data slot") {
		t.Fatalf("want checked method spread refused, got %v", err)
	}
}

func TestViewInControl(t *testing.T) {
	t.Parallel()
	source := viewElementFixture(t, "in")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: 'seven'", "count: 7", 1))
}
