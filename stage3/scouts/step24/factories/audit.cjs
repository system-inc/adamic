// Recounts data and tests boundary decisions against small independent source witnesses.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),assert=require('node:assert/strict'),cp=require('node:child_process'),os=require('node:os');
const base=__dirname,root=path.resolve(process.argv[2]||'/workspace/cache/tsc-census/prepared'),data=path.join(base,'data');
function read(name){return JSON.parse(fs.readFileSync(path.join(data,name+'.json'),'utf8'));}
const f=read('factories'),c=read('counts'),manifest=read('source_manifest');
function audit(rows,counts,pins){for(const pin of pins)assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(root,pin.file))).digest('hex'),pin.sha256,'source pin '+pin.file);const builders=rows.filter(x=>!x.parser&&x.nodeResult&&x.factoryName);assert.equal(builders.length,counts.factoryNodeBuilders,'builder population');assert.equal(builders.filter(x=>x.completeBeforeEscape).length,counts.factoryCompleteBeforeEscape,'complete count');assert.equal(builders.filter(x=>x.completeApartFromParent).length,counts.factoryCompleteApartFromParent,'syntax count');for(const row of rows){for(const field of row.requiredFields){if(field.status==='definitely-written-before-escape'){const ids=new Set(row.returns.map(x=>x.object));const exits=row.escapes.filter(x=>x.first&&ids.has(x.object));assert.ok(exits.length&&exits.every(e=>e.written.includes(field.name)&&(!e.undefined.includes(field.name)||field.allowsUndefined)),'field proof '+row.id+':'+field.name);}}}assert.equal(counts.upstreamDiagnostics.length,0,'upstream checker');assert.equal(read('deferred_fields').length,counts.deferredRequiredFactoryFieldPairs,'deferred count');assert.equal(read('after_escape').length,counts.afterEscapeWriteSites,'after-escape count');assert.equal(read('parser_paths').length,counts.parserBaseAndFinishCalls,'parser path count');}
audit(f,c,manifest);
const artifactMutants=[];for(const[name,mutate]of[
 ['inflate complete count',(rows,counts)=>counts.factoryCompleteBeforeEscape++],
 ['drop required field evidence',(rows)=>{const row=rows.find(r=>r.name==='createBinaryExpression'&&!r.parser);const field=row.requiredFields.find(p=>p.name==='right');for(const e of row.escapes)e.written=e.written.filter(p=>p!=='right');}],
 ['change pinned source hash',(rows,counts,pins)=>pins[0].sha256='0'.repeat(64)],
]){let caught;try{const rows=structuredClone(f),counts=structuredClone(c),pins=structuredClone(manifest);mutate(rows,counts,pins);audit(rows,counts,pins);}catch(error){caught=error.message;}assert.ok(caught,name+' survived');artifactMutants.push({name,caught});console.log('CAUGHT '+name+': '+caught.split('\n')[0]);}
const work=fs.mkdtempSync(path.join(os.tmpdir(),'step24-proof-'));fs.mkdirSync(path.join(work,'src/compiler/factory'),{recursive:true});
// Virtual .ts paths let the stock checker inspect committed .a source without creating authored .ts files.
fs.writeFileSync(path.join(work,'src/compiler/factory/proofs.ts'),'');
fs.writeFileSync(path.join(work,'src/compiler/parser.ts'),'');
fs.writeFileSync(path.join(work,'src/compiler/types.ts'),'');
fs.writeFileSync(path.join(work,'src/compiler/utilities.ts'),'');
fs.writeFileSync(path.join(work,'src/compiler/tsconfig.json'),JSON.stringify({compilerOptions:{strict:true,noEmit:true,target:'es2020'},include:['**/*.ts']}));
const result=cp.spawnSync('node',[path.join(base,'inventory.cjs'),work,path.join(work,'data')],{encoding:'utf8',env:{...process.env,SCOUT_SELFTEST:path.join(base,'proofs.a')},timeout:120000});assert.equal(result.status,0,result.stderr);
const proofs=JSON.parse(fs.readFileSync(path.join(work,'data/factories.json'),'utf8'));
const cases={createComplete:'definitely-written-before-escape',createConditional:'conditional-or-path-incomplete-write',createBothBranches:'definitely-written-before-escape',createPassed:'written-by-return-after-earlier-escape',createAlias:'definitely-written-before-escape',createLoop:'unresolved-control',createCompound:'conditional-or-path-incomplete-write',createUndefined:'written-undefined-before-escape'};
for(const[name,status]of Object.entries(cases)){const row=proofs.find(r=>r.name===name);assert.equal(row.requiredFields.find(p=>p.name==='value').status,status,name);console.log('PASS boundary '+name+': '+status);}
const scannerMutants=[];
const scanner=fs.readFileSync(path.join(base,'inventory.cjs'),'utf8');
for(const [name,old,replacement,witness]of [
 ['branch writes joined by union','u.written=new Set([...u.written].filter(x=>v.written.has(x)))','u.written=new Set([...u.written,...v.written])','createConditional'],
 ['passed node never escapes',"for(const a of args)escape(a,n,'passed',s);",'// mutant: ignore passed objects','createPassed'],
 ['undefined literal treated as initialized',"if(unwrap(e)?.getText()==='undefined')undef.push(fields.at(-1));",'/* mutant: ignore undefined literal */','createUndefined'],
]){assert.ok(scanner.includes(old),name+' edit not found');const file=path.join(work,name.replaceAll(' ','-')+'.cjs'),output=path.join(work,'mutant-'+scannerMutants.length);fs.writeFileSync(file,scanner.replace(old,replacement));const run=cp.spawnSync('node',[file,work,output],{encoding:'utf8',env:{...process.env,SCOUT_SELFTEST:path.join(base,'proofs.a')},timeout:120000});assert.equal(run.status,0,run.stderr);const rows=JSON.parse(fs.readFileSync(path.join(output,'factories.json'),'utf8'));const observed=rows.find(r=>r.name===witness).requiredFields.find(p=>p.name==='value').status;assert.notEqual(observed,cases[witness],name+' survived');scannerMutants.push({name,witness,expected:cases[witness],observed,caughtBy:'independent boundary status assertion'});console.log('CAUGHT scanner '+name+': '+witness+' changed to '+observed);}
fs.writeFileSync(path.join(data,'audit.json'),JSON.stringify({artifactMutants,scannerMutants,boundaryCases:cases},null,2)+'\n');console.log('PASS source pins, population, field evidence, counts, three artifact mutants, three scanner mutants and eight boundary cases');
