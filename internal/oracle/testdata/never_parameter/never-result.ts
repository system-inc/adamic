function never1(x: never): string {
	return "never1 " + typeof x;
}
function fail(): never {
	throw new Error("fail");
}
try {
	console.log(never1(fail()));
} catch {
	console.log("fail");
}
