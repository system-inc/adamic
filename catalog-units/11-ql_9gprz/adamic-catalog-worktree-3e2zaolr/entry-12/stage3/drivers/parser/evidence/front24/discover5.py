from pathlib import Path
import subprocess,json,re
out=Path('/workspace/scratch/parser-front24-discovery5');out.mkdir(); entry=Path('/workspace/scratch/parser-front24-discovery3-slice/parser-proof-main.a'); compiler='/workspace/scratch/parser-front24-adamic';rows=[]
for i in range(4):
 p=subprocess.run([compiler,'c',str(entry)],cwd='/tmp/parser-front7-scratch',stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 (out/f'{i+1:02}.stderr').write_bytes(p.stderr)
 if p.returncode==0:
  (out/'parser.c').write_bytes(p.stdout);rows.append({'step':i+1,'exit':0,'c_bytes':len(p.stdout)});break
 text=p.stderr.decode(); row={'step':i+1,'exit':p.returncode,'diagnostic':text};rows.append(row)
 m=re.search(r'adamic: (.*?):(\d+):(\d+):',text)
 if not m:break
 if i==10:break
 q=subprocess.run(['node','/workspace/scratch/parser-front24-stub3.cjs',*m.groups()],stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 row['stub_exit']=q.returncode
 if q.returncode:row['stub_error']=q.stderr.decode();break
 row['stub']=json.loads(q.stdout)
(out/'report.json').write_text(json.dumps({'warning':'All rows after first found behind cumulative source stubs. No native or Node identity claim. Validated slice unchanged.','rows':rows},indent=2)+'\n')
print(json.dumps(rows,indent=2))
