from pathlib import Path
import hashlib,json,subprocess,concurrent.futures
exec(Path(__file__).with_name('replay.py').read_text().split('pairs=[]')[0])
branches=subprocess.check_output(['git','for-each-ref','--format=%(refname:short)','refs/remotes/origin/compiler/fx[67]*'],text=True).splitlines()
seen={hashlib.sha256((root/r['source']).read_bytes()).hexdigest() for r in json.loads((evidence/'main-results.json').read_text())}
inventory=[];pairs=[]
def item(path):
 name=Path(path).stem; folder=Path(path).parent.name
 if folder=='fx6': return 134 if name in ('p35','p36','p37','p38','p38_number') else 136
 if folder=='fx7' and '/lower/testdata/' in path: return {'p04':147,'p06':147,'p08':149,'p59':149,'p75':148,'p77':148}[name]
 if folder=='callable_producers': return 133 if name=='p21' else 145 if name in ('p60','p61') else 135
 if folder=='fx7_tuple_recognition': return 138
 if folder=='fx7_wrong_aborts': return 146
 if folder=='scalar_union_views': return 144
 if folder=='tuple_object_views': return 137
 if folder=='unsupported_view': return 142
 if folder=='view_diagnostics': return 139
 if folder=='view_element_reads': return 132
 if 'callable_union_result' in name:return 154
 if 'snapshot-maybe' in name:return 143
 if '/interface-downcasts/fx7/' in path:return 151 if name.startswith('depth') else 140 if name in ('p09','p32','p33') else 152
 if path.endswith('p70.a'):return 137
 if 'p19-native' in name:return 132
 return None
for branch in branches:
 base=subprocess.check_output(['git','merge-base','origin/main',branch],text=True).strip()
 paths=subprocess.check_output(['git','diff','--name-only',base,branch],text=True).splitlines()
 for path in paths:
  if not path.endswith('.a') or '/review/' in path:continue
  number=item(path)
  if number is None: continue
  result=subprocess.run(['git','show',branch+':'+path],capture_output=True)
  if result.returncode:continue
  digest=hashlib.sha256(result.stdout).hexdigest()
  dest=evidence/'sources'/digest[:12]/Path(path).name
  inventory.append(dict(branch=branch,fork=base,path=path,item=number,sha256=digest,source=str(dest.relative_to(root))))
  if digest in seen:continue
  seen.add(digest);dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(result.stdout)
  pairs.append((number,dest))
(evidence/'source-inventory.json').write_text(json.dumps(inventory,indent=2)+'\n')
# Distinct source bytes get one replay, with every branch provenance retained above.
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool: rows=list(pool.map(probe,sorted(pairs)))
(evidence/'supplement-results.json').write_text(json.dumps(rows,indent=2)+'\n')

canonical={hashlib.sha256((root/x['source']).read_bytes()).hexdigest():x['source'] for x in json.loads((evidence/'main-results.json').read_text())+rows}
for entry in inventory: entry['source']=canonical[entry['sha256']]
(evidence/'source-inventory.json').write_text(json.dumps(inventory,indent=2)+'\n')
