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
const copy: Source = { ...built };
console.log(typeof built.weight + " " + `${built.weight}` + " " + String(built.weight ?? 0));
console.log(typeof copy.weight + " " + String(copy.weight === undefined));
