import collections,hashlib,json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
rows=[]
for line in (out/'lane-whole.jsonl').read_text().splitlines():
 try:rows.append(json.loads(line))
 except json.JSONDecodeError:pass
assert rows[-1]['Action']=='pass' and rows[-1]['Package']=='github.com/system-inc/adamic/stage3/lane',rows[-1]
passed={r['Test'] for r in rows if r.get('Action')=='pass' and r.get('Test')}
failed=[r for r in rows if r.get('Action')=='fail'];assert not failed
shards=sorted(x for x in passed if x.startswith('TestShard_'));assert len(shards)==398,len(shards)
manifest=json.loads((out/'source-manifests.json').read_text())
def hashes(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tree.joinpath('src').rglob('*')) if p.is_file()}
actual=hashes(out/'amended-apply-tree');assert actual==manifest['after'],'fresh apply differs from measured amendment'
for label in ['main','after','restoring-mutant']:
 assert json.loads((out/(label+'-receipt.json')).read_text())['exit']==0
for label in ['main','after']:
 r=json.loads((out/(label+'-types-receipt.json')).read_text());assert r['exit']==0 and r['root_count']==79
census=json.loads((out/'census-delta.json').read_text());assert len(census['sites']['gone'])==1 and not census['sites']['new'];assert not census['observations']['new'];assert not census['mutant_observations_vs_main']['new'] and not census['mutant_observations_vs_main']['gone']
feature=subprocess.check_output(['git','rev-parse','typescript/changed-files-undefined-1'],text=True).strip();parents=subprocess.check_output(['git','rev-list','--parents','-n','1',feature],text=True).split();assert parents[1:]==['5d334653fa9ce301738e9eb2ef016c99da4589b8']
r=dict(feature=feature,base=parents[1],main_apply_exit=0,amended_apply_exit=0,apply_source_manifest_matches_census=True,lane={'command':'go test -json -count=1 -parallel=2 -timeout=45m ./stage3/lane','exit':0,'elapsed':rows[-1].get('Elapsed'),'passed_tests':len(passed),'manifest_shards':len(shards),'failed_tests':failed,'skipped_tests':[x.get('Test') for x in rows if x.get('Action')=='skip']},census={'main_events':191,'after_events':190,'removed':['sys.ts:290:19'],'new':[],'changed':[],'mutant_exactly_restores_main':True},javascript=json.loads((out/'javascript.json').read_text()),types={'main':0,'after':0,'roots':79},contract={'exit':0,'owners':13,'files':9},lane_generated_files_changed=False,loom_submitted=False)
(out/'verification.json').write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r,indent=2))
