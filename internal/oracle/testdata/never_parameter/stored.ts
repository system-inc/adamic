function never1(x: never): string {
	return "never1 " + typeof x;
}
const saved = never1;
function caller(v: string | number): string {
	if (typeof v === "string") return "s";
	if (typeof v === "number") return "n";
	return saved(v);
}
console.log(caller("a") + caller(1));
