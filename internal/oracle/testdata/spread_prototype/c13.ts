interface Source {
	readonly label: string;
	readonly weight: number;
}
interface Target {
	readonly label: "t";
	readonly weight: number;
}
function isT(x: Source): x is Target {
	return x.label === "t";
}
class Sized {
	readonly label = "t";
	get weight(): number {
		return 20;
	}
}
const honest: Source = { label: "h", weight: 2 };
const built: Source = new Sized();
const value: Source = { ...built };
console.log(typeof honest.weight);
console.log(isT(value) ? "target " + typeof value.weight + " " + String(value.weight) : "no");
