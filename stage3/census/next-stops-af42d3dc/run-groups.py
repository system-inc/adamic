import json,os,subprocess,sys
from pathlib import Path
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;tree=Path('/workspace/after-wave-adapted')
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
(out/'roots.json').write_text(json.dumps(list(map(str,roots)),indent=2)+'\n')
for index in range(0,len(roots),20):
 group=index//20+1; raw=out/f'group-{group}.jsonl';rowsfile=out/f'group-{group}-rows.json'
 rows=json.loads(rowsfile.read_text()) if rowsfile.exists() else []
 assert [r['roots'][0] for r in rows]==list(map(str,roots[index:index+len(rows)]))
 if len(rows)==len(roots[index:index+20]):continue
 assert not raw.exists() or len(raw.read_bytes().splitlines())==2*len(rows)
 with raw.open('ab') as stream:
  for root in roots[index+len(rows):index+20]:
   temporary=scratch/'root.jsonl';temporary.unlink(missing_ok=True)
   name=f'root-{roots.index(root)+1:02}'
   status=run(name,[str(scratch/'census'),str(root),str(temporary)],watch=[temporary])
   (out/(name+'-status.json')).write_bytes((scratch/(name+'-status.json')).read_bytes())
   if status['exit_code']!=0:sys.exit(1)
   data=temporary.read_bytes();observations=[json.loads(l) for l in data.splitlines()]
   assert len(observations)==2
   assert {k:v for k,v in observations[0].items() if k!='seconds'}=={k:v for k,v in observations[1].items() if k!='seconds'}
   stream.write(data);stream.flush();rows.append(observations[0]);rowsfile.write_text(json.dumps(rows,indent=2)+'\n')
   print(f'group {group}: {len(rows)}/{len(roots[index:index+20])} {root.name} {rows[-1]["kind"]}',flush=True)
 p=subprocess.run([sys.executable,str(out/'checkpoint.py'),f'group {group}'],cwd=repo)
 print('checkpoint child_exit='+str(p.returncode),flush=True);p.check_returncode()
