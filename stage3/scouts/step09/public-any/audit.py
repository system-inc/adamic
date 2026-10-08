#!/usr/bin/env python3
"""Require exact residue coverage, complete acceptance capture coverage and fixture counts."""
import gzip
import hashlib
import json
from pathlib import Path
import re
import sys
H=Path(__file__).resolve().parent
ledger=Path(sys.argv[1]) if len(sys.argv)>1 else H/'contracts.json'
l=json.loads(ledger.read_text())
expected={(m[1],int(m[2]),int(m[3])) for m in re.finditer(r'\| (src/compiler/[^:]+):(\d+):(\d+) \|', (H/'evidence/RESIDUE.md').read_text())}
actual={(r['file'],r['line'],r['column']) for r in l['rows']}
assert len(l['rows'])==len(expected)==20 and actual==expected,'dropped or changed residue row'
assert len({r['id'] for r in l['rows']})==20
assert all(r['public_declarations'] and len(r['options'])==3 and r['decision']=='undecided' for r in l['rows'])
for r in l['rows']:
 assert r['acceptance_observation']['projects'] in range(302)
with gzip.open(H/'evidence/project-captures.json.gz','rt') as f:captures=json.load(f)
assert len(captures)==len({r['project'] for r in captures})==301,'dropped acceptance project capture'
o=json.loads((H/'evidence/observations.json').read_text())
assert o['acceptance']['cases']==o['acceptance']['passed']==o['projects']==301 and not o['acceptance']['failed']
assert {r['project'] for r in captures}==set(o['captures_sha256'])
for r in captures:assert hashlib.sha256(json.dumps(r,sort_keys=True).encode()).hexdigest()==o['captures_sha256'][r['project']]
status=json.loads((H/'fixtures/status.json').read_text())
assert len(status)==3 and {r['file'] for r in status}=={p.name for p in (H/'fixtures').glob('*.a')}
for r in status:
 assert r['node']['exit']==r['mutant']['node_exit']==0 and not r['node']['stderr']
 assert r['mutant']['stdout']!=r['node']['stdout'],'fixture mutant survived'
if len(sys.argv)>2:
 tree=Path(sys.argv[2])
 for file,digest in l['source_hashes'].items():assert hashlib.sha256((tree/file).read_bytes()).hexdigest()==digest,file
print('PASS: 19 source contracts plus one checker anomaly, 301 complete captures, three Node fixtures and killed source mutants')
