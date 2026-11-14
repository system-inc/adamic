import os,sys,subprocess,json,hashlib,time
from pathlib import Path
base=Path('/tmp/destructure-price-scratch/adapted').resolve();out=Path('/workspace/scratch/destructure-price-f2989762/r2/census');out.mkdir(exist_ok=True)
sites=json.loads(Path('/tmp/destructure-price-scratch/ast-sites.json').read_text());files=sorted({s['file'] for s in sites},key=lambda name:(base/name).stat().st_size);summary=[]
for i,file in enumerate(files):
 slug=file.removeprefix('src/compiler/').removesuffix('.ts').replace('/','__');prefix=out/slug;raw=prefix.with_suffix('.jsonl')
 env={**os.environ,'LATENT_FULL':'1','LATENT_ASSERT_NO_OUTPUT':'1','LATENT_ONLY_FILE':str(base/file),'LATENT_TIMES':str(out/(slug+'-times.jsonl')),'GOMEMLIMIT':'8GiB','TMPDIR':'/tmp/destructure-price-scratch/tmp','GOTMPDIR':'/tmp/destructure-price-scratch/tmp','GOCACHE':'/tmp/destructure-price-scratch/cache/go'}
 print(f'START {i+1}/{len(files)} {file}',flush=True)
 cmd=[sys.executable,str(out.parent/'measure.py'),'--out',str(prefix),'--seconds','3600','--','/tmp/destructure-price-scratch/census',str(base/'src/compiler'),str(raw)]
 subprocess.run(cmd,env=env,check=False)
 meta=json.loads(prefix.with_suffix('.json').read_text());rows=[]
 if raw.exists():
  for line in raw.read_text().splitlines():
   try:rows.append(json.loads(line))
   except json.JSONDecodeError:pass
 selected=[r for r in rows if r.get('file')==str(base/file)]
 complete=meta['exit']==0 and bool(rows) and rows[-1].get('status')=='relation_report' and len(selected)==1 and rows[0].get('module_format')=='CommonJSBundle' and rows[0].get('latent_mode')=='full'
 findings=[f for row in rows for f in row.get('findings',[]) if f.get('kind')=='Refused' and f.get('reason')=='a method in object destructuring']
 own={f['where']:f for f in findings if f['where'].rsplit(':',2)[0]==str(base/file)}
 dependency={f['where']:f for f in findings if f['where'].rsplit(':',2)[0]!=str(base/file)}
 row={'file':file,'astSites':sum(s['file']==file for s in sites),'complete':complete,'selectedFileRecords':len(selected),'censusCount':len(own) if complete else None,'observedUniqueSites':len(own),'eventCount':len(findings),'sites':[{'where':k.removeprefix(str(base)+'/'),'phase':v.get('phase'),'unit':v.get('unit','').removeprefix(str(base)+'/')} for k,v in sorted(own.items())],'dependencySites':[k.removeprefix(str(base)+'/') for k in sorted(dependency)],'measurement':meta,'header':rows[0] if rows else None}
 summary.append(row);(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print('DONE '+json.dumps({k:row[k] for k in ['file','complete','astSites','censusCount','observedUniqueSites']}),flush=True)
print('ALL FILES FINISHED',flush=True)
