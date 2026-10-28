package lower

import (
	"errors"
	"strings"
	"testing"
)

// A generic instantiated at void whose body reads the value is refused at the call, naming the read
// in the body (#p7t8tmh part 2). all holds Node's discarded 1 and 2, so each body below would read a
// value Node keeps and Adamic released.
func TestGenericVoidReadsAreRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, role string }{
		// Language's road: T = void, read in a callback.
		{"converted in a callback", "function showAll<T>(xs: T[]): void { xs.forEach((x) => { console.log(String(x)); }); }\nshowAll(all);", "as an argument"},
		{"compared", "function hasHole<T>(xs: T[]): boolean { return xs[0] === undefined; }\nconsole.log(`${hasHole(all)}`);", "as an operand of ==="},
		{"a value compared to itself", "function selfSame<T>(xs: T[]): boolean { let same = false; for (const x of xs) { same = x === x; } return same; }\nconsole.log(`${selfSame(all)}`);", "as an operand of ==="},
		{"stringified whole", "function show<T>(xs: T[]): string { return String(JSON.stringify(xs)); }\nconsole.log(show(all));", "as an argument"},
		{"joined", "function joined<T>(xs: T[]): string { return xs.join(','); }\nconsole.log(joined(all));", "through join"},
		{"set of", "function count<T>(xs: T[]): number { return new Set(xs).size; }\nconsole.log(`${count(all)}`);", "as an argument"},
		{"forEach named", "const log = (x: unknown): void => { console.log(typeof x); };\nfunction each<T>(xs: T[]): void { xs.forEach(log); }\neach(all);", "with a callback not written in place"},
		{"includes", "function has<T>(xs: T[]): boolean { let found = false; for (const x of xs) { found = xs.includes(x); } return found; }\nconsole.log(`${has(all)}`);", "through includes"},
		{"a value returned wider", "function give<T>(xs: T[]): unknown { return xs[0]; }\nconsole.log(typeof give(all));", "as a returned value"},
		// T bound to a type holding void, not to void: the container is read whole.
		{"holder converted", "function showWhole<T>(x: T): string { return String(x); }\nconsole.log(showWhole(all));", "as an argument"},
		// Lower's condition 1: inner is instantiated at void only when outer is lowered at void.
		{"nested", "function inner<U>(ys: U[]): string { return ys.join(','); }\nfunction outer<T>(xs: T[]): string { return inner(xs); }\nconsole.log(outer(all));", "through join"},
		// Lower's condition 2: T reaches the signature only through an object, which inferTypes doesn't walk.
		{"unmapped", "function showItems<T>(box: { items: T[] }): string { return box.items.join(','); }\nconsole.log(showItems({ items: all }));", "through join"},
		// Lower's review: ??= reads its target, and Node never calls fail, since each x holds 1 or 2.
		{"logical assignment target", "function fail(): never { throw new Error('read'); }\nfunction fill<T>(xs: T[]): number { let n = 0; for (let x of xs) { x ??= (n++, fail()); } return n; }\nconsole.log(`${fill(all)}`);", "as the target of ??="},
		// Lower's review: T reaches the signature only through a callback's result.
		{"unmapped through a result", "const make: () => void = () => 42;\nfunction called<T>(box: { make: () => T }): boolean { return box.make() === undefined; }\nconsole.log(`${called({ make })}`);", "as an operand of ==="},
		// Lower's review: a base class's body is judged at its own instantiation.
		{"base class", "class Bag<T> { constructor(readonly items: T[]) {} joined(): string { return this.items.join(','); } }\nclass Sub<T> extends Bag<T> {}\nconst sub = new Sub(all); console.log(`${sub.items.length}`);", "through join"},
		// Lower's condition 4, and Compiler's: a generic class at void whose body reads T refuses at new.
		{"generic class", "class Bag<T> { constructor(readonly items: T[]) { console.log(this.items.join(',')); } }\nconst bag = new Bag(all); console.log(`${bag.items.length}`);", "through join"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerVoidProbe(t, extension, voidContainerMark+"\n"+probe.source)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "a generic instantiated at void whose body reads the value") || !strings.Contains(refused.What, probe.role) {
					t.Fatalf("want a generic void read refused %s, got %v", probe.role, err)
				}
			})
		}
	}
}

// A body that only moves its T values stays legal: length, identity of the container, a T returned
// where T is returned, and a T stored where a T is kept.
func TestGenericVoidMovesStayLegal(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"length", "function size<T>(xs: T[]): number { return xs.length; }\nconsole.log(`${size(all)}`);"},
		{"identity", "function same<T>(xs: T[], ys: T[]): boolean { return xs === ys; }\nconsole.log(`${same(all, all)}`);"},
		{"holder passed through", "function pass<T>(x: T): T { return x; }\nconsole.log(`${pass(all).length}`);"},
		{"returned whole", "function keep<T>(xs: T[]): T[] { return xs; }\nconsole.log(`${keep(all).length}`);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNode(t, voidContainerMark+"\n"+probe.source)
		})
	}
}

// Lower's condition 3: a generic used as a value doesn't instantiate through a call. Today each stops
// before it runs (ToPrimitive in the body is not yet, and the classifier refuses forEach over a void
// container with a named callback), so these are pinned as stopped; whichever change lowers them must
// judge the body there. A generic method of a class that isn't generic is the same: it lowers once,
// not per call, and an array of T is not yet.
func TestGenericVoidAsValueStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"typed value", "function showAll<T>(xs: T[]): void { xs.forEach((x) => { console.log(String(x)); }); }\nconst f: (xs: void[]) => void = showAll; f(all);"},
		{"callback", "function showAll<T>(xs: T[]): void { xs.forEach((x) => { console.log(String(x)); }); }\n[all].forEach(showAll);"},
		{"class generic method", "class Holder { show<T>(xs: T[]): string { return xs.join(','); } }\nconsole.log(new Holder().show(all));"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerVoidProbe(t, extension, voidContainerMark+"\n"+probe.source); err == nil {
					t.Fatalf("want it stopped before it runs, and it lowered")
				}
			})
		}
	}
}

// A method of a generic class at void stays refused at its call (Adamic, on #p7t8tmh), and the message
// says what holds the void and where the method reads it, so no reader goes looking for a container.
func TestGenericVoidMethodSaysWhereItReads(t *testing.T) {
	t.Parallel()
	source := voidContainerMark + "\nconst tick: () => void = () => 1;\nclass Cell<T> { constructor(private readonly f: () => T) {} show(): string { return `${this.f()}`; } }\nconsole.log(new Cell(tick).show());"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerVoidProbe(t, extension, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "through show: Cell<void> holds void in its type argument 1 (void), and show reads it at ") || !strings.HasSuffix(refused.What, "as a template literal part") {
				t.Fatalf("want show refused naming what holds the void and where show reads it, got %v", err)
			}
		})
	}
}
