// The same switch over endings[2], an element the tuple's rest holds, so its stored type is the rest's.
// A rest element may be absent, so undefined is checked first, as noUncheckedIndexedAccess asks.
enum ModuleSpecifierEnding {
	Minimal,
	Index,
	JsExtension,
	TsExtension,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function thirdEnding(endings: readonly [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]]): string {
	if (endings[2] === undefined) return "none";
	switch (endings[2]) {
		case ModuleSpecifierEnding.Minimal:
			return "minimal";
		case ModuleSpecifierEnding.Index:
			return "index";
		case ModuleSpecifierEnding.JsExtension:
			return "js";
		case ModuleSpecifierEnding.TsExtension:
			return "ts";
		default:
			return assertNever(endings[2]);
	}
}
function outside(n: number): ModuleSpecifierEnding {
	return n as ModuleSpecifierEnding;
}
console.log(thirdEnding([ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.Index, ModuleSpecifierEnding.TsExtension]) + " " + thirdEnding([ModuleSpecifierEnding.Minimal]) + " " + thirdEnding([ModuleSpecifierEnding.Index, ModuleSpecifierEnding.Index, outside(2), ModuleSpecifierEnding.Minimal]) + " " + thirdEnding([ModuleSpecifierEnding.TsExtension, ModuleSpecifierEnding.JsExtension, ModuleSpecifierEnding.Minimal]));
