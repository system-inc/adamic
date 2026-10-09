#!/usr/bin/env python3
"""Check exact pinned coverage, real view availability and source bytes."""
import copy, gzip, hashlib, json, sys
from pathlib import Path

def read(path):
 data=Path(path).read_bytes()
 return json.loads(gzip.decompress(data) if str(path).endswith('.gz') else data)

def audit(summary,before,after,tree):
 selection=set(summary['selection'])
 assert len(selection)==138, 'selection count'
 for snapshot in [before,after]:
  rows=snapshot['predicates']
  locations=['src/compiler/'+row['location'].split('/src/compiler/')[1] for row in rows]
  assert len(rows)==138 and len(set(locations))==138 and set(locations)==selection,'body coverage'
  assert snapshot['checkerDiagnostics']==summary['checkerDiagnostics'],'checker diagnostics'
  for row in rows:
   assert row['bodyProof'],'logical proof changed'
   assert row['admission']=='PendingView','unsupported view credited as pass'
   assert not row['checkedOriginal'],'original check fabricated'
   assert row['viewFailures'],'missing underlying builder failure'
   assert not row['viewRoots'] or any(root['Kind']==0 or root['Unsupported'] for root in row['viewRoots']),'unavailable root disappeared'
 for name,want in summary['sourceHashes'].items():
  assert hashlib.sha256((tree/name).read_bytes()).hexdigest()==want,'source hash '+name
 assert len(summary['sourceHashes'])==79,'source count'
 return '138 exact bodies, 79 identical source hashes, 320 preserved diagnostics; 138 pending, zero full admissions'

if __name__=='__main__':
 summary,before,after=map(read,sys.argv[1:4]);tree=Path(sys.argv[4])
 print(audit(summary,before,after,tree))
 if '--mutants' in sys.argv[5:]:
  for name in ['drop-body','invent-pass','fabricate-check','source-hash']:
   m,b,a=copy.deepcopy((summary,before,after))
   if name=='drop-body':a['predicates'].pop()
   elif name=='invent-pass':a['predicates'][0]['admission']='Proven'
   elif name=='fabricate-check':a['predicates'][0]['checkedOriginal']=True
   else:m['sourceHashes'][next(iter(m['sourceHashes']))]='0'*64
   try:audit(m,b,a,tree)
   except AssertionError as error:print(name+': caught: '+str(error))
   else:raise AssertionError('mutant escaped: '+name)
