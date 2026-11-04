'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const manifest = require('./proofs/targets.json');
const targets = [
    {name: 'getWidthOfTerminal', before: 'process.stdout.columns', after: 'process.stdout.columns ?? 0', signature: 'getWidthOfTerminal?(): number;'},
    {name: 'writeOutputIsTTY', before: 'process.stdout.isTTY', after: '!!process.stdout.isTTY', signature: 'writeOutputIsTTY?(): boolean;'},
];
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts);
    const file = program.getSourceFile(path.join(tree, manifest[0].file));
    assert(file && !file.parseDiagnostics.length, 'cannot parse sys.ts');
    const hash = text => crypto.createHash('sha256').update(text).digest('hex');
    const owners = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(owners.length, 1, 'getNodeSystem moved');
    const spans = [], states = [];
    const system = nodes(ts, file, n => ts.isInterfaceDeclaration(n) && n.name.text === 'System');
    assert.equal(system.length, 1, 'System interface moved');
    const edits = [], evidence = [];
    for (const target of targets) {
        const signatures = system[0].members.filter(n => n.name?.getText(file) === target.name);
        assert.equal(signatures.length, 1);
        assert.equal(signatures[0].getText(file), target.signature, 'System signature changed');
        const methods = nodes(ts, file, n => ts.isMethodDeclaration(n) && n.name.getText(file) === target.name);
        assert.equal(methods.length, 1, 'Node system method moved');
        const method = methods[0], pin = manifest[0].owners.find(row => row.name === target.name);
        assert(method.getStart(file) >= owners[0].getStart(file) && method.end <= owners[0].end, 'method outside getNodeSystem');
        assert(pin, 'missing owner pin');
        assert([pin.beforeSha256, pin.afterSha256].includes(hash(method.getText(file))), 'target owner moved: ' + target.name);
        states.push(hash(method.getText(file)) === pin.afterSha256 ? 'after' : 'before');
        spans.push(method);
        assert.equal(method.parameters.length, 0);
        assert(method.body && method.body.statements.length === 1 && ts.isReturnStatement(method.body.statements[0]), 'method body moved');
        const expression = method.body.statements[0].expression;
        assert(expression && [target.before,target.after].includes(expression.getText(file)), 'returned expression moved');
        const location = file.getLineAndCharacterOfPosition(method.getStart(file));
        evidence.push({name:target.name,line:location.line+1,before:target.before,after:target.after});
        if (expression.getText(file) === target.before) edits.push({start:expression.getStart(file),end:expression.end,text:target.after});
    }
    assert.equal(new Set(states).size, 1, 'partially applied TTY owners');
    for (const edit of edits) assert(spans.some(owner => edit.start >= owner.getStart(file) && edit.start < owner.end && edit.end >= edit.start && edit.end <= owner.end), 'edit outside TTY owners');
    let text = file.text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    const changed = ts.createSourceFile(file.fileName, text, ts.ScriptTarget.Latest, true);
    const changedOwners = nodes(ts, changed, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(changedOwners.length, 1);
    for (const pin of manifest[0].owners) {
        const methods = nodes(ts, changedOwners[0], n => ts.isMethodDeclaration(n) && n.name.getText(changed) === pin.name);
        assert.equal(methods.length, 1);
        assert.equal(hash(methods[0].getText(changed)), pin.afterSha256, 'unreviewed body change');
    }
    assert.equal(ts.createSourceFile(file.fileName,text,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    // Validate both targets and the unchanged interface before the sole write.
    assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'source changed while planning');
    if (edits.length) fs.writeFileSync(file.fileName,text);
    return {files:edits.length ? 1 : 0,javascriptChanging:true,sites:evidence};
}
module.exports = {adapt,targets};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
