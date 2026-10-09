import pathlib,json,re,subprocess,time,os
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-css-composition_shards';files=['composition_shards_test.go','css_printer_parallel_test.go','css_test.go','gaps_test.go','memory_checks_test.go','optimized_test.go','parser_shards_test.go'];tests=[t for f in files for t in re.findall(r'^func (Test\w+)\(', (r/'stage1/cohere/css'/f).read_text(),re.M)];assert len(tests)==95;listing=(p/'list.log').read_text().splitlines();assert all(t in listing for t in tests);(p/'requested-tests.json').write_text(json.dumps(tests,indent=2))
comp=[t for t in tests if re.match(r'TestCompositionMatchesGo_\d+$',t)];printer=[t for t in tests if re.match(r'TestCSSPrinterAgreesWithGo_\d+$',t)];setup=['TestCompositionMatchesGo_Setup','TestCompositionMatchesGo'];groups={'TestCompositionMatchesGo setup family':setup,'TestCompositionMatchesGo family':comp,'TestCSSPrinterAgreesWithGo family':printer};used=comp+printer+setup;groups.update({t:[t] for t in tests if t not in used});(p/'row-members.json').write_text(json.dumps(groups,indent=2));runs=[]
def run(ts,log):
 start=time.monotonic();cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^('+'|'.join(ts)+')$']
 with (p/log).open('w') as f:q=subprocess.run(cmd,cwd=r,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));(p/'clean-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(runs[-1]['wall'],3),flush=True)
 events=[]
 for l in (p/log).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 realfail=any(x.get('Action')=='fail' and x.get('Test') for x in events);cook=any('test timed out' in x.get('Output','') for x in events) or q.returncode==124
 if realfail and not cook:raise RuntimeError('RED clean baseline: '+log)
 if q.returncode!=0 and not cook:raise RuntimeError('RED/build failure: '+log)
 return not cook
# Complete a clean narrowed baseline across the unit before any mutations.
for row,ts in groups.items():
 ok=run(ts,'baseline-'+row.replace(' ','_')+'.log')
 if not ok:
  (p/'cooked-rows.json').write_text(json.dumps([row],indent=2));print('COOKED row '+row,flush=True)
# The three measurements are separate clean -count=1 invocations.
for row,ts in groups.items():
 for i in range(1,4):run(ts,'timing-'+row.replace(' ','_')+'-'+str(i)+'.log')
