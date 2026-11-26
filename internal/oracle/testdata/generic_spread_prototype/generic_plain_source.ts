interface Weighs {
	readonly label: string;
	readonly weigh: () => number;
}
function copy<T extends Weighs>(source: T): T {
	return { ...source };
}
const value = copy({ label: "x", weigh: () => 5 });
console.log(value.label + " " + String(value.weigh()));
