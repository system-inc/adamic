interface Source {
	readonly label: string;
	readonly weight: number;
}
class Sized {
	readonly label = "t";
	set weight(value: number) {
		console.log("set " + String(value));
	}
}
const built: Source = new Sized();
const copy: Source = { ...built };
const n: number = copy.weight;
console.log(String(n));
