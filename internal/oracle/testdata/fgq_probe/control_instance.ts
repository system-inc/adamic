interface Weighs { weigh(): number }
class Scale implements Weighs {
	#grams: number;
	constructor(grams: number) { this.#grams = grams; }
	weigh(): number { return this.#grams * 2; }
}
function same<T extends Weighs>(source: T): T { return source; }
const c = same(new Scale(10));
try { console.log("weigh " + c.weigh()); } catch (e) { console.log("threw"); }
