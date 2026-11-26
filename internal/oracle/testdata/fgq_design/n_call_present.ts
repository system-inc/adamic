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
const value = copy({ label: "x", weigh: () => 5 });
try {
	console.log(String(value.weigh()));
} catch {
	console.log("threw");
}
