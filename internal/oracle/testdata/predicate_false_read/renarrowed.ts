// q1 (#39wtqdj, Language's program from 08:23Z, exact): the false side narrows empty to undefined and
// the !== test narrows it again, to never, at kept's read, which the flow graph still carries from the
// false condition. Node prints 0.
function some<T>(array: readonly T[] | undefined): array is readonly T[] { return array !== undefined && array.length > 0; }
const empty: readonly string[] | undefined = [] as readonly string[] | undefined;
if (!some(empty)) {
	if (empty !== undefined) {
		const kept: readonly string[] = empty;
		console.log(String(kept.length));
	}
}
