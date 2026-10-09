import json
import pathlib
import subprocess
import time

root = pathlib.Path(__file__).resolve().parents[3]
evidence = root / 'review/compiler/statics-guards'
cases = [
 ('M06', '^TestPrivateAndPublicStaticsAgreeWithNode$'),
 ('M13', '^TestReadinessErrorIncludesReceiverExpression$'),
 ('M12', '^Test(NamespaceFactoryBindingsAgreeWithNode|ParserFactoryBindingHoisting)$'),
 ('u030-P1', 'Test(ClassWrongOutputKeysRepair|ClassWrongOutputPrivateRepair|DefiniteAssignmentSoundNeighbors)$'),
 ('u030-P2', 'TestClockGenericReturnsT01Rejects(Null|Index)BeforeBody$'),
 ('u037-P01', '^TestModuleNamespaceInitializedReadProof$'),
 ('u037-P02', '^TestNamespaceAmbientHostInitialization$'),
 ('u037-P05', '^TestParserFactoryBindingHoisting$'),
 ('u037-P06', '^TestParserFactoryBindingHoisting$'),
 ('u037-P08', '^TestParserFactoryBindingHoisting$'),
]
results = []
for name, selector in cases:
 patch = evidence / (name + '.diff')
 with (evidence / (name + '-apply.log')).open('w') as log:
  subprocess.run(['git', 'apply', '--check', '--verbose', str(patch)], cwd=root, stdout=log, stderr=log, check=True, timeout=10)
  subprocess.run(['git', 'apply', str(patch)], cwd=root, check=True, timeout=10)
 command = ['go', 'test', '-json', './internal/lower', '-run', selector, '-count=1', '-timeout', '90s']
 started = time.monotonic()
 try:
  with (evidence / (name + '.jsonl')).open('w') as log:
   result = subprocess.run(command, cwd=root, stdout=log, stderr=log, timeout=100)
  events = []
  for line in (evidence / (name + '.jsonl')).read_text().splitlines():
   try: event = json.loads(line)
   except ValueError: continue
   if event.get('Action') == 'fail' and event.get('Test'): events.append(event['Test'])
  results.append(dict(probe=name, command=command, exit=result.returncode, seconds=round(time.monotonic()-started,3), failed_tests=events))
  print(results[-1], flush=True)
 finally:
  subprocess.run(['git', 'apply', '-R', str(patch)], cwd=root, check=True, timeout=10)
(evidence / 'probe-results.json').write_text(json.dumps(results, indent=2) + '\n')
if any(row['exit'] != 1 or not row['failed_tests'] for row in results): raise SystemExit('probe did not fail by a test assertion')
