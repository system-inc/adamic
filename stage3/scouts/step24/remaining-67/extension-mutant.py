#!/usr/bin/env python3
import json, os, subprocess, sys
from pathlib import Path
repo=Path(__file__).resolve().parents[4]
tree=Path(sys.argv[1]); file=tree/'src/compiler/builder.ts'
original=file.read_bytes(); old=b'outSignature: buildInfo.outSignature'; new=b'outSignature: true'
assert original.count(old)==1
try:
 file.write_bytes(original.replace(old,new))
 result=subprocess.run(['node',str(repo/'stage3/adapt/20-optional-declarations/adapt.cjs'),str(tree)],capture_output=True,text=True)
 assert result.returncode!=0 and 'builder optional value contract drift: outSignature' in result.stderr,(result.returncode,result.stderr)
 Path(__file__).with_name('extension-mutant.json').write_text(json.dumps({'exit':result.returncode,'caught':'builder optional value contract drift: outSignature','stderr':result.stderr},indent=2)+'\n')
finally:file.write_bytes(original)
print('wrong-type builder initializer caught; restored bytes')
