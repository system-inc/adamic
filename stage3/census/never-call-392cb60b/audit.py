from pathlib import Path
from collections import Counter
import gzip,json,hashlib,subprocess
out=Path(__file__).resolve().parent;repo=out.parents[2];tree=Path('/workspace/never-call-adapted')
def read(case,side):
 p=out/(case+'-'+side+'.jsonl')
 if not p.exists():p=Path(str(p)+'.gz')
 with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as f:return [json.loads(s) for s in f if s.strip()]
closure=json.loads((out/'closure.json').read_text());expected={str(tree/f['file']) for f in closure['files']}
assert len(expected)==81
for f in closure['files']:assert hashlib.sha256((tree/f['file']).read_bytes()).hexdigest()==f['sha256'],f['file']
probe=json.loads((out/'observer-probe.json').read_text())
assert probe['passed'] and probe['actual_child_exit_code']==7
assert probe['peak_rss_kib_at_exit']==probe['rusage_children_peak_rss_kib']
checks={'exit_observer_probe':probe};allcases=['debug-full','semver-full','tsc-entry-ordered-first']
unit_inventory=json.loads((out/'entry-unit-inventory.json').read_text())
for case in allcases:
 streams={s:read(case,s) for s in ['parent','head']}
 assert streams['parent'][0]==streams['head'][0],case
 for side,rows in streams.items():
  assert rows[-1]['status']=='relation_report',(case,side,'missing trailer')
  assert rows[0]['latent_mode']==('first-error' if case=='tsc-entry-ordered-first' else 'full')
  assert not rows[0]['checker_rejected'] and not rows[0]['diagnostics']
  files={r['file'] for r in rows if 'file' in r};units=[u for r in rows for u in r.get('units',[]) or []]
  if case=='tsc-entry-ordered-first':
   assert files==expected,(side,expected-files,files-expected)
   assert len(units)==unit_inventory['total']==4562,(side,len(units))
   assert rows[0]['entry_reach'] is True and rows[0]['entry_first'] is True and rows[0]['root_files']==[str(tree/'src/tsc/tsc.ts')]
   assert next(r['file'] for r in rows if 'file' in r)==str(tree/'src/tsc/tsc.ts')
  else:assert files=={str(tree/('src/compiler/debug.ts' if case=='debug-full' else 'src/compiler/semver.ts'))}
  t=json.loads((out/(case+'-'+side+'.timing.json')).read_text());assert t['child_exit_code']==0 and not t['silent_cutoff']
  if case=='tsc-entry-ordered-first':
   assert t['peak_rss_exact_exit_observation'] is True
   assert t['raw_wait_status']==t['ptrace_exit_event_wait_status']==0
  checks[case+'-'+side]=dict(files=len(files),units=len(units),findings=sum(len(r.get('findings',[]) or []) for r in rows),statuses=dict(Counter(u['status'] for u in units)),child_exit_code=t['child_exit_code'],relation_overflow=rows[-1]['relation_overflow'])
 d=json.loads((out/(case+'-paired.json')).read_text())
 for f,sha in d['source_hashes'].items():assert hashlib.sha256((tree/f).read_bytes()).hexdigest()==sha
 assert d['normalization'] is True
 # Independently inspect all finding-free source identities, including phases beyond refusal_scan.
 if case=='tsc-entry-ordered-first':
  for view in ['raw_findings','distinct_reads']:
   v=d[view];new_stop_units={r['unit']['where']:r for r in v['unit_first_stop_delta'] if r['base'] is None and r['head'] is not None}
   changed_blocked={r['unit']['where'] for r in v['unit_first_stop_delta'] if r['base'] is not None and r['head'] is not None}
   unclassified=[]
   for r in v['new_sites']:
    f=r['finding']
    if f['phase'] in ('lowering','deferred') and f['unit'] in changed_blocked:continue
    if f['unit'] in new_stop_units:continue
    if f['phase']=='refusal_scan':continue
    if f['kind']=='panic':continue
    unclassified.append(r)
   checks['extra_new_site_regressions_'+view]=unclassified
 # The tool itself validates exact unit identity and attempted eligibility while pairing.
for cwd,pin in [(repo,'bdbc7386'),(Path('/workspace/never-call-head'),'392cb60b')]:
 r=subprocess.run(['git','diff','--quiet',pin,'--','internal','stage3/census/latent','stage3/apply.py'],cwd=cwd)
 print('compiler/tool source diff child exit code:',r.returncode,flush=True);assert r.returncode==0
checks['closure_hashes_unchanged']=True
checks['original_source_root_count']=1
checks['unit_inventory_total']=unit_inventory['total']
(out/'audit.json').write_text(json.dumps(checks,indent=2)+'\n')
print(json.dumps(checks,indent=2))
