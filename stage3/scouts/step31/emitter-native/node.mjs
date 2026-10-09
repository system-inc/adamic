// Stock TypeScript transpilation only; no Adamic output participates in the truth.
import {readFileSync, existsSync} from 'node:fs';
import {registerHooks, createRequire} from 'node:module';
import {fileURLToPath, pathToFileURL} from 'node:url';
import {resolve} from 'node:path';
const require = createRequire(import.meta.url);
const ts = require(process.env.STEP31_TYPESCRIPT);
const runtimeUrl = new URL('../../../../oracle/adamic.mjs', import.meta.url).href;
registerHooks({
    resolve(specifier, context, next) {
        if (specifier === 'adamic') return {url: runtimeUrl, shortCircuit:true};
        if (specifier.startsWith('.') && specifier.endsWith('.js')) {
            const url = new URL(specifier.slice(0,-3)+'.ts', context.parentURL);
            if (existsSync(fileURLToPath(url))) return {url:url.href,shortCircuit:true};
        }
        return next(specifier,context);
    },
    load(url,context,next) {
        if (url.endsWith('.a') || url.endsWith('.ts')) return {format:'module',shortCircuit:true,
            source:ts.transpileModule(readFileSync(fileURLToPath(url),'utf8'),{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true}}).outputText};
        return next(url,context);
    }
});
const entry = process.argv[2];
process.argv.splice(1,1);
await import(runtimeUrl);
await import(pathToFileURL(resolve(entry)).href);
