"""Run test overlays; failures must be owned by the selected bridge check."""
import json, os, subprocess
from pathlib import Path
root = Path(__file__).resolve().parents[4]
out = root/'review/compiler/test-split-bridge-main'
work = Path('/workspace/bridge-main-mutants'); work.mkdir(exist_ok=True)
env = dict(os.environ, GOMAXPROCS='4', ADAMIC_UNIT_BUDGET='1', ADAMIC_BUILD_CACHE_DIR='/workspace/bridge-main-products')
env.pop('ADAMIC_TSGO_CORPUS',None); env.pop('ADAMIC_TEST_SHARD',None)
results=[]
def run(label,file,old,new,test,reason,extra=None,expected=None):
 source=root/'bridge/tsgo'/file
 text=source.read_text(); assert text.count(old)==1,(label,text.count(old))
 replacement=work/(label+'.go'); replacement.write_text(text.replace(old,new))
 overlay=work/(label+'.json'); overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
 command=['go','test','-overlay',str(overlay),'./bridge/tsgo','-run',test,'-count=1','-json','-timeout=90s']
 e=env.copy(); e.update(extra or {})
 log=out/(label+'.jsonl')
 with log.open('w') as stream: process=subprocess.run(command,cwd=root,env=e,stdout=stream,stderr=subprocess.STDOUT)
 events=[]
 for line in log.read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError: pass
 failed=[x['Test'] for x in events if x.get('Action')=='fail' and x.get('Test')]
 assert failed==expected if expected is not None else bool(failed),(label,failed)
 assert reason in log.read_text(),(label,reason)
 results.append(dict(mutant=label,command=command,exit=process.returncode,failed_tests=failed,reason=reason))
 (out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
 print(label,failed,flush=True)
run('missing-case','registered_units_test.go','\t{"TestBridgeABI", "abi", "", 0, bridgeABI},\n','','^TestBridgeUnitsCoverEveryPiece$','bridge coverage count')
run('missing-position','units_test.go','positions := make([]int, count)','positions := make([]int, count-1)','^TestBridgeUnitsCoverEveryPiece$','query coverage count')
run('duplicate-owners','units_test.go','&& index%count == shard','&& count > 0 && shard >= 0','^TestBridgeUnitsCoverEveryPiece$','shard coverage')
run('ignore-digest','product_cache_test.go','if hashes[name] != fmt.Sprintf','if false && hashes[name] != fmt.Sprintf','^TestBridgeProductCacheIsVerified$','corrupt product was accepted')
run('ignore-product-count','product_cache_test.go','if len(hashes) != len(bridgeProductNames)','if false && len(hashes) != len(bridgeProductNames)','^TestBridgeProductCacheIsVerified$','wrong product count was accepted')
run('rebuild-product','product_cache_test.go','err = bridgeProductCheck(directory)','err = build(directory)\n\t\tif err == nil { err = bridgeProductCheck(directory) }','^TestBridgeProductCacheIsVerified$','product was rebuilt')
run('budget-enforced','units_test.go','begun := time.Now()','begun := time.Now().Add(-61*time.Second)','^TestBridgeABI$','bridge unit exceeded 60 seconds',expected=['TestBridgeABI'])
run('budget-log-only','units_test.go','begun := time.Now()','begun := time.Now().Add(-61*time.Second)','^TestBridgeABI$','over the 60-second budget',extra={'ADAMIC_UNIT_BUDGET':'0'},expected=[])
old='observed := r.answers("native-asan", r.product("native-asan"), arguments, false)'
new='''nativeArguments := append([]string(nil), arguments...)
 if file == "sample.ts" {
 data, err := os.ReadFile(arguments[1]); if err != nil { r.t.Fatal(err) }
 path := filepath.Join(r.scratch, "planted.tsv")
 data = bytes.Replace(data, []byte("\\t0\\n"), []byte("\\t14\\n"), 1)
 if err := os.WriteFile(path, data, 0600); err != nil { r.t.Fatal(err) }
 nativeArguments[1] = path
 }
 observed := r.answers("native-asan", r.product("native-asan"), nativeArguments, false)'''
for shard in range(4):
 run('planted-shard-'+str(shard),'units_test.go',old,new,'^TestBridge(OracleSample|TimingRound1Sample)$','native oracle mismatch' if shard==3 else ('PASS' if shard==0 else 'skip'),extra={'ADAMIC_TEST_SHARD':str(shard)+'/4'},expected=['TestBridgeOracleSample'] if shard==3 else [])
