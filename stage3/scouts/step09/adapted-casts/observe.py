"""Run Node goldens, real source mutants and in-place native builds."""
import hashlib,json,pathlib,re,subprocess,sys,tempfile
root=pathlib.Path(__file__).resolve().parent
repo=root.parents[3]
binary=pathlib.Path(sys.argv[1]).resolve()
expected=['["a","b"]\n["a","b"]\n["a","b"]\n[""]\n','true\nstring\ntrue\ntrue\n','{"id":1}accepted\nTypeError\n']
mutations=[('split(/\\r\\n?|\\n/)','split(/\\r\\n?|(\\n)?/)'),('.typeExpression = typeExpression;','.typeExpression = undefined;'),('JSON.stringify(descriptor))','JSON.stringify(descriptor) ?? "")')]
rows=[]
for i,p in enumerate(sorted((root/'fixtures').glob('*.a'))):
 def run(cmd,label):
  result=subprocess.run(cmd,cwd=repo,capture_output=True,text=True)
  for channel in ['stdout','stderr']:(root/'evidence'/(p.name+'.'+label+'.'+channel+'.txt')).write_text(getattr(result,channel))
  return dict(stdout=result.stdout,stderr=result.stderr,exit=result.returncode)
 node=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(p)],'node')
 assert node==dict(stdout=expected[i],stderr='',exit=0),(p,node)
 old,new=mutations[i]; text=p.read_text(); assert text.count(old)==1,(p,old)
 with tempfile.TemporaryDirectory(prefix='adapted-cast-mutant-') as tmp:
  mutant=pathlib.Path(tmp)/p.name;mutant.write_text(text.replace(old,new))
  observed=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(mutant)],'mutant')
  assert observed['exit']==0 and observed['stderr']=='' and observed['stdout']!=node['stdout'],(p,observed)
 native=run([str(binary),'build',str(p.relative_to(repo)),'-o','/tmp/adapted-casts-'+p.stem],'build')
 errors=native['stderr']
 if ' error TS' in errors:
  outcome='Checker';header='// a-check: type error '+re.search(r'error (TS\d+)',errors).group(1)
 elif 'Adamic 0.1 refuses ' in errors:
  outcome='Refused';header='// a-check: refused '+errors.split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
 elif "can't lower" in errors and ' yet' in errors:outcome='NotYet';header=None
 elif native['exit']==0:
  outcome='Compiles';header=None
  executed=run(['/tmp/adapted-casts-'+p.stem],'native')
  assert executed==node,('SILENT MISCOMPILE',p,executed,node)
 else:raise AssertionError((p,native))
 if header:
  if text.startswith('// a-check:'):text=text.split('\n',1)[1]
  p.write_text(header+'\n'+text)
 rows.append(dict(file=p.name,node=node,stage0=dict(outcome=outcome,what=errors),header=header,body_sha256=hashlib.sha256(text.encode()).hexdigest(),mutant=dict(change=[old,new],observation=observed,caught_by='independent Node stdout golden')))
(root/'fixtures/status.json').write_text(json.dumps(rows,indent=2)+'\n')
print('PASS: three Node goldens, three source mutants; native results and required first-line headers recorded')
