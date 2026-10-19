import difflib,json,os,subprocess
from pathlib import Path
root=Path.cwd(); evidence=root/'review/compiler/loom-r-slice-6-m2'
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
changes=[('native','internal/native/runtime/exceptions.c', 'if (adamic_string_length(adamic_object_field(error, "name", &name_cache)->reference) != 0 && adamic_string_length(adamic_object_field(error, "message", &message_cache)->reference) != 0) { fputs(": ", stderr); }', 'fputs(": ", stderr);','TestUncaughtErrorEmptyMessage','native uncaught diagnostic'),('javascript','internal/javascript/javascript.go',"name === '' ? message : message === '' ? name : name + ': ' + message", "name + ': ' + message",'TestUncaughtErrorEmptyName','JavaScript uncaught diagnostic')]
results=[]
for name,path,before,after,test,catcher in changes:
 p=root/path; original=p.read_text(); assert original.count(before)==1
 changed=original.replace(before,after,1)
 (evidence/('always-separator-'+name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
 try:
  p.write_text(changed)
  command=['timeout','120','go','test','-p','1','./internal/oracle','-run','^'+test+'$','-count=1','-parallel=4','-timeout','90s','-json']
  log=evidence/('always-separator-'+name+'.jsonl')
  with log.open('w') as output: run=subprocess.run(command,stdout=output,stderr=subprocess.STDOUT,timeout=125)
  text=log.read_text(); caught=run.returncode==1 and catcher in text and 'uncaught diagnostic: got' in text
  results.append({'mutant':name,'command':command,'exit':run.returncode,'caught':caught,'catcher':catcher})
  print(name,run.returncode,'caught',caught,flush=True)
 finally: p.write_text(original)
(evidence/'separator-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(r['caught'] for r in results)
