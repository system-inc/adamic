import sys
sys.dont_write_bytecode = True
import sys
from pathlib import Path
from importlib.machinery import SourceFileLoader
here=Path(__file__).resolve().parent
s=SourceFileLoader('supervise',str(here/'clear-supervise.py')).load_module()
label=sys.argv[1];tree=s.scratch/('baseline' if label=='baseline' else 'before')
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
r=s.run('types-'+label,[str(s.scratch/'adamic'),'types',*[str(p) for p in roots]],limit=1800)
sys.exit(r['exit_code'] if r['exit_code']>=0 else 128-r['exit_code'])
