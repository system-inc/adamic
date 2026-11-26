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
function copy<T extends Source>(source: T): T {
	return { ...source };
}
const value = copy(new Sized());
console.log(typeof value.weight + " " + String(value.weight));
