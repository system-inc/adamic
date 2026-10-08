import gzip, hashlib, importlib.util, json, sys
from pathlib import Path
# Arguments: frozen census worktree, hash-pinned adapted compiler source.
frozen=Path(sys.argv[1])/'stage3/census/hidden'
report=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('hidden', frozen/'hidden.py')
h=importlib.util.module_from_spec(spec);spec.loader.exec_module(h)
stock=h.read_json(frozen/'evidence/stock.json.gz')
source=Path(sys.argv[2])
for name,meta in stock.items():
 data=(source/name).read_bytes()
 assert len(data)==meta['bytes'] and hashlib.sha256(data).hexdigest()==meta['sha256'],name
result={'region':{'file':'utilities.ts','start':48518,'end':64484,'sha256':stock['utilities.ts']['sha256']},'verified_source_files':len(stock),'measurements':{}}
for name,path,commit in [
 ('original',frozen/'evidence/full.jsonl.gz','ed6e29751ee47d86fad450cd1674139883bc0f70'),
 ('area',report/'evidence/base-census.jsonl.gz','dcdbb9098f77f30ad41790c56df1bd63ad462b63'),
 ('existing_fix_archived',report/'evidence/existing-fix-archived.jsonl.gz','a64f77 (archived pre-merge measurement)')]:
 rows=h.read_rows(path);root=Path(rows[1]['file'].split('/src/compiler/')[0])/'src/compiler'
 measured=h.calculate(rows,stock,root)
 spans=[(max(a,48518),min(b,64484)) for a,b in measured['files']['utilities.ts']['hidden_ranges'] if a<64484 and b>48518]
 boundaries=[];any_boundaries=set();any_callees={}
 for row in rows[1:]:
  for f in row['findings']:
   if f['kind']!='Boundary':continue
   if 'a function returning any' in f['text'] or 'a generic function whose resolved return type is any' in f['text']:
    any_boundaries.add((row['file'],f['start'],f['end']))
    diagnostic=f['text'].split(': stage 0',1)[0];any_callees[diagnostic]=any_callees.get(diagnostic,0)+1
   if row['file'].endswith('/utilities.ts') and f['start']<=48518<f['end']:
    boundaries.append({key:f[key] for key in ['unit','where','start','end','text']})
 result['measurements'][name]={'compiler':commit,'ledger_sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'hidden_intersection':spans,'hidden_bytes':h.size(spans),'head_boundaries':boundaries,'any_boundary_count':len(any_boundaries),'any_callees':any_callees}
result['measurements']['existing_fix']=dict(result['measurements']['existing_fix_archived'])
result['measurements']['existing_fix']['compiler']='32b7eed56c293e4c1cec049254db54ec0b38f0c9'
result['measurements']['existing_fix']['basis']='Exact-head replay confirms archived geometry; fresh full corpus run incomplete, no corpus count claimed.'
result['measurements']['existing_fix'].pop('any_boundary_count')
result['measurements']['existing_fix'].pop('any_callees')
result['measurements']['existing_fix'].pop('ledger_sha256')
replay=json.loads(report/'evidence/next-replay.json'.read_text())
boundary=[f for f in replay[0]['findings'] if f['kind']=='Boundary']
assert len(boundary)==1 and (boundary[0]['start'],boundary[0]['end'])==(48518,64484)
result['measurements']['existing_fix']['head_boundaries']=boundary
result['revealed_bytes_from_original']=result['measurements']['original']['hidden_bytes']-result['measurements']['existing_fix']['hidden_bytes']
result['revealed_bytes_from_area']=result['measurements']['area']['hidden_bytes']-result['measurements']['existing_fix']['hidden_bytes']
report/'RESULT.json'.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
