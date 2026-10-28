package lower

import (
	"errors"
	"strings"
	"testing"
)

// A promise that holds void moves through await, through then and catch with callbacks written in
// place, and through Promise.all, allSettled and race, since each yields a value judged in turn
// (#tfycfvw, Adamic's ruling on #p7t8tmh). later resolves to the discarded 1 and 2 on Node.
const voidPromiseMark = voidContainerMark + `
async function later(): Promise<void[]> { return all; }
async function done(): Promise<void> {}`

// await yields an exactly void value, which H2 judges: discarding it and binding it unread stay legal.
func TestVoidPromiseAwaitStaysLegal(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, voidContainerMark+"\n"+`async function done(): Promise<void> {}
const pending: Promise<void> = done();
await pending; const got = await done(); console.log(`+"`${all.length}`"+`);`)
}

// Stage 0 doesn't lower then, catch or the combinators over these yet, so their moves are held where
// this rule stands: it refuses none of them. Whatever else stops the program today stops it later, at
// lowering, and none of it is this rule's.
func TestVoidPromiseMovesAreNotContainerReads(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"then", "const size = await later().then((xs) => xs.length); console.log(`${size}`);"},
		{"then with undefined", "const kept = await later().then(undefined, () => all); console.log(`${kept.length}`);"},
		{"catch", "const kept = await later().catch(() => all); console.log(`${kept.length}`);"},
		{"all", "const both = await Promise.all([later(), later()]); console.log(`${both.length}`);"},
		{"allSettled", "const settled = await Promise.allSettled([done()]); console.log(`${settled.length}`);"},
		{"race", "const first = await Promise.race([later(), later()]); console.log(`${first.length}`);"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerVoidProbe(t, extension, voidPromiseMark+"\n"+probe.source)
				var refused *Refused
				if errors.As(err, &refused) && strings.HasPrefix(refused.What, "a read of a void container's elements") {
					t.Fatalf("want this move kept, got %v", err)
				}
			})
		}
	}
}

// What comes out of each move is still judged, so a read through one is refused where it reads.
func TestVoidPromiseReadsAreRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, role string }{
		{"awaited then read", "const got = await later(); console.log(`${got.includes(undefined)}`);", "through includes"},
		{"then reading", "const seen = await later().then((xs) => xs.includes(undefined)); console.log(`${seen}`);", "through includes"},
		{"then wider", "const seen = await later().then((xs: unknown) => String(xs)); console.log(seen);", "parameter xs, typed wider"},
		{"then named", "const show = (xs: readonly unknown[]): string => String(xs); console.log(await later().then(show));", "with a callback not written in place"},
		{"catch named", "const recover = (): never[] => []; const kept = await later().catch(recover); console.log(`${kept.length}`);", "with a callback not written in place"},
		{"all then joined", `const both = await Promise.all([later(), later()]); console.log(both.join(";"));`, "through join"},
		{"race then joined", `const first = await Promise.race([later()]); console.log(first.join(","));`, "through join"},
		{"finally", `await later().finally(() => { console.log("done"); });`, "through finally"},
		// Part 1's widening refuses this first: the promise's void elements are seen wider in the slot.
		{"resolve wider", "const seen: Promise<readonly unknown[]> = later(); console.log(String((await seen).length));", voidWidenedFirst},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerVoidProbe(t, extension, voidPromiseMark+"\n"+probe.source)
				var refused *Refused
				if probe.role == voidWidenedFirst {
					if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, voidWidenedFirst) {
						t.Fatalf("want part 1's widening to refuse it first, got %v", err)
					}
					return
				}
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "a read of a void container's elements") || !strings.Contains(refused.What, probe.role) {
					t.Fatalf("want a void container read refused %s, got %v", probe.role, err)
				}
			})
		}
	}
}
