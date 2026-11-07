'use strict';
const e=require('./engine.cjs'),{fs,path,assert}=e;
const {adapt}=require('./adapt.cjs');
const inventory=require('./inventory.json'),ownership=require('./ownership.json');
const base=path.resolve(process.argv[2]),out=path.resolve(process.argv[3]);fs.mkdirSync(out,{recursive:true});
const scratch=fs.mkdtempSync(path.join(require('node:os').tmpdir(),'bind2-mutant-'));
const rows=[];
function reset(){fs.rmSync(path.join(scratch,'src'),{recursive:true,force:true});fs.cpSync(path.join(base,'src'),path.join(scratch,'src'),{recursive:true});}
function snapshot(){return e.sources(scratch).map(f=>[f,fs.readFileSync(f),fs.statSync(f,{bigint:true}).mtimeNs]);}
function unchanged(before){for(const [f,b,t] of before){assert.deepEqual(fs.readFileSync(f),b);assert.equal(fs.statSync(f,{bigint:true}).mtimeNs,t);}}
function replace(site){const f=path.join(scratch,'src/compiler',site.file),text=fs.readFileSync(f,'utf8');assert.equal(text.split(site.before).length,2);fs.writeFileSync(f,text.replace(site.before,site.after));}
function refused(name,expected){const before=snapshot();let reason;try{adapt(scratch);}catch(error){reason=error.message;}assert(reason&&reason.includes(expected),reason);unchanged(before);rows.push({name,refused:reason,zeroWrites:true});}
try{
 for(const row of ownership){
  const sites=inventory.left.filter(s=>s.file===row.file&&(s.owner===row.owner||s.owner.startsWith(row.owner+'.')));
  if(sites.length<2)continue;
  reset();replace(sites[0]);
  refused('half owner '+row.file+':'+row.owner,row.file+':'+(row.owner.startsWith('createProgram.')?'createProgram':row.owner));
 }
 reset();const applied=adapt(scratch);assert.equal(applied.rewrites,28);
 const file='builder.ts',f=path.join(scratch,'src/compiler',file);let text=fs.readFileSync(f,'utf8');
 // Reinsert the original reviewed import line into the same import declaration.
 const original=fs.readFileSync(path.join(base,'src/compiler',file),'utf8');const pin=inventory.imports.find(s=>s.file===file);assert(pin);
 const at=original.indexOf(pin.before),following=original.slice(at+pin.before.length).split('\n')[0];assert(following);
 assert.equal(text.split(following).length,2);fs.writeFileSync(f,text.replace(following,pin.before+following));
 refused('unused retained import','builder.ts: half-adapted import: 0 core maybeBind calls');
 reset();for(const site of inventory.left)replace(site);
 // 99x must not inspect replacement bodies, including other core helper call shapes.
 const opaqueFile=path.join(scratch,'src/compiler/checker.ts');
 let opaqueText=fs.readFileSync(opaqueFile,'utf8');
 const opaqueSite=inventory.left.find(s=>s.file==='checker.ts');
 opaqueText=opaqueText.replace(opaqueSite.after,'maybeBind(host, undefined)');
 fs.writeFileSync(opaqueFile,opaqueText);
 const checker=path.join(scratch,'src/compiler/checker.ts');fs.appendFileSync(checker,'\n// foreign owner contract\n');
 const checkerBefore=fs.readFileSync(checker),result=adapt(scratch);assert.equal(result.rewrites,28);assert.equal(result.left,0);
 assert.deepEqual(fs.readFileSync(checker),checkerBefore);
 for(const file of ['watch.ts','program.ts'])assert(!/^    maybeBind,\r?$/m.test(fs.readFileSync(path.join(scratch,'src/compiler',file),'utf8')));
 for(const file of ['watch.ts','program.ts']){
  const name=path.join(scratch,'src/compiler',file),saved=fs.readFileSync(name,'utf8'),original=fs.readFileSync(path.join(base,'src/compiler',file),'utf8');
  const eol=original.includes('\r\n')?'\r\n':'\n',line='    maybeBind,'+eol,at=original.indexOf(line);assert(at>=0);
  const following=original.slice(at+line.length).split(eol)[0];assert.equal(saved.split(following).length,2);
  fs.writeFileSync(name,saved.replace(following,line+following));
  refused('unused retained dynamic import '+file,file+': half-adapted import: 0 core maybeBind calls');
  fs.writeFileSync(name,saved);
 }
 const before=snapshot();assert.equal(adapt(scratch).files,0);unchanged(before);
 rows.push({name:'all handed-off groups absent; dynamic program/watch import removal; checker untouched; byte/mtime replay',pass:true,result});
 reset();for(const site of inventory.left)replace(site);
 const parenthesized=inventory.left.find(s=>s.file==='program.ts'&&s.owner==='createProgram');
 const parentFile=path.join(scratch,'src/compiler/program.ts');
 fs.writeFileSync(parentFile,fs.readFileSync(parentFile,'utf8').replace(parenthesized.after,'(maybeBind)(host, undefined)'));
 assert.equal(adapt(scratch).rewrites,28);
 assert(/^    maybeBind,\r?$/m.test(fs.readFileSync(parentFile,'utf8')));
 const parentState=snapshot();assert.equal(adapt(scratch).files,0);unchanged(parentState);
 rows.push({name:'opaque parenthesized core call retains program import and replays without writes',pass:true});
}finally{fs.rmSync(scratch,{recursive:true,force:true});}
fs.writeFileSync(path.join(out,'composition-mutant.json'),JSON.stringify(rows,null,2)+'\n');console.log(JSON.stringify(rows,null,2));
