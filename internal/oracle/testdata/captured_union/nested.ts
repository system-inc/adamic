// commandLineParser.ts:3577's shape: a let of string | string[] | undefined, written from a
// callback nested in another, then read by the function that declares it.
function parse(entries: readonly string[]): string {
	let extended: string | string[] | undefined;
	const onPropertySet = (key: string, value: string): void => {
		const assign = (): void => {
			if (key === "extends") {
				extended = value;
			} else if (key === "extendsAll") {
				extended = value.split(",");
			} else if (key === "reset") {
				extended = undefined;
			}
		};
		assign();
	};
	entries.forEach((entry) => {
		const at = entry.indexOf("=");
		onPropertySet(entry.slice(0, at), entry.slice(at + 1));
	});
	if (extended === undefined) {
		return "no base";
	}
	if (typeof extended === "string") {
		return `one base ${extended}`;
	}
	return `${extended.length} bases ${extended.join("+")}`;
}

console.log(parse([]));
console.log(parse(["extends=./base.json"]));
console.log(parse(["extendsAll=./a.json,./b.json,./c.json"]));
console.log(parse(["extends=./base.json", "extendsAll=./a.json,./b.json"]));
console.log(parse(["extendsAll=./a.json,./b.json", "extends=./base.json"]));
console.log(parse(["extends=./base.json", "reset="]));
console.log(parse(["target=es2020"]));
