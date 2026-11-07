'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const s='/workspace/scratch/arrayof',hash=x=>crypto.createHash('sha256').update(x).digest('hex'),norm=x=>x.replaceAll('\r\n','\n'),pin=JSON.parse(fs.readFileSync(s+'/instrument.json','utf8'));
const names=fs.readdirSync(s+'/before/src',{recursive:true}).filter(p=>fs.statSync(s+'/before/src/'+p).isFile()).sort(),changed=[];
for(const name of names)if(!fs.readFileSync(s+'/before/src/'+name).equals(fs.readFileSync(s+'/counted/src/'+name)))changed.push(name);
assert.deepEqual(changed,['compiler/core.ts']);const a=fs.readFileSync(s+'/before/src/compiler/core.ts','utf8'),b=fs.readFileSync(s+'/counted/src/compiler/core.ts','utf8');assert.equal(norm(a).replace(pin.before,pin.after),norm(b));
const emit=x=>ts.transpileModule(x,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
function removeOwner(x){const sf=ts.createSourceFile('core.js',x,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),nodes=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='arrayOf');assert.equal(nodes.length,1);return x.slice(0,nodes[0].getStart(sf))+x.slice(nodes[0].end);}
assert.equal(removeOwner(emit(a)),removeOwner(emit(b)));console.log(JSON.stringify({sourceFiles:names.length,changed,onlyNormalizedOwnerDiffers:true,measurementCoreUsesLF:true,outsideOwnerJavascriptIdentical:true,beforeSha256:hash(a),countedSha256:hash(b),instrumentedOwner:pin.after},null,2));
