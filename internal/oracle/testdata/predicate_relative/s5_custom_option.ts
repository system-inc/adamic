interface Base { name: string; type: "string" | "number" | ReadonlyMap<string, string | number> }
interface Custom extends Base { type: ReadonlyMap<string, string | number> }
function isCustom(option: Base): option is Custom {
	return typeof option.type !== "string";
}
const options: Base[] = [{ name: "a", type: "string" }, { name: "b", type: new Map<string, string | number>([["x", 1]]) }];
console.log(String(options.filter(isCustom).length));
