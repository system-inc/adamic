#!/usr/bin/env python3
"""Independently recount retained observations and verify archive hashes."""
import gzip,hashlib,json,pathlib,re,tarfile
p=pathlib.Path(__file__).resolve().parent
for line in (p/'SHA256SUMS').read_text().splitlines():
 digest,name=line.split('  ',1);assert hashlib.sha256((p/name).read_bytes()).hexdigest()==digest,name
for archive in p.rglob('*.tar.gz'):
 with tarfile.open(archive) as t:
  hashes=t.extractfile('SHA256SUMS').read().decode()
  for line in hashes.splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(t.extractfile(name).read()).hexdigest()==digest,(archive,name)
totals={s:dict(elements=0,checks=0,units=0,finding_free_units=0) for s in ['main','head']};seen=set()
for f in (p/'pieces').glob('*.json'):
 summary=json.loads(f.read_text());records=summary.get('records',[summary])
 with tarfile.open(f.with_suffix('.tar.gz')) as t:
  for side in ['main','head']:
   raw=t.extractfile(side+'-count.jsonl').read();baseline=t.extractfile(side+'-meter.jsonl').read()
   assert re.sub(rb'"(?:ir_measured|view_elements|view_element_checks)":(?:true|false|[0-9]+),',b'',raw)==baseline
   rows=list(map(json.loads,raw.splitlines()))
   assert rows[0]['latent_mode']=='full'
   assert rows[0]['checker_rejected']==False
   observed={r['file'].split('-adapted-complete/',1)[1]:r for r in rows[1:]}
   assert set(observed)=={r['file'] for r in records}
   for record in records:
    row=observed[record['file']];units=row['units'];stops={x['unit'] for x in row['findings'] if x['phase']=='lowering'}
    elements=sum(u['view_elements'] for u in units);checks=sum(u['view_element_checks'] for u in units)
    free=sum(u['status']=='attempted' and u['where'] not in stops for u in units)
    expect=record['sides'][side];assert (elements,checks,len(units),free)==(expect['view_elements'],expect['view_element_checks'],expect['units'],expect['finding_free_attempted'])
    for key,value in zip(['elements','checks','units','finding_free_units'],[elements,checks,len(units),free]):totals[side][key]+=value
  for record in records:
   assert record['file'] not in seen;seen.add(record['file'])
with tarfile.open(p/'source-and-build.tar.gz') as t:
 rows=list(map(json.loads,gzip.decompress(t.extractfile('v3-source-bound.jsonl.gz').read()).splitlines()))
s=json.loads((p/'source-summary.json').read_text());assert len(rows)==s['upper_bound'];assert sum(r['readonly'] for r in rows)==s['readonly_subset']
for form,counts in s['forms'].items():
 assert sum(r['form']==form for r in rows)==counts['upper_bound']
 assert sum(r['form']==form and r['readonly'] for r in rows)==counts['readonly_subset']
for file,counts in s['files'].items():
 assert sum(r['file']==file for r in rows)==counts['upper_bound']
 assert sum(r['file']==file and r['readonly'] for r in rows)==counts['readonly_subset']
seven={'index','at','for-of','spread','slice','map','forEach'}
core=[r for r in rows if r['form'] in seven]
assert len(core)==s['seven_form_upper_bound']
assert sum(r['readonly'] for r in core)==s['seven_form_readonly_subset']
for file,counts in s['seven_form_files'].items():
 assert sum(r['file']==file for r in core)==counts['upper_bound']
 assert sum(r['file']==file and r['readonly'] for r in core)==counts['readonly_subset']
assert set(s['files'])==set(json.loads((p/'plan.json').read_text())['files'])
files=[x for x in p.rglob('*') if x.is_file()];assert len(files)<200
for x in files:
 if x.name.endswith('.tar.gz'):continue
 if x.suffix=='.go':raise AssertionError(x)
 text=x.read_text();assert '\u2013' not in text and '\u2014' not in text,x
print(json.dumps(dict(paired_files=len(seen),planned_files=len(s['files']),totals=totals,seven_form_upper_bound=s['seven_form_upper_bound'],seven_form_readonly_subset=s['seven_form_readonly_subset'],source_upper_bound=s['upper_bound'],readonly_subset=s['readonly_subset'],files=len(files),archives='verified',meter_byte_identity='verified'),indent=2))
