from pathlib import Path
import subprocess
source=Path('/workspace/lint-cost-split.py').read_text()
old='for e in events:\n g=guarded(e)';assert source.count(old)==1
Path('/workspace/lint-cost-split-mutant.py').write_text(source.replace(old,'for e in events[:-1]:\n g=guarded(e)'))
with Path('/tmp/lint-cost-partition-mutant.log').open('wb') as out:
 r=subprocess.run(['python3','/workspace/lint-cost-split-mutant.py'],stdout=out,stderr=out)
assert r.returncode!=0
log=Path('/tmp/lint-cost-partition-mutant.log').read_text();assert 'assert sum(counts.values())==len(events)' in log and 'AssertionError' in log
print('partition mutant caught by complete sample-count assertion')
