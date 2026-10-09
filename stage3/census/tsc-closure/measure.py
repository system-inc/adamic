#!/usr/bin/env python3
"""Summarize an explicit source closure using pinned latent and hidden semantics."""
import argparse,collections,functools,hashlib,importlib.util,json
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('manifest',type=Path);p.add_argument('latent',type=Path);p.add_argument('full',type=Path);p.add_argument('stock',type=Path);p.add_argument('hidden_tool',type=Path);p.add_argument('output',type=Path);a=p.parse_args()
spec=importlib.util.spec_from_file_location('closure_hidden',a.hidden_tool);hidden=importlib.util.module_from_spec(spec);spec.loader.exec_module(hidden)
m=json.loads(a.manifest.read_text());root=Path(m['root']);expected={f['file'] for f in m['files']};stock=hidden.read_json(a.stock)
def coverage(rows):
 observed=[Path(r['file']).relative_to(root).as_posix() for r in rows[1:]]
 if set(observed)!=expected or len(observed)!=len(expected):raise ValueError('closure coverage mismatch: '+str(sorted(expected-set(observed))))
 if set(stock)!=expected:raise ValueError('stock closure coverage mismatch')
 for f in m['files']:
  if stock[f['file']]['bytes']!=f['bytes'] or stock[f['file']]['sha256']!=f['sha256']:raise ValueError('stock source identity mismatch')
 if set(rows[0]['root_files'])!={str(root/f) for f in expected}:raise ValueError('root manifest mismatch')
def census(path):
 rows=hidden.read_rows(path);coverage(rows)
 keys={(f['kind'],f['where'],f['reason'],f['text']) for r in rows[1:] for f in r['findings'] if f['kind']!='Boundary'}
 @functools.lru_cache(None)
 def location(site):
  path=Path(site[1].rsplit(':',2)[0]);return path.relative_to(root).as_posix() if path.is_relative_to(root) else None
 def group(names):
  sites=[s for s in keys if names is None or location(s) in names]
  d=[d for d in rows[0]['diagnostic_sites'] if d['file'] and Path(d['file']).is_relative_to(root) and (names is None or Path(d['file']).relative_to(root).as_posix() in names)]
  return {'files':len(expected) if names is None else len(names-{None}),'counts':dict(sorted(collections.Counter(s[0] for s in sites).items())),'per_reason':dict(sorted(collections.Counter(s[0]+': '+s[2] for s in sites).items())),'checker_diagnostics':len(d),'checker_by_code':dict(sorted(collections.Counter(x['text'].split('error ')[1].split(':')[0] for x in d).items()))}
 inside={f for f in expected if f.startswith('src/compiler/')};outside=expected-inside
 return {'mode':rows[0]['latent_mode'],'checker_rejected':rows[0]['checker_rejected'],'all':group(None),'compiler':group(inside),'outside_compiler':group(outside),'unattributed':group({None}), 'per_file':{f:group({f}) for f in sorted(expected)}}
first=census(a.latent);full=census(a.full);rows=hidden.read_rows(a.full)
result=hidden.calculate(rows,stock,root)
def totals(names):
 fields=['bytes','hidden_bytes','blocked_union_bytes','independently_examined_bytes'];d={key:sum(result['files'][name][key] for name in names) for key in fields};d['files']=len(names);d['hidden_share']=d['hidden_bytes']/d['bytes'] if d['bytes'] else 0;return d
inside={f for f in expected if f.startswith('src/compiler/')};outside=expected-inside
result['groups']={'all':totals(expected),'compiler':totals(inside),'outside_compiler':totals(outside)}
result['provenance']={'compiler_root':str(root),'source_root':str(root),'definition':'Only reached source files, including all outside src/compiler; UTF-8 union of blocked ranges minus independent coverage. Non-imported JSON and other assets excluded.'}
for region in result['largest_regions']:
 data=(root/region['file']).read_bytes();region['start_line']=data[:region['start']].count(b'\n')+1;region['end_line']=data[:region['end']-1].count(b'\n')+1
report={'closure':m,'latent':first,'full':full,'hidden':result};a.output.write_text(json.dumps(report,indent=2)+'\n');print(json.dumps({'latent':first['all'],'outside':first['outside_compiler'],'hidden':result['groups']},indent=2))
