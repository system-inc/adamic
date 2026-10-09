#!/usr/bin/env python3
"""Reconcile the complete rule inventory without claiming latent completeness."""
import collections
import json
from pathlib import Path
import sys
before, after, output = map(Path, sys.argv[1:])
def rows(file):
    return [r for line in file.read_text().splitlines() if 'summary' not in (r := json.loads(line))]
b, a = rows(before), rows(after)
# Line numbers move with owner additions. Expression text and relation are stable.
def key(r):
    return tuple(r[n] for n in ['File','Kind','Property','Source','Target','Text'])
left, right = collections.Counter(map(key,b)), collections.Counter(map(key,a))
assert not right-left, 'new optional-widening sites require review'
removed = left-right
cleared = []
for r in b:
    k=key(r)
    if removed[k]: cleared.append(r);removed[k]-=1
ranking=collections.Counter((r['Source'],r['Property']) for r in b)
remaining=[]
for r in a:
    source, prop=r['Source'],r['Property']
    if source=='never': reason='Uninhabited narrowed source has no owning object declaration. Compiler-side relation review required; no source edit.'
    elif source=='PackageJsonPathFields': reason='Raw parsed JSON can hold a non-string version; readPackageJsonField checks its runtime type. Adding version?: string is not truthful.'
    elif prop=='constraint': reason='General Type and its other variants are not TypeParameter shapes. Declaring constraint alone exposes the next missing TypeParameter field; target/mapper contracts require separate subtype proof.'
    elif source=='SourceFile' and prop=='getPositionOfLineAndCharacter': reason='Compiler-only interface omits a method required by the services augmentation. A direct optional member conflicts with that merge; a heritage change would require API proof beyond the permitted optional additions.'
    elif source=='SourceFile' and prop=='skipTrivia': reason='SourceFile is passed through a SourceMapSource view. A truthful universal skipTrivia writer/alias contract is not proved; declined separately from the services method augmentation.'
    elif prop=='source': reason='TextRange/node fallback is a different shape from SourceMapRange. The scratch TextRange.source candidate enlarged the inventory from 399 to 859; a universal source declaration is not justified. Declined pending narrower owner proof.'
    elif prop=='all': reason='CompilerOptions is a distinct shape from the source options/Pick/indexed record. Adding only all would expose more target option fields; no truthful universal declaration proved.'
    else: reason='No complete construction, alias and writer proof for this source owner in this wave. Declined without changing source or weakening checker options.'
    remaining.append({**r,'reason':reason})
output.mkdir(parents=True,exist_ok=True)
(output/'census.json').write_text(json.dumps({'before':len(b),'after':len(a),'cleared':len(cleared),'cleared_sites':cleared,'owners_ranked_before':[{'source':s,'property':p,'sites':c} for (s,p),c in ranking.most_common()]},indent=2)+'\n')
(output/'remaining-sites.json').write_text(json.dumps(remaining,indent=2)+'\n')
lines=['# Remaining optional-widening sites','',f'{len(a)} sites. Each retains its source and target type and exact expression in remaining-sites.json.','', '| Site | Source | Target property | Reason |','|---|---|---|---|']
for r in remaining:
    lines.append(f"| {r['File']}:{r['Line']}:{r['Column']} | {r['Source'].replace('|', '&#124;')} | {r['Property']} | {r['reason']} |")
(output/'remaining-sites.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'before':len(b),'after':len(a),'cleared':len(cleared)}))
