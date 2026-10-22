#!/usr/bin/env python3
"""Reproduce the control's complete apply list, excluding only adaptation 91."""
import json
import os
from pathlib import Path
import sys
sys.dont_write_bytecode=True
repo=Path(__file__).resolve().parents[4]
sys.path.insert(0,str(repo/'stage3'))
import apply
source=apply.typescript_source(json.loads((repo/'stage3/source.json').read_text()))
out=Path('/workspace/scratch/adapt91-control')
apply.run('git','clone','--depth','1',source.resolve().as_uri(),str(out))
adaptations,env=apply.adaptation_inputs()
apply.adapt(out,[p for p in adaptations if p.name!='91-patcher-discard'],env)
