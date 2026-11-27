// The same switch over endings[1], the first element the tuple's rest holds, so its stored type is the
// rest's. A rest element may be absent, so undefined is checked first, as noUncheckedIndexedAccess asks;
// every value built here holds two endings, so the check never answers none.
enum ModuleSpecifierEnding {
	Minimal,
	Index,
	JsExtension,
	TsExtension,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
// The tuple comes back from a function declared to return it, as the adapted tree's allowedEndings comes
// from getAllowedEndingsInPreferredOrder (moduleSpecifiers.ts:199), a fresh literal of the preferred
// ending first. tsc's returns hold one to four endings, but only a literal of two lowers into
// readonly [E, ...E[]] today: one of another length stops in tupleLiteral (object.go), so each value here
// holds two.
function allowedEndingsInPreferredOrder(preferred: ModuleSpecifierEnding, fallback: ModuleSpecifierEnding): readonly [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]] {
	return [preferred, fallback];
}
function secondEnding(endings: readonly [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]]): string {
	if (endings[1] === undefined) return "none";
	switch (endings[1]) {
		case ModuleSpecifierEnding.Minimal:
			return "minimal";
		case ModuleSpecifierEnding.Index:
			return "index";
		case ModuleSpecifierEnding.JsExtension:
			return "js";
		case ModuleSpecifierEnding.TsExtension:
			return "ts";
		default:
			return assertNever(endings[1]);
	}
}
function outside(n: number): ModuleSpecifierEnding {
	return n as ModuleSpecifierEnding;
}
console.log(secondEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.TsExtension)) + " " + secondEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.Index, outside(2))) + " " + secondEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.TsExtension, ModuleSpecifierEnding.Minimal)) + " " + secondEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.JsExtension, outside(1))));
console.log(secondEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.Minimal, outside(8))));
