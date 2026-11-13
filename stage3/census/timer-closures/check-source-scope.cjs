const fs=require('fs'),path=require('path'),cp=require('child_process'),crypto=require('crypto'),assert=require('assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const repo='/workspace/adamic',main=process.argv[2]||'/workspace/scratch/timer-main',after=process.argv[3]||'/workspace/scratch/timer-after';
const normalize=t=>t.replace(/\r\n/g,'\n');
const originals=new Map(),expected=new Map();
for(const file of ts.sys.readDirectory(main+'/src',['.ts','.json'],undefined,['**/*'])) {const name=path.relative(main,file);originals.set(name,normalize(fs.readFileSync(file,'utf8')));expected.set(name,originals.get(name));}
function replace(file,before,after){const text=expected.get(file);assert.equal(text.split(before).length,2,file);expected.set(file,text.replace(before,after));}
const owners=require(repo+'/stage3/adapt/99-correctness-timer-closures/owners.json');
for(const row of owners)replace(row.file,row.before,row.after);
const old=JSON.parse(cp.execFileSync('git',['-C',repo,'show','e3f59bf5509b1feef85b405807a732e11dc383b6:stage3/adapt/99-correctness-axolotl-x/owners.json'],{encoding:'utf8',maxBuffer:8*1024*1024})).find(r=>r.owner==='createWatchProgram');
const amended=require(repo+'/stage3/adapt/99-correctness-axolotl-x/owners.json').find(r=>r.owner==='createWatchProgram');
replace('src/compiler/watchPublic.ts',old.after,amended.after);
const file='src/compiler/sys.ts';let text=expected.get(file);const source=ts.createSourceFile(file,text,99,true);let node;
function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getNodeSystem')node=n;ts.forEachChild(n,visit);}visit(source);assert(node);
const before=node.getText(source);let pairs=0;
const forwarded=before.replace(/^([ \t]*)setTimeout,\n[ \t]*clearTimeout,/gm,(_,indent)=>{pairs++;return `${indent}setTimeout: (callback: (label?: unknown) => void, ms: number, label?: unknown): unknown => setTimeout(callback, ms, label),\n${indent}clearTimeout: (timeoutId: unknown): void => clearTimeout(timeoutId),`;});
assert.equal(pairs,2);replace(file,before,forwarded);
replace(file,'declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;','declare function setTimeout(handler: (label?: unknown) => void, timeout: number, label?: unknown): unknown;');
replace(file,'declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;','declare function clearTimeout(handle: unknown): void;');
const changed=[];for(const [name,text]of expected){assert.equal(normalize(fs.readFileSync(after+'/'+name,'utf8')),text,name+': unowned edit');if(text!==originals.get(name))changed.push(name);}
const calls=require(repo+'/stage3/adapt/99-correctness-timer-closures/calls.json');for(const call of calls)assert(expected.get('src/'+call.file).includes(call.after),call.file+': call inventory drift');
console.log(JSON.stringify({sourceFiles:expected.size,changed:changed.sort(),newWholeOwners:owners.length,amendedWholeOwners:['createWatchProgram','getNodeSystem'],retainedAmbientDeclarations:2,forwardingPairs:pairs,allTenCallsMatchInventory:true,noUnownedSourceEdits:true},null,2));
