"""Validate preserved ordinary observations, compiler pins and frozen source inputs."""
import gzip,hashlib,json,pathlib,subprocess
O=pathlib.Path(__file__).resolve().parent;R=O.parents[2];T=pathlib.Path('/workspace/container-fit-adapted')
sha=lambda b:hashlib.sha256(b).hexdigest()
manifest=json.loads((O/'evidence/source-manifest.json').read_text())
assert manifest=={str(p.relative_to(T)):sha(p.read_bytes()) for p in sorted((T/'src').rglob('*')) if p.is_file()}
expected={str(p.relative_to(T)) for p in (T/'src/compiler').rglob('*.ts')}
for label in ['base','head']:
 D=O/'evidence/per-file'/label/'all';m=json.loads((D/'metrics.json').read_text())
 for name,value in m.items():
  if name.endswith('_sha256'):
   with gzip.open(D/(name[:-7]+'.gz'),'rb') as f:assert sha(f.read())==value
 with gzip.open(D/'records.jsonl.gz','rt') as f:rows=[json.loads(x) for x in f]
 actual=[r['roots'][0].replace(str(T)+'/', '') for r in rows if len(r['roots'])==1 and r['roots'][0].endswith('.ts')]
 if m['returncode']==0:assert len(actual)==79 and set(actual)==expected and len(set(actual))==79
 assert m['wall_seconds']<1250
 print(label,'exit',m['returncode'],'roots',len(actual),'wall_seconds',m['wall_seconds'])
f=json.loads((O/'evidence/latent-build-failure.json').read_text());assert not f['latent_measurements_started']
assert subprocess.check_output(['git','branch','--show-current'],cwd=R,text=True).strip()=='typescript/container-fit-census'
assert all(p.startswith('stage3/census/container-fit/') for p in subprocess.check_output(['git','diff','--name-only','d4c6b07a','HEAD'],cwd=R,text=True).splitlines())
assert '\u2013' not in (O/'REPORT.md').read_text() and '\u2014' not in (O/'REPORT.md').read_text()
print('Source hashes, raw archive hashes, root coverage, time bounds and report-only branch scope verified.')

outcome=json.loads((O/'outcome.json').read_text())
for key in ['latent_first_stops','latent_regressions','latent_cap_rows','pair_counts','latent_time_sums','head_log_byte_identity']:assert outcome[key] is None
print('Unavailable latent outcomes are null, not zero or empty result lists.')
