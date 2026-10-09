from pathlib import Path
import subprocess
paths=[p for p in Path('internal/oracle/testdata/review/agree').glob('fxspptb_oct9_native_*.pending') if any('_p'+n+'_' in p.name for n in ['17','26','33','34','35'])]
originals={p:p.read_bytes() for p in paths}
try:
 for p in paths:p.unlink()
 with Path('review/compiler/fx6-candidates-4/named-146.log').open('w') as log:r=subprocess.run(['go','test','./internal/oracle','-run','^TestReviewProgramsAgreeWithNode$/fxspptb_oct9_native_p(17|18|20|26|33|34|35)_','-v','-count=1','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 print('named-146 exit',r.returncode,flush=True)
finally:
 for p,s in originals.items():p.write_bytes(s)
