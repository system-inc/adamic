"""Change only the walk retain row, require the counts assertion, then restore."""
from pathlib import Path
import subprocess, sys
root=Path(sys.argv[1]).resolve()
path=root/'internal/oracle/counts.md'
original=path.read_text()
rows=[line for line in original.splitlines() if line.startswith('| internal/oracle/testdata/borrow_chain_walk.a |')]
assert len(rows)==1
before=rows[0]
cells=before.split('|');assert cells[4].strip()=='13';cells[4]=' 14 '
after='|'.join(cells)
try:
 path.write_text(original.replace(before,after))
 log=Path('/tmp/borrow-chains-mutant-counts.log')
 with log.open('wb') as output:
  result=subprocess.run(['go','test','./internal/oracle','-run','^TestCountsAreRecorded$/fixtures/internal/oracle/testdata/borrow_chain_walk','-count=1','-timeout','30m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
 text=log.read_text();print(text)
 assert result.returncode!=0 and 'recorded: '+after in text and 'measured: '+before in text
finally:
 path.write_text(original)
