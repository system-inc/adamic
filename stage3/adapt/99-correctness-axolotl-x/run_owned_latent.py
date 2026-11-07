#!/usr/bin/env python3
"""Run and immediately publish each of the twelve bounded ownership-revision units."""
import json
import os
from pathlib import Path
import subprocess
import sys
repo=Path(__file__).resolve().parents[3];p=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch')
for index in range(6):
    for side,tree in [('before','/workspace/bind-before'),('after','/workspace/bind-owned-after')]:
        plan=scratch/'bind-owned-latent-plans'/side/(str(index)+'.json')
        dest=scratch/f'bind-owned-latent-{side}-{index}'
        if not (dest/'metrics.json').exists():
            env=dict(os.environ,LATENT_FULL='0',LATENT_SPECULATIVE='0',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/f'bind-owned-latent-{side}-{index}-times.jsonl'),ANY_MEMBERS_CENSUS_ROOT=tree,ANY_MEMBERS_CENSUS_PIECE=str(plan))
            command=['python3',str(p/'measure_piece.py'),str(dest),'--',str(scratch/'bind-bin/latent'),tree+'/src/compiler',str(scratch/f'bind-owned-latent-{side}-{index}-records.jsonl')]
            subprocess.run(command,cwd=repo,env=env,check=True)
        assert json.loads((dest/'metrics.json').read_text())['exit']==0
        subprocess.run(['python3',str(p/'archive_latent.py'),side,str(index),'--owned'],cwd=repo,check=True)
        subprocess.run(['python3',str(p/'publish_paths.py'),f'Ownership revision CW051 {side} unit {index}',str((p/'owned/latent'/side/str(index)).relative_to(repo))],cwd=repo,check=True)
