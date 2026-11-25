function stop(x: never, mark?: Function): never {
	throw new Error(String(mark || stop));
}
function caller(v: string | number): string {
	if (typeof v === "string") return "s";
	if (typeof v === "number") return "n";
	return stop(v);
}
console.log(caller("a") + caller(1));
