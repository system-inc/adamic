interface Weighs {
	readonly label: string;
	weigh(): number;
}
class Scale implements Weighs {
	readonly label = "t";
	#grams = 20;
	readonly extra = 7;
	weigh(): number {
		return this.#grams + this.extra;
	}
}
function copy(source: Weighs): Weighs {
	return { ...source };
}
const value = copy(new Scale());
console.log(String(!value.weigh));
