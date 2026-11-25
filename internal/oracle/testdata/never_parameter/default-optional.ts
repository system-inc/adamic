function reject(v: never, text = "unused", count?: number): string {
	return text + typeof v + String(count);
}
function kind(v: string | number): string {
	if (typeof v === "string") return "s";
	if (typeof v === "number") return "n";
	return reject(v, "other", 2);
}
console.log(kind("a") + kind(1));
