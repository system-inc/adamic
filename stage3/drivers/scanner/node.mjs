// Run the same driver against upstream source, transpiled by stock TypeScript.
// No Adamic compiler output participates in the oracle.
import { readFileSync, existsSync } from 'node:fs';
import { registerHooks } from 'node:module';
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
const require = createRequire(import.meta.url);
const ts = require(process.env.SCANNER_TYPESCRIPT);
const runtime = pathToFileURL(process.env.SCANNER_RUNTIME).href;
registerHooks({
    resolve(specifier, context, nextResolve) {
        if (specifier === 'adamic') return { url: runtime, shortCircuit: true };
        if (specifier.startsWith('.') && specifier.endsWith('.js')) {
            const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);
            if (existsSync(fileURLToPath(url))) return { url: url.href, shortCircuit: true };
        }
        return nextResolve(specifier, context);
    },
    load(url, context, nextLoad) {
        if (url.endsWith('.ts') || url.endsWith('.a')) {
            const source = readFileSync(fileURLToPath(url), 'utf8');
            const output = ts.transpileModule(source, {
                fileName: fileURLToPath(url).replace(/\.a$/, '.ts'),
                compilerOptions: { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext,
                    verbatimModuleSyntax: true },
            });
            return { format: 'module', source: output.outputText, shortCircuit: true };
        }
        return nextLoad(url, context);
    },
});
const program = process.argv[2];
process.argv.splice(1, 1);
await import(runtime);
try {
    // Upstream initializes its compiler through the barrel entry. Starting
    // the cyclic ESM graph at scanner.ts instead hits parser's load-time read.
    const barrel = new URL('./adapted/src/compiler/_namespaces/ts.ts', pathToFileURL(program));
    if (existsSync(fileURLToPath(barrel))) await import(barrel);
    await import(pathToFileURL(program).href);
} catch (error) {
    console.error(error.stack);
    process.exitCode = 1;
}
