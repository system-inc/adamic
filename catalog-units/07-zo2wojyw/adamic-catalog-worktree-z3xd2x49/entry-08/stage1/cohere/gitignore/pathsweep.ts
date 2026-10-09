// A driver for path.ts alone: it reads the paths file named by its one argument, one path a line with a
// backslash and a newline written \\ and \n, and prints clean, base and dir of each, tab separated, so
// path_test.go can hold them to Go's path.Clean, path.Base and path.Dir.

import { panic, programArguments, readTextFile } from 'adamic';
import { base, clean, dir } from './path.ts';

// unescape reads a line's escapes.
function unescape(line: string): string {
	let text = '';
	let runStart = 0;
	for (let index = 0; index < line.length; index++) {
		if (line.charCodeAt(index) === 0x5c) {
			const letter = line.slice(index + 1, index + 2);
			text += line.slice(runStart, index) + (letter === 'n' ? '\n' : letter);
			index++;
			runStart = index + 1;
		}
	}
	return text + line.slice(runStart);
}

const pathsFile = programArguments()[0] ?? panic('usage: pathsweep.ts <paths file>');
const read = readTextFile(pathsFile);
if (read.kind === 'Error') {
	panic(read.message);
}
const lines = read.text.split('\n');
lines.pop();
const output: string[] = [];
for (const line of lines) {
	const path = unescape(line);
	output.push(`${clean(path)}\t${base(path)}\t${dir(path)}`);
}
console.log(output.join('\n'));
