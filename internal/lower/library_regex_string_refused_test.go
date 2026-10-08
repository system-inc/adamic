package lower

import (
	"strings"
	"testing"
)

func TestRegexStringBoxBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function escape(): String { return new String('a'); }`,
		`const box = new String('a'); const alias = box; console.log(typeof alias);`,
		`const box = new String('a'); const hidden = {box}; console.log(typeof hidden);`,
		`let box = new String('a'); box = new String('b'); box.search(/a/);`,
		`const box = new String('a'); Object.getOwnPropertyNames(box);`,
		`const box = new String('a'); console.log(box[0]);`,
		"const box = new String('a'); console.log(`${box.length}`);",
		`const box = new String('a'); box.search('[');`,
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatalf("must refuse incomplete String box semantics: %s", source)
		}
	}
	for _, source := range []string{
		`const search = String.prototype.search; search.call('a', /a/);`,
		`String.prototype.search.prototype = {};`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "method read as a value") {
			t.Fatalf("must keep detach/mutation refused: %s: %v", source, err)
		}
	}
}
