#!/usr/bin/env python3
"""Generate only three named owner handoffs for source integration validation."""
from pathlib import Path
import difflib
import json
import sys

root=Path(__file__).resolve().parents[3]
out=Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
changes={
 'internal/lower/view_contracts.go':('\t\tif !tagged {','\t\tif !tagged && !l.supportsUntaggedRead(contract) {'),
 'internal/native/view_unions.go':('\tfor _, field := range contract.Fields {','\tif untaggedObjectUnion(e.program.ViewContracts, contract) { e.viewUntaggedObjectUnion(property, object); return }\n\tfor _, field := range contract.Fields {'),
 'internal/javascript/view_unions.go':('\tfor _, field := range contract.Fields {','\tif untaggedObjectUnion(e.program.ViewContracts, contract) { return e.viewUntaggedObjectUnion(property, value) }\n\tfor _, field := range contract.Fields {'),
}
replace={};patch=[]
for name,(before,after) in changes.items():
 source=(root/name).read_text();assert source.count(before)==1,name
 result=source.replace(before,after)
 target=out/Path(name).name
 # Two backends share a basename, so each gets its own path.
 target=out/(name.replace('/','_'))
 target.write_text(result);replace[str(root/name)]=str(target)
 patch.extend(difflib.unified_diff(source.splitlines(True),result.splitlines(True),fromfile='a/'+name,tofile='b/'+name))
(out/'overlay.json').write_text(json.dumps({'Replace':replace},indent=2)+'\n')
(root/'stage3/interface-downcasts/untagged/source-hooks.patch').write_text(''.join(patch))
print('three named source hooks; production shared files unchanged')
