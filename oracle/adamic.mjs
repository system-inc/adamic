// adamic.mjs: the Adamic runtime for the Node oracle: panic, and the files and arguments 0.2 opens.
//
// A panic writes one line to stderr and exits 70. It sets process.exitCode rather than calling
// process.exit, because stdout to a pipe is asynchronous on macOS and process.exit would drop what
// the program already printed. It throws to stop the program, and a catch in the source could take
// that throw, where natively a panic ends the program on the spot; so from the panic on, everything
// the program writes is dropped and the exit is 70 whatever it does, which makes what it does after
// unseen, as it is natively (docs/memory.md, "Exceptions").
import { readFileSync, writeFileSync } from 'node:fs';

class AdamicPanic extends Error {}

// A second failure while the first is being reported (stderr closed under it, say) ends the program
// at once. Without this, the write fails, that failure is another uncaught exception, whose report
// fails the same way, forever, at full speed: two orphaned processes once ran for half an hour so.
let panicking = false;

// panicked is a panic already reported, by panic itself, whether or not something caught its throw.
let panicked = false;

process.on('uncaughtException', (error) => {
	if (panicked) {
		// Reported when it happened; whatever was thrown after it is unseen.
		process.exitCode = 70;
		return;
	}
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
	if (!panicked) {
		panicked = true;
		process.stderr.write(`adamic: panic: ${message}\n`);
		process.exitCode = 70;
		// What's already written goes out; nothing after it does.
		for (const stream of [process.stdout, process.stderr]) {
			stream.write = () => true;
		}
		process.on('exit', () => {
			process.exitCode = 70;
		});
	}
	throw new AdamicPanic(message);
}

// readTextFile reads a file as Node's readFileSync(path, 'utf8') decodes it: invalid bytes become
// U+FFFD by the WHATWG decoder's rules, and a byte-order mark stays, as U+FEFF. A failure is a value,
// since there are no exceptions to throw yet, and its message is Adamic's own words, the same on
// every platform, where Node's own wording isn't. The native runtime says the same (input.c).
export function readTextFile(path) {
	try {
		return { kind: 'Ok', text: readFileSync(path, 'utf8') };
	} catch (error) {
		return { kind: 'Error', message: `cannot read ${path}: ${failure(error.code, false)}` };
	}
}

// writeTextFile writes a file as Node's writeFileSync(path, text) does: made if it isn't there,
// emptied if it is, the text as UTF-8 with a lone surrogate written as U+FFFD.
export function writeTextFile(path, text) {
	if (panicked) {
		// After a panic nothing the program does is seen, a file it writes included.
		return { kind: 'Ok' };
	}
	try {
		writeFileSync(path, text);
		return { kind: 'Ok' };
	} catch (error) {
		return { kind: 'Error', message: `cannot write ${path}: ${failure(error.code, true)}` };
	}
}

// failure is the reason in Adamic's words. When a write can't find its path, what's missing is a
// directory, not the file.
function failure(code, writing) {
	switch (code) {
		case 'ENOENT':
		case 'ENOTDIR':
			return writing ? 'no such directory' : 'no such file';
		case 'EACCES':
		case 'EPERM':
			return 'permission denied';
		case 'EISDIR':
			return 'is a directory';
	}
	return 'failed';
}

// programArguments is the arguments after the program, as process.argv.slice(2) is when Node runs the
// program itself (node.mjs takes its own place out of argv first). A new array every call.
export function programArguments() {
	return process.argv.slice(2);
}
