import json,re,statistics
from pathlib import Path
P=Path('review/test-audit/internal-buildcache');S=Path('/tmp/u015');rows=[x for x in (S/'list.log').read_text().splitlines() if x.startswith('Test')];plan=json.load(open(P/'plan.json'));matrix=[]
def proof(d,row):
 output=''.join(e.get('Output','') for e in d['events'] if e.get('Test','').split('/')[0]==row)
 line=next((l.strip() for l in output.splitlines() if 'buildcache_test.go:' in l and not 'build buildcache_test ' in l),'no failing line')
 return {'command':d['command'],'line':line,'output':output}
for m in plan:
 d=json.load(open(P/(m['id']+'.json')));failed=[r for r in rows if d['status'].get(r)=='fail'];matrix.append(dict(m,failed_rows=failed,status=d['status'],evidence={r:proof(d,r) for r in failed},subcases=d['subcases']))
(P/'matrix.json').write_text(json.dumps(matrix,indent=2));audit=[]
oracles={
'TestEveryInputChangesTheKey':'self-written metamorphic key comparisons; outside-input subcase accepts any error and survived the disabled guard M08',
'TestABuildRunsOncePerKey':'self-written build count, directory equality, product bytes and census text',
'TestAFailedBuildPublishesNothing':'self-written error identity, no leftover directory and next-build callback checks',
'TestUncachedModeBuildsEveryTime':'self-written two-build count and distinct-directory comparison',
'TestParallelCallersBuildOnce':'self-written one-build count and directory equality across eight goroutines',
'TestToolNamesItselfOnce':'self-written Go report prefix and missing-tool markers; does not independently compare the full report or repeat a command; M17 survived'
}
entry={'TestEveryInputChangesTheKey':['P_KEY'],'TestABuildRunsOncePerKey':['P_PRODUCT'],'TestAFailedBuildPublishesNothing':['P_GET','P_PRODUCT'],'TestUncachedModeBuildsEveryTime':['P_PRODUCT'],'TestParallelCallersBuildOnce':['P_GET'],'TestToolNamesItselfOnce':['P_TOOL']}
for row in rows:
 killed=[m for m in matrix if not m['probe'] and row in m['failed_rows']];unique=[m['id'] for m in killed if m['failed_rows']==[row]];last=killed[-1];p=last['evidence'][row];states=[next(m for m in matrix if m['id']==id)['status'].get(row) for id in entry[row]]
 subcases=[s for id in entry[row] for s,status in next(m for m in matrix if m['id']==id)['subcases'].items() if s.split('/')[0]==row and status=='pass']
 audit.append({'test':row,'package':'internal/buildcache','file':'internal/buildcache/buildcache_test.go','seconds':statistics.median(json.load(open(P/(row+'-timing-'+str(n)+'.json')))['binary_seconds'] for n in [1,2,3]),'oracle':oracles[row],'oracle_kind':'self','kills':[m['id'] for m in killed],'unique_kills':unique,'last_proven_fail':last['id']+': '+p['line'],'verdict':'sacred' if unique else 'cannot-judge','subsumed_by':[],'mutants_in_matrix':18,'probe_kills':[m['id'] for m in matrix if m['probe'] and row in m['failed_rows']],'subsumer_seconds':None,'vacuous':all(s=='pass' for s in states),'vacuous_subcases':subcases,'bounded':False,'matrix_rows':[],'evidence':next(m for m in killed if m['id']==unique[-1])['evidence'][row]['command']+'; '+next(m for m in killed if m['id']==unique[-1])['evidence'][row]['line'],'evidence_details':next(m for m in killed if m['id']==unique[-1])['evidence'][row],'entry_probes':entry[row]})
(P/'audit.json').write_text(json.dumps(audit,indent=2))
# Retain clean baseline/list output and all timing records. Raw execution output remained in log files.
events=[]
for line in (S/'baseline.log').read_text().splitlines():
 try:events.append(json.loads(line))
 except ValueError:pass
(P/'baseline.json').write_text(json.dumps(events,indent=2));(P/'scope.json').write_text(json.dumps({'tests':rows,'families':[],'skips':[]},indent=2))
e=json.load(open(P/'environment.json'));e['baseline']['binary_seconds']=next(x['Elapsed'] for x in events if not x.get('Test') and x['Action']=='pass');e['timing_runs_wall_seconds']=sum(json.load(open(f))['wall_seconds'] for f in P.glob('*-timing-*.json'));e['matrix_runs_wall_seconds']=sum(json.load(open(P/(m['id']+'.json')))['wall_seconds'] for m in plan);e['matrix_runs_binary_seconds']=sum(json.load(open(P/(m['id']+'.json')))['binary_seconds'] for m in plan);e['standalone_validation_seconds']=sum(x['seconds'] for x in json.load(open(P/'standalone-validation.json')));e['restored']=json.load(open(P/'restored.json'));e['restored'].pop('events');e['switch_build']=json.load(open(P/'switch-build.json'));e['survivor_probe_build']=json.load(open(P/'survivor-witnesses.json'))[0];(P/'environment.json').write_text(json.dumps(e,indent=2))
print(json.dumps(audit,indent=2));print('TIMES',e)
