interface Weighs {
	readonly label: string;
	readonly weigh: () => number;
}
const value: Weighs = { label: "t", weigh: () => 5 };
const bound = value.weigh.bind(value);
console.log(String(bound()));
