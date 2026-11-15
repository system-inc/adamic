from pathlib import Path
import os,sys,time,json
out=Path(__file__).resolve().parent;s=Path('/workspace/loose-slot-scratch/eleven');sys.path.insert(0,str(out.parent));from measure import run
while not(s/'type-count').exists():time.sleep(3)
for name in ['c04','u00-reconstructed','u00b-reconstructed']:
 src=out/'controls'/(name+'.ts');d=out/'evidence/controls'/name
 run([s/'head-census',src,str(d)+'-production.json'],str(d)+'-production',cwd=s/'head')
 run([s/'type-count',src,str(d)+'-type.json'],str(d)+'-type',cwd=s/'head')
 p=json.loads(Path(str(d)+'-production.json').read_text());t=json.loads(Path(str(d)+'-type.json').read_text());print(name,p,len(t['Events']),t['Catches'],flush=True)
 if name=='u00-reconstructed':assert p['kind']=='NotYet' and '#yeep0sx' in p['reason']and any(e['Catch']for e in t['Events'])
 else:assert p['kind']=='accepted'and not t['Events']
run([s/'type-count',s/'adapted-head/src/compiler',out/'evidence/type-whole.json'],out/'evidence/type-whole',cwd=s/'head')
print('whole type scan complete',flush=True)
