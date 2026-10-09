import pathlib,json,subprocess,re
r=pathlib.Path(__file__).resolve().parent
xs=json.loads((r/'matrix.json').read_text());cs=json.loads((r/'skipped.json').read_text());keep=[];delete=[]
for x in xs:
 lines=(r/x['diff']).read_text().splitlines();f=next(l[6:] for l in lines if l.startswith('+++ b/'));old=next(l[1:] for l in lines if l.startswith('-') and not l.startswith('---'))
 base=subprocess.check_output(['git','show','origin/main:'+f]).decode().splitlines();loc=[i+1 for i,l in enumerate(base) if l==old]
 if loc:x['file_line']=f+':'+str(loc[0])
for c in cs:
 guards=[x for x in xs if c in x['candidates_failed']]
 lost=[x for x in guards if not x['stale'] and not x.get('broken') and x.get('complete') and not x['still_caught_by']]
 if lost:keep.append(dict(test=c,because=', '.join(x['mutant']+' on '+x['branch'] for x in lost)+' loses its last catcher without the set; the two refusal candidates are interchangeable for these mutants'))
 elif all(x.get('still_caught_by') for x in guards):delete.append(c)
report=dict(package='internal/lower',main=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip(),skipped=cs,mutants=[{k:x[k] for k in ['mutant','file_line','branch','candidates_failed','still_caught_by','stale']} for x in xs],keep=keep,deletable=delete)
report['notes']=['Selection follows the requested M*.diff and D*.diff globs; interface defender round2 F01-F03 are excluded.','Keep findings protect the refusal coverage collectively; either refusal candidate alone catches the diagnostic mutants in prior evidence.']
report['replay_wall_seconds']=round(sum(z['wall_seconds'] for x in xs for z in x['runs']),3)
report['witness_failures']={x['branch']+':'+x['mutant']:x['witness_failures'] for x in xs if x['witness_failures']};report['panicking_tests']={x['branch']+':'+x['mutant']:x['panicking_tests'] for x in xs if x['panicking_tests']}
(r/'report.json').write_text(json.dumps(report,indent=2)+'\n');(r/'matrix.json').write_text(json.dumps(xs,indent=2)+'\n');print(json.dumps(report,indent=2))
