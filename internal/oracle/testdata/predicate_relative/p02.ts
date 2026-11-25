interface Loose { kind: string; name?: string }
interface Named { kind: "n"; name: string }
function isNamed(x: Loose): x is Named {
	return x.kind === "n";
}
const loose: Loose = { kind: "n" };
console.log(isNamed(loose) ? "named " + typeof loose.name : "no");
