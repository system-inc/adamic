import pathlib,json,subprocess,time,os,difflib
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');records=[];t=time.monotonic()
with (p/'probe-build.log').open('w') as f:r=subprocess.run(['timeout','90','go','build','-o','/tmp/u067-adamic','./cmd/adamic'],stdout=f,stderr=subprocess.STDOUT)
records.append(dict(command='timeout 90 go build -o /tmp/u067-adamic ./cmd/adamic',exit=r.returncode,wall=time.monotonic()-t))
if r.returncode==0:
 for id in ['M02','M06']:
  for name in ['namespace_callable_properties','namespace_class_registration','namespace_method_receiver']:
   source='internal/oracle/testdata/'+name+'.a'
   for label,selected in [('before',''),('after',id)]:
    env=os.environ.copy();env['ADAMIC_MUTANT']=selected;t=time.monotonic()
    with (p/(id+'.'+name+'.'+label+'.js')).open('w') as f:r=subprocess.run(['/tmp/u067-adamic','js',source],env=env,stdout=f,stderr=subprocess.STDOUT)
    records.append(dict(command='ADAMIC_MUTANT='+selected+' /tmp/u067-adamic js '+source,exit=r.returncode,wall=time.monotonic()-t))
   a=(p/(id+'.'+name+'.before.js')).read_text();b=(p/(id+'.'+name+'.after.js')).read_text();d=''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before',tofile='after'));(p/(id+'.'+name+'.witness.diff')).write_text(d);print(id,name,'changed',a!=b,flush=True)
(p/'survivor-commands.json').write_text(json.dumps(records,indent=2))
