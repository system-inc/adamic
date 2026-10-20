import { readFileSync } from 'node:fs';
import { wasiShim, AdamicExit } from '../internal/native/wasm/shim.mjs';

const shim = wasiShim({ stdout: bytes => process.stdout.write(bytes), stderr: bytes => process.stderr.write(bytes) });
const module = await WebAssembly.compile(readFileSync(process.argv[2]));
const instance = await WebAssembly.instantiate(module, shim.imports);
shim.attach(instance);
try {
	if (instance.exports._start) instance.exports._start();
	else if (instance.exports._initialize) instance.exports._initialize();
	else throw new Error('Adamic module has neither _start nor _initialize');
} catch (error) {
	if (!(error instanceof AdamicExit)) throw error;
	process.exitCode = error.code;
}
