from pathlib import Path
import subprocess,json,time
r=Path('/workspace/scratch/native3-combined-probes');name='16-new';source='const Debug = { fail: (message?: string): never => { throw new Error("failure"); } };\nconsole.log("ok");\n';f=r/(name+'.a');f.write_text(source);row={'name':name,'source':source,'qualification':'This reproduces the throwing Debug object placeholder signature; not a claim about original namespace lowering.'}
for kind,command in [('node',['node','--disable-warning=ExperimentalWarning','/workspace/scanner-native3-combined/oracle/node.mjs',str(f)]),('build',['/workspace/scratch/scanner-combined-adamic','build',str(f),'-o',str(r/(name+'.native'))])]:
 start=time.monotonic()
 with (r/(name+'.'+kind+'.stdout')).open('wb') as out,(r/(name+'.'+kind+'.stderr')).open('wb') as err:row[kind+'_exit']=subprocess.run(command,cwd='/workspace/scanner-native3-combined',stdout=out,stderr=err).returncode
 row[kind+'_seconds']=time.monotonic()-start
 row[kind+'_stdout']=(r/(name+'.'+kind+'.stdout')).read_text();row[kind+'_stderr']=(r/(name+'.'+kind+'.stderr')).read_text()
assert row['node_exit']==0 and row['build_exit']==1 and "stage 0 can't lower a function value with an optional parameter yet" in row['build_stderr'],row
rows=json.loads((r/'results.json').read_text());rows.append(row);(r/'results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row))
