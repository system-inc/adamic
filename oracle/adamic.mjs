// adamic.mjs: the Adamic runtime for the Node oracle: panic, and the files and arguments 0.2 opens.
//
// A panic writes one line to stderr and exits 70. It sets process.exitCode rather than calling
// process.exit, because stdout to a pipe is asynchronous on macOS and process.exit would drop what
// the program already printed. It throws to stop the program, and a catch in the source could take
// that throw, where natively a panic ends the program on the spot; so from the panic on, everything
// the program writes is dropped and the exit is 70 whatever it does, which makes what it does after
// unseen, as it is natively (docs/memory.md, "Exceptions").
import { spawnSync } from 'node:child_process';
import { constants, tmpdir } from 'node:os';
import { mkdtempSync, openSync, closeSync, rmSync, lstatSync, readdirSync, readFileSync, realpathSync, statSync, writeFileSync } from 'node:fs';

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

// utf8Length and utf8At are a string's UTF-8 as TextEncoder writes it, a lone surrogate as U+FFFD's
// three bytes. The last string asked about is kept encoded, since a loop asks about one string byte
// after byte. utf8At panics, in the native runtime's words (utf8.c), where the index isn't a byte.
let utf8Text = '';
let utf8Encoded = Buffer.alloc(0);

export function utf8Length(text) {
	return Buffer.byteLength(text, 'utf8');
}

export function utf8At(text, index) {
	if (text !== utf8Text) {
		utf8Text = text;
		utf8Encoded = Buffer.from(text, 'utf8');
	}
	if (!(index >= 0 && index < utf8Encoded.length && index === Math.trunc(index))) {
		panic(`RangeError: utf8At index ${index} is not a byte of a text of ${utf8Encoded.length} bytes`);
	}
	return utf8Encoded[index];
}

// readDirectory is a directory's names as readdirSync gives them: libuv's scandir, sorted by their
// bytes, without . and .., each decoded as UTF-8. The native runtime sorts and decodes the same way
// (directory.c).
export function readDirectory(path) {
	try {
		return { kind: 'Ok', names: readdirSync(path) };
	} catch (error) {
		return { kind: 'Error', message: `cannot read directory ${path}: ${listingFailure(error.code)}` };
	}
}

function listingFailure(code) {
	switch (code) {
		case 'ENOENT':
			return 'no such directory';
		case 'ENOTDIR':
			return 'not a directory';
		case 'EACCES':
		case 'EPERM':
			return 'permission denied';
	}
	return 'failed';
}

// fileStatus is what a path names, as statSync sees it, following a symbolic link, with whether the
// path is itself one, as lstatSync sees it. A link to nothing is no such file, as statSync says.
export function fileStatus(path) {
	try {
		const link = lstatSync(path);
		const status = link.isSymbolicLink() ? statSync(path) : link;
		const type = status.isFile() ? 'file' : status.isDirectory() ? 'directory' : 'other';
		return { kind: 'Ok', type, size: status.size, symbolicLink: link.isSymbolicLink() };
	} catch (error) {
		return { kind: 'Error', message: `cannot read status of ${path}: ${failure(error.code, false)}` };
	}
}

// programArguments is the arguments after the program, as process.argv.slice(2) is when Node runs the
// program itself (node.mjs takes its own place out of argv first). A new array every call.
export function programArguments() {
	return process.argv.slice(2);
}

// Canonical path identity for directory visitation, following links as the host filesystem does.
export function realPath(path) {
	try {
		return { kind: 'Ok', path: realpathSync(path) };
	} catch (error) {
		return { kind: 'Error', message: `cannot resolve path ${path}: ${failure(error.code, false)}` };
	}
}

// Combined capture uses one descriptor, preserving writes across stdout and stderr and avoiding
// pipe capacity deadlocks. This synchronous door does not schedule parallel children.
export function runProcess(executable, arguments_, directory, environment) {
    const result = { output: '', exitCode: 1, signal: 0, error: '' };
    if (!executable.startsWith('/') || [executable, directory, ...arguments_, ...environment].some(x => x.includes('\0'))) {
        result.error = 'exec:EINVAL';
        return result;
    }
    const env = { ...process.env };
    for (const entry of environment) {
        const equals = entry.indexOf('=');
        if (equals < 0) delete env[entry];
        else env[entry.slice(0, equals)] = entry.slice(equals + 1);
    }
    let temp = '', fd = -1;
    try {
        // Distinguish an unavailable cwd from an unavailable executable, as Go does.
        if (directory !== '') {
            try { if (!statSync(directory).isDirectory()) throw { code: 'ENOTDIR' }; }
            catch (error) { result.error = `chdir:${error.code ?? 'EIO'}`; return result; }
        }
        temp = mkdtempSync(`${tmpdir()}/adamic-process-`);
        fd = openSync(`${temp}/output`, 'w+');
        const child = spawnSync(executable, arguments_, { cwd: directory || undefined, env, stdio: ['ignore', fd, fd] });
        result.output = readFileSync(`${temp}/output`, 'utf8');
        if (child.error) result.error = `exec:${child.error.code ?? 'EIO'}`;
        else if (child.signal) {
            result.signal = constants.signals[child.signal] ?? 0;
            result.error = result.signal === 0 ? 'signal:unsupported' : '';
        } else result.exitCode = child.status;
    } catch (error) { result.error = `capture:${error.code ?? 'EIO'}`; }
    finally { if (fd >= 0) closeSync(fd); if (temp !== '') rmSync(temp, { recursive: true, force: true }); }
    return result;
}
