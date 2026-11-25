const Error = {
	captureStackTrace(target: object, constructorOpt?: (value: never) => never): void {
		if (constructorOpt !== undefined) constructorOpt(target as never);
	},
};
function assertNever(value: never): never {
	throw new globalThis.Error("assertNever reached with " + typeof value);
}
function fail(stackCrawlMark?: (value: never) => never): void {
	Error.captureStackTrace({}, stackCrawlMark || assertNever);
}
try {
	fail();
} catch (caught) {
	console.log("caught " + (caught instanceof globalThis.Error ? caught.message : "other"));
}
