import hashlib,json,subprocess,tarfile
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];tree=Path('/workspace/bcbe5d12-adapted');scratch=Path('/workspace/scratch/b5w3ycg')
def read(name):return json.loads((out/name).read_text())
groups=[read(f'group-{i}-rows.json') for i in range(1,5)]
assert list(map(len,groups))==[20,20,20,19]
rows=sum(groups,[]);roots=[r['roots'][0] for r in rows]
assert roots==sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts')) and len(set(roots))==79
assert all(read(f'root-{i:02}-status.json')['exit_code']==0 for i in range(1,80))
archive=out/'raw-streams.tar.gz';assert hashlib.sha256(archive.read_bytes()).hexdigest()==(out/'raw-streams.tar.gz.sha256').read_text().split()[0]
with tarfile.open(archive) as tar:
 manifest=read('archive-manifest.json');assert set(manifest)==set(tar.getnames())
 for name,m in manifest.items():
  data=tar.extractfile(name).read();assert len(data)==m['bytes'] and hashlib.sha256(data).hexdigest()==m['sha256']
 for group in range(1,5):
  raw=[json.loads(l) for l in tar.extractfile(f'ordinary/group-{group}.jsonl').read().splitlines()]
  assert len(raw)==2*len(groups[group-1])
  for index,row in enumerate(groups[group-1]):
   assert raw[2*index]==row
   assert {k:v for k,v in raw[2*index].items() if k!='seconds'}=={k:v for k,v in raw[2*index+1].items() if k!='seconds'}
summary=read('grouped-first-stops.json');comparisons=read('comparison.json')
assert sum(g['count'] for g in summary)==79 and len(comparisons)==79
assert {r for g in summary for r in g['roots']}=={r['root'] for r in comparisons}
assert {g['where'] for g in summary if g['new_group']}=={r['where'] for r in read('source-excerpts.json')}=={r['where'] for r in read('new-stop-emitters.json')}
for row in read('source-excerpts.json'):assert (out/row['name']).read_bytes()==(row['text']+'\n').encode()
for name,digest in read('adapted-source-manifest.json').items():assert hashlib.sha256((tree/name).read_bytes()).hexdigest()==digest
assert not read('cleanup.json')['remaining']
assert all(p.suffix!='.go' for p in out.rglob('*'))
text=(out/'REPORT.md').read_text();assert '\u2013' not in text and '\u2014' not in text
p=subprocess.run(['git','diff','--exit-code','bcbe5d1244ee8a218d860415dea19cca114bae46','--','.',':(exclude)stage3/census/next-stops-bcbe5d12'],cwd=repo,capture_output=True)
(scratch/'final-input-diff.log').write_bytes(p.stdout+p.stderr);print('git diff child_exit='+str(p.returncode));p.check_returncode()
result=dict(roots=79,group_sizes=[20,20,20,19],raw_observations=158,duplicate_observations_agree=True,archive_sha256_verified=True,archive_member_hashes_verified=True,source_manifest_verified=True,compiler_and_adapter_inputs_unchanged=True,remaining_census_processes=0,root_order_status=read('root-order.json')['status'])
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
