// Transpile source syntax only. The parser is loaded from the adapted source,
// never from stock typescript's createSourceFile or Adamic's emitted output.
import { existsSync, readFileSync } from 'node:fs';
import { registerHooks, createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
const require = createRequire(import.meta.url);
const ts = require(process.env.PARSER_TYPESCRIPT);
const runtime = pathToFileURL(process.env.PARSER_RUNTIME).href;
registerHooks({
    resolve(specifier, context, next) {
        if(specifier === 'adamic') return { url: runtime, shortCircuit: true };
        if(specifier.startsWith('.') && specifier.endsWith('.js')) {
            const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);
            if(existsSync(fileURLToPath(url))) return { url: url.href, shortCircuit: true };
        }
        return next(specifier, context);
    },
    load(url, context, next) {
        if(url.endsWith('.ts') || url.endsWith('.a')) {
            const source = readFileSync(fileURLToPath(url), 'utf8');
            return { format: 'module', shortCircuit: true, source: ts.transpileModule(source, {
                fileName: fileURLToPath(url).replace(/\.a$/, '.ts'),
                compilerOptions: { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true },
            }).outputText };
        }
        return next(url, context);
    },
});
const program = process.argv[2];
process.argv.splice(1, 1);
await import(runtime);
if(process.env.PARSER_TRACE_ERRORS === '1') process.removeAllListeners('uncaughtException');
await import(pathToFileURL(program).href);
