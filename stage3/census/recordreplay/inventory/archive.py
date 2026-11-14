from pathlib import Path
import json,shutil,gzip,hashlib
D=Path('/workspace/scratch/recordreplay');O=D/'inventory-audit';P=D/'head/stage3/adapt/99-wasm-crypto-import/proofs/recordreplay'
for name in ['99-current.json','specific.json','final-readers.json']:
 shutil.copyfile(O/name,P/('inventory-'+name))
readme=D/'head/stage3/adapt/99-wasm-crypto-import/README.md'
readme.write_text(readme.read_text()+'''\n## Recordreplay inventory follow-up\n\nThe current pre-99 tree has 747 source files, including\n`src/deprecatedCompat/controlFlowGraph.ts`; the amended tree has 748. Archived\n746-file proof records describe the older run and are retained as history.\n`proofs/check.cjs` now requires exactly the input inventory plus\n`src/recordreplay.d.ts`, checks the declaration's canonical bytes, hashes every\nsource, and reverses the erased edits before checking crypto scope.\n`proofs/scoped.py` plans the two erased sys edits and declaration, compares the\ncomplete dynamic source map, and updates its two-file and mutation expectations.\nNo foreign proof or production manifest was changed.\n\nThe new `proofs/recordreplay/inventory-*.json` records are current observations.\nThe complete source checker passes and rejects declaration/extra-file mutants.\nThe scoped driver's actual planning/inventory segment passes; its historical\nfull apply/build/lane pipeline was not rerun. Adapted source bytes are unchanged\nfrom the accepted read gate.\n''')
C=D/'base/stage3/census/recordreplay/inventory';C.mkdir(exist_ok=True)
for p in O.iterdir():
 if p.is_file() and (p.suffix in ['.json','.py','.cjs'] or p.name.endswith('-reader.go')):
  name=p.name+'.txt' if p.suffix=='.go' else p.name
  data=p.read_bytes()
  if len(data)>100000:
   (C/(name+'.gz')).write_bytes(gzip.compress(data,mtime=0))
  else:(C/name).write_bytes(data)
# Per-reader classification is an index, not a claim to have run historical pipelines.
cat=json.loads((O/'catalog.json').read_text());js=json.loads((O/'js-readers.json').read_text());py=json.loads((O/'python-readers.json').read_text())
index=[]
for e in cat['readers']:
 probes=[x for x in js+py if x['file']==e['file'] and x.get('results')]
 index.append(dict(file=e['file'],sites=e['sites'],source_sha256=e['sha256'],inventory_projections=len(probes),method='see original-function/comprehension probes' if probes else 'static caller review; no isolated executable projection recorded'))
(C/'reader-index.json').write_text(json.dumps(index,indent=2)+'\n')
