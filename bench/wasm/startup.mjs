import { readFileSync, openSync, closeSync } from 'node:fs';
import { WASI } from 'node:wasi';
import { performance } from 'node:perf_hooks';
import { pathToFileURL } from 'node:url';
import { resolve } from 'node:path';

const [binary, javascript, source] = process.argv.slice(2);
const bytes = readFileSync(binary);
const descriptor = openSync('/dev/null', 'w');
const samples = { compile: [], initialize: [], importJS: [], importSource: [] };
const paths = { importJS: javascript, importSource: source };
const originalLog = console.log;
// Suppress program output symmetrically. Only timing JSON leaves this process.
console.log = () => {};
for (let round = 0; round < 30; round++) {
    const tasks = ['wasm', 'importJS', 'importSource'];
    const offset = round % tasks.length;
    for (const task of [...tasks.slice(offset), ...tasks.slice(0, offset)]) {
        if (task === 'wasm') {
            let started = performance.now();
            const module = await WebAssembly.compile(bytes);
            samples.compile.push(performance.now() - started);
            // WASI host construction is excluded from instantiate/entry timing.
            const wasi = new WASI({ version: 'preview1', args: [binary], env: process.env,
                preopens: { '/': '/' }, returnOnExit: true, stdout: descriptor, stderr: descriptor });
            started = performance.now();
            const instance = await WebAssembly.instantiate(module, wasi.getImportObject());
            if (instance.exports._start) {
                if (wasi.start(instance) !== 0) throw new Error('WASI command failed');
            } else {
                wasi.initialize(instance);
            }
            samples.initialize.push(performance.now() - started);
        } else {
            const url = pathToFileURL(resolve(paths[task]));
            url.searchParams.set('round', String(round));
            const started = performance.now();
            await import(url.href);
            samples[task].push(performance.now() - started);
        }
    }
}
console.log = originalLog;
closeSync(descriptor);
console.log(JSON.stringify(samples));
