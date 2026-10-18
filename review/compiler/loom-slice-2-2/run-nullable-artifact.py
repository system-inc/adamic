import os,subprocess,json,time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-slice-2-2/callback-mutants/nullable-token-omitted';fixture='internal/oracle/testdata/callback_widening/nullable-object-map.a'
records=[]
def run(label,args,env=None,limit=90):
 start=time.monotonic();done=subprocess.run(['timeout',str(limit)]+args,env=env,capture_output=True,timeout=limit+10)
 (out/(label+'.stdout.log')).write_bytes(done.stdout);(out/(label+'.stderr.log')).write_bytes(done.stderr)
 row=dict(label=label,command=['timeout',str(limit)]+args,exit=done.returncode,stdout=done.stdout.decode(errors='replace'),stderr=done.stderr.decode(errors='replace'),seconds=time.monotonic()-start);records.append(row);(out/'artifact-results.json').write_text(json.dumps(records,indent=2)+'\n');print(label,done.returncode,round(row['seconds'],3),flush=True);return row
node=run('node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',fixture]);assert (node['stdout'],node['stderr'],node['exit'])==('1,null\n','',0)
for name,flags in [('release',[]),('sanitized',['--sanitize'])]:
 binary='/tmp/adamic-slice2-nullable-'+name
 build=run(name+'-build',['/tmp/adamic-slice2-nullable-mutant','build',fixture,'-o',binary]+flags,limit=120);assert build['exit']==0
 observed=run(name,[binary],dict(os.environ,ASAN_OPTIONS='detect_leaks=0'));assert tuple(observed[k] for k in ['stdout','stderr','exit']) != tuple(node[k] for k in ['stdout','stderr','exit'])
 assert observed['exit']==70 and observed['stderr']=='adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n'
row=dict(mutant='nullable-token-omitted',caught=True,catcher='Node stdout/stderr/exit comparison catches the narrowed-union checked stop in both release and sanitized artifacts',base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip(),observations=records)
(out/'artifact-result.json').write_text(json.dumps(row,indent=2)+'\n')
