import json, subprocess, time, os, difflib
from pathlib import Path
p=Path(__file__).parent
owned=Path('stage1/cohere/lint/rules/no-underscore-dangle')
base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip()
menu=[('D1', 'new Parser(source, path)', 'new Parser(path, source)', 'swap arguments: parser input and filename'),('D2', '(left, right) => left.start - right.start', '(right, left) => left.start - right.start', 'swap callback arguments: reverse finding ordering'),('D3', 'ancestry(context, parents, child, index)', 'ancestry(context, parents, index, child)', 'swap recursive arguments: revisit parent instead of descending')]
original=subprocess.check_output(['git','show',base+':'+str(owned/'profile.a')],text=True)
records=[{'mutant':i,'file_line':str(owned/'profile.a')+':'+str(original[:original.index(a)].count('\n')+1),'change':desc,'from':a,'to':b} for i,a,b,desc in menu]
(p/'menu.json').write_text(json.dumps(records,indent=2)+'\n')
(p/'witness.ts').write_text("const _lead = 1; const tail_ = 2; this._x; function _fn(_arg) {}\n")
(p/'witness-manifest.txt').write_text(str((p/'witness.ts').resolve())+'\t'+json.dumps({'allowFunctionParams':False})+'\n')
node=['timeout','10','node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(owned/'profile.a'),str(p/'witness-manifest.txt')]
def run(label,cmd,env=None):
 t=time.monotonic()
 with (p/(label+'.log')).open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
 result={'label':label,'command':cmd,'exit':r.returncode,'wall_seconds':time.monotonic()-t}
 with (p/'commands.jsonl').open('a') as f:f.write(json.dumps(result)+'\n')
 return result
baseline=run('witness-baseline',node)
assert baseline['exit']==0
results=[]
for record,(mid,a,b,desc) in zip(records,menu):
 assert original.count(a)==1
 changed=original.replace(a,b)
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(owned/'profile.a'),tofile='b/'+str(owned/'profile.a')))
 (p/(mid+'.diff')).write_text(diff)
 subprocess.run(['git','apply','--check',str(p/(mid+'.diff'))],check=True)
 subprocess.run(['git','apply',str(p/(mid+'.diff'))],check=True)
 try:
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/underscore-defend/cache/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/rules/no-underscore-dangle/','-run','.']
  matrix=run(mid,cmd,env)
  witness=run('witness-'+mid,node)
  events=[]
  for line in (p/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  before=(p/'witness-baseline.log').read_text();after=(p/('witness-'+mid+'.log')).read_text()
  record.update(matrix=matrix,witness=witness,rows_failed=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],rows_passed=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test')],test_seconds=[e['Elapsed'] for e in events if e.get('Action')=='pass' and e.get('Test')=='TestCompileProfiles'],witness_changed=before!=after,witness_diff=list(difflib.unified_diff(before.splitlines(),after.splitlines())))
  results.append(record);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n')
  print(mid,matrix['exit'],record['rows_failed'],record['test_seconds'],flush=True)
 finally:subprocess.run(['git','apply','--reverse',str(p/(mid+'.diff'))],check=True)
