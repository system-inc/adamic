// The namespace oracle executes TypeScript 6.0.3's own emit, including const-enum erasure.
import { createRequire, registerHooks } from 'node:module';
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { homedir } from 'node:os';
import { fileURLToPath, pathToFileURL } from 'node:url';
const require = createRequire(import.meta.url);
const ts = require(process.env.ADAMIC_TYPESCRIPT_RUNTIME || require.resolve('typescript', {paths: [resolve(dirname(fileURLToPath(import.meta.url)), '../api'), resolve(homedir(), '.cache/adamic-stage3/api')]}));
if (ts.version !== '6.0.3') throw new Error(`namespace oracle needs TypeScript 6.0.3, found ${ts.version}`);
const options = {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force, allowNonTsExtensions: true, noEmitOnError: false};
const host = ts.createCompilerHost(options);
const originalSource = host.getSourceFile.bind(host);
host.getSourceFile = (name, target, error, fresh) => name.endsWith('.a') ? ts.createSourceFile(name, readFileSync(name, 'utf8'), target, true, ts.ScriptKind.TS) : originalSource(name, target, error, fresh);
host.resolveModuleNames = (names, from) => names.map(name => name.endsWith('.a') ? {resolvedFileName: resolve(dirname(from), name), extension: ts.Extension.Ts} : ts.resolveModuleName(name, from, options, host).resolvedModule);
const emitted = new Map();
host.writeFile = (name, text, bom, error, sources) => { for (const source of sources || []) emitted.set(resolve(source.fileName), text); };
const entry = resolve(process.argv[2]);
const program = ts.createProgram([entry], options, host);
const result = program.emit();
if (result.emitSkipped || !emitted.has(entry)) throw new Error('TypeScript namespace emit was skipped');
const runtime = new URL('../../oracle/adamic.mjs', import.meta.url).href;
registerHooks({
 resolve(name, context, next) { return name === 'adamic' ? {url: runtime, shortCircuit:true} : next(name, context); },
 load(url, context, next) { if (url.endsWith('.a')) { const text = emitted.get(resolve(fileURLToPath(url))); if (text === undefined) throw new Error(`missing TypeScript emit for ${url}`); return {format:'module', source:text, shortCircuit:true}; } return next(url, context); }
});
await import(runtime);
await import(pathToFileURL(entry).href);
