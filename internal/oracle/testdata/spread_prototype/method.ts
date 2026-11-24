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
const scale = new Scale();
const value = { ...scale };
console.log(value.label);
