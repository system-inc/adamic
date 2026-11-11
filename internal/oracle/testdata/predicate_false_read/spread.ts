// tsc's core.ts some (TypeScript 6.0.3, as tsc_version_zero.ts copies it): its first overload claims
// readonly T[] and answers false for an empty array, so its false side lies there. The call is checked.
// r2 (#39wtqdj, Language's road): the false side spread, typed never where it holds [].
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

function spread(items: readonly string[]): string {
	if (!some(items)) {
		return String([...items].length);
	}
	return "some";
}

console.log(spread(["a"]));
console.log(spread([]));
