'use strict';
const fs=require('fs'),path=require('path'),ts=require('typescript'),crypto=require('crypto'),assert=require('assert/strict');
const repo='/workspace/scratch/single-watcher-forwarder',base='/workspace/scratch/single-watcher-before',out=path.join(repo,'stage3/adapt/99-correctness-single-watcher-forwarder');
assert(fs.existsSync(base+'/patch-set.md'),'baseline apply must complete');
const s=fs.readFileSync(base+'/src/compiler/sys.ts','utf8').replace(/\r\n/g,'\n');
const old=`                (
                    // Cant infer types correctly so lets satisfy checker
                    (param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))
                ) as T,`;
assert.equal(s.split(old).length,2);
let after=s.replace('    createWatcher: (callback: T) => FileWatcher,','    createWatcher: (callback: T) => FileWatcher,\n    createForwarder: (getCallbacks: () => T[] | undefined) => T,').replace(old,'                createForwarder(() => cache.get(path)?.callbacks.slice()),');
for(const [caller,typ,p1,p2] of [['cb => pollingWatchFileWorker(fileName, cb, pollingInterval, options),','FileWatcherCallback','string','FileWatcherEventKind'],['cb => fsWatchHandlingExistenceOnHost(fileOrDirectory, entryKind, cb, recursive, fallbackPollingInterval, fallbackOptions),','FsWatchCallback','"rename" | "change"','string | undefined | null']]){
 assert.equal(after.split(caller).length,2);
 after=after.replace(caller,caller+`\n            (getCallbacks: () => ${typ}[] | undefined): ${typ} => (param1: ${p1}, param2: ${p2}, param3?: Date) => { const callbacks = getCallbacks(); if (callbacks) callbacks.forEach(cb => cb(param1, param2, param3)); },`);
}
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
function owner(text,name){const sf=ts.createSourceFile('sys.ts',text,ts.ScriptTarget.Latest,true);assert.equal(sf.parseDiagnostics.length,0);const ns=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(ns.length,1);return ns[0].getText(sf);}
fs.mkdirSync(out,{recursive:true});
const pins=['createSingleWatcherPerName','createSystemWatchFunctions'].map(name=>{const before=owner(s,name),a=owner(after,name);return {file:'sys.ts',owner:name,beforeDeclarations:1,afterDeclarations:1,ownerBeforeSha256:hash(before),ownerAfterSha256:hash(a),before,after:a};});
fs.writeFileSync(out+'/owners.json',JSON.stringify(pins,null,2)+'\n');
let adapter=fs.readFileSync(repo+'/stage3/adapt/99-correctness-axolotl-k/adapt.cjs','utf8');
adapter=adapter.replace("        if (file === 'core.ts') for (const removeComments of [false, true]) assert.equal(transpile(after, removeComments), transpile(text, removeComments), 'core.ts JavaScript changed');\n",'');
fs.writeFileSync(out+'/adapt.cjs',adapter);
const rows=JSON.parse(fs.readFileSync(repo+'/stage3/adapt/40-explicit-any/evidence/sites.json','utf8'));
function collect(x){if(Array.isArray(x))return x.flatMap(collect);if(x&&typeof x==='object'){if(JSON.stringify(x).length<5000&&x.line===516)return [x];return Object.values(x).flatMap(collect);}return [];}
const matches=collect(rows).filter(x=>JSON.stringify(x).includes('sys.ts'));
fs.writeFileSync(out+'/resolved-sites.json',JSON.stringify({inventory:'stage3/adapt/40-explicit-any/evidence/sites.json',resolves:matches,source:'(param1: any, param2: never, param3: any) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3))',resolution:'Both any parameters disappear with the unchecked generic forwarder. Each caller supplies a concrete, three-parameter callback; Road C, #1vkqj9w.'},null,2)+'\n');
assert.equal(matches.length,2,'two 40 rows');
fs.writeFileSync('/workspace/scratch/single-watcher-evidence/before-sys.ts',s);
fs.writeFileSync('/workspace/scratch/single-watcher-evidence/after-sys.ts',after);
console.log(pins.map(p=>({owner:p.owner,before:p.ownerBeforeSha256,after:p.ownerAfterSha256})));console.log(matches);
