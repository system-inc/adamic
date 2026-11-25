interface Source { kind: string; name: string }
interface Target { kind: "t"; name: string }
function isTarget(x: Source): x is Target {
	return x.kind === "t";
}
const all: Record<string, string> = { kind: "t" };
const viewed: Source = all as unknown as Source;
console.log(isTarget(viewed) ? "target " + typeof viewed.name : "no");
