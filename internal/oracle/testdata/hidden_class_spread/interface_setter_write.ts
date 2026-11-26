interface Source {
	readonly label: string;
	weight: number;
}
class Sized implements Source {
	readonly label = "t";
	get weight(): number {
		return 1;
	}
	set weight(given: number) {
		console.log("set " + String(given));
	}
}
function copy(source: Source): Source {
	return { ...source };
}
const value = copy(new Sized());
value.weight = 5;
console.log(String(JSON.stringify(value)));
