// A stop: an optional element may be absent, so its stored type isn't read from it, and the read keeps
// today's stop.
enum ModuleSpecifierEnding {
	Minimal,
	Index,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function endingName(endings: [ModuleSpecifierEnding?]): string {
	if (endings[0] === undefined) return "none";
	if (endings[0] === ModuleSpecifierEnding.Minimal) return "minimal";
	if (endings[0] === ModuleSpecifierEnding.Index) return "index";
	return assertNever(endings[0]);
}
console.log(endingName([ModuleSpecifierEnding.Index]));
