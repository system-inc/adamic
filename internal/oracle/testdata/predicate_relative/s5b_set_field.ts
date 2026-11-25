interface Base { name: string; values: "none" | Set<string> }
interface WithSet extends Base { values: Set<string> }
function isWithSet(option: Base): option is WithSet {
	return option.values !== "none";
}
const option: Base = { name: "a", values: new Set<string>(["x"]) };
console.log(isWithSet(option) ? "set " + String(option.values.size) : "none");
