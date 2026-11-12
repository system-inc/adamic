// tsc's core.ts some (TypeScript 6.0.3, as tsc_version_zero.ts copies it): its first overload claims
// readonly T[] and answers false for an empty array, so its false side lies there. The call is checked.
// r1 (#39wtqdj, Language's road): the false side read again through a getter, which the checker
// narrows as any property access. That read, narrowed to undefined, stops lowering on main as on this
// head, before the call is reached; getter_capture.ts reaches it.
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
		return typeof holder.items;
	}
	return "some";
}

console.log(show(new Holder(["a"])));
console.log(show(new Holder([])));
