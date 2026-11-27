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
function show(source: Source): void {
	console.log(typeof source.weight + " " + `${source.weight}` + " " + String(source.weight ?? 0));
}
const built: Source = new Sized();
const copy: Source = { ...built };
show(built);
show(copy);
