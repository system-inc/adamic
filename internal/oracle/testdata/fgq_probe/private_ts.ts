interface Weighs { weigh(): number }
class Scale implements Weighs {
	private grams: number;
	constructor(grams: number) { this.grams = grams; }
	weigh(): number { return this.grams * 2; }
}
function copy<T extends Weighs>(source: T): T { return { ...source }; }
const c = copy(new Scale(10));
try { console.log("weigh " + c.weigh()); } catch (e) { console.log("threw"); }
