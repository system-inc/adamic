import pathlib,json,subprocess,os,time,signal,re,shutil
p=pathlib.Path('/tmp/css-delete');items=[e for e in json.loads((p/'mutant-list.json').read_text()) if e.get('original_id','')=='B3'];candidates=['TestCSSPrinterBoundaryProofs','TestCSSPrinterThroughput','TestOptionalBooleanPrinterMatchesGo','TestSharedSliceAppendAgreesWithNode']
witnesses={'TestCSSPrinterShardingCatchesDisagreement','TestCSSParserPlantedDisagreement','TestComposedMemoryChecksCanFail','TestTheCanonicalRangeChecksCanFail'}
printer_count=int(re.search(r'const testCSSPrinterAgreesWithGoShards = (\d+)',pathlib.Path('stage1/cohere/css/css_printer_30s_test.go').read_text())[1])
def witness(n):
 m=re.fullmatch(r'TestCSSPrinterAgreesWithGo_(\d+)',n)
 if m and int(m[1])>=printer_count//4:return True
 return n in witnesses or ('Mutant' in n and n.startswith('TestProduct_')) or bool(re.search(r'(Mutants|MutantKilled|PlantedFailure|PlantedDisagreement)$',n))
def stop(proc):
 # Stop only this invocation and its descendants, including childguard compiler groups.
 pairs=[tuple(map(int,line.split())) for line in subprocess.check_output(['ps','-e','-o','pid=,ppid='],text=True).splitlines()];desc={proc.pid}
 while True:
  more={pid for pid,ppid in pairs if ppid in desc};new=desc|more
  if new==desc:break
  desc=new
 for pid in sorted(desc,reverse=True):
  try:os.kill(pid,signal.SIGTERM)
  except ProcessLookupError:pass
 try:proc.wait(timeout=5)
 except subprocess.TimeoutExpired:
  for pid in sorted(desc,reverse=True):
   try:os.kill(pid,signal.SIGKILL)
   except ProcessLookupError:pass
  proc.wait()
def run(id,extra):
 cache=p/'cache'/id;assert not cache.exists(),cache
 env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=str(cache))
 skip='^('+ '|'.join(candidates+extra)+')($|_|/)';args=['go','test','-json','-count=1','-timeout','30m','./stage1/cohere/css/','-skip',skip]
 logpath=p/(id+'.log');start=time.monotonic();ev=[];offset=0;partial='';caught=[];panics=[];last='';stopped=False
 with logpath.open('w') as log:
  proc=subprocess.Popen(args,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  while True:
   with logpath.open() as rd:rd.seek(offset);chunk=rd.read();offset=rd.tell()
   lines=(partial+chunk).split('\n');partial=lines.pop()
   for line in lines:
    try:e=json.loads(line)
    except:continue
    ev.append(e);top=e.get('Test','').split('/')[0]
    if top:last=top
    text=e.get('Output','')
    if text.startswith('panic:') or text.startswith('fatal error:'):
     panics.append(top or last)
    if e['Action']=='fail' and top and '/'not in e.get('Test','') and not witness(top) and not panics:caught.append(top)
   if caught and proc.poll() is None:
    stop(proc);stopped=True;break
   if proc.poll() is not None:break
   time.sleep(.2)
  proc.wait()
 # Parse the complete flushed log to preserve every observed failure.
 ev=[]
 for line in logpath.read_text().splitlines():
  try:ev.append(json.loads(line))
  except:pass
 failed=sorted({e['Test'] for e in ev if e['Action']=='fail' and 'Test'in e and '/'not in e['Test']})
 pkg=next((e for e in reversed(ev) if 'Test'not in e and e['Action']in ['pass','fail']),{})
 result={'id':id,'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+str(cache)+' '+ ' '.join(args)+' > '+str(logpath)+' 2>&1','wall_seconds':round(time.monotonic()-start,3),'exit':proc.returncode,'stopped_after_clean_catch':stopped,'failed':failed,'witness_failures':[n for n in failed if witness(n)],'panicking_tests':sorted(set(panics)),'passed':sorted({e['Test'] for e in ev if e['Action']=='pass' and 'Test'in e and '/'not in e['Test']}),'skipped':sorted({e['Test'] for e in ev if e['Action']=='skip' and 'Test'in e and '/'not in e['Test']}),'completed_green':pkg.get('Action')=='pass','failing_lines':[e.get('Output','').strip() for e in ev if e.get('Test','').split('/')[0] in failed and e['Action']=='output' and re.search(r'\.go:\d+:',e.get('Output',''))]}
 result['still_caught_by']=[n for n in failed if not witness(n) and n not in result['panicking_tests']]
 (p/(id+'.run.json')).write_text(json.dumps(result,indent=2)+'\n');shutil.rmtree(cache,ignore_errors=True)
 return result
assert (p/'baseline.exit').read_text().strip()=='0','baseline is not green'
shutil.rmtree(p/'cache/baseline',ignore_errors=True)
results=json.loads((p/'matrix.json').read_text())
for e in items:
 id=e['mutant'];print('START',id,flush=True)
 if e['stale']:results.append(dict(e,runs=[]));continue
 diff=p/(id+'.diff');subprocess.run(['git','apply',str(diff)],check=True)
 runs=[]
 try:
  extra=[]
  while True:
   r=run(id+('-skip-panic-'+str(len(extra)) if extra else ''),extra);runs.append(r)
   if r['panicking_tests']:
    new=[n for n in r['panicking_tests'] if n and n not in extra]
    if not new:break
    extra.extend(new);continue
   break
 finally:subprocess.run(['git','apply','-R',str(diff)],check=True)
 row=dict(e,runs=runs,still_caught_by=sorted({n for r in runs if not r['panicking_tests'] for n in r['still_caught_by']}),panicking_tests=sorted({n for r in runs for n in r['panicking_tests']}),witness_failures=sorted({n for r in runs for n in r['witness_failures']}),completed_green=any(r['completed_green'] for r in runs))
 results.append(row);(p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print('DONE',id,row['still_caught_by'],'green',row['completed_green'],flush=True)
print('REPLAY FINISHED',flush=True)
