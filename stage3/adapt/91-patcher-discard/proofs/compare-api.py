#!/usr/bin/env python3
import hashlib
import json
from pathlib import Path
import sys
before, after, output = map(Path, sys.argv[1:])
a,b=before.read_bytes(),after.read_bytes()
assert a==b,'public typescript.d.ts bytes differ'
output.write_text(json.dumps(dict(byteIdentical=True,bytes=len(a),sha256=hashlib.sha256(a).hexdigest()),indent=2)+'\n')
print('PASS byte-identical public typescript.d.ts')
