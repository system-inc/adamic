import 'node:process';
type AnyFunction = (...args: never[]) => void;
function assertNever(member: never): never {
	throw new Error("assertNever reached with " + typeof member);
}
function fail(stackCrawlMark: AnyFunction): void {
	Error.captureStackTrace({}, stackCrawlMark);
	stackCrawlMark();
}
try {
	fail(assertNever);
} catch (caught) {
	console.log("caught " + (caught instanceof Error ? caught.message : "other"));
}
