'use strict';
// Assemble observations and judgments only after completed full streams exist.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [inspectedPath,planPath,comparisonPath,output]=process.argv.slice(2);
const inspected=JSON.parse(fs.readFileSync(inspectedPath)),plan=JSON.parse(fs.readFileSync(planPath)),comparison=JSON.parse(fs.readFileSync(comparisonPath));
assert.deepEqual(plan.introduced,[],'annotation closure has new stock checker errors');
const key=s=>JSON.stringify([s.file,s.line,s.column]);
const planned=new Map(plan.sites.map(s=>[key(s),s]));
const evidence=path.join(__dirname,'evidence');
const proof=name=>JSON.parse(fs.readFileSync(path.join(evidence,name+'.json')));
const hooks=proof('hooks');
const sourceApiProof=proof('source-api-proof');
const verification=proof('verify');
const counts=rows=>Object.fromEntries([...new Set(rows)].sort().map(k=>[k,rows.filter(v=>v===k).length]));
const sites=inspected.sites.map(s=>{
 const p=planned.get(key(s));assert(p,'missing judgment');
 const observation={...s};delete observation.judgment;
 let judgment;
 if((s.file==='src/compiler/factory/nodeFactory.ts'&&s.line===1164)||(s.file==='src/compiler/utilities.ts'&&s.line===8582)){
  const hook=s.file.includes('nodeFactory')?'nodeFactory':'objectAllocator';
  judgment={kind:'real value',reason:'The open internal registration hook accepts a callback that returns a number. Node retains it and forEach stops before the next patcher.',
   implementationInventory:{closed:false,reason:'No registration call exists in compiler/harness/test source; arbitrary external implementations remain possible.',witness:'hooks.cjs'},
   corpus:{hits:0,values:[],processSnapshots:23},witness:hooks.observed.filter(v=>v.hook===hook)};
 }else if(p.eligible){
  judgment={kind:'type-only adaptation',reason:'Retype the closed internal declarations to undefined. Stock TypeScript checks the entire dependency closure and the complete upstream build; emitted JavaScript and public API bytes are unchanged.',rewriteIds:p.editIds,rewrittenType:p.rewrittenType};
 }else{
  judgment={kind:'unclassified',reason:s.expression==='noop'?'noop always returns undefined, but its global retype introduces four harness/test typing errors. No permitted checked retype preserving the full tree has been established.':'No checked closed implementation retype has been established: '+p.reason};
 }
 return {observation,judgment};
});
const typeOnly=sites.filter(s=>s.judgment.kind==='type-only adaptation').length;
const result={date:'2026-10-20',mainSha:inspected.mainSha,branch:'typescript/stage3-void-widening',typescriptVersion:'6.0.3',
 measurement:{scope:comparison.measurementMode,columns:'one-based UTF-16',fullSiteCount:comparison.measurementMode==='full'?sites.length:null,sourceScanSiteCount:sites.length,exhaustiveLanguageCoverage:false,
  limits:'This source-level scan invokes the unchanged production refusal on every AST node, including checker-diagnosed bodies. It does not run other refusal checks or concrete generic instantiations. The requested full latent comparison is blocked.'},
 countsByForm:counts(sites.map(s=>s.observation.form)),countsByFile:counts(sites.map(s=>s.observation.file)),countsByJudgment:counts(sites.map(s=>s.judgment.kind)),
 annotationPlan:{diagnostics:{before:plan.baselineDiagnostics,after:plan.afterDiagnostics,introduced:plan.introduced},edits:plan.edits,erasure:plan.erasure},
 sourceScanComparison:comparison,fullClearanceProven:comparison.measurementMode==='full'&&comparison.voidRemoved===typeOnly&&comparison.new.length===0,sourceScanClearanceProven:comparison.voidRemoved===typeOnly&&comparison.new.length===0,fullLatentBlocker:proof('full-blocker'),
 publicApi:sourceApiProof,localMutants:verification,sites};
const historical=JSON.parse(fs.readFileSync(path.join(evidence,'historical-first-error-sites.json')));
const present=new Set(sites.map(s=>key(s.observation))), previous=new Set(historical.map(key));
result.comparisons={historicalFirstError:{mainSha:'8c318e45',sites:88,commonSites:[...present].filter(k=>previous.has(k)).length,removedSites:historical.filter(s=>!present.has(key(s))),addedSourceScanSites:sites.map(s=>s.observation).filter(s=>!previous.has(key(s)))},lowerNarrowInventory:{previousSites:127,currentDirectInventory:109,exitSitesAlreadyClearedOnMain:18,additionalSourceReturnSites:7,sourceScanSites:sites.length}};
result.sourceScanMutants=proof('source-mutants');
result.provenance=proof('source-scan-provenance');
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({sites:sites.length,judgments:result.countsByJudgment,fullClearanceProven:result.fullClearanceProven,newRefusals:comparison.newRefused.length}));
