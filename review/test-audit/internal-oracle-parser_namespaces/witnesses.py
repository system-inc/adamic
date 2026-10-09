import pathlib,json,subprocess,os,time,difflib
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');records=[];s=pathlib.Path('/tmp/u067-unresolved.a');s.write_text('let read = () => Debug.isDebugging;\nconsole.log(`${read()}`);\nnamespace Debug { export let isDebugging = false; }\n');(p/'M02.unresolved.a.txt').write_bytes(s.read_bytes())
for id,source in [('M02',str(s)),('M06','internal/lower/testdata/namespaces_notyet/reaching_direct.a')]:
 for label,selected in [('before',''),('after',id)]:
  env=os.environ.copy();env['ADAMIC_MUTANT']=selected;cmd=['/tmp/u067-adamic','js',source];t=time.monotonic();target=p/(id+'.independent.'+label+'.mjs')
  with target.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  records.append(dict(command='ADAMIC_MUTANT='+selected+' '+' '.join(cmd),exit=r.returncode,wall=time.monotonic()-t));print(id,label,'compiler',r.returncode,flush=True)
  if r.returncode==0:
   cmd=['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(target.resolve())];t=time.monotonic()
   with (p/(id+'.independent.'+label+'.run.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
   records.append(dict(command=' '.join(cmd),exit=r.returncode,wall=time.monotonic()-t));print(id,label,'runtime',r.returncode,flush=True)
 a=(p/(id+'.independent.before.mjs')).read_text();b=(p/(id+'.independent.after.mjs')).read_text();(p/(id+'.independent.diff')).write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before',tofile='after')))
(p/'independent-commands.json').write_text(json.dumps(records,indent=2))
