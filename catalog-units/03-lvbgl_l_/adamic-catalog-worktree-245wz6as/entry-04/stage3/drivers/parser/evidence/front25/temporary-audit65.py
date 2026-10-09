from pathlib import Path
import subprocess,shutil,json,re
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/parser-front25-slice');out=Path('/workspace/scratch/parser-front25-temporary-audit');out.mkdir(exist_ok=True);rows=[]
for number in range(65,66):
 adapter=next((repo/'stage3/adapt').glob(str(number)+'-*'))
 text=(adapter/'adapt.cjs').read_text()
 def string(name):return json.loads(re.search(r'(?:const |, )'+name+r' = ("(?:[^"\\]|\\.)*")',text)[1])
 file=string('file');before=string('before');after=string('after')
 target=out/str(number);shutil.copytree(root,target)
 p=target/file;s=p.read_text();assert s.count(after)==1,(number,after)
 s=s.replace(after,before)
 if number==65:s=s.replace('declare const process: (NodeJS.Process & { browser?: unknown }) | undefined;','declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;')
 p.write_text(s)
 with (out/f'{number}.stdout').open('wb') as stdout,(out/f'{number}.stderr').open('wb') as stderr:
  r=subprocess.run(['/workspace/scratch/parser-front25-arguments-adamic','c',str(target/'parser-proof-main.a')],cwd='/tmp/parser-front7-scratch',stdout=stdout,stderr=stderr)
 rows.append({'adaptation':number,'removed_site':before,'file':file,'exit':r.returncode,'stderr':(out/f'{number}.stderr').read_text()})
 (out/'report65.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
