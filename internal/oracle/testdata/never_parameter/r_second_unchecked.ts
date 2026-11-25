function label(text: string, x: never): string {
	return text + typeof x;
}
function fail(): never {
	throw new Error("fail");
}
function caller(v: string | number): string {
	if (typeof v === "string") return "s";
	if (typeof v === "number") return "n";
	return label(v, fail());
}
console.log(caller("a") + caller(1));
