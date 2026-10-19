import concurrent.futures,difflib,hashlib,json,os,subprocess
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/chain-slice-6';patches=out/'counts-diffs';patches.mkdir(exist_ok=True)
def table(text):
 return {line.split('|')[1].strip():line for line in text.splitlines() if line.startswith('| internal/')}
old=table(subprocess.check_output(['git','show','origin/main:internal/oracle/counts.md'],text=True));new=table((root/'internal/oracle/counts.md').read_text())
paths=[p for p in new if old.get(p)!=new[p]]
def check(path):
 row={'path':path,'before':old.get(path),'after':new[path],'member':'exceptions-21-main','source':'205586a0 + 71972e18'}
 if 'inherited_fields_guard' in path:row.update(member='inherit-guards',source='16a0b626')
 row["reason"]="new fixture registration" if path not in old else "catchable throw edges and owned unwind/finally slots change generated C"
 patch=patches/(path.replace("/","__")+".patch")
 if patch.exists():
  row["diff"]=str(patch.relative_to(root));return row
 outputs=[]
 for name,binary in [('main','/tmp/chain-slice-6-main'),('slice','/tmp/chain-slice-6-current')]:
  result=subprocess.run([binary,'c',path],capture_output=True,timeout=20,env=dict(os.environ,GOMAXPROCS='1'))
  row[name]={'exit':result.returncode,'sha256':hashlib.sha256(result.stdout).hexdigest(),'diagnostic':result.stderr.decode()[:500]};outputs.append(result.stdout.decode())
 assert row['slice']['exit']==0,path
 row['reason']='new fixture registration' if path not in old else 'catchable throw edges and owned unwind/finally slots change generated C'
 if row['main']['exit']==0:
  assert path not in old or outputs[0]!=outputs[1],path
  diff=''.join(difflib.unified_diff(outputs[0].splitlines(True),outputs[1].splitlines(True),fromfile='main/'+path,tofile='slice/'+path))
  patch=patches/(path.replace('/','__')+'.patch');patch.write_text(diff);row['diff']=str(patch.relative_to(root))
 return row
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=list(pool.map(check,paths))
removed=[p for p in old if p not in new];assert not removed,removed
(out/'counts-attribution.json').write_text(json.dumps({'changed_rows':len(rows),'removed_rows':removed,'rows':rows},indent=2)+'\n')
print('attributed',len(rows),'rows; no removed rows')
