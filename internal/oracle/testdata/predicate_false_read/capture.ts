// tsc's core.ts some (TypeScript 6.0.3, as tsc_version_zero.ts copies it): its first overload claims
// readonly T[] and answers false for an empty array, so its false side lies there. The call is checked.
// p5 (#39wtqdj): the only false-side read is in a closure made on the false path, which captures it.
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

function later(items: readonly string[] | undefined): string {
	if (!some(items)) {
		const read = (): string => typeof items;
		return read();
	}
	return "some";
}

console.log(later(["a"]));
console.log(later([]));
