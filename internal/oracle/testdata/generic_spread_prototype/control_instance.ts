interface Weighs {
	readonly label: string;
	weigh(): number;
}
class Scale {
	readonly label = "t";
	#grams = 20;
	weigh(): number {
		return this.#grams;
	}
}
function same<T extends Weighs>(source: T): T {
	return source;
}
console.log(String(same(new Scale()).weigh()));
