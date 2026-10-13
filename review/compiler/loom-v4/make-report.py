import json,re,subprocess
from pathlib import Path
e=Path('review/compiler/loom-v4'); latest={};attempts=[]
for file in ['results.jsonl','remaining-results.jsonl','fixed-results.jsonl']:
 for line in (e/file).read_text().splitlines():
  row=json.loads(line)
  if 'test' in row:
   row['measurement']='process wall';attempts.append(row);latest[(row['package'],row['test'])]=row
counts=[]
for file in ['regexp-runtime-final.jsonl','restored-controls.jsonl','witness-controls-final.jsonl','whitespace-controls.jsonl','counts-verification.jsonl']:
 for line in (e/file).read_text().splitlines():
  try:row=json.loads(line)
  except json.JSONDecodeError:continue
  if row.get('Action') in ['pass','fail','skip'] and 'Test' in row:
   item=dict(package=row['Package'].removeprefix('github.com/system-inc/adamic/'),test=row['Test'],seconds=row.get('Elapsed',0),exit=int(row['Action']=='fail'),grain=row.get('Elapsed',0)<60,measurement='Go test leaf',status=row['Action'])
   attempts.append(item)
   if file=='counts-verification.jsonl':counts.append(item)
   else:latest[(item['package'],item['test'])]=item
for row in latest.values():
 if row['measurement']=='process wall':
  prefix=row['package'].replace('/','-')+'-'+row['test']
  logs=[e/(prefix+'.log'),e/(prefix+'-remaining.log'),e/(row['test']+'-fixed.log')]
  for log in logs:
   if log.exists() and '--- SKIP: '+row['test'] in log.read_text(errors='replace'):row['status']='skip'
assert all(r['exit']==0 and r['grain'] for r in latest.values())
assert all(r['exit']==0 and r['grain'] for r in counts)
(e/'final-results.json').write_text(json.dumps(dict(focused=list(latest.values()),counts=counts,attempts=attempts),indent=2)+'\n')
mutants=[json.loads(x) for x in (e/'source-mutants.jsonl').read_text().splitlines() if 'blocked' not in json.loads(x)]+[json.loads((e/'javascript-order-fixed-results.json').read_text())]+[json.loads(x) for x in (e/'witness-mutants.jsonl').read_text().splitlines()]
assert len(mutants)==17 and all(x['killed'] for x in mutants)
catchers={'surface-borrow':'ASan heap-use-after-free in adapter root access','bind-name-borrow':'ASan heap-use-after-free in string output','identity-root':'Node stdout comparison, clean exit 0','metadata-inline':'2,885 inline stores violate the zero-store bound','metadata-omitted':'zero runtime calls violate the 2,885-certificate assertion','native-order':'argument output missing; premature exit 70','javascript-order':'argument output missing; premature exit 70','unsupported-domain':'recursive domain refusal missing','unsupported-producer':'recursive producer refusal missing'}
for m in mutants:m['catcher']=m.get('catcher',catchers.get(m['mutant'],'missing escaping-read refusal after clean Node-equivalent execution'))
(e/'mutants-final.json').write_text(json.dumps(mutants,indent=2)+'\n')
observations=[]
for p in sorted(e.glob('*.log')):
 if 'mutant' in p.name:continue
 for line in p.read_text(errors='replace').splitlines():
  if re.search(r'caught|omission|removed .*check|mutant',line,re.I) and len(line)<1800:observations.append(p.name+': '+line)
(e/'mutant-observations.txt').write_text('\n'.join(observations)+'\n')
moves=json.loads((e/'counts-diff.json').read_text());runtime=subprocess.check_output(['git','diff','--name-only','origin/main','--','internal/native/runtime'],text=True).splitlines();base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip()
lines=[
'Built: V4 checked callables with the source branch\'s V3 arrays, squashed for Loom.',
'Commits: source 8a4c135e4f3cf611d2fc7801f4e640d7804af7f9; initial main cd7880cf5145b33a5bf6bb2227e216d7fb9d63c0; landing base '+base+'.',
f'Commands and outputs: {sum(r.get("status")!="skip" for r in latest.values())} focused top-level checks passed, seven carried skips; counts and changed-package vet green.',
'Mutants: 17 independently restored source/witness mutants caught; embedded IR and private-runtime mutants passed their behavioral catchers.',
'Not covered: the full gate and seven carried pending tests. The 30-minute first-green push target was missed.',
'',
'The final commit is the pushed compiler/loom-v4 branch tip, one commit above its landing base. Main advanced only under stage3/census/tsc-closure during validation; its compiler, runtime and counts table did not change. The branch is rebased once onto that refreshed tip before pushing. No PR or Loom submission is made.',
'',
'Code whitespace checks exclude raw review logs and .patch evidence: unified-diff context indentation and sanitizer output are retained verbatim. Two inherited trailing spaces in native test C strings were removed; the affected native probes, reader guard and native/oracle vet passed again. gofmt -l on all changed Go files was empty.',
'',
'The three squash conflicts were resolved once:',
'',
'- internal/flow/trace_test.go: use V3 sparse-array snapshots (length and present keys), retaining main\'s own-property descriptor check so object getters are never called by tracing.',
'- internal/ir/ir.go: retain main\'s SortedMap contract registry and newer fields; include all V3 array/V4 adapter metadata. Use SortedMap for V4\'s new optional-field map too, preserving deterministic gob products.',
'- internal/oracle/checked_views_v2_counts_test.go: retain main\'s per-row countCase registration, omitting the three V4 compile-time refusal rows. Existing V4 refusal readers preserve their Node controls. Port the V3 array count registrations to the same per-row form.',
'',
'Test grain and products: split the large new array-read, array-mutant, array-union and native callable tables into top-level functions. Every executed focused leaf, product and count leaf is under 60 seconds on cpu.max=400000 100000 (4 CPUs); nproc=5. Process wall times include startup and helper setup. Go JSON timings are labeled as leaf elapsed times. Native runtime products are run ahead and their build times are recorded separately. Test-binary compilation is recorded in the build rows, not reported as a test leaf.',
'',
'V4\'s ASan plus counted runs were absent from main\'s runtimeOptions. Declare that mode and three setup products in checked_views_v4_runtime_products_test.go. Also declare the receiver plus RegExp callback feature combination. Main\'s runtime-set check accepts a source-profile witness callback in addition to a persistent file, and still checks the actual emitted features exactly. The same source constant is used by that witness and the RegExp callback oracle. No other package\'s tests were edited to make room. No new Go package consumes the compiler, so cloud/fast-gate/compiler-dependencies.json needs no entry. No new product or test stores a fixed machine/build directory.',
'',
'Conservative assumptions: the exact anchored legacy count name no longer exists on main. Use ^TestCountsAreRecorded to select main\'s generated shards. Existing upstream TypeScript checker-profile witnesses remain references, not new Adamic admissions; new Adamic programs are .a. No cohere implementation was copied.',
'',
'Counts were written once. The requested anchored command ran no tests (5.561 s) and refused to write. The first sharded attempt failed because @types/node 25.3.3 was missing (52.428 s), also without writing. npm ci --ignore-scripts --prefix stage3/api installed the locked dependencies (0.738 s). The single successful update took 57.221 s; independent verification without -update-counts took 45.066 s. All 30 changed rows are below and in counts-diff.json. Twenty-six rows add the carried V3 sparse-array and checked-array fixtures. Three rows disappear because V4 now refuses their escapes before emission. The only changed existing numeric row is namespace_method_receiver.a: 7/7/9/16/5/0 becomes 7/7/7/14/5/0. Attribution from the lowering diff: an explicit this parameter is a declaration contract and is excluded from runtime argument slots; the fixture makes two method calls. This attribution is a code reading, not a separate counterfactual measurement.',
'',
'| Counts row | Before | After | Attribution |','|---|---|---|---|']
for m in moves:
 def nums(row):return 'absent' if row is None else '/'.join(x.strip() for x in row.split('|')[2:-1])
 cause='new V3 fixture' if m['before'] is None else ('V4 compile-time escape refusal' if m['after'] is None else 'declaration-only this excluded from runtime arguments')
 lines.append(f"| {m['fixture']} | {nums(m['before'])} | {nums(m['after'])} | {cause} |")
lines+=['','Runtime files changed against landing main:','']+['- '+p for p in runtime]
lines+=['','Exact validation commands are in run-focused.py, run-remaining.py, run-fixed.py and run-source-mutants.py. No whole package test run or full gate was used. Focused tests were compiled with go test -c, then selected individually with -test.run ^NAME$, -test.count=1, -test.timeout=90s and a 95-second outer kill. The remaining pass used four independent processes with -test.parallel=1; completed tests were excluded. One in-flight serial test was interrupted during the switch and is recorded in focused.log. It was subsequently completed; no test over 60 seconds was rerun. Registry failures were rerun after their declarations were fixed. All original attempts remain in the JSONL/logs.','',
'Final restored controls: timeout 120 go test ./internal/ir ./internal/oracle ./stage1/cohere/json -run <exact source-mutant catcher names and TestCallTargetReaders/TestProduct_CallTargetExports> -count=1 -timeout 90s -parallel 4 -json, 10.034 s wall. Final go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/oracle ./stage1/cohere/json passed with no diagnostics, 2.396 s. The earlier vet caught missing header setup in two split native probe helpers; it was restored before the focused native tests ran. /usr/bin/time was unavailable, so Bash time was used.','',
'Every focused top-level test\'s final seconds follow. Additional subtest and count-leaf seconds are preserved in final-results.json and counts-verification.jsonl. Positive fixtures are held to source Node, release native, ASan/UBSan and backend Node; finished programs also run the leak checks where the source tests request them. Deliberate checked-view negatives pin exit 70 and exact output/diagnostics rather than claiming equality to unchecked Node. Original array checker witnesses are Node reference controls.','',
'| Package | Test | Seconds | Measurement | Result |','|---|---|---:|---|---|']
for row in sorted(latest.values(),key=lambda r:(r['package'],r['test'])):lines.append(f"| {row['package']} | {row['test']} | {row['seconds']:.3f} | {row['measurement']} | {row.get('status','pass')} |")
lines+=['','Independent source and witness mutant catchers:','','| Mutant | Catcher test | Seconds | Observation |','|---|---|---:|---|']
for m in mutants:lines.append(f"| {m['mutant']} | {m['test']} | {m['seconds']:.3f} | exit {m['exit']}; {m['catcher']} |")
lines+=['','All 17 mutants restored exact source bytes. None is counted as a build-warning kill. The original JavaScript-order script\'s mutation seam had moved; its blocked attempt is retained, and javascript-order-fixed.patch targets the current direct-call emitter. Source patches are .patch, never compilable Go. Embedded mutant observations, including native and JavaScript argument/result, aggregate, array-presence/storage/literal/hole, identity/surface, and adapter-cache controls, are collected in mutant-observations.txt with the full per-test logs beside it. Runtime adapter mutants include skip-intern (allocation bound), skip-recheck (winner selection), dying-retain (ASan use-after-free), release-before-remove (lifecycle order), canonical-factory (required exit 70), strong-leak-root (missing required LSan report), and identity-overread (ASan global-buffer-overflow).','',
'Carried pending coverage: TestViewAdapterRuntimeStaticsPending, TestViewAdapterPoolTSanPending, TestViewAdapterRegionAdoptionPending, TestViewAdapterProgramUnderlyingPending, TestV4EscapeAdapterBlameRuntimeCallback, TestV4Lane5FactoryWiderWritePending and TestV4Lane5FactoryReadonlyTargetControl. No whole-gate, ThreadSanitizer pool, region adoption or pending higher-order control is claimed.','',
'Setup: initial cloud/setup.sh was stopped with exit 143 while Git\'s nested historical lookup fetched over 550 MB. An explicit depth-1, blob-filtered fetch of d92d9bfee114c80be2c375d72edae966176e3a4f and detached checkout repaired the submodule setup; setup was rerun successfully. GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh were used. Setup timing lines:', '', '```text']
lines += [x for x in (e/'setup-retry.log').read_text().splitlines() if x.startswith('setup:')]
lines+=['```','']
(e/'REPORT.md').write_text('\n'.join(lines));print('focused',len(latest),'counts leaves',len(counts),'mutants',len(mutants),'longest focused',max(x['seconds'] for x in latest.values()))
