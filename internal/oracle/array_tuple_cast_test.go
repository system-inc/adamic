package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An array cast to a tuple type passes castProof as an interface view, its length the tuple's tag,
// tested on an object placeholder. The array's own value isn't an object, so interfaceCast gives no
// checked cast back; before the guard in cast that nil went on as the cast's value and lowering
// panicked reading its type (tsc's moduleSpecifiers.ts, `return emptyArray as [];`). Each is a NotYet
// at the cast. Mutant: dropping the guard in cast panics both again.
func TestArrayToTupleCastsAreNotYet(test *testing.T) {
	test.Parallel()
	for _, program := range []struct{ name, stdout, want string }{
		{
			"empty.ts",
			"0\n",
			":2:15: stage 0 can't lower a checked cast from never[] to [] whose value isn't a plain object yet",
		},
		{
			"pair.ts",
			"2\n",
			":2:14: stage 0 can't lower a checked cast from number[] to [number] whose value isn't a plain object yet",
		},
	} {
		test.Run(program.name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/array_tuple_casts", program.name),
			)
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != program.stdout {
				test.Fatalf("Node: %+v", truth)
			}
			_, err = lowered(test, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || err.Error() != path+program.want {
				test.Fatalf("got %v; want %s", err, path+program.want)
			}
		})
	}
}
