import subprocess,time,json,pathlib
root=pathlib.Path('/tmp/u059')
rows=['TestTypeScriptConditionsAgreeWithNode','TestConditionNaNMutant','TestConditionLedgerMissingSiteMutant','TestConditionLedgerNullMutant','TestLoopCountersAgreeWithNode','TestCountsAreRecorded','TestDebuggerNativeEmitsNothing','TestElementAccessCompilerAreaBoundaries','TestEntriesAcceptance','TestEntriesProvenance','TestEntriesRuntimeReadiness','TestEnumInitializationNode','TestEnumInitializationUnknownPinned']
(root/'rows.json').write_text(json.dumps(rows))
while not (root/'slice-baseline-exit').exists(): time.sleep(1)
if (root/'slice-baseline-exit').read_text().strip()!='0': raise SystemExit('slice baseline did not pass')
for row in rows:
 for n in range(1,4):
  with (root/f'time-{row}-{n}.log').open('w') as log:
   start=time.monotonic(); p=subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$'],stdout=log,stderr=subprocess.STDOUT)
  (root/f'time-{row}-{n}.meta').write_text(json.dumps({'wall':time.monotonic()-start,'exit':p.returncode}))
  if p.returncode: break
(root/'timings-done').write_text('done')
