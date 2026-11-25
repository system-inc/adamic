interface Dict { kind: string; [field: string]: string }
interface Named { kind: "n"; name: string }
function isNamed(x: Dict): x is Dict & Named {
	return x.kind === "n";
}
const dict: Dict = { kind: "n" };
console.log(isNamed(dict) ? "named " + typeof dict.name : "no");
