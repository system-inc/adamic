from run import run,E
import json
rows=json.loads((E/'scope.json').read_text())
# Fast checks first; then independently bounded native/full-witness rows.
rows.sort(key=lambda r:r[0] in ['TestMutants','TestNodeTableIsLinkOnlyFamily','TestOwnedWitnesses'])
for name,pattern in rows:
 for i in range(1,4):
  r=run('timing-'+name+'-'+str(i),pattern)
  if r['cooked']:break
  if r['exit']!=0:
   (E/'RED-BASELINE').write_text(name+' '+str(r))
   raise SystemExit('red row baseline; stop audit')
