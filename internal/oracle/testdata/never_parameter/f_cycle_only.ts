import 'node:process';
type AnyFunction = (...args: never[]) => void;
function fail(message?: string, stackCrawlMark?: AnyFunction): never {
	const e = new Error(message ? "Debug Failure. " + message : "Debug Failure.");
	if (message === "loop") relay(e, stackCrawlMark || fail);
	throw e;
}
function relay(e: Error, mark: AnyFunction): void {
	fail(e.message, mark);
}
function assertNever(member: never, message = "Illegal value:", stackCrawlMark?: AnyFunction): never {
	return fail(message + " " + typeof member, stackCrawlMark || assertNever);
}
function kind(value: string | number): string {
	if (typeof value === "string") return "s";
	if (typeof value === "number") return "n";
	return assertNever(value);
}
console.log(kind("a") + kind(1));
