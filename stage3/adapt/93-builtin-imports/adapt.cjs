'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const manifest = require('./proofs/targets.json');
const targets = [
    {local: '_fs', module: 'fs', declaration: 'const _fs: typeof import("fs") = require("fs");', uses: 19},
    {local: '_path', module: 'path', declaration: 'const _path: typeof import("path") = require("path");', uses: 6},
    {local: '_os', module: 'os', declaration: 'const _os = require("os");', uses: 2},
];
function hash(text) { return crypto.createHash('sha256').update(text).digest('hex'); }
function validateBindings(program, file, done) {
    const checker = program.getTypeChecker();
    const functions = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(functions.length, 1, 'getNodeSystem moved');
    const imports = file.statements.filter(ts.isImportDeclaration);
    const evidence = [], declarations = [], lines = new Set();
    for (const target of targets) {
        const locals = functions[0].body.statements.filter(ts.isVariableStatement).filter(s => s.declarationList.declarations.some(d => d.name.getText(file) === target.local));
        const existing = imports.filter(n => n.moduleSpecifier.text === `node:${target.module}`);
        let binding;
        if (done) {
            assert.equal(locals.length, 0, 'builtin local remains');
            assert.equal(existing.length, 1, 'builtin import moved');
            assert.equal(existing[0].getText(file), `import * as ${target.local} from "node:${target.module}";`);
            binding = existing[0].importClause.namedBindings.name;
        } else {
            assert.equal(existing.length, 0, 'builtin already imported');
            assert.equal(locals.length, 1, 'builtin local moved');
            const statement = locals[0], list = statement.declarationList;
            assert.equal(list.declarations.length, 1, 'builtin declaration list moved');
            assert.equal(list.flags & ts.NodeFlags.Const, ts.NodeFlags.Const);
            assert.equal(statement.getText(file), target.declaration, 'builtin declaration moved');
            binding = list.declarations[0].name;
            declarations.push(statement);
        }
        const symbol = checker.getSymbolAtLocation(binding);
        assert(symbol, 'builtin binding has no symbol');
        const identifiers = nodes(ts, file, n => ts.isIdentifier(n) && n.text === target.local);
        const uses = identifiers.filter(n => n !== binding);
        for (const use of uses) {
            // Every occurrence outside the sole declaration must be a qualified
            // receiver of that same binding. This also excludes any competing
            // declaration, at module scope or in an enclosing/nested scope.
            assert(ts.isPropertyAccessExpression(use.parent) && use.parent.expression === use,
                `collision or bare namespace use: ${target.local}`);
            assert.equal(checker.getSymbolAtLocation(use), symbol, `captured or shadowed use: ${target.local}`);
            lines.add(file.getLineAndCharacterOfPosition(use.getStart(file)).line + 1);
        }
        assert.equal(uses.length, target.uses, 'qualified use count moved');
        evidence.push({local: target.local, module: `node:${target.module}`, qualifiedUses: uses.length, bareUses: 0, otherDeclarations: 0, allUsesResolveToBinding: true});
    }
    assert.equal(lines.size, 22, 'qualified use line count moved');
    return {declarations, evidence, qualifiedUses: 27, qualifiedUseLines: lines.size};
}
function plannedProgram(program, file, text) {
    const host = ts.createCompilerHost(program.getCompilerOptions(), true);
    const original = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, shouldCreateNewSourceFile) => path.resolve(name) === path.resolve(file.fileName)
        ? ts.createSourceFile(name, text, language, true)
        : original(name, language, onError, shouldCreateNewSourceFile);
    return ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
}
// a-prime: construction belongs to this whole getNodeSystem owner.
const writerWrapper = `        const patchedWriteFile: System["writeFile"] = (path, data, writeBom) =>
            writeFileEnsuringDirectories(
                path,
                data,
                !!writeBom,
                (path, data, writeByteOrderMark) => writeFile(path, data, writeByteOrderMark),
                path => nodeSystem.createDirectory(path),
                path => nodeSystem.directoryExists(path),
            );
`;
function rewriteWriter(raw, restoring = false) {
    const text = raw, wrapper = raw.includes('\r\n') ? writerWrapper.replace(/\n/g, '\r\n') : writerWrapper;
    const file = ts.createSourceFile('sys.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(file.parseDiagnostics.length, 0, 'writer parse error');
    const owner = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(owner.length, 1, 'writer getNodeSystem count');
    const stores = nodes(ts, owner[0], n => ts.isVariableDeclaration(n) && n.name.getText(file) === 'nodeSystem');
    assert.equal(stores.length, 1); assert(ts.isObjectLiteralExpression(stores[0].initializer));
    const props = stores[0].initializer.properties.filter(n => n.name?.getText(file) === 'writeFile'); assert.equal(props.length, 1);
    const declarations = nodes(ts, owner[0], n => ts.isVariableDeclaration(n) && n.name.getText(file) === 'patchedWriteFile');
    const local = nodes(ts, owner[0], n => ts.isFunctionDeclaration(n) && n.name?.text === 'writeFile'); assert.equal(local.length, 1);
    assert.equal(nodes(ts, local[0], n => n.kind === ts.SyntaxKind.ThisKeyword).length, 0, 'raw writer reads this');
    assert.deepEqual(nodes(ts, local[0], ts.isCallExpression).map(n => n.expression.getText(file)), ['_fs.openSync', '_fs.writeSync', '_fs.closeSync'], 'raw writer callees drift');
    const done = declarations.length === 1; let after = text;
    if (done) {
        assert.equal(text.split(wrapper).length, 2, 'writer wrapper drift');
        assert.equal(props[0].getText(file), 'writeFile: patchedWriteFile', 'writer slot drift');
        if (restoring) after = text.replace(wrapper, '').replace('            writeFile: patchedWriteFile,', '            writeFile,');
    } else {
        assert.equal(declarations.length, 0); assert(ts.isShorthandPropertyAssignment(props[0]));
        if (!restoring) {
            const statement = stores[0].parent.parent, start = text.lastIndexOf('\n', statement.getStart(file)-1)+1;
            after = text.slice(0,start)+wrapper+text.slice(start);
            assert.equal(after.split('            writeFile,').length,2);
            after = after.replace('            writeFile,','            writeFile: patchedWriteFile,');
        }
    }
    return after;
}
const constructWriter = text => rewriteWriter(text);
const restoreWriter = text => rewriteWriter(text, true);
function adapt(tree, options = {}) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts);
    const file = program.getSourceFile(path.join(tree, manifest[0].file));
    assert(file && !file.parseDiagnostics.length, 'cannot parse sys.ts');
    const owners = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(owners.length, 1, 'getNodeSystem moved');
    const owner = owners[0], digest = hash(owner.getText(file));
    assert([manifest[0].ownerBeforeSha256, manifest[0].ownerAfterSha256].includes(digest), 'target owner moved: getNodeSystem');
    const done = digest === manifest[0].ownerAfterSha256;
    assert(!options.pinOnly || !done, "pin generation requires predecessor state");
    const importAnchor = source => {
        const anchors = source.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === './hostErrors.js');
        assert.equal(anchors.length, 1, 'hostErrors import anchor moved');
        return anchors[0];
    };
    const importText = source => [importAnchor(source), ...source.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text.startsWith('node:'))].map(n => n.getText(source)).join('\n');
    assert.equal(hash(importText(file)), done ? manifest[0].importsAfterSha256 : manifest[0].importsBeforeSha256, 'import block moved');
    const checked = validateBindings(program, file, done);
    const imports = file.statements.filter(ts.isImportDeclaration);
    const edits = [];
    if (!done) {
        // Remove each entire statement and its own indentation/newline; keep
        // every surrounding byte and every qualified access unchanged.
        for (const statement of checked.declarations) {
            const start = file.text.lastIndexOf('\n', statement.getStart(file) - 1) + 1;
            const end = file.text.indexOf('\n', statement.end) + 1;
            assert.equal(file.text.slice(start, statement.getStart(file)), '        ');
            assert(end > statement.end, 'builtin statement newline moved');
            edits.push({start, end, text: ''});
        }
        assert(imports.length, 'existing imports moved');
        const at = importAnchor(file).end;
        const newline = file.text.includes('\r\n') ? '\r\n' : '\n';
        edits.push({start: at, end: at, text: targets.map(t => `${newline}import * as ${t.local} from "node:${t.module}";`).join('')});
    }
    // 995spev: explicit concrete forwarding closures, in this whole owner.
    if (!done) for (const property of nodes(ts, owner, n => ts.isShorthandPropertyAssignment(n) && ['setTimeout', 'clearTimeout'].includes(n.name.text))) {
        const replacement = property.name.text === 'setTimeout'
            ? 'setTimeout: (callback: (label?: unknown) => void, ms: number, label?: unknown): unknown => setTimeout(callback, ms, label)'
            : 'clearTimeout: (timeoutId: Parameters<typeof clearTimeout>[0]): void => { clearTimeout(timeoutId); }';
        edits.push({start:property.getStart(file),end:property.end,text:replacement});
    }
    const insertion = importAnchor(file).end;
    const newline = file.text.includes('\r\n') ? '\r\n' : '\n';
    let insertions = 0;
    for (const edit of edits) {
        const inOwner = edit.start >= owner.getStart(file) && edit.start < owner.end && edit.end >= edit.start && edit.end <= owner.end;
        const atImports = edit.start === insertion && edit.end === insertion && edit.text === targets.map(t => `${newline}import * as ${t.local} from "node:${t.module}";`).join('');
        assert(inOwner || atImports, 'edit outside getNodeSystem and import insertion');
        if (atImports) insertions++;
    }
    assert.equal(insertions, done ? 0 : 1, 'builtin import insertion count');
    let text = file.text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    if (!done) text = constructWriter(text);
    const changed = ts.createSourceFile(file.fileName, text, ts.ScriptTarget.Latest, true);
    const changedOwners = nodes(ts, changed, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(changedOwners.length, 1);
    if (!options.pinOnly) assert.equal(hash(changedOwners[0].getText(changed)), manifest[0].ownerAfterSha256, 'unreviewed builtin owner change');
    assert.equal(hash(importText(changed)), manifest[0].importsAfterSha256, 'unreviewed import change');
    assert.equal(ts.createSourceFile(file.fileName,text,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    const planned = plannedProgram(program, file, text);
    validateBindings(planned, planned.getSourceFile(file.fileName), true);
    // Validate all bindings and the full planned bytes before the sole write.
    assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'source changed while planning');
    if (options.pinOnly) {
        const pins = [{...manifest[0], ownerBeforeSha256:digest, ownerAfterSha256:hash(changedOwners[0].getText(changed))}];
        fs.writeFileSync(path.join(__dirname,'proofs/targets.json'),JSON.stringify(pins,null,2)+'\n');
        return {pinsRegenerated:true,sourceWrites:0};
    }
    if (edits.length) fs.writeFileSync(file.fileName,text);
    return {files: edits.length ? 1 : 0, javascriptChanging: true, sites: checked.evidence, qualifiedUses: checked.qualifiedUses, qualifiedUseLines: checked.qualifiedUseLines};
}
module.exports = {adapt, targets, validateBindings, plannedProgram, constructWriter, restoreWriter, writerWrapper};
if (require.main === module) {
    if (process.argv[2] === '--pin') { assert.equal(process.argv.length,4); console.log(JSON.stringify(adapt(path.resolve(process.argv[3]),{pinOnly:true}))); process.exit(0); }
    assert.equal(process.argv.length,3,'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
