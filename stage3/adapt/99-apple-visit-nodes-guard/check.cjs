'use strict';
// node check.cjs TREE: stock TypeScript 6.0.3 over src/compiler reports no diagnostics on the adapted tree, and the
// same none with visitArrayWorker's and visitNodes' before owners put back. TREE/node_modules must hold tsc's own
// install (stage3/typescript's), for @types/node; the restored owners are read through the compiler host so nothing
// is copied. It also prints the type the checker gives each new check's subject, which must be what the assertion
// it replaces was handed.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {state, pins} = require('./adapt.cjs');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node check.cjs TREE');
const tree = path.resolve(process.argv[2]);
const visitor = path.join(tree, pins.file);
const adapted = fs.readFileSync(visitor, 'utf8');
const source = ts.createSourceFile(visitor, adapted, ts.ScriptTarget.Latest, true);
const owners = ['visitArrayWorker', 'visitNodes'].map(name => state(source, name));
assert(owners.every(owner => owner.state === 'after'), 'check: the tree must hold both adapted owners');
let restored = adapted;
for (const {owner} of [...owners].sort((left, right) => right.owner.getStart(source) - left.owner.getStart(source))) {
    restored = restored.slice(0, owner.getStart(source)) + pins.owners[owner.name.text].before + restored.slice(owner.end);
}

function measure(visitorText) {
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => { throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')); },
    });
    // declaration stays on, since the config's isolatedDeclarations asks for it; noEmit writes nothing either way.
    const options = {...config.options, noEmit: true, composite: false, declaration: true, declarationMap: false, emitDeclarationOnly: false, incremental: false, tsBuildInfoFile: undefined};
    const host = ts.createCompilerHost(options);
    const readFile = host.readFile.bind(host);
    host.readFile = name => path.resolve(name) === visitor ? visitorText : readFile(name);
    const program = ts.createProgram(config.fileNames, options, host);
    const diagnostics = ts.getPreEmitDiagnostics(program).map(d =>
        `${d.file ? path.relative(tree, d.file.fileName) : ''} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`).sort();
    const checker = program.getTypeChecker(), file = program.getSourceFile(visitor);
    const subjects = [];
    const visited = file.statements.filter(node => ts.isFunctionDeclaration(node) && ['visitArrayWorker', 'visitNodes'].includes(node.name?.text) && node.body);
    assert.equal(visited.length, 2, 'check: the two owners');
    visited.forEach(function visit(node) {
        const call = ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && node.expression.getText(file);
        if (call === 'Debug.assertEachNode' || call === 'Debug.assertNode') subjects.push(`${node.arguments[0].getText(file)}: ${checker.typeToString(checker.getTypeAtLocation(node.arguments[0]))}`);
        if (ts.isForOfStatement(node) && ts.isVariableDeclarationList(node.initializer) && node.initializer.declarations[0]?.name.getText(file) === 'testedNode')
            subjects.push(`${node.expression.getText(file)}: ${checker.typeToString(checker.getTypeAtLocation(node.expression))}`);
        // A single node's check, `test !== undefined && !test(X)`.
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'test' && node.arguments.length === 1 &&
            ts.isPrefixUnaryExpression(node.parent) && ts.isBinaryExpression(node.parent.parent))
            subjects.push(`${node.arguments[0].getText(file)}: ${checker.typeToString(checker.getTypeAtLocation(node.arguments[0]))}`);
        ts.forEachChild(node, visit);
    });
    return {roots: config.fileNames.length, diagnostics, subjects};
}
const before = measure(restored), after = measure(adapted);
console.log(`check: before: ${before.roots} roots, ${before.diagnostics.length} diagnostics, subjects ${JSON.stringify(before.subjects)}`);
console.log(`check: after: ${after.roots} roots, ${after.diagnostics.length} diagnostics, subjects ${JSON.stringify(after.subjects)}`);
for (const line of after.diagnostics.slice(0, 20)) console.log(`check:   after ${line}`);
const same = after.diagnostics.length === 0 && JSON.stringify(before.diagnostics) === JSON.stringify(after.diagnostics);
console.log(`check: ${same ? 'PASS' : 'FAIL'} stock tsc reports no diagnostics over src/compiler, with the before owners and the adapted ones`);
process.exit(same ? 0 : 1);
