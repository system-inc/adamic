// Mirrors oracle/node.mjs loading while supplying only recorded checker dependencies.
import {readFileSync} from 'node:fs';
import {registerHooks, stripTypeScriptTypes} from 'node:module';
import {fileURLToPath, pathToFileURL} from 'node:url';
process.env.ADAMIC_REACT_FACTS = process.argv[2];
const runtime = new URL('./checker_fixture_runtime.mjs', import.meta.url).href;
registerHooks({
    resolve(specifier, context, next) {return specifier === 'adamic' ? {url: runtime, shortCircuit: true} : next(specifier, context);},
    load(url, context, next) {
        if(url.endsWith('.a') || url.endsWith('.ts')) {return {format: 'module', source: stripTypeScriptTypes(readFileSync(fileURLToPath(url), 'utf8')), shortCircuit: true};}
        return next(url, context);
    },
});
const source = process.argv[3];
process.argv.splice(1, 2);
await import(runtime);
await import(pathToFileURL(source).href);
