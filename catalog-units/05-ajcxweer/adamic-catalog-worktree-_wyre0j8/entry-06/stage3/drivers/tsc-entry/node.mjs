// Stock TypeScript transpiles the original source, following scanner's loader.
import { readFileSync, existsSync } from 'node:fs';
import { registerHooks, createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
const require = createRequire(import.meta.url);
const ts = require(process.env.SCANNER_TYPESCRIPT);
// Upstream's bundled CLI has these CommonJS host globals.
globalThis.require = require;
globalThis.module = { exports: {} };
globalThis.__filename = process.argv[2];
globalThis.__dirname = fileURLToPath(new URL('.', pathToFileURL(process.argv[2])));
registerHooks({
    resolve(specifier, context, nextResolve) {
        if (specifier.startsWith('.') && specifier.endsWith('.js')) {
            const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);
            if (existsSync(fileURLToPath(url))) return { url: url.href, shortCircuit: true };
        }
        return nextResolve(specifier, context);
    },
    load(url, context, nextLoad) {
        if (url.endsWith('.ts') || url.endsWith('.a')) {
            const output = ts.transpileModule(readFileSync(fileURLToPath(url), 'utf8'), {
                fileName: fileURLToPath(url).replace(/\.a$/, '.ts'),
                compilerOptions: { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true },
            });
            return { format: 'module', source: output.outputText, shortCircuit: true };
        }
        return nextLoad(url, context);
    },
});
const program = process.argv[2];
process.argv.splice(1, 1);
await import(pathToFileURL(program).href);
