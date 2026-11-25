function assertNever(value: never): never {
	throw new Error("assertNever reached with " + typeof value);
}
let seen = "none";
Error.prepareStackTrace = (_error, frames) => {
	for (const frame of frames) {
		const fn = frame.getFunction();
		if (fn !== undefined) seen = "a function";
	}
	return "";
};
function fail(stackCrawlMark?: Function): string {
	const holder: { stack?: string } = {};
	Error.captureStackTrace(holder, stackCrawlMark || assertNever);
	return String(holder.stack);
}
fail();
console.log("getFunction saw " + seen);
