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
    const digest = crypto.createHash('sha256').update(file.text).digest('hex');
    assert([manifest[0].beforeSha256, manifest[0].afterSha256].includes(digest), 'target source moved: sys.ts');
    const system = nodes(ts, file, n => ts.isInterfaceDeclaration(n) && n.name.text === 'System');
    assert.equal(system.length, 1, 'System interface moved');
    const edits = [], evidence = [];
    for (const target of targets) {
        const signatures = system[0].members.filter(n => n.name?.getText(file) === target.name);
        assert.equal(signatures.length, 1);
        assert.equal(signatures[0].getText(file), target.signature, 'System signature changed');
        const methods = nodes(ts, file, n => ts.isMethodDeclaration(n) && n.name.getText(file) === target.name);
        assert.equal(methods.length, 1, 'Node system method moved');
        const method = methods[0];
        assert.equal(method.parameters.length, 0);
        assert(method.body && method.body.statements.length === 1 && ts.isReturnStatement(method.body.statements[0]), 'method body moved');
        const expression = method.body.statements[0].expression;
        assert(expression && [target.before,target.after].includes(expression.getText(file)), 'returned expression moved');
        const location = file.getLineAndCharacterOfPosition(method.getStart(file));
        evidence.push({name:target.name,line:location.line+1,before:target.before,after:target.after});
        if (expression.getText(file) === target.before) edits.push({start:expression.getStart(file),end:expression.end,text:target.after});
    }
    let text = file.text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    assert.equal(crypto.createHash('sha256').update(text).digest('hex'),manifest[0].afterSha256,'unreviewed body change');
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
