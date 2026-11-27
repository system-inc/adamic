// A stop: the switch over endings[0] narrows endings[first] too, since the checker matches the two by the
// name each reads, a const of literal type 0 naming the same element. But first is no numeric literal, so
// the element it reads isn't placed, and the argument keeps today's stop.
enum ModuleSpecifierEnding {
	Minimal,
	Index,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
const first = 0;
function endingName(endings: readonly [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]]): string {
	switch (endings[0]) {
		case ModuleSpecifierEnding.Minimal:
			return "minimal";
		case ModuleSpecifierEnding.Index:
			return "index";
		default:
			return assertNever(endings[first]);
	}
}
console.log(endingName([ModuleSpecifierEnding.Index]));
