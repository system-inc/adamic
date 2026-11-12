const fs=require('fs'),path=require('path'),os=require('os'),assert=require('assert/strict');
const adapter=require(process.argv[2]),tree=process.argv[3];
const raw=fs.readFileSync(path.join(tree,'src/compiler/utilities.ts'),'utf8');
const root=fs.mkdtempSync(path.join(os.tmpdir(),'supported-extensions-contract-'));fs.mkdirSync(root+'/src/compiler',{recursive:true});
const file=root+'/src/compiler/utilities.ts';
try{
 fs.writeFileSync(file,raw);assert.equal(adapter.adapt(root).edits,1);
 const after=fs.readFileSync(file,'utf8'),stat=fs.statSync(file,{bigint:true});
 assert.equal(adapter.adapt(root).edits,0);assert.equal(fs.readFileSync(file,'utf8'),after);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);
 const before=adapter.owner(raw).text;
 const mutants={
 'changed-before':raw.replace('const flatBuiltins = flatten(builtins);','const flatBuiltins = flatten(builtins); /* drift */'),
 'changed-after':after.replace('const flatBuiltins = flatten(builtins);','const flatBuiltins = flatten(builtins); /* drift */'),
 'missing-owner':raw.replaceAll('getSupportedExtensions(', 'renamedSupportedExtensions('),
 'duplicate-owner':raw+'\n'+before,
 'nested-owner':raw.replace(before,'function wrapper() { '+before+' }'),
 'changed-overload':raw.replace('options?: CompilerOptions): readonly Extension[][];', 'options: CompilerOptions): readonly Extension[][];')
 };
 for(const [name,text] of Object.entries(mutants)){fs.writeFileSync(file,text);const s=fs.statSync(file,{bigint:true});assert.throws(()=>adapter.adapt(root),undefined,name);assert.equal(fs.readFileSync(file,'utf8'),text);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,s.mtimeNs);}
 for(const eol of ['\n','\r\n']){fs.writeFileSync(file,raw.replace(/\r\n/g,'\n').replace(/\n/g,eol));assert.equal(adapter.adapt(root).edits,1);assert.equal(adapter.adapt(root).edits,0);}
 console.log(JSON.stringify({edits:1,replayEdits:0,replayBytesAndMtimeUnchanged:true,driftRefusedWithoutWrites:Object.keys(mutants),lineEndings:['LF','CRLF'],transpileEqualityWithAndWithoutComments:true}));
}finally{fs.rmSync(root,{recursive:true,force:true});}
