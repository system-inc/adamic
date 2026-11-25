function reject(v: never): string {
	return typeof v;
}
function kind(v: boolean | string): string {
	return typeof v === "string" ? v : typeof v === "boolean" ? "bool" : reject(v);
}
console.log(kind("hello") + " " + kind(true));
