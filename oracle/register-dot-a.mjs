// Shared Node 24 source loader for Adamic, used by --import and node.mjs.
import { readFileSync } from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath } from 'node:url';

const runtimeUrl = new URL('./adamic.mjs', import.meta.url).href;

registerHooks({
	resolve(specifier, context, nextResolve) {
		if (specifier === 'adamic') {
			return { url: runtimeUrl, shortCircuit: true };
		}
		return nextResolve(specifier, context);
	},
	load(url, context, nextLoad) {
		if (url.endsWith('.a') || url.endsWith('.ts')) {
			const source = readFileSync(fileURLToPath(url), 'utf8');
			return { format: 'module', source: stripTypeScriptTypes(source, { mode: 'transform', sourceUrl: url }), shortCircuit: true };
		}
		return nextLoad(url, context);
	},
});

// Install runtime error and stream handling even when the entry does not import adamic.
await import(runtimeUrl);
