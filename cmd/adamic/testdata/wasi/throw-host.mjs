import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';

const wasi = new WASI({ version: 'preview1', args: [], env: {}, preopens: {}, returnOnExit: false });
const module = await WebAssembly.compile(readFileSync(process.argv[2]));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
wasi.initialize(instance);
const api = instance.exports;
const bytes = new TextEncoder().encode('boundary');
const pointer = api.malloc(bytes.length);
new Uint8Array(api.memory.buffer, pointer, bytes.length).set(bytes);
api.adamic_request(pointer, bytes.length);
throw new Error('uncaught handler exception returned across the ABI');
