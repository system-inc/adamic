// utilities.ts:8142: a parameter, read in the default of a switch over every member and passed to
// assertNever.
const enum AccessKind {
	Read,
	Write,
	ReadWrite,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function reverseAccessKind(a: AccessKind): AccessKind {
	switch (a) {
		case AccessKind.Read:
			return AccessKind.Write;
		case AccessKind.Write:
			return AccessKind.Read;
		case AccessKind.ReadWrite:
			return AccessKind.ReadWrite;
		default:
			return assertNever(a);
	}
}
function outside(n: number): AccessKind {
	return n as AccessKind;
}
console.log(reverseAccessKind(AccessKind.Read) + " " + reverseAccessKind(AccessKind.Write) + " " + reverseAccessKind(outside(2)));
