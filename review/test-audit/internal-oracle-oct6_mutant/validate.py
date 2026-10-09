import json,pathlib,subprocess,shutil,time
r=pathlib.Path('review/test-audit/internal-oracle-oct6_mutant');menu=json.loads((r/'menu.json').read_text());s=json.loads((r/'scope.json').read_text());s['witness_rows']=[s['rows'][i] for i in [0,1,2,4,6,13,14]];(r/'scope.json').write_text(json.dumps(s,indent=2))
copy=pathlib.Path('/tmp/u066/runtime');shutil.copytree('internal/native/runtime',copy,dirs_exist_ok=True);results=[]
for m in menu:
 f=m['file'];id=m['id'];old=pathlib.Path(f).read_text();new=old.replace(m['before'],m['after']);start=time.monotonic();subprocess.run(['git','apply','--check',str(r/(id+'.diff'))],check=True)
 if f.endswith('.c'):
  p=copy/pathlib.Path(f).name;p.write_text(new);cmd=['clang','-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I',str(copy),'-c',str(p),'-o','/tmp/u066/check.o']
 else:
  p=pathlib.Path('/tmp/u066/'+id+'.go');p.write_text(new);ov=pathlib.Path('/tmp/u066/'+id+'-overlay.json');ov.write_text(json.dumps({'Replace':{str(pathlib.Path(f).resolve()):str(p)}}));cmd=['go','vet','-overlay',str(ov),'./internal/lower/' if '/lower/' in f else './internal/oracle/']
 with open('/tmp/u066/validate-'+id+'.log','w') as out:q=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
 if f.endswith('.c'):p.write_text(old)
 results.append(dict(id=id,returncode=q.returncode,seconds=time.monotonic()-start,command=' '.join(cmd)));print(id,q.returncode,flush=True)
pathlib.Path('/tmp/u066/validation.json').write_text(json.dumps(results,indent=2))
