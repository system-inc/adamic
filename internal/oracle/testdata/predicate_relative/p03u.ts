interface Dict { tag: string; [field: string]: string }
type Named = { tag: string; name: string };
function isNamed(x: Dict): x is Named {
	return x.tag === "n";
}
const dict: Dict = { tag: "n" };
console.log(isNamed(dict) ? "named " + typeof dict.name : "no");
