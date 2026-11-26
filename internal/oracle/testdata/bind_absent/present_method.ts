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
const bound = built.weigh.bind(built);
console.log(String(bound()));
