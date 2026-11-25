function never1(x: never): string {
	return "never1 " + typeof x;
}
function caller(v: string | number): string {
	if (typeof v === "string") return "s";
	if (typeof v === "number") return "n";
	return never1(v);
}
console.log(caller("a") + caller(1));
