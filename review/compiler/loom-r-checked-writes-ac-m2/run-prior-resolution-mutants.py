import pathlib,json,subprocess,difflib,time,os
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-checked-writes-ac-m2';rel='internal/lower/invariance.go';base=subprocess.check_output(['git','show','2dad6ae36c3b:'+rel],text=True);results=[]
for name,selector,needle in [('uncached-and','^TestImpliedTargetAndQueriesOnce$','queried'),('widen-and-left','^TestCheckedLogicalAndDoesNotWidenTruthTestedLeft$','readonly')]:
 old=(root/'review/compiler/loom-r-checked-writes-ac-m-wip'/(name+'.go.txt')).read_text();a=base.splitlines(True);b=old.splitlines(True);current=(root/rel).read_text()
 for g in reversed(list(difflib.SequenceMatcher(None,a,b,autojunk=False).get_grouped_opcodes(3))):
  before=''.join(a[g[0][1]:g[-1][2]]);after=''.join(b[g[0][3]:g[-1][4]]);assert current.count(before)==1,(name,before[:100]);current=current.replace(before,after,1)
 p=out/(name+'.go.txt');p.write_text(current);overlay=out/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/rel):str(p)}}));command=['timeout','120','go','test','-overlay',str(overlay),'./internal/lower','-run',selector,'-count=1','-parallel=4','-v','-timeout=90s'];start=time.monotonic()
 reused=name=='uncached-and'and os.environ.get('REUSE_UNCACHED')=='1'
 if reused:r=subprocess.CompletedProcess(command,1)
 else:
  with(out/(name+'.log')).open('w')as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 text=(out/(name+'.log')).read_text();assert r.returncode==1 and '--- FAIL:'in text and needle in text and 'build failed'not in text and 'error: 'not in text and 'test timed out'not in text
 results.append(dict(mutant=name,command=command,exit=r.returncode,caught=True,reused_recorded_execution=reused,seconds=round(time.monotonic()-start,3)));(out/'prior-resolution-mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(name,'caught',flush=True)
