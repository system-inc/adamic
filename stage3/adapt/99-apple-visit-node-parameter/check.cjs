'use strict';
// node check.cjs BEFORE AFTER: two complete applies, before without this adaptation and after with it. Each tree's
// node_modules must hold tsc's own install (stage3/typescript's), for @types/node.
// (a) Every file under src is byte-identical except src/compiler/visitorPublic.ts, and visitorPublic.ts is the before
//     tree's with only the pinned owner replaced.
// (b) Stock TypeScript 6.0.3 over src/compiler reports the same diagnostics on both trees, and the implementation's
//     node parameter is `Node` before and `Node | undefined` after.
// (c) visitorPublic.ts transpiles to byte-identical JavaScript on both trees, at ES2020 and ESNext, with and without
//     comments: the edit is type-only.
// (d) visitorPublic.ts's declaration emit is byte-identical on both trees, with and without stripInternal: an
//     overloaded function's implementation signature is never emitted, so no declaration moves.
// Each verdict prints PASS or FAIL; the exit is 0 only when all four pass.
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
// The owner is located on LF text, as adapt.cjs locates it; a CRLF tree compares the same after the replacement.
const beforeRaw = fs.readFileSync(path.join(before, pin.file), 'utf8'), afterRaw = fs.readFileSync(path.join(after, pin.file), 'utf8');
const beforeText = beforeRaw.replace(/\r\n/g, '\n'), afterText = afterRaw.replace(/\r\n/g, '\n');
console.log(`check: visitorPublic.ts line endings: before ${/\r\n/.test(beforeRaw) ? 'CRLF' : 'LF'}, after ${/\r\n/.test(afterRaw) ? 'CRLF' : 'LF'}`);
const beforeOwner = locate(beforeText), afterOwner = locate(afterText);
const restored = afterText.slice(0, afterOwner.start) + pin.before + afterText.slice(afterOwner.end);
verdict(
    JSON.stringify(differing) === JSON.stringify([pin.file]) && !beforeOwner.adapted && afterOwner.adapted && restored === beforeText,
    `(a) sources: ${afterFiles.length} files under src, differing ${JSON.stringify(differing)}; before owner adapted ${beforeOwner.adapted}, ` +
        `after owner adapted ${afterOwner.adapted}; visitorPublic.ts with the before owner put back is the before file: ${restored === beforeText}`,
);

function measure(tree) {
    const config = ts.getParsedCommandLineOfConfigFile(path.join(tree, 'src/compiler/tsconfig.json'), {noEmit: true}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => {
            throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'));
        },
    });
    // declaration stays on, since the config's isolatedDeclarations asks for it; noEmit writes nothing either way.
    const options = {...config.options, noEmit: true, composite: false, declaration: true, declarationMap: false, emitDeclarationOnly: false, incremental: false, tsBuildInfoFile: undefined};
    const program = ts.createProgram(config.fileNames, options);
    const diagnostics = ts.getPreEmitDiagnostics(program).map(d => {
        const where = d.file ? `${path.relative(tree, d.file.fileName)}:${d.file.getLineAndCharacterOfPosition(d.start).line + 1}` : '';
        return `${where} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ')}`;
    }).sort();
    const checker = program.getTypeChecker(), source = program.getSourceFile(path.join(tree, pin.file));
    const parameter = source.statements
        .filter(node => ts.isFunctionDeclaration(node) && node.name?.text === pin.owner && node.body)
        .map(node => checker.typeToString(checker.getTypeAtLocation(node.parameters[0])));
    return {roots: config.fileNames.length, diagnostics, parameter};
}
const measured = {before: measure(before), after: measure(after)};
for (const [name, result] of Object.entries(measured)) {
    console.log(`check: ${name}: ${result.roots} roots, ${result.diagnostics.length} diagnostics, ${pin.owner}'s implementation takes node: ${JSON.stringify(result.parameter)}`);
    for (const line of result.diagnostics.slice(0, 20)) console.log(`check:   ${name} ${line}`);
}
// visitorPublic.ts keeps its line count, so diagnostics compare with their lines.
verdict(
    JSON.stringify(measured.before.diagnostics) === JSON.stringify(measured.after.diagnostics) &&
        JSON.stringify(measured.before.parameter) === JSON.stringify(['Node']) && JSON.stringify(measured.after.parameter) === JSON.stringify(['Node | undefined']),
    `(b) types: ${measured.before.roots} and ${measured.after.roots} roots, diagnostics ${measured.before.diagnostics.length} and ${measured.after.diagnostics.length}, ` +
        `node parameter ${JSON.stringify(measured.before.parameter)} and ${JSON.stringify(measured.after.parameter)}`,
);

const javascript = [];
for (const target of [ts.ScriptTarget.ES2020, ts.ScriptTarget.ESNext]) {
    for (const removeComments of [true, false]) {
        const emit = text => ts.transpileModule(text, {fileName: pin.file, compilerOptions: {target, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
        const a = emit(beforeRaw), b = emit(afterRaw);
        javascript.push(a === b);
        if (a !== b) console.log(`check:   JavaScript differs at ${ts.ScriptTarget[target]}, removeComments ${removeComments}`);
    }
}
verdict(javascript.every(Boolean), `(c) JavaScript: ${pin.file} byte-identical at ES2020 and ESNext, with and without comments (${javascript.filter(Boolean).length} of ${javascript.length})`);

const declarations = [];
for (const stripInternal of [true, false]) {
    const emit = text => ts.transpileDeclaration(text, {fileName: pin.file, compilerOptions: {stripInternal}}).outputText;
    const a = emit(beforeRaw), b = emit(afterRaw);
    declarations.push(a === b);
    const signatures = b.split(/\bfunction visitNode\b/).length - 1;
    console.log(`check: declarations, stripInternal ${stripInternal}: ${signatures} visitNode signatures emitted, the same on both trees: ${a === b}`);
}
verdict(declarations.every(Boolean), `(d) declarations: ${pin.file}'s declaration emit byte-identical, with and without stripInternal`);
process.exitCode = failed ? 1 : 0;
