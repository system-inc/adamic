import 'node:process';
type AnyFunction = (...args: never[]) => void;
function fail(message?: string, stackCrawlMark?: Function): never {
	const e = new Error(message ? "Debug Failure. " + message : "Debug Failure.");
	if (Error.captureStackTrace) {
		Error.captureStackTrace(e, stackCrawlMark || fail);
	}
	throw e;
}
function assertNever(member: never, message = "Illegal value:", stackCrawlMark?: Function): never {
	return fail(message + " " + typeof member, stackCrawlMark || assertNever);
}
function kind(value: string | number): string {
	if (typeof value === "string") return "s";
	if (typeof value === "number") return "n";
	return assertNever(value);
}
console.log(kind("a") + kind(1));
