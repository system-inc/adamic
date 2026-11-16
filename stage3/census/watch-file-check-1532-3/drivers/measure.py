import os,pathlib,subprocess as p,json,time,sys
s=pathlib.Path('/workspace/scratch/watch1532-recut');r=s/'work';name=sys.argv[1];tree=s/sys.argv[2];env=dict(os.environ,NODE_PATH=str(r/'stage3/api/node_modules'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'gotmp'),GOCACHE=str(s/'go-cache'),LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler/sys.ts'),LATENT_TIMES=str(s/'logs'/(name+'-times.jsonl')),GOMEMLIMIT='5GiB',GOMAXPROCS='4')
while not (tree/'patch-set.md').exists():time.sleep(2)
rows=[]
def run(label,args):
 start=time.monotonic()
 with (s/'logs'/(name+'-'+label+'.log')).open('w') as f:z=p.run(list(map(str,args)),env=env,cwd=r,stdout=f,stderr=p.STDOUT)
 row=dict(name=name+'-'+label,command=list(map(str,args)),exit=z.returncode,seconds=round(time.monotonic()-start,3));rows.append(row);(s/'logs'/(name+'-results.json')).write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
run('census',[s/'latent-census',tree/'src/compiler',s/'logs'/(name+'-census.jsonl')])
run('stock',['node',s/'stock.cjs',tree])
roots=sorted(p for p in (tree/'src/compiler').rglob('*') if p.suffix in ['.ts','.a']);assert len(roots)==79
run('types',[s/'adamic','types',*roots])
