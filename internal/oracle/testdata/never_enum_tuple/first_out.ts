// moduleSpecifiers.ts:1386: processEnding's switch over allowedEndings[0], the first element of a tuple
// with a rest, read again in the default and passed to assertNever.
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
function processEnding(allowedEndings: readonly [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]]): string {
	switch (allowedEndings[0]) {
		case ModuleSpecifierEnding.Minimal:
			return "minimal";
		case ModuleSpecifierEnding.Index:
			return "index";
		case ModuleSpecifierEnding.JsExtension:
			return "js";
		case ModuleSpecifierEnding.TsExtension:
			return "ts";
		default:
			return assertNever(allowedEndings[0]);
	}
}
function outside(n: number): ModuleSpecifierEnding {
	return n as ModuleSpecifierEnding;
}
console.log(processEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.JsExtension, ModuleSpecifierEnding.Minimal)) + " " + processEnding(allowedEndingsInPreferredOrder(ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.Index)) + " " + processEnding(allowedEndingsInPreferredOrder(outside(3), ModuleSpecifierEnding.Index)) + " " + processEnding(allowedEndingsInPreferredOrder(outside(1), ModuleSpecifierEnding.JsExtension)));
console.log(processEnding(allowedEndingsInPreferredOrder(outside(9), ModuleSpecifierEnding.Minimal)));
