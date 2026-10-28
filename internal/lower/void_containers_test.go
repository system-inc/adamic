package lower

import (
	"errors"
	"strings"
	"testing"
)

// mark hands back its argument through a void type, so Node's all holds 1 and 2 where Adamic, which
// releases a void call's result (#f418fbz), holds undefined twice. Language's attack on #p7t8tmh found
// the library reading those elements with no read in user code: all.includes(undefined) printed true
// natively and on the JavaScript backend, where Node prints false.
const voidContainerMark = `const mark: (i: number) => void = (i) => i;
const all: void[] = [1, 2].map((i) => mark(i));
const nested: void[][] = [all, all];
const box: { readonly items: void[] } = { items: all };`

func TestVoidContainerReadsAreRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, role string }{
		// Language's three, live on main 8d39519a: Node prints false, -1 and -1.
		{"includes", "console.log(`${all.includes(undefined)}`);", "through includes"},
		{"indexOf", "console.log(`${all.indexOf(undefined)}`);", "through indexOf"},
		{"lastIndexOf", "console.log(`${all.lastIndexOf(undefined)}`);", "through lastIndexOf"},
		{"join", `console.log(all.join(","));`, "through join"},
		{"at", "console.log(`${all.at(0) === undefined}`);", "through at"},
		{"find", "console.log(`${all.find((x) => true) === undefined}`);", "through find"},
		// The container is a call's result, a callback parameter or a field read, never a name.
		{"flat result", "console.log(`${nested.flat().includes(undefined)}`);", "through includes"},
		{"callback parameter", "console.log(`${nested.some((inner) => inner.includes(undefined))}`);", "through some"},
		{"field read", "console.log(`${box.items.includes(undefined)}`);", "through includes"},
		{"nested toString", `console.log(nested.toString());`, "through toString"},
		{"nested join", `console.log(nested.join(";"));`, "through join"},
		// Handed whole to something that reads it. A spread is H2's: each element it hands out is a void value.
		{"stringified", `console.log(String(JSON.stringify(all)));`, "as an argument"},
		{"converted", `console.log(String(all));`, "as an argument"},
		{"template", "console.log(`${all}`);", "as a template literal part"},
		{"bound wider", "const seen: readonly unknown[] = all; console.log(`${seen.length}`);", "as a binding's initializer"},
		{"optional slot of two", "const either: void[] | number[] | undefined = all; console.log(`${either !== undefined}`);", "as a binding's initializer"},
		{"field dropped", "const thin: { readonly name: string } = { name: 'n', ...{ items: all } }; console.log(thin.name);", "as a binding's initializer"},
		// forEach is a move only for a callback written in place that keeps its parameters void.
		{"forEach named", "const log = (x: unknown): void => { console.log(typeof x); }; all.forEach(log);", "with a callback not written in place"},
		{"forEach wider", "all.forEach((x: unknown) => { console.log(typeof x); });", "parameter x, typed wider"},
		{"forEach array wider", "all.forEach((x, i, whole: readonly unknown[]) => { console.log(typeof whole[i]); });", "parameter whole, typed wider"},
		// A void element beside undefined is still a void element (Adamic's condition on #p7t8tmh).
		{"tuple at a computed index", "const pair: [void, string] = [undefined, 'b']; const at = (i: number) => pair[i]; console.log(typeof at(0));", "whose value is read as"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerVoidProbe(t, extension, voidContainerMark+"\n"+probe.source)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "a read of a void container's elements") || !strings.Contains(refused.What, probe.role) || !strings.Contains(refused.Fix, "discard them with forEach") {
					t.Fatalf("want a void container read refused %s, got %v", probe.role, err)
				}
			})
		}
	}
}

// pop and shift hand out an element typed void | undefined, which tsc folds to void, and an optional
// call of a void function folds the same way: a read of either is a void read, by H2's rule or by this
// one, even though undefined stands beside it (Adamic's condition on #p7t8tmh).
func TestVoidElementsBesideUndefinedAreRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"pop", "console.log(`${all.pop() === undefined}`);"},
		{"shift", "console.log(`${all.shift() === undefined}`);"},
		{"optional call", "const make: () => void = () => 42; console.log(`${make?.() === undefined}`);"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerVoidProbe(t, extension, voidContainerMark+"\n"+probe.source)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "a read of a void") || !strings.Contains(refused.What, "operand of ===") {
					t.Fatalf("want a void read refused as an operand of ===, got %v", err)
				}
			})
		}
	}
}

// The moves stay legal: length, identity, a container entering a slot that keeps it void, a loop that
// binds the element and never reads it, a forEach that never reads it, and pushing undefined, which
// truly is undefined on both sides.
func TestVoidContainerMovesStayLegal(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"length", "console.log(`${all.length}`);"},
		{"identity", "const other = all; console.log(`${all === other}`);"},
		{"passed as void[]", "function size(values: readonly void[]): number { return values.length; } console.log(`${size(all)}`);"},
		{"optional void[] slot", "function size(values?: void[]): number { return values === undefined ? 0 : values.length; } console.log(`${size(all)}`);"},
		{"field length", "console.log(`${box.items.length}`);"},
		{"loop without a read", "let n = 0; for (const x of all) { n += 1; } console.log(`${n}`);"},
		{"forEach without a read", "let n = 0; all.forEach((x) => { n += 1; }); console.log(`${n}`);"},
		{"pop discarded", "const copy: void[] = all.slice(); copy.pop(); console.log(`${copy.length}`);"},
		{"push undefined", "const more: void[] = []; more.push(undefined); console.log(`${more.length}`);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNode(t, voidContainerMark+"\n"+probe.source)
		})
	}
}
