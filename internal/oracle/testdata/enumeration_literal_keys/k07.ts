class Point {
	x = 1;
	get norm(): number {
		return 5;
	}
}
const instance = new Point();
const spread = { ...instance, y: 2 };
console.log(Object.keys(spread).join(","));
