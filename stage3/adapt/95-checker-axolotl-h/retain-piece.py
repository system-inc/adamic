#!/usr/bin/env python3
"""Retain a completed oracle piece in one cumulative raw shard archive."""
import argparse
import json
from pathlib import Path
import shutil
import subprocess
import time

parser=argparse.ArgumentParser(description=__doc__)
for name in ('pieces','results','evidence','log','exit_file'):
    parser.add_argument(name,type=Path)
parser.add_argument('piece',type=int)
args=parser.parse_args()
args.evidence.mkdir(parents=True,exist_ok=True)
source=args.pieces/f'piece-{args.piece:02d}.json'
report=json.loads(source.read_text())
assert report['status']=='pass' and report['under25Minutes'] and report['seconds']<1500
assert args.exit_file.read_text().strip()=='0'
shutil.copyfile(source,args.evidence/source.name)
raw_pieces=args.results/'pieces';raw_pieces.mkdir(exist_ok=True)
shutil.copyfile(source,raw_pieces/source.name)
shutil.copyfile(args.log,raw_pieces/f'piece-{args.piece:02d}.log')
shutil.copyfile(args.exit_file,raw_pieces/f'piece-{args.piece:02d}.exit')
reports=[json.loads(p.read_text()) for p in sorted(args.evidence.glob('piece-??.json'))]
units=[unit for piece in reports for unit in piece['units']]
assert len({unit['name'] for unit in units})==len(units)
assert all(unit['exit']==0 and unit['status']=='pass' and unit['unit_seconds']<30 and unit['seconds']<30 for unit in units)
paths=[]
for unit in units:paths.extend([unit['name'],unit['name']+'.log'])
for piece in reports:
    for suffix in ('json','log','exit'):paths.append(f'pieces/piece-{piece["piece"]:02d}.{suffix}')
started=time.monotonic()
subprocess.run(['python3','-B',str(Path(__file__).with_name('pack-evidence.py')),str(args.evidence/'shards.tar.gz'),str(args.evidence/'shards.sha256'),str(args.results),*paths,'--prefix','shards'],check=True)
summary=dict(retainedPieces=[piece['piece'] for piece in reports],retainedUnits=len(units),packSeconds=time.monotonic()-started,rawArchive='shards.tar.gz',sha256List='shards.sha256',allMemberHashesVerified=True)
(args.evidence/'packed-shards.json').write_text(json.dumps(summary,indent=2)+'\n')
assert sum(p.is_file() for p in Path(__file__).parent.rglob('*'))<300
print(json.dumps(summary))
