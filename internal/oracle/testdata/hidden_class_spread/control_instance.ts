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
const value: Weighs = new Scale();
console.log(value.label + " " + String(value.weigh()));
