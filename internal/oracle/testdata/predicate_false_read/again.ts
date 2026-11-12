// tsc's core.ts some (TypeScript 6.0.3, as tsc_version_zero.ts copies it): its first overload claims
// readonly T[] and answers false for an empty array, so its false side lies there. The call is checked.
// r3 (#39wtqdj, Language's road): the same predicate called again on the same reference on the
// first call's false side, which reads it as its argument. T is given: inferred from the reference the
// false side types undefined, it is unknown, and stage 0 refuses some's body at unknown on main as on
// this head.
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

function twice(items: readonly string[] | undefined): string {
	if (!some(items)) {
		return some<string>(items) ? "changed" : "none";
	}
	return "some";
}

console.log(twice(["a"]));
console.log(twice([]));
