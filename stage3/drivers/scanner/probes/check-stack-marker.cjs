// Stock checker control for the pinned Node constructor-marker declaration.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT);
const [probe, apiDirectory] = process.argv.slice(2);
if (!probe || !apiDirectory) throw new Error('usage: check-stack-marker.cjs PROBE.a PINNED_API_DIRECTORY');
if (ts.version !== '6.0.3') throw new Error('stock TypeScript 6.0.3 required');
const root = path.resolve(apiDirectory, 'marker-typecheck.ts');
const source = fs.readFileSync(probe, 'utf8');
function check(text) {
    const options = {strict: true, noEmit: true, skipLibCheck: true,
        target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext,
        moduleResolution: ts.ModuleResolutionKind.NodeNext, types: ['node'],
        typeRoots: [path.resolve(apiDirectory, 'node_modules/@types')]};
    const host = ts.createCompilerHost(options), original = host.getSourceFile;
    host.getSourceFile = (file, ...args) => file === root
        ? ts.createSourceFile(file, text, ts.ScriptTarget.ESNext, true)
        : original(file, ...args);
    return ts.getPreEmitDiagnostics(ts.createProgram([root], options, host))
        .map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
}
const callable = check(source);
const mutant = source.replace('stackCrawlMark?: AnyFunction', 'stackCrawlMark?: {}');
if (mutant === source) throw new Error('opaque-marker mutant did not change source');
const opaque = check(mutant);
if (callable.length) throw new Error(JSON.stringify(callable));
if (!opaque.some(d => d.code === 2345 && d.message.includes("parameter of type 'Function'"))) throw new Error('opaque marker mutant survived');
console.log(JSON.stringify({typescript: ts.version, callable_diagnostics: callable, opaque_marker_mutant: opaque}, null, 2));
