"""Change only the walk retain row, require the counts assertion, then restore."""
from pathlib import Path
import subprocess, sys
root=Path(sys.argv[1]).resolve()
path=root/'internal/oracle/testdata/counts/runtime/internal/oracle/testdata/borrow_chain_walk.a.counts'
original=path.read_text()
rows=[line for line in original.splitlines() if line.startswith('| internal/oracle/testdata/borrow_chain_walk.a |')]
assert len(rows)==1
before=rows[0]
cells=before.split('|');cells[4]=' '+str(int(cells[4])+1)+' '
after='|'.join(cells)
try:
 path.write_text(original.replace(before,after))
 log=Path('/tmp/borrow-chains-mutant-counts.log')
 with log.open('wb') as output:
  result=subprocess.run(['go','test','./internal/oracle','-run','^TestCountsAreRecorded_/internal/oracle/testdata/borrow_chain_walk','-count=1','-timeout','90s'],cwd=root,stdout=output,stderr=subprocess.STDOUT,timeout=120)
 text=log.read_text();print(text)
 assert result.returncode!=0 and 'recorded: '+after in text and 'measured: '+before in text
finally:
 path.write_text(original)
