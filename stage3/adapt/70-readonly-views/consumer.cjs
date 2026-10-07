#!/usr/bin/env node
'use strict';
// A downstream compatibility witness against the emitted public declarations.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
if (process.argv.length !== 4) throw new Error('usage: node consumer.cjs <before.d.ts> <after.d.ts>');
const consumer = `import type * as ts from './published';
type View = NonNullable<Parameters<typeof ts.setTextRange>[1]>;
function adjust(view: View): void { view.pos = 1; view.end = 2; }
void adjust;
`;
const root = path.resolve('/readonly-view-consumer'), source = path.join(root, 'consumer.ts'), published = path.join(root, 'published.d.ts');
function check(file) {
    const options = { strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
        verbatimModuleSyntax: true, erasableSyntaxOnly: true, noEmit: true,
        module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
        target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: [] };
    const host = ts.createCompilerHost(options);
    const read = host.readFile, exists = host.fileExists;
    host.readFile = name => name === source ? consumer : name === published ? fs.readFileSync(file, 'utf8') : read(name);
    host.fileExists = name => name === source || name === published || exists(name);
    const getSourceFile = host.getSourceFile;
    host.getSourceFile = (name, languageVersion, onError, fresh) => name === source || name === published ?
        ts.createSourceFile(name, host.readFile(name), languageVersion, true) : getSourceFile(name, languageVersion, onError, fresh);
    host.resolveModuleNames = names => names.map(name => name === './published' ?
        { resolvedFileName: published, extension: ts.Extension.Dts } : undefined);
    const program = ts.createProgram([source, published], options, host);
    return ts.getPreEmitDiagnostics(program).map(d => ({ code: d.code,
        where: d.file && `${path.basename(d.file.fileName)}:${d.file.getLineAndCharacterOfPosition(d.start).line + 1}:${d.file.getLineAndCharacterOfPosition(d.start).character + 1}`,
        message: ts.flattenDiagnosticMessageText(d.messageText, '\n') }));
}
const before = check(path.resolve(process.argv[2])), after = check(path.resolve(process.argv[3]));
if (before.length || after.length !== 2 || after.some(d => d.code !== 2540 || !d.where?.startsWith('consumer.ts:')))
    throw new Error(`unexpected downstream witness: ${JSON.stringify({ before, after })}`);
console.log(JSON.stringify({ consumer, before, after }, null, 2));
