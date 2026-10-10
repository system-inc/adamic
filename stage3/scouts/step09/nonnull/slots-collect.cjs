const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const [results,logs]=process.argv.slice(2),driver=path.resolve(__dirname,'../../../drivers/tsc');
const ids=[...JSON.parse(fs.readFileSync(path.join(driver,'selection.json'))).cases.map(c=>c.id),'tiny'].sort();
const report=JSON.parse(fs.readFileSync(path.join(results,'report.json')));assert.equal(report.passed,301);assert.deepEqual(report.failed,[]);
const manifest=JSON.parse(fs.readFileSync(path.join(__dirname,'placeholder-slots.json'))),rows=new Map(manifest.slots.map(s=>[s.id,{site:s.id,slot:s.slot,initializations:0,first_reads:0,before_write:0,after_write:0,nullish_first_reads:0,writes:0,opaque_calls:0,locations:{},before_write_inputs:[],opaque_inputs:[]}])),records=[];
for(const id of ids){
 const record=JSON.parse(fs.readFileSync(path.join(logs,id+'.json')));assert.equal(path.basename(record.input),id);
 const expected=path.join(driver,id==='tiny'?'tiny':'corpus/'+id);
 for(const suffix of ['stdout','stderr','exit'])assert.deepEqual(fs.readFileSync(path.join(results,id,'actual.'+suffix)),fs.readFileSync(path.join(expected,'golden.'+suffix)),id+':'+suffix);
 for(const [site,r] of Object.entries(record.slots)){
  const row=rows.get(site);assert.ok(row,site);assert.equal(r.first_reads,r.before_write+r.after_write);assert.ok(r.first_reads<=r.initializations);assert.ok(r.nullish_first_reads<=r.first_reads);
  for(const k of ['initializations','first_reads','before_write','after_write','nullish_first_reads','writes','opaque_calls']){assert.ok(Number.isSafeInteger(r[k])&&r[k]>=0);row[k]+=r[k];}
  for(const [where,count] of Object.entries(r.locations))row.locations[where]=(row.locations[where]||0)+count;
  if(r.before_write)row.before_write_inputs.push(id);if(r.opaque_calls)row.opaque_inputs.push(id);
 }
 records.push({input:id,slots:record.slots});
}
const all=[...rows.values()];for(const row of all){row.unread_epochs=row.initializations-row.first_reads;row.answer=row.before_write?'NO: first read before write':row.opaque_calls?'UNKNOWN: opaque array operations':!row.first_reads?'UNOBSERVED: no first read':'YES for observed first reads';}
const counts={};for(const r of all)counts[r.answer]=(counts[r.answer]||0)+1;
const summary={projects:301,byte_identical_project_outputs:301,placeholder_sites:50,counts,initializations:all.reduce((n,r)=>n+r.initializations,0),first_reads:all.reduce((n,r)=>n+r.first_reads,0),before_write:all.reduce((n,r)=>n+r.before_write,0),opaque_calls:all.reduce((n,r)=>n+r.opaque_calls,0),driver:report};
for(const [name,data] of Object.entries({'slot-observations.json':all,'slot-summary.json':summary}))fs.writeFileSync(path.join(__dirname,name),JSON.stringify(data,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'slot-projects.jsonl'),records.map(r=>JSON.stringify(r)).join('\n')+'\n');
fs.writeFileSync(path.join(__dirname,'SLOTS.tsv'),'site\tslot\tanswer\tinitializations\tfirst_reads\tbefore_write\tafter_write\tunread_epochs\topaque_calls\tbefore_write_inputs\n'+all.map(r=>[r.site,r.slot.target||r.slot.key,r.answer,r.initializations,r.first_reads,r.before_write,r.after_write,r.unread_epochs,r.opaque_calls,r.before_write_inputs.join(',')||'-'].join('\t')).join('\n')+'\n');
console.log(JSON.stringify(summary,null,2));
