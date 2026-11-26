// moduleSpecifiers.ts:231: a const declared in the arrow itself, whose initializer reads the enclosing
// function's const, read in the default of a switch over every member and passed to assertNever.
const enum ModuleSpecifierEnding {
	Minimal,
	Index,
	JsExtension,
	TsExtension,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function getPreferredEnding(flag: number): ModuleSpecifierEnding {
	if (flag === 0) return ModuleSpecifierEnding.Minimal;
	if (flag === 1) return ModuleSpecifierEnding.Index;
	if (flag === 2) return ModuleSpecifierEnding.JsExtension;
	if (flag === 3) return ModuleSpecifierEnding.TsExtension;
	return flag as ModuleSpecifierEnding;
}
interface Preferences {
	getAllowedEndingsInPreferredOrder: (syntax: number | undefined) => ModuleSpecifierEnding[];
}
function getPreferences(flag: number, allowTs: boolean): Preferences {
	const filePreferredEnding = getPreferredEnding(flag);
	return {
		getAllowedEndingsInPreferredOrder: syntax => {
			const preferredEnding = syntax !== undefined ? getPreferredEnding(syntax) : filePreferredEnding;
			switch (preferredEnding) {
				case ModuleSpecifierEnding.JsExtension:
					return allowTs
						? [ModuleSpecifierEnding.JsExtension, ModuleSpecifierEnding.TsExtension, ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.Index]
						: [ModuleSpecifierEnding.JsExtension, ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.Index];
				case ModuleSpecifierEnding.TsExtension:
					return [ModuleSpecifierEnding.TsExtension, ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.JsExtension, ModuleSpecifierEnding.Index];
				case ModuleSpecifierEnding.Index:
					return allowTs
						? [ModuleSpecifierEnding.Index, ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.TsExtension, ModuleSpecifierEnding.JsExtension]
						: [ModuleSpecifierEnding.Index, ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.JsExtension];
				case ModuleSpecifierEnding.Minimal:
					return allowTs
						? [ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.Index, ModuleSpecifierEnding.TsExtension, ModuleSpecifierEnding.JsExtension]
						: [ModuleSpecifierEnding.Minimal, ModuleSpecifierEnding.Index, ModuleSpecifierEnding.JsExtension];
				default:
					assertNever(preferredEnding);
			}
		},
	};
}
const preferences = getPreferences(2, false);
console.log(preferences.getAllowedEndingsInPreferredOrder(undefined).join(",") + " " + preferences.getAllowedEndingsInPreferredOrder(1).join(",") + " " + getPreferences(0, true).getAllowedEndingsInPreferredOrder(3).join(","));
console.log(getPreferences(9, true).getAllowedEndingsInPreferredOrder(undefined).join(","));
