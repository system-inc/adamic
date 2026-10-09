const fs=require('node:fs'),path=require('node:path'),zlib=require('node:zlib'),assert=require('node:assert/strict');
const evidence=path.resolve(process.argv[2]),callers=path.resolve(process.argv[3]),out=path.resolve(process.argv[4]||__dirname);
const parserFunctions=new Set(JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json'),'utf8')).parserFunctionSites);
const phaseCorrections=[];
const summary=JSON.parse(fs.readFileSync(path.join(evidence,'summary.json'),'utf8'));
const callerSummary=JSON.parse(fs.readFileSync(path.join(callers,'summary.json'),'utf8'));assert.equal(summary.acceptance.projects,301);assert.equal(callerSummary.projects,301);
const groups=new Map(),callerGroups=new Map(),cases=[];
function add(map,p,workload){const key=p.origin+'|'+p.allocation+'|'+p.frames.join(' > ')+(p.callerStack?'|'+p.callerStack.join(' > '):'');let group=map.get(key);if(!group){group={origin:p.origin,allocation:p.allocation,frames:p.frames,...(p.callerStack?{callerStack:p.callerStack}:{}),workloads:{}};map.set(key,group);}let row=group.workloads[workload];if(!row){row={created:0,noParent:0,notGraphReachable:0,detached:0,kinds:{}};group.workloads[workload]=row;}for(const k of ['created','noParent','notGraphReachable','detached'])row[k]+=p[k];for(const [k,n] of Object.entries(p.kinds))row.kinds[k]=(row.kinds[k]||0)+n;}
for(const id of ['scanner',...summary.cases.map(c=>c.id)]){
 const report=JSON.parse(fs.readFileSync(path.join(evidence,id+'.json'),'utf8'));
 let moved=0;for(const p of report.paths){const actual=p.frames.some(f=>parserFunctions.has(f))?'parser':'factory outside parsing';if(p.origin!==actual){assert.equal(p.origin,'parser');moved+=p.created;p.origin=actual;}}
 report.totals.parser-=moved;report.totals.synthetic+=moved;if(moved)phaseCorrections.push({id,movedFromParserToSynthetic:moved});
 const workload=id==='scanner'?'scanner':'acceptance';for(const p of report.paths)if(p.detached||p.origin==='factory outside parsing')add(groups,p,workload);
 if(id!=='scanner'){assert(report.instrumentationAgrees&&report.goldenAgrees);const synthetic=JSON.parse(fs.readFileSync(path.join(callers,id+'.json'),'utf8'));assert.equal(synthetic.totals.synthetic,report.totals.synthetic,id+' supplemental count mismatch');for(const p of synthetic.paths)add(callerGroups,p,workload);cases.push({id,totals:report.totals,rootSourceFiles:report.rootSourceFiles,syntaxDigest:report.syntaxDigest,stdoutSha256:report.stdoutSha256,expectedExit:report.expectedExit,instrumentationAgrees:true,goldenAgrees:true});}
}
summary.acceptance.totals={};for(const row of cases)for(const [key,value]of Object.entries(row.totals))summary.acceptance.totals[key]=(summary.acceptance.totals[key]||0)+value;
for(const workload of ['scanner','acceptance']){const counted=[...groups.values()].reduce((sum,g)=>sum+(g.workloads[workload]?.detached||0),0);assert.equal(counted,summary[workload].totals.detachedLiteralUnion);}
fs.writeFileSync(path.join(out,'summary.json'),JSON.stringify({...summary,syntheticCallerPassSeconds:callerSummary.wallSeconds,phaseCorrections},null,2)+'\n');
fs.writeFileSync(path.join(out,'projects.json'),JSON.stringify(cases,null,2)+'\n');
fs.copyFileSync(path.join(evidence,'scanner-inputs.json'),path.join(out,'scanner-inputs.json'));
const paths={note:'All full parser/factory paths with a detached candidate or synthetic allocation. Counts include matched attached creations at these paths. Synthetic callers have a separate source-mapped full-stack inventory.',allocationPaths:[...groups.values()],syntheticCallerPaths:[...callerGroups.values()]};
fs.writeFileSync(path.join(out,'paths.json.gz'),zlib.gzipSync(JSON.stringify(paths),{level:9}));
// Readable site totals complement the full dynamic paths, without repeating library paths 301 times.
const sites=new Map();
for(const g of groups.values()){
 const parser=g.frames.filter(f=>f.startsWith('src/compiler/parser.ts:')).at(-1)||g.allocation;
 const factory=g.frames.find(f=>f.startsWith('src/compiler/factory/nodeFactory.ts:'))||g.allocation;
 const site=g.origin==='parser'?parser:factory;const key=g.origin+'|'+site;let row=sites.get(key);if(!row){row={origin:g.origin,site,scanner:0,acceptance:0,unreachable:0,kinds:new Set()};sites.set(key,row);}for(const [w,v]of Object.entries(g.workloads)){row[w]+=v.detached;row.unreachable+=v.notGraphReachable;for(const kind of Object.keys(v.kinds))row.kinds.add(kind);}
}
const syntheticSites=new Map();
for(const g of callerGroups.values()){
 const caller=g.callerStack.find(line=>line.includes('src/compiler/')&&!line.includes('src/compiler/factory/'))||g.callerStack[0];const match=caller.match(/src\/compiler\/[^:)]+:\d+/);const site=match?match[0]:caller;const key=site;let row=syntheticSites.get(key);if(!row){row={site,created:0,unreachable:0,kinds:new Set()};syntheticSites.set(key,row);}for(const v of Object.values(g.workloads)){row.created+=v.created;row.unreachable+=v.notGraphReachable;for(const k of Object.keys(v.kinds))row.kinds.add(k);}
}
const lines=['# Observed creation paths','','Exact constructor expression sites are in sites.json. Function-definition sites below identify the deepest parser creation path. Every full dynamic parser/factory path and every source-mapped synthetic caller stack is in paths.json.gz. “Detached” here is the requested literal union: no parent OR not graph-reachable from a SourceFile. It includes legitimate parentless SourceFile roots; summary.json also reports the root-exempt union. Counts do not estimate bytes or membership.','', '## Parser and factory paths with detached candidates','','| Origin | Creation function site | Scanner detached | Acceptance detached | Unreachable across both | Kinds |','|---|---|---:|---:|---:|---|'];
for(const row of [...sites.values()].sort((a,b)=>(b.scanner+b.acceptance)-(a.scanner+a.acceptance)))lines.push(`| ${row.origin} | ${row.site} | ${row.scanner} | ${row.acceptance} | ${row.unreachable} | ${[...row.kinds].sort().join(', ')} |`);
lines.push('','## Synthetic allocations by calling code site','','These are the first non-factory compiler frames in source-mapped stacks, with line and column capture preserved in the compressed full paths. Modules, CLI launchers and instrumentation frames are omitted from this site summary.','', '| Caller site | Created | Unreachable | Kinds |','|---|---:|---:|---|');
for(const row of [...syntheticSites.values()].sort((a,b)=>b.created-a.created))lines.push(`| ${row.site} | ${row.created} | ${row.unreachable} | ${[...row.kinds].sort().join(', ')} |`);
fs.writeFileSync(path.join(out,'code-paths.md'),lines.join('\n')+'\n');
console.log(`301 goldens and instrumentation controls pass. ${groups.size} full allocation paths, ${callerGroups.size} synthetic caller paths; compressed evidence ${fs.statSync(path.join(out,'paths.json.gz')).size} bytes.`);
