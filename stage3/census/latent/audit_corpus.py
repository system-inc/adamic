"""Plant one overlay-only NotYet in a real eligible compiler function."""
import collections, json, os, pathlib, subprocess, sys
scratch=pathlib.Path(sys.argv[1]).resolve(); adapted=pathlib.Path(sys.argv[2]).resolve()
LABEL='measured on a checker-rejected program'
configuration=sys.argv[3] if len(sys.argv)>3 else 'main'
evidence=pathlib.Path(sys.argv[4]).resolve() if len(sys.argv)>4 else scratch
evidence.mkdir(parents=True,exist_ok=True)
baseline=[json.loads(l) for l in (scratch/(configuration+'.jsonl')).read_text().splitlines()]
where=str(adapted/'src/compiler/binder.ts')+':330:1'; name='getModuleInstanceState'
unit=next(u for row in baseline[1:] for u in row['units'] if u['where']==where)
assert unit['kind']=='KindFunctionDeclaration' and unit['status']!='skipped_checker_body'
env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1',LATENT_MUTANT_FUNCTION=name,LATENT_MUTANT_WHERE=where)
with (evidence/'corpus-mutant-run.log').open('w') as log:
 subprocess.run([str(scratch/(configuration+'-census')),str(adapted/'src/compiler'),str(evidence/'corpus-mutant.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
mutant=[json.loads(l) for l in (evidence/'corpus-mutant.jsonl').read_text().splitlines()]
assert mutant[0]==baseline[0]
by_file={r['file']:r for r in mutant[1:]};deltas={}
for before in baseline[1:]:
 after=by_file[before['file']];assert before['units']==after['units']
 key=lambda f:(f['kind'],f['where'],f['reason'],f['text'])
 old={key(f) for f in before['findings']};new={key(f) for f in after['findings']}
 assert not old-new
 deltas[before['file'].replace(str(adapted)+'/', '')]=len(new)-len(old)
 if before['file']==str(adapted/'src/compiler/binder.ts'):
  added=new-old;assert len(added)==1
  finding=added.pop();assert finding[0]=='NotYet' and finding[1]==where and finding[2]=='latent planted extra NotYet'
 else:assert after==before
assert sum(deltas.values())==1
result=dict(measurement=LABEL,function=name,where=where.replace(str(adapted)+'/', ''),per_file_delta=deltas,extra_NotYet=1,unchanged_other_files=len(deltas)-1,pass_no_IR_output=True)
(evidence/'corpus-mutant-audit.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
