// adamic.mjs: the Adamic runtime for the Node oracle. Only panic so far.
//
// A panic writes one line to stderr and exits 70. It sets process.exitCode rather than calling
// process.exit, because stdout to a pipe is asynchronous on macOS and process.exit would drop what
// the program already printed. 0.1 has no try, so nothing catches the throw.
class AdamicPanic extends Error {}

// A second failure while the first is being reported (stderr closed under it, say) ends the program
// at once. Without this, the write fails, that failure is another uncaught exception, whose report
// fails the same way, forever, at full speed: two orphaned processes once ran for half an hour so.
let panicking = false;

process.on('uncaughtException', (error) => {
	if (panicking) {
		process.exit(70);
	}
	panicking = true;
	const message = error instanceof AdamicPanic ? error.message : String(error);
	process.stderr.write(`adamic: panic: ${message}\n`);
	process.exitCode = 70;
});

for (const stream of [process.stdout, process.stderr]) {
	stream.on('error', () => process.exit(70));
}

export function panic(message) {
	throw new AdamicPanic(message);
}
