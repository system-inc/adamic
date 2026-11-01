#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const crypto=require('node:crypto'),ts=require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const manifest=require('./sites.json');
const [beforeTree,afterTree]=process.argv.slice(2);
assert(beforeTree && afterTree,'usage: node verify.cjs BEFORE AFTER');assert.equal(ts.version,'6.0.3');
const hash=t=>crypto.createHash('sha256').update(t).digest('hex');
const walk=d=>fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]).sort();
const files=t=>walk(path.join(t,'src')).filter(f=>f.endsWith('.ts')).map(f=>path.relative(t,f));
assert.deepEqual(files(beforeTree),files(afterTree));let changed=0,evidence;
for(const file of files(beforeTree)){
    const before=fs.readFileSync(path.join(beforeTree,file),'utf8'),after=fs.readFileSync(path.join(afterTree,file),'utf8');
    if(file!=='src/compiler/checker.ts'){assert.equal(after===before,true,'other source: '+file);continue;}
    const a=ts.createSourceFile(file,before,ts.ScriptTarget.Latest,true),b=ts.createSourceFile(file,after,ts.ScriptTarget.Latest,true);
    assert.equal(a.parseDiagnostics.length+b.parseDiagnostics.length,0);
    const inventory=source=>{const owners=[],assertions=[];function visit(n){if(ts.isFunctionDeclaration(n)&&manifest.owners.some(o=>o.name===n.name?.text))owners.push(n);if(ts.isNonNullExpression(n))assertions.push(n);ts.forEachChild(n,visit);}visit(source);return{owners,assertions};};
    const old=inventory(a),actual=inventory(b);
    assert.equal(actual.assertions.length-old.assertions.length,8,'all eight required-read edits must be present');
    assert.equal(old.owners.length,6);assert.equal(actual.owners.length,6);
    let reconstructed=after;
    for(const owner of [...actual.owners].sort((x,y)=>y.pos-x.pos)){
        const pin=manifest.owners.find(o=>o.name===owner.name.text),original=old.owners.find(o=>o.name.text===pin.name);
        assert.equal(hash(original.getText(a)),pin.beforeSha256,'control owner moved');
        assert.equal(hash(owner.getText(b)),pin.afterSha256,'required owner checks missing or changed: '+pin.name);
        reconstructed=reconstructed.slice(0,owner.getStart(b))+original.getText(a)+reconstructed.slice(owner.end);
    }
    assert.equal(reconstructed===before,true,'changes outside the six inventoried private owners');
    const crlf={before:(before.match(/\r\n/g)||[]).length,after:(after.match(/\r\n/g)||[]).length};assert.equal(crlf.after,crlf.before);
    const javascript=[];
    for(const removeComments of [false,true]){
        const options={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
        const expected=ts.transpileModule(before,options).outputText,actual=ts.transpileModule(after,options).outputText;
        assert.equal(actual===expected,true,'stock emitted JavaScript bytes');javascript.push({removeComments,bytes:Buffer.byteLength(actual),sha256:hash(actual)});
    }
    changed+=Number(before!==after);evidence={crlf,javascript};
}
assert.equal(changed,1);console.log(JSON.stringify({sourceFiles:files(beforeTree).length,changedFiles:changed,assertions:8,owners:6,...evidence},null,2));
