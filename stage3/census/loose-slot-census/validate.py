"""Audit experiment receipts and reporting boundaries; no compiler acceptance path."""
from pathlib import Path
import collections,hashlib,json,re,subprocess
out=Path(__file__).resolve().parent;s=Path('/workspace/loose-slot-scratch');repo=s/'head'
def read(p):return json.loads(p.read_text())
provenance=read(out/'evidence/provenance.json');assert provenance['adapted_identical']
for which in ['main','head']:
 tree=s/('adapted-'+which)/'src/compiler';files={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(tree.rglob('*.ts'))};assert len(files)==79 and files==provenance[which]['files']
rows={};resources=[]
for which in ['main','head']:
 rows[which]=[]
 for i in range(1,80):
  base=out/'evidence'/which/f'root-{i:02}';resource=read(base.with_suffix('.time.json'));assert resource['exit_code']==0 and resource['seconds']>0 and resource['peak_rss_KiB']>0;resources.append(resource)
  row=read(base.with_suffix('.json'));assert row['module_format']=='CommonJSBundle';row.pop('seconds');row.pop('roots')
  for key in ['where','reason','fix']:
   if key in row:row[key]=row[key].replace(str(s/('adapted-'+which)/'src/compiler')+'/','')
  rows[which].append(row)
diff=[i+1 for i,(a,b)in enumerate(zip(rows['main'],rows['head']))if a!=b]
comparison=read(out/'evidence/comparison.json');assert comparison['unchanged']==79-len(diff) and len(comparison['differences'])==len(diff)
controls=out/'evidence/controls';production=read(controls/'arrays-production.json');assert production['kind']=='NotYet' and production['where'].endswith('arrays.a:48:1') and production['reason'].endswith("a cycle the cycle finder can't follow through unknown or object (#yeep0sx links these slots)")
latent=[json.loads(line)for line in (controls/'arrays-latent.jsonl').read_text().splitlines()];assert latent[0]['checker_rejected']is False and latent[-1]['status']=='relation_report' and all(not row.get('findings')for row in latent)
for normal,mutant in [('arrays-type','arrays-mutant')]:
 a=read(controls/(normal+'.json'));m=read(controls/(mutant+'.json'));assert any(e['Where'].endswith('arrays.a:48:1')for e in a['Events'])and not m['Events'];assert {k:v for k,v in a.items()if k!='Events'}=={k:v for k,v in m.items()if k!='Events'}
for name in ['k17_captured_unknown','k18_entry_then_declared']:
 ordinary=read(controls/(name+'-production.json'));typed=read(controls/(name+'-type.json'));assert ordinary['kind']=='NotYet' and any(e['Where']==ordinary['where']for e in typed['Events'])
typed=read(out/'evidence/type-whole.json');mutant=read(out/'evidence/type-whole-mutant.json');assert typed['Files']==79 and not mutant['Events'] and {k:v for k,v in typed.items()if k!='Events'}=={k:v for k,v in mutant.items()if k!='Events'}
for p in [*controls.glob('*.time.json'),*out.glob('evidence/type-*.time.json')]:
 r=read(p);assert r['exit_code']==0 and r['seconds']>0 and r['peak_rss_KiB']>0
sites={e['Where']for e in typed['Events']};listing=(out/'SITES.md').read_text()
for where in sites:
 p,line,col=where.rsplit(':',2);file=str(Path(p).relative_to(s/'adapted-head/src/compiler'));assert f'## {file}:{line}:{col}'in listing;assert Path(p).read_text().splitlines()[int(line)-1]in listing
report=(out/'REPORT.md').read_text();assert report.index('## Control first')<report.index('## Production next stops')<report.index('## Sites past the walls');assert 'UNKNOWN'in report and 'partial type-level'in report
hashes=read(out/'evidence/binary-sha256.json')
for name in ['head-census','main-census','latent','type-count','type-count-mutant']:assert hashlib.sha256((s/name).read_bytes()).hexdigest()==hashes[name]
assert hashlib.sha256((out/'type_count.go.txt').read_bytes()).hexdigest()==hashes['type-instrument-source']
assert hashlib.sha256((s/'mutant-loose-slots.go').read_bytes()).hexdigest()==hashes['mutant-loose-slots-source']
assert read(out/'evidence/type-whole-mutant.time.json')['command'][0]==str(s/'type-count-mutant')
assert 'judge 0/33' in (out/'evidence/type-whole-mutant.log').read_text()
changed=subprocess.check_output(['git','diff','--name-only','7e8cfaafb'],cwd=repo,text=True).splitlines();assert all(p.startswith('stage3/census/loose-slot-census/')for p in changed);assert not any(p.endswith('.go')for p in changed)
result={'status':'passed','production_observations':158,'roots_per_sha':79,'unchanged_pairs':79-len(diff),'changed_indices':diff,'positive_control':'arrays production stop / latent zero / type positive / mutant zero','captured_and_literal_controls':'k17 and k18 exact-site match','type_observed_events':len(typed['Events']),'type_distinct_sites':len(sites),'type_forms':dict(collections.Counter(e['Form']for e in typed['Events'])),'exhaustive_production_site_count':'UNKNOWN','whole_mutant_events':0,'site_quotes_verified':len(sites),'ordinary_lower_source_edits':0}
(out/'evidence/validation.json').write_text(json.dumps(result,indent=2)+'\n');print(result)
