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
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts);
    const file = program.getSourceFile(path.join(tree, manifest[0].file));
    assert(file && !file.parseDiagnostics.length, 'cannot parse sys.ts');
    const digest = hash(file.text);
    assert([manifest[0].beforeSha256, manifest[0].afterSha256].includes(digest), 'target source moved: sys.ts');
    const done = digest === manifest[0].afterSha256;
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
        const at = imports[imports.length - 1].end;
        const newline = file.text.includes('\r\n') ? '\r\n' : '\n';
        edits.push({start: at, end: at, text: targets.map(t => `${newline}import * as ${t.local} from "node:${t.module}";`).join('')});
    }
    let text = file.text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    assert.equal(hash(text), manifest[0].afterSha256, 'unreviewed builtin import change');
    assert.equal(ts.createSourceFile(file.fileName,text,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    const planned = plannedProgram(program, file, text);
    validateBindings(planned, planned.getSourceFile(file.fileName), true);
    // Validate all bindings and the full planned bytes before the sole write.
    assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'source changed while planning');
    if (edits.length) fs.writeFileSync(file.fileName,text);
    return {files: edits.length ? 1 : 0, javascriptChanging: true, sites: checked.evidence, qualifiedUses: checked.qualifiedUses, qualifiedUseLines: checked.qualifiedUseLines};
}
module.exports = {adapt, targets, validateBindings, plannedProgram};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
