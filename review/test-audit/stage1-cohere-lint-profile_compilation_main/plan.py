from pathlib import Path
import subprocess,json,difflib
p=Path('review/test-audit/stage1-cohere-lint-profile_compilation_main');base=lambda f:subprocess.check_output(['git','show','HEAD:'+f],text=True)
plan=[('M01','stage1/cohere/lint/main.ts','const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + 1;','const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + 2;','change constant'),('M02','stage1/cohere/lint/main.ts','        const fixed = linter.fixed();\n','', 'drop statement'),('M03','stage1/cohere/lint/lint.ts','        rules.visit(index, parent);\n','', 'drop statement')]
# M02 drops the matching output statement as well, so no unused binding remains.
result=[]
for mid,f,old,new,menu in plan:
 a=base(f);assert a.count(old)==1;b=a.replace(old,new)
 if mid=='M02':b=b.replace('        console.log(`fixed\\t${written(fixed)}`);\n','')
 (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 result.append(dict(id=mid,file=f,line=a[:a.index(old)].count('\n')+1,before=old,after=new,menu=menu))
(p/'plan.json').write_text(json.dumps(result,indent=2))
checks=[dict(id='W01',file='stage1/cohere/lint/lint_test.go',entry='difference',change='return empty comparison at entry',kind='weaken agreement comparison'),dict(id='W02',file='stage1/cohere/lint/registration_test.go',entry='TestRegistrationMutant',change='bytes.Equal(side.run.output, want) becomes true',kind='weaken equality comparison'),dict(id='S01',file='stage1/cohere/lint/profile_compilation_main_test.go',entry='compilationShard',change='testProfileCompilationShards = 1 becomes 2',kind='change construction constant'),dict(id='S02',file='stage1/cohere/lint/factory_hooks_shards_test.go',entry='factoryHooksProducts',change='value.rows length must be 5 rather than 4',kind='change construction constant'),dict(id='W03',file='stage1/cohere/lint/registration_test.go',entry='TestNestedOutsideModuleCopy',change='weaken the bad-import detection to treat a Node error and a Lower error as no disagreement',kind='weaken nested loading checks')]
(p/'check-plan.json').write_text(json.dumps(checks,indent=2))
# Primary port answer probe. Standalone whole-body early return removes now-unused imports later verified by own build.
f='stage1/cohere/lint/main.ts';a=base(f);start=a.index('function run(');end=a.index('\nconst args =',start);b=a[:start]+a[start:a.index('{',start)+1]+'\n    return 0;\n}\n'+a[end:];(p/'P01.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
