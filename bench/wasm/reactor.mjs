import { readFileSync } from 'node:fs';
import { WASI } from 'node:wasi';
import { pathToFileURL } from 'node:url';
const binary = process.argv[2];
const wasi = new WASI({ version: 'preview1', args: [binary], env: process.env, returnOnExit: true });
const module = await WebAssembly.compile(readFileSync(binary));
const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
wasi.initialize(instance);
if (process.argv.includes('--check')) {
    const source = await import(new URL('source.mts', pathToFileURL(binary)));
    const exports = instance.exports;
    for (const request of ['test', '世界 🌍', 'a\0b', 'x'.repeat(1024)]) {
        const bytes = new TextEncoder().encode(request);
        const input = exports.malloc(bytes.length);
        new Uint8Array(exports.memory.buffer, input, bytes.length).set(bytes);
        const response = exports.adamic_request(input, bytes.length);
        const actual = new TextDecoder().decode(new Uint8Array(exports.memory.buffer,
            exports.adamic_response_bytes(response), exports.adamic_response_length(response)));
        exports.adamic_release(response);
        exports.free(input);
        if (actual !== source.handleRequest(request)) throw new Error('response differs');
    }
}
