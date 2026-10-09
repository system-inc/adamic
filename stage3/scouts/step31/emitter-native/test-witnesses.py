#!/usr/bin/env python3
"""Node truth, real checker outcomes, and independent source-input mutants."""
import argparse,json,os,subprocess
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('compiler');parser.add_argument('candidate',type=Path);parser.add_argument('output',type=Path);args=parser.parse_args()
args.output.mkdir();root=Path(__file__).resolve().parent
rows=json.loads((root/'witnesses.json').read_text());observations=[]
def run(command,prefix,env=None,cwd=None):
    with prefix.with_suffix('.stdout').open('wb') as out,prefix.with_suffix('.stderr').open('wb') as err:
        code=subprocess.run(command,stdout=out,stderr=err,env=env,cwd=cwd).returncode
    prefix.with_suffix('.exit').write_text(str(code)+'\n');return code,prefix.with_suffix('.stdout').read_bytes(),prefix.with_suffix('.stderr').read_bytes()
for row in rows:
    source=root/'witnesses'/(row['name']+'.a');prefix=args.output/row['name']
    node=run(['node',str(root/'node.mjs'),str(source)],prefix.with_name(prefix.name+'-node'))
    assert node==(0,row['stdout'].encode(),b''),(row['name'],'Node truth',node)
    builds=[]
    for split in ['0','1']:
        native=run([args.compiler,'build',str(source),'-o',str(args.output/'unused')],prefix.with_name(prefix.name+'-split-'+split),dict(os.environ,ADAMIC_NATIVE_SPLIT=split),args.candidate)
        assert native[0]==1 and ('error TS'+row['checkerCode']+':').encode() in native[2],(row['name'],'checker stop',native)
        builds.append(dict(split=split,exit=native[0],diagnostic=native[2].decode()))
    text=source.read_text();before=row['mutant']['before'];assert text.count(before)>=1
    mutant=args.output/(row['name']+'-mutant.a');mutant.write_text(text.replace(before,row['mutant']['after']))
    changed=run(['node',str(root/'node.mjs'),str(mutant)],prefix.with_name(prefix.name+'-mutant'))
    assert changed[0]==0 and changed[2]==b'' and changed[1]!=node[1],(row['name'],'source mutant survived',changed)
    observations.append(dict(name=row['name'],node=dict(exit=0,stdout=node[1].decode(),stderr=''),builds=builds,mutant=dict(caught_by='original Node stdout bytes',exit=0,stdout=changed[1].decode())))
    print('PASS '+row['name']+': Node, both checker stops, source mutant caught',flush=True)
(args.output/'report.json').write_text(json.dumps(observations,indent=2)+'\n')
