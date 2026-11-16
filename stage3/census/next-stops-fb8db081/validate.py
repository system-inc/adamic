"""Validate all paired roots in each requested format and frozen source/binary evidence."""
import hashlib,json,os,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/next-stops-fb8db081');tree=scratch/'adapted';repo=scratch/'repo'
read=lambda p:json.loads(p.read_text())
provenance=read(base/'provenance.json');sha=provenance['sha']
assert subprocess.check_output(['git','rev-parse',sha],cwd=repo,text=True).strip()==sha
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=scratch/'typescript',text=True).strip()==provenance['upstream']
expected=list(map(str,sorted((tree/'src/compiler').rglob('*.ts'))));assert len(expected)==79
modes=os.environ.get('CENSUS_ANALYZE_FORMATS','CommonJSBundle EcmaScript').split()
statuses={str(p.relative_to(base)):read(p) for p in base.rglob('*-status.json')}
for s in statuses.values():assert s['exit_code']==0 and s['seconds']<1800 and s['process_group_cleaned'] and not s['stopping_reason']
if 'EcmaScript' in modes:
 transition=read(base/'format-transition.json')
 assert transition['binary']==read(base/'binaries.json') and transition['new_apply_or_build'] is False
 checkpoint=transition['commonjs_validated_checkpoint']
 saved=json.loads(subprocess.check_output(['git','show',checkpoint+':stage3/census/next-stops-fb8db081/validation-CommonJSBundle.json'],cwd=repo,text=True))
 assert saved['raw_observations']==158 and saved['duplicates_agree'] and saved['formats']==['CommonJSBundle']
 assert subprocess.call(['git','merge-base','--is-ancestor',checkpoint,'HEAD'],cwd=repo)==0
for mode in modes:
 out=base/'head'/mode;rows=[]
 for i in range(1,80):
  pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
  assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
  assert (out/f'root-{i:02}-status.json').exists();rows.append(pair[0])
 assert [r['roots'][0] for r in rows]==expected
 assert all(r['module_format']==mode for r in rows)
 assert sum(g['count'] for g in read(out/'grouped-first-stops.json'))==79
 assert all(g.get('phase') and g.get('owner') for g in read(out/'grouped-first-stops.json'))
 for a,z in [(1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79)]: assert (out/f'group-{a}-{z}-status.json').exists()
binary=read(base/'binaries.json');assert hashlib.sha256(Path(binary['path']).read_bytes()).hexdigest()==binary['sha256']
info=subprocess.check_output(['go','version','-m',binary['path']],text=True);assert info==(base/'build-info.txt').read_text()
assert 'vcs.revision='+sha in info
manifest=read(base/'adapted-source-manifest.json');assert manifest==read(base/'adapted-source-before-measurement.json')
assert all(hashlib.sha256((tree/n).read_bytes()).hexdigest()==h for n,h in manifest.items())
cs=read(base/'comparison.json');baseline=read(base/'baseline.json');assert len(cs)==79*len(modes)
archived=(base/'base/CommonJSBundle/census.jsonl').read_bytes()
assert hashlib.sha256(archived).hexdigest()==provenance['baseline_rows_sha256']
oldrows=[json.loads(line) for line in archived.splitlines()];assert len(oldrows)==79
assert {r['roots'][0].split('/src/compiler/')[-1]:[r.get('where','-').split('/src/compiler/')[-1],r['kind'],r.get('reason','-')] for r in oldrows}==baseline
for c in cs:assert c['before']==baseline[c['root']] and c['unchanged']==(c['before']==c['after'])
changed=subprocess.check_output(['git','diff',sha,'--name-only'],cwd=repo,text=True).splitlines()
untracked=subprocess.check_output(['git','ls-files','--others','--exclude-standard'],cwd=repo,text=True).splitlines()
assert all(n.startswith('stage3/census/next-stops-fb8db081/') for n in changed+untracked)
assert not list(base.rglob('*.go'))
assert not [p for p in base.rglob('*') if p.is_file() and p.stat().st_size>20_000_000]
processes=subprocess.check_output(['ps','-eo','stat=,args='],text=True).splitlines()
assert not [line for line in processes if len(line.split(None,1))==2 and line.split(None,1)[1].startswith(binary['path']+' ') and not line.split()[0].startswith('Z')]
assert len(list(base.glob('head/*/apply-status.json')))==1 and len(list(base.glob('head/*/build-status.json')))==1
result=dict(no_live_census_children=True,single_apply_and_build=True,formats=modes,roots_per_format=79,raw_observations=158*len(modes),duplicates_agree=True,all_child_exits={n:s['exit_code'] for n,s in statuses.items()},maximum_child_seconds=max(s['seconds'] for s in statuses.values()),adapted_source_hashes_verified=True,binary_verified=True,only_census_directory_changed=True)
(base/('validation-'+modes[-1]+'.json')).write_text(json.dumps(result,indent=2)+'\n')
(base/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print('validation passed',modes)
