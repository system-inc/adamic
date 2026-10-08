// Independent stock transpilation of a real compiler source tree for source mutants.
import {readFileSync, existsSync} from 'node:fs';
import {registerHooks, createRequire} from 'node:module';
import {fileURLToPath, pathToFileURL} from 'node:url';
import {resolve} from 'node:path';
const require = createRequire(import.meta.url);
const stock = require(process.env.STEP31_TYPESCRIPT);
const {dump} = require('./binder-dump.cjs');
registerHooks({
    resolve(specifier, context, nextResolve) {
        if (specifier.startsWith('.') && specifier.endsWith('.js')) {
            const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);
            if (existsSync(fileURLToPath(url))) return {url: url.href, shortCircuit: true};
        }
        return nextResolve(specifier, context);
    },
    load(url, context, nextLoad) {
        if (url.endsWith('.ts')) return {format: 'module', shortCircuit: true,
            source: stock.transpileModule(readFileSync(fileURLToPath(url), 'utf8'), {
                fileName: fileURLToPath(url), compilerOptions: {target: stock.ScriptTarget.ESNext,
                    module: stock.ModuleKind.ESNext, verbatimModuleSyntax: true}}).outputText};
        return nextLoad(url, context);
    }
});
const api = await import(pathToFileURL(resolve(process.argv[2], 'src/compiler/_namespaces/ts.ts')).href);
if (process.argv[3] === '--probes') {
    const {probes} = require('./piece-probes.cjs');
    console.log(JSON.stringify(probes(api), null, 2));
} else {
    process.stdout.write(dump(api, JSON.parse(readFileSync(process.argv[3], 'utf8'))));
}
