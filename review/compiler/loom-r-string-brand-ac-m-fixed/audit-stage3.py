import concurrent.futures,json,re,subprocess
from pathlib import Path
root=Path.cwd();out=Path(__file__).resolve().parent;decoder=json.JSONDecoder()
items=[]
for manifest in sorted(Path('stage3/fixtures').glob('*/status.json')):
 for record in json.loads(manifest.read_text()):items.append((manifest,record))
def observe(item):
 manifest,record=item;path=manifest.parent/record['file'];r=subprocess.run(['/tmp/string-brand-fixed-adamic','c',str(path)],capture_output=True,text=True,timeout=30)
 errors=r.stderr.removeprefix('adamic: ').replace(str(root)+'/', '').rstrip('\n')
 outcome='Compiles' if r.returncode==0 else 'NotYet' if "can't lower" in errors and ' yet' in errors else 'Refused' if 'Adamic 0.1 refuses' in errors else 'Checker' if ' error TS' in errors else 'Unknown'
 return {'manifest':str(manifest),'file':record['file'],'before':record['stage0'],'actual':{'outcome':outcome,'what':errors},'exit':r.returncode}
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=list(pool.map(observe,items))
changes=[];pending=[]
for row in rows:
 if row['before']['outcome']=='CheckedStop':continue
 if row['before'].get('outcome')==row['actual']['outcome'] and row['before'].get('what','')==row['actual']['what']:continue
 if row['actual']['outcome']=='Compiles':pending.append(row);continue
 assert row['before']['outcome'] not in ('Compiles','CheckedStop'),row
 assert row['actual']['outcome'] in ('Refused','NotYet','Checker'),row
 changes.append(row)
# Replace only stage0's JSON span. Every other field and byte stays unchanged.
for name in sorted({r['manifest'] for r in changes}):
 p=Path(name);text=p.read_text();original=text;targets={r['file']:r for r in changes if r['manifest']==name};cursor=0;replacements=[]
 while cursor<len(text):
  if text[cursor] not in '{':cursor+=1;continue
  record,end=decoder.raw_decode(text,cursor)
  if isinstance(record,dict) and record.get('file') in targets:
   raw=text[cursor:end];match=re.search(r'"stage0"\s*:\s*',raw);assert match
   _,value_end=decoder.raw_decode(raw,match.end());replacement=json.dumps(targets[record['file']]['actual'],separators=(',',':'))
   replacements.append((cursor+match.end(),cursor+value_end,replacement))
  cursor=end
 for start,end,replacement in reversed(replacements):text=text[:start]+replacement+text[end:]
 assert len(replacements)==len(targets)
 p.write_text(text)
(out/'stage3-audit.json').write_text(json.dumps({'observations':rows,'diagnostic_updates':changes,'pending_native_proof':pending},indent=2))
print('Audited',len(rows),'records; diagnostic updates',[(r['manifest'],r['file']) for r in changes],'new admissions needing Node/native proof',[(r['manifest'],r['file']) for r in pending],flush=True)
