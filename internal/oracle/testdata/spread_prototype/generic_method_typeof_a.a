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
function copy<T extends Weighs>(source: T): T {
	return { ...source };
}
const value = copy(new Scale());
console.log(value.label + " " + typeof value.weigh);
