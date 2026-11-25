function bump(x: never): string { x ||= x; return "bump"; }
function caller(v: string | number): string {
	if (typeof v === "string") return "s";
	if (typeof v === "number") return "n";
	return bump(v);
}
console.log(caller("a") + caller(1));
