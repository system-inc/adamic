// A copy of a SpreadWith literal's copy of a Sized seen as Source: neither copy holds weight, Sized's prototype
// getter, though Source says both do, and the read of it stops (Memory's pin on #k7g6f9h, #7mjgxkc's rule).
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
const inner = { ...built, extra: 1 };
const value: Source = { ...inner };
console.log(typeof honest.weight);
console.log(typeof value.weight + " " + String(value.weight));
