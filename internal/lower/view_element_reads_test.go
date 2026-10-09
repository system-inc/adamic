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

func TestViewStringElementReadRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, viewElementSource)
	if err == nil || !strings.Contains(err.Error(), "view member read by element access is not yet checked") || !strings.Contains(err.Error(), "use view.count") || !strings.Contains(err.Error(), ".a:") {
		t.Fatalf("want located checked-view refusal and fix, got %v", err)
	}
}

func TestViewLiteralTypedElementReadRefused(t *testing.T) {
	t.Parallel()
	source := strings.Replace(viewElementSource, "return view['count'];", "const key = 'count'; return view[key];", 1)
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "view member read by element access is not yet checked") {
		t.Fatalf("unchecked literal typed key: %v", err)
	}
}

func TestViewElementFixAgreesWithNode(t *testing.T) {
	t.Parallel()
	source := strings.Replace(viewElementSource, "view['count']", "view.count", 1)
	source = strings.Replace(source, "count: 'wrong'", "count: 2", 1)
	lowersAndAgreesWithNode(t, source)
}

func TestViewElementP05Refused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, viewElementFixture(t, "p05"))
	requireViewElementRefusal(t, err)
}

func TestViewElementP06Refused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, viewElementFixture(t, "p06"))
	requireViewElementRefusal(t, err)
}

func TestViewElementP01Refused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, viewElementFixture(t, "p01"))
	requireViewElementRefusal(t, err)
}

func TestViewElementP07Refused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, viewElementFixture(t, "p07"))
	requireViewElementRefusal(t, err)
}

func TestViewElementP05CorrectControl(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(viewElementFixture(t, "p05"), "view['count']", "view.count")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: 'seven'", "count: 7", 1))
}

func TestViewElementP06CorrectControl(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(viewElementFixture(t, "p06"), "view['count']", "view.count")
	lowersAndAgreesWithNode(t, strings.Replace(source, "count: true", "count: 1", 1))
}

func TestViewElementP01CorrectControl(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(viewElementFixture(t, "p01"), "view['text']", "view.text")
	lowersAndAgreesWithNode(t, strings.Replace(source, "text: 7", "text: 'seven'", 1))
}

func TestViewElementP07CorrectControl(t *testing.T) {
	t.Parallel()
	source := strings.ReplaceAll(viewElementFixture(t, "p07"), "view['value']", "view.value")
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

func requireViewElementRefusal(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "view member read by element access is not yet checked") || !strings.Contains(err.Error(), "use view.") {
		t.Fatalf("unchecked view element read: %v", err)
	}
}
