// tsc's core.ts some (TypeScript 6.0.3, as tsc_version_zero.ts copies it): its first overload claims
// readonly T[] and answers false for an empty array, so its false side lies there. The call is checked.
// r1 reached (#39wtqdj, Language's road): the false side read again through a getter, inside a closure made
// on the false path. A getter's read narrowed to undefined stops on main as on this head (getter.ts), and
// the checker doesn't narrow a property read inside a closure, so this read reaches the call; the flow
// graph still carries the false answer to it, through the closure's start.
function some<T>(array: readonly T[] | undefined): array is readonly T[];
function some<T>(array: readonly T[] | undefined, predicate: (value: T) => boolean): boolean;
function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean {
	if (array !== undefined) {
		if (predicate !== undefined) {
			for (let i = 0; i < array.length; i++) {
				if (predicate(array[i]!)) {
					return true;
				}
			}
		}
		else {
			return array.length > 0;
		}
	}
	return false;
}

class Holder {
	stored: readonly string[] | undefined;

	constructor(stored: readonly string[] | undefined) {
		this.stored = stored;
	}

	get items(): readonly string[] | undefined {
		return this.stored;
	}
}

function show(holder: Holder): string {
	if (!some(holder.items)) {
		const read = (): string => typeof holder.items;
		return read();
	}
	return "some";
}

console.log(show(new Holder(["a"])));
console.log(show(new Holder([])));
