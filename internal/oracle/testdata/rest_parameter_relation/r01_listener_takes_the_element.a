class Logger {
	onError(seen: Error): void { console.log(`logged ${seen.message}`); }
}
interface Handler {
	onError(...args: [error: Error]): void;
}
const logger = new Logger();
const handler: Handler = logger;
handler.onError(new Error(`boom${String(1)}`));
