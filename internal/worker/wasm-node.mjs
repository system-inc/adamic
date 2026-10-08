// Node's platform witness for Workers' CompiledWasm import rule.
import { readFileSync } from 'node:fs';
import { registerHooks } from 'node:module';
import { fileURLToPath } from 'node:url';

export function registerWasm(witness = false) {
	if (witness) globalThis.wasmWitness = { calls: 0, instances: [] };
	registerHooks({
		load(url, context, nextLoad) {
			if (url.endsWith('.wasm')) {
				const bytes = readFileSync(fileURLToPath(url));
				return { format: 'module', source: `export default await WebAssembly.compile(Uint8Array.from(${JSON.stringify(Array.from(bytes))}));`, shortCircuit: true };
			}
			if (witness && url.endsWith('/crossing-runtime.mjs')) {
				let source = readFileSync(fileURLToPath(url), 'utf8');
				// Count only calls reaching an actual Wasm export, and retain instance diagnostics.
				source = source.replace('invoke(name, ...args) {', "invoke(name, ...args) { if (name.startsWith('adamic_export_')) globalThis.wasmWitness.calls++;");
				source = source.replace('next.exports._initialize();', 'globalThis.wasmWitness.instances.push(next); next.exports._initialize();');
				return { format: 'module', source, shortCircuit: true };
			}
			return nextLoad(url, context);
		},
	});
}
