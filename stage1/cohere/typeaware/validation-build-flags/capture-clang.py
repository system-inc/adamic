#!/usr/bin/env python3
import os, sys, json, shlex
from pathlib import Path
real='/workspace/adamic-tools/llvm/bin/clang'
args=[real,*sys.argv[1:]]
Path(os.environ['CLANG_CAPTURE']).write_text(shlex.join(args)+'\n')
Path(os.environ['CLANG_CAPTURE']+'.json').write_text(json.dumps(args,indent=2)+'\n')
os.execv(real,args)
