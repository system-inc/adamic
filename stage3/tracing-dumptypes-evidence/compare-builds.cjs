'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),cp=require('node:child_process');
const [before,after,out]=process.argv.slice(2),rows=[],patches=[];
const digest=x=>crypto.createHash('sha256').update(x).digest('hex');
function files(root,prefix=''){return fs.readdirSync(path.join(root,prefix),{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(root,path.join(prefix,e.name)):[path.join(prefix,e.name)]);}
const names=files(path.join(before,'built')).filter(f=>/\.(js|mjs|cjs)$/.test(f)).sort();
assert.deepEqual(files(path.join(after,'built')).filter(f=>/\.(js|mjs|cjs)$/.test(f)).sort(),names);
for(const name of names){
 const a=fs.readFileSync(path.join(before,'built',name),'utf8'),b=fs.readFileSync(path.join(after,'built',name),'utf8');
 if(a===b){rows.push({name,equal:true});continue;}
 const marker='function dumpTypes(types) {';const x=a.indexOf(marker),y=b.indexOf(marker);assert(x>=0&&y>=0,name);
 const endA=a.indexOf('function dumpLegend()',x),endB=b.indexOf('function dumpLegend()',y);assert(endA>x&&endB>y,name);
 assert.equal(b.slice(0,y),a.slice(0,x),name+' prefix outside dumpTypes changed');assert.equal(b.slice(endB),a.slice(endA),name+' suffix outside dumpTypes changed');
 const bodyA=a.slice(x,endA),bodyB=b.slice(y,endB);
 let expected=bodyA.replace('const objectFlags = type.objectFlags;','const objectFlags = getObjectFlags(type);');assert.notEqual(expected,bodyA,name+' original objectFlags site missing');
 expected=expected.replace('intrinsicName: type.intrinsicName,','intrinsicName: type.flags & 402431 /* Intrinsic */ ? type.intrinsicName : void 0,');
 assert.equal(bodyB,expected,name+' unexpected dumpTypes JavaScript');
 const p=cp.spawnSync('git',['diff','--no-index','--',path.join(before,'built',name),path.join(after,'built',name)],{encoding:'utf8'});assert.equal(p.status,1);patches.push(p.stdout);
 rows.push({name,equal:false,diffExit:p.status,hunks:(p.stdout.match(/^@@/gm)||[]).length,allHunksInDumpTypes:true,beforeSha256:digest(a),afterSha256:digest(b)});
}
const decl=files(path.join(before,'built')).filter(f=>f.endsWith('.d.ts')).sort();assert.deepEqual(files(path.join(after,'built')).filter(f=>f.endsWith('.d.ts')).sort(),decl);for(const name of decl)assert.equal(digest(fs.readFileSync(path.join(before,'built',name))),digest(fs.readFileSync(path.join(after,'built',name))),name+' declaration changed');
function keys(tree){const ts=require(path.resolve(tree,'built/local/typescript.js'));return {own:Reflect.ownKeys(ts).filter(k=>typeof k==='string').sort(),symbols:Reflect.ownKeys(ts).filter(k=>typeof k==='symbol').map(String).sort(),enumerable:Object.keys(ts).sort()};}
const ka=keys(before),kb=keys(after);assert.deepEqual(kb,ka);
fs.writeFileSync(path.join(out,'build-comparison-results.json'),JSON.stringify({files:rows,declarationFiles:decl.length,exportKeys:ka.own.length,symbolKeys:ka.symbols.length},null,2)+'\n');fs.writeFileSync(path.join(out,'runtime-export-keys.json'),JSON.stringify(ka,null,2)+'\n');fs.writeFileSync(path.join(out,'javascript-hunks.patch'),patches.join('\n'));console.log(JSON.stringify({javascriptFiles:names.length,changed:rows.filter(r=>!r.equal).map(r=>r.name),declarationFiles:decl.length,exportKeys:ka.own.length}));
