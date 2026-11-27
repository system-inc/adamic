interface Weighs {
	readonly label: string;
	weigh(grams: number): number;
}
class Scale {
	readonly label = "t";
	weigh(grams: number): number {
		return grams;
	}
}
const built: Weighs = new Scale();
const value: Weighs = { ...built };
function note(): number {
	console.log("note");
	return 1;
}
try {
	const bound = value.weigh.bind(value, note());
	console.log(String(bound()));
} catch {
	console.log("threw");
}
