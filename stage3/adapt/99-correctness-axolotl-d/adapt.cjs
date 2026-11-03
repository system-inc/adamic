'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const {helper} = require('../94-axolotl-property-read/adapt.cjs');
function nodes(root, predicate) {
    const found = [];
    (function visit(n) { if (predicate(n)) found.push(n); ts.forEachChild(n, visit); })(root);
    return found;
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const helperText = fs.readFileSync(path.join(tree, 'src/compiler/hostErrors.ts'), 'utf8');
    assert(helperText.endsWith(helper), '94 propertyOf helper missing or changed');
    const fileName = path.join(tree, 'src/compiler/commandLineParser.ts');
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'tryReadFile');
    assert.equal(functions.length, 1, 'tryReadFile moved');
    const fn = functions[0];
    const before = 'return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1, fileName, e.message);';
    const after = 'return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1, fileName, String(propertyOf(e, "message")));';
    const catches = nodes(fn, ts.isCatchClause);
    assert.equal(catches.length, 1);
    assert.equal(catches[0].variableDeclaration?.getText(source), 'e');
    assert.equal(catches[0].block.statements.length, 1);
    const statement = catches[0].block.statements[0];
    assert([before, after].includes(statement.getText(source)), 'unreviewed catch text');
    const imports = source.statements.filter(ts.isImportDeclaration);
    const hosts = imports.filter(n => n.moduleSpecifier.text === './hostErrors.js');
    assert(hosts.length <= 1, 'duplicate hostErrors imports');
    const imported = hosts.length && hosts[0].importClause?.namedBindings;
    if (hosts.length) assert(imported && ts.isNamedImports(imported) && !hosts[0].importClause.isTypeOnly, 'unreviewed hostErrors import');
    const propertyImports = imported ? imported.elements.filter(n => n.name.text === 'propertyOf' || n.propertyName?.text === 'propertyOf') : [];
    assert(propertyImports.length <= 1);
    if (propertyImports.length) assert.equal(propertyImports[0].getText(source), 'propertyOf');
    const done = statement.getText(source) === after;
    assert.equal(done, propertyImports.length === 1, 'half-adapted commandLineParser.ts');
    if (done) return {files: 0, javascriptChanging: true};
    const edits = [{start: statement.getStart(source), end: statement.end, value: after}];
    if (imported) {
        edits.push({start: imported.end - 1, end: imported.end - 1, value: `${imported.elements.length ? ', ' : ''}propertyOf `});
    } else {
        const at = imports[imports.length - 1].end;
        edits.push({start: at, end: at, value: `${text.includes('\r\n') ? '\r\n' : '\n'}import { propertyOf } from "./hostErrors.js";`});
    }
    let result = text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) result = result.slice(0,edit.start)+edit.value+result.slice(edit.end);
    assert.equal(ts.createSourceFile(fileName,result,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    assert.equal(fs.readFileSync(fileName,'utf8'),text,'source changed while planning');
    fs.writeFileSync(fileName,result);
    return {files:1,javascriptChanging:true};
}
module.exports = {adapt};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
