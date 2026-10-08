"""Compile isolated original member probes, retaining exact diagnostics."""
from pathlib import Path
import concurrent.futures, json, subprocess
root=Path(__file__).resolve().parents[4]
unit=Path(__file__).resolve().parent
manifest=json.loads((unit/'manifest.json').read_text())
logs=Path('/tmp/lane5-c-probes'); logs.mkdir(exist_ok=True)
def probe(row):
    if row['status']!='generated': return row
    path=unit/'families'/('rank-'+str(row['rank']))/'good.a'
    result=[]
    for backend in ['c','js']:
        destination=logs/(str(row['rank'])+'-'+backend+'.log')
        with destination.open('w') as log:
            p=subprocess.run(['/tmp/lane5-c-adamic',backend,str(path)],cwd=root,stdout=log,stderr=subprocess.STDOUT)
        result.append({'backend':backend,'exit':p.returncode,'log':str(destination)})
    row['probes']=result
    row['status']='ready for oracle' if all(p['exit']==0 for p in result) else 'compiler read refusal'
    print(row['rank'],row['status'],flush=True)
    return row
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
    result=list(pool.map(probe,manifest))
(unit/'probe-results.json').write_text(json.dumps(result,indent=2)+'\n')
