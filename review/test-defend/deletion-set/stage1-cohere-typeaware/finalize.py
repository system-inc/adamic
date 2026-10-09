import json,pathlib,subprocess
p=pathlib.Path(__file__).resolve().parent
planned=json.loads((p/'mutant-list.json').read_text());matrix=json.loads((p/'matrix.json').read_text())
candidate='TestVolumeProfileCorpora family'
result=dict(package='stage1/cohere/typeaware',main=(p/'main.txt').read_text().strip(),skipped=[candidate],mutants=[],keep=[],deletable=[])
lost=[];unresolved=[]
for m in planned:
 rs=[r for r in matrix if r['mutant']==m['mutant'] and r['branch']==m['branch']]
 if not m['stale']:assert rs,m['mutant']
 r=rs[-1] if rs else None
 catch=r['known_non_witness_catchers'] if r else []
 row=dict(mutant=m['mutant'],file_line=m['file_line'],branch=m['branch'],candidates_failed=[candidate],still_caught_by=catch,stale=m['stale'])
 if r:
  row['wall_seconds']=round(r['wall'],3)
  row['observed_failures']=r['failed']
  row['panicking_tests']=sorted({n for x in rs for n in x['panicking_tests']})
  row['stopped_after_clean_catch']=r['stopped_at_clean_failure']
  if not catch:
   if r['exit']==0 and r['binary_seconds'] is not None:lost.append(m['mutant'])
   else:unresolved.append(m['mutant'])
 else:unresolved.append(m['mutant'])
 result['mutants'].append(row)
if lost:result['keep']=[dict(test=candidate,because=', '.join(lost)+' loses its last catcher without it')]
elif unresolved:result['keep']=[dict(test=candidate,because='Replay remains unresolved for '+', '.join(unresolved)+'; deletion not established')]
else:result['deletable']=[candidate]
result['notes']=['Only shown mutants establish this decision.','The requested skip expression retains the enumeration-only TestVolumeProfileCorporaUnion.','Historical repository corpus manifest enabled for D1/D3 after a second clean baseline.']
(p/'result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
