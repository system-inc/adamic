// adamic.mjs: the Adamic runtime for the Node oracle. Only panic so far.
//
// A panic writes one line to stderr and exits 70. It sets process.exitCode rather than calling
// process.exit, because stdout to a pipe is asynchronous on macOS and process.exit would drop what
// the program already printed. 0.1 has no try, so nothing catches the throw.
class AdamicPanic extends Error {}

process.on('uncaughtException', (error) => {
	const message = error instanceof AdamicPanic ? error.message : String(error);
	process.stderr.write(`adamic: panic: ${message}\n`);
	process.exitCode = 70;
});

export function panic(message) {
	throw new AdamicPanic(message);
}
