// tsc's core.ts some (TypeScript 6.0.3, as tsc_version_zero.ts copies it): its first overload claims
// readonly T[] and answers false for an empty array, so its false side lies there. The call is checked.
// p1 (#39wtqdj): semver.ts's Version.toString shape, nothing reads this.build on the false side, so
// the false check is dropped and an empty build prints as Node prints it.
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

class Version {
	readonly build: readonly string[];

	constructor(build: readonly string[]) {
		this.build = build;
	}

	toString(): string {
		let result = "1.2.3";
		if (some(this.build)) result += "+" + this.build.join(".");
		return result;
	}
}

console.log(new Version([]).toString());
console.log(new Version(["build", "5"]).toString());
