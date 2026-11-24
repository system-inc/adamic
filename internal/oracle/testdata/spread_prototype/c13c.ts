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
const built: Source = new Sized();
const value: Source = { ...built };
console.log(typeof value.weight + " " + String(value.weight));
