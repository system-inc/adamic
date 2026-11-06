class Logger {
	onError(seen: TypeError): void { console.log(`logged ${seen.message}`); }
}
interface Handler {
	onError(...args: [error: Error]): void;
}
const logger = new Logger();
const handler: Handler = logger;
handler.onError(new TypeError(`boom${String(1)}`));
