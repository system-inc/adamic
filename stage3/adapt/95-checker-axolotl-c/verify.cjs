#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const sites = require('./sites.json');
const [beforeTree, afterTree] = process.argv.slice(2);
assert(beforeTree && afterTree, 'usage: verify.cjs BEFORE AFTER');
assert.equal(ts.version,'6.0.3');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const walk = dir => fs.readdirSync(dir,{withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(dir,e.name)) : [path.join(dir,e.name)]).sort();
const files = tree => walk(path.join(tree,'src')).map(f=>path.relative(tree,f));
assert.deepEqual(files(beforeTree),files(afterTree));
const evidence = [];
let changedFiles = 0;
for (const file of files(beforeTree)) {
    const oldBytes = fs.readFileSync(path.join(beforeTree,file)), newBytes = fs.readFileSync(path.join(afterTree,file));
    const rows = sites.filter(row=>row.file===file);
    if (!rows.length) { assert(oldBytes.equals(newBytes),'unowned source changed: '+file); continue; }
    const before = oldBytes.toString(), after = newBytes.toString();
    let expected = before;
    const source = ts.createSourceFile(file,before,ts.ScriptTarget.Latest,true);
    const owners = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) || ts.isInterfaceDeclaration(node)) owners.push(node);
        ts.forEachChild(node,visit);
    }
    visit(source);
    // Independent exact-source reconstruction, without invoking the adapter.
    for (const row of rows) {
        const owner = owners.filter(node=>node.name?.text===row.owner)[row.ownerOccurrence];
        assert(owner,'inventoried owner');
        const oldOwner = owner.getText(source);
        assert.equal(hash(oldOwner),row.ownerBeforeSha256,'before owner pin');
        assert.equal(oldOwner.split(row.before).length,2,'unique before spelling within owner');
        const newOwner = oldOwner.replace(row.before,row.after);
        assert.equal(after.split(newOwner).length,2,'every declaration repair must be present');
        assert.equal(expected.split(oldOwner).length,2,'unique owner');
        expected = expected.replace(oldOwner,newOwner);
    }
    assert.equal(after,expected,'only the inventoried declaration texts may change');
    assert.equal(ts.createSourceFile(file,after,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    const crlf = {before:(before.match(/\r\n/g)||[]).length,after:(after.match(/\r\n/g)||[]).length};
    assert.equal(crlf.before,crlf.after);
    const javascript = [];
    for (const removeComments of [false,true]) {
        const options={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
        const oldJS=ts.transpileModule(before,options).outputText, newJS=ts.transpileModule(after,options).outputText;
        assert.equal(newJS,oldJS,'JavaScript must be byte-identical');
        javascript.push({removeComments,bytes:Buffer.byteLength(newJS),sha256:hash(newJS)});
    }
    evidence.push({file,beforeSha256:hash(before),afterSha256:hash(after),crlf,javascript});
    changedFiles += Number(before!==after);
}
assert.equal(changedFiles,2);
console.log(JSON.stringify({sourceFiles:files(beforeTree).length,changedFiles,declarations:sites.length,evidence},null,2));
