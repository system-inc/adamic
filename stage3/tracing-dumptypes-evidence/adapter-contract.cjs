const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const a=require('/workspace/adamic/stage3/adapt/99-correctness-tracing-dumptypes/adapt.cjs');
const before=fs.readFileSync('/workspace/scratch/dp24r5m-base/src/compiler/tracing.ts','utf8');
const after=fs.readFileSync('/workspace/scratch/dp24r5m-flags-road/src/compiler/tracing.ts','utf8');
const root=fs.mkdtempSync('/workspace/scratch/dp24r5m-contract-'),rows=[];
function fixture(text){const dir=fs.mkdtempSync(path.join(root,'case-'));fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});const file=path.join(dir,'src/compiler/tracing.ts');fs.writeFileSync(file,text);return {dir,file};}
function accept(name,text,expected){const {dir,file}=fixture(text);assert.equal(a.adapt(dir).edits,2);assert.equal(fs.readFileSync(file,'utf8'),expected);const stat=fs.statSync(file,{bigint:true});assert.equal(a.adapt(dir).edits,0);assert.equal(fs.readFileSync(file,'utf8'),expected);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);rows.push({name,accepted:true,replayWrites:0});}
function reject(name,text){const {dir,file}=fixture(text),stat=fs.statSync(file,{bigint:true});let error;try{a.adapt(dir);}catch(e){error=e;}assert(error,name);assert.equal(fs.readFileSync(file,'utf8'),text);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);rows.push({name,accepted:false,writes:0,reason:error.message});}
accept('CRLF',before,after);accept('LF',before.replace(/\r\n/g,'\n'),after.replace(/\r\n/g,'\n'));
accept('foreign outside owner',before+'\n// preserved foreign text\n',after+'\n// preserved foreign text\n');
accept('foreign namespace import',before.replace('    type Node,','    type SourceFile,\r\n    type Node,'),after.replace('    type Node,','    type SourceFile,\r\n    type Node,'));
reject('whole owner drift',before.replace('const numTypes = types.length;','const numTypes = types.length + 0;'));
reject('restored object cast only',after.replace('const objectFlags = getObjectFlags(type);','const objectFlags = (type as any).objectFlags;'));
reject('restored intrinsic cast only',after.replace('intrinsicName: type.flags & TypeFlags.Intrinsic ? (type as IntrinsicType).intrinsicName : undefined,','intrinsicName: (type as any).intrinsicName,'));
reject('after missing value import',after.replace('    getObjectFlags,\r\n',''));
reject('after missing type import',after.replace('    type IntrinsicType,\r\n',''));
reject('duplicate import',after.replace('    getObjectFlags,','    getObjectFlags,\r\n    getObjectFlags,'));
reject('wrong import kind',after.replace('    getObjectFlags,','    type getObjectFlags,'));
reject('missing owner',before.replace('function dumpTypes(','function omittedDumpTypes('));
reject('duplicate owner',before+'\n'+a.owner(before).text+'\n');
reject('namespace drift',before.replace('export namespace tracingEnabled','export namespace otherTracing'));
reject('before owner with after imports',before.replace('    getLineAndCharacterOfPosition,','    getLineAndCharacterOfPosition,\r\n    getObjectFlags,'));
const result={cases:rows.length,accepted:rows.filter(r=>r.accepted).length,noWriteRefusals:rows.filter(r=>!r.accepted).length,rows};
fs.writeFileSync('/workspace/scratch/dp24r5m-evidence/adapter-contract-results.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({cases:result.cases,accepted:result.accepted,noWriteRefusals:result.noWriteRefusals}));
