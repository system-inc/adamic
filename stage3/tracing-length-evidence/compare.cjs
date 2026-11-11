'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process');
const [beforeTree,afterTree,out]=process.argv.slice(2),beforeDir=path.join(beforeTree,'built/local'),afterDir=path.join(afterTree,'built/local');
const names=fs.readdirSync(beforeDir).filter(x=>x.endsWith('.js')).sort();assert.deepEqual(fs.readdirSync(afterDir).filter(x=>x.endsWith('.js')).sort(),names);
const report=[];
for(const name of names){const before=fs.readFileSync(path.join(beforeDir,name),'utf8'),after=fs.readFileSync(path.join(afterDir,name),'utf8');
 if(before===after){report.push({name,classification:'byte-identical'});continue;}
 let pops=0,alls=0;const expected=before.replace(/^([ \t]*)(eventStack\d*)\.length--;/gm,(_,indent,s)=>{pops++;return indent+'if ('+s+'.length === 0) throw new RangeError("Invalid array length");\n'+indent+s+'.splice('+s+'.length - 1);';}).replace(/(eventStack\d*)\.length = 0;/g,(_,s)=>{alls++;return s+'.splice(0);';});
 assert.equal(pops,1,name);assert.equal(alls,1,name);assert.equal(after,expected,'unreviewed emitted bytes: '+name);
 const diff=cp.spawnSync('diff',['-u',path.join(beforeDir,name),path.join(afterDir,name)],{encoding:'utf8'});assert.equal(diff.status,1);fs.writeFileSync(path.join(path.dirname(out),name+'.diff'),diff.stdout);
 report.push({name,popEdits:pops,popAllEdits:alls,hunks:diff.stdout.split('\n').filter(x=>x.startsWith('@@')),classification:'exact two owner edits; no esbuild rename'});
}
const b=require(path.resolve(beforeDir,'typescript.js')),a=require(path.resolve(afterDir,'typescript.js'));const keys=x=>Reflect.ownKeys(x).map(k=>typeof k==='symbol'?['symbol',String(k)]:['string',k]).sort((a,b)=>a[1].localeCompare(b[1]));assert.deepEqual(keys(a),keys(b));
fs.writeFileSync(out,JSON.stringify({files:report,exportKeys:keys(b),exportKeyCount:keys(b).length,exportKeysIdentical:true},null,2)+'\n');console.log(JSON.stringify(report));
