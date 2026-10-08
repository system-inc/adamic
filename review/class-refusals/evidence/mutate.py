from pathlib import Path
import subprocess,json,os
root=Path.cwd(); out=Path('/tmp/class-refusals'); observations=[]
def run(args):
 p=subprocess.run(args,capture_output=True,text=True,timeout=180)
 return dict(exit=p.returncode,stdout=p.stdout,stderr=p.stderr)
def observe(name,path):
 record={'name':name,'path':path,'node':run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',path])}
 js=run(['go','run','./cmd/adamic','js',path]); record['js_build']=dict(js)
 if js['exit']==0:
  p=out/(name+'-mutant.mjs'); p.write_text(js['stdout']); record['backend']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(p)])
 for mode,flags in [('release',[]),('sanitized',['--sanitize'])]:
  p=out/(name+'-mutant-'+mode); record[mode+'_build']=run(['go','run','./cmd/adamic','build',path,'-o',str(p)]+flags)
  if record[mode+'_build']['exit']==0: record[mode]=run([str(p)])
 return record
mutants=[
 ('super','internal/lower/class_inheritance.go','TestClassWrongOutput103',['classfeat_init_super_number','classfeat_init_super_getter','classfeat_init_super'],['super_deferred']),
 ('iterator','internal/lower/iteration.go','TestClassWrongOutput107',['iterators_override_this','iterators_hidden_return'],['iterator_base_parameter']),
 ('keys','internal/lower/class_features.go','TestClassWrongOutput108',['iterators_sym_keys_view'],['keys_unrelated']),
 ('private','internal/lower/class_static.go','TestClassWrongOutput106',['classfeat_static_private_instance','classfeat_static_private_method'],[])]
for name,file,test,originals,safe in mutants:
 path=root/file; original=path.read_text()
 if name=='super':
  old=subprocess.check_output(['git','show','5348895b:'+file],text=True)
  start='func (l *lowering) initializerReads('; end='// Private names with the same spelling'
  replacement=old[old.index(start):old.index(end)]
  changed=original[:original.index(start)]+replacement+original[original.index(end):]
  changed=changed.replace('\n\t"strings"','')
 elif name=='iterator': changed=original.replace('if l.iterationFactoryReturnsThis(entry) &&', 'if false && l.iterationFactoryReturnsThis(entry) &&',1)
 elif name=='keys': changed=original.replace('if !fresh &&', 'if false && !fresh &&',1)
 else: changed=original.replace('if target.Parent == declaration &&', 'if false && target.Parent == declaration &&',1)
 try:
  path.write_text(changed)
  result=run(['go','test','./internal/oracle','-run','^'+test+'$','-count=1','-timeout','3m'])
  (out/(name+'-mutant-test.log')).write_text(result['stdout']+result['stderr'])
  entry={'guard':name,'test':test,'test_result':result,'observations':[]}
  for probe in originals: entry['observations'].append(observe(probe,'internal/oracle/testdata/'+probe+'.a'))
  for probe in safe: entry['observations'].append(observe(probe,str(out/'probes'/(probe+'.a'))))
  observations.append(entry); (out/'mutants.json').write_text(json.dumps(observations,indent=2))
  print(name,'test exit',result['exit'],flush=True)
  for r in entry['observations']:
   print(r['name'],'Node',r['node'],'release',r.get('release',r['release_build']),'sanitized',r.get('sanitized',r['sanitized_build']),'backend',r.get('backend',r['js_build']),flush=True)
 finally: path.write_text(original)
