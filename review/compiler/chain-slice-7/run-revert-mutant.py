from pathlib import Path
import subprocess
p=Path('review/compiler/chain-slice-7')
with (p/'revert-mutant.log').open('w') as log:
 result=subprocess.run(['go','test','-overlay='+str(p/'revert-overlay.json'),'./internal/lower','-run','^TestEEPPresence','-v','-count=1','-timeout','80s'],stdout=log,stderr=subprocess.STDOUT,timeout=90)
s=(p/'revert-mutant.log').read_text()
assert result.returncode==1
assert all('--- FAIL: TestEEPPresence'+str(i) in s and 'got <nil>' in s for i in range(7))
assert '--- PASS: TestEEPPresenceSupported' in s
print('Revert mutant caught by all seven refusal tests; supported control passed')
