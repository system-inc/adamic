"""Verify the complete fixed denominator and report Go identity by import capability."""
import argparse,collections,gzip,hashlib,json
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('census_directory',type=Path);a=p.parse_args()
out=Path(__file__).resolve().parent
key=lambda r:(r['repo'],r['pin'],r['path'])
baseline=json.loads(gzip.open(out/'side-effect.json.gz','rt').read()) if (out/'side-effect.json.gz').exists() else json.loads(gzip.open(out/'after.json.gz','rt').read())
base={key(r):r for r in baseline['files']}
go={key(r):r for r in map(json.loads,gzip.open(out/'go.jsonl.gz','rt'))}
assert len(base)==len(go)==152660
report={'port_source_sha256':{str(x.relative_to(out.parents[3])):hashlib.sha256(x.read_bytes()).hexdigest() for x in sorted(out.parent.glob('*.ts'))},'denominator':152660,'definition':'complete-file output SHA-256 equals pinned Go output; all input paths and source SHA-256 must equal the side-effect census','capability_ablations':['default','namespace','named'],'stages':{}}
previous=base
for label,filename in [('merged','merged'),('default','default'),('namespace','namespace'),('named','named'),('type','clauses'),('attributes','attributes')]:
 data=json.loads((a.census_directory/(filename+'.json')).read_text());files={key(r):r for r in data['files']}
 assert files.keys()==base.keys() and len(data['files'])==152660
 assert dict(collections.Counter(r['outcome'] for r in files.values()))==data['counts']
 assert all(r['sha256']==base[k]['sha256'] for k,r in files.items())
 identical=[];differences=[];conflicts=[];changes=[]
 for k,r in files.items():
  if (r['outcome'],r.get('answer'))!=(previous[k]['outcome'],previous[k].get('answer')):changes.append(r)
  if r['outcome']!='accepted':continue
  g=go[k]
  if 'error' in g:conflicts.append(dict(**r,go_error=g['error']))
  elif hashlib.sha256(r['answer'].encode()).hexdigest()==g['sha256']:identical.append(k)
  else:differences.append(r)
 report['stages'][label]=dict(counts=data['counts'],byte_identical=len(identical),output_differences=differences,acceptance_conflicts=conflicts,changes_from_previous=changes)
 previous=files
 print(label,len(identical),len(differences),len(conflicts))
# Retain the original side-effect census and the complete final census. The small
# intermediate changes plus final/base file hashes retain every stage outcome.
if not (out/'side-effect.json.gz').exists():(out/'side-effect.json.gz').write_bytes((out/'after.json.gz').read_bytes())
(out/'after.json.gz').write_bytes(gzip.compress(json.dumps(data).encode(),mtime=0))
(out/'import-stages.json').write_text(json.dumps(report,indent=2)+'\n')
summary=json.loads((out/'summary.json').read_text());summary['side_effect_after']=report['stages']['merged']['counts'];summary['side_effect_byte_identical']=report['stages']['merged']['byte_identical'];summary['after']=data['counts'];summary['after_byte_identical']=report['stages']['attributes']['byte_identical'];summary['after_mismatches']=differences;summary['after_accepted_go_refused']=conflicts
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
