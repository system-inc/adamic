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
function same<T extends Weighs>(source: T): T {
	return source;
}
console.log(String(same(new Scale()).weigh()));
