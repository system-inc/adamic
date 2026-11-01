'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('typescript');
const control = path.resolve(process.argv[2]), retyped = path.resolve(process.argv[3]);
assert.equal(ts.version, '6.0.3');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function files(directory, root = directory) {
    return fs.readdirSync(directory, {withFileTypes: true}).flatMap(e => e.isDirectory() ? files(path.join(directory, e.name), root) : [path.relative(root, path.join(directory, e.name))]).sort();
}
const sourceFiles = files(path.join(control, 'src'));
assert.deepEqual(files(path.join(retyped, 'src')), sourceFiles);
const changed = [];
for (const relative of sourceFiles) {
    const a = fs.readFileSync(path.join(control, 'src', relative));
    const b = fs.readFileSync(path.join(retyped, 'src', relative));
    if (!a.equals(b)) {
        assert.equal(relative, 'compiler/tsbuildPublic.ts');
        const before = 'timerToBuildInvalidatedProject: any;';
        assert.equal(a.toString().split(before).length, 2);
        assert.equal(b.toString(), a.toString().replace(before, 'timerToBuildInvalidatedProject: unknown;'));
        changed.push({file: 'src/' + relative, beforeSha256: hash(a), afterSha256: hash(b)});
    }
}
assert.equal(changed.length, 1);
const outputFiles = files(path.join(control, 'built'));
const javascript = [], declarations = [];
for (const relative of outputFiles.filter(f => /\.(js|mjs|cjs)$/.test(f) || f.endsWith('.d.ts'))) {
    const a = fs.readFileSync(path.join(control, 'built', relative));
    const b = fs.readFileSync(path.join(retyped, 'built', relative));
    assert(a.equals(b), 'emitted bytes changed: ' + relative);
    const row = {file: 'built/' + relative, sha256: hash(a), bytes: a.length};
    (/\.(js|mjs|cjs)$/.test(relative) ? javascript : declarations).push(row);
}
assert.deepEqual(files(path.join(retyped, 'built')).filter(f => /\.(js|mjs|cjs)$/.test(f) || f.endsWith('.d.ts')), outputFiles.filter(f => /\.(js|mjs|cjs)$/.test(f) || f.endsWith('.d.ts')));
assert(javascript.length > 0 && declarations.length > 0);
const reference = 'tests/baselines/reference/api/typescript.d.ts';
assert(fs.readFileSync(path.join(control, reference)).equals(fs.readFileSync(path.join(retyped, reference))));
for (const file of ['built/local/typescript.d.ts', 'built/local/typescript.internal.d.ts', reference]) {
    assert(!/\bSolutionBuilderState\b|\btimerToBuildInvalidatedProject\b/.test(fs.readFileSync(path.join(retyped, file), 'utf8')), 'private state leaked into API');
}
const compiler = require(path.join(process.argv[4], 'stage3/adapt/47-host-errors/adapt.cjs')).compiler;
const diagnosticsFor = tree => ts.getPreEmitDiagnostics(compiler(tree, ts)).map(d => {
    const position = d.file && d.file.getLineAndCharacterOfPosition(d.start);
    return {file: d.file && path.relative(tree, d.file.fileName), line: position && position.line + 1, column: position && position.character + 1, code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')};
});
const controlDiagnostics = diagnosticsFor(control), retypedDiagnostics = diagnosticsFor(retyped);
assert.deepEqual(retypedDiagnostics, controlDiagnostics, 'strict helper diagnostics changed');
const diagnostics = {control: controlDiagnostics.length, retyped: retypedDiagnostics.length, identical: true, added: []};
console.log(JSON.stringify({status: 'pass', typescript: ts.version, sourceFilesChecked: sourceFiles.length, changed, compilerDiagnostics: diagnostics, javascript, declarations, publicApiDiff: [], timerStatePublic: false, referenceIdentical: true}, null, 2));
