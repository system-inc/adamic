interface Weighs {
	readonly label: string;
	readonly weigh: () => number;
}
class Scale {
	readonly label = "t";
	weigh(): number {
		return 20;
	}
}
const built: Weighs = new Scale();
const value: Weighs = { ...built };
try {
	const bound = value.weigh.bind(value);
	console.log(String(bound()));
} catch {
	console.log("threw");
}
