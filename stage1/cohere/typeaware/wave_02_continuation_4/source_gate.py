#!/usr/bin/env python3
"""Audit .a modules with a pinned cohere CLI that does not yet enumerate .a sources."""
import argparse,json,subprocess
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('--scratch',required=True);parser.add_argument('--cohere',required=True);args=parser.parse_args()
own=Path(__file__).resolve().parent;root=own.parents[3];scratch=Path(args.scratch).resolve();scratch.mkdir(parents=True,exist_ok=True)
for file in list(own.parent.glob('*.ts'))+list(own.parent.glob('*.a')):
 source=file.read_text().replace("from '../../typescript/", "from '"+str(root/'stage1/typescript')+'/').replace("from '../lint/", "from '"+str(root/'stage1/cohere/lint')+'/').replace(".a'",".ts'")
 (scratch/file.with_suffix('.ts').name).write_text(source)
for file in own.glob('*.a'):
 source=file.read_text().replace("from '../../../typescript/", "from '"+str(root/'stage1/typescript')+'/').replace("from '../", "from './").replace(".a'",".ts'")
 (scratch/file.with_suffix('.ts').name).write_text(source)
config=json.loads((root/'tsconfig.json').read_text());config['files']=[str(root/'internal/load/prelude.d.ts')];config['include']=['*.ts'];config.pop('exclude',None);(scratch/'tsconfig.json').write_text(json.dumps(config))
(scratch/'CohereSettings.json').write_text((root/'CohereSettings.json').read_text())
files=[str(scratch/file.with_suffix('.ts').name) for file in own.glob('*.a')]
for name,flags in [('format',['-format-only','-no-fix']),('lint',['-lint','-no-fix'])]:
 with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:
  result=subprocess.run([args.cohere,'-directory',str(scratch),'-no-cache',*flags,*files],cwd=root,stdout=out,stderr=err)
 print(name,'exit',result.returncode,flush=True)
 if result.returncode:raise SystemExit(result.returncode)
