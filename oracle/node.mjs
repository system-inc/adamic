// node.mjs: run an Adamic program on Node, the oracle every native binary is compared against.
//
//	node --disable-warning=ExperimentalWarning oracle/node.mjs main.a [arguments...]
//
// It runs the source itself, with its types stripped, and never anything Adamic lowered: an oracle
// built from Adamic's own output would agree with Adamic's bugs. Both .ts and .a load the same way,
// and 'adamic' resolves to the runtime beside this file.
import { readFileSync } from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';

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
			return { format: 'module', source: stripTypeScriptTypes(source), shortCircuit: true };
		}
		return nextLoad(url, context);
	},
});

// The program sees argv as if Node ran it directly, [node, program, ...arguments]: this file takes its
// own place out, so programArguments() is process.argv.slice(2) here too.
const program = process.argv[2];
process.argv.splice(1, 1);

await import(runtimeUrl);
await import(pathToFileURL(program).href);
