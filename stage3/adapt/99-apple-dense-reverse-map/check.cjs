'use strict';
// node check.cjs TREE: stock TypeScript 6.0.3 over src/compiler reports the same diagnostics on the adapted tree and
// on the same tree with makeReverseMap's before owner put back, read through the compiler host so nothing is copied.
// Diagnostics compare by file and message; scanner.ts's lines below the owner move by the owner's growth. It also
// prints the type each reverse map's reader returns, which the widened result must leave as it was.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {owner, pins} = require('./adapt.cjs');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node check.cjs TREE');
const tree = path.resolve(process.argv[2]);
const scanner = path.join(tree, pins.file);
const adapted = fs.readFileSync(scanner, 'utf8');
const found = owner(adapted);
assert.equal(found.state, 'after', 'check: the tree must hold the dense owner');
const restored = adapted.slice(0, found.start) + pins.before + adapted.slice(found.end);

function measure(scannerText) {
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => { throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')); },
    });
    const options = {...config.options, noEmit: true, composite: false, declaration: false, declarationMap: false, emitDeclarationOnly: false, incremental: false, tsBuildInfoFile: undefined};
    const host = ts.createCompilerHost(options);
    const readFile = host.readFile.bind(host);
    host.readFile = name => path.resolve(name) === scanner ? scannerText : readFile(name);
    const program = ts.createProgram(config.fileNames, options, host);
    const diagnostics = ts.getPreEmitDiagnostics(program).map(d =>
        `${d.file ? path.relative(tree, d.file.fileName) : ''} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`).sort();
    const checker = program.getTypeChecker(), source = program.getSourceFile(scanner);
    const returns = {};
    ts.forEachChild(source, function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.body && ['tokenToString', 'regularExpressionFlagToCharacterCode'].includes(node.name?.text)) {
            const statement = node.body.statements[0];
            returns[node.name.text] = checker.typeToString(checker.getTypeAtLocation(statement.expression));
        }
        ts.forEachChild(node, visit);
    });
    return {roots: config.fileNames.length, diagnostics, returns};
}
const before = measure(restored), after = measure(adapted);
console.log(`check: before: ${before.roots} roots, ${before.diagnostics.length} diagnostics, reads ${JSON.stringify(before.returns)}`);
console.log(`check: after: ${after.roots} roots, ${after.diagnostics.length} diagnostics, reads ${JSON.stringify(after.returns)}`);
for (const line of after.diagnostics.slice(0, 20)) console.log(`check:   after ${line}`);
const same = JSON.stringify(before.diagnostics) === JSON.stringify(after.diagnostics);
console.log(`check: ${same ? 'PASS' : 'FAIL'} stock diagnostics are the same with the before owner and the dense one`);
process.exit(same ? 0 : 1);
