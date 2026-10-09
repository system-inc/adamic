#!/usr/bin/env python3
"""Plant a changed native input in one owned unit and test coverage/cache guards."""
import json,os,subprocess,time,signal
from pathlib import Path
root=Path(__file__).resolve().parents[3]
logs=Path(os.environ.get('ADAMIC_BRIDGE_MUTANT_LOGS','/tmp/test-split-bridge/mutants'));logs.mkdir(parents=True,exist_ok=True)
env={**os.environ,'GOMAXPROCS':'4','ADAMIC_GATE_UNCACHED':'1','ADAMIC_UNIT_BUDGET':'1'}
results=[]

def run(name,path,change,selector,expected,catcher,shard=None,success=False,budget=None):
 original=(root/path).read_text(); mutated=change(original); assert mutated!=original,name
 replacement=logs/(name+'.go'); replacement.write_text(mutated)
 overlay=logs/(name+'-overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/path):str(replacement)}}))
 command=['go','test','./bridge/tsgo','-json','-overlay='+str(overlay),'-run',selector,'-count=1','-timeout=90s']
 logfile=logs/(name+'.jsonl'); environment=env.copy()
 if shard is not None:environment['ADAMIC_TEST_SHARD']=shard
 if budget is not None:environment['ADAMIC_UNIT_BUDGET']=budget
 begun=time.monotonic()
 with logfile.open('w') as out:
  process=subprocess.Popen(command,cwd=root,env=environment,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
  try:process.wait(timeout=90)
  except subprocess.TimeoutExpired:
   os.killpg(process.pid,signal.SIGKILL);process.wait()
   (logs/'killed.json').write_text(json.dumps(dict(name=name,seconds=time.monotonic()-begun,killed=True,priority="P0"))+'\n')
   raise SystemExit('P0: killed at 90 seconds: '+name)
 elapsed=time.monotonic()-begun
 assert elapsed<60,('over budget: split smaller',name,elapsed)
 text=logfile.read_text(); events=[]
 for line in text.splitlines():
  try:events.append(json.loads(line))
  except json.JSONDecodeError:pass
 failed=sorted({e['Test'] for e in events if e['Action']=='fail' and e.get('Test')})
 passed=[e['Test'] for e in events if e['Action']=='pass' and e.get('Test')]
 skipped=[e['Test'] for e in events if e['Action']=='skip' and e.get('Test')]
 assert '[build failed]' not in text,(name,text[-5000:])
 assert process.returncode==(0 if success else 1) and failed==expected and catcher in text,(name,process.returncode,failed,text[-5000:])
 results.append(dict(seconds=elapsed,cooked=False,hard_limit_seconds=90,name=name,shard=shard,budget=environment.get("ADAMIC_UNIT_BUDGET"),command=command,exit=process.returncode,failed=failed,passed=len(passed),skipped=len(skipped),catcher=catcher,log=str(logfile)))
 (Path(os.environ.get('ADAMIC_BRIDGE_MERGE_LOGS','/tmp/bridge-b2-validation'))/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
 print(name,shard or '',':',failed or 'all selected units passed',flush=True)

def changed_query(s):
 needle='observed := r.answers("native-asan", r.product("native-asan"), arguments, false)'
 assert s.count(needle)==1
 return s.replace(needle,'''nativeArguments := append([]string{}, arguments...)
 if file == "sample.ts" {
  data, err := os.ReadFile(arguments[1]); if err != nil { r.t.Fatal(err) }
  manifest := filepath.Join(r.scratch, "planted.tsv")
  if err := os.WriteFile(manifest, bytes.Replace(data, []byte("\\t0\\n"), []byte("\\t14\\n"), 1), 0o644); err != nil { r.t.Fatal(err) }
  nativeArguments[1] = manifest
 }
 observed := r.answers("native-asan", r.product("native-asan"), nativeArguments, false)''')
for shard in range(4):
 owner=shard==3
 run('changed-native-query-'+str(shard),'bridge/tsgo/units_test.go',changed_query,'^TestBridge',['TestBridgeOracleSample'] if owner else [],'native oracle mismatch' if owner else 'PASS',str(shard)+'/4',not owner)

run('missing-piece','bridge/tsgo/registered_units_test.go',lambda s:s.replace('\t{"TestBridgeABI", "abi", "", 0, bridgeABI},\n',''),'TestBridgeUnitsCoverEveryPiece',['TestBridgeUnitsCoverEveryPiece'],'bridge coverage count')
run('missing-query','bridge/tsgo/units_test.go',lambda s:s.replace('positions := make([]int, count)','positions := make([]int, count-1)'),'TestBridgeUnitsCoverEveryPiece',['TestBridgeUnitsCoverEveryPiece'],'query coverage count')
run('drop-shard-filter','bridge/tsgo/units_test.go',lambda s:s.replace('&& index%count == shard','&& count > 0 && shard >= 0'),'TestBridgeUnitsCoverEveryPiece',['TestBridgeUnitsCoverEveryPiece'],'shard coverage')
run('accept-corrupt-product','bridge/tsgo/product_cache_test.go',lambda s:s.replace('if hashes[name] != fmt.Sprintf','if false && hashes[name] != fmt.Sprintf'),'TestBridgeProductCacheIsVerified',['TestBridgeProductCacheIsVerified'],'corrupt product was accepted')
run('accept-extra-product','bridge/tsgo/product_cache_test.go',lambda s:s.replace('if len(hashes) != len(bridgeProductNames)','if false && len(hashes) != len(bridgeProductNames)'),'TestBridgeProductCacheIsVerified',['TestBridgeProductCacheIsVerified'],'wrong product count was accepted')
run('rebuild-product','bridge/tsgo/product_cache_test.go',lambda s:s.replace('Flags: []string{key}}, build)\n', 'Flags: []string{key}}, build)\n\tif err == nil { err = build(directory) }\n',1),'TestBridgeProductCacheIsVerified',['TestBridgeProductCacheIsVerified'],'product was rebuilt')
run('over-budget','bridge/tsgo/units_test.go',lambda s:s.replace('begun := time.Now()','begun := time.Now().Add(-31*time.Second)'),'TestBridgeABI',['TestBridgeABI'],'bridge unit exceeded 30 seconds')
run('over-budget-without-budget-gate','bridge/tsgo/units_test.go',lambda s:s.replace('begun := time.Now()','begun := time.Now().Add(-31*time.Second)'),'TestBridgeABI',[],'over the 30-second budget',success=True,budget='0')
run('poison-build-callback','bridge/tsgo/products_test.go',lambda s:s.replace('if err := buildBridgeProducts(repository, directory); err != nil {','if err := fmt.Errorf("shared products rebuilt after fetch"); err != nil {'),'^TestBridge',[],'PASS',success=True)
