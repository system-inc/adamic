import importlib.util,json,gzip,hashlib
from pathlib import Path
spec=importlib.util.spec_from_file_location('hidden','/tmp/hidden-11-pin/stage3/census/hidden/hidden.py');h=importlib.util.module_from_spec(spec);spec.loader.exec_module(h)
root=Path('/tmp/hidden-adapted/src/compiler');pin=Path('/tmp/hidden-11-pin/stage3/census/hidden');stock=h.read_json(pin/'evidence/stock.json.gz');rows=h.read_rows(Path('/tmp/hidden-11-full.jsonl'));assert len(rows)==80,len(rows)
result=h.calculate(rows,stock,root)
old=h.read_json(pin/'RESULT.json')
start,end=2778143,2786286
intersection=lambda ranges:[[max(a,start),min(b,end)] for a,b in ranges if a<end and b>start]
oldranges=intersection(old['files']['checker.ts']['hidden_ranges']);newranges=intersection(result['files']['checker.ts']['hidden_ranges'])
oldbytes=h.size(oldranges);newbytes=h.size(newranges)
# Independent per-byte assertion for the exact assigned region.
mask=bytearray(end-start)
for row in rows[1:]:
 for f in row['findings']:
  if f['kind']=='Boundary' and f['where'].rsplit(':',2)[0]==str(root/'checker.ts'):
   a,b=max(start,f['start']),min(end,f['end'])
   if a<b:mask[a-start:b-start]=b'\x01'*(b-a)
 skips=[]
checker=next(row for row in rows if row.get('file')==str(root/'checker.ts'))
for unit in checker['units']:
 if unit['status'] in ('split_checker_body','skipped_checker_body'):skips.append((unit['body_start'],unit['body_end']))
for a,b in skips:
 a,b=max(start,a),min(end,b)
 if a<b:mask[a-start:b-start]=b'\x01'*(b-a)
for unit in checker['units']:
 if unit['status'] not in ('attempted','panic'):continue
 external=next(x for x in stock['checker.ts']['units'] if x['where']==h.local_where(unit['where'],root))
 a,b=max(start,external['start']),min(end,external['end'])
 if a>=b:continue
 expose=bytearray(b'\x01'*(b-a))
 cuts=[(f['start'],f['end']) for r in rows[1:] for f in r['findings'] if f['kind']=='Boundary' and f['unit']==unit['where'] and f['where'].rsplit(':',2)[0]==str(root/'checker.ts')]
 cuts += [(x,y) for x,y in skips if external['start']<=x and y<=external['end']]
 for x,y in cuts:
  x,y=max(a,x),min(b,y)
  if x<y:expose[x-a:y-a]=b'\x00'*(y-x)
 for i,v in enumerate(expose,a-start):
  if v:mask[i]=0
assert sum(mask)==newbytes,(sum(mask),newbytes)
heads=sorted([f for row in rows[1:] for f in row['findings'] if f['kind']=='Boundary' and f['where'].rsplit(':',2)[0]==str(root/'checker.ts') and f['start']<=start<f['end']],key=lambda f:f['end']-f['start'])
summary={'region':[start,end],'old_hidden_intersection':oldranges,'new_hidden_intersection':newranges,'old_hidden_bytes':oldbytes,'new_hidden_bytes':newbytes,'revealed_bytes':oldbytes-newbytes,'next_boundary':heads[0] if heads else None,'independent_region_byte_mask':'pass','production_compiler_unchanged_from_base':'dcdbb9098f77f30ad41790c56df1bd63ad462b63','source_files':len(stock),'source_hash_mismatches':[n for n,v in stock.items() if hashlib.sha256((root/n).read_bytes()).hexdigest()!=v['sha256']]}
assert not summary['source_hash_mismatches']
Path('review/hidden-11/evidence/region.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
