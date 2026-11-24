interface Source {
	readonly label: string;
	readonly weight: number;
}
class Sized {
	readonly label = "t";
	get weight(): number {
		return 20;
	}
}
const honest: Source = { label: "h", weight: 2 };
const built: Source = new Sized();
const value: Source = { ...built, weight: built.weight + 1 };
console.log(typeof honest.weight);
console.log(typeof value.weight + " " + String(value.weight));
