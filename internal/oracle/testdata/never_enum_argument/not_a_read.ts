// A stop: the argument is a call that returns never, no enum's never read and no rulings.md:76 one.
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function fail(): never {
	throw new Error("fail");
}
console.log("before");
assertNever(fail());
