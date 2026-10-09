# Compare complete frozen step-18 censuses. Run from the repository root.
import json,pathlib,collections,hashlib,sys,re
before,after,out=sys.argv[1:]
prefix='/workspace/scratch/scout-optional-adapted/src/compiler/'
def read(p):return [json.loads(x) for x in pathlib.Path(p).read_text().splitlines()]
def findings(rows):return [dict((k,v.replace(prefix,'') if isinstance(v,str) else v) for k,v in f.items()) for r in rows for f in r.get('findings',[]) if f['kind'] in ('NotYet','Refused')]
def key(f):return tuple(f.get(k,'') for k in ('kind','where','reason','text'))
b,a=read(before),read(after)
assert len(b)==len(a)==80
assert b[0]==a[0]
assert {r['file'] for r in b if 'file'in r}=={r['file'] for r in a if 'file'in r}
bf,af=findings(b),findings(a)
reasons=set(r['reason'] for r in json.load(open('docs/step-18/inventory.json'))['rows'])
reasons.update(f['reason'] for f in bf+af if f['reason'].startswith(('an optional call','an optional method','observing an optional void','an optional chain')))
for source in ('internal/lower/optional_chain.go','internal/lower/optional_callable.go'):
 reasons.update(re.findall(r'l\.notYet\([^\n]*?, "([^"]+)"\)', pathlib.Path(source).read_text()))
reasons.add("a chain call without one represented callable signature")
rows=[]
for reason in sorted(reasons):
 bb={key(f):f for f in bf if f['reason']==reason};aa={key(f):f for f in af if f['reason']==reason}
 gone=[dict(base_root=f,barriers_at_same_position=[x for x in af if x['where']==f['where']],barriers_at_same_line=[x for x in af if x['where'].rsplit(':',1)[0]==f['where'].rsplit(':',1)[0]]) for k,f in bb.items() if k not in aa]
 rows.append(dict(reason=reason,before=len(bb),after=len(aa),disappeared=gone,new=[f for k,f in aa.items() if k not in bb]))
 print(len(bb),len(aa),len(gone),reason)
 for f in gone:print(' ',f['base_root']['where'],[(x['where'],x['reason'])for x in f['barriers_at_same_line']])
r=dict(base='4885cec5',before_sha256=hashlib.sha256(pathlib.Path(before).read_bytes()).hexdigest(),after_sha256=hashlib.sha256(pathlib.Path(after).read_bytes()).hexdigest(),files=79,checker_metadata_identical=True,measurement='checker-rejected; no usable IR; exact local roots only, never whole-program acceptance or hidden-byte retirement',rows=rows)
r['source_sha256']={p:hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest() for p in ('internal/lower/optional_chain.go','internal/lower/optional_callable.go','internal/lower/expression.go','internal/native/optional_calls.go','internal/native/taste.go','internal/native/emit_maps.go','internal/native/emit_arrays.go','internal/javascript/optional_calls.go')}
r['credited_roots']=[x['base_root'] for row in rows for x in row['disappeared'] if not x['barriers_at_same_line']]
r['reclassified_roots']=[x for row in rows for x in row['disappeared'] if x['barriers_at_same_line']]
pathlib.Path(out).write_text(json.dumps(r,indent=2)+'\n')
