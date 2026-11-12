import collections,importlib.util,json,pathlib
S=pathlib.Path('/workspace/ldpi')
spec=importlib.util.spec_from_file_location('control',S/'fixture-control.py');control=importlib.util.module_from_spec(spec);spec.loader.exec_module(control)
manifest=json.loads((S/'chunks.json').read_text());base_header=None;result={}
for kind in ['shallow','fixed']:
 data=[];states=[];reports=[];statuses=[];seen=set()
 for g in manifest['groups']:
  p=S/'slices'/f'{g["id"]}-{kind}';st=json.loads((p/'status.json').read_text())
  assert st['child_exit_code']==0 and not st['deadline_expired'],(p,st)
  rows=[json.loads(l) for l in (p/'census.jsonl').open()]
  assert rows[0]['status']=='measurement' and rows[0]['latent_mode']=='first-error' and not rows[0]['checker_rejected']
  if base_header is None:base_header=rows[0]
  assert rows[0]==base_header
  assert rows[-1]['status']=='relation_report'
  expected={f['path']:f for f in g['files']};actual={r['file']:r for r in rows[1:-1]}
  assert expected.keys()==actual.keys() and not seen.intersection(actual)
  seen.update(actual)
  for f,r in actual.items():assert len(r['units'])==expected[f]['units'],f
  report=rows[-1]['relation_overflow']
  assert report['after_load']==report['end']=='' and report['cache_sizes']['flaggedPairs']==report['cache_sizes']['ts2859']==0
  assert not rows[-1]['findings']
  reports.append({'group':g['id'],'report':report});statuses.append(st);data.append(rows)
  states.append(json.loads((p/'instantiations.json').read_text()))
 assert seen==set(base_header['sources']) and len(seen)==81
 files=control.merge(data,states)
 assert sum(len(r['units']) for r in files.values())==manifest['units']==4566
 out=S/f'combined-{kind}';out.mkdir(exist_ok=True)
 header=dict(base_header,segmented_observation={'groups':6,'files':81,'units':4566,'selection':'disjoint complete source files; full entry program retained'})
 trailer={'measurement':header['measurement'],'status':'relation_report','relation_overflow':{'per_unit':False,'segmented_reports':reports,'attribution':'per child; no fabricated whole-run cache budget'},'findings':[]}
 with (out/'census.jsonl').open('w') as stream:
  for row in [header,*[files[p] for p in sorted(files)],trailer]:stream.write(json.dumps(row,sort_keys=True,separators=(',',':'),ensure_ascii=False)+'\n')
 metadata={'complete':True,'children':statuses,'file_records':len(files),'unit_records':4566,'status_counts':dict(collections.Counter(u['status'] for r in files.values() for u in r['units']))}
 (out/'metadata.json').write_text(json.dumps(metadata,indent=2)+'\n');result[kind]={k:v for k,v in metadata.items() if k!='children'}
(S/'combined-summary.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
