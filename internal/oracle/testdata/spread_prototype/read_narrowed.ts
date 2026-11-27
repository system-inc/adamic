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
if (typeof copy.weight === "number") {
	console.log(String(copy.weight + 1));
} else {
	console.log("absent");
}
