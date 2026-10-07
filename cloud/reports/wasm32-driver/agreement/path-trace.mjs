import { readFileSync, writeFileSync } from 'node:fs';
import { WASI } from 'node:wasi';
const wasi = new WASI({ version: 'preview1', args: [process.argv[2]], env: process.env, preopens: { '/': '/' }, returnOnExit: true });
const imports = wasi.getImportObject();
const pathOpen = imports.wasi_snapshot_preview1.path_open;
let instance;
const trace = [];
imports.wasi_snapshot_preview1.path_open = (...args) => {
	const path = new TextDecoder().decode(new Uint8Array(instance.exports.memory.buffer, args[2], args[3]));
	const result = pathOpen(...args);
	trace.push({ function: 'path_open', preopen: args[0], path, errno: result });
	return result;
};
instance = await WebAssembly.instantiate(await WebAssembly.compile(readFileSync(process.argv[2])), imports);
const exit = wasi.start(instance);
writeFileSync(process.argv[3], JSON.stringify({ exit, trace }, null, 2) + '\n');
process.exitCode = exit;
