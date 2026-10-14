// Retain per-input observations and correlate every checked site, including unseen ones.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [controlArg,instrumentedArg,logsArg]=process.argv.slice(2),control=path.resolve(controlArg),instrumented=path.resolve(instrumentedArg),logs=path.resolve(logsArg);
const driver=path.resolve(__dirname,'../../../drivers/tsc');
const selection=JSON.parse(fs.readFileSync(path.join(driver,'selection.json')));
const inputs=[...selection.cases.map(r=>({id:r.id,source:r.source,path:r.path,sha256:r.source_sha256})),{id:'tiny',source:'stage3/drivers/tsc/tiny',path:'tiny'}].sort((a,b)=>a.id.localeCompare(b.id));
const sites=JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json'))),checked=sites.filter(s=>s.status==='checked'),byID=new Map();
for(const s of checked){const id=s.file+':'+s.line+':'+s.column+'@'+s.start+'-'+s.end;byID.set(id,{site:id,expression:s.expression,visits:0,nullish:0,null:0,undefined:0,inputs:[],nullish_inputs:[]});}
for(const folder of [control,instrumented]){const report=JSON.parse(fs.readFileSync(path.join(folder,'report.json')));assert.equal(report.cases,301);assert.equal(report.passed,301);assert.deepEqual(report.failed,[]);}
assert.equal(fs.readdirSync(logs).filter(n=>n.endsWith('.json')).length,301);
const records=[];
for(const input of inputs){
 if(input.id!=='tiny')assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(driver,input.path))).digest('hex'),input.sha256);
 for(const suffix of ['stdout','stderr','exit'])assert.deepEqual(fs.readFileSync(path.join(control,input.id,'actual.'+suffix)),fs.readFileSync(path.join(instrumented,input.id,'actual.'+suffix)),input.id+':'+suffix);
 const data=JSON.parse(fs.readFileSync(path.join(logs,input.id+'.json')));assert.equal(path.basename(data.input),input.id);require('./probe-contract.cjs')(data,sites);
 for(const [id,count] of Object.entries(data.hits)){assert.ok(byID.has(id),id);assert.ok(Number.isSafeInteger(count)&&count>0);const row=byID.get(id);row.visits+=count;row.inputs.push(input.id);}
 const nullishSites=new Set();
 for(const [key,count] of Object.entries(data.nullish)){
  const pos=key.lastIndexOf(':'),id=key.slice(0,pos),kind=key.slice(pos+1),row=byID.get(id);assert.ok(row);assert.ok(kind==='null'||kind==='undefined');assert.ok(count>0&&count<=data.hits[id]);row.nullish+=count;row[kind]+=count;nullishSites.add(id);
 }
 for(const id of nullishSites)byID.get(id).nullish_inputs.push(input.id);
 const first=Object.keys(data.nullish)[0]||null;
 records.push({...input,exit:Number(fs.readFileSync(path.join(instrumented,input.id,'actual.exit'),'utf8')),first_nullish:first,hits:data.hits,nullish:data.nullish});
}
const rows=[...byID.values()].sort((a,b)=>a.site.localeCompare(b.site));
const sets=[],setMap=new Map();
function inputSet(ids){const key=ids.join('\n');if(!setMap.has(key)){const name=ids.length===301?'ALL301':'INPUTS'+(sets.length+1);setMap.set(key,name);sets.push({id:name,inputs:ids});}return setMap.get(key);}
for(const row of rows){row.reached_input_set=inputSet(row.inputs);row.nullish_input_set=inputSet(row.nullish_inputs);delete row.inputs;delete row.nullish_inputs;}
const offenders=rows.filter(r=>r.nullish>0),unvisited=rows.filter(r=>r.visits===0);
const summary={projects:301,control:JSON.parse(fs.readFileSync(path.join(control,'report.json'))),instrumented:JSON.parse(fs.readFileSync(path.join(instrumented,'report.json'))),byte_identical_project_outputs:301,instrumented_sites:checked.length,reached_sites:rows.length-unvisited.length,unvisited_sites:unvisited.length,nullish_sites:offenders.length,nullish_evaluations:offenders.reduce((n,r)=>n+r.nullish,0),null_evaluations:offenders.reduce((n,r)=>n+r.null,0),undefined_evaluations:offenders.reduce((n,r)=>n+r.undefined,0),projects_with_nullish:records.filter(r=>Object.keys(r.nullish).length).length,first_nullish_sites:[...new Set(records.map(r=>r.first_nullish))]};
fs.writeFileSync(path.join(__dirname,'runtime-projects.jsonl'),records.map(r=>JSON.stringify(r)).join('\n')+'\n');
for(const [name,data] of Object.entries({'runtime-summary.json':summary,'runtime-sites.json':rows,'input-sets.json':sets}))fs.writeFileSync(path.join(__dirname,name),JSON.stringify(data,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'NULLISH.tsv'),'site\tvalue\tevaluations\tinput set\texpression\n'+offenders.map(r=>[r.site,r.null?'null/undefined':'undefined',r.nullish,r.nullish_input_set,r.expression.replace(/\s+/g,' ')].join('\t')).join('\n')+'\n');
console.log(JSON.stringify(summary,null,2));
