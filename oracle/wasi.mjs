import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';

const binary = process.argv[2];
const wasi = new WASI({ version: 'preview1', args: [binary, ...process.argv.slice(3)], env: { ...process.env, ADAMIC_WASI_CWD: process.cwd() }, preopens: { '/': '/' }, returnOnExit: true });
const module = await WebAssembly.compile(readFileSync(binary));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
process.exitCode = wasi.start(instance);
