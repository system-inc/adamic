import sys
sys.dont_write_bytecode = True
import json,os,sys
from pathlib import Path
from importlib.machinery import SourceFileLoader
here=Path(__file__).resolve().parent
s=SourceFileLoader('supervise',str(here/'clear-supervise.py')).load_module()
label=sys.argv[1];tree=s.scratch/({'baseline':'baseline','after':'before','mutant':'mutant'}[label]);out=s.scratch/('census-'+label);out.mkdir(exist_ok=True)
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'unit-times.jsonl'))
r=s.run('census-'+label,[str(s.scratch/'latent'),str(tree/'src/compiler'),str(out/'raw.jsonl')],env=env,watch=[out/'raw.jsonl',out/'unit-times.jsonl'],limit=3600)
(out/'result.json').write_text(json.dumps(r,indent=2)+'\n')
sys.exit(r['exit_code'] if r['exit_code']>=0 else 128-r['exit_code'])
