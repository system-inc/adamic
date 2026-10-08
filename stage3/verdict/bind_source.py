#!/usr/bin/env python3
"""Bind an executable to its builder-supplied source and runtime dependencies."""
import argparse
import json
from pathlib import Path
import tempfile
from differential import inventory, sha
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('binary',type=Path);p.add_argument('source',type=Path)
p.add_argument('--artifact',action='append',type=Path,default=[])
a=p.parse_args();binary=a.binary.resolve();source=a.source.resolve()
with tempfile.TemporaryDirectory(prefix='verdict-bind-') as scratch:
 data=inventory(source,Path(scratch)/'inventory.json')
result={'schema':1,'source_root':str(source),'source_sha256':data['source_sha256'],
 'binary_sha256':sha(binary.read_bytes()),'artifacts':{str(f.resolve()):sha(f.read_bytes()) for f in a.artifact}}
Path(str(binary)+'.source.json').write_text(json.dumps(result,indent=2)+'\n')
