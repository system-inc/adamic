function reject(v: never): string {
	return typeof v;
}
function kind(v: "s" | "n"): string {
	switch (v) {
		case "s": return "string";
		case "n": return "number";
		default: return reject(v);
	}
}
console.log(kind("s") + " " + kind("n"));
