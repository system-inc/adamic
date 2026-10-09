import subprocess,json,time,os,signal,difflib
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-lint-bv61ffb_decoded_options');base=(p/'base.txt').read_text().strip();source='stage1/cohere/lint/main.ts'
original=subprocess.check_output(['git','show',base+':'+source],text=True);old='function run(row: string, countOnly: boolean): number {';new=old+'\n    if (row.length >= 0) return 0;';changed=original.replace(old,new,1)
(p/'P1.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+source,tofile='b/'+source)))
with (p/'menu.jsonl').open('a') as f:f.write(json.dumps({'id':'P1','file':source,'line':original[:original.index(old)].count('\n')+1,'old':old,'new':new,'kind':'probe'})+'\n')
for mid in ('M1','M2','M3','P1'):
 before=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
 subprocess.run(['git','apply',str(p/(mid+'.diff'))],check=True)
 menu=[json.loads(l) for l in (p/'menu.jsonl').read_text().splitlines()];file=next(x['file'] for x in menu if x['id']==mid)
 subprocess.run(['git','add',file],check=True);subprocess.run(['git','commit','-m','audit scratch '+mid],check=True,stdout=subprocess.DEVNULL)
 variant=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
 try:
  pattern='^(TestDecodedOptionsAndMutant_[0-9]+|TestDecodedOptionsAndMutantUnion|TestCompilerAndStage1Agree_(002|007|012))$' if mid=='P1' else '^TestCompilerAndStage1Agree_(002|007|012)$'
  cmd=['timeout','100','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',pattern]
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u106/cache/'+mid+'-clean'
  if (p/(mid+'.log')).exists():(p/(mid+'.log')).rename(p/(mid+'.initial.log'))
  start=time.monotonic()
  with (p/(mid+'.log')).open('w') as out:
   proc=subprocess.Popen(cmd,env=env,stdout=out,stderr=subprocess.STDOUT,start_new_session=True);rc=proc.wait()
   try:os.killpg(proc.pid,signal.SIGKILL)
   except ProcessLookupError:pass
  records=[]
  for l in (p/(mid+'.log')).read_text().splitlines():
   try:records.append(json.loads(l))
   except:pass
  results=[json.loads(l) for l in (p/'results.jsonl').read_text().splitlines()];r=next(r for r in results if r['id']==mid)
  preserved=[] if mid=='P1' else [x for x in r['failures'] if not x.startswith('TestCompilerAndStage1Agree_')]
  oldcomplete=[] if mid=='P1' else [x for x in r['completed'] if not x.startswith('TestCompilerAndStage1Agree_')]
  r.update({'command':cmd,'initial_command':r['command'],'initial_precondition_failures_excluded':True,'scratch_commit':variant,'wall_replay':time.monotonic()-start,'returncode_replay':rc,'failures':preserved+[x['Test'] for x in records if x.get('Action')=='fail' and x.get('Test')],'completed':oldcomplete+[x['Test'] for x in records if x.get('Action') in ('pass','fail','skip') and x.get('Test')]})
  (p/'results.jsonl').write_text(''.join(json.dumps(x)+'\n' for x in results))
 finally:
  # This removes only our own temporary variant commit; evidence remains untracked.
  subprocess.run(['git','reset','--mixed',before],check=True,stdout=subprocess.DEVNULL)
  Path(file).write_text(subprocess.check_output(['git','show',base+':'+file],text=True))
