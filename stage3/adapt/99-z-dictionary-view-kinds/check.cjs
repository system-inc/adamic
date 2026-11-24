'use strict';
// node check.cjs TREE: stock TypeScript 6.0.3 over src/compiler reports the same diagnostics on the adapted tree and
// on the same tree with both before owners put back, isJsonEqual's and loadModuleFromExports', read through the
// compiler host so nothing is copied. Diagnostics compare by file and message, since the lines below each owner move
// by its growth. It also prints the type stock tsc gives each value the after owners cast, which must be object, the
// type the checked dictionary view takes (internal/lower/cast_dictionary_view.go).
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {owner, pins} = require('./adapt.cjs');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node check.cjs TREE');
const tree = path.resolve(process.argv[2]);

const restored = new Map();
for (const pin of pins.owners) {
    const file = path.join(tree, pin.file), adapted = fs.readFileSync(file, 'utf8'), found = owner(adapted, pin);
    assert.equal(found.state, 'after', `check: the tree must hold ${pin.owner}'s after owner`);
    restored.set(path.resolve(file), adapted.slice(0, found.start) + pin.before + adapted.slice(found.end));
}

function measure(replaced) {
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => { throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')); },
    });
    const options = {...config.options, noEmit: true, composite: false, declaration: false, declarationMap: false, emitDeclarationOnly: false, incremental: false, tsBuildInfoFile: undefined};
    const host = ts.createCompilerHost(options);
    const readFile = host.readFile.bind(host);
    host.readFile = name => replaced.get(path.resolve(name)) ?? readFile(name);
    const program = ts.createProgram(config.fileNames, options, host);
    const diagnostics = ts.getPreEmitDiagnostics(program).map(d =>
        `${d.file ? path.relative(tree, d.file.fileName) : ''} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`).sort();
    // Every `as MapLike<unknown>` inside the two owners, with the type of the value it casts.
    const checker = program.getTypeChecker(), casts = [];
    for (const pin of pins.owners) {
        const source = program.getSourceFile(path.join(tree, pin.file));
        ts.forEachChild(source, function visit(node) {
            if (ts.isAsExpression(node) && node.type.getText(source) === 'MapLike<unknown>') {
                for (let at = node.parent; at; at = at.parent) {
                    if (ts.isFunctionDeclaration(at) && at.name?.text === pin.owner && at.parent === source) {
                        casts.push(`${pin.owner}: ${node.expression.getText(source)} is ${checker.typeToString(checker.getTypeAtLocation(node.expression))}`);
                        break;
                    }
                }
            }
            ts.forEachChild(node, visit);
        });
    }
    return {roots: config.fileNames.length, diagnostics, casts};
}
const before = measure(restored), after = measure(new Map());
console.log(`check: before: ${before.roots} roots, ${before.diagnostics.length} diagnostics`);
console.log(`check: after: ${after.roots} roots, ${after.diagnostics.length} diagnostics`);
for (const line of after.diagnostics.slice(0, 20)) console.log(`check:   after ${line}`);
for (const line of after.casts) console.log(`check:   after cast ${line}`);
const same = JSON.stringify(before.diagnostics) === JSON.stringify(after.diagnostics);
console.log(`check: ${same ? 'PASS' : 'FAIL'} stock diagnostics are the same with the before owners and the after ones`);
// isJsonEqual's four and loadModuleFromExports' four, 2626's, 2629's, 2630's and 2637's.
const objects = after.casts.length === 8 && after.casts.every(line => line.endsWith(' is object'));
console.log(`check: ${objects ? 'PASS' : 'FAIL'} the after owners' ${after.casts.length} casts each take a value typed object`);
process.exit(same && objects ? 0 : 1);
