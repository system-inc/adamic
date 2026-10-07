// A WASI Preview 1 command, with the same cwd and arguments as the source oracle.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { WASI } from 'node:wasi';

const [path, ...args] = process.argv.slice(2);
const wasi = new WASI({ version: 'preview1', args: [path, ...args],
	env: {}, preopens: { '.': process.cwd() }, returnOnExit: true });
const module = await WebAssembly.compile(await readFile(path));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
// A canary in the reserved bottom of the linear stack catches corruption even if the
// eventual panic text happens to match Node after a missing stack check.
const low = instance.exports.__stack_low.value;
new Uint8Array(instance.exports.memory.buffer, low, 32).fill(0xa5);
process.exitCode = wasi.start(instance);
assert.deepEqual(Array.from(new Uint8Array(instance.exports.memory.buffer, low, 32)),
	Array(32).fill(0xa5), 'the reserved linear-stack margin was overwritten');
