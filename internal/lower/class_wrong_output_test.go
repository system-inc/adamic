package lower

import (
	"os"
	"strings"
	"testing"
)

func TestClassWrongOutputIteratorReceiver(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/iterators_override_this.a")
	if err != nil {
		t.Fatal(err)
	}
	prefix := strings.Split(string(source), "for(const value")[0]
	for _, use := range []string{
		"for (const value of new ScaledIterator()) { console.log(`${value}`); break; }",
		"console.log([...new ScaledIterator()].join(','));",
		"const viewed: BaseIterator = new ScaledIterator(); console.log(Array.from(viewed).join(','));",
		"const [first = -1] = new ScaledIterator(); console.log(`${first}`);",
	} {
		_, err := lowerSource(t, prefix+use)
		if err == nil || !strings.Contains(err.Error(), "adamic/iterator-receiver-origin") {
			t.Fatalf("unsafe protocol accepted: %s: %v", use, err)
		}
	}
	for _, use := range []string{
		"console.log([...new BaseIterator()].join(','));",
		"class Unchanged extends BaseIterator {} console.log([...new Unchanged()].join(','));",
	} {
		if _, err := lowerSource(t, prefix+use); err != nil {
			t.Fatalf("unchanged protocol refused: %v", err)
		}
	}
}

func TestClassWrongOutputKeysRepair(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/iterators_sym_keys_view.a")
	if err != nil {
		t.Fatal(err)
	}
	repaired := strings.Replace(string(source), "Object.keys(view)", "Object.keys({ value: view.value })", 1)
	if _, err := lowerSource(t, repaired); err != nil {
		t.Fatalf("explicit string-key copy refused: %v", err)
	}
}
