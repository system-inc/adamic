import subprocess, pathlib, concurrent.futures, json, argparse
p=argparse.ArgumentParser();p.add_argument('destination');p.add_argument('--repos',default=str(pathlib.Path(__file__).parent/'testdata/public-repos.tsv'));a=p.parse_args()
root=pathlib.Path(a.destination).resolve();root.mkdir(parents=True,exist_ok=True)
def fetch(row):
 repo,sha=row.split();name=repo.replace('/','__')+'-'+sha[:8];path=root/name;path.mkdir(exist_ok=True)
 with (root/(name+'.log')).open('w') as log:
  for command in [['git','init','-q',str(path)],['git','-C',str(path),'fetch','--depth=1','https://github.com/'+repo+'.git',sha],['git','-C',str(path),'checkout','--detach','FETCH_HEAD']]:
   result=subprocess.run(command,stdout=log,stderr=log)
   if result.returncode:return dict(repo=repo,requested_sha=sha,status='failed',log=str(root/(name+'.log')))
 actual=subprocess.check_output(['git','-C',str(path),'rev-parse','HEAD'],text=True).strip()
 return dict(repo=repo,requested_sha=sha,sha=actual,path=str(path),status='fetched')
rows=pathlib.Path(a.repos).read_text().splitlines()
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 result=list(pool.map(fetch,rows))
(root/'fetch.json').write_text(json.dumps(result,indent=2)+'\n')
