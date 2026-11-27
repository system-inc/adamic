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
console.log(String(copy.weight === undefined));
