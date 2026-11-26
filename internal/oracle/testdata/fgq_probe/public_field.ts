interface Weighs { weigh(): number }
class Scale implements Weighs {
	constructor(readonly grams: number, readonly label: string) {}
	weigh(): number { return this.grams * 2 + this.label.length; }
}
function copy<T extends Weighs>(source: T): T { return { ...source }; }
const c = copy(new Scale(10, "kg"));
try { console.log("weigh " + c.weigh()); } catch (e) { console.log("threw"); }
