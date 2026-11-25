import 'node:process';
function assertNever(value: never): never {
	throw new Error("assertNever reached with " + typeof value);
}
Error.captureStackTrace = (target: object, constructorOpt?: Function): void => {
	if (constructorOpt !== undefined) constructorOpt(target);
};
function fail(stackCrawlMark?: Function): void {
	Error.captureStackTrace({}, stackCrawlMark || assertNever);
}
try {
	fail();
} catch (caught) {
	console.log("caught " + (caught instanceof Error ? caught.message : "other"));
}
