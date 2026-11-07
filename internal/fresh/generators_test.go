package fresh_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A nested generator's call is a CallClosure through its carrier, direct to the generator (#qk8rztp
// S3a), and a step hands back what its frame holds: its arguments, and through its carrier every
// cell of its parent the carrier holds (Memory, on fresh.go's GeneratorNext). The CallClosure arm
// holds both, since it lets every operand escape, the carrier and the arguments alike, and gives an
// outside result. The carrier is outside whatever made it (MakeClosure's arm, ClosureSelf's
// library-language arm, and a parent's binding, which a MakeClosure made), and every captured cell's
// value has escaped, so what the frame reaches through it is never confined; the arguments are what
// an arm that dropped its operands would lose. Here a confined Box is handed to the generator, and
// each step's value is written into it: a step may give back the Box itself, so that write must stay
// unproven. The control is a write into the same Box before the call, of a value made right there,
// which is proven. Mutant: the CallClosure arm gives its result without letting its operands escape;
// the step's write is proven, and this goes red.
func TestNestedGeneratorCallHoldsWhatItsFrameHolds(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `class Tag {
    name: string = 'tag';
}
class Box {
    first: Tag | undefined = undefined;
    stepped: Tag | undefined = undefined;
}
function host(label: string): string {
    function* tags(box: Box): Generator<Tag, void, undefined> {
        const made = new Tag();
        made.name = label + String(box.first === undefined);
        yield made;
    }
    const box = new Box();
    box.first = new Tag();
    for (const tag of tags(box)) {
        box.stepped = tag;
    }
    const last = box.stepped;
    return last === undefined ? 'none' : last.name;
}
console.log(host('t'));
`
	if writeError := os.WriteFile(path, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatalf("the program must lower: %v", lowerError)
	}
	host := -1
	for index, function := range program.Functions {
		if function.Name == "host" {
			host = index
		}
	}
	if host < 0 {
		test.Fatal("no function host")
	}
	found := map[string]fresh.Write{}
	for _, write := range fresh.ProveWrites(program) {
		if write.Kind == fresh.WriteField && write.Function == host {
			found[write.Name] = write
		}
	}
	first, isFirst := found["first"]
	stepped, isStepped := found["stepped"]
	if !isFirst || !isStepped {
		test.Fatalf("host's writes %+v, want first and stepped", found)
	}
	if !first.Proven {
		test.Errorf("the control, a fresh Tag written into a confined Box, isn't proven: %+v", first)
	}
	if stepped.Proven || stepped.Why == "" {
		test.Errorf("a step's value written into the Box handed to the generator is proven: %+v", stepped)
	}
}
