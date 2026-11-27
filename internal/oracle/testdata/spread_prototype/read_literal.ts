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
const plain: Source = { label: "l", weight: 3 };
console.log(typeof plain.weight);
console.log(typeof copy.weight);
