'use strict';
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict'),crypto=require('node:crypto'),cp=require('node:child_process'),ts=require('typescript');
const [baseLane,changedLane,out]=process.argv.slice(2).map(p=>path.resolve(p)),base=path.join(baseLane,'adapted-tree'),changed=path.join(changedLane,'adapted-tree');
fs.mkdirSync(out,{recursive:true});
const sites=require('./sites.json'),files=[...new Set(sites.map(s=>s.file))];
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
const walk=d=>fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]).sort();
function artifacts(tree){return Object.fromEntries(walk(path.join(tree,'built')).filter(f=>/\.(js|mjs|cjs)$/.test(f)||/\/typescript\.d\.(ts|mts|cts)$/.test(f)).map(f=>[path.relative(tree,f),hash(fs.readFileSync(f))]));}
const old=artifacts(base),current=artifacts(changed);assert(Object.keys(old).length);assert.deepEqual(current,old,'all real JS and public API bytes');
const a=JSON.parse(fs.readFileSync(path.join(baseLane,'report.json'))),b=JSON.parse(fs.readFileSync(path.join(changedLane,'report.json')));assert.equal(a.status,'pass');assert.equal(b.status,'pass');for(const k of ['status','counts','failed_tests','baseline_diffs','api','platform','verdict'])assert.deepEqual(b[k],a[k],k);
assert.deepEqual(fs.readFileSync(path.join(baseLane,'oracle/baseline.diff')),fs.readFileSync(path.join(changedLane,'oracle/baseline.diff')));
const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'adapt77-mutants-'));
function reset(){for(const f of files){const dest=path.join(scratch,f);fs.mkdirSync(path.dirname(dest),{recursive:true});fs.copyFileSync(path.join(base,f),dest);}}
function apply(){return cp.spawnSync(process.execPath,[path.join(__dirname,'adapt.cjs'),scratch],{encoding:'utf8'});}
reset();let r=apply();assert.equal(r.status,0,r.stderr);
const expected=Object.fromEntries(files.map(f=>[f,fs.readFileSync(path.join(scratch,f))]));
const oldSources=walk(path.join(base,'src')).filter(f=>f.endsWith('.ts'));assert.deepEqual(walk(path.join(changed,'src')).filter(f=>f.endsWith('.ts')).map(f=>path.relative(changed,f)),oldSources.map(f=>path.relative(base,f)));
for(const oldFile of oldSources){const f=path.relative(base,oldFile);assert.deepEqual(fs.readFileSync(path.join(changed,f)),expected[f]||fs.readFileSync(oldFile),'only reviewed changes: '+f);}
r=apply();assert.equal(r.status,0);assert(JSON.parse(r.stdout).removed===0);for(const f of files)assert.deepEqual(fs.readFileSync(path.join(scratch,f)),expected[f]);
reset();for(const f of files)fs.writeFileSync(path.join(scratch,f),fs.readFileSync(path.join(scratch,f),'utf8').replaceAll('\r\n','\n'));assert.equal(apply().status,0);for(const f of files)assert.equal(fs.readFileSync(path.join(scratch,f),'utf8'),expected[f].toString().replaceAll('\r\n','\n'));
const mutants=[];
for(let i=0;i<sites.length;i++){
 reset();const s=sites[i],f=path.join(scratch,s.file),text=fs.readFileSync(f,'utf8');const ordinal=sites.slice(0,i).filter(t=>t.file===s.file&&t.before===s.before).length;
 let start=-1;for(let n=0;n<=ordinal;n++)start=text.indexOf(s.before,start+1);assert(start>=0);
 const replacement=s.before.replace(/ as [^]*$/, ' as string');assert.notEqual(replacement,s.before);
 const mutated=text.slice(0,start)+replacement+text.slice(start+s.before.length);fs.writeFileSync(f,mutated);
 r=apply();assert.notEqual(r.status,0,'site mutant: '+s.id);assert.match(r.stderr,/reviewed site drift/);assert.equal(fs.readFileSync(f,'utf8'),mutated,'reject before write');
 fs.writeFileSync(path.join(out,s.id+'-mutant.log'),r.stderr);mutants.push({site:s.id,location:s.scoutLocation,mutation:'outer bridge/cast chain replaced by string assertion',caughtBy:'reviewed AST site multiplicity guard',exit:r.status});
}
const artifactMutants=[];
for(const suffix of ['.js','/typescript.d.ts']){const f=Object.keys(old).find(f=>f.endsWith(suffix));assert(f);const dest=path.join(scratch,path.basename(f)+'.byte-mutant');fs.writeFileSync(dest,Buffer.concat([fs.readFileSync(path.join(changed,f)),Buffer.from('\n')]));assert.throws(()=>assert.equal(hash(fs.readFileSync(dest)),old[f]),assert.AssertionError);artifactMutants.push({file:f,mutation:'append one real artifact byte',caughtBy:'SHA256 equality'});}
assert.throws(()=>assert.deepEqual({...b.counts,passing:b.counts.passing-1},a.counts),assert.AssertionError);
function diagnostics(tree,mutate){const cfg=path.join(tree,'src/compiler/tsconfig.json'),config=ts.parseJsonConfigFileContent(ts.readConfigFile(cfg,ts.sys.readFile).config,ts.sys,path.dirname(cfg),undefined,cfg);assert.equal(config.errors.length,0);const host=ts.createCompilerHost(config.options),read=host.readFile;host.readFile=f=>{let text=read(f);if(mutate&&f===path.join(tree,'src/compiler/transformer.ts'))text=text.replace('(node as T & SourceFile).path','(node as string).path');return text;};return ts.getPreEmitDiagnostics(ts.createProgram(config.fileNames,config.options,host)).map(d=>({file:d.file?path.relative(tree,d.file.fileName):null,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));}
const originalDiagnostics=diagnostics(base),finalDiagnostics=diagnostics(changed),checkerMutant=diagnostics(changed,true);assert.deepEqual(originalDiagnostics,[]);assert.deepEqual(finalDiagnostics,[]);assert(checkerMutant.some(d=>d.file==='src/compiler/transformer.ts'&&[2339,2352].includes(d.code)),'actual source type mutant');
function debugRuntime(compiler) {
 const d=compiler.Debug,k=compiler.TypeMapKind,oldFlag=d.isDebugging;d.isDebugging=true;
 try {
  const left=d.attachDebugPrototypeIfDebug({kind:k.Function,func:t=>t,debugInfo:()=>"left\ncontinued"});
  const right=d.attachDebugPrototypeIfDebug({kind:k.Function,func:t=>t,debugInfo:()=>"right"});
  assert(left instanceof d.DebugTypeMapper);assert(right instanceof d.DebugTypeMapper);
  const values=[k.Composite,k.Merged].map(kind=>d.attachDebugPrototypeIfDebug({kind,mapper1:left,mapper2:right}).__debugToString());
  for(const value of values)assert.equal(value,"m1: left\n    continued\nm2: right");
  // The class view is not an unconditional promise: a child from before debug mode can lack the prototype.
  const bad=d.attachDebugPrototypeIfDebug({kind:k.Composite,mapper1:{kind:k.Function,func:t=>t},mapper2:right});
  assert.throws(()=>bad.__debugToString(),TypeError);
  return values;
 } finally {d.isDebugging=oldFlag;}
}
const oldDebug=debugRuntime(require(path.join(base,'built/local/typescript.js'))),newDebug=debugRuntime(require(path.join(changed,'built/local/typescript.js')));assert.deepEqual(newDebug,oldDebug);
const library=fs.readFileSync(path.join(changed,'built/local/typescript.js'),'utf8');assert(library.includes('m1: ${'));
const debugMutantFile=path.join(scratch,'debug-mutant.cjs');fs.writeFileSync(debugMutantFile,library.replace('m1: ${','m0: ${'));
assert.throws(()=>debugRuntime(require(debugMutantFile)),assert.AssertionError);
const report={base:a.execution.commit,node:process.version,typescript:ts.version,reviewed:18,removed:4,unresolved:14,sites,sourceFiles:oldSources.length,idempotent:true,LFReconstruction:true,outputs:current,jsFiles:Object.keys(current).filter(f=>/\.(js|mjs|cjs)$/.test(f)).length,apiFiles:Object.keys(current).filter(f=>f.endsWith('/typescript.d.ts')),lane:{counts:b.counts,failed_tests:b.failed_tests,baseline_diffs:b.baseline_diffs,status:b.status},checker:{before:originalDiagnostics,after:finalDiagnostics,mutant:checkerMutant},mutants,artifactMutants,laneCountMutant:'caught',debugRuntime:oldDebug,runtimeMutant:{mutation:"actual built nested-mapper m1 label changed to m0",caughtBy:"exact successful debugger output"},totalMutants:23};
fs.writeFileSync(path.join(out,'proof.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify({reviewed:18,removed:4,unresolved:14,mutants:23,checkerDiagnostics:0,jsFiles:report.jsFiles,apiFiles:report.apiFiles,lane:report.lane}));
