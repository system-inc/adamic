'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict');
const adapter=process.argv[2],source=process.argv[3],destination=process.argv[4];
const a=require(adapter+'/adapt.cjs');fs.mkdirSync(destination+'/src/compiler',{recursive:true});
const f=destination+'/src/compiler/types.ts',original=fs.readFileSync(source);fs.writeFileSync(f,original);
assert.deepEqual(a.adapt(destination),{files:1,edits:1,javascriptChanging:false});
const after=fs.readFileSync(f),beforeStat=fs.statSync(f,{bigint:true});
assert(after.includes(Buffer.from('\r\n')));assert(!after.toString().replace(/\r\n/g,'').includes('\n'));
assert.deepEqual(a.adapt(destination),{files:0,edits:0,javascriptChanging:false});
assert.equal(fs.statSync(f,{bigint:true}).mtimeNs,beforeStat.mtimeNs);
assert.deepEqual(a.rewrite(destination,true),{files:1,edits:1,javascriptChanging:false});assert.deepEqual(fs.readFileSync(f),original);
for(const [name,text] of [
 ['owner drift',original.toString().replace('interface '+a.owners[0].owner+' ', 'interface '+a.owners[0].owner+' /* drift */ ')],
 ['half adaptation',original.toString().replace(a.owners[0].before.replace(/\n/g,'\r\n'),a.owners[0].after.replace(/\n/g,'\r\n').replace(a.owners[0].owner==='Node'?'allowEdits?: true':'constraint?: Type',a.owners[0].owner==='Node'?'allowEdits?: boolean':'constraint?: string'))],
 ['mixed endings',original.toString().replace('\r\n','\n')],
 ['duplicate owner',original.toString()+'\r\nexport interface '+a.owners[0].owner+' {}\r\n']]){
 fs.writeFileSync(f,text);assert.throws(()=>a.adapt(destination),undefined,name);assert.equal(fs.readFileSync(f,'utf8'),text,name+' writes');
}
fs.writeFileSync(f,original);console.log('PASS owner pins, CRLF, replay no-write, restoration, drift/half/mixed/duplicate rejection');
