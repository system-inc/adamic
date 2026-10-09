import os,sys,subprocess,pathlib,time,json,re
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-oracle-checked_views_v2_source'
rows='TestCheckedViewUntaggedSourceDispatch TestCheckedViewUntaggedCallableUnion TestCheckedViewUntaggedOptionalCallableControl TestCheckedViewUntaggedSourceFlows TestCheckedViewUntaggedOwnClassData TestCheckedViewUntaggedRecursive TestCheckedViewUntaggedArrayPending TestClassWrongOutput103 TestClassWrongOutput107 TestClassWrongOutput108 TestClassWrongOutput106 TestClockGenericReturnsT01Mutant TestClosureMergeRefusals'.split();pattern='^('+'|'.join(rows)+')$'
def run(name,args,env=None):
 st=time.monotonic()
 with (out/(name+'.log')).open('w') as f:p=subprocess.run(args,cwd=root,env=env or os.environ,stdout=f,stderr=subprocess.STDOUT)
 with (out/'commands.jsonl').open('a') as f:f.write(json.dumps(dict(name=name,command=args,wall=time.monotonic()-st,exit=p.returncode))+'\n')
 return p.returncode
if __name__=='__main__':
 assert all(r in (out/'list.log').read_text().splitlines() for r in rows)
 (out/'rows.json').write_text(json.dumps(rows,indent=2));(out/'scope.regex').write_text(pattern)
 code=run('scope-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(out/'scope.cover'),'./internal/oracle/','-run',pattern]);assert code==0,'red scoped baseline'
 run('coverage-functions',['go','tool','cover','-func='+str(out/'scope.cover')])
 for row in rows:
  for n in range(1,4):assert run(f'timing-{row}-{n}',['timeout','120','go','test','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$'])==0,(row,n)
