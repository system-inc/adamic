from pathlib import Path
import json,os,subprocess,time
repo=Path('/workspace/adamic');base=Path('/workspace/scratch/scanner-main-next-records');out=base/'historical-controls';out.mkdir()
env=dict(os.environ,SCANNER_TYPESCRIPT='/workspace/scratch/native3-cache/api/node_modules/typescript/lib/typescript.js',ADAMIC_NATIVE_SPLIT='0')
rows=[]
def run(command,stem):
 start=time.monotonic()
 with (out/(stem+'.stdout')).open('wb') as stdout,(out/(stem+'.stderr')).open('wb') as stderr:
  code=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; exec "$@"','tools',*map(str,command)],cwd=base/'compiler-tree',env=env,stdout=stdout,stderr=stderr).returncode
 return code,round(time.monotonic()-start,3)
for original in json.loads((repo/'stage3/drivers/scanner/evidence/land-area-next/witnesses/results.json').read_text()):
 name=original['name'];file=out/(name+'.a');file.write_text(original['source']);binary=out/(name+'-native')
 row={'name':name,'source':original['source']}
 row['node_exit'],row['node_wall_seconds']=run(['node',repo/'stage3/drivers/scanner/scratch-witness.cjs',file],name+'-node')
 row['build_exit'],row['build_wall_seconds']=run([base/'adamic','build',file,'-o',binary],name+'-build')
 if row['build_exit']==0 and row['node_exit']==0:
  row['native_exit'],row['native_wall_seconds']=run([binary],name+'-native')
  row['diff_exit'],_=run(['diff','-u',out/(name+'-node.stdout'),out/(name+'-native.stdout')],name+'-diff')
  content=(out/(name+'-native.stdout')).read_bytes()
  if content:
   mutant=out/(name+'-mutant.stdout');mutant.write_bytes(bytes([content[0]^1])+content[1:]);row['mutant_diff_exit'],_=run(['diff','-u',out/(name+'-node.stdout'),mutant],name+'-mutant-diff')
 for phase in ['node','build','native']:
  for stream in ['stdout','stderr']:
   p=out/(name+'-'+phase+'.'+stream)
   if p.exists():row[phase+'_'+stream]=p.read_text(errors='replace')
 row['status']='probe-admitted-and-matches-Node' if row.get('native_exit')==0 and row.get('diff_exit')==0 and row.get('mutant_diff_exit')==1 else 'probe-still-blocked' if row['node_exit']==0 and row['build_exit']!=0 else 'probe-failed'
 rows.append(row);(out/'results.json').write_text(json.dumps(rows,indent=2)+'\n');print(name,row['status'],flush=True)
assert len(rows)==15 and not any(r['status']=='probe-failed' for r in rows)
