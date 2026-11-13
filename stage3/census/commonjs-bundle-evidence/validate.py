"""Audit frozen input, all raw observations, child bounds, and validation exits."""
import hashlib,json,re,subprocess
from pathlib import Path
out=Path('/workspace/scratch/995spev');repo=Path('/workspace/adamic')
read=lambda name:json.loads((out/name).read_text())
p=read('provenance.json');roots=read('roots.json');assert len(roots)==79 and len(set(roots))==79
for label in ['main','head']:
 for i,root in enumerate(roots,1):
  name=f'{label}-root-{i:02}'
  status=read(name+'-status.json');assert status['exit_code']==0 and status['seconds']<1786,status
  rows=[json.loads(line) for line in (out/(name+'.jsonl')).read_text().splitlines()]
  assert len(rows)==2 and all(row['roots']==[root] for row in rows)
  assert {k:v for k,v in rows[0].items() if k!='seconds'}=={k:v for k,v in rows[1].items() if k!='seconds'},name
  assert all(row.get('module_format','EcmaScript')==('CommonJSBundle' if label=='head' else 'EcmaScript') for row in rows),name
 for checkpoint in range(1,5):
  status=read(f'{label}-roots-{checkpoint}-status.json');assert status['exit_code']==0 and status['seconds']<1786,status
 status=read(label+'-sys-status.json');assert status['exit_code']==0 and status['seconds']<1786
 rows=[json.loads(line) for line in (out/(label+'-sys.jsonl')).read_text().splitlines()]
 assert len(rows)==3 and rows[0]['latent_mode']=='full' and rows[-1]['status']=='relation_report'
 assert rows[1]['file']==roots[41] and Path(rows[1]['file']).name=='sys.ts'
 assert any(u.get('name')=='getNodeSystem' and u['where'].endswith('sys.ts:1475:5') and u['status']=='attempted' for u in rows[1]['units'])
r=read('sys-delta.json');assert not r['new']
assert {(f['kind'],f['where']) for f in r['gone']}=={('NotYet','sys.ts:1494:35'),('NotYet','sys.ts:1494:92'),('NotYet','sys.ts:1494:121'),('NotYet','sys.ts:1724:41'),('Boundary','sys.ts:1494:9'),('Boundary','sys.ts:1724:13')}
for key in ['overlay-tests','guards-final','latent-python-tests','meter-tests-full','lower-source-tests','head-build','latent-build','entry-build','main-build','main-latent-build','sys-analysis','push-change']:
 assert read(key+'-status.json')['exit_code']==0,key
assert read('guard-mutant-status.json')['exit_code']==1
assert 'moduleFormat disagrees: production "moduleFormat", latent ""' in (out/'guard-mutant.log').read_text()
assert 'Ran 94 tests' in (out/'meter-tests-full.log').read_text() and 'skipped' not in (out/'meter-tests-full.log').read_text()
warm=(out/'lower-tests-warm.log').read_text()
assert set(re.findall(r'^--- FAIL: (\w+)',warm,re.M))=={'TestProgramRegionCensusMembership','TestProduct_ProgramRegionCensusPairwise','TestProduct_ProgramRegionCensusIndexed','TestCycleFinderCensusEquality','TestProduct_CycleScanningCensus'}
assert warm.count('ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout')==5
source=out/'adapted';manifest=read('adapted-source-manifest.json')
actual={str(f.relative_to(source)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((source/'src').rglob('*')) if f.is_file()};assert actual==manifest
for name,digest in p['binaries'].items():assert hashlib.sha256((out/name).read_bytes()).hexdigest()==digest
parent=subprocess.check_output(['git','rev-parse',p['head']+'^'],cwd=repo,text=True).strip();assert parent==p['main']
assert read('remote-code-status.json')['exit_code']==0
remote=(out/'remote-code.log').read_text().split()[0];assert remote==p['head']
assert read('measurement-pause.json')['resumed']
for file in out.glob('*-status.json'):
 status=json.loads(file.read_text())
 if 'seconds' in status:assert status['seconds']<1786,(file.name,status)
result=dict(roots_per_pin=79,root_children=158,observations=316,root_children_exit_zero=True,duplicate_observations_agree=True,source_unchanged=True,source_files=len(actual),sys_notyet_gone=4,sys_boundaries_gone=2,sys_new_sites=0,code_one_commit_on_main=True,code_pushed=True,meter_tests=94,meter_skips=0)
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
