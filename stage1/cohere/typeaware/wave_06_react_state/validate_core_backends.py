"""Compare prepared-HIR source and emitted JavaScript on independent Node."""
import pathlib,subprocess,json,time,argparse
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);a=p.parse_args();scratch=pathlib.Path(a.scratch).resolve();root=pathlib.Path(__file__).resolve().parents[4];records=[]
def run(name,cmd):
 start=time.monotonic()
 with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:r=subprocess.run([str(x) for x in cmd],cwd=root,stdout=out,stderr=err)
 records.append(dict(name=name,command=[str(x) for x in cmd],exit=r.returncode,seconds=time.monotonic()-start));(scratch/'backend-runs.json').write_text(json.dumps(records,indent=2)+'\n');assert r.returncode==0,(name,(scratch/(name+'.stderr')).read_text());return (scratch/(name+'.stdout')).read_bytes()
for row in json.loads((scratch/'batches.json').read_text())+[dict(name=n,entry=str(scratch/(n+'.a')),truth=str(scratch/(n+'-go.stdout'))) for n in ['compiler','repository']]:
 name=row['name'];entry=pathlib.Path(row['entry']);truth=pathlib.Path(row['truth']).read_bytes();assert run(name+'-source-node',['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs',entry])==truth;assert (scratch/(name+'-source-node.stderr')).read_bytes()==b''
 js=scratch/(name+'-emitted.mjs');js.write_bytes(run(name+'-js-build',[scratch/'adamic','js',entry]));assert run(name+'-emitted-node',['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs',js])==truth;assert (scratch/(name+'-emitted-node.stderr')).read_bytes()==b'';print(name,'source and emitted JavaScript match Go bytes',flush=True)
print('PASS prepared-HIR backend comparisons; native source lowering remains absent',flush=True)
