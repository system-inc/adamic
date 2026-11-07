package lower

import (
	"os"
	"strings"
	"testing"
)

func TestClassWrongOutputIteratorReceiver(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../oracle/testdata/class_wrong_output_refused/iterators_override_this.a")
	if readError != nil {
		test.Fatal(readError)
	}
	prefix := strings.Split(string(source), "for(const value")[0]
	for _, use := range []string{
		"for (const value of new ScaledIterator()) { console.log(`${value}`); break; }",
		"console.log([...new ScaledIterator()].join(','));",
		"const viewed: BaseIterator = new ScaledIterator(); console.log(Array.from(viewed).join(','));",
		"const [first = -1] = new ScaledIterator(); console.log(`${first}`);",
	} {
		_, lowerSourceError := lowerSource(test, prefix+use)
		if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), "adamic/iterator-receiver-origin") {
			test.Fatalf("unsafe protocol accepted: %s: %v", use, lowerSourceError)
		}
	}
	for _, use := range []string{
		"console.log([...new BaseIterator()].join(','));",
		"class Unchanged extends BaseIterator {} console.log([...new Unchanged()].join(','));",
	} {
		lowersAndAgreesWithNode(test, prefix+use)
	}
}

func TestClassWrongOutputKeysRepair(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../oracle/testdata/class_wrong_output_refused/iterators_sym_keys_view.a")
	if readError != nil {
		test.Fatal(readError)
	}
	repaired := strings.Replace(string(source), "Object.keys(view)", "Object.keys({ value: view.value })", 1)
	lowersAndAgreesWithNode(test, repaired)
}

func TestClassWrongOutputPrivateRepair(test *testing.T) {
	test.Parallel()
	source := "class Box { #secret: string; constructor(secret: string) { this.#secret = secret; } reveal(): string { return this.#secret; } static peek(box: Box): string { return box.reveal(); } } console.log(Box.peek(new Box(`s${1}`)));"
	lowersAndAgreesWithNode(test, source)
}
