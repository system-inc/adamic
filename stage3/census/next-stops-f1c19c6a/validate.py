import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
rows=[]
for n,size in enumerate([20,20,20,19],start=1):
 group=json.loads((out/f'group-{n}-rows.json').read_text());assert len(group)==size
 raw=[json.loads(line) for line in (out/f'group-{n}.jsonl').read_text().splitlines()];assert len(raw)==2*size
 for i,row in enumerate(group):
  first,second=raw[2*i:2*i+2]
  assert first==row
  assert {k:v for k,v in first.items() if k!='seconds'}=={k:v for k,v in second.items() if k!='seconds'}
 rows.extend(group)
roots=json.loads((out/'roots.json').read_text());assert len(roots)==79 and roots==sorted(roots,key=Path)
assert [r['roots'][0] for r in rows]==roots and len(set(roots))==79
assert rows==[json.loads(line) for line in (out/'census.jsonl').read_text().splitlines()]
for i in range(1,80):
 s=json.loads((out/f'root-{i:02}-status.json').read_text());assert s['exit_code']==0 and s['process_group_cleaned']
entry=json.loads((out/'entry-census-status.json').read_text());assert entry['process_group_cleaned']
import hashlib
manifest=json.loads((out/'apply-source-manifest.json').read_text())
adapted=Path('/workspace/scratch/next-stops-adapted')
assert manifest=={str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((adapted/'src/compiler').rglob('*')) if p.is_file()}
for binary in json.loads((out/'binaries.json').read_text()).values():
 assert hashlib.sha256(Path(binary['path']).read_bytes()).hexdigest()==binary['sha256']
assert not any(arg.startswith('-overlay') for arg in json.loads((out/'build-status.json').read_text())['command'])
assert any(arg.startswith('-overlay') for arg in json.loads((out/'entry-build-status.json').read_text())['command'])
assert subprocess.check_output(['git','-C','/workspace/scratch/typescript-source','rev-parse','HEAD'],text=True).strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
single={}
for name in ['debug','semver']:
 records=[json.loads(line) for line in (out/(name+'-full.jsonl')).read_text().splitlines()]
 assert records[0]['latent_mode']=='full'
 assert records[-1]['status']=='relation_report'
 files=[r for r in records if 'file' in r]
 assert len(files)==1 and files[0]['file']==str(adapted/'src/compiler'/ (name+'.ts'))
 findings=[f for r in records for f in r.get('findings') or []]
 ordered=json.loads((out/(name+'-full-findings.json')).read_text())
 order=lambda f:(f['where'].rsplit(':',2)[0],int(f['where'].rsplit(':',2)[1]),int(f['where'].rsplit(':',2)[2]))
 assert sorted(findings,key=order)==ordered
 status=json.loads((out/(name+'-full-status.json')).read_text())
 assert status['exit_code']==0 and status['process_group_cleaned']
 single[name]=dict(records=len(findings),complete=True,selected_file=files[0]['file'],child_exit=0)
pending=[]
for line in subprocess.check_output(['ps','-eo','pid=,comm='],text=True).splitlines():
 pid,comm=line.split(maxsplit=1)
 if comm.startswith(('next-stops-cens','next-stops-entr','next-stops-late')):pending.append({'pid':int(pid),'comm':comm})
assert not pending,pending
source=subprocess.run(['git','diff','--quiet','HEAD','--','.',':(exclude)stage3/census/next-stops-f1c19c6a'])
assert source.returncode==0
sha=json.loads((out/'provenance.json').read_text())['sha'];assert subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()==sha
for f in out.glob('*.py'):
 import ast
 ast.parse(f.read_text(),filename=str(f))
text=(out/'REPORT.md').read_text();assert '\u2014' not in text and '\u2013' not in text
statuses={}
for label in ['apply','group-1','group-2','group-3','group-4']:
 s=json.loads((out/('checkpoint-push-'+label+'-status.json')).read_text());assert s['exit_code']==0;statuses[label]=s['command'][3].split(':')[0]
result=dict(single_file_full_mode=single,roots=79,groups=[20,20,20,19],raw_observations=158,duplicate_non_timing_agreement=True,adapted_source_hashes_unchanged=True,binary_hashes_verified=True,ordinary_overlay=False,typescript_pin_verified=True,ordinary_child_exits=0,entry_child_exit=entry['exit_code'],all_census_process_groups_cleaned=True,surviving_census_processes=pending,compiler_adapter_source_diff_exit=source.returncode,base_sha=sha,checkpoint_commits=statuses,report_uses_no_em_or_en_dashes=True)
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
