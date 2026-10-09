import datetime,gzip,json,pathlib,re,shutil
P=pathlib.Path(__file__).resolve().parent
items=json.loads((P/'inventory.json').read_text());results=json.loads((P/'results.json').read_text());by={r['diff']:r for r in results};assert set(by)=={i['diff'] for i in items}
extra=json.loads((P/'skipped-extra.json').read_text());extra=['TestCountsAreRecorded']+sorted(t for t in extra if t!='TestCountsAreRecorded');skip=re.compile((P/'skip.regex').read_text().strip());details={};default_skips=set()
for item in items:
 r=by[item['diff']];assert r['completed'] and not r.get('broken');assert not r['runs'][-1]['panics'];caught=set();cases=set();outputs=[]
 for run in r['runs']:
  events=[]
  with gzip.open(P/run['log'],'rt') as f:
   for line in f:
    try:events.append(json.loads(line))
    except ValueError:pass
  assert not any(e.get('FailedBuild') for e in events)
  assert any(e.get('Action')=='start' for e in events)
  panic={p['test'] for p in run['panics']}
  failed={e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')}
  for e in events:
   top=e.get('Test','').split('/')[0]
   if e.get('Action') in ('run','fail') and top:assert not skip.search(top)
   if e.get('Action')=='skip' and top and e.get('Test')==top:default_skips.add(top)
   if e.get('Action')=='output' and e.get('Test') in failed and '_test.go:' in e.get('Output',''):outputs.append({'test':e['Test'],'line':e['Output'].strip(),'log':run['log']})
  caught.update(t.split('/')[0] for t in failed if t.split('/')[0] not in panic and t.split('/')[0] not in extra)
  cases.update(t for t in failed if t.split('/')[0] not in panic and t.split('/')[0] not in extra)
 assert sorted(caught)==r['all_catchers']
 expected={t.strip() for t in (P/'test-list.log').read_text().splitlines() if t.startswith('Test') and not skip.search(t.strip())}
 observed={e.get('Test','').split('/')[0] for e in events if e.get('Action')=='run'}
 assert expected - {p['test'] for p in r['panics']} <= observed, (item['diff'],sorted(expected-observed))
 details[item['diff']]={'failed_cases':sorted(cases),'assertion_output':outputs}
summary={'base':'7b9d4272','skipped_extra':extra,'mutants':[{k:by[i['diff']][k] for k in ('diff','file_line','all_catchers','panics','completed')} for i in items]}
(P/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');(P/'failure-details.json').write_text(json.dumps(details,indent=2)+'\n');(P/'default-skips.json').write_text(json.dumps(sorted(default_skips),indent=2)+'\n')
status=json.loads((P/'status.json').read_text());context=json.loads((P/'context.json').read_text());start=datetime.datetime.fromisoformat(context['started_utc']);end=datetime.datetime.now(datetime.timezone.utc)
timing={'elapsed_wall_seconds_including_build_diagnostics_and_restart':round((end-start).total_seconds(),3),'accepted_replay_sum_wall_seconds':round(sum(r['wall_seconds'] for r in results),3),'resumed_runner_wall_seconds':status['wall_seconds'],'completed_records':len(results),'concurrency':2,'ended_utc':end.isoformat()};(P/'timing.json').write_text(json.dumps(timing,indent=2)+'\n')
# Compress diagnostic and runner logs only after every replay has finished.
for f in list(P.rglob('*')):
 if f.is_file() and '.log' in f.name and not f.name.endswith('.gz'):
  with f.open('rb') as src,gzip.open(f.with_name(f.name+'.gz'),'wb',compresslevel=6) as dst:shutil.copyfileobj(src,dst)
  f.unlink()
with (P/'notes.txt').open('a') as f:f.write('All 19 accepted replays completed without failfast. summary.json lists all non-excluded top-level FAILs under the requested name-based exclusions; failure-details.json and raw compressed logs retain fixture subcases. Interrupted and build-only diagnostic attempts are excluded from accepted results.\n')
print(json.dumps({'records':len(results),'timing':timing,'catch_counts':[(r['diff'],len(r['all_catchers']),len(r['panics'])) for r in results]},indent=2))
