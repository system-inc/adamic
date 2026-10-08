// Parser manifests print many short lines. Format each call as Node does, then
// write bounded chunks to the test's stdout file. The exit hook also preserves
// lines printed before a failing port exits.
import { writeSync } from 'node:fs';
import { format } from 'node:util';
import { pathToFileURL } from 'node:url';

let pending = '';
function flush() {
    if (pending === '') return;
    const bytes = Buffer.from(pending);
    pending = '';
    for (let offset = 0; offset < bytes.length;) {
        offset += writeSync(1, bytes, offset, bytes.length - offset);
    }
}
console.log = (...args) => {
    pending += format(...args) + '\n';
    if (pending.length >= 65536) flush();
};
process.once('exit', flush);
const runner = process.argv[2];
process.argv.splice(1, 1);
try {
    await import(pathToFileURL(runner).href);
} finally {
    flush();
}
