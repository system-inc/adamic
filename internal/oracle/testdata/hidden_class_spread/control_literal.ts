interface Weighs {
	readonly label: string;
	weigh(): number;
}
function copy(source: Weighs): Weighs {
	return { ...source };
}
const value = copy({ label: "x", weigh: () => 5 });
console.log(value.label + " " + String(value.weigh()));
