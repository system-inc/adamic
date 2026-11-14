'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),cp=require('node:child_process'),assert=require('node:assert/strict'),ts=require('typescript');
const adapter=require('../adapt.cjs'),spec=require('./watch-directory.json'),pins=require('./targets.json');
const tree=path.resolve(process.argv[2]),before=cp.execFileSync('git',['-C',tree,'show','refs/stage3/99-wasm-crypto-import-before:src/compiler/sys.ts'],{encoding:'utf8'});
const expected=fs.readFileSync(path.join(tree,'src/compiler/sys.ts'),'utf8');adapter.validateWatchDirectory(expected);
const root=fs.mkdtempSync(path.join(os.tmpdir(),'watch-directory-contract-')),file=path.join(root,'src/compiler/sys.ts');
const cases=[];let writes=0;const write=fs.writeFileSync;
function noWrites(fn){writes=0;fs.writeFileSync=(...a)=>{writes++;return write(...a);};try{fn();assert.equal(writes,0);}finally{fs.writeFileSync=write;}}
function snapshot(){return [fs.readFileSync(file,'utf8'),String(fs.statSync(file,{bigint:true}).mtimeNs)];}
function refused(label,text){fs.writeFileSync(file,text);const snap=snapshot();noWrites(()=>assert.throws(()=>adapter.adapt(root),undefined,label));assert.deepEqual(snapshot(),snap);cases.push(label);}
try {
 fs.cpSync(path.join(tree,'src'),path.join(root,'src'),{recursive:true});
 for(const eol of ['\n','\r\n']){
  const original=before.replace(/\r\n/g,'\n').replace(/\n/g,eol);fs.writeFileSync(file,original);assert.equal(adapter.adapt(root).files,1);
  assert.equal(fs.readFileSync(file,'utf8'),expected.replace(/\r\n/g,'\n').replace(/\n/g,eol));const snap=snapshot();noWrites(()=>assert.equal(adapter.adapt(root).files,0));assert.deepEqual(snapshot(),snap);
 }
 refused('restoring-shorthand',expected.replace(spec.after,spec.before));
 refused('unnormalized',expected.replace(spec.after,spec.after.replace('!!recursive','recursive')));
 refused('half-adapted-forwarder',before.replace('            watchDirectory,','            '+spec.after+','));
 refused('owner-drift',before.replace('function getNodeSystem()', 'function getNodeSystem(/* drift */)'));
 refused('duplicate-owner',before+'\r\n'+pins[0].ownerBeforeText);
 refused('mixed-endings',before.replace('\r\n','\n'));
 refused('duplicate-forwarder',expected.replace(spec.after,spec.after+', '+spec.after));
 const main=expected.replace(spec.after,spec.before);const planned=adapter.forwardWatchDirectory(main);assert.equal(planned,expected);
 const sf=ts.createSourceFile('sys.ts',expected,ts.ScriptTarget.Latest,true);
 const factory=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createSystemWatchFunctions');assert(factory.getText(sf).includes('recursive: boolean, options: WatchOptions | undefined'));
 console.log(JSON.stringify({exactSystemParameters:true,noRest:true,innerBooleanUnchanged:true,LF:true,CRLF:true,zeroWriteReplay:true,bytesAndMtimePreserved:true,refusals:cases},null,2));
} finally {fs.rmSync(root,{recursive:true,force:true});}
