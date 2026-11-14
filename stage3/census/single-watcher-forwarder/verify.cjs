'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const out=__dirname,repo='/workspace/scratch/single-watcher-forwarder',lane='/workspace/scratch/single-watcher-lane';
const read=f=>JSON.parse(fs.readFileSync(path.join(out,f),'utf8'));
const sum=read('summary.json');assert.equal(sum.adapted.helper_cast_survives,false);assert.equal(sum.mutant.helper_cast_survives,true);assert.equal(sum.mutant.source_aligned_added,0);assert.equal(sum.mutant.source_aligned_gone,0);assert.equal(sum.adapted.source_aligned_added,2);assert.equal(sum.adapted.source_aligned_gone,2);assert.equal(sum.adapted.source_aligned_changes,1);
for(const label of ['base','adapted','mutant']){const s=read(label+'/stock.json');assert.equal(s.version,'6.0.3');assert.equal(s.roots.length,79);assert.equal(s.diagnostics.length,0);const rows=fs.readFileSync(path.join(out,label,'latent.jsonl'),'utf8').trim().split('\n').map(JSON.parse);assert.equal(rows[0].latent_mode,'full');assert.equal(rows[0].module_format,'CommonJSBundle');assert.equal(rows[0].diagnostics.length,0);assert.equal(rows.at(-1).status,'relation_report');assert.equal(read(label+'/latent-status.json').exit_code,0);}
const traces=read('runtime-traces.json');for(const kind of ['polling','fs'])assert.deepEqual(traces.results[kind].before,traces.results[kind].after);
const count=read('callers.json');assert.equal(count.callCount,2);assert.equal(count.otherCalls,0);assert.equal(count.identifierUses,3);
const pins=JSON.parse(fs.readFileSync(repo+'/stage3/adapt/99-correctness-single-watcher-forwarder/owners.json'));for(const p of pins)for(const k of ['before','after'])assert.equal(crypto.createHash('sha256').update(p[k]).digest('hex'),p[k==='before'?'ownerBeforeSha256':'ownerAfterSha256']);
const resolved=JSON.parse(fs.readFileSync(repo+'/stage3/adapt/99-correctness-single-watcher-forwarder/resolved-sites.json'));assert.deepEqual(resolved.resolves.map(r=>[r.file,r.line,r.column]),[['src/compiler/sys.ts',516,30],['src/compiler/sys.ts',516,58]]);
const commands=fs.readFileSync(path.join(out,'commands.jsonl'),'utf8').trim().split('\n').map(JSON.parse);for(const name of ['contract','scope','runtime-after-event','types-base','types-adapted','stock-base','stock-adapted','stock-mutant','generated-freshness','latent-audit'])assert(commands.some(x=>x.name===name&&x.exit===0),name);
if(process.argv.includes('--lane')){const execution=JSON.parse(fs.readFileSync(lane+'/execution.json'));assert.equal(execution.apply_exit,0);assert.equal(JSON.parse(fs.readFileSync(lane+'/report.json')).status,'pass');
 for(const f of ['src/compiler/sys.ts'])assert.equal(fs.readFileSync('/workspace/scratch/single-watcher-after/'+f,'utf8').replace(/\r\n/g,'\n'),fs.readFileSync(lane+'/adapted-tree/'+f,'utf8').replace(/\r\n/g,'\n'));
}
console.log('PASS: evidence internally complete'+(process.argv.includes('--lane')?' and full lane matches measured source':''));
