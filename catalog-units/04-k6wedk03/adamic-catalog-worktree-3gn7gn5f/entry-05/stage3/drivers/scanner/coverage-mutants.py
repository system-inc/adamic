#!/usr/bin/env python3
"""Plant source mutants in scratch slices and compare the complete scanner dump."""
import argparse, hashlib, json, os, shutil, subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('slice',type=Path);p.add_argument('oracle',type=Path);p.add_argument('inputs',type=Path);p.add_argument('output',type=Path);a=p.parse_args()
here=Path(__file__).resolve().parent; a.output.mkdir(parents=True,exist_ok=False)
mutations={
 'escape':('case CharacterCodes.n:\n                return "\\n";', 'case CharacterCodes.n:\n                return "\\r";'),
 'error':('onError(message, length || 0, arg0);','if (!scannerProofOmittedError) { scannerProofOmittedError = true; } else onError(message, length || 0, arg0);'),
 'trivia':('return token = SyntaxKind.SingleLineCommentTrivia;','return token = SyntaxKind.MultiLineCommentTrivia;'),
}
results={}
for name,(before,after) in mutations.items():
 tree=a.output/(name+'-tree');shutil.copytree(a.slice,tree)
 file=tree/'src/compiler/scanner.ts';text=file.read_text();assert text.count(before)==1,(name,text.count(before));text=text.replace(before,after)
 if name=='error':text='let scannerProofOmittedError = false;\n'+text
 file.write_text(text)
 run=a.output/(name+'-run')
 with (a.output/(name+'.log')).open('wb') as log:
  code=subprocess.run(['python3',str(here/'run.py'),str(run),'--tree',str(tree),'--inputs',str(a.inputs),'--oracle',str(a.oracle),'--node-only'],stdout=log,stderr=subprocess.STDOUT).returncode
 assert code!=0,(name,'mutant survived')
 assert (run/'node.stderr').read_bytes()==b'',(name,'not a successful Node execution')
 assert (run/'full-tree-diff.stdout').stat().st_size>0,(name,'no byte difference')
 original=a.oracle.read_bytes().splitlines();changed=(run/'node.stdout').read_bytes().splitlines()
 if name=='error':
  assert len(changed)==len(original)-1,'exactly one scanner error must disappear'
  assert sum(l.startswith(b'error\t') for l in changed)==sum(l.startswith(b'error\t') for l in original)-1
 elif name=='escape':
  differences=[(x.split(b'\t'),y.split(b'\t')) for x,y in zip(original,changed) if x!=y]
  assert differences and all(x[:6]==y[:6] and x[6]!=y[6] for x,y in differences),'escape mutation must change cooked values only'
 else:
  assert len(original)==len(changed)
  differences=[(x.split(b'\t'),y.split(b'\t')) for x,y in zip(original,changed) if x!=y]
  assert differences and all(x[0] in (b'SingleLineCommentTrivia', b'FirstTriviaToken') and y[0]==b'MultiLineCommentTrivia' and x[1:]==y[1:] for x,y in differences)
 results[name]={'node_exit':0,'comparison_exit':1,'rows_before':len(original),'rows_after':len(changed),'sha256':hashlib.sha256((run/'node.stdout').read_bytes()).hexdigest()}
 (a.output/'report.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
