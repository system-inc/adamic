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
function copy(source: Weighs): Weighs {
	return { ...source };
}
const value = copy(new Scale());
try {
	console.log(String(value.weigh()));
} catch {
	console.log("threw");
}
