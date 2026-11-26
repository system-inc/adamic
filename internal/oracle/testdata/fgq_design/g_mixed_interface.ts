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
function copy<T extends Weighs>(source: T): T {
	return { ...source };
}
const values: Weighs[] = [new Scale(), copy(new Scale())];
for (const value of values) {
	try {
		console.log(String(value.weigh()));
	} catch {
		console.log("threw");
	}
}
