package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A read the checker narrowed to undefined or null, of a binding or a property a call may write, stops as
// not yet, in .a and .ts (#r0vzaaa). The checker keeps the narrowing across a call it doesn't look
// inside, and a consumer that asks it for the read's type folds the read to that value, so both backends
// printed undefined where Node prints 4. Each stops at the read: the callee writes the module binding
// (n01), the null twin (n03), a let a closure the call runs writes (n04), a parameter the same way (n08),
// a module binding another module's function writes, read in the importing module under its own check
// (n05), a module binding its own setter writes, called from the module it imports (n06,
// n06_module_setter/back calling setLater), and ??= in the callee (n07, #aw65psm's g05). A write of
// another form counts as one: ++ (n09, which writes nothing on this run, so Node prints undefined and the
// stop is the rule's conservative answer), an array or object destructuring target (n10, n11) and a
// for-of head (n12), each in a callee declared below the read. So does a read right after its assignment
// when something before it in its statement can run a function that writes it: a call (n15) or a
// template's conversion of an object (n16). After a call that writes it, a template read stops (n19), and
// so does a read passed to a program function whose parameter holds only undefined (n17) or a generic one
// inferred from the read as undefined (n18), since the callee folds it inside. There's no compound read
// here: tsc refuses `s += 'a'` on a binding narrowed to undefined (TS2322), so it never reaches lowering.
// A property narrowed the same way, with code that can run between its narrowing and the read, stops when
// its name is written anywhere but that narrowing: a class field a callee writes (f01), an object
// literal's field (f02), a template read of it (f03), the null twin (f04), a delete in a callee, which
// leaves undefined where the checker says null (f08), a write through an alias in the read's own code
// (f09), and a recursive call back into the read's own function (f10), where Node prints 4 and both
// backends printed undefined. So does one in a program with a write that could reach any name,
// Object.assign in a callee (f05) or a computed key in the read's own function (f06), which the checker
// doesn't see as a write of the name, and one of a name a number spells, an element an array method
// writes (f07). A getter of the name stops the read whatever its flow says, since it answers each read
// itself: one whose value changes with other state, read after a check and a call that writes nothing
// named x (f11), and one read right after an assignment its setter takes (f12), where Node prints 4 and
// both backends printed undefined. Their controls lower and agree with Node on release native, sanitized native and the
// JavaScript backend: a callee that doesn't write the binding (n02), a local no other function writes,
// read narrowed after a call (c01), a binding narrowed to number before the call that writes it (c02,
// narrowed.go's own road), a field the callee writes but nothing narrowed (c03), a field read with
// nothing that runs code since its assignment (c04, the shape of the counted nbody_field_values.a:18:47),
// a field read after a call where nothing writes the name but its narrowing (c06), a read in the
// statement right after its assignment, with nothing before it that runs code (c05, the shape of the
// counted maybe_numbers.a:10:16), a read passed whole to a program function whose parameter holds the
// declared type, which reads storage (c07, the shape of the counted regions.a:78:40), and an imported
// binding read after the call that sets it, which the checker doesn't narrow in the importing module
// (n14). Each counted program the first property rule stopped has a control of its shape: an optional
// chain between the assignment and the read (c08, graph_regions_list.a), a delete as the narrowing with a
// literal union's delete elsewhere (c09, the records fixtures'), a check as the narrowing (c10,
// scout_nullable_object_null_field.a), a call between that reaches no function writing the name (c11,
// weak_parent.a), a method call between, in a function only the module's top level calls, whose own
// earlier writes can't run again (c12, class_oct6_subclass_holder.a), and a library push between, after
// an element's assignment (c13, scout_nullable_strings.a). Their negatives stop: a check followed by a
// call in the same condition that writes the name (f13), a check that proves less than the narrowing,
// `!== null` over a read narrowed to undefined (f14), a write two direct calls away (f15), a write in the
// read's own function entered again through a method (f16), and a push after a check of an index, which
// can write that index (f17). (== is refused, and so is a delete on a fixed object or an array, so f08
// and c09 delete from a record.)
// Mutants: narrowedNullishReadGap's binding test forced to let every binding through, and every binding
// stop here lowers instead, printing undefined where Node prints 4. assignedJustBefore forced to answer
// no, and c05 stops. passedAtDeclaredType's parameter test dropped, and n17 lowers. The property rule's
// stop forced off, and f01 to f10 lower; the write that could reach any name ignored, and f05 and f06
// lower; code between the property's assignment and its read always assumed, and c04 stops; the getter
// test (accessorName) dropped, and f11 and f12 lower; reach taken as nothing, and f15 and f16 lower;
// calledOnlyFromTopLevel answering yes, and f16 lowers; the push exemption kept after a check, and f17
// lowers.
func TestNarrowedToNullishReadsArePinned(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/narrowed_to_nullish"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				test.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					test.Fatalf("lowered, want %q: Node prints %q", stop, onNode(test, absolute).stdout)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					test.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				test.Fatalf("Lower: %v", err)
			}
			truth := onNode(test, absolute)
			sanitized, _ := natively(test, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"release", released(test, program)},
				{"sanitized", sanitized},
				{"JavaScript", onJavaScriptBackend(test, program)},
			} {
				if difference := disagreement(truth, backend.got); difference != "" {
					test.Fatalf(
						"%s: %s: Node %q, %s %q",
						backend.name,
						difference,
						truth.stdout,
						backend.name,
						backend.got.stdout,
					)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 96 {
		test.Fatalf("found %d programs, want 96", programs)
	}
}
