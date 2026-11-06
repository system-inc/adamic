class Logger {
	onError(seen: { text: string; code: number }): void { console.log(`${seen.text} ${String(seen.code)}`); }
}
interface Handler {
	onError(...args: [event: { text: string }]): void;
}
const logger = new Logger();
const handler: Handler = logger;
handler.onError({ text: `boom${String(1)}` });
