from pathlib import Path
import subprocess,json,re
root=Path.cwd();ev=root/'review/compiler/loom-implied-target/followup'
source=(ev.parent/'objects-sequential.a').read_text()
rows=[]
for n in [512,1024,2048,4096]:
 text=re.sub(r'(const items = \[).*?(\n    \];)',lambda m:m[1]+'\n'+''.join(f'    {{ value: {i}, done: false }},\n' for i in range(n))+m[2],source,flags=re.S)
 assert text!=source or n==2048
 # Verify generated literal size rather than silently measuring an unchanged source.
 assert len(re.findall(r'\{ value: \d+, done: false \}',text))==n
 p=ev/f'objects-{n}.a';p.write_text(text)
 for label,binary in [('before','/tmp/implied-profile'),('after','/tmp/implied-profile-after')]:
  result=subprocess.run([binary,str(p),f'/tmp/implied-{label}-{n}.cpu'],capture_output=True,text=True,timeout=90)
  assert result.returncode==0,result.stderr
  row=json.loads(result.stdout);row.update(elements=n,state=label);rows.append(row)
  (ev/'sizes.json').write_text(json.dumps(rows,indent=2)+'\n')
  print(n,label,row['lower_seconds'],flush=True)
