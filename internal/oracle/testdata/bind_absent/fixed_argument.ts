interface Weighs {
	readonly label: string;
	weigh(): number;
}
class Scale {
	readonly label = "t";
	weigh(): number {
		return 20;
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
