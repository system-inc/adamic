from pathlib import Path
import json,hashlib,re,sys,gzip
out=Path(__file__).resolve().parent;repo=Path('/workspace/ancestor-adapt');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def read(name):return json.loads((out/name).read_text())
required=['contract','apply','replay','existing-pins','types-before','types-after','build-before','build-after','build-comparison','stock-before','stock-after','census-before','census-after','census-mutant','census-delta','generate-shards','lane']
for n in required:
 s=read(n+'.status.json');assert s['child_exit']==0,(n,s['child_exit']);assert not s['timed_out']and s['seconds']<1800
for n in ['types-before-extended','types-after-extended']:
 s=read(n+'.status.json');assert s['child_exit']==1 and not s['timed_out'] and s['seconds']<1800
assert read('extra-adamic-extended-comparison.json')['diagnostics_identical_after_root_normalization']
for label in ['before','after']:
 s=read('stock-'+label+'.json');assert len(s['roots'])==79 and len(s['extendedRoots'])==80;assert not s['diagnostics']and not s['extendedDiagnostics']and not s['emitSkipped'];assert set(s['roots'])==set(read('types-'+label+'-roots.json'))
assert read('stock-before.json')['outputs']==read('stock-after.json')['outputs'];assert len(read('stock-before.json')['outputs'])==89
b=read('build-comparison.json');assert b['javascriptOutputs']==10 and b['publicDeclarations']==2 and all(x['identical']and x['beforeSha256']==x['afterSha256']for x in b['rows'])
assert read('census-delta.json')['removed']==[['Refused','path.ts:1105:28',"a cast the runtime can't check"]]and not read('census-delta.json')['added']
for label in ['before','after','mutant']:
 s=read('census-'+label+'.status.json');assert 'LATENT_FULL=1'in s['command']and 'LATENT_ASSERT_NO_OUTPUT=1'in s['command']and any(x.startswith('LATENT_ONLY_FILE=')for x in s['command'])
 rows=[json.loads(l)for l in (out/('census-'+label+'.jsonl')).read_text().splitlines()];assert not rows[0]['diagnostics']and not rows[0]['checker_rejected'];assert len([r for r in rows if 'file'in r])==1
p=read('provenance.json')
for f,h in p['adaptation_hashes'].items():assert sha(repo/f)==h
for f,key in [('sanctioned-api.json','sanctioned_api_sha256'),('shard-plan.json','shard_plan_sha256')]:assert sha(repo/'stage3/lane'/f)==p[key]
for proof in read('sanction-proof-hashes.json')['proofs']:assert sha(repo/proof['file'])==proof['recorded']==proof['actual']
rows=[]
lane_text=(out/'lane.log').read_text() if (out/'lane.log').exists() else gzip.decompress((out/'lane.log.gz').read_bytes()).decode()
for l in lane_text.splitlines():
 try:rows.append(json.loads(l))
 except json.JSONDecodeError:raise AssertionError('invalid Go JSON stream')
fail=[r for r in rows if r.get('Action')=='fail'];assert not fail,fail
passes={r['Test']for r in rows if r.get('Action')=='pass'and'Test'in r};shards={t for t in passes if t.startswith('TestShard_')}
manifest=json.loads((repo/'stage3/lane/shards.json').read_text());expected={'TestShard_'+''.join(p[:1].upper()+p[1:]for p in u['name'].split('-'))for u in manifest['units']};assert shards==expected and len(shards)==398
assert 'TestLaneSanctionedApiFreshnessPython'in passes
package=[r for r in rows if r.get('Action')=='pass'and'Test'not in r];assert len(package)==1
command=read('lane.status.json')['command'];assert './stage3/lane'in command and not any(x.startswith('-run')for x in command)
skipped=[r.get('Test')for r in rows if r.get('Action')=='skip']
summary={'passed':True,'wholeLaneChildExit':0,'packageElapsedSeconds':package[0]['Elapsed'],'testPassEvents':sum(r.get('Action')=='pass'and'Test'in r for r in rows),'uniqueTestsPassed':len(passes),'shardsPassed':len(shards),'shardsExpected':len(expected),'failureEvents':0,'skippedTests':skipped,'sanctionedApiFreshness':'PASS','extra80RootAdamicProbe':'32 unchanged baseline diagnostics; child 1 before and after'}
(out/'audit.json').write_text(json.dumps(summary,indent=2)+'\n');(out/'lane-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary))
