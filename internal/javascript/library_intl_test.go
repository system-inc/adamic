package javascript

import (
	"os/exec"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The collator stand-in is Node's collator in shape (#zjndrx5 head 1-4, Runtime's ruling at 16:11Z): no
// own property, so Object.keys, getOwnPropertyNames and hasOwn see nothing, compare is a prototype
// getter, so 'compare' in it is true and a write throws a TypeError in a module's strict mode, and a
// call of compare stops in ir's words. Each line is what Node's own collator answers, the last aside.
// Mutant: compare made an own data property in the constructor, so keys and hasOwn see it and the
// write goes through.
func TestIntlCollatorStandInHasNoOwnProperty(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("missing Node: %v", err)
	}
	source := "const panic = message => { console.log('panic: ' + message); };\n" + intlCollatorRuntime() +
		"const collator = adamicIntlCollator(undefined);\n" +
		"console.log(JSON.stringify(Object.keys(collator)));\n" +
		"console.log(Object.getOwnPropertyNames(collator).length);\n" +
		"console.log(Object.hasOwn(collator, 'compare'));\n" +
		"console.log('compare' in collator);\n" +
		"try { collator.compare = () => 0; console.log('wrote'); } catch (error) { console.log(error instanceof TypeError ? 'TypeError' : 'other'); }\n" +
		"collator.compare(collator, 'a', 'b');\n"
	result, err := exec.Command("node", "--input-type=module", "-e", source).CombinedOutput()
	want := "[]\n0\nfalse\ntrue\nTypeError\npanic: " + ir.IntlCollatorCompareStop + "\n"
	if err != nil || string(result) != want {
		t.Fatalf("got %v\n%s\nwant\n%s", err, result, want)
	}
}
