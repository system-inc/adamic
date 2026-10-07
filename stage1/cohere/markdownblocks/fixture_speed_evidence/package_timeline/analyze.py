import json,datetime,pathlib,csv,sys,math
root=pathlib.Path(sys.argv[1]);out=pathlib.Path(sys.argv[2]);out.mkdir(parents=True,exist_ok=True)
events=[json.loads(l) for l in (root/'package.jsonl').read_text().splitlines() if l.strip()]
def epoch(s):return datetime.datetime.fromisoformat(s.replace('Z','+00:00')).timestamp()
start=next(epoch(e['Time']) for e in events if e['Action']=='start' and 'Test' not in e)
end=max(epoch(e['Time']) for e in events if e['Action'] in ('pass','fail','skip') and 'Test' not in e)
tests={}; changes=[]
for e in events:
 name=e.get('Test');action=e['Action'];sec=epoch(e['Time'])-start
 if not name or action not in ('run','cont','pause','pass','fail','skip'):continue
 changes.append({'seconds':sec,'action':action,'test':name})
 if action=='run':
  assert name not in tests
  tests[name]={'test':name,'parent':name.rsplit('/',1)[0] if '/' in name else None,'run_seconds':sec,'first_cont_seconds':None,'end_seconds':None,'running_intervals':[],'paused_intervals':[],'state':'running','since':sec}
  continue
 r=tests[name]
 if action=='pause':
  assert r['state']=='running';r['running_intervals'].append([r['since'],sec]);r['state']='paused';r['since']=sec
 elif action=='cont':
  assert r['state']=='paused';r['paused_intervals'].append([r['since'],sec]);r['state']='running';r['since']=sec
  if r['first_cont_seconds'] is None:r['first_cont_seconds']=sec
 else:
  r[r['state']+'_intervals'].append([r['since'],sec]);r['end_seconds']=sec;r['result']=action;r['go_elapsed_seconds']=e.get('Elapsed');del r['state'];del r['since']
def union(intervals):
 result=[]
 for a,b in sorted(intervals):
  if result and a<=result[-1][1]:result[-1][1]=max(b,result[-1][1])
  else:result.append([a,b])
 return result
def duration(intervals):return sum(b-a for a,b in intervals)
def clip(intervals,a,b):return [[max(a,x),min(b,y)] for x,y in intervals if y>a and x<b]
source_waits=json.loads((root/'source-waits.json').read_text())
top_barrier=min(r['first_cont_seconds'] for n,r in tests.items() if '/' not in n and r['first_cont_seconds'] is not None)
for name,r in tests.items():
 assert r['end_seconds'] is not None
 r['running_seconds']=duration(r['running_intervals']);r['paused_seconds']=duration(r['paused_intervals'])
 child=union([i for n,d in tests.items() if d['parent']==name for i in d['running_intervals']])
 r['running_child_overlap_seconds']=sum(duration(clip(child,a,b)) for a,b in r['running_intervals'])
 r['outside_running_children_seconds']=r['running_seconds']-r['running_child_overlap_seconds']
 r['event_interval_minus_go_elapsed_seconds']=r['running_seconds']-(r['go_elapsed_seconds'] or 0)
 r['elapsed_caveat']='Event completion can be buffered at parent. Go elapsed is rounded to 0.01 s and excludes parallel-child barrier wait; event interval includes that wait.'
 r['wait_evidence']=source_waits[name.split('/')[0]]
 r['top_serial_barrier_pause_seconds']=duration(clip(r['paused_intervals'],0,top_barrier)) if not r['parent'] else 0
 r['after_top_barrier_pause_seconds']=r['paused_seconds']-r['top_serial_barrier_pause_seconds']
 if '/' in name:
  if name=='TestMarkdownListLayout/unordered_marker':r['wait_evidence']='lists_test.go:397,439: synchronous Node output mutant, load/lower, fresh -O0 sanitized native canary build/run; no layoutOnce inside subtest'
  elif name.startswith('TestParserRepresentationProbes/'):r['wait_evidence']='gaps_test.go:33: parallel child; parent barrier/-parallel slot before CONT, Node/load/lower then native/leaks or expected compiler refusal'
  elif name.startswith('TestMarkdownWhitespaceLayout/policy_'):r['wait_evidence']='whitespace_test.go:50: synchronous Node policy mutant; baseline whitespace lowering/native/leaks happened in parent before these subtests'
  elif name.startswith('TestWholeDocumentOraclePreflight/'):r['wait_evidence']='audit_test.go:276: synchronous Go oracle control, Go build/run and comparison'
  elif name.startswith('TestMdastMalformedEvents/'):r['wait_evidence']='mdast_errors_test.go:50: synchronous native/Node/backend diagnostic control; native artifact promise shares identical source build'
  else:r['wait_evidence']='Synchronous output mutant child; parent waits in t.Run; '+r['wait_evidence']
 if r['paused_intervals']:r['pause_wait']='Go testing parent barrier and/or -parallel slot; exact pause/CONT interval measured'
 else:r['pause_wait']='none'
samples=[json.loads(l) for l in (root/'samples.jsonl').read_text().splitlines()]
hist=[]
for a in range(0,math.ceil(end-start),30):
 b=min(a+30,end-start);cpu=throttle=0
 for left,right in zip(samples,samples[1:]):
  x=max(start+a,left['epoch']);y=min(start+b,right['epoch'])
  if y<=x:continue
  frac=(y-x)/(right['epoch']-left['epoch'])
  cpu+=(right['cpu_stat']['usage_usec']-left['cpu_stat']['usage_usec'])*frac/1e6
  throttle+=(right['cpu_stat']['throttled_usec']-left['cpu_stat']['throttled_usec'])*frac/1e6
 topsec=sum(duration(clip(r['running_intervals'],a,b)) for n,r in tests.items() if '/' not in n)
 leafsec=0;maxleaf=0
 bounds=sorted({a,b}|{v for r in tests.values() for i in r['running_intervals'] for v in i if a<v<b})
 for x,y in zip(bounds,bounds[1:]):
  middle=(x+y)/2;active={n for n,r in tests.items() if any(l<=middle<h for l,h in r['running_intervals'])}
  leaves={n for n in active if not any(k.startswith(n+'/') for k in active)}
  leafsec+=(y-x)*len(leaves);maxleaf=max(maxleaf,len(leaves))
 hist.append({'start_seconds':a,'end_seconds':b,'average_busy_cpu_equivalents_cgroup':cpu/(b-a),'average_active_top_tests_events':topsec/(b-a),'average_active_leaf_tests_events':leafsec/(b-a),'maximum_active_leaf_tests_events':maxleaf,'cgroup_throttled_seconds':throttle})
# Exact intervals with only one active top-level test; tests can have many child intervals.
bounds=sorted({0,end-start}|{v for n,r in tests.items() if '/' not in n for i in r['running_intervals'] for v in i})
alone=[]
for a,b in zip(bounds,bounds[1:]):
 active=[n for n,r in tests.items() if '/' not in n and any(l<=(a+b)/2<h for l,h in r['running_intervals'])]
 if len(active)==1:
  if alone and alone[-1]['test']==active[0] and abs(alone[-1]['end_seconds']-a)<1e-6:alone[-1]['end_seconds']=b
  else:alone.append({'start_seconds':a,'end_seconds':b,'test':active[0]})
(out/'tests.json').write_text(json.dumps(list(tests.values()),indent=2)+'\n')
(out/'events.json').write_text(json.dumps(changes,indent=2)+'\n')
(out/'histogram.json').write_text(json.dumps(hist,indent=2)+'\n')
(out/'alone.json').write_text(json.dumps(alone,indent=2)+'\n')
columns=['test','parent','run_seconds','first_cont_seconds','end_seconds','running_seconds','go_elapsed_seconds','event_interval_minus_go_elapsed_seconds','paused_seconds','top_serial_barrier_pause_seconds','after_top_barrier_pause_seconds','running_child_overlap_seconds','outside_running_children_seconds','result','pause_wait','wait_evidence']
with (out/'tests.csv').open('w') as f:
 w=csv.DictWriter(f,fieldnames=columns);w.writeheader();w.writerows({k:r[k] for k in columns} for r in sorted(tests.values(),key=lambda r:r['run_seconds']))
summary={'package_start_epoch':start,'package_event_wall_seconds':end-start,'package_reported_elapsed_seconds':next(e.get('Elapsed') for e in reversed(events) if 'Test' not in e and e['Action'] in ('pass','fail','skip')),'tests':len(tests),'top_level_tests':sum('/' not in n for n in tests),'results':{a:sum(r['result']==a for r in tests.values()) for a in ('pass','fail','skip')},'last_finishing_top_tests':sorted([r for n,r in tests.items() if '/' not in n],key=lambda r:r['end_seconds'],reverse=True)[:12]}
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('Package seconds',end-start,'tests',len(tests))
for r in summary['last_finishing_top_tests']:print(r['test'],round(r['run_seconds'],3),r['first_cont_seconds'],round(r['end_seconds'],3),round(r['running_seconds'],3))
