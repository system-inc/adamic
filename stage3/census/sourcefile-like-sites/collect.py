import pathlib,json,gzip,collections
out=pathlib.Path(__file__).resolve().parent
seen={}; coverage=[]
for status in sorted((out/'raw').glob('*.status.json')):
 s=json.loads(status.read_text()); stem=str(status).removesuffix('.status.json'); p=pathlib.Path(stem+'.jsonl'); p=p if p.exists() else pathlib.Path(str(p)+'.gz')
 records=[]
 if p.exists():
  with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as src:
   for line in src:
    try: records.append(json.loads(line))
    except json.JSONDecodeError: s['truncated_line']=True
 s['recorded_files']=[r['file'] for r in records if 'file' in r]
 s['checker_rejected']=records[0].get('checker_rejected') if records else None
 s['diagnostic_count']=len(records[0].get('diagnostics',[])) if records else None
 s['full_header']=bool(records and records[0].get('latent_mode')=='full')
 s['relation_trailer']=any(r.get('status')=='relation_report' for r in records)
 for r in records:
  for f in r.get('findings') or []:
   if 'SourceFileLike' not in f.get('reason',''):continue
   key=tuple(f.get(k,'') for k in ['kind','where','reason','text','phase','unit'])
   if key not in seen:seen[key]=dict(f,observed_in=[])
   if s['file'] not in seen[key]['observed_in']:seen[key]['observed_in'].append(s['file'])
 coverage.append(s)
(out/'findings.json').write_text(json.dumps(list(seen.values()),indent=2)+'\n')
(out/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
print(json.dumps({'attempted':len(coverage),'complete':sum(s['child_exit']==0 and s['relation_trailer'] for s in coverage),'SourceFileLike_findings':len(seen),'exit_counts':dict(collections.Counter(s['child_exit'] for s in coverage))}))
