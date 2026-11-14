from pathlib import Path
import subprocess,os,sys,json
base=Path(sys.argv[1])/'src';tag=sys.argv[2];out=Path(__file__).parent
files=['compiler/sys.ts','compiler/watchPublic.ts','compiler/tsbuildPublic.ts','server/utilities.ts','server/session.ts','server/typingInstallerAdapter.ts'] if tag=='1' else ['compiler/sys.ts']
for file in files:
 dest=out/('sources.'+tag)/file;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes((base/file).read_bytes())
 name=('server-' if file.startswith('server/') else '')+Path(file).stem
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(base/file))
 root=base/file if file.startswith('server/') else base/'compiler'
 stream=out/('census.'+tag+'.'+name+'.jsonl')
 with (out/('census.'+tag+'.'+name+'.log')).open('w') as log:
  r=subprocess.run(['/workspace/timer-call-queue-scratch/census',str(root),str(stream)],env=env,stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==0,(file,r.returncode)
 rows=[json.loads(x)for x in stream.read_text().splitlines()]
 assert rows[-1]['status']=='relation_report' and any(r.get('file','').endswith('/src/'+file)for r in rows),file
 print(file,r.returncode,flush=True)
