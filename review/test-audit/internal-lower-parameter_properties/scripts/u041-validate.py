from pathlib import Path
import subprocess,json,time
r=Path('review/test-audit/internal-lower-parameter_properties'); temp=Path('/tmp/u041/validation');temp.mkdir(parents=True,exist_ok=True)
menu=json.loads((r/'menu.json').read_text()); files={m['file'] for m in menu}; bases={f:subprocess.check_output(['git','show','origin/main:'+f],text=True) for f in files}; replacements={}; results={}
for file,b in bases.items():
 dest=temp/(file.replace('/','_')+'.original');dest.write_text(b);replacements[str(Path(file).resolve())]=str(dest)
for m in menu:
 mid=m['id'];candidate=temp/(mid+'.go');candidate.write_text(bases[m['file']].replace(m['old'],m['new']))
 overlay=dict(replacements);overlay[str(Path(m['file']).resolve())]=str(candidate)
 path=temp/(mid+'.json');path.write_text(json.dumps({'Replace':overlay}))
 start=time.monotonic()
 with (r/(mid+'-vet.log')).open('w') as f:
  diffcheck=subprocess.run(['git','apply','--check','--cached',str(r/(mid+'.diff'))],stdout=f,stderr=subprocess.STDOUT)
  package='./'+str(Path(m['file']).parent)+'/'
  c=subprocess.run(['timeout','90','go','vet','-overlay='+str(path),package],stdout=f,stderr=subprocess.STDOUT)
 results[mid]={'apply_exit':diffcheck.returncode,'vet_exit':c.returncode,'wall_seconds':round(time.monotonic()-start,3),'command':'go vet -overlay='+str(path)+' '+package}
 (r/'standalone-validation.json').write_text(json.dumps(results,indent=2)+'\n')
 assert diffcheck.returncode==0 and c.returncode==0,(mid,results[mid])
