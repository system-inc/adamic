#!/usr/bin/env python3
"""The policy audit must reject trusting an unresolved helper receiver."""
import gzip,json,pathlib,subprocess,sys
lane=pathlib.Path(__file__).resolve().parent;artifact=lane/'read-demand-sites.json.gz';original=artifact.read_bytes()
rows=json.loads(gzip.decompress(original))
helper=next(r for r in rows if r['file']=='src/compiler/utilities.ts' and r['line']==993 and r['text']=='module.valueDeclaration')
assert helper['shared_graph']['retain_check']
helper['shared_graph']['retain_check']=False
try:
 artifact.write_bytes(gzip.compress((json.dumps(rows)+'\n').encode(),mtime=0))
 caught=subprocess.run([sys.executable,str(lane/'audit-read-demand.py'),sys.argv[1],str(lane)],capture_output=True,text=True)
 print(caught.stdout,end='');print(caught.stderr,end='')
 assert caught.returncode!=0 and 'Unknown receiver lost its read check' in caught.stderr,'trust-unknown helper mutant escaped'
 print('PASS: trust-Unknown shared-helper measurement mutant caught by guard policy assertion')
finally:artifact.write_bytes(original)
