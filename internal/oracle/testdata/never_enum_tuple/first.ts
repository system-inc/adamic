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
console.log(processEnding([ModuleSpecifierEnding.JsExtension, ModuleSpecifierEnding.Minimal]) + " " + processEnding([ModuleSpecifierEnding.Minimal]) + " " + processEnding([outside(3), ModuleSpecifierEnding.Index, ModuleSpecifierEnding.JsExtension]) + " " + processEnding([outside(1)]));
