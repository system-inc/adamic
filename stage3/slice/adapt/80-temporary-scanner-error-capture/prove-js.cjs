const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.SLICE_TYPESCRIPT);
const [beforeFile, afterFile, destination] = process.argv.slice(2);
if (!destination) throw new Error('usage: prove-js.cjs BEFORE_DEBUG AFTER_DEBUG NEW_OUTPUT');
fs.mkdirSync(destination);
const before = fs.readFileSync(beforeFile, 'utf8');
const after = fs.readFileSync(afterFile, 'utf8');
const emit = source => ts.transpileModule(source, {
    fileName: 'debug.ts',
    compilerOptions: { target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true },
}).outputText;
const oldJavaScript = emit(before);
const newJavaScript = emit(after);
const mutantSource = after.replace('captureStackTrace(e,', 'captureStackTrace(new Error(),');
if (mutantSource === after) throw new Error('capture-call mutant did not change source');
const mutantJavaScript = emit(mutantSource);
for (const [name, source] of [['before', oldJavaScript], ['after', newJavaScript], ['mutant', mutantJavaScript]]) {
    fs.writeFileSync(path.join(destination, `${name}.mjs`), source);
}
if (oldJavaScript !== newJavaScript) throw new Error('cast removal changes emitted JavaScript');
if (newJavaScript === mutantJavaScript) throw new Error('capture-call mutant was not caught');
console.log(JSON.stringify({ typescript: ts.version, identical: true, bytes: Buffer.byteLength(newJavaScript),
    sha256: crypto.createHash('sha256').update(newJavaScript).digest('hex'), captureCallMutantCaught: true }));
