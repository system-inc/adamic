from pathlib import Path
import os,json,time,subprocess,difflib
p=Path('/tmp/u098');root=Path('/workspace/adamic');d=root/'stage1/cohere/graphql/printer';env=os.environ.copy()
source=(p/'post.py').read_text();exec(source[source.index('original='):source.index('records=[]')])
name,s,desc=mods['S4'];old='\tdata, err := os.ReadFile(loweredProduct + "/port.c")\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n';assert old in s;s=s.replace(old,'',1);mods['S4']=(name,s,desc+'; drop now-unused source-read statement and guard')
for mid,imp in [('P2','runtime'),('P4','github.com/system-inc/adamic/internal/native'),('P5','path/filepath'),('P5','github.com/system-inc/adamic/internal/buildcache')]:
 name,s,desc=mods[mid];mods[mid]=(name,s.replace('\t"'+imp+'"\n',''),desc+'; remove unused '+imp+' import')
records=[]
for mid in ['W1','W2','S1','S2','S3']:records.append(dict(id=mid,**json.loads((p/(mid+'-vet.meta')).read_text())))
try:
 for mid,(name,s,desc) in mods.items():
  if mid in ['W1','W2','S1','S2','S3']:continue
  (d/name).write_text(s);subprocess.run(['gofmt','-w',str(d/name)],check=True);s=(d/name).read_text();fn='stage1/cohere/graphql/printer/'+name
  (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original[name].splitlines(True),s.splitlines(True),fromfile='a/'+fn,tofile='b/'+fn)))
  cmd=['go','vet','./stage1/cohere/graphql/printer/'];t=time.monotonic()
  with (p/(mid+'-vet.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  meta=dict(command=cmd,exit=r.returncode,wall=time.monotonic()-t,description=desc);(p/(mid+'-vet.meta')).write_text(json.dumps(meta,indent=2));records.append(dict(id=mid,**meta));(d/name).write_text(original[name]);assert r.returncode==0,(mid,meta)
 s=body_replaced(original['printer.ts'],'export function format(',"    return { kind: 'Ok', text: '' };");name='printer.ts';(d/name).write_text(s);fn='stage1/cohere/graphql/printer/'+name
 (p/'P1.diff').write_text(''.join(difflib.unified_diff(original[name].splitlines(True),s.splitlines(True),fromfile='a/'+fn,tofile='b/'+fn)))
 cmd=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/graphql/printer/main.ts','-o','/tmp/u098/standalone/P1'];t=time.monotonic()
 with (p/'P1-build.log').open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 meta=dict(command=cmd,exit=r.returncode,wall=time.monotonic()-t,artifact_exists=(p/'standalone/P1').is_file());(p/'P1-build.meta').write_text(json.dumps(meta,indent=2));assert r.returncode==0 and meta['artifact_exists']
finally:
 for name,s in original.items():(d/name).write_text(s)
(p/'probe-validation.json').write_text(json.dumps(records,indent=2));(p/'post-done').write_text('done')
