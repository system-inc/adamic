from pathlib import Path
import json
import subprocess
root=Path.cwd()
out=root/'review/compiler-generators-main/count-comparison'
out.mkdir(exist_ok=True)
rows=[]
for label,relative in [('regexp','internal/oracle/testdata/regexp.a'),('regexp-tree','internal/fresh/testdata/regexp_tree.ts')]:
 sources={}
 counts={}
 for name,binary in [('main','/tmp/generators-baseline-adamic'),('generators','/tmp/generators-current-adamic')]:
  c=subprocess.run([binary,'c',str(root/relative)],cwd=root,capture_output=True,check=True)
  (out/(label+'-'+name+'.c.txt')).write_bytes(c.stdout)
  sources[name]=c.stdout
  target='/tmp/generators-count-'+label+'-'+name
  with (out/(label+'-'+name+'-build.log')).open('w') as log:
   subprocess.run([binary,'build',str(root/relative),'-o',target,'--count'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  with (out/(label+'-'+name+'-run.log')).open('w') as log:
   subprocess.run([target],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  counts[name]=(out/(label+'-'+name+'-run.log')).read_text().splitlines()[-1]
 rows.append({'fixture':relative,'C_identical':sources['main']==sources['generators'],'counts':counts})
(out/'RESULT.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
