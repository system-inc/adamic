package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A computed write to a fixed object (finiteElementWrite) is an if-chain over the key's union: each test
// of the held key stores exactly its own field, once, and past the last test the key is outside the union
// the checker proved, which panics loudly rather than writing nothing. No fixture reaches that panic (the
// checker proved it can't be reached), so only the lowered shape shows it's there: finite_keyed_write's
// mutant (d) drops it, and this test is the one that goes red.
func TestFiniteElementWriteStoresEachNamedFieldThenPanics(t *testing.T) {
	t.Parallel()
	// plain_local.a's shape (testdata/finite_keyed_write).
	program, err := lowerSource(t, `interface Levels { Low: number; Medium: number; High: number; }
function setLevel(level: keyof Levels, value: number): Levels {
	const levels: Levels = { Low: 0, Medium: 0, High: 0 };
	levels[level] = value;
	return levels;
}
const medium = setLevel('Medium', 1);
console.log(String(medium.Low) + ' ' + String(medium.Medium) + ' ' + String(medium.High));
`)
	if err != nil {
		t.Fatal(err)
	}
	// keyTest is the field name an arm's condition tests the held key against, if it's one.
	keyTest := func(test ir.If) (string, bool) {
		condition, isBinary := test.Condition.(ir.Binary)
		if !isBinary || condition.Operator != ir.Equal {
			return "", false
		}
		name, isString := condition.Right.(ir.StringConstant)
		if !isString {
			return "", false
		}
		return program.Strings[name.Index], true
	}
	var chain *ir.If
	for _, function := range program.Functions {
		if function.Name != "setLevel" {
			continue
		}
		walk(function.Body, func(node any) bool {
			if chain != nil {
				return false
			}
			if test, isIf := node.(ir.If); isIf {
				if _, named := keyTest(test); named {
					chain = &test
					return false
				}
			}
			return true
		})
	}
	if chain == nil {
		t.Fatal("setLevel lowered no if-chain testing its key against a field name")
	}
	stores := map[string]int{}
	arm := *chain
	for {
		name, named := keyTest(arm)
		if !named {
			t.Fatalf("an arm tests %#v, not the key against a field name", arm.Condition)
		}
		walk(arm.Then, func(node any) bool {
			if store, isStore := node.(ir.SetProperty); isStore {
				if store.Name != name {
					t.Errorf("the arm testing %q stores %q", name, store.Name)
				}
				stores[store.Name]++
			}
			return true
		})
		if len(arm.Else) != 1 {
			break
		}
		next, isIf := arm.Else[0].(ir.If)
		if !isIf {
			break
		}
		arm = next
	}
	if len(stores) != 3 || stores["Low"] != 1 || stores["Medium"] != 1 || stores["High"] != 1 {
		t.Errorf("stores %v, want one SetProperty for each of Low, Medium and High", stores)
	}
	if len(arm.Else) != 1 {
		t.Fatalf("after the last key test: %#v, want one panic", arm.Else)
	}
	panicked, isPanic := arm.Else[0].(ir.Panic)
	if !isPanic {
		t.Fatalf("after the last key test: %#v, want a panic", arm.Else[0])
	}
	message, isString := panicked.Message.(ir.StringConstant)
	if !isString || program.Strings[message.Index] != "computed object key outside its proven finite union" {
		t.Fatalf("the panic after the last key test says %#v", panicked.Message)
	}
}
