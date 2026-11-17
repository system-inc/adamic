'use strict';
// node check.cjs BEFORE AFTER: two complete applies, before without this adaptation and after with it.
// (a) Every file under src is byte-identical except src/compiler/checker.ts, and checker.ts is the before tree's with
//     only the pinned owner replaced.
// (b) Stock TypeScript 6.0.3 over the 79 src/compiler roots reports the same diagnostics on both trees, and the owner's
//     inferred return type is the same: the new statements narrow to what the old expression returned.
// Each verdict prints PASS or FAIL; the exit is 0 only when both pass.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {locate, owners} = require('./adapt.cjs');
assert.equal(process.argv.length, 4, 'usage: node check.cjs BEFORE AFTER');
const [before, after] = process.argv.slice(2).map(tree => path.resolve(tree));
assert.equal(ts.version, '6.0.3');
const pin = owners[0];
let failed = 0;
function verdict(ok, line) {
    console.log(`check: ${ok ? 'PASS' : 'FAIL'} ${line}`);
    if (!ok) failed++;
}

function files(tree) {
    const found = [];
    (function walk(directory) {
        for (const entry of fs.readdirSync(directory, {withFileTypes: true})) {
            const full = path.join(directory, entry.name);
            if (entry.isDirectory()) walk(full);
            else if (entry.isFile()) found.push(path.relative(tree, full));
        }
    })(path.join(tree, 'src'));
    return found.sort();
}
const beforeFiles = files(before), afterFiles = files(after);
const differing = [...new Set([...beforeFiles, ...afterFiles])].filter(name => {
    const a = path.join(before, name), b = path.join(after, name);
    return !fs.existsSync(a) || !fs.existsSync(b) || !fs.readFileSync(a).equals(fs.readFileSync(b));
}).sort();
const beforeChecker = fs.readFileSync(path.join(before, pin.file), 'utf8');
const afterChecker = fs.readFileSync(path.join(after, pin.file), 'utf8');
const beforeOwner = locate(beforeChecker), afterOwner = locate(afterChecker);
const restored = afterChecker.slice(0, afterOwner.start) + pin.before + afterChecker.slice(afterOwner.end);
verdict(
    JSON.stringify(differing) === JSON.stringify([pin.file]) && !beforeOwner.adapted && afterOwner.adapted && restored === beforeChecker,
    `(a) sources: ${afterFiles.length} files under src, differing ${JSON.stringify(differing)}; before owner adapted ${beforeOwner.adapted}, ` +
        `after owner adapted ${afterOwner.adapted}; checker.ts with the before owner put back is the before file: ${restored === beforeChecker}`,
);

function measure(tree) {
    const roots = files(tree).filter(name => name.startsWith('src/compiler/') && name.endsWith('.ts') && !name.endsWith('.d.ts')).map(name => path.join(tree, name));
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => {
            throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'));
        },
    });
    const options = {...config.options, noEmit: true, composite: false, declaration: false, declarationMap: false, emitDeclarationOnly: false, incremental: false, tsBuildInfoFile: undefined};
    const program = ts.createProgram(roots, options);
    const diagnostics = ts.getPreEmitDiagnostics(program).map(d => {
        const where = d.file ? `${path.relative(tree, d.file.fileName)}:${d.file.getLineAndCharacterOfPosition(d.start).line + 1}` : '';
        return `${where} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`;
    });
    const checker = program.getTypeChecker(), source = program.getSourceFile(path.join(tree, pin.file));
    let returned = [];
    ts.forEachChild(source, function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === pin.owner) {
            returned.push(checker.typeToString(checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(node))));
        }
        ts.forEachChild(node, visit);
    });
    return {roots: roots.length, diagnostics, returned};
}
const measured = {before: measure(before), after: measure(after)};
for (const [name, result] of Object.entries(measured)) {
    console.log(`check: ${name}: ${result.roots} roots, ${result.diagnostics.length} diagnostics, ${pin.owner} returns ${JSON.stringify(result.returned)}`);
    for (const line of result.diagnostics.slice(0, 20)) console.log(`check:   ${name} ${line}`);
}
// checker.ts's lines below the owner move by the owner's growth, so its diagnostics compare by message alone.
const shape = list => list.map(line => line.replace(/^(\S+?):\d+ /, '$1 ')).sort();
verdict(
    JSON.stringify(shape(measured.before.diagnostics)) === JSON.stringify(shape(measured.after.diagnostics)) &&
        measured.before.returned.length === 1 && JSON.stringify(measured.before.returned) === JSON.stringify(measured.after.returned),
    `(b) types: ${measured.before.roots} and ${measured.after.roots} roots, diagnostics ${measured.before.diagnostics.length} and ${measured.after.diagnostics.length}, ` +
        `return type ${JSON.stringify(measured.before.returned)} and ${JSON.stringify(measured.after.returned)}`,
);
process.exitCode = failed ? 1 : 0;
