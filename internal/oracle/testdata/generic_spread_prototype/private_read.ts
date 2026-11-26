interface Weighs {
	readonly label: string;
	weigh(): number;
}
class Scale {
	readonly label = "t";
	#grams = 20;
	readonly extra = 7;
	weigh(): number {
		return this.#grams + this.extra;
	}
}
function copy<T extends Weighs>(source: T): T {
	return { ...source };
}
const value = copy(new Scale());
try {
	console.log(String(value.weigh()));
} catch {
	console.log("threw");
}
