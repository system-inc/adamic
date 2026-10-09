import json,pathlib,collections
p=pathlib.Path(__file__).parent
items=json.loads((p/'apply-check.json').read_text());rows=json.loads((p/'matrix.json').read_text())
assert len(items)==len(rows)==9, 'Replay incomplete'
candidates=['TestInterfaceDefaultGap','TestInterfaceTypeMethodGap','TestUnattachedDecoratorControl']
main=(p/'main.txt').read_text().strip()
mutants=[]
for r in rows:
 assert not r.get('broken'),r
 if not r['stale']:
  assert r['runs'],r
  if not r['still_caught_by']:
   last=r['runs'][-1]
   assert last['completed'] and not last['early_stopped'] and last['exit']==0,r
   assert len(last['statuses'])==252-len(r['panicking_tests']),r
 mutants.append({k:r[k] for k in ['mutant','file_line','branch','candidates_failed','still_caught_by','stale']} | {'wall_seconds':round(sum(x['wall_seconds'] for x in r['runs']),3),'witness_failures':r['witness_failures'],'panicking_tests':r['panicking_tests'],'early_stopped':any(x['early_stopped'] for x in r['runs'])})
keep=[];deletable=[]
for c in candidates:
 relevant=[r for r in rows if c in r['candidates_failed']]
 lost=[r for r in relevant if not r['stale'] and not r['still_caught_by']]
 assert not any(r['stale'] for r in relevant), 'Stale evidence prevents deletion judgment'
 if lost:
  keep.append({'test':c,'because':'; '.join(r['branch']+'/'+r['mutant']+' loses its last catcher without the set' for r in lost)})
 else:deletable.append(c)
report={'package':'stage1/cohere/estree','main':main,'skipped':candidates,'mutants':mutants,'keep':keep,'deletable':deletable}
(p/'report.json').write_text(json.dumps(report,indent=2)+'\n')
b=json.loads((p/'baseline-result.json').read_text());seconds=sum(sum(x['wall_seconds'] for x in r['runs']) for r in rows)
(p/'REPORT.md').write_text(f'''# ESTree deletion-set replay

Starting commit: `{main}`. All nine gathered standalone diffs applied unchanged.

The clean baseline with all three candidates skipped passed in {b['wall_seconds']:.3f} wall seconds ({b['binary_seconds']:.3f} binary seconds). Completed mutant replays used {seconds:.3f} wall seconds. All commands use ADAMIC_GATE_UNCACHED=1 and a fresh ADAMIC_BUILD_CACHE_DIR, with a 30 minute test timeout.

Mutant runs with ordinary failures were stopped after the first clean non-witness failure and the buffered events were drained. Their matrix statuses list only observed rows; later rows are unknown. Runs with no ordinary catcher completed the entire default package gate. Witness failures and panicking tests are reported separately and never credited as sole catchers.

The two interface candidates jointly protect the lost catches. At least one of that pair must remain; this replay does not establish that both are individually necessary. The unattached decorator control is deletable only in the narrow sense that every gathered mutant it failed has an observed ordinary catcher outside the set. Nothing was deleted or weakened.

Twelve optional rows skipped in the baseline: {', '.join(b['rows_skipped'])}. Conclusions cover the default gate, including newly discovered tests, rather than unavailable optional corpora.

Costs and interruptions: historical branches reused mutant names in root and session matrices, so each diff was paired with its adjacent matrix and given a branch-qualified filename. An initial partial M06 replay was abandoned to fix panic recording; it has no verdict. The workspace restarted during gaps D3; that partial log has no verdict, and remaining replays restarted with fresh replay-v3 caches. /tmp has a total capacity below the requested 15 GB free, despite removing prior-unit scratch. No baseline or completed replay failed from disk exhaustion.

See mutant-list.json for the catalog written before replay, matrix.json for commands and all observed statuses, report.json for the decision, and the adjacent raw JSON logs. Production diffs were restored after each run. Only this evidence directory is committed and pushed.
''')
print(json.dumps(report,indent=2));print('replay wall seconds',seconds)
