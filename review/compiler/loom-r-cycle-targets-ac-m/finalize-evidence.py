import pathlib,json
p=pathlib.Path('review/compiler/loom-r-cycle-targets-ac-m')
initial=json.loads((p/'final-package-results.json').read_text())
remaining=json.loads((p/'remaining-package-results.json').read_text())
native=json.loads((p/'native-wasi-ready-result.json').read_text())
rows=initial[:2]+[native]+remaining
expected=['internal/ir','internal/lower','internal/native','internal/javascript','internal/flow','internal/fresh','internal/oracle']
assert [r['package']for r in rows]==expected
assert all(r['exit']==0 and not r['over_budget']for r in rows)
for n in ['a-check','vet','readers','guards','focused','counts-update','effects-final','mutants','admission-mutant']:
 assert(p/(n+'.exit.log')).read_text().strip()=='0',n
assert not(p/'gofmt.log').read_text()
assert not(p/'diff-check.log').read_text()
all_events=[]
for name in ['final-package-00.log','final-package-01.log','native-wasi-ready.log','final-package-03.log','final-package-04.log','final-package-05.log','final-package-06.log','effects-final.log']:
 for line in(p/name).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Test')and e.get('Action')in('pass','fail','skip'):all_events.append(dict(log=name,package=e.get('Package'),test=e['Test'],status=e['Action'],seconds=e.get('Elapsed',0)))
assert not any(e['status']=='fail'for e in all_events)
names=json.loads((p/'changed-test-functions.json').read_text())
changed=[e for e in all_events if e['test']in names or e['test']=='TestClosureTargetsBoundOnlyProvenValues']
assert set(names)<=set(e['test']for e in changed)
(p/'all-final-test-seconds.json').write_text(json.dumps(all_events,indent=2)+'\n')
(p/'changed-test-seconds.json').write_text(json.dumps(changed,indent=2)+'\n')
(p/'verification-summary.json').write_text(json.dumps(dict(source='b1d42ea3917096b83841c570f05185cbe9a0b337',base='67733635b0252eb9d3bb3645c72262c7b18361e1',packages=rows,all_pass=True,runtime_additional_lines=[],acheck=dict(checked=0,NotYet=0,refused=3,type_error=0),mutant_types_caught=9),indent=2)+'\n')
with(p/'REPORT.md').open('a')as f:
 f.write('Final complete package results\n\n| Package | Package seconds | Slowest leaf seconds | Result |\n|---|---:|---:|---|\n')
 for r in rows:
  seconds=r.get('seconds',r.get('packages',[{}])[-1].get('Elapsed'))
  f.write(f"| {r['package']} | {seconds} | {r['max_leaf_seconds']} | PASS |\n")
 f.write('\nEvery completed final leaf is under 60 seconds. all-final-test-seconds.json records all terminal test events. New/changed declared test timings from complete final suites (the explicit capture-layout refinement is also checked separately in effects-final.log):\n\n| Test | Seconds |\n|---|---:|\n')
 for name in names:
  timings=[e['seconds']for e in changed if e['test']==name]
  f.write(f"| {name} | {max(timings)} |\n")
 f.write('\nDelivery is one commit with parent 67733635b0252eb9d3bb3645c72262c7b18361e1, retaining the complete original message plus the requested re-cut line, conservative assumption and Ahra co-author trailer. Only compiler/loom-r-cycle-targets-ac-m is pushed, without force. The final response supplies the full delivery SHA and remote verification. No ahra CLI, PR or Loom submission is run.\n')
print(json.dumps(rows,indent=2))
